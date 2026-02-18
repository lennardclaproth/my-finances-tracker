package jobs

import (
	"context"
	"sync"

	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

type Job interface {
	Name() string
	Start(ctx context.Context) error
}

type Manager struct {
	jobs []Job
	wg   sync.WaitGroup
	log  logging.Logger
}

func NewManager(log logging.Logger, jobs ...Job) *Manager {
	return &Manager{
		jobs: jobs,
		log:  log,
	}
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
				m.log.Error(ctx, "Job %s exited with error: %v", err, j.Name())
			}
		}(job)
	}

	<-ctx.Done() // Wait for cancellation signal

	m.wg.Wait()
	return nil
}
