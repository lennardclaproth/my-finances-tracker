package portfolio

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

// GetPortfolioSnapshots returns snapshot series for charting the portfolio.
//
// @Summary Get portfolio snapshots
// @Description Returns snapshot points for the given account and optional date range.
// @Tags portfolio
// @Accept json
// @Produce json
// @Param account_id query string true "Account ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} api.PortfolioSnapshotPointResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/snapshots [get]
func GetPortfolioSnapshots(
	log logging.Logger,
	fetcher account.Fetcher,
	lister portfolioSnapshotLister,
) http.Handler {
	endpoint := func(ctx context.Context, req api.GetPortfolioSnapshotsRequest) (status int, res any, err error) {
		if _, err := fetcher.FetchByID(ctx, req.AccountID); err != nil {
			if errors.Is(err, account.ErrAccountNotFound) {
				return http.StatusNotFound, map[string]string{"account_id": account.ErrAccountNotFound.Error()}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}

		from, to, err := parsePortfolioDateRange(req.From, req.To)
		if err != nil {
			return http.StatusBadRequest, portfolioDateRangeProblem(err), nil
		}

		snapshots, err := lister.ListForAccount(ctx, req.AccountID, from, to)
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}

		out := make([]api.PortfolioSnapshotPointResponse, 0, len(snapshots))
		for _, snap := range snapshots {
			if snap == nil {
				continue
			}
			out = append(out, api.PortfolioSnapshotPointResponse{
				OccurredAt:            snap.OccurredAt,
				MarketValue:           int64(snap.MarketValue),
				TotalPnL:              int64(snap.TotalPnL),
				TotalPnLPct:           snap.TotalPnLPct,
				TotalCostBasis:        int64(snap.CostBasis),
				ReturnVsCostBasisPct:  portfolio.SnapshotReturnVsCostBasisPct(snap),
				DailyReturnPct:        snap.DailyDeltaPnLPct,
				TimeWeightedReturnPct: snap.TimeWeightedReturnPct,
				ValueIndex:            portfolio.SnapshotValueIndex(snap),
			})
		}
		return http.StatusOK, out, nil
	}

	decodeFn := httpx.DecoderFunc[api.GetPortfolioSnapshotsRequest](func(r *http.Request) (api.GetPortfolioSnapshotsRequest, error) {
		var req api.GetPortfolioSnapshotsRequest
		res, err := httpx.DecodeQuery[api.GetPortfolioSnapshotsRequest](r)
		if err != nil {
			return req, fmt.Errorf("GetPortfolioSnapshots failed to decode query: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}
