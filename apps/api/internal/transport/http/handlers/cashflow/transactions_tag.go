package cashflow

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/api"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

// GetCashflowTagDistribution returns grouped totals per tag for combined, incoming and outgoing transactions.
//
// @Summary     Get cashflow tag distribution
// @Description Returns tag distribution for combined, incoming and outgoing transactions in the selected range.
// @Tags        Transactions
// @Accept      application/json
// @Produce     application/json
// @Param       from query string false "Start date (YYYY-MM-DD)"
// @Param       to query string false "End date (YYYY-MM-DD)"
// @Param       include_ignored query bool false "Include ignored transactions"
// @Success     200 {object} api.CashflowTagDistributionResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/analytics/tags [get]
func GetCashflowTagDistribution(log logging.Logger, store *storage.SQLXBankTransactionStore) http.Handler {
	endpoint := func(ctx context.Context, req api.GetCashflowAnalyticsRequest) (status int, res any, err error) {
		from, to, parseErr := parseCashflowAnalyticsRange(req.From, req.To)
		if parseErr != nil {
			return http.StatusBadRequest, map[string]string{"date_range": parseErr.Error()}, nil
		}

		dist, err := store.FetchTagDistribution(ctx, storage.CashflowAnalyticsQuery{
			From:           from,
			To:             to,
			IncludeIgnored: req.IncludeIgnored,
		})
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}

		toEntry := func(entries []storage.CashflowTagDistributionEntry) []api.CashflowTagDistributionEntry {
			out := make([]api.CashflowTagDistributionEntry, 0, len(entries))
			for _, entry := range entries {
				out = append(out, api.CashflowTagDistributionEntry{
					Tag:        entry.Tag,
					TotalCents: entry.TotalCents,
				})
			}
			return out
		}

		return http.StatusOK, api.CashflowTagDistributionResponse{
			Combined: toEntry(dist.Combined),
			Incoming: toEntry(dist.Incoming),
			Outgoing: toEntry(dist.Outgoing),
		}, nil
	}

	decodeFn := httpx.DecoderFunc[api.GetCashflowAnalyticsRequest](func(r *http.Request) (api.GetCashflowAnalyticsRequest, error) {
		var req api.GetCashflowAnalyticsRequest
		res, err := httpx.DecodeQuery[api.GetCashflowAnalyticsRequest](r)
		if err != nil {
			return req, fmt.Errorf("GetCashflowTagDistribution failed to decode query: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}

// TagTransaction applies a tag to an existing transaction.
//
// @Summary     Tag a transaction
// @Description Apply a tag to a transaction by id
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.TagTransactionRequest true "Tag request"
// @Success     200 {object} map[string]string "OK"
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/tag [post]
// @Tags        Transactions
func TagTransaction(log logging.Logger, tagger *storage.SQLXBankTransactionStore) http.HandlerFunc {
	// endpoint closure uses the injected tagger to construct the use-case handler
	endpoint := func(ctx context.Context, req api.TagTransactionRequest) (status int, res struct{}, err error) {
		err = tagger.Tag(ctx, req.Id, req.Tag)
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, struct{}{}, nil
	}
	decoderFn := httpx.DecoderFunc[api.TagTransactionRequest](func(r *http.Request) (api.TagTransactionRequest, error) {
		return httpx.DecodeJSON[api.TagTransactionRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}

// TagTransactionsBySelection applies a tag to a list of transaction IDs.
//
// @Summary     Tag selected cashflow transactions
// @Description Apply/overwrite tag for selected transaction IDs.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.TagTransactionsBySelectionRequest true "Selection tag request"
// @Success     200 {object} api.TagTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/tag/selection [post]
// @Tags        Transactions
func TagTransactionsBySelection(log logging.Logger, tagger *storage.SQLXBankTransactionStore) http.HandlerFunc {
	endpoint := func(ctx context.Context, req api.TagTransactionsBySelectionRequest) (status int, res api.TagTransactionsResponse, err error) {
		updated, err := tagger.UpdateTagByIDs(ctx, req.IDs, req.Tag)
		if err != nil {
			return http.StatusInternalServerError, api.TagTransactionsResponse{}, err
		}
		return http.StatusOK, api.TagTransactionsResponse{
			UpdatedCount: updated,
			Status:       fmt.Sprintf("updated %d selected transactions", updated),
		}, nil
	}
	decoderFn := httpx.DecoderFunc[api.TagTransactionsBySelectionRequest](func(r *http.Request) (api.TagTransactionsBySelectionRequest, error) {
		return httpx.DecodeJSON[api.TagTransactionsBySelectionRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}

// TagTransactionsByFilter applies a tag to all transactions matching the filter.
//
// @Summary     Tag filtered cashflow transactions
// @Description Apply/overwrite tag for all transactions that match the supplied filter.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.TagTransactionsByFilterRequest true "Filter tag request"
// @Success     200 {object} api.TagTransactionsResponse
// @Success     202 {object} api.TagTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/tag/filter [post]
// @Tags        Transactions
func TagTransactionsByFilter(
	log logging.Logger,
	tagger *storage.SQLXBankTransactionStore,
	enqueuer jobs.BulkTagEnqueuer,
) http.HandlerFunc {
	svc := cashflowservice.NewBulkTagService(tagger, enqueuer, cashflowservice.DefaultBulkTagAsyncCutoff)
	endpoint := func(ctx context.Context, req api.TagTransactionsByFilterRequest) (status int, res any, err error) {
		result, err := svc.TagByFilter(ctx, cashflowservice.BulkTagRequest{
			Tag:       req.Tag,
			AccountID: req.AccountID,
			Filters:   toBulkTagFilters(req.Filters),
		})
		if err != nil {
			if errors.Is(err, cashflowservice.ErrBulkTagFiltersInvalid) {
				return http.StatusBadRequest, map[string]string{"filters": err.Error()}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		if result.Mode == cashflowservice.TagByFilterModeAsync {
			return http.StatusAccepted, api.TagTransactionsResponse{
				UpdatedCount: 0,
				Status:       fmt.Sprintf("scheduled background bulk tag job for %d transactions", result.TotalMatched),
			}, nil
		}
		return http.StatusOK, api.TagTransactionsResponse{
			UpdatedCount: result.UpdatedCount,
			Status:       fmt.Sprintf("updated %d filtered transactions", result.UpdatedCount),
		}, nil
	}
	decoderFn := httpx.DecoderFunc[api.TagTransactionsByFilterRequest](func(r *http.Request) (api.TagTransactionsByFilterRequest, error) {
		return httpx.DecodeJSON[api.TagTransactionsByFilterRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}
