package portfolio

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

// recordingLocker refuses the build lock so Builder.Build returns
// ErrBuildInProgress before touching its other dependencies, letting the test
// observe whether (and for which account) a rebuild was attempted.
type recordingLocker struct {
	acquireCalls int
	lastAccID    uuid.UUID
}

func (l *recordingLocker) TryAcquireBuildLock(_ context.Context, accID uuid.UUID) (bool, error) {
	l.acquireCalls++
	l.lastAccID = accID
	return false, nil
}
func (l *recordingLocker) ReleaseBuildLock(_ context.Context, _ uuid.UUID) error { return nil }

func TestImportCompletedTriggersRebuildForPortfolioImports(t *testing.T) {
	locker := &recordingLocker{}
	builder := portfolio.NewBuilder(nil, nil, nil, nil, locker, nil)
	handler := NewImportCompletedHandler(builder, nil)

	accID := uuid.New()
	err := handler.Handle(context.Background(), importer.Completed{
		ImportID:  uuid.New(),
		Type:      importer.ImportTypePortfolio,
		AccountID: &accID,
	}, eventbus.Metadata{})

	if !errors.Is(err, portfolio.ErrBuildInProgress) {
		t.Fatalf("expected a rebuild attempt (ErrBuildInProgress), got %v", err)
	}
	if locker.acquireCalls != 1 || locker.lastAccID != accID {
		t.Fatalf("expected one build-lock acquire for %s, got calls=%d id=%s", accID, locker.acquireCalls, locker.lastAccID)
	}
}

func TestImportCompletedIgnoresNonPortfolioImports(t *testing.T) {
	locker := &recordingLocker{}
	builder := portfolio.NewBuilder(nil, nil, nil, nil, locker, nil)
	handler := NewImportCompletedHandler(builder, nil)

	accID := uuid.New()
	err := handler.Handle(context.Background(), importer.Completed{
		ImportID:  uuid.New(),
		Type:      importer.ImportTypeCashflow,
		AccountID: &accID,
	}, eventbus.Metadata{})

	if err != nil {
		t.Fatalf("expected nil for non-portfolio import, got %v", err)
	}
	if locker.acquireCalls != 0 {
		t.Fatalf("expected no rebuild for a cashflow import, got %d acquire calls", locker.acquireCalls)
	}
}
