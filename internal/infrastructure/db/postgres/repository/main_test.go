package repository_test

import (
	"ZVideo/internal/infrastructure/config"
	postgresdb "ZVideo/internal/infrastructure/db/postgres"
	"ZVideo/internal/testing/db"
	"context"
	"log"
	"os"
	"testing"

	"gorm.io/gorm"
)

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	var cleanup func() error
	if os.Getenv("TEST_DATABASE_EXTERNAL") == "true" {
		log.Println("Connecting to external Postgres test database...")
		cfg := config.LoadConfig()
		var err error
		sharedDB, err = postgresdb.NewConnection(cfg.Database)
		if err != nil {
			log.Fatalf("Failed to connect to external Postgres test database: %v", err)
		}
		cleanup = closeDatabase
	} else {
		log.Println("Starting shared Postgres Testcontainer...")
		pgContainer, err := db.NewPostgresContainer(ctx)
		if err != nil {
			log.Fatalf("Failed to start Postgres container: %v", err)
		}
		sharedDB = pgContainer.DB
		cleanup = func() error {
			log.Println("Terminating shared Postgres Testcontainer...")
			return pgContainer.Terminate(ctx)
		}
	}

	code := m.Run()

	if err := cleanup(); err != nil {
		log.Printf("Failed to clean up Postgres test database: %v", err)
		code = 1
	}

	os.Exit(code)
}

func closeDatabase() error {
	sqlDB, err := sharedDB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
