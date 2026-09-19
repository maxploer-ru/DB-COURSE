package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseDriver string
	Database       DatabaseConfig
	Redis          RedisConfig
	JWT            JWTConfig
	HTTP           HTTPConfig
	Auth           AuthConfig
	Minio          MinioConfig
	Logging        LoggingConfig
}

type LoggingConfig struct {
	Level      string
	OutputPath string
	AddSource  bool
}

type HTTPConfig struct {
	Port            string
	Env             string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type MinioConfig struct {
	Endpoint         string
	ExternalEndpoint string
	AccessKey        string
	SecretKey        string
	Bucket           string
	UseSSL           bool
}

type AuthConfig struct {
	Password PasswordConfig
}

type PasswordConfig struct {
	MinLength int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	Database int
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		DatabaseDriver: getEnv("DB_DRIVER", "postgres"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "1488"),
			DBName:   getEnv("DB_NAME", "zvideo"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			Database: getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenTTL:  getEnvAsDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvAsDuration("JWT_REFRESH_TTL", 168*time.Hour),
		},
		HTTP: HTTPConfig{
			Port:            getEnv("HTTP_PORT", "8080"),
			Env:             getEnv("APP_ENV", "development"),
			ReadTimeout:     getEnvAsDuration("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvAsDuration("HTTP_WRITE_TIMEOUT", 60*time.Second),
			IdleTimeout:     getEnvAsDuration("HTTP_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: getEnvAsDuration("HTTP_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Auth: AuthConfig{
			Password: PasswordConfig{
				MinLength: 8,
			},
		},
		Minio: MinioConfig{
			Endpoint:         getEnv("MINIO_ENDPOINT", "localhost:9000"),
			ExternalEndpoint: getEnv("MINIO_EXTERNAL_ENDPOINT", "localhost:9000"),
			AccessKey:        getEnv("MINIO_ROOT_USER", "minioadmin"),
			SecretKey:        getEnv("MINIO_ROOT_PASSWORD", "minioadmin"),
			Bucket:           getEnv("MINIO_BUCKET", "zvideo-videos"),
			UseSSL:           getEnvAsBool("MINIO_USE_SSL", false),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "debug"),
			OutputPath: getEnv("LOG_OUTPUT", "stdout"),
			AddSource:  getEnvAsBool("LOG_ADD_SOURCE", false),
		},
	}
}

// Validate rejects credentials and settings that are unsafe to run in production.
// Development defaults remain available for local development and tests.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if c.Database.Host == "" || c.Database.Port <= 0 || c.Database.User == "" || c.Database.DBName == "" {
		return fmt.Errorf("database connection settings are incomplete")
	}
	if c.Minio.Endpoint == "" || c.Minio.Bucket == "" {
		return fmt.Errorf("minio settings are incomplete")
	}
	if !strings.EqualFold(strings.TrimSpace(c.HTTP.Env), "production") {
		return nil
	}

	if len(c.JWT.Secret) < 32 || isPlaceholder(c.JWT.Secret, "your-secret-key-change-in-production") {
		return fmt.Errorf("JWT_SECRET must be a non-default secret of at least 32 characters in production")
	}
	if c.Database.Password == "" || isPlaceholder(c.Database.Password, "1488") {
		return fmt.Errorf("DB_PASSWORD must be configured in production")
	}
	if isPlaceholder(c.Minio.AccessKey, "minioadmin") || isPlaceholder(c.Minio.SecretKey, "minioadmin") {
		return fmt.Errorf("MinIO credentials must be configured in production")
	}
	if strings.EqualFold(strings.TrimSpace(c.Database.SSLMode), "disable") {
		return fmt.Errorf("DB_SSLMODE must enable transport security in production")
	}
	return nil
}

func isPlaceholder(value string, placeholders ...string) bool {
	value = strings.TrimSpace(value)
	for _, placeholder := range placeholders {
		if strings.EqualFold(value, placeholder) {
			return true
		}
	}
	return false
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
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
