package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseDriver string
	Database       DatabaseConfig
	Mongo          MongoConfig
	Redis          RedisConfig
	JWT            JWTConfig
	Server         ServerConfig
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
}

type MongoConfig struct {
	URI                    string
	Host                   string
	Port                   int
	User                   string
	Password               string
	Database               string
	AuthSource             string
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
	MaxPoolSize            uint64
	MinPoolSize            uint64
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

type ServerConfig struct {
	Port string
	Env  string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		DatabaseDriver: getEnv("DB_DRIVER", "mongo"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "1488"),
			DBName:   getEnv("DB_NAME", "zvideo"),
		},
		Mongo: MongoConfig{
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
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
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
			Endpoint:         getEnv("MINIO_ENDPOINT", "minio:9000"),
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
