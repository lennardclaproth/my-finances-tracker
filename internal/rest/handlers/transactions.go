package handlers

import (
	"context"

	"github.com/lennardclaproth/my-finances-tracker/api/contracts"
	"github.com/lennardclaproth/my-finances-tracker/internal/app/transactions"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/rest"

	"net/http"
)

// GetUntaggedTransactions returns a paginated list of transactions without tags.
//
// @Summary     Get untagged transactions
// @Description Retrieve paginated untagged transactions
// @Accept      json
// @Produce     application/json
// @Param       page      query     int  false "Page number"       default(1)
// @Param       page_size query     int  false "Page size"         default(20)
// @Success     200 {array} contracts.Transaction "List of transactions"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /transactions/untagged [get]
// @Tags        Transactions
func GetUntaggedTransactions(log logging.Logger, fetcher transactions.UntaggedTransactionFetcher) http.HandlerFunc {
	// endpoint closure uses the injected fetcher to construct the use-case handler
	endpoint := func(ctx context.Context, req contracts.GetUntaggedTransactionsRequest) (status int, res []contracts.Transaction, err error) {
		handler := transactions.NewUntaggedHandler(fetcher)
		transactions, err := handler.Handle(ctx, req.Page, req.PageSize)
		if err != nil {
			return http.StatusInternalServerError, nil, err
		}
		return http.StatusOK, transactions, nil
	}
	decoderFn := rest.DecoderFunc[contracts.GetUntaggedTransactionsRequest](func(r *http.Request) (contracts.GetUntaggedTransactionsRequest, error) {
		return rest.QueryDecoder[contracts.GetUntaggedTransactionsRequest](r)
	})

	return rest.Endpoint(decoderFn, log, endpoint)
}

// TagTransaction applies a tag to an existing transaction.
//
// @Summary     Tag a transaction
// @Description Apply a tag to a transaction by id
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     contracts.TagTransactionRequest true "Tag request"
// @Success     200 {object} map[string]string "OK"
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /transactions/tag [post]
// @Tags        Transactions
func TagTransaction(log logging.Logger, tagger transactions.TransactionTagger) http.HandlerFunc {
	// endpoint closure uses the injected tagger to construct the use-case handler
	endpoint := func(ctx context.Context, req contracts.TagTransactionRequest) (status int, res struct{}, err error) {
		handler := transactions.NewTagHandler(tagger)
		err = handler.Handle(ctx, req.Id, req.Tag)
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, struct{}{}, nil
	}
	decoderFn := rest.DecoderFunc[contracts.TagTransactionRequest](func(r *http.Request) (contracts.TagTransactionRequest, error) {
		return rest.JSONDecoder[contracts.TagTransactionRequest](r)
	})
	return rest.Endpoint(decoderFn, log, endpoint)
}
