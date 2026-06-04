package portfolio

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	portfoliodomain "github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// GetPositionsRequest contains query filters for current portfolio positions.
type GetPositionsRequest struct {
	AccountID     uuid.UUID `query:"account_id"`
	IncludeClosed bool      `query:"include_closed"`
}

func (r GetPositionsRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	return len(problems) == 0, problems
}

// PositionResponse represents one current or closed position for an account.
type PositionResponse struct {
	ID               uuid.UUID  `json:"id"`
	Symbol           *string    `json:"symbol,omitempty"`
	Name             *string    `json:"name,omitempty"`
	Quantity         float64    `json:"quantity"`
	CostBasis        int64      `json:"cost_basis"`
	RealizedPnL      int64      `json:"realized_pnl"`
	MarketValue      *int64     `json:"market_value,omitempty"`
	UnrealizedPnLPct *float64   `json:"unrealized_pnl_pct,omitempty"`
	LastSnapshotAt   *time.Time `json:"last_snapshot_at,omitempty"`
	OpenDate         time.Time  `json:"open_date"`
	CloseDate        *time.Time `json:"close_date,omitempty"`
	IsClosed         bool       `json:"is_closed"`
}

// PositionsResponse returns account position rows with include-closed metadata.
type PositionsResponse struct {
	IncludeClosed bool               `json:"include_closed"`
	Data          []PositionResponse `json:"data"`
}

// GetPortfolioPositions returns positions with their latest snapshot metrics.
//
// @Summary Get portfolio positions
// @Description Returns portfolio positions for the given account and include_closed filter.
// @Tags portfolio
// @Accept json
// @Produce json
// @Param account_id query string true "Account ID"
// @Param include_closed query bool false "Include closed positions"
// @Success 200 {object} PositionsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/positions [get]
func GetPortfolioPositions(
	log logging.Logger,
	queries *portfoliodomain.Queries,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[GetPositionsRequest](r)
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
		if queries == nil {
			log.Error(r.Context(), "portfolio positions: queries are not configured", errors.New("portfolio queries not configured"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get portfolio positions"})
			return
		}

		rows, err := queries.PositionsForAccount(r.Context(), req.AccountID, req.IncludeClosed)
		if err != nil {
			log.Error(r.Context(), "portfolio positions: failed to list positions", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get portfolio positions"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, PositionsResponse{
			IncludeClosed: req.IncludeClosed,
			Data:          toPositionResponses(rows),
		})
	})
}

func toPositionResponses(positions []*portfoliodomain.PositionWithLatestSnapshot) []PositionResponse {
	out := make([]PositionResponse, 0, len(positions))
	for _, pos := range positions {
		if pos == nil {
			continue
		}
		out = append(out, PositionResponse{
			ID:               pos.ID,
			Symbol:           pos.Symbol,
			Name:             pos.Name,
			Quantity:         pos.Quantity,
			CostBasis:        int64(pos.CostBasis),
			RealizedPnL:      int64(pos.RealizedPnL),
			MarketValue:      pos.MarketValue,
			UnrealizedPnLPct: pos.UnrealizedPnLPct,
			LastSnapshotAt:   pos.LastSnapshotAt,
			OpenDate:         pos.OpenDate,
			CloseDate:        pos.CloseDate,
			IsClosed:         pos.IsClosed,
		})
	}
	return out
}
