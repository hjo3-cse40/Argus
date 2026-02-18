package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvDev   Environment = "dev"
	EnvStage Environment = "stage"
	EnvProd  Environment = "prod"
)

type Config struct {
	Environment Environment
	Port        string
	RabbitMQ    RabbitMQConfig
	API         APIConfig
	Database    DatabaseConfig
	Destinations DestinationsConfig
}

type RabbitMQConfig struct {
	URL      string
	Username string
	Password string
	Host     string
	Port     string
}

type APIConfig struct {
	BaseURL string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type DestinationsConfig struct {
	DiscordWebhookURL string
}

func Load() (*Config, error) {
	// Load .env from current dir or infra/ so DISCORD_WEBHOOK_URL etc. are set
	for _, path := range []string{".env", "infra/.env", "../infra/.env"} {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			break
		}
	}
	// Also try next to the binary (e.g. when run from backend/cmd/worker)
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for _, name := range []string{"infra/.env", "../infra/.env", ".env"} {
			p := filepath.Join(dir, name)
			if _, err := os.Stat(p); err == nil {
				_ = godotenv.Load(p)
				break
			}
		}
	}

	env := getEnv("ENV", string(EnvDev))
	environment := Environment(env)
	if environment != EnvDev && environment != EnvStage && environment != EnvProd {
		return nil, fmt.Errorf("invalid ENV value: %s (must be dev, stage, or prod)", env)
	}

	cfg := &Config{
		Environment: environment,
		Port:        getEnv("PORT", "8080"),
		RabbitMQ: RabbitMQConfig{
			URL:      getEnv("RABBITMQ_URL", ""),
			Username: getEnv("RABBITMQ_USER", "argus"),
			Password: getEnv("RABBITMQ_PASS", "argus"),
			Host:     getEnv("RABBITMQ_HOST", "localhost"),
			Port:     getEnv("RABBITMQ_PORT", "5672"),
		},
		API: APIConfig{
			BaseURL: getEnv("API_BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "argus"),
			Password: getEnv("DB_PASSWORD", "argus"),
			DBName:   getEnv("DB_NAME", "argus"),
		},
		Destinations: DestinationsConfig{
			DiscordWebhookURL: getEnv("DISCORD_WEBHOOK_URL", ""),
		},
	}

	// Build RabbitMQ URL if not provided
	if cfg.RabbitMQ.URL == "" {
		cfg.RabbitMQ.URL = fmt.Sprintf("amqp://%s:%s@%s:%s/",
			cfg.RabbitMQ.Username,
			cfg.RabbitMQ.Password,
			cfg.RabbitMQ.Host,
			cfg.RabbitMQ.Port,
		)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func (c *Config) IsDev() bool {
	return c.Environment == EnvDev
}

func (c *Config) IsStage() bool {
	return c.Environment == EnvStage
}

func (c *Config) IsProd() bool {
	return c.Environment == EnvProd
}
