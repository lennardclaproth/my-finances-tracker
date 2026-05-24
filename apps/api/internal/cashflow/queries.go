package cashflow

import (
	"context"
	"fmt"
	"time"
)

// Queries exposes read-side cashflow use cases.
type Queries struct {
	qs queryStore
}

type queryStore interface {
	GetMonthlyAnalytics(ctx context.Context, filter AnalyticsFilter) ([]MonthlyAnalyticsPoint, error)
}

// NewQueries creates cashflow read-side use cases.
func NewQueries(qs queryStore) *Queries {
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

// MonthlyAnalytics returns incoming, outgoing, and net totals grouped by month.
func (q *Queries) MonthlyAnalytics(ctx context.Context, filter AnalyticsFilter) ([]MonthlyAnalyticsPoint, error) {
	points, err := q.qs.GetMonthlyAnalytics(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("cashflow monthly analytics: %w", err)
	}
	return points, nil
}
