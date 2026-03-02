package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/api"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

// CreateListing creates a new market-data listing.
//
// @Summary Create listing
// @Description Create a listing and trigger async daily data accumulation.
// @Tags listings
// @Accept json
// @Produce json
// @Param request body api.CreateListingRequest true "Listing payload"
// @Success 200 {object} api.ListingResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/listing [post]
func CreateListing(
	log logging.Logger,
	mds *marketdata.Service,
) http.Handler {
	endpoint := func(ctx context.Context, r api.CreateListingRequest) (status int, res any, err error) {
		var options []marketdata.ListingOption
		if r.ISIN != nil {
			options = append(options, marketdata.ListingWithISIN(*r.ISIN))
		}
		if r.Exchange != nil {
			options = append(options, marketdata.ListingWithExchange(*r.Exchange))
		}
		if r.Currency != nil {
			cur := money.Currency(*r.Currency)
			if !cur.IsValid() {
				return http.StatusBadRequest, struct{}{}, fmt.Errorf("invalid currency: %s", *r.Currency)
			}
			options = append(options, marketdata.ListingWithCurrency(cur))
		}
		if r.Description != nil {
			options = append(options, marketdata.ListingWithDescription(*r.Description))
		}
		if r.Region != nil {
			options = append(options, marketdata.ListingWithRegion(*r.Region))
		}
		if r.Type != nil {
			options = append(options, marketdata.ListingWithType(*r.Type))
		}
		if r.Ticker != nil {
			options = append(options, marketdata.ListingWithTicker(*r.Ticker))
		}
		listing, err := mds.CreateListing(ctx, r.Symbol, r.Name, marketdata.Source(r.Source), options...)
		if err != nil {
			if errors.Is(err, marketdata.ErrListingAlreadyExists) {
				return http.StatusConflict, map[string]string{"listing": "listing already exists"}, nil
			}
			if errors.Is(err, marketdata.ErrListingNameEmpty) ||
				errors.Is(err, marketdata.ErrListingSymbolEmpty) ||
				errors.Is(err, marketdata.ErrListingSourceEmpty) ||
				errors.Is(err, marketdata.ErrInvalidListingCurrency) {
				return http.StatusBadRequest, map[string]string{"listing": err.Error()}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, toListingResponse(listing), nil
	}

	decodeFn := httpx.DecoderFunc[api.CreateListingRequest](func(r *http.Request) (api.CreateListingRequest, error) {
		var req api.CreateListingRequest
		res, err := httpx.DecodeJSON[api.CreateListingRequest](r)
		if err != nil {
			return req, fmt.Errorf("CreateListing failed to decode request: %w", err)
		}
		return res, err
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

// UpdateListingFields updates specific fields for an existing listing.
//
// @Summary Update listing fields
// @Description Patch selected listing fields; omitted fields remain unchanged.
// @Tags listings
// @Accept json
// @Produce json
// @Param request body api.UpdateListingFieldsRequest true "Listing patch payload"
// @Success 200 {object} api.ListingResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/listing [patch]
func UpdateListingFields(
	log logging.Logger,
	mds *marketdata.Service,
) http.Handler {
	endpoint := func(ctx context.Context, r api.UpdateListingFieldsRequest) (status int, res any, err error) {
		listing, err := mds.UpdateListingFields(
			ctx,
			r.Id,
			r.Description,
			r.Exchange,
			r.Region,
			r.Currency,
			r.ISIN,
			r.Ticker,
			r.Type,
		)
		if err != nil {
			if errors.Is(err, marketdata.ErrNoListingFieldsToUpdate) ||
				errors.Is(err, marketdata.ErrInvalidListingCurrency) {
				return http.StatusBadRequest, map[string]string{"listing": err.Error()}, nil
			}
			if errors.Is(err, marketdata.ErrListingNotFound) {
				return http.StatusNotFound, map[string]string{"listing": "listing not found"}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, toListingResponse(listing), nil
	}

	decodeFn := httpx.DecoderFunc[api.UpdateListingFieldsRequest](func(r *http.Request) (api.UpdateListingFieldsRequest, error) {
		var req api.UpdateListingFieldsRequest
		res, err := httpx.DecodeJSON[api.UpdateListingFieldsRequest](r)
		if err != nil {
			return req, fmt.Errorf("UpdateListingFields failed to decode request: %w", err)
		}
		return res, err
	})
	return httpx.Endpoint(decodeFn, log, endpoint)
}

// GetListings returns all listings.
//
// @Summary List listings
// @Description Return all market-data listings ordered by symbol.
// @Tags listings
// @Accept json
// @Produce json
// @Success 200 {array} api.ListingResponse
// @Failure 500 {object} map[string]string
// @Router /marketdata/listings [get]
func GetListings(
	log logging.Logger,
	mds *marketdata.Service,
) http.Handler {
	endpoint := func(ctx context.Context, _ struct{}) (status int, res []api.ListingResponse, err error) {
		listings, err := mds.ListListings(ctx)
		if err != nil {
			return http.StatusInternalServerError, nil, err
		}
		response := make([]api.ListingResponse, 0, len(listings))
		for _, listing := range listings {
			if listing == nil {
				continue
			}
			response = append(response, toListingResponse(listing))
		}
		return http.StatusOK, response, nil
	}

	decodeFn := httpx.DecoderFunc[struct{}](func(r *http.Request) (struct{}, error) {
		return struct{}{}, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}

func toListingResponse(listing *marketdata.Listing) api.ListingResponse {
	var currency *string
	if listing != nil && listing.Currency != nil {
		s := string(*listing.Currency)
		currency = &s
	}
	if listing == nil {
		return api.ListingResponse{}
	}
	return api.ListingResponse{
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
