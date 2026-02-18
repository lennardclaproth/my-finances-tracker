package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/api"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

func CreateListing(
	log logging.Logger,
	mds *marketdata.Service,
) http.Handler {
	endpoint := func(ctx context.Context, r api.CreateListingRequest) (status int, res struct{}, err error) {
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
		_, err = mds.CreateListing(ctx, r.Symbol, r.Name, marketdata.Source(r.Source), options...)
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, struct{}{}, nil
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

func UpdateListingFields(
	log logging.Logger,
) http.Handler {
	endpoint := func(ctx context.Context, r api.UpdateListingFieldsRequest) (status int, res struct{}, err error) {
		return http.StatusOK, struct{}{}, nil
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
