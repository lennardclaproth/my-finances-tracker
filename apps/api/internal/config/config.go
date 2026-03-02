package config

import (
	"fmt"
	"log/slog"
	"os"
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
	Agent       AgentConfig `yaml:"agent"`
	Providers   Providers   `yaml:"providers"`
}

type AgentConfig struct {
	AgentBaseURL      string `yaml:"agent_base_url"`
	DefaultTagAgentID string `yaml:"default_tag_agent_id"`
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
	if c.Agent.AgentBaseURL == "" {
		return fmt.Errorf("agent base URL cannot be empty")
	}
	if c.Agent.DefaultTagAgentID == "" {
		return fmt.Errorf("default tag agent ID cannot be empty")
	}
	return nil
}

func ReadConfig() (*Config, error) {
	f, err := os.ReadFile(configPath)

	if err != nil {
		return nil, fmt.Errorf("config: error opening config file at %s: %w", configPath, err)
	}

	var cfg Config

	err = yaml.Unmarshal(f, &cfg)

	if err != nil {
		panic(fmt.Errorf("config: error decoding config: %w", err))
	}

	cfg.hydrateProviderEnv()

	os.Setenv("ELASTIC_APM_SERVER_URL", cfg.APM.ServerURL)
	os.Setenv("ELASTIC_APM_SERVICE_NAME", cfg.APM.ServiceName)
	os.Setenv("ELASTIC_APM_ENVIRONMENT", cfg.APM.Environment)
	os.Setenv("ELASTIC_APM_SECRET_TOKEN", cfg.APM.SecretToken)
	os.Setenv("ELASTIC_APM_VERIFY_SERVER_CERT", fmt.Sprintf("%t", cfg.APM.VerifyServerCert))
	os.Setenv("ELASTIC_APM_LOG_LEVEL", cfg.APM.LogLevel)

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
