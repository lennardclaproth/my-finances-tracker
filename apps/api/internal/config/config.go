package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

const configPath = "config.yaml"

type Config struct {
	Server      Server      `yaml:"server"`
	Database    Database    `yaml:"database"`
	Logging     Logging     `yaml:"logging"`
	APM         APMConfig   `yaml:"apm"`
	DiskStorage DiskStorage `yaml:"disk_storage"`
	Providers   Providers   `yaml:"providers"`
}

type DiskStorage struct {
	BasePath string `yaml:"base_path"`
}

type Logging struct {
	Level string `yaml:"level"`
}

type Server struct {
	Environment string `yaml:"environment"`
	Port        int    `yaml:"port"`
}

type Database struct {
	ConnStr string `yaml:"connection_string"`
	Type    string `yaml:"type"`
}

type APMConfig struct {
	ServerURL             string  `yaml:"server_url"`
	ServiceName           string  `yaml:"service_name"`
	Environment           string  `yaml:"environment"`
	SecretToken           string  `yaml:"secret_token"`
	VerifyServerCert      bool    `yaml:"verify_server_cert"`
	LogLevel              string  `yaml:"log_level"`
	TransactionSampleRate float64 `yaml:"transaction_sample_rate"`
}

type Providers struct {
	MarketStack  ProviderConfig `yaml:"marketstack"`
	AlphaVantage ProviderConfig `yaml:"alphavantage"`
}

type ProviderConfig struct {
	BaseURI string   `yaml:"base_uri"`
	APIKeys []string `yaml:"-"`
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}
	if c.Database.ConnStr == "" {
		return fmt.Errorf("database connection string cannot be empty")
	}
	if c.Database.Type != "sqlite3" && c.Database.Type != "postgres" {
		return fmt.Errorf("unsupported database type: %s", c.Database.Type)
	}
	if c.Logging.Level != "debug" && c.Logging.Level != "info" && c.Logging.Level != "warn" && c.Logging.Level != "error" {
		return fmt.Errorf("invalid logging level: %s", c.Logging.Level)
	}
	if c.APM.ServerURL == "" {
		return fmt.Errorf("APM server URL cannot be empty")
	}
	if c.APM.ServiceName == "" {
		return fmt.Errorf("APM service name cannot be empty")
	}
	if c.APM.LogLevel != "debug" && c.APM.LogLevel != "info" && c.APM.LogLevel != "warn" && c.APM.LogLevel != "error" {
		return fmt.Errorf("invalid APM log level: %s", c.APM.LogLevel)
	}
	if c.APM.TransactionSampleRate < 0 || c.APM.TransactionSampleRate > 1 {
		return fmt.Errorf("APM transaction sample rate must be between 0 and 1")
	}
	if c.DiskStorage.BasePath == "" {
		return fmt.Errorf("disk storage base path cannot be empty")
	}
	return nil
}

func ReadConfig() (*Config, error) {
	f, err := os.ReadFile(configPath)

	if err != nil {
		return nil, fmt.Errorf("config: error opening config file at %s: %w", configPath, err)
	}

	var cfg Config

	if err := yaml.Unmarshal(f, &cfg); err != nil {
		return nil, fmt.Errorf("config: error decoding config: %w", err)
	}

	cfg.hydrateProviderEnv()
	cfg.applyAPMDefaults()

	apmEnv := []struct {
		key   string
		value string
	}{
		{"ELASTIC_APM_SERVER_URL", cfg.APM.ServerURL},
		{"ELASTIC_APM_SERVICE_NAME", cfg.APM.ServiceName},
		{"ELASTIC_APM_ENVIRONMENT", cfg.APM.Environment},
		{"ELASTIC_APM_SECRET_TOKEN", cfg.APM.SecretToken},
		{"ELASTIC_APM_VERIFY_SERVER_CERT", strconv.FormatBool(cfg.APM.VerifyServerCert)},
		{"ELASTIC_APM_LOG_LEVEL", cfg.APM.LogLevel},
		{"ELASTIC_APM_TRANSACTION_SAMPLE_RATE", strconv.FormatFloat(cfg.APM.TransactionSampleRate, 'f', 2, 64)},
	}
	for _, env := range apmEnv {
		if err := os.Setenv(env.key, env.value); err != nil {
			return nil, fmt.Errorf("config: failed setting %s: %w", env.key, err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) hydrateProviderEnv() {
	marketStackBaseURI := strings.TrimSpace(c.Providers.MarketStack.BaseURI)
	if marketStackBaseURI == "" {
		marketStackBaseURI = strings.TrimSpace(os.Getenv("MARKETSTACK_BASE_URI"))
	}
	if marketStackBaseURI == "" {
		marketStackBaseURI = "https://api.marketstack.com/v2"
	}
	c.Providers.MarketStack.BaseURI = marketStackBaseURI
	c.Providers.MarketStack.APIKeys = splitAndDedupeCommaValues(os.Getenv("MARKETSTACK_API_KEY"))

	alphaVantageBaseURI := strings.TrimSpace(c.Providers.AlphaVantage.BaseURI)
	if alphaVantageBaseURI == "" {
		alphaVantageBaseURI = strings.TrimSpace(os.Getenv("ALPHA_VANTAGE_BASE_URI"))
	}
	if alphaVantageBaseURI == "" {
		alphaVantageBaseURI = "https://www.alphavantage.co"
	}
	c.Providers.AlphaVantage.BaseURI = alphaVantageBaseURI
	c.Providers.AlphaVantage.APIKeys = splitAndDedupeCommaValues(os.Getenv("ALPHA_VANTAGE_API_KEY"))
}

func splitAndDedupeCommaValues(raw string) []string {
	parts := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))

	for _, part := range parts {
		v := strings.TrimSpace(part)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	return out
}

func (c *Config) applyAPMDefaults() {
	if raw := strings.TrimSpace(os.Getenv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE")); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
			c.APM.TransactionSampleRate = parsed
			return
		}
	}

	if c.APM.TransactionSampleRate > 0 {
		return
	}

	environment := strings.ToLower(strings.TrimSpace(c.Server.Environment))
	if environment == "" {
		environment = strings.ToLower(strings.TrimSpace(c.APM.Environment))
	}

	switch environment {
	case "prod", "production":
		c.APM.TransactionSampleRate = 0.2
	default:
		c.APM.TransactionSampleRate = 1.0
	}
}

func (l *Logging) GetLogLevel() slog.Leveler {
	switch l.Level {
	case "debug":
		return slog.LevelDebug // Debug
	case "info":
		return slog.LevelInfo // Info
	case "warn":
		return slog.LevelWarn // Warn
	case "error":
		return slog.LevelError // Error
	default:
		return slog.LevelInfo // Default to Info
	}
}
