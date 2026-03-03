package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

type testLogger struct{}

func (l *testLogger) Info(ctx context.Context, msg string, fields ...any) {}

func (l *testLogger) Error(ctx context.Context, msg string, err error, fields ...any) {}

var _ logging.Logger = (*testLogger)(nil)

type fakeBus struct {
	publishFn   func(ctx context.Context, env bus.Envelope) error
	lastEnv     bus.Envelope
	publishCall int
}

func (f *fakeBus) Publish(ctx context.Context, env bus.Envelope) error {
	f.publishCall++
	f.lastEnv = env
	if f.publishFn != nil {
		return f.publishFn(ctx, env)
	}
	return nil
}

func (f *fakeBus) Subscribe(eventType string, h bus.Handler) (bus.Subscription, error) {
	return nil, nil
}

func (f *fakeBus) Close() error { return nil }

type fakeAccountFetcher struct {
	fetchFn func(ctx context.Context, id uuid.UUID) (*account.Account, error)
}

func (f *fakeAccountFetcher) FetchByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	if f.fetchFn != nil {
		return f.fetchFn(ctx, id)
	}
	return nil, account.ErrAccountNotFound
}

type fakePortfolioSnapshotLister struct {
	listFn func(ctx context.Context, accountID uuid.UUID, from, to *time.Time) ([]*portfolio.PortfolioSnapshot, error)
}

func (f *fakePortfolioSnapshotLister) ListForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	from, to *time.Time,
) ([]*portfolio.PortfolioSnapshot, error) {
	if f.listFn != nil {
		return f.listFn(ctx, accountID, from, to)
	}
	return nil, nil
}

type fakePortfolioPositionLister struct {
	listFn func(ctx context.Context, accountID uuid.UUID, includeClosed bool) ([]*portfolio.PositionWithLatestSnapshot, error)
}

func (f *fakePortfolioPositionLister) ListForAccountWithLatestSnapshot(
	ctx context.Context,
	accountID uuid.UUID,
	includeClosed bool,
) ([]*portfolio.PositionWithLatestSnapshot, error) {
	if f.listFn != nil {
		return f.listFn(ctx, accountID, includeClosed)
	}
	return nil, nil
}

type fakePortfolioTransactionLister struct {
	listFn func(ctx context.Context, accountID uuid.UUID, from, to *time.Time) ([]portfolio.TransactionWithListingID, error)
}

func (f *fakePortfolioTransactionLister) FetchForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	from, to *time.Time,
) ([]portfolio.TransactionWithListingID, error) {
	if f.listFn != nil {
		return f.listFn(ctx, accountID, from, to)
	}
	return nil, nil
}

func TestRebuildPortfolio_PublishesRequestedEvent(t *testing.T) {
	accID := uuid.New()
	b := &fakeBus{}
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return &account.Account{ID: id, Name: "test"}, nil
		},
	}

	h := RebuildPortfolio(&testLogger{}, b, fetcher)
	body, _ := json.Marshal(api.RebuildPortfolioRequest{AccountID: accID})
	req := httptest.NewRequest(http.MethodPost, "/portfolio/rebuild", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rec.Code, rec.Body.String())
	}
	if b.publishCall != 1 {
		t.Fatalf("expected publish to be called once, got %d", b.publishCall)
	}

	var event api.PortfolioRebuildRequested
	if err := json.Unmarshal(b.lastEnv.Payload, &event); err != nil {
		t.Fatalf("failed decoding published payload: %v", err)
	}
	if event.AccID != accID {
		t.Fatalf("expected account id %s, got %s", accID, event.AccID)
	}

	var resp api.AsyncEventAcceptedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if resp.AccountID != accID {
		t.Fatalf("expected response account id %s, got %s", accID, resp.AccountID)
	}
	if resp.Topic != (api.PortfolioRebuildRequested{}).MessageTopic() {
		t.Fatalf("expected topic %s, got %s", (api.PortfolioRebuildRequested{}).MessageTopic(), resp.Topic)
	}
}

func TestRebuildPortfolio_UnknownAccountReturns400(t *testing.T) {
	b := &fakeBus{}
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return nil, account.ErrAccountNotFound
		},
	}
	h := RebuildPortfolio(&testLogger{}, b, fetcher)
	body, _ := json.Marshal(api.RebuildPortfolioRequest{AccountID: uuid.New()})
	req := httptest.NewRequest(http.MethodPost, "/portfolio/rebuild", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if b.publishCall != 0 {
		t.Fatalf("expected no publish call for unknown account")
	}
}

func TestRebuildPortfolio_PublishFailureReturns503(t *testing.T) {
	b := &fakeBus{
		publishFn: func(ctx context.Context, env bus.Envelope) error {
			return errors.New("bus unavailable")
		},
	}
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return &account.Account{ID: id, Name: "test"}, nil
		},
	}

	h := RebuildPortfolio(&testLogger{}, b, fetcher)
	body, _ := json.Marshal(api.RebuildPortfolioRequest{AccountID: uuid.New()})
	req := httptest.NewRequest(http.MethodPost, "/portfolio/rebuild", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRebuildPortfolio_InvalidRequestValidation(t *testing.T) {
	b := &fakeBus{}
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return &account.Account{ID: id, Name: "test", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}, nil
		},
	}
	h := RebuildPortfolio(&testLogger{}, b, fetcher)
	body, _ := json.Marshal(api.RebuildPortfolioRequest{AccountID: uuid.Nil})
	req := httptest.NewRequest(http.MethodPost, "/portfolio/rebuild", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetPortfolioSnapshots_UnknownAccountReturns400(t *testing.T) {
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return nil, account.ErrAccountNotFound
		},
	}
	lister := &fakePortfolioSnapshotLister{}
	h := GetPortfolioSnapshots(&testLogger{}, fetcher, lister)
	req := httptest.NewRequest(http.MethodGet, "/portfolio/snapshots?account_id="+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetPortfolioSnapshots_InvalidDateRangeReturns400(t *testing.T) {
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return &account.Account{ID: id, Name: "ok"}, nil
		},
	}
	lister := &fakePortfolioSnapshotLister{}
	accID := uuid.New()
	h := GetPortfolioSnapshots(&testLogger{}, fetcher, lister)
	req := httptest.NewRequest(http.MethodGet, "/portfolio/snapshots?account_id="+accID.String()+"&from=2026-02-01&to=2026-01-01", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetPortfolioPositions_UnknownAccountReturns400(t *testing.T) {
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return nil, account.ErrAccountNotFound
		},
	}
	lister := &fakePortfolioPositionLister{}
	h := GetPortfolioPositions(&testLogger{}, fetcher, lister)
	req := httptest.NewRequest(http.MethodGet, "/portfolio/positions?account_id="+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetPortfolioPositions_DefaultOpenOnlyAndIncludeClosed(t *testing.T) {
	accID := uuid.New()
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return &account.Account{ID: id, Name: "ok"}, nil
		},
	}
	lister := &fakePortfolioPositionLister{
		listFn: func(ctx context.Context, accountID uuid.UUID, includeClosed bool) ([]*portfolio.PositionWithLatestSnapshot, error) {
			if accountID != accID {
				t.Fatalf("unexpected account id: %s", accountID)
			}
			if includeClosed {
				return []*portfolio.PositionWithLatestSnapshot{
					{ID: uuid.New(), Quantity: 1, IsClosed: false},
					{ID: uuid.New(), Quantity: 0, IsClosed: true},
				}, nil
			}
			return []*portfolio.PositionWithLatestSnapshot{
				{ID: uuid.New(), Quantity: 1, IsClosed: false},
			}, nil
		},
	}
	h := GetPortfolioPositions(&testLogger{}, fetcher, lister)

	reqOpenOnly := httptest.NewRequest(http.MethodGet, "/portfolio/positions?account_id="+accID.String(), nil)
	recOpenOnly := httptest.NewRecorder()
	h.ServeHTTP(recOpenOnly, reqOpenOnly)
	if recOpenOnly.Code != http.StatusOK {
		t.Fatalf("expected 200 for open-only, got %d body=%s", recOpenOnly.Code, recOpenOnly.Body.String())
	}
	var openOnly api.PortfolioPositionsResponse
	if err := json.Unmarshal(recOpenOnly.Body.Bytes(), &openOnly); err != nil {
		t.Fatalf("failed decoding open-only payload: %v", err)
	}
	if openOnly.IncludeClosed {
		t.Fatalf("expected include_closed=false")
	}
	if len(openOnly.Data) != 1 || openOnly.Data[0].IsClosed {
		t.Fatalf("expected one open position, got %+v", openOnly.Data)
	}

	reqAll := httptest.NewRequest(http.MethodGet, "/portfolio/positions?account_id="+accID.String()+"&include_closed=true", nil)
	recAll := httptest.NewRecorder()
	h.ServeHTTP(recAll, reqAll)
	if recAll.Code != http.StatusOK {
		t.Fatalf("expected 200 for include_closed=true, got %d body=%s", recAll.Code, recAll.Body.String())
	}
	var all api.PortfolioPositionsResponse
	if err := json.Unmarshal(recAll.Body.Bytes(), &all); err != nil {
		t.Fatalf("failed decoding include_closed payload: %v", err)
	}
	if !all.IncludeClosed {
		t.Fatalf("expected include_closed=true")
	}
	if len(all.Data) != 2 {
		t.Fatalf("expected two positions, got %d", len(all.Data))
	}
}

func TestGetPortfolioTransactions_UnknownAccountReturns404(t *testing.T) {
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return nil, account.ErrAccountNotFound
		},
	}
	lister := &fakePortfolioTransactionLister{}
	h := GetPortfolioTransactions(&testLogger{}, fetcher, lister)
	req := httptest.NewRequest(http.MethodGet, "/portfolio/transactions?account_id="+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetPortfolioTransactions_MapsResponseAndDateRange(t *testing.T) {
	accID := uuid.New()
	listingID := uuid.New()
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return &account.Account{ID: id, Name: "ok"}, nil
		},
	}
	lister := &fakePortfolioTransactionLister{
		listFn: func(ctx context.Context, accountID uuid.UUID, from, to *time.Time) ([]portfolio.TransactionWithListingID, error) {
			if accountID != accID {
				t.Fatalf("unexpected account id: %s", accountID)
			}
			if from == nil || to == nil {
				t.Fatalf("expected from and to to be set")
			}
			if from.Format("2006-01-02") != "2026-01-01" {
				t.Fatalf("unexpected from: %s", from.Format("2006-01-02"))
			}
			if to.Format("2006-01-02") != "2026-01-31" {
				t.Fatalf("unexpected to date: %s", to.Format("2006-01-02"))
			}
			symbol := "VWCE"
			isin := "IE00BK5BQT80"
			now := time.Date(2026, 1, 31, 9, 0, 0, 0, time.UTC)
			return []portfolio.TransactionWithListingID{
				{
					Transaction: portfolio.Transaction{
						ID:          uuid.New(),
						AccountID:   &accID,
						Origin:      portfolio.TransactionOriginImport,
						Source:      "DEGIRO",
						OccurredAt:  now,
						Type:        portfolio.TxBuy,
						Symbol:      &symbol,
						ISIN:        &isin,
						Description: "Buy",
						Quantity:    2,
						UnitPrice:   1000000,
						AmountCents: 2000000,
						CreatedAt:   now,
						UpdatedAt:   now,
					},
					ListingID: &listingID,
				},
				{
					Transaction: portfolio.Transaction{
						ID:          uuid.New(),
						AccountID:   &accID,
						Origin:      portfolio.TransactionOriginManual,
						Source:      "DEGIRO",
						OccurredAt:  now,
						Type:        portfolio.TxCash,
						Description: "Cash withdrawal",
						Quantity:    -1,
						UnitPrice:   0,
						AmountCents: 5050000,
						CreatedAt:   now,
						UpdatedAt:   now,
					},
					ListingID: nil,
				},
			}, nil
		},
	}
	h := GetPortfolioTransactions(&testLogger{}, fetcher, lister)
	req := httptest.NewRequest(http.MethodGet, "/portfolio/transactions?account_id="+accID.String()+"&from=2026-01-01&to=2026-01-31", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var response api.PortfolioTransactionsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if len(response.Data) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(response.Data))
	}
	if response.Data[0].ListingID == nil || *response.Data[0].ListingID != listingID {
		t.Fatalf("expected first listing id %s, got %v", listingID, response.Data[0].ListingID)
	}
	if response.Data[0].Amount != "2" {
		t.Fatalf("expected BUY amount 2, got %s", response.Data[0].Amount)
	}
	if response.Data[1].Type != string(portfolio.TxCash) {
		t.Fatalf("expected CASH type, got %s", response.Data[1].Type)
	}
	if response.Data[1].Amount != "-5.05" {
		t.Fatalf("expected signed CASH amount -5.05, got %s", response.Data[1].Amount)
	}
}

func TestGetPortfolioTransactions_InvalidDateRangeReturns400(t *testing.T) {
	accID := uuid.New()
	fetcher := &fakeAccountFetcher{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*account.Account, error) {
			return &account.Account{ID: id, Name: "ok"}, nil
		},
	}
	lister := &fakePortfolioTransactionLister{}
	h := GetPortfolioTransactions(&testLogger{}, fetcher, lister)
	req := httptest.NewRequest(http.MethodGet, "/portfolio/transactions?account_id="+accID.String()+"&from=2026-02-01&to=2026-01-01", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func ptrString(v string) *string {
	return &v
}

func ptrTime(v time.Time) *time.Time {
	return &v
}
