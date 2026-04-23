package config

import (
	"testing"

	"github.com/google/uuid"
)

func TestApplyAPMDefaults_ProductionDefault(t *testing.T) {
	t.Setenv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE", "")

	cfg := &Config{
		Server: Server{Environment: "production"},
		APM:    APMConfig{TransactionSampleRate: 0},
	}
	cfg.applyAPMDefaults()
	if cfg.APM.TransactionSampleRate != 0.2 {
		t.Fatalf("expected production sample rate 0.2, got %v", cfg.APM.TransactionSampleRate)
	}
}

func TestApplyAPMDefaults_DevDefault(t *testing.T) {
	t.Setenv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE", "")

	cfg := &Config{
		Server: Server{Environment: "development"},
		APM:    APMConfig{TransactionSampleRate: 0},
	}
	cfg.applyAPMDefaults()
	if cfg.APM.TransactionSampleRate != 1.0 {
		t.Fatalf("expected development sample rate 1.0, got %v", cfg.APM.TransactionSampleRate)
	}
}

func TestApplyAPMDefaults_EnvOverride(t *testing.T) {
	t.Setenv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE", "0.42")

	cfg := &Config{
		Server: Server{Environment: "production"},
		APM:    APMConfig{TransactionSampleRate: 0},
	}
	cfg.applyAPMDefaults()
	if cfg.APM.TransactionSampleRate != 0.42 {
		t.Fatalf("expected env override sample rate 0.42, got %v", cfg.APM.TransactionSampleRate)
	}
}

func TestApplyAPMDefaults_ExistingConfigValueWins(t *testing.T) {
	t.Setenv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE", "")

	cfg := &Config{
		Server: Server{Environment: "production"},
		APM:    APMConfig{TransactionSampleRate: 0.75},
	}
	cfg.applyAPMDefaults()
	if cfg.APM.TransactionSampleRate != 0.75 {
		t.Fatalf("expected existing config value 0.75, got %v", cfg.APM.TransactionSampleRate)
	}
}

func validConfigForValidation() Config {
	return Config{
		Server: Server{
			Environment: "development",
			Port:        6060,
		},
		Database: Database{
			ConnStr: "sqlite://tmp.db",
			Type:    "sqlite3",
		},
		Logging: Logging{
			Level: "info",
		},
		APM: APMConfig{
			ServerURL:             "http://localhost:8200",
			ServiceName:           "my-finances-tracker",
			Environment:           "development",
			LogLevel:              "info",
			TransactionSampleRate: 1,
		},
		DiskStorage: DiskStorage{
			BasePath: "C:\\mft",
		},
		Agent: AgentConfig{
			Enabled:           true,
			AgentBaseURL:      "http://localhost:8001/api",
			DefaultTagAgentID: uuid.NewString(),
		},
	}
}

func TestValidate_AllowsDisabledAgentWithoutConnectionSettings(t *testing.T) {
	cfg := validConfigForValidation()
	cfg.Agent.Enabled = false
	cfg.Agent.AgentBaseURL = ""
	cfg.Agent.DefaultTagAgentID = ""

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected disabled agent validation to pass, got %v", err)
	}
}

func TestValidate_RejectsEnabledAgentWithInvalidDefaultTagAgentID(t *testing.T) {
	cfg := validConfigForValidation()
	cfg.Agent.Enabled = true
	cfg.Agent.DefaultTagAgentID = "not-a-uuid"

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid default tag agent id")
	}
}
