package cashflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Commands struct {
	cs  commandStore
	eac accountExistenceChecker
	p   publisher
}

type commandStore interface {
	CreateTransaction(ctx context.Context, tx *Transaction) error
	CountByFilter(ctx context.Context, filters TransactionFilters) (int, error)
	UpdateTagByFilter(ctx context.Context, filters TransactionFilters, tag string) (int, error)
}

type accountExistenceChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type publisher interface {
	Publish(ctx context.Context, payload any) error
}

const (
	manualCashflowBatchMaxSize = 100
)

// CreateMany validates and persists manual cashflow transactions for one account.
func (c *Commands) CreateMany(ctx context.Context, accID, impID uuid.UUID, transactions []TransactionData) ([]*Transaction, error) {
	if len(transactions) == 0 {
		return nil, fmt.Errorf("create many: %w", ErrTransactionsRequired)
	}
	if len(transactions) > manualCashflowBatchMaxSize {
		return nil, fmt.Errorf("create many: %w", ErrTransactionLimitExceeded)
	}

	exists, err := c.eac.Exists(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("create many: failed to check account existence: %w", err)
	}
	if exists == false {
		return nil, fmt.Errorf("create many: %w", ErrAccountNotFound)
	}

	out := make([]*Transaction, 0, len(transactions))
	for i, row := range transactions {
		rowNumber := manualCashflowRowNumber(i)
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
			nil,
			&accID,
		)
		if err != nil {
			return nil, fmt.Errorf("manual cashflow create: build transaction: %w", err)
		}
		tx.Tag = row.Tag
		// TODO: consider batch insert if this becomes a bottleneck
		if err := c.cs.CreateTransaction(ctx, tx); err != nil {
			return nil, err
		}
		out = append(out, tx)
	}
	//TODO: Publish transactions imported event

	//TODO: change return type
	return out, nil
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

// TagByFilter applies or schedules tagging based on total matched rows and async policy.
func (c *Commands) TagByFilter(ctx context.Context, tag string, accID uuid.UUID, filters TransactionFilters) (BulkTagResult, error) {
	total, err := c.cs.CountByFilter(ctx, filters)
	if err != nil {
		return BulkTagResult{}, err
	}
	// TODO: Refactor this into a nice pub sub flow
	if total > manualCashflowBatchMaxSize && c.p != nil && accID != uuid.Nil {
		if err := c.p.Publish(ctx, TagByFilterCommand{
			Tag:       tag,
			AccountID: &accID,
			Filters:   filters,
		}); err != nil {
			return BulkTagResult{}, err
		}
		return BulkTagResult{
			Mode:         TagByFilterModeAsync,
			UpdatedCount: 0,
			TotalMatched: total,
		}, nil
	}

	updated, err := c.cs.UpdateTagByFilter(ctx, filters, tag)
	if err != nil {
		return BulkTagResult{}, err
	}
	return BulkTagResult{
		Mode:         TagByFilterModeSync,
		UpdatedCount: updated,
		TotalMatched: total,
	}, nil
}
