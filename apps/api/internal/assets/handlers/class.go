package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
)

type ClassHandler struct {
	aec accountExistenceChecker
	cs  classStorer
	cg  classGetter
}

type classStorer interface {
	Create(ctx context.Context, class *assets.Class) error
	Update(ctx context.Context, class *assets.Class) error
}

func NewClassHandler() *ClassHandler {
	return &ClassHandler{}
}

// CreateClass creates a manual class for an account.
func (h *ClassHandler) CreateClass(ctx context.Context, accID uuid.UUID, name string) (*Class, error) {
	// Guard against account existence
	exists, err := h.aec.Exists(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("handlers: failed to create class: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("handlers: failed to create class: %w", ErrAccountNotFound)
	}
	// Create new class
	class, err := assets.NewClass(accID, nil, name)
	if err != nil {
		return fmt.Errorf("handlers: create class failed: %w", err), nil
	}
	// Store class
	err = h.cs.Create(ctx, class)
	if err != nil {
		return nil, fmt.Errorf("handlers: failed to create class: %w", err)
	}
	// TODO: publish rebuild event
	return class, nil
}

// UpdateClass mutates a manual class name/archive status.
func (h *ClassHandler) UpdateClass(ctx context.Context, accID, classID uuid.UUID, name *string, archived *bool) error {
	// if err := s.EnsureAccountProjection(ctx, input.AccountID); err != nil {
	// 	return err
	// }
	exists, err := h.aec.Exists(ctx, accID)
	if err != nil {
		return fmt.Errorf("handlers: failed to update class: %w", err)
	}
	if !exists {
		return fmt.Errorf("handlers: failed to update class: %w", ErrAccountNotFound)
	}
	class, err := h.cg.Get(ctx, accID, classID)
	if err != nil {
		return fmt.Errorf("handlers: failed to update class: %w", err)
	}
	if class == nil {
		return fmt.Errorf("handlers: failed to update class: %w", ErrClassNotFound)
	}
	if class.Source != assets.ClassSourceManual {
		return fmt.Errorf("handlers: failed to update class: %w", ErrClassSourceNotManual)
	}
	
	// var name *string
	// if input.Name != nil {
	// 	normalized, normErr := normalizeClassName(*input.Name)
	// 	if normErr != nil {
	// 		return normErr
	// 	}
	// 	name = &normalized
	// }
	err = class.Update(name, archived)
	if err != nil {
		return fmt.Errorf("handlers: failed to update class: %w", err)
	}
	err = h.cs.Update(ctx, class)
	if err != nil {
		return fmt.Errorf("handlers: failed to update class: %w", err)
	}
	
	//TODO: publish rebuild
	return nil
}

// DeleteClass removes a manual class and related items/history.
func (s *Service) DeleteClass(ctx context.Context, accountID, classID uuid.UUID) error {
	if err := s.EnsureAccountProjection(ctx, accountID); err != nil {
		return err
	}
	class, err := s.store.FetchClassByID(ctx, accountID, classID)
	if err != nil {
		return fmt.Errorf("assets service: fetch class: %w", err)
	}
	if class == nil {
		return ErrAssetClassNotFound
	}
	if class.Source != ClassSourceManual {
		return ErrAssetClassNotManual
	}
	deleted, err := s.store.DeleteClass(ctx, accountID, classID)
	if err != nil {
		return fmt.Errorf("assets service: delete class: %w", err)
	}
	if deleted == 0 {
		return ErrAssetClassNotFound
	}
	s.requestSnapshotsRebuild(ctx, accountID)
	return nil
}

// ListClasses returns table rows for account classes.
func (s *Service) ListClasses(ctx context.Context, accountID uuid.UUID, includeArchived bool) ([]ClassSummary, error) {
	if err := s.EnsureAccountProjection(ctx, accountID); err != nil {
		return nil, err
	}
	classes, err := s.store.ListClassesForAccount(ctx, accountID, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("assets service: list classes: %w", err)
	}

	out := make([]ClassSummary, 0, len(classes))
	for _, class := range classes {
		if class == nil {
			continue
		}
		currentWorth, err := s.store.SumClassWorth(ctx, accountID, class.ID)
		if err != nil {
			return nil, fmt.Errorf("assets service: sum class worth: %w", err)
		}
		latestHistory, err := s.store.ListHistoryByClass(ctx, accountID, class.ID, 1, false)
		if err != nil {
			return nil, fmt.Errorf("assets service: list class history: %w", err)
		}
		earliestHistory, err := s.store.ListHistoryByClass(ctx, accountID, class.ID, 1, true)
		if err != nil {
			return nil, fmt.Errorf("assets service: list class inception history: %w", err)
		}

		var lastChangeAt *time.Time
		var growthPct *float64
		if len(latestHistory) > 0 {
			last := latestHistory[0].EffectiveDate
			lastChangeAt = &last
		}
		if len(latestHistory) > 0 && len(earliestHistory) > 0 {
			growthPct = growthPctFromInception(earliestHistory[0].ClassTotalWorth, latestHistory[0].ClassTotalWorth)
		}

		out = append(out, ClassSummary{
			ID:           class.ID,
			Name:         class.Name,
			Source:       class.Source,
			Archived:     class.Archived,
			CurrentWorth: currentWorth,
			LastChangeAt: lastChangeAt,
			GrowthPct:    growthPct,
			UpdatedAt:    class.UpdatedAt,
		})
	}
	return out, nil
}

// GetClassDetails returns class items, growth points, and history.
func (s *Service) GetClassDetails(ctx context.Context, input ListClassDetailsInput) (*ClassDetails, error) {
	if err := s.EnsureAccountProjection(ctx, input.AccountID); err != nil {
		return nil, err
	}
	class, err := s.store.FetchClassByID(ctx, input.AccountID, input.ClassID)
	if err != nil {
		return nil, fmt.Errorf("assets service: fetch class: %w", err)
	}
	if class == nil {
		return nil, ErrAssetClassNotFound
	}

	currentWorth, err := s.store.SumClassWorth(ctx, input.AccountID, input.ClassID)
	if err != nil {
		return nil, fmt.Errorf("assets service: sum class worth: %w", err)
	}
	latestHistory, err := s.store.ListHistoryByClass(ctx, input.AccountID, input.ClassID, 1, false)
	if err != nil {
		return nil, fmt.Errorf("assets service: list recent history: %w", err)
	}
	earliestHistory, err := s.store.ListHistoryByClass(ctx, input.AccountID, input.ClassID, 1, true)
	if err != nil {
		return nil, fmt.Errorf("assets service: list inception history: %w", err)
	}
	var lastChangeAt *time.Time
	var growthPct *float64
	if len(latestHistory) > 0 {
		last := latestHistory[0].EffectiveDate
		lastChangeAt = &last
	}
	if len(latestHistory) > 0 && len(earliestHistory) > 0 {
		growthPct = growthPctFromInception(earliestHistory[0].ClassTotalWorth, latestHistory[0].ClassTotalWorth)
	}

	items, err := s.store.ListItemsByClass(ctx, input.AccountID, input.ClassID, false)
	if err != nil {
		return nil, fmt.Errorf("assets service: list items: %w", err)
	}
	itemRows := make([]ItemSummary, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		itemRows = append(itemRows, ItemSummary{
			ID:           item.ID,
			Name:         item.Name,
			CurrentWorth: item.CurrentWorth,
			Archived:     item.Archived,
			UpdatedAt:    item.UpdatedAt,
		})
	}

	historyDesc, err := s.store.ListHistoryByClass(ctx, input.AccountID, input.ClassID, 500, false)
	if err != nil {
		return nil, fmt.Errorf("assets service: list history desc: %w", err)
	}
	historyLatestDesc, err := s.store.ListHistoryByClass(ctx, input.AccountID, input.ClassID, ClassGrowthHistoryWindow, false)
	if err != nil {
		return nil, fmt.Errorf("assets service: list class growth window: %w", err)
	}
	historyAsc := reverseHistory(historyLatestDesc)

	growth := toGrowthPoints(historyAsc)
	history := make([]HistoryEntry, 0, len(historyDesc))
	for _, entry := range historyDesc {
		if entry == nil {
			continue
		}
		history = append(history, *entry)
	}
	return &ClassDetails{
		Class: ClassSummary{
			ID:           class.ID,
			Name:         class.Name,
			Source:       class.Source,
			Archived:     class.Archived,
			CurrentWorth: currentWorth,
			LastChangeAt: lastChangeAt,
			GrowthPct:    growthPct,
			UpdatedAt:    class.UpdatedAt,
		},
		Items:   itemRows,
		Growth:  growth,
		History: history,
	}, nil
}
