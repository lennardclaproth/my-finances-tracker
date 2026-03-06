package config

import "testing"

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
