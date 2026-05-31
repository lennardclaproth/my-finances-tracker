package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
)

// Queries exposes portfolio read-side use cases.
type Queries struct {
	qs queryStore
}

type queryStore interface {
	SnapshotsForAccount(ctx context.Context, accountID uuid.UUID, limit, offset *int, from, to *time.Time, sort *sorting.Direction) ([]*PortfolioSnapshot, error)
	PositionsWithLatestSnapshot(ctx context.Context, accID uuid.UUID, includeClosed bool) ([]*PositionWithLatestSnapshot, error)
}

// NewQueries creates portfolio read-side use cases.
func NewQueries(qs queryStore) *Queries {
	return &Queries{
		qs: qs,
	}
}

// SnapshotsForAccount returns portfolio snapshots for the given account.
// When limit, offset, from and to are nil, it returns all snapshots using the store's default ordering.
func (q *Queries) SnapshotsForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	limit, offset *int,
	from, to *time.Time,
	sort *sorting.Direction,
) ([]*PortfolioSnapshot, error) {
	return q.qs.SnapshotsForAccount(ctx, accountID, limit, offset, from, to, sort)
}

// PositionsForAccount returns account positions with their latest snapshot metadata.
func (q *Queries) PositionsForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	includeClosed bool,
) ([]*PositionWithLatestSnapshot, error) {
	positions, err := q.qs.PositionsWithLatestSnapshot(ctx, accountID, includeClosed)
	if err != nil {
		return nil, fmt.Errorf("portfolio positions: %w", err)
	}
	if positions == nil {
		return []*PositionWithLatestSnapshot{}, nil
	}
	return positions, nil
}
