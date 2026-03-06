package portfolio

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type TransactionSortBy string

const (
	TransactionSortByDate TransactionSortBy = "date"
)

type TransactionSortOrder string

const (
	TransactionSortOrderAsc  TransactionSortOrder = "asc"
	TransactionSortOrderDesc TransactionSortOrder = "desc"
)

type TransactionListQuery struct {
	AccountID uuid.UUID
	From      *time.Time
	To        *time.Time
	Limit     int
	Offset    int
	SortBy    TransactionSortBy
	SortOrder TransactionSortOrder
	Q         string
	Type      *TransactionType
	Origin    *TransactionOrigin
	Source    string
	Listing   string
}

type TransactionListResult struct {
	Total        int
	Transactions []TransactionWithListingID
}

func NormalizeTransactionSortBy(raw string) TransactionSortBy {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(TransactionSortByDate):
		return TransactionSortByDate
	default:
		return TransactionSortByDate
	}
}

func NormalizeTransactionSortOrder(raw string) TransactionSortOrder {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(TransactionSortOrderAsc):
		return TransactionSortOrderAsc
	case string(TransactionSortOrderDesc):
		return TransactionSortOrderDesc
	default:
		return TransactionSortOrderDesc
	}
}
