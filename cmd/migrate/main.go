package main

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/config"
	applogger "ZVideo/internal/infrastructure/logger"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	cfg := config.LoadConfig()
	appLogger, closeLog := applogger.NewConfigured(cfg.Logging)
	defer closeLog()

	if err := run(appLogger); err != nil {
		appLogger.ErrorContext(context.Background(), "migration process failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(appLogger domain.Logger) error {
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
		return fmt.Errorf("invalid migration direction %q", direction)
	}

	if driver == "" {
		driver = getEnv("DB_DRIVER", "postgres")
	}
	if migrationsPath == "" {
		migrationsPath = "./migrations/postgres"
	}

	switch driver {
	case "postgres":
		if err := migratePostgres(migrationsPath, direction, steps, appLogger); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	default:
		return fmt.Errorf("unknown driver %q", driver)
	}

	appLogger.InfoContext(context.Background(), "migration completed successfully")
	return nil
}

func migratePostgres(migrationsPath, direction string, steps int, appLogger domain.Logger) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "1488"),
		getEnv("DB_NAME", "zvideo"),
		getEnv("DB_SSLMODE", "disable"),
	)

	appLogger.InfoContext(context.Background(), "connecting to Postgres")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	appLogger.InfoContext(context.Background(), "connected to Postgres")

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

	return applyPgMigrations(db, files, direction, steps, appLogger)
}

type pgMigrationFile struct {
	Version  string
	UpPath   string
	DownPath string
}

func applyPgMigrations(db *gorm.DB, files []pgMigrationFile, direction string, steps int, appLogger domain.Logger) error {
	applied, err := listAppliedPgMigrations(db)
	if err != nil {
		return err
	}

	if direction == "up" {
		return applyPgUp(db, files, applied, steps, appLogger)
	}

	return applyPgDown(db, files, applied, steps, appLogger)
}

func applyPgUp(db *gorm.DB, files []pgMigrationFile, applied map[string]string, steps int, appLogger domain.Logger) error {
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
		appLogger.InfoContext(context.Background(), "no migrations to apply")
		return nil
	}

	appLogger.InfoContext(context.Background(), "migrations to apply", slog.Int("count", len(pending)))
	for _, file := range pending {
		if file.UpPath == "" {
			return fmt.Errorf("missing up migration for %s", file.Version)
		}
		if err := runPgMigration(db, file.Version, file.UpPath, true, applied, appLogger); err != nil {
			return err
		}
	}
	return nil
}

func applyPgDown(db *gorm.DB, files []pgMigrationFile, applied map[string]string, steps int, appLogger domain.Logger) error {
	orderedApplied, err := listAppliedPgMigrationsOrdered(db, "desc")
	if err != nil {
		return err
	}

	if steps > 0 && steps < len(orderedApplied) {
		orderedApplied = orderedApplied[:steps]
	}

	if len(orderedApplied) == 0 {
		appLogger.InfoContext(context.Background(), "no migrations to rollback")
		return nil
	}

	fileByVersion := map[string]pgMigrationFile{}
	for _, file := range files {
		fileByVersion[file.Version] = file
	}

	appLogger.InfoContext(context.Background(), "migrations to rollback", slog.Int("count", len(orderedApplied)))
	for _, version := range orderedApplied {
		file := fileByVersion[version]
		if file.DownPath == "" {
			return fmt.Errorf("missing down migration for %s", version)
		}
		if err := runPgMigration(db, file.Version, file.DownPath, false, applied, appLogger); err != nil {
			return err
		}
	}
	return nil
}

func runPgMigration(db *gorm.DB, version, path string, isUp bool, applied map[string]string, appLogger domain.Logger) error {
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

	appLogger.InfoContext(context.Background(), "applying migration",
		slog.String("version", version), slog.String("file", filepath.Base(path)), slog.Bool("up", isUp))
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
