package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

// GetDailies fetches daily market data for a given symbol.
//
// @Summary Fetch daily market data
// @Description Get daily market data for a symbol with optional date range and pagination
// @Tags dailies
// @Accept json
// @Produce json
// @Param listing_id query string false "Listing ID (preferred when duplicate symbols exist across sources)"
// @Param symbol query string false "Listing symbol (required when listing_id is omitted, e.g. TDT.AS)"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param sort_order query string false "Date sort order: asc or desc (default asc)"
// @Param limit query int false "Page size"
// @Param offset query int false "Offset"
// @Success 200 {object} marketdata.DailyResponse
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /dailies [get]
func GetDailies(
	log logging.Logger,
	mds *marketdata.Service,
) http.Handler {
	endpoint := func(ctx context.Context, req api.GetDailiesRequest) (status int, res any, err error) {
		var from *time.Time
		if req.From != "" {
			parsedFrom, parseErr := time.Parse("2006-01-02", req.From)
			if parseErr != nil {
				return http.StatusBadRequest, map[string]string{"from": "from must be in YYYY-MM-DD format"}, nil
			}
			from = &parsedFrom
		}

		var to *time.Time
		if req.To != "" {
			parsedTo, parseErr := time.Parse("2006-01-02", req.To)
			if parseErr != nil {
				return http.StatusBadRequest, map[string]string{"to": "to must be in YYYY-MM-DD format"}, nil
			}
			to = &parsedTo
		}

		limit := req.Limit
		if limit == 0 {
			limit = 100
		}
		sortOrder := marketdata.NormalizeDailyDateSortOrder(req.SortOrder)

		var resp *marketdata.DailyResponse
		if strings.TrimSpace(req.ListingID) != "" {
			listingID, parseErr := uuid.Parse(strings.TrimSpace(req.ListingID))
			if parseErr != nil {
				return http.StatusBadRequest, map[string]string{"listing_id": "listing_id must be a valid UUID"}, nil
			}
			resp, err = mds.GetDailiesByListingIDWithSort(ctx, listingID, from, to, limit, req.Offset, sortOrder)
		} else {
			resp, err = mds.GetDailiesWithSort(ctx, req.Symbol, from, to, limit, req.Offset, sortOrder)
		}
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}

		return http.StatusOK, resp, nil
	}

	decodeFn := httpx.DecoderFunc[api.GetDailiesRequest](func(r *http.Request) (api.GetDailiesRequest, error) {
		var req api.GetDailiesRequest
		res, err := httpx.DecodeQuery[api.GetDailiesRequest](r)
		if err != nil {
			return req, fmt.Errorf("GetDailies failed to decode query: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}
