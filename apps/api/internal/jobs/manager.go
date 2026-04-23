package jobs

import (
	"context"
	"sync"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

// Job represents a background worker lifecycle.
type Job interface {
	Name() string
	Start(ctx context.Context) error
}

// Manager coordinates startup and shutdown of background jobs.
type Manager struct {
	jobs []Job
	wg   sync.WaitGroup
	log  logging.Logger
}

// NewManager constructs a job manager for the provided jobs.
func NewManager(log logging.Logger, jobs ...Job) *Manager {
	return &Manager{
		jobs: jobs,
		log:  log,
	}
}

// JobNames returns the managed job names in registration order.
func (m *Manager) JobNames() []string {
	names := make([]string, 0, len(m.jobs))
	for _, job := range m.jobs {
		if job == nil {
			continue
		}
		names = append(names, job.Name())
	}
	return names
}

// Start runs all jobs managed by the Manager.
func (m *Manager) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, job := range m.jobs {
		m.wg.Add(1)
		go func(j Job) {
			defer m.wg.Done()
			if err := j.Start(ctx); err != nil {
				m.log.Error(ctx, "job exited with error", err, "operation", "job."+j.Name())
			}
		}(job)
	}

	<-ctx.Done() // Wait for cancellation signal

	m.wg.Wait()
	return nil
}
