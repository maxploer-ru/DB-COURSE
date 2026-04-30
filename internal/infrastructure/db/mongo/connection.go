package mongo

import (
	"ZVideo/internal/infrastructure/config"
	"context"
	"fmt"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Connection struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewConnection(cfg config.MongoConfig) (*Connection, error) {
	uri := cfg.URI
	if uri == "" {
		uri = buildMongoURI(cfg)
	}

	opts := options.Client().ApplyURI(uri)
	if cfg.ConnectTimeout > 0 {
		opts.SetConnectTimeout(cfg.ConnectTimeout)
	}
	if cfg.ServerSelectionTimeout > 0 {
		opts.SetServerSelectionTimeout(cfg.ServerSelectionTimeout)
	}
	if cfg.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize > 0 {
		opts.SetMinPoolSize(cfg.MinPoolSize)
	}

	connectTimeout := cfg.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("connect to mongo: %w", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	db := client.Database(cfg.Database)
	if err := EnsureIndexes(ctx, db); err != nil {
		return nil, fmt.Errorf("ensure mongo indexes: %w", err)
	}

	return &Connection{Client: client, DB: db}, nil
}

func (c *Connection) Close(ctx context.Context) error {
	return c.Client.Disconnect(ctx)
}

func buildMongoURI(cfg config.MongoConfig) string {
	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Port
	if port == 0 {
		port = 27017
	}
	database := cfg.Database
	if database == "" {
		database = "zvideo"
	}
	authSource := cfg.AuthSource
	if authSource == "" {
		authSource = database
	}

	auth := ""
	if cfg.User != "" {
		user := url.QueryEscape(cfg.User)
		pass := url.QueryEscape(cfg.Password)
		auth = fmt.Sprintf("%s:%s@", user, pass)
	}

	if auth == "" {
		return fmt.Sprintf("mongodb://%s:%d/%s", host, port, database)
	}
	return fmt.Sprintf("mongodb://%s%s:%d/%s?authSource=%s", auth, host, port, database, authSource)
}
