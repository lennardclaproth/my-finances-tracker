package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	cashflowservice "github.com/lennardclaproth/my-finances-tracker/internal/cashflow/service"
	httpx "github.com/lennardclaproth/my-finances-tracker/internal/http"
	"github.com/lennardclaproth/my-finances-tracker/internal/jobs"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

type manualCashflowTransactionCreator interface {
	CreateMany(ctx context.Context, input cashflowservice.ManualCashflowCreateInput) (*cashflowservice.ManualCashflowCreateResult, error)
}

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

// GetCashflowMonthlyAnalytics returns incoming, outgoing and net totals grouped per month.
//
// @Summary     Get cashflow monthly analytics
// @Description Returns monthly incoming, outgoing and net totals for the selected range.
// @Tags        Transactions
// @Accept      application/json
// @Produce     application/json
// @Param       from query string false "Start date (YYYY-MM-DD)"
// @Param       to query string false "End date (YYYY-MM-DD)"
// @Param       include_ignored query bool false "Include ignored transactions"
// @Success     200 {object} api.CashflowMonthlyAnalyticsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/analytics/monthly [get]
func GetCashflowMonthlyAnalytics(log logging.Logger, store *storage.SQLXBankTransactionStore) http.Handler {
	endpoint := func(ctx context.Context, req api.GetCashflowAnalyticsRequest) (status int, res any, err error) {
		from, to, parseErr := parseCashflowAnalyticsRange(req.From, req.To)
		if parseErr != nil {
			return http.StatusBadRequest, map[string]string{"date_range": parseErr.Error()}, nil
		}

		points, err := store.FetchMonthlyAnalytics(ctx, storage.CashflowAnalyticsQuery{
			From:           from,
			To:             to,
			IncludeIgnored: req.IncludeIgnored,
		})
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}

		out := make([]api.CashflowMonthlyAnalyticsPoint, 0, len(points))
		for _, point := range points {
			out = append(out, api.CashflowMonthlyAnalyticsPoint{
				Month:         point.Month.Format("2006-01-02"),
				IncomingCents: point.IncomingCents,
				OutgoingCents: point.OutgoingCents,
				NetCents:      point.NetCents,
			})
		}

		return http.StatusOK, api.CashflowMonthlyAnalyticsResponse{Data: out}, nil
	}

	decodeFn := httpx.DecoderFunc[api.GetCashflowAnalyticsRequest](func(r *http.Request) (api.GetCashflowAnalyticsRequest, error) {
		var req api.GetCashflowAnalyticsRequest
		res, err := httpx.DecodeQuery[api.GetCashflowAnalyticsRequest](r)
		if err != nil {
			return req, fmt.Errorf("GetCashflowMonthlyAnalytics failed to decode query: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
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
// @Success     200 {object} api.CashflowTagDistributionResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/analytics/tags [get]
func GetCashflowTagDistribution(log logging.Logger, store *storage.SQLXBankTransactionStore) http.Handler {
	endpoint := func(ctx context.Context, req api.GetCashflowAnalyticsRequest) (status int, res any, err error) {
		from, to, parseErr := parseCashflowAnalyticsRange(req.From, req.To)
		if parseErr != nil {
			return http.StatusBadRequest, map[string]string{"date_range": parseErr.Error()}, nil
		}

		dist, err := store.FetchTagDistribution(ctx, storage.CashflowAnalyticsQuery{
			From:           from,
			To:             to,
			IncludeIgnored: req.IncludeIgnored,
		})
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}

		toEntry := func(entries []storage.CashflowTagDistributionEntry) []api.CashflowTagDistributionEntry {
			out := make([]api.CashflowTagDistributionEntry, 0, len(entries))
			for _, entry := range entries {
				out = append(out, api.CashflowTagDistributionEntry{
					Tag:        entry.Tag,
					TotalCents: entry.TotalCents,
				})
			}
			return out
		}

		return http.StatusOK, api.CashflowTagDistributionResponse{
			Combined: toEntry(dist.Combined),
			Incoming: toEntry(dist.Incoming),
			Outgoing: toEntry(dist.Outgoing),
		}, nil
	}

	decodeFn := httpx.DecoderFunc[api.GetCashflowAnalyticsRequest](func(r *http.Request) (api.GetCashflowAnalyticsRequest, error) {
		var req api.GetCashflowAnalyticsRequest
		res, err := httpx.DecodeQuery[api.GetCashflowAnalyticsRequest](r)
		if err != nil {
			return req, fmt.Errorf("GetCashflowTagDistribution failed to decode query: %w", err)
		}
		return res, nil
	})

	return httpx.Endpoint(decodeFn, log, endpoint)
}

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

func toManualCashflowCreateTransactions(
	transactions []api.CreateManualCashflowTransactionEntryRequest,
) []cashflowservice.ManualCashflowCreateTransactionInput {
	out := make([]cashflowservice.ManualCashflowCreateTransactionInput, 0, len(transactions))
	for _, tx := range transactions {
		out = append(out, cashflowservice.ManualCashflowCreateTransactionInput{
			Date:        tx.Date,
			Amount:      tx.Amount,
			Type:        tx.Type,
			Description: tx.Description,
			Note:        tx.Note,
			Tag:         tx.Tag,
			Vendor:      tx.Vendor,
		})
	}
	return out
}

// TagTransaction applies a tag to an existing transaction.
//
// @Summary     Tag a transaction
// @Description Apply a tag to a transaction by id
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.TagTransactionRequest true "Tag request"
// @Success     200 {object} map[string]string "OK"
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/tag [post]
// @Tags        Transactions
func TagTransaction(log logging.Logger, tagger *storage.SQLXBankTransactionStore) http.HandlerFunc {
	// endpoint closure uses the injected tagger to construct the use-case handler
	endpoint := func(ctx context.Context, req api.TagTransactionRequest) (status int, res struct{}, err error) {
		err = tagger.Tag(ctx, req.Id, req.Tag)
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, struct{}{}, nil
	}
	decoderFn := httpx.DecoderFunc[api.TagTransactionRequest](func(r *http.Request) (api.TagTransactionRequest, error) {
		return httpx.DecodeJSON[api.TagTransactionRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}

// TagTransactionsBySelection applies a tag to a list of transaction IDs.
//
// @Summary     Tag selected cashflow transactions
// @Description Apply/overwrite tag for selected transaction IDs.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.TagTransactionsBySelectionRequest true "Selection tag request"
// @Success     200 {object} api.TagTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/tag/selection [post]
// @Tags        Transactions
func TagTransactionsBySelection(log logging.Logger, tagger *storage.SQLXBankTransactionStore) http.HandlerFunc {
	endpoint := func(ctx context.Context, req api.TagTransactionsBySelectionRequest) (status int, res api.TagTransactionsResponse, err error) {
		updated, err := tagger.UpdateTagByIDs(ctx, req.IDs, req.Tag)
		if err != nil {
			return http.StatusInternalServerError, api.TagTransactionsResponse{}, err
		}
		return http.StatusOK, api.TagTransactionsResponse{
			UpdatedCount: updated,
			Status:       fmt.Sprintf("updated %d selected transactions", updated),
		}, nil
	}
	decoderFn := httpx.DecoderFunc[api.TagTransactionsBySelectionRequest](func(r *http.Request) (api.TagTransactionsBySelectionRequest, error) {
		return httpx.DecodeJSON[api.TagTransactionsBySelectionRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}

// TagTransactionsByFilter applies a tag to all transactions matching the filter.
//
// @Summary     Tag filtered cashflow transactions
// @Description Apply/overwrite tag for all transactions that match the supplied filter.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.TagTransactionsByFilterRequest true "Filter tag request"
// @Success     200 {object} api.TagTransactionsResponse
// @Success     202 {object} api.TagTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/tag/filter [post]
// @Tags        Transactions
func TagTransactionsByFilter(
	log logging.Logger,
	tagger *storage.SQLXBankTransactionStore,
	enqueuer jobs.BulkTagEnqueuer,
) http.HandlerFunc {
	svc := cashflowservice.NewBulkTagService(tagger, enqueuer, cashflowservice.DefaultBulkTagAsyncCutoff)
	endpoint := func(ctx context.Context, req api.TagTransactionsByFilterRequest) (status int, res any, err error) {
		result, err := svc.TagByFilter(ctx, cashflowservice.BulkTagRequest{
			Tag:       req.Tag,
			AccountID: req.AccountID,
			Filters:   toBulkTagFilters(req.Filters),
		})
		if err != nil {
			if errors.Is(err, cashflowservice.ErrBulkTagFiltersInvalid) {
				return http.StatusBadRequest, map[string]string{"filters": err.Error()}, nil
			}
			return http.StatusInternalServerError, struct{}{}, err
		}
		if result.Mode == cashflowservice.TagByFilterModeAsync {
			return http.StatusAccepted, api.TagTransactionsResponse{
				UpdatedCount: 0,
				Status:       fmt.Sprintf("scheduled background bulk tag job for %d transactions", result.TotalMatched),
			}, nil
		}
		return http.StatusOK, api.TagTransactionsResponse{
			UpdatedCount: result.UpdatedCount,
			Status:       fmt.Sprintf("updated %d filtered transactions", result.UpdatedCount),
		}, nil
	}
	decoderFn := httpx.DecoderFunc[api.TagTransactionsByFilterRequest](func(r *http.Request) (api.TagTransactionsByFilterRequest, error) {
		return httpx.DecodeJSON[api.TagTransactionsByFilterRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}

func toBulkTagFilters(filters api.CashflowTagFilters) cashflowservice.BulkTagFilters {
	return cashflowservice.BulkTagFilters{
		Q:           filters.Q,
		Description: filters.Description,
		Note:        filters.Note,
		Source:      filters.Source,
		Direction:   filters.Direction,
		Tags:        filters.Tags,
		Untagged:    filters.Untagged,
		HideIgnored: filters.HideIgnored,
		From:        filters.From,
		To:          filters.To,
	}
}

// IgnoreTransactionsBySelection sets the ignored status for selected transaction IDs.
//
// @Summary     Ignore selected cashflow transactions
// @Description Set ignored=true/false for selected transaction IDs.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.IgnoreTransactionsBySelectionRequest true "Selection ignore request"
// @Success     200 {object} api.TagTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/ignore/selection [post]
// @Tags        Transactions
func IgnoreTransactionsBySelection(log logging.Logger, store *storage.SQLXBankTransactionStore) http.HandlerFunc {
	endpoint := func(ctx context.Context, req api.IgnoreTransactionsBySelectionRequest) (status int, res api.TagTransactionsResponse, err error) {
		ignored := true
		if req.Ignored != nil {
			ignored = *req.Ignored
		}
		updated, err := store.UpdateIgnoredByIDs(ctx, req.IDs, ignored)
		if err != nil {
			return http.StatusInternalServerError, api.TagTransactionsResponse{}, err
		}
		return http.StatusOK, api.TagTransactionsResponse{
			UpdatedCount: updated,
			Status:       fmt.Sprintf("updated %d selected transactions as %s", updated, ignoredStatusLabel(ignored)),
		}, nil
	}
	decoderFn := httpx.DecoderFunc[api.IgnoreTransactionsBySelectionRequest](func(r *http.Request) (api.IgnoreTransactionsBySelectionRequest, error) {
		return httpx.DecodeJSON[api.IgnoreTransactionsBySelectionRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}

// IgnoreTransactionsByFilter sets the ignored status for transactions matching the filter.
//
// @Summary     Ignore filtered cashflow transactions
// @Description Set ignored=true/false for all transactions that match the supplied filter.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body     api.IgnoreTransactionsByFilterRequest true "Filter ignore request"
// @Success     200 {object} api.TagTransactionsResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/ignore/filter [post]
// @Tags        Transactions
func IgnoreTransactionsByFilter(log logging.Logger, store *storage.SQLXBankTransactionStore) http.HandlerFunc {
	endpoint := func(ctx context.Context, req api.IgnoreTransactionsByFilterRequest) (status int, res any, err error) {
		query, parseErr := cashflowservice.BuildBulkTagQuery(toBulkTagFilters(req.Filters))
		if parseErr != nil {
			return http.StatusBadRequest, map[string]string{"filters": parseErr.Error()}, nil
		}

		ignored := true
		if req.Ignored != nil {
			ignored = *req.Ignored
		}
		updated, err := store.UpdateIgnoredByQuery(ctx, query, ignored)
		if err != nil {
			return http.StatusInternalServerError, struct{}{}, err
		}
		return http.StatusOK, api.TagTransactionsResponse{
			UpdatedCount: updated,
			Status:       fmt.Sprintf("updated %d filtered transactions as %s", updated, ignoredStatusLabel(ignored)),
		}, nil
	}
	decoderFn := httpx.DecoderFunc[api.IgnoreTransactionsByFilterRequest](func(r *http.Request) (api.IgnoreTransactionsByFilterRequest, error) {
		return httpx.DecodeJSON[api.IgnoreTransactionsByFilterRequest](r)
	})
	return httpx.Endpoint(decoderFn, log, endpoint)
}

func ignoredStatusLabel(ignored bool) string {
	if ignored {
		return "ignored"
	}
	return "not ignored"
}

func splitTags(tags string) []string {
	if strings.TrimSpace(tags) == "" {
		return nil
	}
	raw := strings.Split(tags, ",")
	out := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, entry := range raw {
		tag := strings.TrimSpace(entry)
		if tag == "" {
			continue
		}
		key := strings.ToLower(tag)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func normalizeDirectionFilter(raw string) (string, error) {
	direction := strings.ToLower(strings.TrimSpace(raw))
	if direction == "" {
		return "", nil
	}
	if direction != "in" && direction != "out" {
		return "", fmt.Errorf("direction must be either in or out")
	}
	return direction, nil
}

func parseCashflowAnalyticsRange(fromRaw, toRaw string) (*time.Time, *time.Time, error) {
	var from *time.Time
	if fromRaw != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromRaw)
		if err != nil {
			return nil, nil, fmt.Errorf("from must be in YYYY-MM-DD format")
		}
		from = &parsedFrom
	}

	var to *time.Time
	if toRaw != "" {
		parsedTo, err := time.Parse("2006-01-02", toRaw)
		if err != nil {
			return nil, nil, fmt.Errorf("to must be in YYYY-MM-DD format")
		}
		to = &parsedTo
	}

	if from != nil && to != nil && from.After(*to) {
		return nil, nil, fmt.Errorf("from must be before or equal to to")
	}

	return from, to, nil
}
