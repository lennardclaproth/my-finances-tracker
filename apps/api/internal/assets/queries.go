package assets

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type Queries struct {
	qs queryStore
}

type queryStore interface {
	ClassesForAccount(ctx context.Context, accID uuid.UUID, includeArchived bool) (map[uuid.UUID]*Class, error)
	Class(ctx context.Context, classID uuid.UUID) (*Class, error)
	ClassBounds(ctx context.Context, classIDs []uuid.UUID) (map[uuid.UUID]*ClassBounds, error)
	Snapshots(ctx context.Context, accID uuid.UUID, from, to *time.Time) ([]*Snapshot, error)
}

type ClassBounds struct {
	ClassID uuid.UUID
	First   *Mutation
	Last    *Mutation
}

type ClassSummary struct {
	ID           uuid.UUID
	Name         string
	Source       ClassSource
	Archived     bool
	CurrentWorth money.Price
	LastChangeAt *time.Time
	GrowthPct    *float64
	UpdatedAt    time.Time
}

// ListClasses returns table rows for account classes.
func (q *Queries) ListClasses(ctx context.Context, accountID uuid.UUID, includeArchived bool) ([]ClassSummary, error) {
	// Get classes
	classes, err := q.qs.ClassesForAccount(ctx, accountID, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("list classes: failed to get classes for account %s: %w", accountID, err)
	}
	// Make list of class ID's, initialize empty and allocate memory
	// for len(classes)
	ids := make([]uuid.UUID, 0, len(classes))
	for id := range classes {
		ids = append(ids, id)
	}
	// Get bounds of classes (first mutation, last mutation)
	bounds, err := q.qs.ClassBounds(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list classes: failed to get bounds: %w", err)
	}
	summaries := make([]ClassSummary, 0, len(ids))
	// Determine growth
	for _, id := range ids {
		class, ok := classes[id]
		if !ok {
			return nil, fmt.Errorf("list classes: class not in map")
		}
		summary := ClassSummary{
			ID:       class.ID,
			Name:     class.Name,
			Archived: class.Archived,
		}
		bound, ok := bounds[id]
		// if bound.First is nil we can assume that there have been
		// no mutations. When bound.First is not null then we assume bound.Last
		// is always filled, if a mutation is done and it is the first mutation then
		// it is automatically also the last mutation.
		if !ok {
			summaries = append(summaries, summary)
			continue
		}
		// This might seem redundant but it clarifies the code
		if bound.First == nil {
			summaries = append(summaries, summary)
			continue
		}
		growth := growthPctFromBouds(*bound)
		summary.GrowthPct = &growth
		if bound.Last != nil {
			summary.CurrentWorth = bound.Last.ClassTotalWorth
			summary.LastChangeAt = &bound.Last.EffectiveDate
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

type AssetSummary struct {
	ID           uuid.UUID
	Name         string
	CurrentWorth money.Price
	Archived     bool
	UpdatedAt    time.Time
}

type GrowthPoint struct {
	Date       time.Time
	TotalWorth money.Price
}

type ClassDetails struct {
	Class     ClassSummary
	Assets    []AssetSummary
	Growth    []GrowthPoint
	Mutations []Mutation
}

// ClassDetails returns class items, growth points, and history.
func (q *Queries) ClassDetails(ctx context.Context, classID, accID uuid.UUID) (*ClassDetails, error) {
	// Get class including assets and mutations (mutations sorted DESC)
	class, err := q.qs.Class(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("class details: failed to get class: %w", err)
	}
	if class == nil {
		return nil, fmt.Errorf("class details: %w", ErrClassNotFound)
	}
	// Determine growth, because we assume that
	// mutations are already sorted DESC the last mutation
	// is the first entry on the list, the first mutation is
	// the last in the list.
	var lastChangeAt *time.Time
	bound := &ClassBounds{
		ClassID: classID,
		Last:    &class.Mutations[0],
		First:   &class.Mutations[len(class.Mutations)-1],
	}
	lastChangeAt = &bound.Last.EffectiveDate
	growthPct := growthPctFromBouds(*bound)
	// Build growth points for visualization
	growth := growthPointsFromMutations(class.Mutations)
	// Build asset summaries
	assets := make([]AssetSummary, 0, len(class.Assets))
	for _, asset := range class.Assets {
		assets = append(assets, AssetSummary{
			ID:           asset.ID,
			Name:         asset.Name,
			CurrentWorth: asset.CurrentWorth,
			Archived:     asset.Archived,
			UpdatedAt:    asset.UpdatedAt,
		})
	}
	return &ClassDetails{
		Class: ClassSummary{
			ID:           class.ID,
			Name:         class.Name,
			Source:       class.Source,
			Archived:     class.Archived,
			CurrentWorth: class.Mutations[0].ClassTotalWorth,
			LastChangeAt: lastChangeAt,
			GrowthPct:    &growthPct,
			UpdatedAt:    class.UpdatedAt,
		},
		Assets:    assets,
		Growth:    growth,
		Mutations: class.Mutations,
	}, nil
}

// ListSnapshots returns account-level daily total snapshots.
func (q *Queries) ListSnapshots(ctx context.Context, accID uuid.UUID, from, to *time.Time) ([]GrowthPoint, error) {
	snapshots, err := q.qs.Snapshots(ctx, accID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: failed to get snapshots: %w", err)
	}
	out := make([]GrowthPoint, 0, len(snapshots))
	for _, row := range snapshots {
		if row == nil {
			continue
		}
		out = append(out, GrowthPoint{
			Date:       row.OccurredAt,
			TotalWorth: row.TotalWorth,
		})
	}
	return out, nil
}
