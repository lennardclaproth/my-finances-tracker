package cashflow

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// GetTagDistributionRequest contains date filters for cashflow tag analytics.
type GetTagDistributionRequest struct {
	From           string `query:"from"`
	To             string `query:"to"`
	IncludeIgnored bool   `query:"include_ignored"`
}

// TagDistributionEntryResponse represents one tag total in analytics output.
type TagDistributionEntryResponse struct {
	Tag        string `json:"tag"`
	TotalCents int64  `json:"totalCents"`
}

// TagDistributionResponse returns grouped tag totals by cashflow direction.
type TagDistributionResponse struct {
	Combined []TagDistributionEntryResponse `json:"combined"`
	Incoming []TagDistributionEntryResponse `json:"incoming"`
	Outgoing []TagDistributionEntryResponse `json:"outgoing"`
}

// TagTransactionRequest tags one transaction with a single tag.
type TagTransactionRequest struct {
	ID  uuid.UUID `json:"id"`
	Tag string    `json:"tag"`
}

// TagTransactionsBySelectionRequest applies a tag to selected transaction IDs.
type TagTransactionsBySelectionRequest struct {
	Tag string      `json:"tag"`
	IDs []uuid.UUID `json:"ids"`
}

// TagTransactionsByFilterRequest applies a tag to transactions matching filters.
type TagTransactionsByFilterRequest struct {
	Tag       string             `json:"tag"`
	AccountID *uuid.UUID         `json:"account_id,omitempty"`
	Filters   TransactionFilters `json:"filters"`
}

// TagTransactionsResponse returns the result of a tag mutation operation.
type TagTransactionsResponse struct {
	UpdatedCount int    `json:"updated_count"`
	Status       string `json:"status"`
}

func (r TagTransactionsBySelectionRequest) isValid() map[string]string {
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
// @Success     200 {object} TagDistributionResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/analytics/tags [get]
func GetCashflowTagDistribution(log logging.Logger, queries *cashflow.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[GetTagDistributionRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
			return
		}

		from, to, parseErr := date.ParseFromTo(req.From, req.To)
		if parseErr != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"date_range": parseErr.Error()})
			return
		}

		dist, err := queries.TagDistribution(r.Context(), cashflow.AnalyticsFilter{
			From:           from,
			To:             to,
			IncludeIgnored: req.IncludeIgnored,
		})
		if err != nil {
			log.Error(r.Context(), "cashflow tag distribution: failed to get analytics", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get cashflow tag distribution"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, TagDistributionResponse{
			Combined: toTagDistributionEntryResponse(dist.Combined),
			Incoming: toTagDistributionEntryResponse(dist.Incoming),
			Outgoing: toTagDistributionEntryResponse(dist.Outgoing),
		})
	})
}

// TagTransaction applies a tag to an existing transaction.
//
// @Summary     Tag a transaction
// @Description Apply a tag to a transaction by id
// @Accept      application/json
// @Produce     application/json
// @Param       payload body TagTransactionRequest true "Tag request"
// @Success     200 {object} map[string]string "OK"
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/tag [post]
// @Tags        Transactions
func TagTransaction(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[TagTransactionRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}

		if err := commands.TagByID(r.Context(), req.ID, req.Tag); err != nil {
			log.Error(r.Context(), "cashflow tag transaction: failed to update transaction", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to tag cashflow transaction"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, struct{}{})
	})
}

// TagTransactionsBySelection applies a tag to a list of transaction IDs.
//
// @Summary     Tag selected cashflow transactions
// @Description Apply/overwrite tag for selected transaction IDs.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body TagTransactionsBySelectionRequest true "Selection tag request"
// @Success     200 {object} TagTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/tag/selection [post]
// @Tags        Transactions
func TagTransactionsBySelection(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[TagTransactionsBySelectionRequest](r)
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

		updated, err := commands.TagByIDs(r.Context(), req.IDs, req.Tag)
		if err != nil {
			log.Error(r.Context(), "cashflow tag selection: failed to update transactions", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to tag selected cashflow transactions"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, TagTransactionsResponse{
			UpdatedCount: updated,
			Status:       fmt.Sprintf("updated %d selected transactions", updated),
		})
	})
}

// TagTransactionsByFilter applies a tag to all transactions matching the filter.
//
// @Summary     Tag filtered cashflow transactions
// @Description Apply/overwrite tag for all transactions that match the supplied filter.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body TagTransactionsByFilterRequest true "Filter tag request"
// @Success     200 {object} TagTransactionsResponse
// @Success     202 {object} TagTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/tag/filter [post]
// @Tags        Transactions
func TagTransactionsByFilter(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[TagTransactionsByFilterRequest](r)
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

		accountID := uuid.Nil
		if req.AccountID != nil {
			accountID = *req.AccountID
		}

		result, err := commands.TagByFilter(r.Context(), req.Tag, accountID, filters)
		if err != nil {
			log.Error(r.Context(), "cashflow tag filter: failed to update transactions", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to tag filtered cashflow transactions"})
			return
		}

		if result.Mode == cashflow.TagByFilterModeAsync {
			_ = httpx.JSONEncode(w, http.StatusAccepted, TagTransactionsResponse{
				UpdatedCount: 0,
				Status:       fmt.Sprintf("scheduled background bulk tag job for %d transactions", result.TotalMatched),
			})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, TagTransactionsResponse{
			UpdatedCount: result.UpdatedCount,
			Status:       fmt.Sprintf("updated %d filtered transactions", result.UpdatedCount),
		})
	})
}

func toTagDistributionEntryResponse(entries []cashflow.TagDistributionEntry) []TagDistributionEntryResponse {
	out := make([]TagDistributionEntryResponse, 0, len(entries))
	for _, entry := range entries {
		out = append(out, TagDistributionEntryResponse{
			Tag:        entry.Tag,
			TotalCents: entry.TotalCents,
		})
	}
	return out
}
