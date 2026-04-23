package cashflow

import (
	"context"

	"github.com/google/uuid"
)

// AutoTagFallbackTag is used when the agent returns a deterministic client-side failure.
const AutoTagFallbackTag = "unk"

// AutoTagOutcome describes how auto-tag processing completed for one transaction.
type AutoTagOutcome string

const (
	// AutoTagOutcomeTagged indicates the agent tagged the transaction successfully.
	AutoTagOutcomeTagged AutoTagOutcome = "tagged"
	// AutoTagOutcomeFallbackTagged indicates the fallback tag was applied after client-side agent failure.
	AutoTagOutcomeFallbackTagged AutoTagOutcome = "fallback_tagged"
	// AutoTagOutcomeRetry indicates no tag was applied and the caller should retry later.
	AutoTagOutcomeRetry AutoTagOutcome = "retry"
)

// AutoTagRunner calls the external agent to tag transactions.
type AutoTagRunner interface {
	RunTagAgent(ctx context.Context, tx *Transaction) error
}

// AutoTagWriter persists transaction tags.
type AutoTagWriter interface {
	Tag(ctx context.Context, id uuid.UUID, tag string) error
}

// AutoTagProcessor encapsulates business policy for agent failures and fallback tags.
type AutoTagProcessor struct {
	runner        AutoTagRunner
	writer        AutoTagWriter
	isClientError func(error) bool
}

// NewAutoTagProcessor constructs an auto-tag processor.
func NewAutoTagProcessor(runner AutoTagRunner, writer AutoTagWriter, isClientError func(error) bool) *AutoTagProcessor {
	return &AutoTagProcessor{
		runner:        runner,
		writer:        writer,
		isClientError: isClientError,
	}
}

// Process executes one auto-tag attempt and applies fallback policy when appropriate.
func (p *AutoTagProcessor) Process(ctx context.Context, tx *Transaction) (AutoTagOutcome, error) {
	err := p.runner.RunTagAgent(ctx, tx)
	if err == nil {
		return AutoTagOutcomeTagged, nil
	}
	if p.isClientError == nil || !p.isClientError(err) {
		return AutoTagOutcomeRetry, err
	}
	if tagErr := p.writer.Tag(ctx, tx.ID, AutoTagFallbackTag); tagErr != nil {
		return AutoTagOutcomeRetry, tagErr
	}
	return AutoTagOutcomeFallbackTagged, nil
}
