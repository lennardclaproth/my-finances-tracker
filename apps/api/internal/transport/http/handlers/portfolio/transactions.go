package portfolio

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
// @Success 200 {object} api.PortfolioTransactionsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/transactions [get]
func GetPortfolioTransactions(
	log logging.Logger,
	fetcher account.Fetcher,
	lister portfolioTransactionLister,
) http.Handler {
	endpoint := func(ctx context.Context, req api.GetPortfolioTransactionsRequest) (status int, res any, err error) {
		limit := req.Limit
		if limit == 0 {
			limit = 25
		}
		sortBy := portfolio.NormalizeTransactionSortBy(req.SortBy)
		sortOrder := portfolio.NormalizeTransactionSortOrder(req.SortOrder)

		var txType *portfolio.TransactionType
		if normalizedType := strings.ToUpper(strings.TrimSpace(req.Type)); normalizedType != "" {
			typed := portfolio.TransactionType(normalizedType)
			txType = &typed
		}

		var origin *portfolio.TransactionOrigin
		if normalizedOrigin := strings.ToUpper(strings.TrimSpace(req.Origin)); normalizedOrigin != "" {
			typed := portfolio.TransactionOrigin(normalizedOrigin)
			origin = &typed
		}

		if _, err := fetcher.FetchByID(ctx, req.AccountID); err != nil {
			if errors.Is(err, account.ErrAccountNotFound) {
				return http.StatusNotFound, map[string]string{"account_id": account.ErrAccountNotFound.Error()}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}

		from, to, err := parsePortfolioDateRange(req.From, req.To)
		if err != nil {
			return http.StatusBadRequest, portfolioDateRangeProblem(err), nil
		}

		result, err := lister.FetchForAccount(ctx, portfolio.TransactionListQuery{
			AccountID: req.AccountID,
			From:      from,
			To:        to,
			Limit:     limit,
			Offset:    req.Offset,
			SortBy:    sortBy,
			SortOrder: sortOrder,
			Q:         strings.TrimSpace(req.Q),
			Type:      txType,
			Origin:    origin,
			Source:    strings.TrimSpace(req.Source),
			Listing:   strings.TrimSpace(req.Listing),
		})
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}

		out := make([]api.PortfolioTransactionResponse, 0, len(result.Transactions))
		for _, tx := range result.Transactions {
			accountID := uuid.Nil
			if tx.AccountID != nil {
				accountID = *tx.AccountID
			}
			out = append(out, api.PortfolioTransactionResponse{
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
				Amount:      formatDecimal(portfolio.SignedAmountForRead(tx.Type, tx.Quantity, tx.AmountCents.Float64())),
				Quantity:    formatDecimal(tx.Quantity),
				UnitPrice:   formatDecimal(tx.UnitPrice.Float64()),
				CreatedAt:   tx.CreatedAt,
				UpdatedAt:   tx.UpdatedAt,
			})
		}
		return http.StatusOK, api.PortfolioTransactionsResponse{
			Pagination: api.Pagination{
				Limit:  limit,
				Offset: req.Offset,
				Count:  len(out),
				Total:  result.Total,
			},
			Data: out,
		}, nil
	}

	decodeFn := httpx.DecoderFunc[api.GetPortfolioTransactionsRequest](func(r *http.Request) (api.GetPortfolioTransactionsRequest, error) {
		var req api.GetPortfolioTransactionsRequest
		res, err := httpx.DecodeQuery[api.GetPortfolioTransactionsRequest](r)
		if err != nil {
			return req, fmt.Errorf("GetPortfolioTransactions failed to decode query: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
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
