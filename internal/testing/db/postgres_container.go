package db

import (
	"context"
	"path/filepath"
	"runtime"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	driver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresContainer struct {
	DB        *gorm.DB
	container *postgres.PostgresContainer
}

func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	_, b, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(b), "../../..")

	scripts := []string{
		filepath.Join(root, "migrations/postgres/001_init.up.sql"),
		filepath.Join(root, "migrations/postgres/002_notification_function.up.sql"),
		filepath.Join(root, "migrations/postgres/003_community.up.sql"),
	}

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithInitScripts(scripts...),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(3*time.Minute),
		),
	)
	if err != nil {
		return nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(driver.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{
		DB:        db,
		container: pgContainer,
	}, nil
}

func (c *PostgresContainer) Terminate(ctx context.Context) error {
	return c.container.Terminate(ctx)
}
