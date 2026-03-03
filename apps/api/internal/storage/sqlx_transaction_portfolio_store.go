package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lib/pq"
)

type SQLXPortfolioTransactionStore struct {
	db             *DB
	tableName      string
	positionsTable string
}

func NewSQLXPortfolioTransactionStore(db *DB) *SQLXPortfolioTransactionStore {
	return &SQLXPortfolioTransactionStore{
		db:             db,
		tableName:      qualifyPortfolioTransactionsTable(db),
		positionsTable: qualifyTable(db, SchemaPortfolio, TablePositions),
	}
}

var _ portfolio.TransactionStore = (*SQLXPortfolioTransactionStore)(nil)

func qualifyPortfolioTransactionsTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "portfolio_transactions"
	}
	return fmt.Sprintf("%s.%s", SchemaPortfolio, TableTransactions)
}

func (s *SQLXPortfolioTransactionStore) Create(ctx context.Context, tx *portfolio.Transaction) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, import_id, origin, source, occurred_at, position_id, isin, symbol, description,
			type, quantity, unit_price, amount_cents, checksum, row_number, created_at, updated_at
		) VALUES (
			:id, :account_id, :import_id, :origin, :source, :occurred_at, :position_id, :isin, :symbol, :description,
			:type, :quantity, :unit_price, :amount_cents, :checksum, :row_number, :created_at, :updated_at
		)
	`, s.tableName)
	executor := s.db.GetExecutor(ctx)
	namedQuery, args, err := sqlx.Named(query, tx)
	if err != nil {
		return fmt.Errorf("sqlx_portfolio_transaction_store: bind named params: %w", err)
	}
	namedQuery = sqlx.Rebind(sqlx.BindType(s.db.DriverName()), namedQuery)
	_, err = executor.ExecContext(ctx, namedQuery, args...)
	if err != nil {
		if isPortfolioDuplicate(err) {
			return portfolio.ErrDuplicateTransaction
		}
		return fmt.Errorf("sqlx_portfolio_transaction_store: create transaction: %w", err)
	}
	return nil
}

func (s *SQLXPortfolioTransactionStore) GetASC(ctx context.Context, accID uuid.UUID) ([]portfolio.Transaction, error) {
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE account_id = ?
		ORDER BY occurred_at ASC, row_number ASC
	`, s.tableName)
	query = s.db.Rebind(query)
	executor := s.db.GetExecutor(ctx)
	var txs []portfolio.Transaction
	if err := sqlx.SelectContext(ctx, executor, &txs, query, accID); err != nil {
		return nil, fmt.Errorf("sqlx_portfolio_transaction_store: get asc: %w", err)
	}
	return txs, nil
}

func (s *SQLXPortfolioTransactionStore) UpdatePositions(ctx context.Context, transactions []portfolio.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}
	query := fmt.Sprintf(`
		UPDATE %s
		SET position_id = ?, updated_at = ?
		WHERE id = ?
	`, s.tableName)
	query = s.db.Rebind(query)
	executor := s.db.GetExecutor(ctx)
	now := time.Now().UTC()
	for i := range transactions {
		tx := transactions[i]
		if tx.PositionID == nil {
			continue
		}
		if _, err := executor.ExecContext(ctx, query, *tx.PositionID, now, tx.ID); err != nil {
			return fmt.Errorf("sqlx_portfolio_transaction_store: update position mapping: %w", err)
		}
	}
	return nil
}

func (s *SQLXPortfolioTransactionStore) GetByPositionID(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]portfolio.Transaction, error) {
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE position_id = ?
	`, s.tableName)
	args := []any{positionID}
	if from != nil {
		query += " AND occurred_at >= ?"
		args = append(args, *from)
	}
	query += " ORDER BY occurred_at ASC, row_number ASC"
	query = s.db.Rebind(query)
	executor := s.db.GetExecutor(ctx)
	var txs []portfolio.Transaction
	if err := sqlx.SelectContext(ctx, executor, &txs, query, args...); err != nil {
		return nil, fmt.Errorf("sqlx_portfolio_transaction_store: get by position: %w", err)
	}
	return txs, nil
}

func (s *SQLXPortfolioTransactionStore) FetchForAccount(
	ctx context.Context,
	accID uuid.UUID,
	from, to *time.Time,
) ([]portfolio.TransactionWithListingID, error) {
	query := fmt.Sprintf(`
		SELECT
			t.*,
			p.listing_id AS listing_id
		FROM %s t
		LEFT JOIN %s p ON p.id = t.position_id
		WHERE t.account_id = ?
	`, s.tableName, s.positionsTable)
	args := []any{accID}
	if from != nil {
		query += " AND t.occurred_at >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND t.occurred_at <= ?"
		args = append(args, *to)
	}
	query += " ORDER BY t.occurred_at DESC, t.created_at DESC, t.id DESC"
	query = s.db.Rebind(query)

	executor := s.db.GetExecutor(ctx)
	var txs []portfolio.TransactionWithListingID
	if err := sqlx.SelectContext(ctx, executor, &txs, query, args...); err != nil {
		return nil, fmt.Errorf("sqlx_portfolio_transaction_store: fetch for account: %w", err)
	}
	return txs, nil
}

func isPortfolioDuplicate(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && strings.Contains(strings.ToLower(pqErr.Constraint), "checksum")
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") && strings.Contains(msg, "checksum")
}
