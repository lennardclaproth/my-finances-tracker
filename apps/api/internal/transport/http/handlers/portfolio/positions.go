package portfolio

// GetPortfolioPositions returns positions (open-only by default) with their latest snapshot metrics.
//
// @Summary Get portfolio positions
// @Description Returns portfolio positions for the given account and include_closed filter.
// @Tags portfolio
// @Accept json
// @Produce json
// @Param account_id query string true "Account ID"
// @Param include_closed query bool false "Include closed positions"
// @Success 200 {object} api.PortfolioPositionsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/positions [get]
func GetPortfolioPositions(
	log logging.Logger,
	fetcher account.Fetcher,
	lister portfolioPositionLister,
) http.Handler {
	endpoint := func(ctx context.Context, req api.GetPortfolioPositionsRequest) (status int, res any, err error) {
		if _, err := fetcher.FetchByID(ctx, req.AccountID); err != nil {
			if errors.Is(err, account.ErrAccountNotFound) {
				return http.StatusNotFound, map[string]string{"account_id": account.ErrAccountNotFound.Error()}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}

		positions, err := lister.ListForAccountWithLatestSnapshot(ctx, req.AccountID, req.IncludeClosed)
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}

		out := make([]api.PortfolioPositionResponse, 0, len(positions))
		for _, pos := range positions {
			if pos == nil {
				continue
			}
			out = append(out, api.PortfolioPositionResponse{
				ID:               pos.ID,
				Symbol:           pos.Symbol,
				Name:             pos.Name,
				Quantity:         pos.Quantity,
				CostBasis:        int64(pos.CostBasis),
				RealizedPnL:      int64(pos.RealizedPnL),
				MarketValue:      pos.MarketValue,
				UnrealizedPnLPct: pos.UnrealizedPnLPct,
				LastSnapshotAt:   pos.LastSnapshotAt,
				OpenDate:         pos.OpenDate,
				CloseDate:        pos.CloseDate,
				IsClosed:         pos.IsClosed,
			})
		}

		return http.StatusOK, api.PortfolioPositionsResponse{
			IncludeClosed: req.IncludeClosed,
			Data:          out,
		}, nil
	}

	decodeFn := httpx.DecoderFunc[api.GetPortfolioPositionsRequest](func(r *http.Request) (api.GetPortfolioPositionsRequest, error) {
		var req api.GetPortfolioPositionsRequest
		res, err := httpx.DecodeQuery[api.GetPortfolioPositionsRequest](r)
		if err != nil {
			return req, fmt.Errorf("GetPortfolioPositions failed to decode query: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}
