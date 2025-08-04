package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	Weather     WeatherConfig
	Redis       RedisConfig
	Monitoring  MonitoringConfig
	Stripe      StripeConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
	MaxConns int
	MinConns int
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
	Issuer     string
}

type WeatherConfig struct {
	OpenWeatherMapAPIKey string
	WeatherAPIKey        string
	DefaultProvider      string
	CacheTTL            time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type MonitoringConfig struct {
	MetricsEnabled bool
	TracingEnabled bool
	LogLevel       string
}

type StripeConfig struct {
	APIKey          string
	WebhookSecret   string
	PublishableKey  string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", "30s"),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", "30s"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Name:     getEnv("DB_NAME", "smart_garden_bot"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvAsInt("DB_MAX_CONNS", 25),
			MinConns: getEnvAsInt("DB_MIN_CONNS", 5),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			Expiration: getEnvAsDuration("JWT_EXPIRATION", "24h"),
			Issuer:     getEnv("JWT_ISSUER", "smart-garden-bot"),
		},
		Weather: WeatherConfig{
			OpenWeatherMapAPIKey: getEnv("OPENWEATHERMAP_API_KEY", ""),
			WeatherAPIKey:        getEnv("WEATHER_API_KEY", ""),
			DefaultProvider:      getEnv("DEFAULT_WEATHER_PROVIDER", "openweathermap"),
			CacheTTL:            getEnvAsDuration("WEATHER_CACHE_TTL", "30m"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Monitoring: MonitoringConfig{
			MetricsEnabled: getEnvAsBool("METRICS_ENABLED", true),
			TracingEnabled: getEnvAsBool("TRACING_ENABLED", false),
			LogLevel:       getEnv("LOG_LEVEL", "info"),
		},
		Stripe: StripeConfig{
			APIKey:          getEnv("STRIPE_API_KEY", ""),
			WebhookSecret:   getEnv("STRIPE_WEBHOOK_SECRET", ""),
			PublishableKey:  getEnv("STRIPE_PUBLISHABLE_KEY", ""),
		},
	}

	return cfg, nil
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}