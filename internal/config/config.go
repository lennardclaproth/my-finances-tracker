package config

import (
	"fmt"
	"log/slog"
	"os"

	"go.yaml.in/yaml/v3"
)

const configPath = "config.yaml"

type Config struct {
	Server   Server    `yaml:"server"`
	Database Database  `yaml:"database"`
	Logging  Logging   `yaml:"logging"`
	APM      APMConfig `yaml:"apm"`
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

	os.Setenv("ELASTIC_APM_SERVER_URL", cfg.APM.ServerURL)
	os.Setenv("ELASTIC_APM_SERVICE_NAME", cfg.APM.ServiceName)
	os.Setenv("ELASTIC_APM_ENVIRONMENT", cfg.APM.Environment)
	os.Setenv("ELASTIC_APM_SECRET_TOKEN", cfg.APM.SecretToken)
	os.Setenv("ELASTIC_APM_VERIFY_SERVER_CERT", fmt.Sprintf("%t", cfg.APM.VerifyServerCert))
	os.Setenv("ELASTIC_APM_LOG_LEVEL", cfg.APM.LogLevel)

	return &cfg, nil
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
