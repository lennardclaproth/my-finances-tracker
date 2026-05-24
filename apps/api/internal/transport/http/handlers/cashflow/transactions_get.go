package cashflow

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/transport/http"
)

// GetTransactionsRequest contains filters, sorting, and pagination for cashflow transactions.
type GetTransactionsRequest struct {
	Limit       int    `query:"limit"`
	Offset      int    `query:"offset"`
	SortBy      string `query:"sort_by"`
	SortOrder   string `query:"sort_order"`
	Q           string `query:"q"`
	Description string `query:"description"`
	Note        string `query:"note"`
	Source      string `query:"source"`
	Direction   string `query:"direction"`
	Tags        string `query:"tags"`
	Untagged    bool   `query:"untagged"`
	HideIgnored bool   `query:"hide_ignored"`
	From        string `query:"from"`
	To          string `query:"to"`
}

// GetTransactionResponse represents one cashflow transaction in a list response.
type GetTransactionResponse struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	Note        string    `json:"note"`
	Source      string    `json:"source"`
	AmountCents int64     `json:"amountCents"`
	Direction   string    `json:"direction"`
	Date        time.Time `json:"date"`
	Tag         string    `json:"tag"`
	Ignored     bool      `json:"ignored"`
}

// PaginationResponse describes offset pagination metadata.
type PaginationResponse struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

// GetTransactionsResponse returns paginated cashflow transactions.
type GetTransactionsResponse struct {
	Pagination PaginationResponse          `json:"pagination"`
	Data       []CreateTransactionResponse `json:"data"`
}

// GetTransactions searches and filters cashflow transactions.
//
// @Summary     Search cashflow transactions
// @Description Filter cashflow transactions with explicit field filters and optional fuzzy search over description, note, and tag.
// @Tags        Transactions
// @Accept      application/json
// @Produce     application/json
// @Param       limit query int false "Page size"
// @Param       offset query int false "Offset"
// @Param       sort_by query string false "Sort field: date, description, note, tag, source, amount"
// @Param       sort_order query string false "Sort order: asc or desc"
// @Param       q query string false "Fuzzy query over description, note, tag"
// @Param       description query string false "Case-insensitive contains filter on description"
// @Param       note query string false "Case-insensitive contains filter on note"
// @Param       source query string false "Case-insensitive contains filter on source"
// @Param       direction query string false "Direction filter: in or out"
// @Param       tags query string false "Comma-separated tags, e.g. food,travel"
// @Param       untagged query bool false "Only untagged transactions (empty tag)"
// @Param       hide_ignored query bool false "Hide ignored transactions"
// @Param       from query string false "Start date (YYYY-MM-DD)"
// @Param       to query string false "End date (YYYY-MM-DD)"
// @Success     200 {object} TransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions [get]
func GetTransactions(log logging.Logger, queries *cashflow.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[GetTransactionsRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}

			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{
				"error": "invalid query parameters",
			})
			return
		}

		query, problems := toTransactionListQuery(req)
		if len(problems) > 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		result, err := queries.ListTransactions(r.Context(), query)
		if err != nil {
			log.Error(r.Context(), "cashflow transactions: failed to list transactions", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to get cashflow transactions",
			})
			return
		}

		items := make([]CreateTransactionResponse, 0, len(result.Transactions))
		for _, tx := range result.Transactions {
			items = append(items, CreateTransactionResponse{
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

		_ = httpx.JSONEncode(w, http.StatusOK, GetTransactionsResponse{
			Pagination: PaginationResponse{
				Limit:  query.Limit,
				Offset: query.Offset,
				Count:  len(items),
				Total:  result.Total,
			},
			Data: items,
		})
	})
}

func toTransactionListQuery(req GetTransactionsRequest) (cashflow.TransactionListQuery, map[string]string) {
	problems := make(map[string]string)

	if req.Limit < 0 {
		problems["limit"] = "limit must be greater than or equal to 0"
	}
	if req.Offset < 0 {
		problems["offset"] = "offset must be greater than or equal to 0"
	}

	sort, sortErr := cashflow.ParseTransactionSort(req.SortBy, req.SortOrder)
	if sortErr != nil {
		if strings.Contains(sortErr.Error(), "sort_order") {
			problems["sort_order"] = sortErr.Error()
		} else {
			problems["sort_by"] = sortErr.Error()
		}
	}
	direction, directionErr := cashflow.ParseDirection(req.Direction)
	if directionErr != nil {
		problems["direction"] = directionErr.Error()
	}
	from, to, dateErr := date.ParseFromTo(req.From, req.To)
	if dateErr != nil {
		problems["date_range"] = dateErr.Error()
	}

	limit := req.Limit
	if limit == 0 {
		limit = 100
	}

	return cashflow.TransactionListQuery{
		Limit:       limit,
		Offset:      req.Offset,
		Sort:        sort,
		Q:           req.Q,
		Description: req.Description,
		Note:        req.Note,
		Source:      req.Source,
		Direction:   direction,
		Tags:        cashflow.SplitTags(req.Tags),
		Untagged:    req.Untagged,
		HideIgnored: req.HideIgnored,
		From:        from,
		To:          to,
	}, problems
}
