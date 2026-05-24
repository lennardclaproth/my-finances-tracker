package cashflow

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/api"
	cashflowservice "github.com/lennardclaproth/my-finances-tracker/internal/cashflow/service"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

// IgnoreTransactionsBySelection sets the ignored status for selected transaction IDs.
//
// @Summary     Ignore selected cashflow transactions
// @Description Set ignored=true/false for selected transaction IDs.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.IgnoreTransactionsBySelectionRequest true "Selection ignore request"
// @Success     200 {object} api.TagTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/ignore/selection [post]
// @Tags        Transactions
func IgnoreTransactionsBySelection(log logging.Logger, store *storage.SQLXBankTransactionStore) http.HandlerFunc {
	endpoint := func(ctx context.Context, req api.IgnoreTransactionsBySelectionRequest) (status int, res api.TagTransactionsResponse, err error) {
		ignored := true
		if req.Ignored != nil {
			ignored = *req.Ignored
		}
		updated, err := store.UpdateIgnoredByIDs(ctx, req.IDs, ignored)
		if err != nil {
			return http.StatusInternalServerError, api.TagTransactionsResponse{}, err
		}
		return http.StatusOK, api.TagTransactionsResponse{
			UpdatedCount: updated,
			Status:       fmt.Sprintf("updated %d selected transactions as %s", updated, ignoredStatusLabel(ignored)),
		}, nil
	}
	decoderFn := httpx.DecoderFunc[api.IgnoreTransactionsBySelectionRequest](func(r *http.Request) (api.IgnoreTransactionsBySelectionRequest, error) {
		return httpx.DecodeJSON[api.IgnoreTransactionsBySelectionRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}

// IgnoreTransactionsByFilter sets the ignored status for transactions matching the filter.
//
// @Summary     Ignore filtered cashflow transactions
// @Description Set ignored=true/false for all transactions that match the supplied filter.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.IgnoreTransactionsByFilterRequest true "Filter ignore request"
// @Success     200 {object} api.TagTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/ignore/filter [post]
// @Tags        Transactions
func IgnoreTransactionsByFilter(log logging.Logger, store *storage.SQLXBankTransactionStore) http.HandlerFunc {
	endpoint := func(ctx context.Context, req api.IgnoreTransactionsByFilterRequest) (status int, res any, err error) {
		query, parseErr := cashflowservice.BuildBulkTagQuery(toBulkTagFilters(req.Filters))
		if parseErr != nil {
			return http.StatusBadRequest, map[string]string{"filters": parseErr.Error()}, nil
		}

		ignored := true
		if req.Ignored != nil {
			ignored = *req.Ignored
		}
		updated, err := store.UpdateIgnoredByQuery(ctx, query, ignored)
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, api.TagTransactionsResponse{
			UpdatedCount: updated,
			Status:       fmt.Sprintf("updated %d filtered transactions as %s", updated, ignoredStatusLabel(ignored)),
		}, nil
	}
	decoderFn := httpx.DecoderFunc[api.IgnoreTransactionsByFilterRequest](func(r *http.Request) (api.IgnoreTransactionsByFilterRequest, error) {
		return httpx.DecodeJSON[api.IgnoreTransactionsByFilterRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}
