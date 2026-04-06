package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Server   ServerConfig
	HTTP     HTTPConfig
	Auth     AuthConfig
}

type HTTPConfig struct {
	Port            string
	Env             string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	ExposedHeaders []string

	RateLimit       int
	RateLimitWindow time.Duration

	MaxBodySize       int64
	EnableCompression bool
	EnableSwagger     bool
	ServeStatic       bool
	StaticDir         string
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
	err := godotenv.Load()
	if err != nil {
		return nil
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "1488"),
			DBName:   getEnv("DB_NAME", "zvideo"),
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

			AllowedOrigins: getEnvAsSlice("HTTP_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),
			AllowedMethods: getEnvAsSlice("HTTP_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}),
			AllowedHeaders: getEnvAsSlice("HTTP_ALLOWED_HEADERS", []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"}),
			ExposedHeaders: getEnvAsSlice("HTTP_EXPOSED_HEADERS", []string{"Link", "X-Request-ID"}),

			RateLimit:       getEnvAsInt("HTTP_RATE_LIMIT", 100),
			RateLimitWindow: getEnvAsDuration("HTTP_RATE_LIMIT_WINDOW", 1*time.Minute),

			MaxBodySize:       getEnvAsInt64("HTTP_MAX_BODY_SIZE", 10<<20),
			EnableCompression: getEnvAsBool("HTTP_ENABLE_COMPRESSION", true),
			EnableSwagger:     getEnvAsBool("HTTP_ENABLE_SWAGGER", false),
			ServeStatic:       getEnvAsBool("HTTP_SERVE_STATIC", true),
			StaticDir:         getEnv("HTTP_STATIC_DIR", "./uploads"),
		},
		Auth: AuthConfig{
			Password: PasswordConfig{
				MinLength: 8,
			},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
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

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		var parts []string
		for _, part := range splitByComma(value) {
			if part != "" {
				parts = append(parts, part)
			}
		}
		if len(parts) > 0 {
			return parts
		}
	}
	return defaultValue
}

func splitByComma(s string) []string {
	var result []string
	current := ""
	for _, ch := range s {
		if ch == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
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
