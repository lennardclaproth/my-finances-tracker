package portfolio

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	portfoliodomain "github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

type portfolioTransactionLister interface {
	FetchForAccount(
		ctx context.Context,
		query portfoliodomain.TransactionListQuery,
	) (*portfoliodomain.TransactionListResult, error)
}

// GetPortfolioTransactionsRequest contains filters, sorting, and pagination for portfolio transactions.
type GetPortfolioTransactionsRequest struct {
	AccountID uuid.UUID `query:"account_id"`
	From      string    `query:"from"`
	To        string    `query:"to"`
	Limit     int       `query:"limit"`
	Offset    int       `query:"offset"`
	SortBy    string    `query:"sort_by"`
	SortOrder string    `query:"sort_order"`
	Q         string    `query:"q"`
	Type      string    `query:"type"`
	Origin    string    `query:"origin"`
	Source    string    `query:"source"`
	Listing   string    `query:"listing"`
}

func (r GetPortfolioTransactionsRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if r.Limit < 0 {
		problems["limit"] = "limit must be greater than or equal to 0"
	}
	if r.Limit != 0 && r.Limit != 10 && r.Limit != 25 && r.Limit != 50 && r.Limit != 100 {
		problems["limit"] = "limit must be one of: 10, 25, 50, 100"
	}
	if r.Offset < 0 {
		problems["offset"] = "offset must be greater than or equal to 0"
	}
	if r.SortBy != "" && strings.ToLower(strings.TrimSpace(r.SortBy)) != string(portfoliodomain.TransactionSortByDate) {
		problems["sort_by"] = "sort_by must be one of: date"
	}
	if r.SortOrder != "" {
		switch strings.ToLower(strings.TrimSpace(r.SortOrder)) {
		case string(portfoliodomain.TransactionSortOrderAsc), string(portfoliodomain.TransactionSortOrderDesc):
		default:
			problems["sort_order"] = "sort_order must be either asc or desc"
		}
	}
	if r.Type != "" {
		switch strings.ToUpper(strings.TrimSpace(r.Type)) {
		case string(portfoliodomain.TxBuy),
			string(portfoliodomain.TxSell),
			string(portfoliodomain.TxDividend),
			string(portfoliodomain.TxTax),
			string(portfoliodomain.TxFee),
			string(portfoliodomain.TxCash):
		default:
			problems["type"] = "type must be one of: BUY, SELL, DIVIDEND, TAX, FEE, CASH"
		}
	}
	if r.Origin != "" {
		switch strings.ToUpper(strings.TrimSpace(r.Origin)) {
		case string(portfoliodomain.TransactionOriginImport), string(portfoliodomain.TransactionOriginManual):
		default:
			problems["origin"] = "origin must be one of: IMPORT, MANUAL"
		}
	}
	return len(problems) == 0, problems
}

// CreateManualPortfolioTransactionRequest creates a manual portfolio transaction.
type CreateManualPortfolioTransactionRequest struct {
	AccountID   uuid.UUID  `json:"account_id"`
	VendorID    uuid.UUID  `json:"vendor_id"`
	OccurredAt  string     `json:"occurred_at"`
	Type        string     `json:"type"`
	ListingID   *uuid.UUID `json:"listing_id,omitempty"`
	Amount      string     `json:"amount"`
	Quantity    *string    `json:"quantity,omitempty"`
	Description *string    `json:"description,omitempty"`
}

func (r CreateManualPortfolioTransactionRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.AccountID == uuid.Nil {
		problems["account_id"] = "account_id is required"
	}
	if r.VendorID == uuid.Nil {
		problems["vendor_id"] = "vendor_id is required"
	}
	if strings.TrimSpace(r.OccurredAt) == "" {
		problems["occurred_at"] = "occurred_at is required"
	}
	if strings.TrimSpace(r.Type) == "" {
		problems["type"] = "type is required"
	}
	if strings.TrimSpace(r.Amount) == "" {
		problems["amount"] = "amount is required"
	}
	return len(problems) == 0, problems
}

// PaginationResponse describes offset pagination metadata.
type PaginationResponse struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

// PortfolioTransactionResponse represents one portfolio transaction in read APIs.
type PortfolioTransactionResponse struct {
	ID          uuid.UUID  `json:"id"`
	AccountID   uuid.UUID  `json:"account_id"`
	Origin      string     `json:"origin"`
	Source      string     `json:"source"`
	OccurredAt  time.Time  `json:"occurred_at"`
	Type        string     `json:"type"`
	ListingID   *uuid.UUID `json:"listing_id,omitempty"`
	ISIN        *string    `json:"isin,omitempty"`
	Symbol      *string    `json:"symbol,omitempty"`
	Description string     `json:"description"`
	Amount      string     `json:"amount"`
	Quantity    string     `json:"quantity"`
	UnitPrice   string     `json:"unit_price"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PortfolioTransactionsResponse returns paginated portfolio transactions.
type PortfolioTransactionsResponse struct {
	Pagination PaginationResponse             `json:"pagination"`
	Data       []PortfolioTransactionResponse `json:"data"`
}

// ManualPortfolioTransactionResponse represents a created manual portfolio transaction.
type ManualPortfolioTransactionResponse struct {
	ID          uuid.UUID  `json:"id"`
	AccountID   uuid.UUID  `json:"account_id"`
	Origin      string     `json:"origin"`
	Source      string     `json:"source"`
	OccurredAt  time.Time  `json:"occurred_at"`
	Type        string     `json:"type"`
	ListingID   *uuid.UUID `json:"listing_id,omitempty"`
	ISIN        *string    `json:"isin,omitempty"`
	Symbol      *string    `json:"symbol,omitempty"`
	Description string     `json:"description"`
	Amount      string     `json:"amount"`
	Quantity    string     `json:"quantity"`
	UnitPrice   string     `json:"unit_price"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// GetPortfolioTransactions returns transaction history for a portfolio account.
//
// @Summary Get portfolio transactions
// @Description Returns portfolio transactions for the given account and optional date range.
// @Tags portfolio
// @Accept json
// @Produce json
// @Param account_id query string true "Account ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param limit query int false "Page size (10, 25, 50, 100)"
// @Param offset query int false "Offset"
// @Param sort_by query string false "Sort field: date"
// @Param sort_order query string false "Sort order: asc or desc"
// @Param q query string false "Contains search over description, source, symbol, isin"
// @Param type query string false "Transaction type: BUY, SELL, DIVIDEND, TAX, FEE, CASH"
// @Param origin query string false "Transaction origin: IMPORT, MANUAL"
// @Param source query string false "Case-insensitive contains filter on source"
// @Param listing query string false "Case-insensitive contains filter on symbol/isin"
// @Success 200 {object} PortfolioTransactionsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/transactions [get]
func GetPortfolioTransactions(
	log logging.Logger,
	fetcher account.QueryStore,
	lister portfolioTransactionLister,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[GetPortfolioTransactionsRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
			return
		}
		if ok, problems := req.isValid(); !ok {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}
		if fetcher == nil {
			log.Error(r.Context(), "portfolio transactions: account fetcher is not configured", errors.New("account fetcher not configured"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get portfolio transactions"})
			return
		}
		if lister == nil {
			log.Error(r.Context(), "portfolio transactions: lister is not configured", errors.New("portfolio transaction lister not configured"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get portfolio transactions"})
			return
		}

		if _, err := fetcher.GetByID(r.Context(), req.AccountID); err != nil {
			if errors.Is(err, account.ErrAccountNotFound) {
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"account_id": account.ErrAccountNotFound.Error()})
				return
			}
			log.Error(r.Context(), "portfolio transactions: failed to fetch account", err, "account_id", req.AccountID.String())
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get portfolio transactions"})
			return
		}

		query, problems := toPortfolioTransactionListQuery(req)
		if len(problems) > 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		result, err := lister.FetchForAccount(r.Context(), query)
		if err != nil {
			log.Error(r.Context(), "portfolio transactions: failed to list transactions", err, "account_id", req.AccountID.String())
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get portfolio transactions"})
			return
		}
		if result == nil {
			result = &portfoliodomain.TransactionListResult{}
		}

		items := toPortfolioTransactionResponses(result.Transactions)
		_ = httpx.JSONEncode(w, http.StatusOK, PortfolioTransactionsResponse{
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

func toPortfolioTransactionListQuery(req GetPortfolioTransactionsRequest) (portfoliodomain.TransactionListQuery, map[string]string) {
	problems := make(map[string]string)
	limit := req.Limit
	if limit == 0 {
		limit = 25
	}

	var txType *portfoliodomain.TransactionType
	if normalizedType := strings.ToUpper(strings.TrimSpace(req.Type)); normalizedType != "" {
		typed := portfoliodomain.TransactionType(normalizedType)
		txType = &typed
	}

	var origin *portfoliodomain.TransactionOrigin
	if normalizedOrigin := strings.ToUpper(strings.TrimSpace(req.Origin)); normalizedOrigin != "" {
		typed := portfoliodomain.TransactionOrigin(normalizedOrigin)
		origin = &typed
	}

	from, to, err := date.ParseFromTo(req.From, req.To)
	if err != nil {
		problems["date_range"] = err.Error()
	}
	if to != nil {
		toEnd := date.EndOfDayUTC(*to)
		to = &toEnd
	}

	return portfoliodomain.TransactionListQuery{
		AccountID: req.AccountID,
		From:      from,
		To:        to,
		Limit:     limit,
		Offset:    req.Offset,
		SortBy:    portfoliodomain.NormalizeTransactionSortBy(req.SortBy),
		SortOrder: portfoliodomain.NormalizeTransactionSortOrder(req.SortOrder),
		Q:         strings.TrimSpace(req.Q),
		Type:      txType,
		Origin:    origin,
		Source:    strings.TrimSpace(req.Source),
		Listing:   strings.TrimSpace(req.Listing),
	}, problems
}

func toPortfolioTransactionResponses(rows []portfoliodomain.TransactionWithListingID) []PortfolioTransactionResponse {
	out := make([]PortfolioTransactionResponse, 0, len(rows))
	for _, tx := range rows {
		accountID := uuid.Nil
		if tx.AccountID != nil {
			accountID = *tx.AccountID
		}
		out = append(out, PortfolioTransactionResponse{
			ID:          tx.ID,
			AccountID:   accountID,
			Origin:      string(tx.Origin),
			Source:      tx.Source,
			OccurredAt:  tx.OccurredAt,
			Type:        string(tx.Type),
			ListingID:   tx.ListingID,
			ISIN:        tx.ISIN,
			Symbol:      tx.Symbol,
			Description: tx.Description,
			Amount:      formatDecimal(portfoliodomain.SignedAmountForRead(tx.Type, tx.Quantity, tx.AmountCents.Float64())),
			Quantity:    formatDecimal(tx.Quantity),
			UnitPrice:   formatDecimal(tx.UnitPrice.Float64()),
			CreatedAt:   tx.CreatedAt,
			UpdatedAt:   tx.UpdatedAt,
		})
	}
	return out
}

// CreateManualPortfolioTransaction creates a manual portfolio transaction without triggering rebuilds.
//
// @Summary Create manual portfolio transaction
// @Description Creates a manual portfolio transaction and persists it without publishing rebuild events.
// @Tags portfolio
// @Accept json
// @Produce json
// @Param request body CreateManualPortfolioTransactionRequest true "Manual portfolio transaction payload"
// @Success 201 {object} ManualPortfolioTransactionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/transactions/manual [post]
func CreateManualPortfolioTransaction(
	log logging.Logger,
	commands *portfoliodomain.Commands,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[CreateManualPortfolioTransactionRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		if ok, problems := req.isValid(); !ok {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}
		if commands == nil {
			log.Error(r.Context(), "manual portfolio transaction: creator is not configured", errors.New("manual portfolio transaction creator not configured"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create portfolio transaction"})
			return
		}

		result, err := commands.CreateTransaction(r.Context(), portfoliodomain.ManualTransactionInput{
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
			writeManualPortfolioTransactionError(w, r, log, err)
			return
		}
		if result == nil || result.Transaction == nil || result.Transaction.AccountID == nil {
			log.Error(r.Context(), "manual portfolio transaction: invalid create result", errors.New("manual transaction missing account or transaction"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create portfolio transaction"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusCreated, toManualPortfolioTransactionResponse(result))
	})
}

func writeManualPortfolioTransactionError(w http.ResponseWriter, r *http.Request, log logging.Logger, err error) {
	switch {
	case errors.Is(err, portfoliodomain.ErrManualAccountNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"account_id": err.Error()})
	case errors.Is(err, portfoliodomain.ErrManualVendorNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"vendor_id": err.Error()})
	case errors.Is(err, portfoliodomain.ErrManualListingNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"listing_id": err.Error()})
	case errors.Is(err, portfoliodomain.ErrManualVendorTypeNotSupported):
		_ = httpx.JSONEncode(w, http.StatusUnprocessableEntity, map[string]string{"vendor_id": err.Error()})
	case errors.Is(err, portfoliodomain.ErrDuplicateTransaction):
		_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{"transaction": "duplicate transaction"})
	case isManualValidationErr(err):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"transaction": err.Error()})
	default:
		log.Error(r.Context(), "manual portfolio transaction: failed to create transaction", err)
		_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create portfolio transaction"})
	}
}

func toManualPortfolioTransactionResponse(result *portfoliodomain.ManualTransactionCreateResult) ManualPortfolioTransactionResponse {
	tx := result.Transaction
	return ManualPortfolioTransactionResponse{
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
	}
}

func isManualValidationErr(err error) bool {
	return errors.Is(err, portfoliodomain.ErrManualVendorNotActive) ||
		errors.Is(err, portfoliodomain.ErrManualInvalidOccurredAt) ||
		errors.Is(err, portfoliodomain.ErrManualInvalidType) ||
		errors.Is(err, portfoliodomain.ErrManualInvalidAmount) ||
		errors.Is(err, portfoliodomain.ErrManualInvalidQuantity) ||
		errors.Is(err, portfoliodomain.ErrManualQuantityRequired) ||
		errors.Is(err, portfoliodomain.ErrManualQuantityForbidden) ||
		errors.Is(err, portfoliodomain.ErrManualListingRequired) ||
		errors.Is(err, portfoliodomain.ErrManualListingForbidden) ||
		errors.Is(err, portfoliodomain.ErrManualNonCashAmountMustBePos) ||
		errors.Is(err, portfoliodomain.ErrManualCashAmountMustBeNonZero) ||
		errors.Is(err, portfoliodomain.ErrManualQuantityMustBePositive) ||
		errors.Is(err, portfoliodomain.ErrManualListingIdentityMissing)
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
