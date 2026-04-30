package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"ZVideo/internal/infrastructure/config"
	"ZVideo/internal/infrastructure/db/mongo"
	"ZVideo/internal/infrastructure/db/mongo/migrate"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	var direction string
	var steps int
	var driver string
	var migrationsPath string

	flag.StringVar(&direction, "direction", "up", "migration direction (up/down)")
	flag.IntVar(&steps, "steps", 0, "number of steps to migrate (0 = all)")
	flag.StringVar(&driver, "driver", "", "database driver (postgres/mongo)")
	flag.StringVar(&migrationsPath, "path", "", "path to SQL migrations for postgres")
	flag.Parse()

	if direction != "up" && direction != "down" {
		log.Fatal("Invalid migration direction: ", direction)
	}

	if driver == "" {
		driver = getEnv("DB_DRIVER", "postgres")
	}
	if migrationsPath == "" {
		migrationsPath = "./migrations/postgres"
	}

	switch driver {
	case "postgres":
		if err := migratePostgres(migrationsPath, direction, steps); err != nil {
			log.Fatal("Migration failed:", err)
		}
	case "mongo":
		if err := migrateMongo(direction, steps); err != nil {
			log.Fatal("Migration failed:", err)
		}
	default:
		log.Fatal("Unknown driver: ", driver)
	}

	log.Println("Migration completed successfully!")
}

func migratePostgres(migrationsPath, direction string, steps int) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "1488"),
		getEnv("DB_NAME", "zvideo"),
	)

	log.Println("Connecting to Postgres...")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Connected successfully!")

	if err := ensurePgMigrationTable(db); err != nil {
		return err
	}

	const lockID int64 = 742394821
	if err := db.Exec("SELECT pg_advisory_lock(?)", lockID).Error; err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_ = db.Exec("SELECT pg_advisory_unlock(?)", lockID).Error
	}()

	files, err := getPgMigrationFiles(migrationsPath)
	if err != nil {
		return err
	}

	return applyPgMigrations(db, files, direction, steps)
}

func migrateMongo(direction string, steps int) error {
	cfg := config.MongoConfig{
		URI:                    getEnv("MONGO_URI", ""),
		Host:                   getEnv("MONGO_HOST", "localhost"),
		Port:                   getEnvAsInt("MONGO_PORT", 27017),
		User:                   getEnv("MONGO_USER", ""),
		Password:               getEnv("MONGO_PASSWORD", ""),
		Database:               getEnv("MONGO_DB", "zvideo"),
		AuthSource:             getEnv("MONGO_AUTH_SOURCE", "zvideo"),
		ConnectTimeout:         getEnvAsDuration("MONGO_CONNECT_TIMEOUT", 10*time.Second),
		ServerSelectionTimeout: getEnvAsDuration("MONGO_SERVER_SELECTION_TIMEOUT", 5*time.Second),
		MaxPoolSize:            uint64(getEnvAsInt("MONGO_MAX_POOL_SIZE", 50)),
		MinPoolSize:            uint64(getEnvAsInt("MONGO_MIN_POOL_SIZE", 0)),
	} // TODO: refactor

	log.Println("Connecting to MongoDB...")
	conn, err := mongo.NewConnection(cfg)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close(context.Background())
	}()

	log.Println("Connected successfully!")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return migrate.Apply(ctx, conn.DB, direction, steps, migrate.DefaultMigrations())
}

type pgMigrationFile struct {
	Version  string
	UpPath   string
	DownPath string
}

func applyPgMigrations(db *gorm.DB, files []pgMigrationFile, direction string, steps int) error {
	applied, err := listAppliedPgMigrations(db)
	if err != nil {
		return err
	}

	if direction == "up" {
		return applyPgUp(db, files, applied, steps)
	}

	return applyPgDown(db, files, applied, steps)
}

func applyPgUp(db *gorm.DB, files []pgMigrationFile, applied map[string]string, steps int) error {
	var pending []pgMigrationFile
	for _, file := range files {
		if _, ok := applied[file.Version]; !ok {
			pending = append(pending, file)
		}
	}

	if steps > 0 && steps < len(pending) {
		pending = pending[:steps]
	}

	if len(pending) == 0 {
		log.Println("No migrations to apply")
		return nil
	}

	log.Printf("Found %d migration(s) to apply\n", len(pending))
	for _, file := range pending {
		if file.UpPath == "" {
			return fmt.Errorf("missing up migration for %s", file.Version)
		}
		if err := runPgMigration(db, file.Version, file.UpPath, true, applied); err != nil {
			return err
		}
	}
	return nil
}

func applyPgDown(db *gorm.DB, files []pgMigrationFile, applied map[string]string, steps int) error {
	orderedApplied, err := listAppliedPgMigrationsOrdered(db, "desc")
	if err != nil {
		return err
	}

	if steps > 0 && steps < len(orderedApplied) {
		orderedApplied = orderedApplied[:steps]
	}

	if len(orderedApplied) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	fileByVersion := map[string]pgMigrationFile{}
	for _, file := range files {
		fileByVersion[file.Version] = file
	}

	log.Printf("Found %d migration(s) to rollback\n", len(orderedApplied))
	for _, version := range orderedApplied {
		file := fileByVersion[version]
		if file.DownPath == "" {
			return fmt.Errorf("missing down migration for %s", version)
		}
		if err := runPgMigration(db, file.Version, file.DownPath, false, applied); err != nil {
			return err
		}
	}
	return nil
}

func runPgMigration(db *gorm.DB, version, path string, isUp bool, applied map[string]string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}
	checksum := checksumSQL(content)
	if existing, ok := applied[version]; ok && isUp {
		if existing != checksum {
			return fmt.Errorf("checksum mismatch for %s", version)
		}
		return nil
	}

	log.Printf("Applying: %s", filepath.Base(path))
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("failed to execute %s: %w", path, err)
		}
		if isUp {
			if err := tx.Exec("INSERT INTO schema_migrations(version, checksum) VALUES (?, ?)", version, checksum).Error; err != nil {
				return fmt.Errorf("record migration %s: %w", version, err)
			}
		} else {
			if err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", version).Error; err != nil {
				return fmt.Errorf("remove migration %s: %w", version, err)
			}
		}
		return nil
	})
}

func ensurePgMigrationTable(db *gorm.DB) error {
	stmt := `
CREATE TABLE IF NOT EXISTS schema_migrations
(
    version   TEXT PRIMARY KEY,
    checksum  TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`
	return db.Exec(stmt).Error
}

func listAppliedPgMigrations(db *gorm.DB) (map[string]string, error) {
	rows, err := db.Raw("SELECT version, checksum FROM schema_migrations").Rows()
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var version, checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, err
		}
		result[version] = checksum
	}
	return result, nil
}

func listAppliedPgMigrationsOrdered(db *gorm.DB, order string) ([]string, error) {
	if order != "asc" && order != "desc" {
		return nil, errors.New("invalid order")
	}
	query := fmt.Sprintf("SELECT version FROM schema_migrations ORDER BY applied_at %s", order)
	rows, err := db.Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, nil
}

func getPgMigrationFiles(migrationsPath string) ([]pgMigrationFile, error) {
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory not found: %s", migrationsPath)
	}

	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, err
	}

	fileMap := map[string]*pgMigrationFile{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		var direction string
		if strings.HasSuffix(name, ".up.sql") {
			direction = "up"
		} else if strings.HasSuffix(name, ".down.sql") {
			direction = "down"
		} else {
			continue
		}
		version := strings.TrimSuffix(name, fmt.Sprintf(".%s.sql", direction))
		if version == "" {
			continue
		}

		entryValue := fileMap[version]
		if entryValue == nil {
			entryValue = &pgMigrationFile{Version: version}
			fileMap[version] = entryValue
		}
		fullPath := filepath.Join(migrationsPath, name)
		if direction == "up" {
			entryValue.UpPath = fullPath
		} else {
			entryValue.DownPath = fullPath
		}
	}

	var files []pgMigrationFile
	for _, value := range fileMap {
		files = append(files, *value)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Version < files[j].Version
	})

	return files, nil
}

func checksumSQL(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
