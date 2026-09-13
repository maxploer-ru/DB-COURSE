package repository_test

import (
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

	log.Println("Starting shared Postgres Testcontainer...")
	pgContainer, err := db.NewPostgresContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to start Postgres container: %v", err)
	}

	sharedDB = pgContainer.DB

	code := m.Run()

	log.Println("Terminating shared Postgres Testcontainer...")
	if err := pgContainer.Terminate(ctx); err != nil {
		log.Fatalf("Failed to terminate Postgres container: %v", err)
	}

	os.Exit(code)
}
