package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/config"
)

type noopSetupJobsLogger struct{}

func (noopSetupJobsLogger) Debug(context.Context, string, ...any) {}
func (noopSetupJobsLogger) Info(context.Context, string, ...any)  {}
func (noopSetupJobsLogger) Warn(context.Context, string, ...any)  {}
func (noopSetupJobsLogger) Error(context.Context, string, error, ...any) {
}

func containsJobName(names []string, target string) bool {
	for _, name := range names {
		if name == target {
			return true
		}
	}
	return false
}

func TestSetupJobs_AgentEnabled_IncludesTaggerJob(t *testing.T) {
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Enabled:           true,
			AgentBaseURL:      "http://localhost:8001/api",
			DefaultTagAgentID: uuid.NewString(),
		},
	}

	mgr, _, _, _ := setupJobs(noopSetupJobsLogger{}, &appDependencies{}, cfg, nil)
	names := mgr.JobNames()
	if !containsJobName(names, "TaggerJob") {
		t.Fatalf("expected TaggerJob when agent is enabled, jobs=%v", names)
	}
}

func TestSetupJobs_AgentDisabled_ExcludesTaggerJob(t *testing.T) {
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Enabled:           false,
			AgentBaseURL:      "",
			DefaultTagAgentID: "",
		},
	}

	mgr, _, _, _ := setupJobs(noopSetupJobsLogger{}, &appDependencies{}, cfg, nil)
	names := mgr.JobNames()
	if containsJobName(names, "TaggerJob") {
		t.Fatalf("did not expect TaggerJob when agent is disabled, jobs=%v", names)
	}
}
