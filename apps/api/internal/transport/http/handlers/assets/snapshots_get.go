package assets

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/transport/http"
)

// GetAssetSnapshotsRequest contains query filters for account-level asset snapshots.
type GetSnapshotsRequest struct {
	AccountID uuid.UUID `json:"account_id" query:"account_id"`
	From      *string   `json:"from,omitempty" query:"from"`
	To        *string   `json:"to,omitempty" query:"to"`
}

type SnapshotResponse struct {
	Date       string `json:"date"`
	TotalWorth string `json:"total_worth"`
}

// GetAssetSnapshots returns account-level total asset snapshots.
//
// @Summary List asset snapshots
// @Description Returns account-level daily total worth snapshot points.
// @Tags assets
// @Accept json
// @Produce json
// @Param account_id query string true "Account ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} SnapshotResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/snapshots [get]
func GetSnapshots(log logging.Logger, queries assets.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[GetSnapshotsRequest](r)
		if err != nil {
			httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
		}
		var from *time.Time
		if req.From != nil && len(strings.TrimSpace(*req.From)) != 0 {
			parsed, err := time.Parse("2026-01-02", *req.From)
			if err != nil {
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{
					"from": "from must be in YYYY-MM-DD format",
				})
				return
			}
			from = &parsed
		}
		var to *time.Time
		if req.To != nil && len(strings.TrimSpace(*req.To)) != 0 {
			parsed, err := time.Parse("2026-01-02", *req.To)
			if err != nil {
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{
					"from": "from must be in YYYY-MM-DD format",
				})
				return
			}
			to = &parsed
		}
		snapshots, err := queries.ListSnapshots(r.Context(), req.AccountID, from, to)
		if err != nil {
			log.Error(r.Context(), "An error occurred while trying to get snapshots", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get snapshots"})
			return
		}
		res := make([]SnapshotResponse, 0, len(snapshots))
		for _, snapshot := range snapshots {
			res = append(res, SnapshotResponse{
				Date:       snapshot.Date.Format("2026-01-02"),
				TotalWorth: snapshot.TotalWorth.String(),
			})
		}
		_ = httpx.JSONEncode(w, http.StatusOK, res)
	})
}
