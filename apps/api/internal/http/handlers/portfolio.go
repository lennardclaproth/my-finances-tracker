package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

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
				return http.StatusBadRequest, map[string]string{"account_id": "account not found"}, nil
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

type portfolioSnapshotLister interface {
	ListForAccount(ctx context.Context, accountID uuid.UUID, from, to *time.Time) ([]*portfolio.PortfolioSnapshot, error)
}

type portfolioPositionLister interface {
	ListForAccountWithLatestSnapshot(
		ctx context.Context,
		accountID uuid.UUID,
		includeClosed bool,
	) ([]*portfolio.PositionWithLatestSnapshot, error)
}

type portfolioTransactionLister interface {
	FetchForAccount(
		ctx context.Context,
		query portfolio.TransactionListQuery,
	) (*portfolio.TransactionListResult, error)
}

// GetPortfolioSnapshots returns snapshot series for charting the portfolio.
//
// @Summary Get portfolio snapshots
// @Description Returns snapshot points for the given account and optional date range.
// @Tags portfolio
// @Accept json
// @Produce json
// @Param account_id query string true "Account ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} api.PortfolioSnapshotPointResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/snapshots [get]
func GetPortfolioSnapshots(
	log logging.Logger,
	fetcher account.Fetcher,
	lister portfolioSnapshotLister,
) http.Handler {
	endpoint := func(ctx context.Context, req api.GetPortfolioSnapshotsRequest) (status int, res []api.PortfolioSnapshotPointResponse, err error) {
		if _, err := fetcher.FetchByID(ctx, req.AccountID); err != nil {
			if errors.Is(err, account.ErrAccountNotFound) {
				return http.StatusBadRequest, nil, nil
			}
			return http.StatusInternalServerError, nil, err
		}

		from, to, err := parsePortfolioDateRange(req.From, req.To)
		if err != nil {
			return http.StatusBadRequest, nil, nil
		}

		snapshots, err := lister.ListForAccount(ctx, req.AccountID, from, to)
		if err != nil {
			return http.StatusInternalServerError, nil, err
		}

		out := make([]api.PortfolioSnapshotPointResponse, 0, len(snapshots))
		for _, snap := range snapshots {
			if snap == nil {
				continue
			}
			returnVsCostBasisPct := 0.0
			if snap.CostBasis != 0 {
				returnVsCostBasisPct = (snap.TotalPnL.Float64() / snap.CostBasis.Float64()) * 100
			}
			valueIndex := 100.0 * (1 + (snap.TimeWeightedReturnPct / 100))
			out = append(out, api.PortfolioSnapshotPointResponse{
				OccurredAt:            snap.OccurredAt,
				MarketValue:           int64(snap.MarketValue),
				TotalPnL:              int64(snap.TotalPnL),
				TotalPnLPct:           snap.TotalPnLPct,
				TotalCostBasis:        int64(snap.CostBasis),
				ReturnVsCostBasisPct:  returnVsCostBasisPct,
				DailyReturnPct:        snap.DailyDeltaPnLPct,
				TimeWeightedReturnPct: snap.TimeWeightedReturnPct,
				ValueIndex:            valueIndex,
			})
		}
		return http.StatusOK, out, nil
	}

	decodeFn := httpx.DecoderFunc[api.GetPortfolioSnapshotsRequest](func(r *http.Request) (api.GetPortfolioSnapshotsRequest, error) {
		var req api.GetPortfolioSnapshotsRequest
		res, err := httpx.DecodeQuery[api.GetPortfolioSnapshotsRequest](r)
		if err != nil {
			return req, fmt.Errorf("GetPortfolioSnapshots failed to decode query: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}

func parsePortfolioDateRange(fromRaw, toRaw string) (*time.Time, *time.Time, error) {
	var from *time.Time
	if fromRaw != "" {
		d, err := time.Parse("2006-01-02", fromRaw)
		if err != nil {
			return nil, nil, fmt.Errorf("from must be in YYYY-MM-DD format")
		}
		v := d.UTC()
		from = &v
	}
	var to *time.Time
	if toRaw != "" {
		d, err := time.Parse("2006-01-02", toRaw)
		if err != nil {
			return nil, nil, fmt.Errorf("to must be in YYYY-MM-DD format")
		}
		v := d.UTC().AddDate(0, 0, 1).Add(-time.Nanosecond)
		to = &v
	}
	if from != nil && to != nil && from.After(*to) {
		return nil, nil, fmt.Errorf("from must be before or equal to to")
	}
	return from, to, nil
}

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
// @Failure 500 {object} map[string]string
// @Router /portfolio/positions [get]
func GetPortfolioPositions(
	log logging.Logger,
	fetcher account.Fetcher,
	lister portfolioPositionLister,
) http.Handler {
	endpoint := func(ctx context.Context, req api.GetPortfolioPositionsRequest) (status int, res api.PortfolioPositionsResponse, err error) {
		if _, err := fetcher.FetchByID(ctx, req.AccountID); err != nil {
			if errors.Is(err, account.ErrAccountNotFound) {
				return http.StatusBadRequest, api.PortfolioPositionsResponse{}, nil
			}
			return http.StatusInternalServerError, api.PortfolioPositionsResponse{}, err
		}

		positions, err := lister.ListForAccountWithLatestSnapshot(ctx, req.AccountID, req.IncludeClosed)
		if err != nil {
			return http.StatusInternalServerError, api.PortfolioPositionsResponse{}, err
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
	endpoint := func(ctx context.Context, req api.GetPortfolioTransactionsRequest) (status int, res api.PortfolioTransactionsResponse, err error) {
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
				return http.StatusNotFound, api.PortfolioTransactionsResponse{}, nil
			}
			return http.StatusInternalServerError, api.PortfolioTransactionsResponse{}, err
		}

		from, to, err := parsePortfolioDateRange(req.From, req.To)
		if err != nil {
			msg := err.Error()
			switch {
			case msg == "from must be in YYYY-MM-DD format":
				return http.StatusBadRequest, api.PortfolioTransactionsResponse{}, nil
			case msg == "to must be in YYYY-MM-DD format":
				return http.StatusBadRequest, api.PortfolioTransactionsResponse{}, nil
			default:
				return http.StatusBadRequest, api.PortfolioTransactionsResponse{}, nil
			}
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
			return http.StatusInternalServerError, api.PortfolioTransactionsResponse{}, err
		}

		out := make([]api.PortfolioTransactionResponse, 0, len(result.Transactions))
		for _, tx := range result.Transactions {
			amount := tx.AmountCents.Float64()
			if tx.Type == portfolio.TxCash && tx.Quantity < 0 {
				amount = -amount
			}
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
				Amount:      formatDecimal(amount),
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
