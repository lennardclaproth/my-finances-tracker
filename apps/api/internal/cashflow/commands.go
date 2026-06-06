package cashflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type Commands struct {
	cs  CommandStore
	qs  QueryStore
	aec accountExistenceChecker
}

type accountExistenceChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

// CommandStore persists cashflow transaction mutations.
type CommandStore interface {
	CreateTransactions(ctx context.Context, txs []*Transaction) (int, error)
	UpdateTagByIDs(ctx context.Context, ids []uuid.UUID, tag string) (int, error)
	UpdateTagByFilter(ctx context.Context, filters TransactionFilters, tag string) (int, error)
	UpdateIgnoredByIDs(ctx context.Context, ids []uuid.UUID, ignored bool) (int, error)
	UpdateIgnoredByFilter(ctx context.Context, filters TransactionFilters, ignored bool) (int, error)
}

const (
	manualCashflowBatchMaxSize = 100
)

// NewCommands creates cashflow write-side use cases.
func NewCommands(
	cStore CommandStore,
	qStore QueryStore,
	aec accountExistenceChecker,
) *Commands {
	return &Commands{
		cs:  cStore,
		qs:  qStore,
		aec: aec,
	}
}

type TransactionData struct {
	Description string
	Note        string
	Source      string
	Direction   CashFlowDirection
	Amount      money.Price
	Date        time.Time
	AccountType *AccountType
	Tag         string
	// RowNumber is the source row number (e.g. CSV line) and feeds the dedup
	// checksum. Manual entries leave it zero and receive a generated row number.
	RowNumber int
}

// NewTransactionData validates and maps manual cashflow input into transaction data.
func NewTransactionData(
	dateRaw, amountRaw, typeRaw, descriptionRaw, noteRaw, tagRaw, vendorRaw, source string,
	rowNumber *int,
) (TransactionData, error) {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(dateRaw))
	if err != nil {
		return TransactionData{}, ErrManualCashflowInvalidDate
	}

	amount, err := money.ParsePrice(amountRaw)
	if err != nil {
		return TransactionData{}, err
	}

	direction, err := ParseDirection(typeRaw)
	if err != nil {
		return TransactionData{}, err
	}

	description := strings.TrimSpace(descriptionRaw)
	if description == "" {
		return TransactionData{}, ErrManualCashflowDescriptionRequired
	}
	note := strings.TrimSpace(noteRaw)
	if note == "" {
		return TransactionData{}, ErrManualCashflowNoteRequired
	}
	tag := strings.TrimSpace(tagRaw)
	if tag == "" {
		return TransactionData{}, ErrManualCashflowTagRequired
	}

	if vendorName := strings.TrimSpace(vendorRaw); vendorName != "" {
		source = "manual:" + vendorName
	}

	return TransactionData{
		Description: description,
		Note:        note,
		Source:      source,
		Direction:   *direction,
		Amount:      amount,
		Date:        date.UTC(),
		Tag:         tag,
	}, nil
}

// CreateManyResult reports the outcome of a batch cashflow create.
type CreateManyResult struct {
	// Transactions are the transactions built from the input. On a successful
	// create with no duplicates these are the persisted rows.
	Transactions []*Transaction
	// Imported is the number of rows newly inserted.
	Imported int
	// Duplicates is the number of rows skipped because they already existed.
	Duplicates int
}

// CreateMany validates a batch of cashflow transactions and persists them with a single
// bulk insert, skipping and counting rows that already exist. It serves both manual
// entry and CSV imports: manual callers pass a nil import ID and are capped at
// manualCashflowBatchMaxSize; import callers pass the import ID and set RowNumber on
// each TransactionData. Rows without a row number fall back to a generated manual one.
func (c *Commands) CreateMany(ctx context.Context, accID uuid.UUID, impID *uuid.UUID, transactions []TransactionData) (CreateManyResult, error) {
	if len(transactions) == 0 {
		return CreateManyResult{}, fmt.Errorf("create many: %w", ErrTransactionsRequired)
	}
	if impID == nil && len(transactions) > manualCashflowBatchMaxSize {
		return CreateManyResult{}, fmt.Errorf("create many: %w", ErrTransactionLimitExceeded)
	}

	exists, err := c.aec.Exists(ctx, accID)
	if err != nil {
		return CreateManyResult{}, fmt.Errorf("create many: failed to check if account exists: %w", err)
	}
	if !exists {
		return CreateManyResult{}, fmt.Errorf("create many: %w", ErrAccountNotFound)
	}

	txs := make([]*Transaction, 0, len(transactions))
	for i, row := range transactions {
		rowNumber := row.RowNumber
		if rowNumber == 0 {
			rowNumber = manualCashflowRowNumber(i)
		}
		tx, err := NewTransaction(
			row.Description,
			row.Note,
			row.Source,
			row.Tag,
			row.Direction,
			row.Amount,
			row.Date,
			rowNumber,
			impID,
			row.AccountType,
			accID,
		)
		if err != nil {
			return CreateManyResult{}, fmt.Errorf("create many: build transaction: %w", err)
		}
		txs = append(txs, tx)
	}

	inserted, err := c.cs.CreateTransactions(ctx, txs)
	if err != nil {
		return CreateManyResult{}, fmt.Errorf("create many: %w", err)
	}

	return CreateManyResult{
		Transactions: txs,
		Imported:     inserted,
		Duplicates:   len(txs) - inserted,
	}, nil
}

type ExecutionMode string

const (
	TagByFilterModeSync  ExecutionMode = "sync"
	TagByFilterModeAsync ExecutionMode = "async"
)

type BulkTagResult struct {
	Mode         ExecutionMode
	UpdatedCount int
	TotalMatched int
}

type TagByFilterCommand struct {
	Tag       string
	AccountID *uuid.UUID
	Filters   TransactionFilters
}

// TagByID applies a tag to one cashflow transaction.
func (c *Commands) TagByID(ctx context.Context, id uuid.UUID, tag string) error {
	_, err := c.TagByIDs(ctx, []uuid.UUID{id}, tag)
	if err != nil {
		return fmt.Errorf("cashflow tag by id: %w", err)
	}
	return nil
}

// TagByIDs applies a tag to the selected cashflow transactions.
func (c *Commands) TagByIDs(ctx context.Context, ids []uuid.UUID, tag string) (int, error) {
	updated, err := c.cs.UpdateTagByIDs(ctx, ids, tag)
	if err != nil {
		return 0, fmt.Errorf("cashflow tag by ids: %w", err)
	}
	return updated, nil
}

// TagByFilter applies or schedules tagging based on total matched rows and async policy.
func (c *Commands) TagByFilter(ctx context.Context, tag string, accID uuid.UUID, filters TransactionFilters) (BulkTagResult, error) {
	total, err := c.qs.CountByFilter(ctx, filters)
	if err != nil {
		return BulkTagResult{}, fmt.Errorf("cashflow tag by filter count: %w", err)
	}
	// TODO: implement async tagging
	updated, err := c.cs.UpdateTagByFilter(ctx, filters, tag)
	if err != nil {
		return BulkTagResult{}, fmt.Errorf("cashflow tag by filter update: %w", err)
	}
	return BulkTagResult{
		Mode:         TagByFilterModeSync,
		UpdatedCount: updated,
		TotalMatched: total,
	}, nil
}

// IgnoreByIDs sets the ignored flag for the selected cashflow transactions.
func (c *Commands) IgnoreByIDs(ctx context.Context, ids []uuid.UUID, ignored bool) (int, error) {
	updated, err := c.cs.UpdateIgnoredByIDs(ctx, ids, ignored)
	if err != nil {
		return 0, fmt.Errorf("cashflow ignore by ids: %w", err)
	}
	return updated, nil
}

// IgnoreByFilter sets the ignored flag for cashflow transactions matching the supplied filters.
func (c *Commands) IgnoreByFilter(ctx context.Context, filters TransactionFilters, ignored bool) (int, error) {
	updated, err := c.cs.UpdateIgnoredByFilter(ctx, filters, ignored)
	if err != nil {
		return 0, fmt.Errorf("cashflow ignore by filter: %w", err)
	}
	return updated, nil
}
