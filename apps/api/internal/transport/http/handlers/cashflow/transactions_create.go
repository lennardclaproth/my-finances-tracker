package cashflow

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	cashflowservice "github.com/lennardclaproth/my-finances-tracker/internal/cashflow/service"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

// CreateManualCashflowTransactions creates manual cashflow transactions in bulk.
//
// @Summary     Create manual cashflow transactions
// @Description Creates one or more manual cashflow transactions for an account.
// @Tags        Transactions
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.CreateManualCashflowTransactionsRequest true "Manual cashflow transactions payload"
// @Success     201 {object} api.ManualCashflowTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     409 {object} map[string]string "Conflict"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/manual [post]
func CreateManualCashflowTransactions(
	log logging.Logger,
	svc manualCashflowTransactionCreator,
) http.Handler {
	endpoint := func(ctx context.Context, req api.CreateManualCashflowTransactionsRequest) (status int, res any, err error) {
		result, err := svc.CreateMany(ctx, cashflowservice.ManualCashflowCreateInput{
			AccountID: req.AccountID,
			Transactions: toManualCashflowCreateTransactions(
				req.Transactions,
			),
		})
		if err != nil {
			switch {
			case errors.Is(err, cashflowservice.ErrManualCashflowAccountNotFound):
				return http.StatusNotFound, map[string]string{"account_id": err.Error()}, nil
			case errors.Is(err, cashflowservice.ErrManualCashflowTransactionsRequired),
				errors.Is(err, cashflowservice.ErrManualCashflowTransactionLimitExceeded),
				errors.Is(err, cashflowservice.ErrManualCashflowInvalidDate),
				errors.Is(err, cashflowservice.ErrManualCashflowInvalidType),
				errors.Is(err, cashflowservice.ErrManualCashflowInvalidAmount),
				errors.Is(err, cashflowservice.ErrManualCashflowDescriptionRequired),
				errors.Is(err, cashflowservice.ErrManualCashflowNoteRequired),
				errors.Is(err, cashflowservice.ErrManualCashflowTagRequired):
				return http.StatusBadRequest, map[string]string{"transaction": err.Error()}, nil
			case errors.Is(err, cashflow.ErrDuplicateTransaction):
				return http.StatusConflict, map[string]string{"transaction": "duplicate transaction"}, nil
			default:
				return http.StatusInternalServerError, struct{}{}, err
			}
		}
		if result == nil {
			return http.StatusInternalServerError, struct{}{}, fmt.Errorf("manual cashflow create: empty result")
		}

		data := make([]api.Transaction, 0, len(result.Transactions))
		for _, tx := range result.Transactions {
			data = append(data, api.Transaction{
				ID:          tx.ID,
				Description: tx.Description,
				Note:        tx.Note,
				Source:      tx.Source,
				AmountCents: int64(tx.AmountCents),
				Direction:   string(tx.Direction),
				Date:        tx.Date,
				Tag:         tx.Tag,
				Ignored:     tx.Ignored,
			})
		}

		return http.StatusCreated, api.ManualCashflowTransactionsResponse{
			CreatedCount: len(data),
			Data:         data,
		}, nil
	}

	decodeFn := httpx.DecoderFunc[api.CreateManualCashflowTransactionsRequest](func(r *http.Request) (api.CreateManualCashflowTransactionsRequest, error) {
		var req api.CreateManualCashflowTransactionsRequest
		res, err := httpx.DecodeJSON[api.CreateManualCashflowTransactionsRequest](r)
		if err != nil {
			return req, fmt.Errorf("CreateManualCashflowTransactions failed to decode request: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}
