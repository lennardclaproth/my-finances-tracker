package cashflow

import (
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// MonthlyAnalyticsRequest contains query filters for monthly cashflow analytics.
type MonthlyAnalyticsRequest struct {
	From           string `query:"from"`
	To             string `query:"to"`
	IncludeIgnored bool   `query:"include_ignored"`
}

// MonthlyAnalyticsPointResponse represents one month of aggregated cashflow metrics.
type MonthlyAnalyticsPointResponse struct {
	Month         string `json:"month"`
	IncomingCents int64  `json:"incoming_cents"`
	OutgoingCents int64  `json:"outgoing_cents"`
	NetCents      int64  `json:"net_cents"`
}

// CashflowMonthlyAnalyticsResponse returns monthly analytics time-series data.
type CashflowMonthlyAnalyticsResponse struct {
	Data []MonthlyAnalyticsPointResponse `json:"data"`
}

// GetMonthlyAnalytics returns incoming, outgoing and net totals grouped per month.
//
// @Summary     Get cashflow monthly analytics
// @Description Returns monthly incoming, outgoing and net totals for the selected range.
// @Tags        Transactions
// @Accept      application/json
// @Produce     application/json
// @Param       from query string false "Start date (YYYY-MM-DD)"
// @Param       to query string false "End date (YYYY-MM-DD)"
// @Param       include_ignored query bool false "Include ignored transactions"
// @Success     200 {object} CashflowMonthlyAnalyticsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/analytics/monthly [get]
func GetMonthlyAnalytics(log logging.Logger, queries *cashflow.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[MonthlyAnalyticsRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}

			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{
				"error": "invalid query parameters",
			})
			return
		}

		from, to, err := date.ParseFromTo(req.From, req.To)
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"date_range": err.Error()})
			return
		}

		points, err := queries.MonthlyAnalytics(r.Context(), cashflow.AnalyticsFilter{
			From:           from,
			To:             to,
			IncludeIgnored: req.IncludeIgnored,
		})
		if err != nil {
			log.Error(r.Context(), "cashflow monthly analytics: failed to get monthly analytics", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get cashflow monthly analytics"})
			return
		}

		out := make([]MonthlyAnalyticsPointResponse, 0, len(points))
		for _, point := range points {
			out = append(out, MonthlyAnalyticsPointResponse{
				Month:         point.Month.Format("2006-01-02"),
				IncomingCents: point.IncomingCents,
				OutgoingCents: point.OutgoingCents,
				NetCents:      point.NetCents,
			})
		}

		_ = httpx.JSONEncode(w, http.StatusOK, CashflowMonthlyAnalyticsResponse{Data: out})
	})
}
