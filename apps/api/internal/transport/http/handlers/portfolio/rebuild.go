package portfolio

// RebuildPortfolio publishes an async rebuild request for a specific account.
//
// @Summary Rebuild portfolio
// @Description Triggers a portfolio rebuild by publishing an event to the internal bus.
// @Tags portfolio
// @Accept json
// @Produce json
// @Param request body api.RebuildPortfolioRequest true "Rebuild request payload"
// @Success 202 {object} api.AsyncEventAcceptedResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/rebuild [post]
func RebuildPortfolio(
	log logging.Logger,
	publisher bus.Bus,
	fetcher account.Fetcher,
) http.Handler {
	endpoint := func(ctx context.Context, req api.RebuildPortfolioRequest) (status int, res any, err error) {
		if publisher == nil {
			return http.StatusServiceUnavailable, map[string]string{"error": "event bus unavailable"}, nil
		}
		if _, err := fetcher.FetchByID(ctx, req.AccountID); err != nil {
			if errors.Is(err, account.ErrAccountNotFound) {
				return http.StatusNotFound, map[string]string{"account_id": account.ErrAccountNotFound.Error()}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		env, err := bus.NewJSONEnvelopeFromContext(ctx, api.PortfolioRebuildRequested{AccID: req.AccountID})
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, fmt.Errorf("failed to encode portfolio rebuild event: %w", err)
		}
		if err := publisher.Publish(ctx, env); err != nil {
			return http.StatusServiceUnavailable, map[string]string{"error": "failed to publish rebuild event"}, nil
		}
		return http.StatusAccepted, api.AsyncEventAcceptedResponse{
			MessageID: env.MessageID,
			Topic:     env.Topic,
			AccountID: req.AccountID,
		}, nil
	}

	decodeFn := httpx.DecoderFunc[api.RebuildPortfolioRequest](func(r *http.Request) (api.RebuildPortfolioRequest, error) {
		var req api.RebuildPortfolioRequest
		res, err := httpx.DecodeJSON[api.RebuildPortfolioRequest](r)
		if err != nil {
			return req, fmt.Errorf("RebuildPortfolio failed to decode request: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}
