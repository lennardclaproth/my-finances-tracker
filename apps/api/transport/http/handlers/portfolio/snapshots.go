package portfolio

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	portfoliodomain "github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// GetSnapshotsRequest contains query filters for portfolio snapshot history.
type GetSnapshotsRequest struct {
	AccountID uuid.UUID `query:"account_id"`
	From      string    `query:"from"`
	To        string    `query:"to"`
}

func (r GetSnapshotsRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return len(problems) == 0, problems
}

// SnapshotPointResponse represents one portfolio snapshot point in the timeline.
type SnapshotPointResponse struct {
	OccurredAt            time.Time `json:"occurred_at"`
	MarketValue           int64     `json:"market_value"`
	TotalPnL              int64     `json:"total_pnl"`
	TotalPnLPct           float64   `json:"total_pnl_pct"`
	TotalCostBasis        int64     `json:"total_cost_basis"`
	ReturnVsCostBasisPct  float64   `json:"return_vs_cost_basis_pct"`
	DailyReturnPct        float64   `json:"daily_return_pct"`
	TimeWeightedReturnPct float64   `json:"time_weighted_return_pct"`
	ValueIndex            float64   `json:"value_index"`
}

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
// @Success 200 {array} SnapshotPointResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/snapshots [get]
func GetPortfolioSnapshots(
	log logging.Logger,
	fetcher account.Fetcher,
	queries *portfoliodomain.Queries,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[GetSnapshotsRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
			return
		}
		if ok, problems := req.isValid(); !ok {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		if fetcher == nil {
			log.Error(r.Context(), "portfolio snapshots: account fetcher is not configured", errors.New("account fetcher not configured"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get portfolio snapshots"})
			return
		}
		if queries == nil {
			log.Error(r.Context(), "portfolio snapshots: queries are not configured", errors.New("portfolio queries not configured"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get portfolio snapshots"})
			return
		}

		if _, err := fetcher.FetchByID(r.Context(), req.AccountID); err != nil {
			if errors.Is(err, account.ErrAccountNotFound) {
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"account_id": account.ErrAccountNotFound.Error()})
				return
			}
			log.Error(r.Context(), "portfolio snapshots: failed to fetch account", err, "account_id", req.AccountID.String())
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get portfolio snapshots"})
			return
		}

		from, to, err := date.ParseFromTo(req.From, req.To)
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"date_range": err.Error()})
			return
		}
		if to != nil {
			toEnd := date.EndOfDayUTC(*to)
			to = &toEnd
		}

		sort := sorting.ASC
		snapshots, err := queries.SnapshotsForAccount(r.Context(), req.AccountID, nil, nil, from, to, &sort)
		if err != nil {
			log.Error(r.Context(), "portfolio snapshots: failed to list snapshots", err, "account_id", req.AccountID.String())
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get portfolio snapshots"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toSnapshotPointResponses(snapshots))
	})
}

func toSnapshotPointResponses(snapshots []*portfoliodomain.PortfolioSnapshot) []SnapshotPointResponse {
	out := make([]SnapshotPointResponse, 0, len(snapshots))
	for _, snap := range snapshots {
		if snap == nil {
			continue
		}
		out = append(out, SnapshotPointResponse{
			OccurredAt:            snap.OccurredAt,
			MarketValue:           int64(snap.MarketValue),
			TotalPnL:              int64(snap.TotalPnL),
			TotalPnLPct:           snap.TotalPnLPct,
			TotalCostBasis:        int64(snap.CostBasis),
			ReturnVsCostBasisPct:  portfoliodomain.SnapshotReturnVsCostBasisPct(snap),
			DailyReturnPct:        snap.DailyDeltaPnLPct,
			TimeWeightedReturnPct: snap.TimeWeightedReturnPct,
			ValueIndex:            portfoliodomain.SnapshotValueIndex(snap),
		})
	}
	return out
}
