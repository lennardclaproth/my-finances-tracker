package assets

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
)

// BuilderStore reads mutations and rewrites account-level snapshots.
type BuilderStore interface {
	Mutations(ctx context.Context, accID uuid.UUID, sort *sorting.Direction, skip, take *uint64) ([]*Mutation, error)
	DeleteSnapshots(ctx context.Context, accID uuid.UUID) error
	StoreSnapshots(ctx context.Context, snapshots []*Snapshot) error
}

type Builder struct {
	bs  BuilderStore
	uow UnitOfWork
}

// NewBuilder constructs the assets snapshot Builder.
func NewBuilder(bs BuilderStore, uow UnitOfWork) *Builder {
	return &Builder{bs: bs, uow: uow}
}

// RebuildAll rebuilds account-level assets snapshots from account mutations.
func (b *Builder) RebuildAll(ctx context.Context, accountID uuid.UUID) error {
	sort := sorting.ASC
	mutations, err := b.bs.Mutations(ctx, accountID, &sort, nil, nil)
	if err != nil {
		return fmt.Errorf("assets service: list account mutations for snapshots: %w", err)
	}

	today := date.StartOfDayUTC(time.Now().UTC())
	rows := buildSnapshotsFromMutations(accountID, mutations, today)
	if len(rows) == 0 {
		// If there are no rows we assume no mutations have
		// been done technically speaking this should never occur
		// considering a new class always creates a mutation with
		// an inital worth.
		rows = []*Snapshot{
			{
				ID:         uuid.New(),
				AccountID:  accountID,
				OccurredAt: today,
				TotalWorth: 0,
				CreatedAt:  time.Now().UTC(),
				UpdatedAt:  time.Now().UTC(),
			},
		}
	}

	if err := b.uow.Do(ctx, func(txCtx context.Context) error {
		if deleteErr := b.bs.DeleteSnapshots(txCtx, accountID); deleteErr != nil {
			return deleteErr
		}
		return b.bs.StoreSnapshots(txCtx, rows)
	}); err != nil {
		return fmt.Errorf("rebuild all: failed to execute transaction: %w", err)
	}
	return nil
}

func buildSnapshotsFromMutations(accountID uuid.UUID, mutations []*Mutation, today time.Time) []*Snapshot {
	if len(mutations) == 0 {
		return nil
	}
	type classKey = uuid.UUID
	type dayKey = string

	latestByClassAndDay := make(map[classKey]map[dayKey]money.Price)
	earliestDay := date.StartOfDayUTC(today)
	earliestSet := false

	for _, row := range mutations {
		if row == nil {
			continue
		}
		day := date.StartOfDayUTC(row.EffectiveDate)
		key := day.Format("2006-01-02")
		perClass := latestByClassAndDay[row.ClassID]
		if perClass == nil {
			perClass = make(map[dayKey]money.Price)
			latestByClassAndDay[row.ClassID] = perClass
		}
		// Mutations is ordered ascending, so the latest row for each class/day wins.
		perClass[key] = row.ClassTotalWorth
		if !earliestSet || day.Before(earliestDay) {
			earliestDay = day
			earliestSet = true
		}
	}
	if !earliestSet {
		return nil
	}

	lastByClass := make(map[classKey]money.Price)
	out := make([]*Snapshot, 0)
	now := time.Now().UTC()

	for day := earliestDay; !day.After(today); day = day.AddDate(0, 0, 1) {
		dayToken := day.Format("2006-01-02")
		total := money.Price(0)
		for classID, series := range latestByClassAndDay {
			if nextWorth, ok := series[dayToken]; ok {
				lastByClass[classID] = nextWorth
			}
			total += lastByClass[classID]
		}
		out = append(out, &Snapshot{
			ID:         uuid.New(),
			AccountID:  accountID,
			OccurredAt: day,
			TotalWorth: total,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	return out
}
