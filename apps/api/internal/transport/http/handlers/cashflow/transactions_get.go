package cashflow

// GetCashflowTransactions searches and filters cashflow transactions.
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
// @Success     200 {object} api.CashflowTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions [get]
func GetCashflowTransactions(log logging.Logger, store *storage.SQLXBankTransactionStore) http.Handler {
	endpoint := func(ctx context.Context, req api.GetCashflowTransactionsRequest) (status int, res any, err error) {
		var from *time.Time
		if req.From != "" {
			parsedFrom, parseErr := time.Parse("2006-01-02", req.From)
			if parseErr != nil {
				return http.StatusBadRequest, map[string]string{"from": "from must be in YYYY-MM-DD format"}, nil
			}
			from = &parsedFrom
		}

		var to *time.Time
		if req.To != "" {
			parsedTo, parseErr := time.Parse("2006-01-02", req.To)
			if parseErr != nil {
				return http.StatusBadRequest, map[string]string{"to": "to must be in YYYY-MM-DD format"}, nil
			}
			to = &parsedTo
		}

		if from != nil && to != nil && from.After(*to) {
			return http.StatusBadRequest, map[string]string{"from": "from must be before or equal to to"}, nil
		}

		direction, directionErr := normalizeDirectionFilter(req.Direction)
		if directionErr != nil {
			return http.StatusBadRequest, map[string]string{"direction": directionErr.Error()}, nil
		}

		limit := req.Limit
		if limit == 0 {
			limit = 100
		}

		result, err := store.Fetch(ctx, storage.CashflowTransactionQuery{
			Limit:       limit,
			Offset:      req.Offset,
			SortBy:      req.SortBy,
			SortOrder:   req.SortOrder,
			Q:           req.Q,
			Description: req.Description,
			Note:        req.Note,
			Source:      req.Source,
			Direction:   direction,
			Tags:        splitTags(req.Tags),
			Untagged:    req.Untagged,
			HideIgnored: req.HideIgnored,
			From:        from,
			To:          to,
		})
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}

		items := make([]api.Transaction, 0, len(result.Transactions))
		for _, tx := range result.Transactions {
			items = append(items, api.Transaction{
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

		return http.StatusOK, api.CashflowTransactionsResponse{
			Pagination: api.Pagination{
				Limit:  limit,
				Offset: req.Offset,
				Count:  len(items),
				Total:  result.Total,
			},
			Data: items,
		}, nil
	}

	decodeFn := httpx.DecoderFunc[api.GetCashflowTransactionsRequest](func(r *http.Request) (api.GetCashflowTransactionsRequest, error) {
		var req api.GetCashflowTransactionsRequest
		res, err := httpx.DecodeQuery[api.GetCashflowTransactionsRequest](r)
		if err != nil {
			return req, fmt.Errorf("GetCashflowTransactions failed to decode query: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}
