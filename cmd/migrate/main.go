package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	var direction string
	var steps int

	flag.StringVar(&direction, "direction", "up", "migration direction (up/down)")
	flag.IntVar(&steps, "steps", 0, "number of steps to migrate (0 = all)")
	flag.Parse()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "1488"),
		getEnv("DB_NAME", "zvideo"),
	)

	log.Println("Connecting to database...")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Connected successfully!")

	migrationsPath := "./migrations/postgres"

	if err := applyMigrations(db, migrationsPath, direction, steps); err != nil {
		log.Fatal("Migration failed:", err)
	}

	log.Println("Migration completed successfully!")
}

func applyMigrations(db *gorm.DB, migrationsPath, direction string, steps int) error {
	files, err := getMigrationFiles(migrationsPath, direction)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	if steps > 0 && steps < len(files) {
		files = files[:steps]
	}

	if len(files) == 0 {
		log.Println("No migrations to apply")
		return nil
	}

	log.Printf("Found %d migration(s) to apply\n", len(files))

	for _, file := range files {
		log.Printf("Applying: %s", filepath.Base(file))

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		if err := db.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("failed to execute %s: %w", file, err)
		}

		log.Printf("✅ Applied: %s", filepath.Base(file))
	}

	return nil
}

func getMigrationFiles(migrationsPath, direction string) ([]string, error) {
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory not found: %s", migrationsPath)
	}

	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, err
	}

	var files []string
	suffix := fmt.Sprintf(".%s.sql", direction)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, suffix) {
			fullPath := filepath.Join(migrationsPath, name)
			files = append(files, fullPath)
		}
	}

	sort.Strings(files)

	return files, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
