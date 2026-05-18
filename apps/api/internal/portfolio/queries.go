package portfolio

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
)

type Queries struct {
	qs queryStore
}

type queryStore interface {
	SnapshotsForAccount(ctx context.Context, accountID uuid.UUID, limit, offset *int, from, to *time.Time, sort *sorting.Direction) ([]*PortfolioSnapshot, error)
}

type SnapshotSortField string

// const (
// 	SortByDate        SnapshotSortField = "date"
// )

// type SnapshotSort struct {
// 	Field     sorting.Field
// 	Direction sorting.Direction
// }

// snapshotsForAccount returns a list of portfolio snapshots for the given account, ordered by date descending.
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
