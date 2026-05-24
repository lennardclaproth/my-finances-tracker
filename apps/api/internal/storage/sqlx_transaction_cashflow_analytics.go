package storage

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
)

type CashflowAnalyticsQuery struct {
	From           *time.Time
	To             *time.Time
	IncludeIgnored bool
}

type CashflowMonthlyAnalyticsPoint struct {
	Month         time.Time
	IncomingCents int64
	OutgoingCents int64
	NetCents      int64
}

type CashflowTagDistributionEntry struct {
	Tag        string
	TotalCents int64
}

type CashflowTagDistribution struct {
	Combined []CashflowTagDistributionEntry
	Incoming []CashflowTagDistributionEntry
	Outgoing []CashflowTagDistributionEntry
}

func (s *SQLXBankTransactionStore) FetchMonthlyAnalytics(ctx context.Context, query CashflowAnalyticsQuery) ([]CashflowMonthlyAnalyticsPoint, error) {
	points, err := s.FetchCashflowMonthlyAnalytics(ctx, cashflow.AnalyticsFilter{
		From:           query.From,
		To:             query.To,
		IncludeIgnored: query.IncludeIgnored,
	})
	if err != nil {
		return nil, err
	}

	out := make([]CashflowMonthlyAnalyticsPoint, 0, len(points))
	for _, point := range points {
		out = append(out, CashflowMonthlyAnalyticsPoint{
			Month:         point.Month,
			IncomingCents: point.IncomingCents,
			OutgoingCents: point.OutgoingCents,
			NetCents:      point.NetCents,
		})
	}
	return out, nil
}

// FetchCashflowMonthlyAnalytics returns monthly cashflow totals for the read-side cashflow use case.
func (s *SQLXBankTransactionStore) FetchCashflowMonthlyAnalytics(ctx context.Context, filter cashflow.AnalyticsFilter) ([]cashflow.MonthlyAnalyticsPoint, error) {
	monthExpr := "TO_CHAR(DATE_TRUNC('month', date), 'YYYY-MM-01')"
	if s.db.DriverName() == string(Sqlite) {
		monthExpr = "COALESCE(STRFTIME('%Y-%m-01', date), SUBSTR(CAST(date AS TEXT), 1, 7) || '-01')"
	}

	whereClause, args := buildCashflowAnalyticsWhereClause(CashflowAnalyticsQuery{
		From:           filter.From,
		To:             filter.To,
		IncludeIgnored: filter.IncludeIgnored,
	})
	rawQuery := fmt.Sprintf(`
		SELECT
			%s AS month_start,
			COALESCE(SUM(CASE WHEN direction = 'in' THEN amount_cents ELSE 0 END), 0) AS incoming_cents,
			COALESCE(SUM(CASE WHEN direction = 'out' THEN amount_cents ELSE 0 END), 0) AS outgoing_cents
		FROM %s
		%s
		GROUP BY 1
		ORDER BY 1 ASC
	`, monthExpr, s.tableName, whereClause)
	rawQuery = s.db.Rebind(rawQuery)

	type row struct {
		MonthStart    string `db:"month_start"`
		IncomingCents int64  `db:"incoming_cents"`
		OutgoingCents int64  `db:"outgoing_cents"`
	}

	var rows []row
	executor := s.db.GetExecutor(ctx)
	if err := sqlx.SelectContext(ctx, executor, &rows, rawQuery, args...); err != nil {
		return nil, fmt.Errorf("sqlx_transaction_store: failed to fetch monthly analytics: %w", err)
	}

	points := make([]cashflow.MonthlyAnalyticsPoint, 0, len(rows))
	for _, r := range rows {
		month, err := time.Parse("2006-01-02", r.MonthStart)
		if err != nil {
			return nil, fmt.Errorf("sqlx_transaction_store: failed to parse month %q: %w", r.MonthStart, err)
		}
		month = month.UTC()
		points = append(points, cashflow.MonthlyAnalyticsPoint{
			Month:         month,
			IncomingCents: r.IncomingCents,
			OutgoingCents: r.OutgoingCents,
			NetCents:      r.IncomingCents - r.OutgoingCents,
		})
	}

	return points, nil
}

func (s *SQLXBankTransactionStore) FetchTagDistribution(ctx context.Context, query CashflowAnalyticsQuery) (*CashflowTagDistribution, error) {
	whereClause, args := buildCashflowAnalyticsWhereClause(query)
	rawQuery := fmt.Sprintf(`
		SELECT
			COALESCE(NULLIF(TRIM(tag), ''), 'untagged') AS tag_name,
			direction,
			COALESCE(SUM(amount_cents), 0) AS total_cents
		FROM %s
		%s
		GROUP BY tag_name, direction
	`, s.tableName, whereClause)
	rawQuery = s.db.Rebind(rawQuery)

	type row struct {
		TagName    string `db:"tag_name"`
		Direction  string `db:"direction"`
		TotalCents int64  `db:"total_cents"`
	}

	var rows []row
	executor := s.db.GetExecutor(ctx)
	if err := sqlx.SelectContext(ctx, executor, &rows, rawQuery, args...); err != nil {
		return nil, fmt.Errorf("sqlx_transaction_store: failed to fetch tag distribution: %w", err)
	}

	combined := make(map[string]int64)
	incoming := make(map[string]int64)
	outgoing := make(map[string]int64)

	for _, r := range rows {
		tag := strings.TrimSpace(r.TagName)
		if tag == "" {
			tag = "untagged"
		}
		combined[tag] += r.TotalCents

		switch r.Direction {
		case "in":
			incoming[tag] += r.TotalCents
		case "out":
			outgoing[tag] += r.TotalCents
		}
	}

	return &CashflowTagDistribution{
		Combined: mapToSortedDistributionEntries(combined),
		Incoming: mapToSortedDistributionEntries(incoming),
		Outgoing: mapToSortedDistributionEntries(outgoing),
	}, nil
}

func buildCashflowAnalyticsWhereClause(query CashflowAnalyticsQuery) (string, []any) {
	var (
		conditions []string
		args       []any
	)

	if !query.IncludeIgnored {
		conditions = append(conditions, "ignored = ?")
		args = append(args, false)
	}

	if query.From != nil {
		conditions = append(conditions, "date >= ?")
		args = append(args, *query.From)
	}

	if query.To != nil {
		conditions = append(conditions, "date <= ?")
		args = append(args, *query.To)
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func mapToSortedDistributionEntries(input map[string]int64) []CashflowTagDistributionEntry {
	out := make([]CashflowTagDistributionEntry, 0, len(input))
	for tag, total := range input {
		out = append(out, CashflowTagDistributionEntry{
			Tag:        tag,
			TotalCents: total,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].TotalCents == out[j].TotalCents {
			return out[i].Tag < out[j].Tag
		}
		return out[i].TotalCents > out[j].TotalCents
	})
	return out
}
