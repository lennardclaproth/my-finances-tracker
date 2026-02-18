package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lib/pq"
)

type SQLXPortfolioTransactionStore struct {
	db        *DB
	tableName string
}

func NewSQLXPortfolioTransactionStore(db *DB) *SQLXPortfolioTransactionStore {
	return &SQLXPortfolioTransactionStore{
		db:        db,
		tableName: qualifyPortfolioTransactionsTable(db),
	}
}

func qualifyPortfolioTransactionsTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "portfolio_transactions"
	}
	return fmt.Sprintf("%s.%s", SchemaPortfolio, TableTransactions)
}

func (s *SQLXPortfolioTransactionStore) Create(ctx context.Context, tx *portfolio.Transaction) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, import_id, source, occurred_at, value_date, listing_id, isin, symbol,
			type, quantity, price_cents, amount_cents, checksum, raw_ref, row_number, created_at, updated_at
		) VALUES (
			:id, :account_id, :import_id, :source, :occurred_at, :value_date, :listing_id, :isin, :symbol,
			:type, :quantity, :price_cents, :amount_cents, :checksum, :raw_ref, :row_number, :created_at, :updated_at
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

func isPortfolioDuplicate(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && strings.Contains(strings.ToLower(pqErr.Constraint), "checksum")
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") && strings.Contains(msg, "checksum")
}
