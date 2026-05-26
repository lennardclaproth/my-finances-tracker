package cashflow

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/transport/http"
)

// IgnoreTransactionsBySelectionRequest toggles ignored-state for selected transaction IDs.
type IgnoreTransactionsBySelectionRequest struct {
	Ignored *bool       `json:"ignored"`
	IDs     []uuid.UUID `json:"ids"`
}

// IgnoreTransactionsByFilterRequest toggles ignored-state for transactions matching filters.
type IgnoreTransactionsByFilterRequest struct {
	Ignored *bool              `json:"ignored"`
	Filters TransactionFilters `json:"filters"`
}

// IgnoreTransactionsResponse returns the result of an ignore mutation operation.
type IgnoreTransactionsResponse struct {
	UpdatedCount int    `json:"updated_count"`
	Status       string `json:"status"`
}

func (r IgnoreTransactionsBySelectionRequest) isValid() map[string]string {
	problems := make(map[string]string)
	if len(r.IDs) == 0 {
		problems["ids"] = "ids is required"
		return problems
	}
	for i, id := range r.IDs {
		if id == uuid.Nil {
			problems[fmt.Sprintf("ids[%d]", i)] = "id must be a valid UUID"
		}
	}
	return problems
}

// IgnoreTransactionsBySelection sets the ignored status for selected transaction IDs.
//
// @Summary     Ignore selected cashflow transactions
// @Description Set ignored=true/false for selected transaction IDs.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body IgnoreTransactionsBySelectionRequest true "Selection ignore request"
// @Success     200 {object} IgnoreTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/ignore/selection [post]
// @Tags        Transactions
func IgnoreTransactionsBySelection(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[IgnoreTransactionsBySelectionRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}

		if problems := req.isValid(); len(problems) > 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		ignored := ignoreState(req.Ignored)
		updated, err := commands.IgnoreByIDs(r.Context(), req.IDs, ignored)
		if err != nil {
			log.Error(r.Context(), "cashflow ignore selection: failed to update transactions", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to ignore selected cashflow transactions"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, IgnoreTransactionsResponse{
			UpdatedCount: updated,
			Status:       fmt.Sprintf("updated %d selected transactions as %s", updated, ignoredStatusLabel(ignored)),
		})
	})
}

// IgnoreTransactionsByFilter sets the ignored status for transactions matching the filter.
//
// @Summary     Ignore filtered cashflow transactions
// @Description Set ignored=true/false for all transactions that match the supplied filter.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body IgnoreTransactionsByFilterRequest true "Filter ignore request"
// @Success     200 {object} IgnoreTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/ignore/filter [post]
// @Tags        Transactions
func IgnoreTransactionsByFilter(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[IgnoreTransactionsByFilterRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}

		filters, problems := req.Filters.ToAppFilters()
		if len(problems) > 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		ignored := ignoreState(req.Ignored)
		updated, err := commands.IgnoreByFilter(r.Context(), filters, ignored)
		if err != nil {
			log.Error(r.Context(), "cashflow ignore filter: failed to update transactions", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to ignore filtered cashflow transactions"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, IgnoreTransactionsResponse{
			UpdatedCount: updated,
			Status:       fmt.Sprintf("updated %d filtered transactions as %s", updated, ignoredStatusLabel(ignored)),
		})
	})
}

func ignoreState(value *bool) bool {
	if value == nil {
		return true
	}
	return *value
}

func ignoredStatusLabel(ignored bool) string {
	if ignored {
		return "ignored"
	}
	return "not ignored"
}
