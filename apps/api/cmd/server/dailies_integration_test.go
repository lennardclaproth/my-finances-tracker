package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

func TestIntegration_GetDailiesEndpoint_TotalCountRespectsFiltersAndPagination(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()
	seedBrandNewDayManualProvider(t, app.db)

	listing, err := marketdata.NewListing("BND.TEST", "Brand New Day Test", marketdata.SourceBrandNewDay)
	if err != nil {
		t.Fatalf("failed creating listing seed: %v", err)
	}
	// Avoid triggering async sync for this deterministic test.
	accEnd := time.Now().UTC().AddDate(0, 0, 1)
	listing.AccumulatedEnd = &accEnd

	listingStore := storage.NewSQLXListingStore(app.db)
	if err := listingStore.Create(ctx, listing); err != nil {
		t.Fatalf("failed storing listing seed: %v", err)
	}

	dailyStore := storage.NewSQLXDailyStore(app.db)
	for _, day := range []time.Time{
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 4, 0, 0, 0, 0, time.UTC),
	} {
		d, err := marketdata.NewDaily(listing.Symbol, day, 10, 10, 10, 10, 0)
		if err != nil {
			t.Fatalf("failed creating daily seed: %v", err)
		}
		d.ListingID = listing.ID
		if err := dailyStore.Create(ctx, &d); err != nil {
			t.Fatalf("failed storing daily seed: %v", err)
		}
	}

	req, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/marketdata/dailies?symbol=BND.TEST&from=2026-02-02&to=2026-02-04&limit=1&offset=1",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /marketdata/dailies failed: %v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatalf("failed closing response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	var payload struct {
		Data []struct {
			Date string `json:"Date"`
		} `json:"Data"`
		Metadata struct {
			ResultCount int `json:"ResultCount"`
			TotalCount  int `json:"TotalCount"`
		} `json:"Metadata"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}

	if payload.Metadata.TotalCount != 3 {
		t.Fatalf("expected total_count 3 for filtered range, got %d", payload.Metadata.TotalCount)
	}
	if payload.Metadata.ResultCount != 1 {
		t.Fatalf("expected result_count 1 for paginated response, got %d", payload.Metadata.ResultCount)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected one row in page, got %d", len(payload.Data))
	}
}

func TestIntegration_GetDailiesEndpoint_ManualProviderDoesNotAutoSync(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()
	seedBrandNewDayManualProvider(t, app.db)

	listing, err := marketdata.NewListing("BND.MANUAL", "Brand New Day Manual", marketdata.SourceBrandNewDay)
	if err != nil {
		t.Fatalf("failed creating listing seed: %v", err)
	}
	if err := storage.NewSQLXListingStore(app.db).Create(ctx, listing); err != nil {
		t.Fatalf("failed storing listing seed: %v", err)
	}

	req, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/marketdata/dailies?symbol=BND.MANUAL&limit=5&offset=0",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /marketdata/dailies failed: %v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatalf("failed closing response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	var payload struct {
		Metadata struct {
			Message string `json:"Message"`
		} `json:"Metadata"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}
	if payload.Metadata.Message != "Manual provider configured; automatic sync disabled" {
		t.Fatalf("unexpected metadata message: %q", payload.Metadata.Message)
	}
}

func TestIntegration_GetDailiesEndpoint_ListingIDSelectsCorrectSourceWhenSymbolDuplicatesExist(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()
	seedBrandNewDayManualProvider(t, app.db)

	apiListing, err := marketdata.NewListing("DUP.AS", "API Source Listing", marketdata.SourceMarketStack)
	if err != nil {
		t.Fatalf("failed creating api listing seed: %v", err)
	}
	manualListing, err := marketdata.NewListing("DUP.AS", "Manual Source Listing", marketdata.SourceBrandNewDay)
	if err != nil {
		t.Fatalf("failed creating manual listing seed: %v", err)
	}

	listingStore := storage.NewSQLXListingStore(app.db)
	if err := listingStore.Create(ctx, apiListing); err != nil {
		t.Fatalf("failed storing api listing seed: %v", err)
	}
	if err := listingStore.Create(ctx, manualListing); err != nil {
		t.Fatalf("failed storing manual listing seed: %v", err)
	}

	dailyStore := storage.NewSQLXDailyStore(app.db)
	apiDaily, err := marketdata.NewDaily(apiListing.Symbol, time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), 10, 10, 10, 10, 0)
	if err != nil {
		t.Fatalf("failed creating api daily: %v", err)
	}
	apiDaily.ListingID = apiListing.ID
	if err := dailyStore.Create(ctx, &apiDaily); err != nil {
		t.Fatalf("failed storing api daily: %v", err)
	}

	manualDaily, err := marketdata.NewDaily(manualListing.Symbol, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), 20, 20, 20, 20, 0)
	if err != nil {
		t.Fatalf("failed creating manual daily: %v", err)
	}
	manualDaily.ListingID = manualListing.ID
	if err := dailyStore.Create(ctx, &manualDaily); err != nil {
		t.Fatalf("failed storing manual daily: %v", err)
	}

	req, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/marketdata/dailies?listing_id="+manualListing.ID.String()+"&limit=10&offset=0",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /marketdata/dailies failed: %v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatalf("failed closing response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	var payload struct {
		Data []struct {
			Date string `json:"Date"`
		} `json:"Data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected exactly one daily row for selected listing, got %d", len(payload.Data))
	}
	if payload.Data[0].Date[:10] != "2026-01-02" {
		t.Fatalf("expected manual listing daily date 2026-01-02, got %s", payload.Data[0].Date)
	}
}

func TestIntegration_GetDailiesEndpoint_SortOrderByDateDesc(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()
	seedBrandNewDayManualProvider(t, app.db)

	listing, err := marketdata.NewListing("BND.SORT", "Brand New Day Sort", marketdata.SourceBrandNewDay)
	if err != nil {
		t.Fatalf("failed creating listing seed: %v", err)
	}
	accEnd := time.Now().UTC().AddDate(0, 0, 1)
	listing.AccumulatedEnd = &accEnd

	listingStore := storage.NewSQLXListingStore(app.db)
	if err := listingStore.Create(ctx, listing); err != nil {
		t.Fatalf("failed storing listing seed: %v", err)
	}

	dailyStore := storage.NewSQLXDailyStore(app.db)
	for _, day := range []time.Time{
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
	} {
		d, err := marketdata.NewDaily(listing.Symbol, day, 10, 10, 10, 10, 0)
		if err != nil {
			t.Fatalf("failed creating daily seed: %v", err)
		}
		d.ListingID = listing.ID
		if err := dailyStore.Create(ctx, &d); err != nil {
			t.Fatalf("failed storing daily seed: %v", err)
		}
	}

	req, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/marketdata/dailies?symbol=BND.SORT&sort_order=desc&limit=1&offset=0",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /marketdata/dailies failed: %v", err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatalf("failed closing response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	var payload struct {
		Data []struct {
			Date string `json:"Date"`
		} `json:"Data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected one row, got %d", len(payload.Data))
	}
	if payload.Data[0].Date[:10] != "2026-02-03" {
		t.Fatalf("expected latest date first for desc sort, got %s", payload.Data[0].Date)
	}
}
