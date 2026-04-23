package assets

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

const (
	// PortfolioClassName is the fixed name for the portfolio-linked class.
	PortfolioClassName = "Portfolio"
	// PortfolioItemName is the fixed item name for the portfolio-linked class.
	PortfolioItemName = "Portfolio Worth"
	// ClassGrowthHistoryWindow is the maximum class history rows used for growth chart points.
	ClassGrowthHistoryWindow = 5000
)

var (
	// ErrAssetAccountNotFound indicates the provided account id does not exist.
	ErrAssetAccountNotFound = fmt.Errorf("account not found")
	// ErrAssetClassNotFound indicates the selected class does not exist for the account.
	ErrAssetClassNotFound = fmt.Errorf("asset class not found")
	// ErrAssetClassNameRequired indicates class name is blank.
	ErrAssetClassNameRequired = fmt.Errorf("class name is required")
	// ErrAssetClassNameReserved indicates manual classes cannot use reserved names.
	ErrAssetClassNameReserved = fmt.Errorf("class name is reserved")
	// ErrAssetClassNotManual indicates mutation was attempted on a non-manual class.
	ErrAssetClassNotManual = fmt.Errorf("asset class is not manually managed")
	// ErrAssetClassAlreadyExists indicates the class name already exists for account.
	ErrAssetClassAlreadyExists = fmt.Errorf("asset class already exists")
	// ErrAssetItemNotFound indicates the selected item does not exist for the class/account.
	ErrAssetItemNotFound = fmt.Errorf("asset item not found")
	// ErrAssetItemNameRequired indicates item name is blank.
	ErrAssetItemNameRequired = fmt.Errorf("item name is required")
	// ErrAssetItemAlreadyExists indicates the item name already exists in class.
	ErrAssetItemAlreadyExists = fmt.Errorf("asset item already exists")
	// ErrAssetWorthInvalid indicates a set-worth value is malformed.
	ErrAssetWorthInvalid = fmt.Errorf("worth must be a decimal string with up to 6 decimals")
	// ErrAssetAmountInvalid indicates an adjust amount is malformed.
	ErrAssetAmountInvalid = fmt.Errorf("amount must be a positive decimal string with up to 6 decimals")
	// ErrAssetDirectionInvalid indicates adjust direction is malformed.
	ErrAssetDirectionInvalid = fmt.Errorf("direction must be one of: increase, decrease")
	// ErrAssetEffectiveDateInvalid indicates effective date is malformed.
	ErrAssetEffectiveDateInvalid = fmt.Errorf("effective_date must be in YYYY-MM-DD format")
	// ErrAssetEffectiveDateFuture indicates future effective dates are not allowed.
	ErrAssetEffectiveDateFuture = fmt.Errorf("effective_date cannot be in the future")
)

var (
	signedDecimalPattern = regexp.MustCompile(`^-?\d+(\.\d{1,6})?$`)
	positiveDecimalRegex = regexp.MustCompile(`^\d+(\.\d{1,6})?$`)
)

type assetAccountFetcher interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*account.Account, error)
}

type portfolioSnapshotLister interface {
	ListForAccount(ctx context.Context, accountID uuid.UUID, from, to *time.Time) ([]*portfolio.PortfolioSnapshot, error)
}

type assetStore interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error

	EnsureAccount(ctx context.Context, acc *Account) error

	CreateClass(ctx context.Context, class *Class) error
	FetchClassByID(ctx context.Context, accountID, classID uuid.UUID) (*Class, error)
	FetchClassBySource(ctx context.Context, accountID uuid.UUID, source ClassSource) (*Class, error)
	ListClassesForAccount(ctx context.Context, accountID uuid.UUID, includeArchived bool) ([]*Class, error)
	UpdateClass(ctx context.Context, accountID, classID uuid.UUID, name *string, archived *bool) error
	DeleteClass(ctx context.Context, accountID, classID uuid.UUID) (int64, error)

	CreateItem(ctx context.Context, item *Item) error
	FetchItemByID(ctx context.Context, accountID, classID, itemID uuid.UUID) (*Item, error)
	FetchItemByClassAndName(ctx context.Context, accountID, classID uuid.UUID, name string) (*Item, error)
	ListItemsByClass(ctx context.Context, accountID, classID uuid.UUID, includeArchived bool) ([]*Item, error)
	UpdateItemWorth(ctx context.Context, accountID, classID, itemID uuid.UUID, worth money.Price) error
	SumClassWorth(ctx context.Context, accountID, classID uuid.UUID) (money.Price, error)

	CreateHistory(ctx context.Context, entry *HistoryEntry) error
	DeleteHistoryByClass(ctx context.Context, accountID, classID uuid.UUID) error
	ListHistoryByClass(ctx context.Context, accountID, classID uuid.UUID, limit int, ascending bool) ([]*HistoryEntry, error)
	ListHistoryForAccount(ctx context.Context, accountID uuid.UUID, limit int, ascending bool) ([]*HistoryEntry, error)

	DeleteSnapshotsByAccount(ctx context.Context, accountID uuid.UUID) error
	UpsertSnapshots(ctx context.Context, snapshots []*Snapshot) error
	ListSnapshotsForAccount(ctx context.Context, accountID uuid.UUID, from, to *time.Time) ([]*Snapshot, error)
	SumAccountWorth(ctx context.Context, accountID uuid.UUID, includeArchived bool) (money.Price, error)
}

// Service orchestrates assets-domain behavior.
type Service struct {
	accounts          assetAccountFetcher
	portfolioSnapshot portfolioSnapshotLister
	store             assetStore
	publisher         bus.Bus
	now               func() time.Time
}

// NewService constructs an assets domain service.
func NewService(accounts assetAccountFetcher, snapshots portfolioSnapshotLister, store assetStore, publisher bus.Bus) *Service {
	return &Service{
		accounts:          accounts,
		portfolioSnapshot: snapshots,
		store:             store,
		publisher:         publisher,
		now:               time.Now,
	}
}

// CreateClassInput contains required fields to create a manual class.
type CreateClassInput struct {
	AccountID uuid.UUID
	Name      string
}

// UpdateClassInput contains mutable class fields.
type UpdateClassInput struct {
	AccountID uuid.UUID
	ClassID   uuid.UUID
	Name      *string
	Archived  *bool
}

// CreateItemInput creates an item under a manual class.
type CreateItemInput struct {
	AccountID     uuid.UUID
	ClassID       uuid.UUID
	Name          string
	InitialWorth  string
	EffectiveDate string
	Note          *string
}

// SetItemWorthInput replaces item worth with an absolute value.
type SetItemWorthInput struct {
	AccountID     uuid.UUID
	ClassID       uuid.UUID
	ItemID        uuid.UUID
	Worth         string
	EffectiveDate string
	Note          *string
}

// AdjustItemWorthInput adjusts item worth with explicit direction and amount.
type AdjustItemWorthInput struct {
	AccountID     uuid.UUID
	ClassID       uuid.UUID
	ItemID        uuid.UUID
	Direction     string
	Amount        string
	EffectiveDate string
	Note          *string
}

// ListClassDetailsInput loads slider details for one class.
type ListClassDetailsInput struct {
	AccountID uuid.UUID
	ClassID   uuid.UUID
}

// ListSnapshotsInput loads account-level daily total snapshots.
type ListSnapshotsInput struct {
	AccountID uuid.UUID
	From      *time.Time
	To        *time.Time
}

// EnsureAccountProjection ensures the assets account projection exists for accountID.
func (s *Service) EnsureAccountProjection(ctx context.Context, accountID uuid.UUID) error {
	if _, err := s.accounts.FetchByID(ctx, accountID); err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			return ErrAssetAccountNotFound
		}
		return fmt.Errorf("assets service: fetch account: %w", err)
	}
	if err := s.store.EnsureAccount(ctx, NewAccount(accountID)); err != nil {
		return fmt.Errorf("assets service: ensure assets account: %w", err)
	}
	return nil
}

// CreateClass creates a manual class for an account.
func (s *Service) CreateClass(ctx context.Context, input CreateClassInput) (*Class, error) {
	name, err := normalizeClassName(input.Name)
	if err != nil {
		return nil, err
	}
	if err := s.EnsureAccountProjection(ctx, input.AccountID); err != nil {
		return nil, err
	}

	now := s.now().UTC()
	class := &Class{
		ID:        uuid.New(),
		AccountID: input.AccountID,
		Name:      name,
		Source:    ClassSourceManual,
		Archived:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateClass(ctx, class); err != nil {
		if errors.Is(err, ErrAssetClassAlreadyExists) {
			return nil, err
		}
		return nil, fmt.Errorf("assets service: create class: %w", err)
	}
	s.requestSnapshotsRebuild(ctx, input.AccountID)
	return class, nil
}

// UpdateClass mutates a manual class name/archive status.
func (s *Service) UpdateClass(ctx context.Context, input UpdateClassInput) error {
	if err := s.EnsureAccountProjection(ctx, input.AccountID); err != nil {
		return err
	}
	class, err := s.store.FetchClassByID(ctx, input.AccountID, input.ClassID)
	if err != nil {
		return fmt.Errorf("assets service: fetch class: %w", err)
	}
	if class == nil {
		return ErrAssetClassNotFound
	}
	if class.Source != ClassSourceManual {
		return ErrAssetClassNotManual
	}

	var name *string
	if input.Name != nil {
		normalized, normErr := normalizeClassName(*input.Name)
		if normErr != nil {
			return normErr
		}
		name = &normalized
	}

	if err := s.store.UpdateClass(ctx, input.AccountID, input.ClassID, name, input.Archived); err != nil {
		if errors.Is(err, ErrAssetClassAlreadyExists) {
			return err
		}
		return fmt.Errorf("assets service: update class: %w", err)
	}
	s.requestSnapshotsRebuild(ctx, input.AccountID)
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

// CreateItem creates a new tracked asset item under a manual class.
func (s *Service) CreateItem(ctx context.Context, input CreateItemInput) (*Item, error) {
	name, err := normalizeItemName(input.Name)
	if err != nil {
		return nil, err
	}
	initialWorth, err := parseSignedWorth(input.InitialWorth)
	if err != nil {
		return nil, err
	}
	effectiveDate, err := parseEffectiveDate(input.EffectiveDate, s.now)
	if err != nil {
		return nil, err
	}
	note := normalizeOptionalNote(input.Note)

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
	if class.Source != ClassSourceManual {
		return nil, ErrAssetClassNotManual
	}

	now := s.now().UTC()
	item := &Item{
		ID:           uuid.New(),
		ClassID:      input.ClassID,
		AccountID:    input.AccountID,
		Name:         name,
		CurrentWorth: 0,
		Archived:     false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		if createErr := s.store.CreateItem(txCtx, item); createErr != nil {
			return createErr
		}
		if updateErr := s.store.UpdateItemWorth(txCtx, input.AccountID, input.ClassID, item.ID, initialWorth); updateErr != nil {
			return updateErr
		}
		item.CurrentWorth = initialWorth

		classTotal, sumErr := s.store.SumClassWorth(txCtx, input.AccountID, input.ClassID)
		if sumErr != nil {
			return sumErr
		}
		history := &HistoryEntry{
			ID:              uuid.New(),
			AccountID:       input.AccountID,
			ClassID:         input.ClassID,
			ItemID:          item.ID,
			ChangeType:      ChangeTypeSet,
			Direction:       nil,
			Amount:          initialWorth,
			PreviousWorth:   0,
			NewWorth:        initialWorth,
			ClassTotalWorth: classTotal,
			EffectiveDate:   effectiveDate,
			Note:            note,
			CreatedAt:       now,
		}
		return s.store.CreateHistory(txCtx, history)
	}); err != nil {
		if errors.Is(err, ErrAssetItemAlreadyExists) {
			return nil, err
		}
		return nil, fmt.Errorf("assets service: create item: %w", err)
	}
	s.requestSnapshotsRebuild(ctx, input.AccountID)
	return item, nil
}

// SetItemWorth sets absolute worth for an item.
func (s *Service) SetItemWorth(ctx context.Context, input SetItemWorthInput) error {
	worth, err := parseSignedWorth(input.Worth)
	if err != nil {
		return err
	}
	effectiveDate, err := parseEffectiveDate(input.EffectiveDate, s.now)
	if err != nil {
		return err
	}
	note := normalizeOptionalNote(input.Note)
	return s.applyWorthChange(ctx, input.AccountID, input.ClassID, input.ItemID, ChangeTypeSet, nil, worth, effectiveDate, note)
}

// AdjustItemWorth adjusts worth with explicit direction and positive amount.
func (s *Service) AdjustItemWorth(ctx context.Context, input AdjustItemWorthInput) error {
	amount, err := parsePositiveAmount(input.Amount)
	if err != nil {
		return err
	}
	effectiveDate, err := parseEffectiveDate(input.EffectiveDate, s.now)
	if err != nil {
		return err
	}
	direction, err := parseDirection(input.Direction)
	if err != nil {
		return err
	}
	note := normalizeOptionalNote(input.Note)
	return s.applyWorthChange(ctx, input.AccountID, input.ClassID, input.ItemID, ChangeTypeAdjust, &direction, amount, effectiveDate, note)
}

func (s *Service) applyWorthChange(
	ctx context.Context,
	accountID uuid.UUID,
	classID uuid.UUID,
	itemID uuid.UUID,
	changeType ChangeType,
	direction *ChangeDirection,
	amount money.Price,
	effectiveDate time.Time,
	note string,
) error {
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
	if class.Source != ClassSourceManual && changeType != ChangeTypeSet {
		// Portfolio worth is set only by sync flow.
		return ErrAssetClassNotManual
	}
	if class.Source == ClassSourcePortfolio {
		return ErrAssetClassNotManual
	}

	now := s.now().UTC()
	if err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		item, fetchErr := s.store.FetchItemByID(txCtx, accountID, classID, itemID)
		if fetchErr != nil {
			return fetchErr
		}
		if item == nil {
			return ErrAssetItemNotFound
		}
		previousWorth := item.CurrentWorth
		nextWorth := amount
		if changeType == ChangeTypeAdjust {
			if direction != nil && *direction == ChangeDirectionDecrease {
				nextWorth = previousWorth - amount
			} else {
				nextWorth = previousWorth + amount
			}
		}
		if updateErr := s.store.UpdateItemWorth(txCtx, accountID, classID, itemID, nextWorth); updateErr != nil {
			return updateErr
		}

		classTotal, sumErr := s.store.SumClassWorth(txCtx, accountID, classID)
		if sumErr != nil {
			return sumErr
		}

		entry := &HistoryEntry{
			ID:              uuid.New(),
			AccountID:       accountID,
			ClassID:         classID,
			ItemID:          itemID,
			ChangeType:      changeType,
			Direction:       direction,
			Amount:          amount,
			PreviousWorth:   previousWorth,
			NewWorth:        nextWorth,
			ClassTotalWorth: classTotal,
			EffectiveDate:   effectiveDate,
			Note:            note,
			CreatedAt:       now,
		}
		return s.store.CreateHistory(txCtx, entry)
	}); err != nil {
		return err
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

// ListSnapshots returns account-level daily total snapshots.
func (s *Service) ListSnapshots(ctx context.Context, input ListSnapshotsInput) ([]GrowthPoint, error) {
	if err := s.EnsureAccountProjection(ctx, input.AccountID); err != nil {
		return nil, err
	}
	rows, err := s.store.ListSnapshotsForAccount(ctx, input.AccountID, input.From, input.To)
	if err != nil {
		return nil, fmt.Errorf("assets service: list snapshots: %w", err)
	}
	out := make([]GrowthPoint, 0, len(rows))
	for _, row := range rows {
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

// RebuildTotalSnapshots rebuilds account-level assets snapshots from account history.
func (s *Service) RebuildTotalSnapshots(ctx context.Context, accountID uuid.UUID) error {
	if err := s.EnsureAccountProjection(ctx, accountID); err != nil {
		return err
	}

	historyAsc, err := s.store.ListHistoryForAccount(ctx, accountID, 0, true)
	if err != nil {
		return fmt.Errorf("assets service: list account history for snapshots: %w", err)
	}

	today := startOfDayUTC(s.now().UTC())
	rows := buildSnapshotsFromHistory(accountID, historyAsc, today)
	if len(rows) == 0 {
		totalWorth, sumErr := s.store.SumAccountWorth(ctx, accountID, true)
		if sumErr != nil {
			return fmt.Errorf("assets service: sum account worth for fallback snapshot: %w", sumErr)
		}
		now := s.now().UTC()
		rows = []*Snapshot{
			{
				ID:         uuid.New(),
				AccountID:  accountID,
				OccurredAt: today,
				TotalWorth: totalWorth,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		}
	}

	if err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		if deleteErr := s.store.DeleteSnapshotsByAccount(txCtx, accountID); deleteErr != nil {
			return deleteErr
		}
		return s.store.UpsertSnapshots(txCtx, rows)
	}); err != nil {
		return fmt.Errorf("assets service: rebuild account snapshots: %w", err)
	}
	return nil
}

func reverseHistory(rows []*HistoryEntry) []*HistoryEntry {
	out := make([]*HistoryEntry, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		out = append(out, rows[i])
	}
	return out
}

func startOfDayUTC(value time.Time) time.Time {
	v := value.UTC()
	return time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, time.UTC)
}

func buildSnapshotsFromHistory(accountID uuid.UUID, history []*HistoryEntry, today time.Time) []*Snapshot {
	if len(history) == 0 {
		return nil
	}
	type classKey = uuid.UUID
	type dayKey = string

	latestByClassAndDay := make(map[classKey]map[dayKey]money.Price)
	earliestDay := startOfDayUTC(today)
	earliestSet := false

	for _, row := range history {
		if row == nil {
			continue
		}
		day := startOfDayUTC(row.EffectiveDate)
		key := day.Format("2006-01-02")
		perClass := latestByClassAndDay[row.ClassID]
		if perClass == nil {
			perClass = make(map[dayKey]money.Price)
			latestByClassAndDay[row.ClassID] = perClass
		}
		// History is ordered ascending, so the latest row for each class/day wins.
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

func toGrowthPoints(entries []*HistoryEntry) []GrowthPoint {
	type point struct {
		date  time.Time
		worth money.Price
	}
	byDate := map[string]point{}
	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		date := time.Date(entry.EffectiveDate.Year(), entry.EffectiveDate.Month(), entry.EffectiveDate.Day(), 0, 0, 0, 0, time.UTC)
		key := date.Format("2006-01-02")
		if _, exists := byDate[key]; !exists {
			keys = append(keys, key)
		}
		byDate[key] = point{
			date:  date,
			worth: entry.ClassTotalWorth,
		}
	}
	sort.Strings(keys)
	out := make([]GrowthPoint, 0, len(keys))
	for _, key := range keys {
		entry := byDate[key]
		out = append(out, GrowthPoint{
			Date:       entry.date,
			TotalWorth: entry.worth,
		})
	}
	return out
}

func growthPctFromInception(inceptionWorth, latestWorth money.Price) *float64 {
	inception := inceptionWorth.Float64()
	if inception == 0 {
		return nil
	}
	latest := latestWorth.Float64()
	value := ((latest - inception) / math.Abs(inception)) * 100
	return &value
}

// SyncPortfolioClassFromRebuild updates the portfolio-linked class from portfolio snapshots.
func (s *Service) SyncPortfolioClassFromRebuild(ctx context.Context, accountID uuid.UUID) error {
	if err := s.EnsureAccountProjection(ctx, accountID); err != nil {
		return err
	}
	snapshots, err := s.portfolioSnapshot.ListForAccount(ctx, accountID, nil, nil)
	if err != nil {
		return fmt.Errorf("assets service: list portfolio snapshots: %w", err)
	}
	if len(snapshots) == 0 {
		return nil
	}
	filtered := make([]*portfolio.PortfolioSnapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		if snapshot == nil {
			continue
		}
		filtered = append(filtered, snapshot)
	}
	if len(filtered) == 0 {
		return nil
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].OccurredAt.UTC().Before(filtered[j].OccurredAt.UTC())
	})
	if err := s.syncPortfolioHistoryFromSnapshots(ctx, accountID, filtered); err != nil {
		return err
	}
	s.requestSnapshotsRebuild(ctx, accountID)
	return nil
}

func (s *Service) syncPortfolioHistoryFromSnapshots(ctx context.Context, accountID uuid.UUID, snapshots []*portfolio.PortfolioSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	now := s.now().UTC()
	return s.store.WithTx(ctx, func(txCtx context.Context) error {
		class, err := s.store.FetchClassBySource(txCtx, accountID, ClassSourcePortfolio)
		if err != nil {
			return err
		}
		if class == nil {
			class = &Class{
				ID:        uuid.New(),
				AccountID: accountID,
				Name:      PortfolioClassName,
				Source:    ClassSourcePortfolio,
				Archived:  false,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := s.store.CreateClass(txCtx, class); err != nil {
				return err
			}
		}

		item, err := s.store.FetchItemByClassAndName(txCtx, accountID, class.ID, PortfolioItemName)
		if err != nil {
			return err
		}
		if item == nil {
			item = &Item{
				ID:           uuid.New(),
				ClassID:      class.ID,
				AccountID:    accountID,
				Name:         PortfolioItemName,
				CurrentWorth: 0,
				Archived:     false,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := s.store.CreateItem(txCtx, item); err != nil {
				return err
			}
		}

		latest := snapshots[len(snapshots)-1]
		if latest == nil {
			return nil
		}
		if err := s.store.UpdateItemWorth(txCtx, accountID, class.ID, item.ID, latest.MarketValue); err != nil {
			return err
		}

		if err := s.store.DeleteHistoryByClass(txCtx, accountID, class.ID); err != nil {
			return err
		}

		previousWorth := money.Price(0)
		for _, snapshot := range snapshots {
			if snapshot == nil {
				continue
			}
			date := snapshot.OccurredAt.UTC()
			entry := &HistoryEntry{
				ID:              uuid.New(),
				AccountID:       accountID,
				ClassID:         class.ID,
				ItemID:          item.ID,
				ChangeType:      ChangeTypeSet,
				Direction:       nil,
				Amount:          snapshot.MarketValue,
				PreviousWorth:   previousWorth,
				NewWorth:        snapshot.MarketValue,
				ClassTotalWorth: snapshot.MarketValue,
				EffectiveDate:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
				Note:            "synced from portfolio snapshot",
				CreatedAt:       now,
			}
			if err := s.store.CreateHistory(txCtx, entry); err != nil {
				return err
			}
			previousWorth = snapshot.MarketValue
		}
		return nil
	})
}

func (s *Service) requestSnapshotsRebuild(ctx context.Context, accountID uuid.UUID) {
	if s.publisher == nil || accountID == uuid.Nil {
		return
	}
	env, err := bus.NewJSONEnvelopeFromContext(ctx, api.AssetsSnapshotsRebuildRequested{AccID: accountID})
	if err != nil {
		return
	}
	_ = s.publisher.Publish(ctx, env)
}

func normalizeClassName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", ErrAssetClassNameRequired
	}
	if strings.EqualFold(name, PortfolioClassName) {
		return "", ErrAssetClassNameReserved
	}
	return name, nil
}

func normalizeItemName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", ErrAssetItemNameRequired
	}
	return name, nil
}

func parseSignedWorth(raw string) (money.Price, error) {
	value := strings.TrimSpace(raw)
	if !signedDecimalPattern.MatchString(value) {
		return 0, ErrAssetWorthInvalid
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(floatValue) || math.IsInf(floatValue, 0) {
		return 0, ErrAssetWorthInvalid
	}

	sign := 1.0
	if floatValue < 0 {
		sign = -1.0
		floatValue = math.Abs(floatValue)
	}
	parsed, err := money.NewPrice(floatValue)
	if err != nil {
		return 0, ErrAssetWorthInvalid
	}
	if sign < 0 {
		return -parsed, nil
	}
	return parsed, nil
}

func parsePositiveAmount(raw string) (money.Price, error) {
	value := strings.TrimSpace(raw)
	if !positiveDecimalRegex.MatchString(value) {
		return 0, ErrAssetAmountInvalid
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil || floatValue <= 0 {
		return 0, ErrAssetAmountInvalid
	}
	parsed, err := money.NewPrice(floatValue)
	if err != nil {
		return 0, ErrAssetAmountInvalid
	}
	return parsed, nil
}

func parseDirection(raw string) (ChangeDirection, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "increase":
		return ChangeDirectionIncrease, nil
	case "decrease":
		return ChangeDirectionDecrease, nil
	default:
		return "", ErrAssetDirectionInvalid
	}
}

func parseEffectiveDate(raw string, nowFn func() time.Time) (time.Time, error) {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, ErrAssetEffectiveDateInvalid
	}
	date = date.UTC()
	now := nowFn().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if date.After(today) {
		return time.Time{}, ErrAssetEffectiveDateFuture
	}
	return date, nil
}

func normalizeOptionalNote(note *string) string {
	if note == nil {
		return ""
	}
	return strings.TrimSpace(*note)
}
