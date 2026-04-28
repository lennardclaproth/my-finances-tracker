package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/lennardclaproth/my-finances-tracker/api"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

type manualPortfolioTransactionCreator interface {
	Create(ctx context.Context, input portfolio.ManualTransactionInput) (*portfolio.ManualTransactionCreateResult, error)
}

// CreateManualPortfolioTransaction creates a manual portfolio transaction without triggering rebuilds.
//
// @Summary Create manual portfolio transaction
// @Description Creates a manual portfolio transaction and persists it without publishing rebuild events.
// @Tags portfolio
// @Accept json
// @Produce json
// @Param request body api.CreateManualPortfolioTransactionRequest true "Manual portfolio transaction payload"
// @Success 201 {object} api.ManualPortfolioTransactionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/transactions/manual [post]
func CreateManualPortfolioTransaction(
	log logging.Logger,
	svc manualPortfolioTransactionCreator,
) http.Handler {
	endpoint := func(ctx context.Context, req api.CreateManualPortfolioTransactionRequest) (status int, res any, err error) {
		result, err := svc.Create(ctx, portfolio.ManualTransactionInput{
			AccountID:   req.AccountID,
			VendorID:    req.VendorID,
			OccurredAt:  req.OccurredAt,
			Type:        req.Type,
			ListingID:   req.ListingID,
			Amount:      req.Amount,
			Quantity:    req.Quantity,
			Description: req.Description,
		})
		if err != nil {
			switch {
			case errors.Is(err, portfolio.ErrManualAccountNotFound):
				return http.StatusNotFound, map[string]string{"account_id": err.Error()}, nil
			case errors.Is(err, portfolio.ErrManualVendorNotFound):
				return http.StatusNotFound, map[string]string{"vendor_id": err.Error()}, nil
			case errors.Is(err, portfolio.ErrManualListingNotFound):
				return http.StatusNotFound, map[string]string{"listing_id": err.Error()}, nil
			case errors.Is(err, portfolio.ErrManualVendorTypeNotSupported):
				return http.StatusUnprocessableEntity, map[string]string{"vendor_id": err.Error()}, nil
			case errors.Is(err, portfolio.ErrDuplicateTransaction):
				return http.StatusConflict, map[string]string{"transaction": "duplicate transaction"}, nil
			case isManualValidationErr(err):
				return http.StatusBadRequest, map[string]string{"transaction": err.Error()}, nil
			default:
				return http.StatusInternalServerError, struct{}{}, err
			}
		}
		if result == nil || result.Transaction == nil || result.Transaction.AccountID == nil {
			return http.StatusInternalServerError, struct{}{}, fmt.Errorf("manual transaction: invalid create result")
		}

		tx := result.Transaction
		return http.StatusCreated, api.ManualPortfolioTransactionResponse{
			ID:          tx.ID,
			AccountID:   *tx.AccountID,
			Origin:      string(tx.Origin),
			Source:      tx.Source,
			OccurredAt:  tx.OccurredAt,
			Type:        string(tx.Type),
			ListingID:   result.ListingID,
			ISIN:        tx.ISIN,
			Symbol:      tx.Symbol,
			Description: tx.Description,
			Amount:      formatDecimal(result.SignedAmount),
			Quantity:    formatDecimal(tx.Quantity),
			UnitPrice:   formatDecimal(tx.UnitPrice.Float64()),
			CreatedAt:   tx.CreatedAt,
			UpdatedAt:   tx.UpdatedAt,
		}, nil
	}

	decodeFn := httpx.DecoderFunc[api.CreateManualPortfolioTransactionRequest](func(r *http.Request) (api.CreateManualPortfolioTransactionRequest, error) {
		var req api.CreateManualPortfolioTransactionRequest
		res, err := httpx.DecodeJSON[api.CreateManualPortfolioTransactionRequest](r)
		if err != nil {
			return req, fmt.Errorf("CreateManualPortfolioTransaction failed to decode request: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}

func isManualValidationErr(err error) bool {
	return errors.Is(err, portfolio.ErrManualVendorNotActive) ||
		errors.Is(err, portfolio.ErrManualInvalidOccurredAt) ||
		errors.Is(err, portfolio.ErrManualInvalidType) ||
		errors.Is(err, portfolio.ErrManualInvalidAmount) ||
		errors.Is(err, portfolio.ErrManualInvalidQuantity) ||
		errors.Is(err, portfolio.ErrManualQuantityRequired) ||
		errors.Is(err, portfolio.ErrManualQuantityForbidden) ||
		errors.Is(err, portfolio.ErrManualListingRequired) ||
		errors.Is(err, portfolio.ErrManualListingForbidden) ||
		errors.Is(err, portfolio.ErrManualNonCashAmountMustBePos) ||
		errors.Is(err, portfolio.ErrManualCashAmountMustBeNonZero) ||
		errors.Is(err, portfolio.ErrManualQuantityMustBePositive) ||
		errors.Is(err, portfolio.ErrManualListingIdentityMissing)
}

func formatDecimal(v float64) string {
	raw := strconv.FormatFloat(v, 'f', 6, 64)
	raw = strings.TrimRight(raw, "0")
	raw = strings.TrimRight(raw, ".")
	if raw == "-0" || raw == "" {
		return "0"
	}
	return raw
}
