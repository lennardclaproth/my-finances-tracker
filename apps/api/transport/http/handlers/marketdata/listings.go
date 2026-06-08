package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// CreateListingRequest creates a market listing.
type CreateListingRequest struct {
	Name        string  `json:"name"`
	Symbol      string  `json:"symbol"`
	Source      string  `json:"source"`
	Description *string `json:"description,omitempty"`
	Exchange    *string `json:"exchange,omitempty"`
	Region      *string `json:"region,omitempty"`
	Currency    *string `json:"currency,omitempty"`
	ISIN        *string `json:"isin,omitempty"`
	Ticker      *string `json:"ticker,omitempty"`
	Type        *string `json:"type,omitempty"`
}

func (r CreateListingRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Name) == "" {
		problems["name"] = "name is required"
	}
	if strings.TrimSpace(r.Symbol) == "" {
		problems["symbol"] = "symbol is required"
	}
	if strings.TrimSpace(r.Source) == "" {
		problems["source"] = "source is required"
	}
	if r.Currency != nil {
		currency := money.Currency(strings.TrimSpace(*r.Currency))
		if !currency.IsValid() {
			problems["currency"] = "currency is invalid"
		}
	}
	return len(problems) == 0, problems
}

// UpdateListingFieldsRequest updates mutable listing metadata fields.
type UpdateListingFieldsRequest struct {
	ID          uuid.UUID `json:"id"`
	Description *string   `json:"description,omitempty"`
	Exchange    *string   `json:"exchange,omitempty"`
	Region      *string   `json:"region,omitempty"`
	Currency    *string   `json:"currency,omitempty"`
	ISIN        *string   `json:"isin,omitempty"`
	Ticker      *string   `json:"ticker,omitempty"`
	Type        *string   `json:"type,omitempty"`
}

func (r UpdateListingFieldsRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.ID == uuid.Nil {
		problems["id"] = "id is required"
	}
	if r.Description == nil &&
		r.Exchange == nil &&
		r.Region == nil &&
		r.Currency == nil &&
		r.ISIN == nil &&
		r.Ticker == nil &&
		r.Type == nil {
		problems["listing"] = marketdata.ErrNoListingFieldsToUpdate.Error()
	}
	if r.Currency != nil {
		currency := money.Currency(strings.TrimSpace(*r.Currency))
		if !currency.IsValid() {
			problems["currency"] = "currency is invalid"
		}
	}
	return len(problems) == 0, problems
}

// SearchListingsRequest contains search and pagination inputs for listings.
type SearchListingsRequest struct {
	Q      string `query:"q"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}

func (r SearchListingsRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Q) == "" {
		problems["q"] = "q is required"
	}
	if r.Limit < 0 {
		problems["limit"] = "limit must be greater than or equal to 0"
	}
	if r.Limit > 100 {
		problems["limit"] = "limit must be less than or equal to 100"
	}
	if r.Offset < 0 {
		problems["offset"] = "offset must be greater than or equal to 0"
	}
	return len(problems) == 0, problems
}

// ListingResponse represents one market listing record.
type ListingResponse struct {
	ID          uuid.UUID `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Source      string    `json:"source"`
	Description *string   `json:"description,omitempty"`
	Exchange    *string   `json:"exchange,omitempty"`
	Region      *string   `json:"region,omitempty"`
	Currency    *string   `json:"currency,omitempty"`
	ISIN        *string   `json:"isin,omitempty"`
	Ticker      *string   `json:"ticker,omitempty"`
	Type        *string   `json:"type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PaginationResponse describes offset/limit paging metadata.
type PaginationResponse struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

// ListingsSearchResponse returns paginated listing search results.
type ListingsSearchResponse struct {
	Pagination PaginationResponse `json:"pagination"`
	Data       []ListingResponse  `json:"data"`
}

// CreateListing creates a new market-data listing.
//
// @Summary Create listing
// @Description Create a listing and trigger async end-of-day data accumulation.
// @Tags listings
// @Accept json
// @Produce json
// @Param request body CreateListingRequest true "Listing payload"
// @Success 200 {object} ListingResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/listing [post]
func CreateListing(
	log logging.Logger,
	commands *marketdata.Commands,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[CreateListingRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		listing, err := commands.CreateListing(
			r.Context(),
			strings.TrimSpace(req.Symbol),
			strings.TrimSpace(req.Name),
			marketdata.Source(strings.TrimSpace(req.Source)),
			listingOptions(req)...,
		)
		if err != nil {
			switch {
			case errors.Is(err, marketdata.ErrListingAlreadyExists):
				_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{"listing": "listing already exists"})
			case isListingValidationError(err):
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"listing": err.Error()})
			default:
				log.Error(r.Context(), "create listing: failed to create listing", err)
				_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create listing"})
			}
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toListingResponse(listing))
	})
}

// UpdateListingFields updates specific fields for an existing listing.
//
// @Summary Update listing fields
// @Description Patch selected listing fields; omitted fields remain unchanged.
// @Tags listings
// @Accept json
// @Produce json
// @Param request body UpdateListingFieldsRequest true "Listing patch payload"
// @Success 200 {object} ListingResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/listing [patch]
func UpdateListingFields(
	log logging.Logger,
	commands *marketdata.Commands,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[UpdateListingFieldsRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		listing, err := commands.UpdateListingFields(
			r.Context(),
			req.ID,
			req.Description,
			req.Exchange,
			req.Region,
			req.Currency,
			req.ISIN,
			req.Ticker,
			req.Type,
		)
		if err != nil {
			switch {
			case errors.Is(err, marketdata.ErrListingNotFound):
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"listing": "listing not found"})
			case errors.Is(err, marketdata.ErrNoListingFieldsToUpdate) ||
				errors.Is(err, marketdata.ErrInvalidListingCurrency):
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"listing": err.Error()})
			default:
				log.Error(r.Context(), "update listing fields: failed to update listing", err)
				_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to update listing"})
			}
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toListingResponse(listing))
	})
}

// GetListings returns all listings.
//
// @Summary List listings
// @Description Return all market-data listings ordered by symbol.
// @Tags listings
// @Accept json
// @Produce json
// @Success 200 {array} ListingResponse
// @Failure 500 {object} map[string]string
// @Router /marketdata/listings [get]
func GetListings(
	log logging.Logger,
	queries *marketdata.Queries,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listings, err := queries.ListListings(r.Context())
		if err != nil {
			log.Error(r.Context(), "list listings: failed to list listings", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to list listings"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toListingResponses(listings))
	})
}

// SearchListings searches listings by symbol, name or ISIN.
//
// @Summary Search listings
// @Description Search market-data listings using a case-insensitive partial query over symbol, name and isin.
// @Tags listings
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Page size (max 100, default 25)"
// @Param offset query int false "Offset"
// @Success 200 {object} ListingsSearchResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/listings/search [get]
func SearchListings(
	log logging.Logger,
	queries *marketdata.Queries,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[SearchListingsRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		limit := req.Limit
		if limit == 0 {
			limit = 25
		}

		listings, total, err := queries.SearchListings(r.Context(), strings.TrimSpace(req.Q), limit, req.Offset)
		if err != nil {
			log.Error(r.Context(), "search listings: failed to search listings", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to search listings"})
			return
		}

		data := toListingResponses(listings)
		_ = httpx.JSONEncode(w, http.StatusOK, ListingsSearchResponse{
			Pagination: PaginationResponse{
				Limit:  limit,
				Offset: req.Offset,
				Count:  len(data),
				Total:  total,
			},
			Data: data,
		})
	})
}

func listingOptions(req CreateListingRequest) []marketdata.ListingOption {
	var options []marketdata.ListingOption
	if req.ISIN != nil {
		options = append(options, marketdata.ListingWithISIN(*req.ISIN))
	}
	if req.Exchange != nil {
		options = append(options, marketdata.ListingWithExchange(*req.Exchange))
	}
	if req.Currency != nil {
		options = append(options, marketdata.ListingWithCurrency(money.Currency(strings.TrimSpace(*req.Currency))))
	}
	if req.Description != nil {
		options = append(options, marketdata.ListingWithDescription(*req.Description))
	}
	if req.Region != nil {
		options = append(options, marketdata.ListingWithRegion(*req.Region))
	}
	if req.Type != nil {
		options = append(options, marketdata.ListingWithType(*req.Type))
	}
	if req.Ticker != nil {
		options = append(options, marketdata.ListingWithTicker(*req.Ticker))
	}
	return options
}

func isListingValidationError(err error) bool {
	return errors.Is(err, marketdata.ErrListingNameEmpty) ||
		errors.Is(err, marketdata.ErrListingSymbolEmpty) ||
		errors.Is(err, marketdata.ErrListingSourceEmpty) ||
		errors.Is(err, marketdata.ErrInvalidListingCurrency)
}

func toListingResponses(listings []*marketdata.Listing) []ListingResponse {
	response := make([]ListingResponse, 0, len(listings))
	for _, listing := range listings {
		if listing == nil {
			continue
		}
		response = append(response, toListingResponse(listing))
	}
	return response
}

func toListingResponse(listing *marketdata.Listing) ListingResponse {
	if listing == nil {
		return ListingResponse{}
	}

	var currency *string
	if listing.Currency != nil {
		s := string(*listing.Currency)
		currency = &s
	}

	return ListingResponse{
		ID:          listing.ID,
		Symbol:      listing.Symbol,
		Name:        listing.Name,
		Source:      string(listing.Source),
		Description: listing.Description,
		Exchange:    listing.Exchange,
		Region:      listing.Region,
		Currency:    currency,
		ISIN:        listing.ISIN,
		Ticker:      listing.Ticker,
		Type:        listing.Type,
		CreatedAt:   listing.CreatedAt,
		UpdatedAt:   listing.UpdatedAt,
	}
}
