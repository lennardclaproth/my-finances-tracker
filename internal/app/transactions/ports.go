package transactions

import (
	"context"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/domain"
)

type TransactionCreator interface {
	Create(ctx context.Context, tx *domain.Transaction) error
}

type UntaggedTransactionFetcher interface {
	FetchUntagged(ctx context.Context, page, pageSize int) ([]*domain.Transaction, error)
}

type TransactionTagger interface {
	Tag(ctx context.Context, id uuid.UUID, tag string) error
}

type VendorFetcher interface {
	FetchById(ctx context.Context, id uuid.UUID) (*domain.Vendor, error)
}

type CsvParser interface {
	ParseRow(row []string, rowNumber int, importId uuid.UUID) (TransactionData, error)
	ParseHeader(header []string) (error)
}
