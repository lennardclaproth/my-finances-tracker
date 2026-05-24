package cashflow

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/api"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

// GetCashflowMonthlyAnalytics returns incoming, outgoing and net totals grouped per month.
//
// @Summary     Get cashflow monthly analytics
// @Description Returns monthly incoming, outgoing and net totals for the selected range.
// @Tags        Transactions
// @Accept      application/json
// @Produce     application/json
// @Param       from query string false "Start date (YYYY-MM-DD)"
// @Param       to query string false "End date (YYYY-MM-DD)"
// @Param       include_ignored query bool false "Include ignored transactions"
// @Success     200 {object} api.CashflowMonthlyAnalyticsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/analytics/monthly [get]
func GetCashflowMonthlyAnalytics(log logging.Logger, store *storage.SQLXBankTransactionStore) http.Handler {
	endpoint := func(ctx context.Context, req api.GetCashflowAnalyticsRequest) (status int, res any, err error) {
		from, to, parseErr := parseCashflowAnalyticsRange(req.From, req.To)
		if parseErr != nil {
			return http.StatusBadRequest, map[string]string{"date_range": parseErr.Error()}, nil
		}

		points, err := store.FetchMonthlyAnalytics(ctx, storage.CashflowAnalyticsQuery{
			From:           from,
			To:             to,
			IncludeIgnored: req.IncludeIgnored,
		})
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}

		out := make([]api.CashflowMonthlyAnalyticsPoint, 0, len(points))
		for _, point := range points {
			out = append(out, api.CashflowMonthlyAnalyticsPoint{
				Month:         point.Month.Format("2006-01-02"),
				IncomingCents: point.IncomingCents,
				OutgoingCents: point.OutgoingCents,
				NetCents:      point.NetCents,
			})
		}

		return http.StatusOK, api.CashflowMonthlyAnalyticsResponse{Data: out}, nil
	}

	decodeFn := httpx.DecoderFunc[api.GetCashflowAnalyticsRequest](func(r *http.Request) (api.GetCashflowAnalyticsRequest, error) {
		var req api.GetCashflowAnalyticsRequest
		res, err := httpx.DecodeQuery[api.GetCashflowAnalyticsRequest](r)
		if err != nil {
			return req, fmt.Errorf("GetCashflowMonthlyAnalytics failed to decode query: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}
