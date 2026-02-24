package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvDev   Environment = "dev"
	EnvStage Environment = "stage"
	EnvProd  Environment = "prod"
)

type Config struct {
	Environment  Environment
	Port         string
	RabbitMQ     RabbitMQConfig
	API          APIConfig
	Database     DatabaseConfig
	Destinations DestinationsConfig
	RSSHub       RSSHubConfig
}

type RSSHubConfig struct {
	BaseURL string
	Feeds   []Feed
}

type Feed struct {
	SourceType string // e.g. "youtube", "reddit", "x", "github"
	URL        string // full RSSHub URL
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
		RSSHub: RSSHubConfig{
			BaseURL: getEnv("RSSHUB_BASE_URL", "http://localhost:1200"),
		},
	}

	// Parse comma-separated feeds in "type:path" format
	// e.g. "youtube:youtube/channel/ABC,reddit:reddit/subreddit/golang"
	if feeds := getEnv("RSSHUB_FEEDS", ""); feeds != "" {
		for _, f := range strings.Split(feeds, ",") {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}
			parts := strings.SplitN(f, ":", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				continue
			}
			cfg.RSSHub.Feeds = append(cfg.RSSHub.Feeds, Feed{
				SourceType: parts[0],
				URL:        cfg.RSSHub.BaseURL + "/" + parts[1],
			})
		}
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

	// Validate DISCORD_WEBHOOK_URL format when set (same rule as sources API)
	if cfg.Destinations.DiscordWebhookURL != "" &&
		!strings.HasPrefix(cfg.Destinations.DiscordWebhookURL, "https://discord.com/api/webhooks/") {
		return nil, fmt.Errorf("DISCORD_WEBHOOK_URL must be a valid Discord webhook URL (https://discord.com/api/webhooks/...)")
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
