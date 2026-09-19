package integration_test

import (
	"ZVideo/internal/infrastructure/config"
	"ZVideo/internal/infrastructure/db/postgres"
	"log"
	"os"
	"testing"

	"gorm.io/gorm"
)

var integrationDB *gorm.DB

func TestMain(m *testing.M) {
	cfg := config.LoadConfig()
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		log.Printf("integration database is unavailable: %v", err)
		os.Exit(1)
	}
	integrationDB = db

	code := m.Run()

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("close integration database: %v", err)
		os.Exit(1)
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("close integration database: %v", err)
		os.Exit(1)
	}
	os.Exit(code)
}
