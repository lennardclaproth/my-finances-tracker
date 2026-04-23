package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lib/pq"
)

// SQLXAssetStore persists assets-domain projections and mutation history.
type SQLXAssetStore struct {
	db             *DB
	accountsTable  string
	classesTable   string
	itemsTable     string
	historyTable   string
	snapshotsTable string
}

// NewSQLXAssetStore constructs a SQLX-backed assets store.
func NewSQLXAssetStore(db *DB) *SQLXAssetStore {
	return &SQLXAssetStore{
		db:             db,
		accountsTable:  qualifyAssetsAccountsTable(db),
		classesTable:   qualifyAssetsClassesTable(db),
		itemsTable:     qualifyAssetsItemsTable(db),
		historyTable:   qualifyAssetsHistoryTable(db),
		snapshotsTable: qualifyAssetsSnapshotsTable(db),
	}
}

func qualifyAssetsAccountsTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "assets_accounts"
	}
	return fmt.Sprintf("%s.%s", SchemaAssets, TableAccounts)
}

func qualifyAssetsClassesTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "asset_classes"
	}
	return fmt.Sprintf("%s.%s", SchemaAssets, TableAssetClasses)
}

func qualifyAssetsItemsTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "asset_items"
	}
	return fmt.Sprintf("%s.%s", SchemaAssets, TableAssetItems)
}

func qualifyAssetsHistoryTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "asset_histories"
	}
	return fmt.Sprintf("%s.%s", SchemaAssets, TableAssetHistory)
}

func qualifyAssetsSnapshotsTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "asset_snapshots"
	}
	return fmt.Sprintf("%s.%s", SchemaAssets, TableAssetSnapshot)
}

// WithTx executes a function in a transaction.
func (s *SQLXAssetStore) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return s.db.WithTx(ctx, fn)
}

// EnsureAccount ensures account projection row exists.
func (s *SQLXAssetStore) EnsureAccount(ctx context.Context, acc *assets.Account) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, created_at, updated_at)
		VALUES (:id, :account_id, :created_at, :updated_at)
		ON CONFLICT (account_id) DO NOTHING
	`, s.accountsTable)
	_, err := s.db.NamedExecContext(ctx, query, acc)
	if err != nil {
		return fmt.Errorf("sqlx_asset_store: ensure account: %w", err)
	}
	return nil
}

// CreateClass inserts one class.
func (s *SQLXAssetStore) CreateClass(ctx context.Context, class *assets.Class) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, name, source, archived, created_at, updated_at)
		VALUES (:id, :account_id, :name, :source, :archived, :created_at, :updated_at)
	`, s.classesTable)
	executor := s.db.GetExecutor(ctx)
	if _, err := sqlx.NamedExecContext(ctx, executor, query, class); err != nil {
		if isAssetClassDuplicate(err) {
			return assets.ErrAssetClassAlreadyExists
		}
		return fmt.Errorf("sqlx_asset_store: create class: %w", err)
	}
	return nil
}

// FetchClassByID returns one class for account.
func (s *SQLXAssetStore) FetchClassByID(ctx context.Context, accountID, classID uuid.UUID) (*assets.Class, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s
		WHERE id = ? AND account_id = ?
		LIMIT 1
	`, s.classesTable))
	executor := s.db.GetExecutor(ctx)
	var class assets.Class
	if err := sqlx.GetContext(ctx, executor, &class, query, classID, accountID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("sqlx_asset_store: fetch class by id: %w", err)
	}
	return &class, nil
}

// FetchClassBySource returns one class by source for account.
func (s *SQLXAssetStore) FetchClassBySource(ctx context.Context, accountID uuid.UUID, source assets.ClassSource) (*assets.Class, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s
		WHERE account_id = ? AND source = ?
		ORDER BY created_at ASC
		LIMIT 1
	`, s.classesTable))
	executor := s.db.GetExecutor(ctx)
	var class assets.Class
	if err := sqlx.GetContext(ctx, executor, &class, query, accountID, source); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("sqlx_asset_store: fetch class by source: %w", err)
	}
	return &class, nil
}

// ListClassesForAccount returns classes sorted by name.
func (s *SQLXAssetStore) ListClassesForAccount(ctx context.Context, accountID uuid.UUID, includeArchived bool) ([]*assets.Class, error) {
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE account_id = ?
	`, s.classesTable)
	args := []any{accountID}
	if !includeArchived {
		query += " AND archived = ?"
		args = append(args, false)
	}
	query += " ORDER BY name ASC, created_at ASC"
	query = s.db.Rebind(query)

	executor := s.db.GetExecutor(ctx)
	rows := make([]*assets.Class, 0)
	if err := sqlx.SelectContext(ctx, executor, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("sqlx_asset_store: list classes: %w", err)
	}
	return rows, nil
}

// UpdateClass mutates class name/archive values.
func (s *SQLXAssetStore) UpdateClass(ctx context.Context, accountID, classID uuid.UUID, name *string, archived *bool) error {
	assignments := []string{"updated_at = ?"}
	args := []any{time.Now().UTC()}

	if name != nil {
		assignments = append(assignments, "name = ?")
		args = append(args, *name)
	}
	if archived != nil {
		assignments = append(assignments, "archived = ?")
		args = append(args, *archived)
	}
	args = append(args, classID, accountID)

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE id = ? AND account_id = ?
	`, s.classesTable, strings.Join(assignments, ", "))
	query = s.db.Rebind(query)
	executor := s.db.GetExecutor(ctx)
	res, err := executor.ExecContext(ctx, query, args...)
	if err != nil {
		if isAssetClassDuplicate(err) {
			return assets.ErrAssetClassAlreadyExists
		}
		return fmt.Errorf("sqlx_asset_store: update class: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlx_asset_store: update class rows affected: %w", err)
	}
	if affected == 0 {
		return assets.ErrAssetClassNotFound
	}
	return nil
}

// DeleteClass deletes one class for account.
func (s *SQLXAssetStore) DeleteClass(ctx context.Context, accountID, classID uuid.UUID) (int64, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		DELETE FROM %s
		WHERE id = ? AND account_id = ?
	`, s.classesTable))
	executor := s.db.GetExecutor(ctx)
	res, err := executor.ExecContext(ctx, query, classID, accountID)
	if err != nil {
		return 0, fmt.Errorf("sqlx_asset_store: delete class: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("sqlx_asset_store: delete class rows affected: %w", err)
	}
	return affected, nil
}

// CreateItem inserts one item.
func (s *SQLXAssetStore) CreateItem(ctx context.Context, item *assets.Item) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, class_id, account_id, name, current_worth, archived, created_at, updated_at)
		VALUES (:id, :class_id, :account_id, :name, :current_worth, :archived, :created_at, :updated_at)
	`, s.itemsTable)
	executor := s.db.GetExecutor(ctx)
	if _, err := sqlx.NamedExecContext(ctx, executor, query, item); err != nil {
		if isAssetItemDuplicate(err) {
			return assets.ErrAssetItemAlreadyExists
		}
		return fmt.Errorf("sqlx_asset_store: create item: %w", err)
	}
	return nil
}

// FetchItemByID returns one item scoped by account/class.
func (s *SQLXAssetStore) FetchItemByID(ctx context.Context, accountID, classID, itemID uuid.UUID) (*assets.Item, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s
		WHERE id = ? AND class_id = ? AND account_id = ?
		LIMIT 1
	`, s.itemsTable))
	executor := s.db.GetExecutor(ctx)
	var item assets.Item
	if err := sqlx.GetContext(ctx, executor, &item, query, itemID, classID, accountID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("sqlx_asset_store: fetch item by id: %w", err)
	}
	return &item, nil
}

// FetchItemByClassAndName returns one item by class/name.
func (s *SQLXAssetStore) FetchItemByClassAndName(ctx context.Context, accountID, classID uuid.UUID, name string) (*assets.Item, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s
		WHERE class_id = ? AND account_id = ? AND LOWER(name) = LOWER(?)
		LIMIT 1
	`, s.itemsTable))
	executor := s.db.GetExecutor(ctx)
	var item assets.Item
	if err := sqlx.GetContext(ctx, executor, &item, query, classID, accountID, name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("sqlx_asset_store: fetch item by name: %w", err)
	}
	return &item, nil
}

// ListItemsByClass lists items sorted by name.
func (s *SQLXAssetStore) ListItemsByClass(ctx context.Context, accountID, classID uuid.UUID, includeArchived bool) ([]*assets.Item, error) {
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE account_id = ? AND class_id = ?
	`, s.itemsTable)
	args := []any{accountID, classID}
	if !includeArchived {
		query += " AND archived = ?"
		args = append(args, false)
	}
	query += " ORDER BY name ASC, created_at ASC"
	query = s.db.Rebind(query)
	executor := s.db.GetExecutor(ctx)
	items := make([]*assets.Item, 0)
	if err := sqlx.SelectContext(ctx, executor, &items, query, args...); err != nil {
		return nil, fmt.Errorf("sqlx_asset_store: list items: %w", err)
	}
	return items, nil
}

// UpdateItemWorth updates current worth for one item.
func (s *SQLXAssetStore) UpdateItemWorth(ctx context.Context, accountID, classID, itemID uuid.UUID, worth money.Price) error {
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s
		SET current_worth = ?, updated_at = ?
		WHERE id = ? AND class_id = ? AND account_id = ?
	`, s.itemsTable))
	executor := s.db.GetExecutor(ctx)
	res, err := executor.ExecContext(ctx, query, worth, time.Now().UTC(), itemID, classID, accountID)
	if err != nil {
		return fmt.Errorf("sqlx_asset_store: update item worth: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlx_asset_store: update item worth rows affected: %w", err)
	}
	if affected == 0 {
		return assets.ErrAssetItemNotFound
	}
	return nil
}

// SumClassWorth sums all non-archived item worth values for one class.
func (s *SQLXAssetStore) SumClassWorth(ctx context.Context, accountID, classID uuid.UUID) (money.Price, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT COALESCE(SUM(current_worth), 0)
		FROM %s
		WHERE account_id = ? AND class_id = ? AND archived = ?
	`, s.itemsTable))
	executor := s.db.GetExecutor(ctx)
	var total int64
	if err := sqlx.GetContext(ctx, executor, &total, query, accountID, classID, false); err != nil {
		return 0, fmt.Errorf("sqlx_asset_store: sum class worth: %w", err)
	}
	return money.Price(total), nil
}

// SumAccountWorth sums item worth values for one account.
func (s *SQLXAssetStore) SumAccountWorth(ctx context.Context, accountID uuid.UUID, includeArchived bool) (money.Price, error) {
	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(i.current_worth), 0)
		FROM %s i
		WHERE i.account_id = ?
	`, s.itemsTable)
	args := []any{accountID}
	if !includeArchived {
		query += " AND i.archived = ?"
		args = append(args, false)
	}
	query = s.db.Rebind(query)

	executor := s.db.GetExecutor(ctx)
	var total int64
	if err := sqlx.GetContext(ctx, executor, &total, query, args...); err != nil {
		return 0, fmt.Errorf("sqlx_asset_store: sum account worth: %w", err)
	}
	return money.Price(total), nil
}

// CreateHistory inserts one history entry.
func (s *SQLXAssetStore) CreateHistory(ctx context.Context, entry *assets.HistoryEntry) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, class_id, item_id, change_type, direction, amount, previous_worth,
			new_worth, class_total_worth, effective_date, note, created_at
		) VALUES (
			:id, :account_id, :class_id, :item_id, :change_type, :direction, :amount, :previous_worth,
			:new_worth, :class_total_worth, :effective_date, :note, :created_at
		)
	`, s.historyTable)
	executor := s.db.GetExecutor(ctx)
	if _, err := sqlx.NamedExecContext(ctx, executor, query, entry); err != nil {
		return fmt.Errorf("sqlx_asset_store: create history: %w", err)
	}
	return nil
}

// DeleteHistoryByClass deletes all history entries for an account/class pair.
func (s *SQLXAssetStore) DeleteHistoryByClass(ctx context.Context, accountID, classID uuid.UUID) error {
	query := s.db.Rebind(fmt.Sprintf(`
		DELETE FROM %s
		WHERE account_id = ? AND class_id = ?
	`, s.historyTable))
	executor := s.db.GetExecutor(ctx)
	if _, err := executor.ExecContext(ctx, query, accountID, classID); err != nil {
		return fmt.Errorf("sqlx_asset_store: delete history by class: %w", err)
	}
	return nil
}

// ListHistoryByClass lists class history by effective date.
func (s *SQLXAssetStore) ListHistoryByClass(ctx context.Context, accountID, classID uuid.UUID, limit int, ascending bool) ([]*assets.HistoryEntry, error) {
	order := "DESC"
	if ascending {
		order = "ASC"
	}
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE account_id = ? AND class_id = ?
		ORDER BY effective_date %s, created_at %s
	`, s.historyTable, order, order)
	args := []any{accountID, classID}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	query = s.db.Rebind(query)
	executor := s.db.GetExecutor(ctx)

	rows, err := executor.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlx_asset_store: list history: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	out := make([]*assets.HistoryEntry, 0)
	for rows.Next() {
		var row assets.HistoryEntry
		if err := rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("sqlx_asset_store: scan history row: %w", err)
		}
		out = append(out, &row)
	}
	return out, nil
}

// ListHistoryForAccount lists account-scoped history by effective date.
func (s *SQLXAssetStore) ListHistoryForAccount(ctx context.Context, accountID uuid.UUID, limit int, ascending bool) ([]*assets.HistoryEntry, error) {
	order := "DESC"
	if ascending {
		order = "ASC"
	}
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE account_id = ?
		ORDER BY effective_date %s, created_at %s
	`, s.historyTable, order, order)
	args := []any{accountID}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	query = s.db.Rebind(query)

	executor := s.db.GetExecutor(ctx)
	rows, err := executor.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlx_asset_store: list account history: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	out := make([]*assets.HistoryEntry, 0)
	for rows.Next() {
		var row assets.HistoryEntry
		if err := rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("sqlx_asset_store: scan account history row: %w", err)
		}
		out = append(out, &row)
	}
	return out, nil
}

// DeleteSnapshotsByAccount deletes all snapshot points for an account.
func (s *SQLXAssetStore) DeleteSnapshotsByAccount(ctx context.Context, accountID uuid.UUID) error {
	query := s.db.Rebind(fmt.Sprintf(`
		DELETE FROM %s
		WHERE account_id = ?
	`, s.snapshotsTable))
	executor := s.db.GetExecutor(ctx)
	if _, err := executor.ExecContext(ctx, query, accountID); err != nil {
		return fmt.Errorf("sqlx_asset_store: delete snapshots by account: %w", err)
	}
	return nil
}

// UpsertSnapshots creates or updates account snapshot points.
func (s *SQLXAssetStore) UpsertSnapshots(ctx context.Context, snapshots []*assets.Snapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, occurred_at, total_worth, created_at, updated_at
		) VALUES (
			:id, :account_id, :occurred_at, :total_worth, :created_at, :updated_at
		)
		ON CONFLICT (account_id, occurred_at) DO UPDATE SET
			total_worth = EXCLUDED.total_worth,
			updated_at = EXCLUDED.updated_at
	`, s.snapshotsTable)
	executor := s.db.GetExecutor(ctx)
	for _, snapshot := range snapshots {
		if snapshot == nil {
			continue
		}
		if snapshot.CreatedAt.IsZero() {
			snapshot.CreatedAt = time.Now().UTC()
		}
		if snapshot.ID == uuid.Nil {
			snapshot.ID = uuid.New()
		}
		snapshot.OccurredAt = time.Date(
			snapshot.OccurredAt.UTC().Year(),
			snapshot.OccurredAt.UTC().Month(),
			snapshot.OccurredAt.UTC().Day(),
			0, 0, 0, 0,
			time.UTC,
		)
		snapshot.UpdatedAt = time.Now().UTC()
		if _, err := sqlx.NamedExecContext(ctx, executor, query, snapshot); err != nil {
			return fmt.Errorf("sqlx_asset_store: upsert snapshots: %w", err)
		}
	}
	return nil
}

// ListSnapshotsForAccount lists snapshots for an account in ascending day order.
func (s *SQLXAssetStore) ListSnapshotsForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	from, to *time.Time,
) ([]*assets.Snapshot, error) {
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE account_id = ?
	`, s.snapshotsTable)
	args := []any{accountID}
	if from != nil {
		query += " AND occurred_at >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND occurred_at <= ?"
		args = append(args, *to)
	}
	query += " ORDER BY occurred_at ASC, id ASC"
	query = s.db.Rebind(query)

	executor := s.db.GetExecutor(ctx)
	rows := make([]*assets.Snapshot, 0)
	if err := sqlx.SelectContext(ctx, executor, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("sqlx_asset_store: list snapshots for account: %w", err)
	}
	return rows, nil
}

func isAssetClassDuplicate(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && strings.Contains(strings.ToLower(pqErr.Constraint), "uq_asset_classes_account_name")
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") && strings.Contains(msg, "account_id") && strings.Contains(msg, "name")
}

func isAssetItemDuplicate(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && strings.Contains(strings.ToLower(pqErr.Constraint), "uq_asset_items_class_name")
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") && strings.Contains(msg, "class_id") && strings.Contains(msg, "name")
}
