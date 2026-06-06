package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
)

// SQLXAssetsStore persists the assets aggregate: account projection, classes, items,
// mutation history, and account-level snapshots. It satisfies the assets command,
// getter, aggregator, query, builder, sync, and unit-of-work contracts.
//
// The mutation history table column is item_id, while the domain Mutation models it
// as AssetID (db:"asset_id"); reads alias item_id AS asset_id and writes map
// :asset_id into item_id.
type SQLXAssetsStore struct {
	db             *DB
	accountsTable  string
	classesTable   string
	itemsTable     string
	historyTable   string
	snapshotsTable string
}

var (
	_ assets.CommandStore    = (*SQLXAssetsStore)(nil)
	_ assets.CommandGetter   = (*SQLXAssetsStore)(nil)
	_ assets.ClassAggregator = (*SQLXAssetsStore)(nil)
	_ assets.UnitOfWork      = (*SQLXAssetsStore)(nil)
	_ assets.QueryStore      = (*SQLXAssetsStore)(nil)
	_ assets.BuilderStore    = (*SQLXAssetsStore)(nil)
	_ assets.SyncStore       = (*SQLXAssetsStore)(nil)
)

// assetHistorySelectColumns aliases item_id to the domain's asset_id db tag.
const assetHistorySelectColumns = `
	id, account_id, class_id, item_id AS asset_id, change_type, direction,
	amount, previous_worth, new_worth, class_total_worth, effective_date, note, created_at
`

// NewSQLXAssetsStore creates an assets store backed by SQLX.
func NewSQLXAssetsStore(db *DB) *SQLXAssetsStore {
	return &SQLXAssetsStore{
		db:             db,
		accountsTable:  qualifyTableAs(db, SchemaAssets, TableAccounts, "assets_accounts"),
		classesTable:   qualifyTableAs(db, SchemaAssets, TableAssetClasses, "asset_classes"),
		itemsTable:     qualifyTableAs(db, SchemaAssets, TableAssetItems, "asset_items"),
		historyTable:   qualifyTableAs(db, SchemaAssets, TableAssetHistory, "asset_histories"),
		snapshotsTable: qualifyTableAs(db, SchemaAssets, TableAssetSnapshot, "asset_snapshots"),
	}
}

// Do runs fn within a single database transaction.
func (s *SQLXAssetsStore) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	return s.db.WithTx(ctx, fn)
}

// --- Account projection -----------------------------------------------------

// CreateAccount inserts the assets account projection, ignoring an existing row.
func (s *SQLXAssetsStore) CreateAccount(ctx context.Context, acc *assets.Account) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, created_at, updated_at)
		VALUES (:id, :account_id, :created_at, :updated_at)
		ON CONFLICT (account_id) DO NOTHING
	`, s.accountsTable)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, acc); err != nil {
		return fmt.Errorf("assets store: create account: %w", err)
	}
	return nil
}

// --- Classes ----------------------------------------------------------------

// CreateClass inserts one class, mapping the per-account name unique violation to
// assets.ErrClassAlreadyExists.
func (s *SQLXAssetsStore) CreateClass(ctx context.Context, class *assets.Class) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, name, source, archived, created_at, updated_at)
		VALUES (:id, :account_id, :name, :source, :archived, :created_at, :updated_at)
	`, s.classesTable)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, class); err != nil {
		if isUniqueViolation(err) {
			return assets.ErrClassAlreadyExists
		}
		return fmt.Errorf("assets store: create class: %w", err)
	}
	return nil
}

// UpdateClass persists a class's name and archived flag.
func (s *SQLXAssetsStore) UpdateClass(ctx context.Context, class *assets.Class) error {
	class.UpdatedAt = time.Now().UTC()
	query := fmt.Sprintf(`
		UPDATE %s SET name = :name, archived = :archived, updated_at = :updated_at WHERE id = :id
	`, s.classesTable)
	res, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, class)
	if err != nil {
		if isUniqueViolation(err) {
			return assets.ErrClassAlreadyExists
		}
		return fmt.Errorf("assets store: update class: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return assets.ErrClassNotFound
	}
	return nil
}

// DeleteClass removes a class by ID; items and history cascade via foreign keys.
func (s *SQLXAssetsStore) DeleteClass(ctx context.Context, classID uuid.UUID) error {
	query := s.db.Rebind(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, s.classesTable))
	if _, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, classID); err != nil {
		return fmt.Errorf("assets store: delete class: %w", err)
	}
	return nil
}

// Class returns a class with its items and mutations (newest first), or (nil, nil)
// when it does not exist.
func (s *SQLXAssetsStore) Class(ctx context.Context, classID uuid.UUID) (*assets.Class, error) {
	var class assets.Class
	query := s.db.Rebind(fmt.Sprintf(`SELECT * FROM %s WHERE id = ? LIMIT 1`, s.classesTable))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &class, query, classID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("assets store: get class: %w", err)
	}

	itemsQuery := s.db.Rebind(fmt.Sprintf(`SELECT * FROM %s WHERE class_id = ? ORDER BY name ASC, created_at ASC`, s.itemsTable))
	var items []assets.Asset
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &items, itemsQuery, classID); err != nil {
		return nil, fmt.Errorf("assets store: get class items: %w", err)
	}
	class.Assets = items

	mutQuery := s.db.Rebind(fmt.Sprintf(`SELECT %s FROM %s WHERE class_id = ? ORDER BY effective_date DESC, created_at DESC`, assetHistorySelectColumns, s.historyTable))
	var muts []assets.Mutation
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &muts, mutQuery, classID); err != nil {
		return nil, fmt.Errorf("assets store: get class mutations: %w", err)
	}
	class.Mutations = muts

	return &class, nil
}

// ClassBySource returns the account's class for the given source, or (nil, nil) when absent.
func (s *SQLXAssetsStore) ClassBySource(ctx context.Context, accountID uuid.UUID, source assets.ClassSource) (*assets.Class, error) {
	var class assets.Class
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s WHERE account_id = ? AND source = ? ORDER BY created_at ASC LIMIT 1
	`, s.classesTable))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &class, query, accountID, source); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("assets store: class by source: %w", err)
	}
	return &class, nil
}

// ClassesForAccount returns the account's classes keyed by class ID.
func (s *SQLXAssetsStore) ClassesForAccount(ctx context.Context, accID uuid.UUID, includeArchived bool) (map[uuid.UUID]*assets.Class, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ?`, s.classesTable)
	args := []any{accID}
	if !includeArchived {
		query += " AND archived = ?"
		args = append(args, false)
	}
	query += " ORDER BY name ASC, created_at ASC"

	var rows []*assets.Class
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("assets store: classes for account: %w", err)
	}
	out := make(map[uuid.UUID]*assets.Class, len(rows))
	for _, class := range rows {
		out[class.ID] = class
	}
	return out, nil
}

// ClassBounds returns the earliest and latest mutation per class for the given IDs.
func (s *SQLXAssetsStore) ClassBounds(ctx context.Context, classIDs []uuid.UUID) (map[uuid.UUID]*assets.ClassBounds, error) {
	result := make(map[uuid.UUID]*assets.ClassBounds, len(classIDs))
	if len(classIDs) == 0 {
		return result, nil
	}
	query, args, err := sqlx.In(fmt.Sprintf(`
		SELECT %s FROM %s WHERE class_id IN (?)
		ORDER BY class_id, effective_date ASC, created_at ASC
	`, assetHistorySelectColumns, s.historyTable), classIDs)
	if err != nil {
		return nil, fmt.Errorf("assets store: class bounds expand: %w", err)
	}
	var muts []assets.Mutation
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &muts, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("assets store: class bounds: %w", err)
	}
	for i := range muts {
		m := muts[i]
		bound, ok := result[m.ClassID]
		if !ok {
			first := m
			result[m.ClassID] = &assets.ClassBounds{ClassID: m.ClassID, First: &first, Last: &first}
			continue
		}
		last := m
		bound.Last = &last
	}
	return result, nil
}

// AggregateValue returns the summed worth of a class's non-archived items.
func (s *SQLXAssetsStore) AggregateValue(ctx context.Context, accID, classID uuid.UUID) (money.Price, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT COALESCE(SUM(current_worth), 0) FROM %s
		WHERE account_id = ? AND class_id = ? AND archived = ?
	`, s.itemsTable))
	var total int64
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &total, query, accID, classID, false); err != nil {
		return 0, fmt.Errorf("assets store: aggregate value: %w", err)
	}
	return money.Price(total), nil
}

// --- Items ------------------------------------------------------------------

// CreateAsset inserts one asset item.
func (s *SQLXAssetsStore) CreateAsset(ctx context.Context, asset *assets.Asset) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, class_id, account_id, name, current_worth, archived, created_at, updated_at)
		VALUES (:id, :class_id, :account_id, :name, :current_worth, :archived, :created_at, :updated_at)
	`, s.itemsTable)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, asset); err != nil {
		return fmt.Errorf("assets store: create asset: %w", err)
	}
	return nil
}

// Asset returns an asset item with its class populated, or (nil, nil) when absent.
func (s *SQLXAssetsStore) Asset(ctx context.Context, assetID uuid.UUID) (*assets.Asset, error) {
	var item assets.Asset
	query := s.db.Rebind(fmt.Sprintf(`SELECT * FROM %s WHERE id = ? LIMIT 1`, s.itemsTable))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &item, query, assetID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("assets store: get asset: %w", err)
	}

	var class assets.Class
	classQuery := s.db.Rebind(fmt.Sprintf(`SELECT * FROM %s WHERE id = ? LIMIT 1`, s.classesTable))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &class, classQuery, item.ClassID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("assets store: get asset class: %w", err)
		}
	} else {
		item.Class = &class
	}
	return &item, nil
}

// AssetByClassAndName returns an item by case-insensitive name within a class, or
// (nil, nil) when absent.
func (s *SQLXAssetsStore) AssetByClassAndName(ctx context.Context, accountID, classID uuid.UUID, name string) (*assets.Asset, error) {
	var item assets.Asset
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s WHERE class_id = ? AND account_id = ? AND LOWER(name) = LOWER(?) LIMIT 1
	`, s.itemsTable))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &item, query, classID, accountID, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("assets store: asset by class and name: %w", err)
	}
	return &item, nil
}

// SetWorth persists an asset's current worth.
func (s *SQLXAssetsStore) SetWorth(ctx context.Context, asset *assets.Asset) error {
	asset.UpdatedAt = time.Now().UTC()
	query := fmt.Sprintf(`UPDATE %s SET current_worth = :current_worth, updated_at = :updated_at WHERE id = :id`, s.itemsTable)
	res, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, asset)
	if err != nil {
		return fmt.Errorf("assets store: set worth: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return assets.ErrAssetNotFound
	}
	return nil
}

// UpdateAssetWorth sets an item's worth scoped by account/class/item.
func (s *SQLXAssetsStore) UpdateAssetWorth(ctx context.Context, accountID, classID, assetID uuid.UUID, worth money.Price) error {
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s SET current_worth = ?, updated_at = ? WHERE id = ? AND class_id = ? AND account_id = ?
	`, s.itemsTable))
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, worth, time.Now().UTC(), assetID, classID, accountID)
	if err != nil {
		return fmt.Errorf("assets store: update asset worth: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return assets.ErrAssetNotFound
	}
	return nil
}

// --- Mutations --------------------------------------------------------------

// CreateMutation inserts one mutation history entry. AssetID maps to the item_id column.
func (s *SQLXAssetsStore) CreateMutation(ctx context.Context, mut *assets.Mutation) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, class_id, item_id, change_type, direction, amount,
			previous_worth, new_worth, class_total_worth, effective_date, note, created_at
		) VALUES (
			:id, :account_id, :class_id, :asset_id, :change_type, :direction, :amount,
			:previous_worth, :new_worth, :class_total_worth, :effective_date, :note, :created_at
		)
	`, s.historyTable)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, mut); err != nil {
		return fmt.Errorf("assets store: create mutation: %w", err)
	}
	return nil
}

// Mutations returns an account's mutation history ordered by effective date, optionally paginated.
func (s *SQLXAssetsStore) Mutations(ctx context.Context, accID uuid.UUID, sort *sorting.Direction, skip, take *uint64) ([]*assets.Mutation, error) {
	dir := "ASC"
	if sort != nil {
		dir = normalizeSortOrder(string(*sort))
	}
	query := fmt.Sprintf(`
		SELECT %s FROM %s WHERE account_id = ?
		ORDER BY effective_date %s, created_at %s
	`, assetHistorySelectColumns, s.historyTable, dir, dir)
	args := []any{accID}
	if take != nil {
		query += " LIMIT ?"
		args = append(args, int64(*take))
		if skip != nil {
			query += " OFFSET ?"
			args = append(args, int64(*skip))
		}
	}
	var muts []*assets.Mutation
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &muts, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("assets store: mutations: %w", err)
	}
	return muts, nil
}

// --- Snapshots --------------------------------------------------------------

// Snapshots returns an account's daily total-worth snapshots within an optional range.
func (s *SQLXAssetsStore) Snapshots(ctx context.Context, accID uuid.UUID, from, to *time.Time) ([]*assets.Snapshot, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ?`, s.snapshotsTable)
	args := []any{accID}
	if from != nil {
		query += " AND occurred_at >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND occurred_at <= ?"
		args = append(args, *to)
	}
	query += " ORDER BY occurred_at ASC, id ASC"

	rows := make([]*assets.Snapshot, 0)
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("assets store: snapshots: %w", err)
	}
	return rows, nil
}

// DeleteSnapshots removes all snapshots for an account.
func (s *SQLXAssetsStore) DeleteSnapshots(ctx context.Context, accID uuid.UUID) error {
	query := s.db.Rebind(fmt.Sprintf(`DELETE FROM %s WHERE account_id = ?`, s.snapshotsTable))
	if _, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, accID); err != nil {
		return fmt.Errorf("assets store: delete snapshots: %w", err)
	}
	return nil
}

// StoreSnapshots upserts account snapshots keyed by (account, day).
func (s *SQLXAssetsStore) StoreSnapshots(ctx context.Context, snapshots []*assets.Snapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, occurred_at, total_worth, created_at, updated_at)
		VALUES (:id, :account_id, :occurred_at, :total_worth, :created_at, :updated_at)
		ON CONFLICT (account_id, occurred_at) DO UPDATE SET
			total_worth = EXCLUDED.total_worth,
			updated_at = EXCLUDED.updated_at
	`, s.snapshotsTable)
	executor := s.db.GetExecutor(ctx)
	for _, snapshot := range snapshots {
		if snapshot == nil {
			continue
		}
		if snapshot.ID == uuid.Nil {
			snapshot.ID = uuid.New()
		}
		if snapshot.CreatedAt.IsZero() {
			snapshot.CreatedAt = time.Now().UTC()
		}
		snapshot.OccurredAt = date.StartOfDayUTC(snapshot.OccurredAt)
		snapshot.UpdatedAt = time.Now().UTC()
		if _, err := sqlx.NamedExecContext(ctx, executor, query, snapshot); err != nil {
			return fmt.Errorf("assets store: store snapshots: %w", err)
		}
	}
	return nil
}

// --- Sync -------------------------------------------------------------------

// CleanPortfolio removes the account's portfolio-sourced class so the sync flow can
// rebuild it; its items and history cascade via foreign keys.
func (s *SQLXAssetsStore) CleanPortfolio(ctx context.Context, accountID uuid.UUID) error {
	query := s.db.Rebind(fmt.Sprintf(`DELETE FROM %s WHERE account_id = ? AND source = ?`, s.classesTable))
	if _, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, accountID, assets.ClassSourcePortfolio); err != nil {
		return fmt.Errorf("assets store: clean portfolio: %w", err)
	}
	return nil
}

// TryAcquireSyncLock is a best-effort no-op: the assets schema has no per-account
// lock column, so it always reports the lock as acquired. Proper locking would
// require an assets account lock column (a migration, out of scope here).
func (s *SQLXAssetsStore) TryAcquireSyncLock(ctx context.Context, id uuid.UUID) (bool, error) {
	return true, nil
}

// ReleaseSyncLock is the no-op counterpart to TryAcquireSyncLock.
func (s *SQLXAssetsStore) ReleaseSyncLock(ctx context.Context, id uuid.UUID) error {
	return nil
}
