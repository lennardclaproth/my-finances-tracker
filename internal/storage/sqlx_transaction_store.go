package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/transaction"
	"github.com/lib/pq"
)

type SQLXTransactionStore struct {
	db *DB
}

func NewSQLXTransactionStore(db *DB) *SQLXTransactionStore {
	return &SQLXTransactionStore{db: db}
}

func parseRows(rows *sqlx.Rows) ([]*transaction.Transaction, error) {
	var transactions []*transaction.Transaction
	for rows.Next() {
		var tx transaction.Transaction
		if err := rows.StructScan(&tx); err != nil {
			return nil, fmt.Errorf("sqlx_transaction_store: failed to scan transaction record: %w", err)
		}
		transactions = append(transactions, &tx)
	}
	return transactions, nil
}

func (s *SQLXTransactionStore) Create(ctx context.Context, tx *transaction.Transaction) error {
	query := fmt.Sprintf(`
        INSERT INTO %s (
            id, description, note, source, amount_cents,
            direction, date, checksum, created_at, updated_at, tag,
			row_number, ignored, import_id
        ) VALUES (
            :id, :description, :note, :source, :amount_cents,
            :direction, :date, :checksum, :created_at, :updated_at, :tag,
			:row_number, :ignored, :import_id
        )
    `, TableTransactions)
	executor := s.db.GetExecutor(ctx)
	namedQuery, args, err := sqlx.Named(query, tx)
	if err != nil {
		return fmt.Errorf("sqlx_transaction_store: failed to bind named params: %w", err)
	}
	namedQuery, args, err = sqlx.In(namedQuery, args...)
	if err != nil {
		return fmt.Errorf("sqlx_transaction_store: failed to expand query: %w", err)
	}
	namedQuery = sqlx.Rebind(sqlx.DOLLAR, namedQuery) // or executor.Rebind if you have it
	_, err = executor.ExecContext(ctx, namedQuery, args...)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" && pqErr.Constraint == "transactions_checksum_key" {
				return transaction.ErrDuplicateTransaction
			}
		}
		return fmt.Errorf("sqlx_transaction_store: failed to save transaction: %w", err)
	}
	return nil
}

func (s *SQLXTransactionStore) FetchUntagged(ctx context.Context, page, pageSize int) ([]*transaction.Transaction, error) {
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT * FROM %s WHERE tag IS NULL OR tag = '' ORDER BY date DESC LIMIT $1 OFFSET $2`, TableTransactions)

	executor := s.db.GetExecutor(ctx)
	rows, err := executor.QueryxContext(ctx, query, pageSize, offset)

	if err == sql.ErrNoRows {
		return []*transaction.Transaction{}, nil // return empty slice if no transactions found
	}
	if err != nil {
		return nil, fmt.Errorf("sqlx_transaction_store: failed to fetch untagged transactions: %w", err)
	}
	defer rows.Close()
	transactions, err := parseRows(rows)
	if err != nil {
		return nil, fmt.Errorf("sqlx_transaction_store: failed to parse transaction rows: %w", err)
	}
	return transactions, nil
}

func (s *SQLXTransactionStore) Tag(ctx context.Context, id uuid.UUID, tag string) error {
	query := fmt.Sprintf(`UPDATE %s SET tag = $1, updated_at = NOW() WHERE id = $2`, TableTransactions)
	executor := s.db.GetExecutor(ctx)
	_, err := executor.ExecContext(ctx, query, tag, id)
	if err != nil {
		return fmt.Errorf("sqlx_transaction_store: failed to tag transaction: %w", err)
	}
	return nil
}
