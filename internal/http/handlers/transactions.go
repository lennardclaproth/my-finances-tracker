package handlers

import (
	"context"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/api"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

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
// @Router      /transactions/tag [post]
// @Tags        Transactions
func TagTransaction(log logging.Logger, tagger *storage.SQLXTransactionStore) http.HandlerFunc {
	// endpoint closure uses the injected tagger to construct the use-case handler
	endpoint := func(ctx context.Context, req api.TagTransactionRequest) (status int, res struct{}, err error) {
		err = tagger.Tag(ctx, req.Id, req.Tag)
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, struct{}{}, nil
	}
	decoderFn := httpx.DecoderFunc[api.TagTransactionRequest](func(r *http.Request) (api.TagTransactionRequest, error) {
		return httpx.JSONDecoder[api.TagTransactionRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}
