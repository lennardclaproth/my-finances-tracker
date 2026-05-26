package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
)

type Queries struct {
	qs queryStore
}

type queryStore interface {
	SnapshotsForAccount(ctx context.Context, accountID uuid.UUID, limit, offset *int, from, to *time.Time, sort *sorting.Direction) ([]*PortfolioSnapshot, error)
	PositionsWithLatestSnapshot(ctx context.Context, accID uuid.UUID, includeClosed bool) ([]*PositionWithLatestSnapshot, error)
}

// SnapshotsForAccount returns a list of portfolio snapshots for the given account, ordered by date descending.
// When limit, offset, from and to are nil, it returns all snapshots for the account.
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
	if q.qs == nil {
		return nil, fmt.Errorf("portfolio positions: query store is not configured")
	}

	positions, err := q.qs.PositionsWithLatestSnapshot(ctx, accountID, includeClosed)
	if err != nil {
		return nil, fmt.Errorf("portfolio positions: %w", err)
	}
	if positions == nil {
		return []*PositionWithLatestSnapshot{}, nil
	}
	return positions, nil
}
