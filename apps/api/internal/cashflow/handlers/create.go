package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
)

const (
	manualCashflowBatchMaxSize = 100
)

var (
	// ErrCreateAccountNotFound indicates the account does not exist.
	ErrCreateAccountNotFound = fmt.Errorf("account not found")
	// ErrCreateImportNotFound indicates the import does not exist.
	ErrCreateImportNotFound = fmt.Errorf("import not found")
	// ErrCreateTransactionsRequired indicates no transactions were supplied.
	ErrCreateTransactionsRequired = fmt.Errorf("transactions is required")
	// ErrCreateTransactionLimitExceeded indicates the request is above the supported batch size.
	ErrCreateTransactionLimitExceeded = fmt.Errorf("transactions exceeds maximum batch size of 100")
)

// CreateCommand describes a bulk create request for one account.
type CreateCommand struct {
	AccountID    uuid.UUID
	Transactions []cashflow.TransactionData
}

// ManualCashflowCreateResult contains persisted rows.
type ManualCashflowCreateResult struct {
	Transactions []*cashflow.Transaction
}

type accountFetcher interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type importFetcher interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type transactionCreator interface {
	Create(ctx context.Context, tx *cashflow.Transaction) error
}

// CreateHandler creates manual cashflow transactions in bulk.
type CreateHandler struct {
	accounts     accountFetcher
	imports      importFetcher
	transactions transactionCreator
}

// NewCreateHandler constructs a manual cashflow transaction create service.
func NewCreateHandler(
	accounts accountFetcher,
	imports importFetcher,
	transactions transactionCreator,
) *CreateHandler {
	return &CreateHandler{
		accounts:     accounts,
		imports:      imports,
		transactions: transactions,
	}
}

// CreateMany validates and persists manual cashflow transactions for one account.
func (s *CreateHandler) CreateMany(ctx context.Context, accID, impID uuid.UUID, transactions []cashflow.TransactionData) ([]*cashflow.Transaction, error) {
	if len(transactions) == 0 {
		return nil, ErrCreateTransactionsRequired
	}
	if len(transactions) > manualCashflowBatchMaxSize {
		return nil, ErrCreateTransactionLimitExceeded
	}

	exists, err := s.accounts.Exists(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("handlers: failed to create transactions: %w", err)
	}
	if exists == false {
		return nil, fmt.Errorf("handlers: failed to create transactions: %w", ErrCreateAccountNotFound)
	}

	exists, err = s.imports.Exists(ctx, impID)
	if err != nil {
		return nil, fmt.Errorf("handlers: failed to create transactions: %w", err)
	}
	if exists == false {
		return nil, fmt.Errorf("handlers: failed to create transactions: %w", ErrCreateImportNotFound)
	}

	out := make([]*cashflow.Transaction, 0, len(transactions))
	for i, row := range transactions {
		rowNumber := manualCashflowRowNumber(i)
		tx, err := cashflow.NewTransaction(
			row.Description,
			row.Note,
			row.Source,
			row.Tag,
			row.Direction,
			row.Amount,
			row.Date,
			rowNumber,
			impID,
			nil,
			&accID,
		)
		if err != nil {
			return nil, fmt.Errorf("manual cashflow create: build transaction: %w", err)
		}
		tx.Tag = row.Tag
		// TODO: consider batch insert if this becomes a bottleneck
		if err := s.transactions.Create(ctx, tx); err != nil {
			return nil, err
		}
		out = append(out, tx)
	}
	//TODO: Publish transactions imported event

	//TODO: change return type
	return out, nil
}

func manualCashflowRowNumber(idx int) int {
	const maxRowNumber = 2_147_483_647
	row := int(time.Now().UnixNano()%2_000_000_000) + idx + 1
	if row <= 0 {
		return idx + 1
	}
	if row > maxRowNumber {
		return (row % maxRowNumber) + 1
	}
	return row
}
