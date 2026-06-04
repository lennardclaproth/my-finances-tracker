package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// GetEODRequest contains filters for listing daily market data retrieval.
type GetEODRequest struct {
	ListingID *uuid.UUID `query:"listing_id"`
	Symbol    string     `query:"symbol"`
	From      string     `query:"from"`
	To        string     `query:"to"`
	SortOrder string     `query:"sort_order"`
	Limit     int        `query:"limit"`
	Offset    int        `query:"offset"`
}

func (r GetEODRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)

	if r.ListingID == nil && strings.TrimSpace(r.Symbol) == "" {
		problems["symbol"] = "symbol is required when listing_id is not provided"
	}
	if r.Limit < 0 {
		problems["limit"] = "limit must be greater than or equal to 0"
	}
	if r.Offset < 0 {
		problems["offset"] = "offset must be greater than or equal to 0"
	}
	if r.SortOrder != "" {
		switch strings.ToLower(strings.TrimSpace(r.SortOrder)) {
		case sorting.ASC.String(), sorting.DESC.String():
		default:
			problems["sort_order"] = "sort_order must be either asc or desc"
		}
	}

	return len(problems) == 0, problems
}

// GetEODResponse returns daily rows and freshness metadata.
type GetEODResponse struct {
	Data     []marketdata.EOD
	Metadata GetEODMetadataResponse
}

// GetEODMetadataResponse describes daily retrieval metadata.
type GetEODMetadataResponse struct {
	Message     string
	ResultCount int
	TotalCount  int
}

// GetEOD fetches end of day market data for a given symbol.
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
// @Success 200 {object} GetEODResponse
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /marketdata/dailies [get]
func GetEOD(
	log logging.Logger,
	queries *marketdata.Queries,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[GetEODRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}

			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{
				"error": "invalid query parameters",
			})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		from, to, err := date.ParseFromTo(req.From, req.To)
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, err.Error())
			return
		}

		limit := req.Limit
		if limit == 0 {
			limit = 100
		}

		listingID, err := resolveListingID(r, queries, req)
		if err != nil {
			if errors.Is(err, marketdata.ErrListingNotFound) {
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"listing": "listing not found"})
				return
			}

			log.Error(r.Context(), "marketdata dailies: failed to resolve listing", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to get market data dailies",
			})
			return
		}

		result, err := queries.GetEODByListing(r.Context(), listingID, from, to, limit, req.Offset, sorting.MustParse(req.SortOrder))
		if err != nil {
			if errors.Is(err, marketdata.ErrListingNotFound) {
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"listing": "listing not found"})
				return
			}

			log.Error(r.Context(), "marketdata dailies: failed to get EOD data", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to get market data dailies",
			})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toEODResponse(result))
	})
}

func resolveListingID(r *http.Request, queries *marketdata.Queries, req GetEODRequest) (uuid.UUID, error) {
	if req.ListingID != nil {
		return *req.ListingID, nil
	}

	listing, err := queries.ListingBySymbol(r.Context(), strings.TrimSpace(req.Symbol))
	if err != nil {
		return uuid.Nil, err
	}
	if listing == nil {
		return uuid.Nil, marketdata.ErrListingNotFound
	}

	return listing.ID, nil
}

func toEODResponse(result *marketdata.EODResult) GetEODResponse {
	if result == nil {
		return GetEODResponse{
			Data: []marketdata.EOD{},
		}
	}

	data := make([]marketdata.EOD, 0, len(result.Data))
	data = append(data, result.Data...)

	return GetEODResponse{
		Data: data,
		Metadata: GetEODMetadataResponse{
			Message:     result.Metadata.Message,
			ResultCount: result.Metadata.ResultCount,
			TotalCount:  result.Metadata.TotalCount,
		},
	}
}
