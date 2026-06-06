package cashflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
)

// Queries exposes read-side cashflow use cases.
type Queries struct {
	qs QueryStore
}

// QueryStore reads cashflow transactions and analytics.
type QueryStore interface {
	GetMonthlyAnalytics(ctx context.Context, filter AnalyticsFilter) ([]MonthlyAnalyticsPoint, error)
	GetTagDistribution(ctx context.Context, filter AnalyticsFilter) (*TagDistribution, error)
	ListTransactions(ctx context.Context, query TransactionListQuery) (*TransactionListResult, error)
	CountByFilter(ctx context.Context, filters TransactionFilters) (int, error)
}

// NewQueries creates cashflow read-side use cases.
func NewQueries(qs QueryStore) *Queries {
	return &Queries{qs: qs}
}

// AnalyticsFilter describes the filters used for cashflow analytics queries.
type AnalyticsFilter struct {
	From           *time.Time
	To             *time.Time
	IncludeIgnored bool
}

// MonthlyAnalyticsPoint contains cashflow totals for a single calendar month.
type MonthlyAnalyticsPoint struct {
	Month         time.Time
	IncomingCents int64
	OutgoingCents int64
	NetCents      int64
}

// TagDistributionEntry contains the total amount for a single cashflow tag.
type TagDistributionEntry struct {
	Tag        string
	TotalCents int64
}

// TagDistribution contains cashflow tag totals grouped by direction.
type TagDistribution struct {
	Combined []TagDistributionEntry
	Incoming []TagDistributionEntry
	Outgoing []TagDistributionEntry
}

// MonthlyAnalytics returns incoming, outgoing, and net totals grouped by month.
func (q *Queries) MonthlyAnalytics(ctx context.Context, filter AnalyticsFilter) ([]MonthlyAnalyticsPoint, error) {
	points, err := q.qs.GetMonthlyAnalytics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("cashflow monthly analytics: %w", err)
	}
	return points, nil
}

// TagDistribution returns tag totals for combined, incoming, and outgoing cashflows.
func (q *Queries) TagDistribution(ctx context.Context, filter AnalyticsFilter) (*TagDistribution, error) {
	dist, err := q.qs.GetTagDistribution(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("cashflow tag distribution: %w", err)
	}
	if dist == nil {
		return &TagDistribution{}, nil
	}
	return dist, nil
}

const (
	// TransactionSortFieldDate sorts cashflow transactions by transaction date.
	TransactionSortFieldDate sorting.Field = "date"
	// TransactionSortFieldDescription sorts cashflow transactions by description.
	TransactionSortFieldDescription sorting.Field = "description"
	// TransactionSortFieldNote sorts cashflow transactions by note.
	TransactionSortFieldNote sorting.Field = "note"
	// TransactionSortFieldTag sorts cashflow transactions by tag.
	TransactionSortFieldTag sorting.Field = "tag"
	// TransactionSortFieldSource sorts cashflow transactions by source.
	TransactionSortFieldSource sorting.Field = "source"
	// TransactionSortFieldAmount sorts cashflow transactions by amount.
	TransactionSortFieldAmount sorting.Field = "amount"
)

// TransactionListQuery describes filters, sorting, and pagination for cashflow transactions.
type TransactionListQuery struct {
	Limit       int
	Offset      int
	Sort        sorting.Sort
	Q           string
	Description string
	Note        string
	Source      string
	Direction   *CashFlowDirection
	Tags        []string
	Untagged    bool
	HideIgnored bool
	From        *time.Time
	To          *time.Time
}

// TransactionListResult contains a page of cashflow transactions and the total match count.
type TransactionListResult struct {
	Total        int
	Transactions []*Transaction
}

// ListTransactions returns a filtered and paginated cashflow transaction page.
func (q *Queries) ListTransactions(ctx context.Context, query TransactionListQuery) (*TransactionListResult, error) {
	result, err := q.qs.ListTransactions(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("cashflow transactions: %w", err)
	}
	if result == nil {
		return &TransactionListResult{Transactions: []*Transaction{}}, nil
	}
	if result.Transactions == nil {
		result.Transactions = []*Transaction{}
	}
	return result, nil
}

// ParseTransactionSort converts raw transaction sort inputs into a validated
// sort value using the supported cashflow transaction fields.
func ParseTransactionSort(fieldRaw, directionRaw string) (sorting.Sort, error) {
	return sorting.ParseSort(
		fieldRaw,
		directionRaw,
		sorting.DESC,
		parseTransactionSortField,
	)
}

func parseTransactionSortField(raw string) (sorting.Field, error) {
	switch sorting.Field(strings.ToLower(strings.TrimSpace(raw))) {
	case "":
		return TransactionSortFieldDate, nil
	case TransactionSortFieldDate:
		return TransactionSortFieldDate, nil
	case TransactionSortFieldDescription:
		return TransactionSortFieldDescription, nil
	case TransactionSortFieldNote:
		return TransactionSortFieldNote, nil
	case TransactionSortFieldTag:
		return TransactionSortFieldTag, nil
	case TransactionSortFieldSource:
		return TransactionSortFieldSource, nil
	case TransactionSortFieldAmount:
		return TransactionSortFieldAmount, nil
	default:
		return "", fmt.Errorf("sort_by must be one of: date, description, note, tag, source, amount")
	}
}
