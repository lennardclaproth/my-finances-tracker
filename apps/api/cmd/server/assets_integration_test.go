package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/config"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

func TestIntegration_AssetsManualClassAndWorthMutations(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	acc, err := account.NewAccount("asset-account", nil)
	if err != nil {
		t.Fatalf("failed creating account seed: %v", err)
	}
	if err := storage.NewSQLXAccountStore(app.db).Create(ctx, acc); err != nil {
		t.Fatalf("failed storing account seed: %v", err)
	}

	createClassBody, _ := json.Marshal(api.CreateAssetClassRequest{
		AccountID: acc.ID,
		Name:      "Property",
	})
	createClassReq, err := http.NewRequest(http.MethodPost, app.server.URL+"/assets/classes", bytes.NewReader(createClassBody))
	if err != nil {
		t.Fatalf("failed creating class request: %v", err)
	}
	createClassReq.Header.Set("Content-Type", "application/json")
	createClassRes, err := http.DefaultClient.Do(createClassReq)
	if err != nil {
		t.Fatalf("POST /assets/classes failed: %v", err)
	}
	defer func() {
		if err := createClassRes.Body.Close(); err != nil {
			t.Fatalf("failed closing create class response body: %v", err)
		}
	}()
	if createClassRes.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(createClassRes.Body)
		t.Fatalf("expected status 201, got %d body=%s", createClassRes.StatusCode, string(raw))
	}
	var createdClass api.AssetClassResponse
	if err := json.NewDecoder(createClassRes.Body).Decode(&createdClass); err != nil {
		t.Fatalf("failed decoding create class response: %v", err)
	}

	createItemBody, _ := json.Marshal(api.CreateAssetItemRequest{
		AccountID:     acc.ID,
		ClassID:       createdClass.ID,
		Name:          "Apartment Amsterdam",
		InitialWorth:  "100000",
		EffectiveDate: "2026-03-01",
	})
	createItemReq, err := http.NewRequest(http.MethodPost, app.server.URL+"/assets/items", bytes.NewReader(createItemBody))
	if err != nil {
		t.Fatalf("failed creating item request: %v", err)
	}
	createItemReq.Header.Set("Content-Type", "application/json")
	createItemRes, err := http.DefaultClient.Do(createItemReq)
	if err != nil {
		t.Fatalf("POST /assets/items failed: %v", err)
	}
	defer func() {
		if err := createItemRes.Body.Close(); err != nil {
			t.Fatalf("failed closing create item response body: %v", err)
		}
	}()
	if createItemRes.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(createItemRes.Body)
		t.Fatalf("expected status 201, got %d body=%s", createItemRes.StatusCode, string(raw))
	}
	var createdItem api.AssetItemResponse
	if err := json.NewDecoder(createItemRes.Body).Decode(&createdItem); err != nil {
		t.Fatalf("failed decoding create item response: %v", err)
	}

	adjustBody, _ := json.Marshal(api.AdjustAssetItemWorthRequest{
		AccountID:     acc.ID,
		ClassID:       createdClass.ID,
		ItemID:        createdItem.ID,
		Direction:     "increase",
		Amount:        "5000",
		EffectiveDate: "2026-03-02",
	})
	adjustReq, err := http.NewRequest(http.MethodPost, app.server.URL+"/assets/items/worth/adjust", bytes.NewReader(adjustBody))
	if err != nil {
		t.Fatalf("failed creating adjust request: %v", err)
	}
	adjustReq.Header.Set("Content-Type", "application/json")
	adjustRes, err := http.DefaultClient.Do(adjustReq)
	if err != nil {
		t.Fatalf("POST /assets/items/worth/adjust failed: %v", err)
	}
	defer func() {
		if err := adjustRes.Body.Close(); err != nil {
			t.Fatalf("failed closing adjust response body: %v", err)
		}
	}()
	if adjustRes.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(adjustRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", adjustRes.StatusCode, string(raw))
	}

	setBody, _ := json.Marshal(api.SetAssetItemWorthRequest{
		AccountID:     acc.ID,
		ClassID:       createdClass.ID,
		ItemID:        createdItem.ID,
		Worth:         "125000",
		EffectiveDate: "2026-03-03",
	})
	setReq, err := http.NewRequest(http.MethodPost, app.server.URL+"/assets/items/worth/set", bytes.NewReader(setBody))
	if err != nil {
		t.Fatalf("failed creating set request: %v", err)
	}
	setReq.Header.Set("Content-Type", "application/json")
	setRes, err := http.DefaultClient.Do(setReq)
	if err != nil {
		t.Fatalf("POST /assets/items/worth/set failed: %v", err)
	}
	defer func() {
		if err := setRes.Body.Close(); err != nil {
			t.Fatalf("failed closing set response body: %v", err)
		}
	}()
	if setRes.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(setRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", setRes.StatusCode, string(raw))
	}

	classesReq, err := http.NewRequest(http.MethodGet, app.server.URL+"/assets/classes?account_id="+acc.ID.String(), nil)
	if err != nil {
		t.Fatalf("failed creating list classes request: %v", err)
	}
	classesRes, err := http.DefaultClient.Do(classesReq)
	if err != nil {
		t.Fatalf("GET /assets/classes failed: %v", err)
	}
	defer func() {
		if err := classesRes.Body.Close(); err != nil {
			t.Fatalf("failed closing list classes response body: %v", err)
		}
	}()
	if classesRes.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(classesRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", classesRes.StatusCode, string(raw))
	}
	var classRows []api.AssetClassResponse
	if err := json.NewDecoder(classesRes.Body).Decode(&classRows); err != nil {
		t.Fatalf("failed decoding classes response: %v", err)
	}
	if len(classRows) != 1 {
		t.Fatalf("expected one class row, got %d", len(classRows))
	}
	if classRows[0].CurrentWorth != "125000" {
		t.Fatalf("expected class current_worth=125000, got %s", classRows[0].CurrentWorth)
	}

	detailsReq, err := http.NewRequest(http.MethodGet, app.server.URL+"/assets/classes/"+createdClass.ID.String()+"?account_id="+acc.ID.String(), nil)
	if err != nil {
		t.Fatalf("failed creating details request: %v", err)
	}
	detailsRes, err := http.DefaultClient.Do(detailsReq)
	if err != nil {
		t.Fatalf("GET /assets/classes/{class_id} failed: %v", err)
	}
	defer func() {
		if err := detailsRes.Body.Close(); err != nil {
			t.Fatalf("failed closing details response body: %v", err)
		}
	}()
	if detailsRes.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(detailsRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", detailsRes.StatusCode, string(raw))
	}
	var details api.AssetClassDetailsResponse
	if err := json.NewDecoder(detailsRes.Body).Decode(&details); err != nil {
		t.Fatalf("failed decoding details response: %v", err)
	}
	if len(details.Items) != 1 {
		t.Fatalf("expected one class item, got %d", len(details.Items))
	}
	if len(details.History) != 3 {
		t.Fatalf("expected three history entries (initial, adjust, set), got %d", len(details.History))
	}
}

func TestIntegration_AssetsMutationsTriggerSnapshotsRebuildEvent(t *testing.T) {
	logger := logging.NewSlogLogger(slog.LevelDebug)
	db, _ := setupIntegrationDB(t, logger)
	cfg := &config.Config{
		DiskStorage: config.DiskStorage{
			BasePath: t.TempDir(),
		},
	}
	deps := newAppDependencies(logger, db, cfg)
	b, err := setupBus(logger, deps)
	if err != nil {
		t.Fatalf("failed setting up bus: %v", err)
	}
	router := setupRouterWithDeps(logger, deps, b, nil, nil)
	mux := http.NewServeMux()
	router.Register(mux)
	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
		if err := b.Close(); err != nil {
			t.Errorf("failed closing bus: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Errorf("failed closing db: %v", err)
		}
	})

	ctx := context.Background()
	createSvc := account.NewCreateService(deps.accountStore, b)
	acc, err := account.NewAccount("assets-snapshots-mutation-account", nil)
	if err != nil {
		t.Fatalf("failed creating account: %v", err)
	}
	if err := createSvc.Create(ctx, acc); err != nil {
		t.Fatalf("failed creating account with bus projection: %v", err)
	}

	createClassBody, _ := json.Marshal(api.CreateAssetClassRequest{
		AccountID: acc.ID,
		Name:      "Property",
	})
	createClassReq, err := http.NewRequest(http.MethodPost, server.URL+"/assets/classes", bytes.NewReader(createClassBody))
	if err != nil {
		t.Fatalf("failed creating class request: %v", err)
	}
	createClassReq.Header.Set("Content-Type", "application/json")
	createClassRes, err := http.DefaultClient.Do(createClassReq)
	if err != nil {
		t.Fatalf("POST /assets/classes failed: %v", err)
	}
	defer func() {
		if err := createClassRes.Body.Close(); err != nil {
			t.Fatalf("failed closing create class response body: %v", err)
		}
	}()
	if createClassRes.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(createClassRes.Body)
		t.Fatalf("expected status 201, got %d body=%s", createClassRes.StatusCode, string(raw))
	}
	var createdClass api.AssetClassResponse
	if err := json.NewDecoder(createClassRes.Body).Decode(&createdClass); err != nil {
		t.Fatalf("failed decoding create class response: %v", err)
	}

	createItemBody, _ := json.Marshal(api.CreateAssetItemRequest{
		AccountID:     acc.ID,
		ClassID:       createdClass.ID,
		Name:          "Apartment Amsterdam",
		InitialWorth:  "100000",
		EffectiveDate: "2026-03-01",
	})
	createItemReq, err := http.NewRequest(http.MethodPost, server.URL+"/assets/items", bytes.NewReader(createItemBody))
	if err != nil {
		t.Fatalf("failed creating item request: %v", err)
	}
	createItemReq.Header.Set("Content-Type", "application/json")
	createItemRes, err := http.DefaultClient.Do(createItemReq)
	if err != nil {
		t.Fatalf("POST /assets/items failed: %v", err)
	}
	defer func() {
		if err := createItemRes.Body.Close(); err != nil {
			t.Fatalf("failed closing create item response body: %v", err)
		}
	}()
	if createItemRes.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(createItemRes.Body)
		t.Fatalf("expected status 201, got %d body=%s", createItemRes.StatusCode, string(raw))
	}
	var createdItem api.AssetItemResponse
	if err := json.NewDecoder(createItemRes.Body).Decode(&createdItem); err != nil {
		t.Fatalf("failed decoding create item response: %v", err)
	}

	setBody, _ := json.Marshal(api.SetAssetItemWorthRequest{
		AccountID:     acc.ID,
		ClassID:       createdClass.ID,
		ItemID:        createdItem.ID,
		Worth:         "125000",
		EffectiveDate: "2026-03-03",
	})
	setReq, err := http.NewRequest(http.MethodPost, server.URL+"/assets/items/worth/set", bytes.NewReader(setBody))
	if err != nil {
		t.Fatalf("failed creating set request: %v", err)
	}
	setReq.Header.Set("Content-Type", "application/json")
	setRes, err := http.DefaultClient.Do(setReq)
	if err != nil {
		t.Fatalf("POST /assets/items/worth/set failed: %v", err)
	}
	defer func() {
		if err := setRes.Body.Close(); err != nil {
			t.Fatalf("failed closing set response body: %v", err)
		}
	}()
	if setRes.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(setRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", setRes.StatusCode, string(raw))
	}

	waitForAssetSnapshotsLatestWorth(t, server.URL, acc.ID, "125000")
}

func TestIntegration_AssetsPortfolioClassSyncOnPortfolioRebuiltEvent(t *testing.T) {
	logger := logging.NewSlogLogger(slog.LevelDebug)
	db, _ := setupIntegrationDB(t, logger)
	cfg := &config.Config{
		DiskStorage: config.DiskStorage{
			BasePath: t.TempDir(),
		},
	}
	deps := newAppDependencies(logger, db, cfg)
	b, err := setupBus(logger, deps)
	if err != nil {
		t.Fatalf("failed setting up bus: %v", err)
	}
	router := setupRouterWithDeps(logger, deps, b, nil, nil)
	mux := http.NewServeMux()
	router.Register(mux)
	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
		if err := b.Close(); err != nil {
			t.Errorf("failed closing bus: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Errorf("failed closing db: %v", err)
		}
	})

	ctx := context.Background()
	createSvc := account.NewCreateService(deps.accountStore, b)
	acc, err := account.NewAccount("portfolio-sync-asset-account", nil)
	if err != nil {
		t.Fatalf("failed creating account: %v", err)
	}
	if err := createSvc.Create(ctx, acc); err != nil {
		t.Fatalf("failed creating account with bus projection: %v", err)
	}

	snapshot := &portfolio.PortfolioSnapshot{
		ID:          uuid.New(),
		AccountID:   acc.ID,
		OccurredAt:  time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
		MarketValue: 333000000000,
		CostBasis:   300000000000,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := deps.portfolioSnapshotStore.CreateSnapshot(ctx, snapshot); err != nil {
		t.Fatalf("failed creating portfolio snapshot: %v", err)
	}
	snapshot2 := &portfolio.PortfolioSnapshot{
		ID:          uuid.New(),
		AccountID:   acc.ID,
		OccurredAt:  time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		MarketValue: 345000000000,
		CostBasis:   301000000000,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := deps.portfolioSnapshotStore.CreateSnapshot(ctx, snapshot2); err != nil {
		t.Fatalf("failed creating second portfolio snapshot: %v", err)
	}

	env, err := bus.NewJSONEnvelope(api.PortfolioRebuilt{AccID: acc.ID})
	if err != nil {
		t.Fatalf("failed creating portfolio rebuilt envelope: %v", err)
	}
	if err := b.Publish(ctx, env); err != nil {
		t.Fatalf("failed publishing portfolio rebuilt event: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/assets/classes?account_id="+acc.ID.String(), nil)
		if err != nil {
			t.Fatalf("failed creating classes request: %v", err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET /assets/classes failed: %v", err)
		}
		var rows []api.AssetClassResponse
		decodeErr := json.NewDecoder(res.Body).Decode(&rows)
		closeErr := res.Body.Close()
		if closeErr != nil {
			t.Fatalf("failed closing classes response body: %v", closeErr)
		}
		if res.StatusCode == http.StatusOK && decodeErr == nil && len(rows) == 1 && rows[0].Source == string(assets.ClassSourcePortfolio) {
			if rows[0].CurrentWorth != "345000" {
				t.Fatalf("expected portfolio class current_worth=345000, got %s", rows[0].CurrentWorth)
			}

			detailsReq, err := http.NewRequest(
				http.MethodGet,
				server.URL+"/assets/classes/"+rows[0].ID.String()+"?account_id="+acc.ID.String(),
				nil,
			)
			if err != nil {
				t.Fatalf("failed creating class details request: %v", err)
			}
			detailsRes, err := http.DefaultClient.Do(detailsReq)
			if err != nil {
				t.Fatalf("GET /assets/classes/{class_id} failed: %v", err)
			}
			var details api.AssetClassDetailsResponse
			detailsDecodeErr := json.NewDecoder(detailsRes.Body).Decode(&details)
			detailsCloseErr := detailsRes.Body.Close()
			if detailsCloseErr != nil {
				t.Fatalf("failed closing class details response body: %v", detailsCloseErr)
			}
			if detailsRes.StatusCode != http.StatusOK {
				t.Fatalf("expected details status 200, got %d", detailsRes.StatusCode)
			}
			if detailsDecodeErr != nil {
				t.Fatalf("failed decoding class details response: %v", detailsDecodeErr)
			}
			if len(details.History) != 2 {
				t.Fatalf("expected portfolio class history rebuilt from snapshots (2 entries), got %d", len(details.History))
			}
			if len(details.Growth) != 2 {
				t.Fatalf("expected 2 growth points from snapshots, got %d", len(details.Growth))
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for portfolio class sync (status=%d rows=%d decode_err=%v)", res.StatusCode, len(rows), decodeErr)
		}
		time.Sleep(50 * time.Millisecond)
	}

	waitForAssetSnapshotsLatestWorth(t, server.URL, acc.ID, "345000")
}

func TestIntegration_GetAssetSnapshots_Validation(t *testing.T) {
	app := newIntegrationApp(t)

	missingAccountReq, err := http.NewRequest(http.MethodGet, app.server.URL+"/assets/snapshots", nil)
	if err != nil {
		t.Fatalf("failed creating missing-account request: %v", err)
	}
	missingAccountRes, err := http.DefaultClient.Do(missingAccountReq)
	if err != nil {
		t.Fatalf("GET /assets/snapshots without account_id failed: %v", err)
	}
	defer func() {
		if err := missingAccountRes.Body.Close(); err != nil {
			t.Fatalf("failed closing missing-account response body: %v", err)
		}
	}()
	if missingAccountRes.StatusCode != http.StatusBadRequest {
		raw, _ := io.ReadAll(missingAccountRes.Body)
		t.Fatalf("expected status 400 for missing account_id, got %d body=%s", missingAccountRes.StatusCode, string(raw))
	}

	invalidFromReq, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/assets/snapshots?account_id="+uuid.New().String()+"&from=2026/03/01",
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating invalid-from request: %v", err)
	}
	invalidFromRes, err := http.DefaultClient.Do(invalidFromReq)
	if err != nil {
		t.Fatalf("GET /assets/snapshots with invalid from failed: %v", err)
	}
	defer func() {
		if err := invalidFromRes.Body.Close(); err != nil {
			t.Fatalf("failed closing invalid-from response body: %v", err)
		}
	}()
	if invalidFromRes.StatusCode != http.StatusBadRequest {
		raw, _ := io.ReadAll(invalidFromRes.Body)
		t.Fatalf("expected status 400 for invalid from date, got %d body=%s", invalidFromRes.StatusCode, string(raw))
	}

	var payload map[string]string
	if err := json.NewDecoder(invalidFromRes.Body).Decode(&payload); err != nil {
		t.Fatalf("failed decoding invalid-from response: %v", err)
	}
	if payload["from"] != "from must be in YYYY-MM-DD format" {
		t.Fatalf("expected from field error, got %+v", payload)
	}
}

func TestIntegration_AssetsClassDetailsGrowthUsesLatestHistoryWindow(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	acc, err := account.NewAccount("asset-growth-window-account", nil)
	if err != nil {
		t.Fatalf("failed creating account seed: %v", err)
	}
	if err := storage.NewSQLXAccountStore(app.db).Create(ctx, acc); err != nil {
		t.Fatalf("failed storing account seed: %v", err)
	}

	createClassBody, _ := json.Marshal(api.CreateAssetClassRequest{
		AccountID: acc.ID,
		Name:      "Long Growth",
	})
	createClassReq, err := http.NewRequest(http.MethodPost, app.server.URL+"/assets/classes", bytes.NewReader(createClassBody))
	if err != nil {
		t.Fatalf("failed creating class request: %v", err)
	}
	createClassReq.Header.Set("Content-Type", "application/json")
	createClassRes, err := http.DefaultClient.Do(createClassReq)
	if err != nil {
		t.Fatalf("POST /assets/classes failed: %v", err)
	}
	defer func() {
		if err := createClassRes.Body.Close(); err != nil {
			t.Fatalf("failed closing class response body: %v", err)
		}
	}()
	if createClassRes.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(createClassRes.Body)
		t.Fatalf("expected status 201, got %d body=%s", createClassRes.StatusCode, string(raw))
	}
	var createdClass api.AssetClassResponse
	if err := json.NewDecoder(createClassRes.Body).Decode(&createdClass); err != nil {
		t.Fatalf("failed decoding create class response: %v", err)
	}

	createItemBody, _ := json.Marshal(api.CreateAssetItemRequest{
		AccountID:     acc.ID,
		ClassID:       createdClass.ID,
		Name:          "Long Growth Item",
		InitialWorth:  "1",
		EffectiveDate: "2020-01-01",
	})
	createItemReq, err := http.NewRequest(http.MethodPost, app.server.URL+"/assets/items", bytes.NewReader(createItemBody))
	if err != nil {
		t.Fatalf("failed creating item request: %v", err)
	}
	createItemReq.Header.Set("Content-Type", "application/json")
	createItemRes, err := http.DefaultClient.Do(createItemReq)
	if err != nil {
		t.Fatalf("POST /assets/items failed: %v", err)
	}
	defer func() {
		if err := createItemRes.Body.Close(); err != nil {
			t.Fatalf("failed closing item response body: %v", err)
		}
	}()
	if createItemRes.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(createItemRes.Body)
		t.Fatalf("expected status 201, got %d body=%s", createItemRes.StatusCode, string(raw))
	}
	var createdItem api.AssetItemResponse
	if err := json.NewDecoder(createItemRes.Body).Decode(&createdItem); err != nil {
		t.Fatalf("failed decoding create item response: %v", err)
	}

	assetStore := storage.NewSQLXAssetStore(app.db)
	previousWorth, err := money.NewPrice(1)
	if err != nil {
		t.Fatalf("failed creating initial previous worth: %v", err)
	}
	startDate := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	const historyRows = 1205
	for i := 0; i < historyRows; i++ {
		worthValue := float64(1000 + i)
		nextWorth, err := money.NewPrice(worthValue)
		if err != nil {
			t.Fatalf("failed creating worth for i=%d: %v", i, err)
		}
		effectiveDate := startDate.AddDate(0, 0, i)
		createdAt := effectiveDate.Add(12 * time.Hour)

		if err := assetStore.WithTx(ctx, func(txCtx context.Context) error {
			if err := assetStore.UpdateItemWorth(txCtx, acc.ID, createdClass.ID, createdItem.ID, nextWorth); err != nil {
				return err
			}
			return assetStore.CreateHistory(txCtx, &assets.Mutation{
				ID:              uuid.New(),
				AccountID:       acc.ID,
				ClassID:         createdClass.ID,
				AssetID:         createdItem.ID,
				ChangeType:      assets.ChangeTypeSet,
				Direction:       nil,
				Amount:          nextWorth,
				PreviousWorth:   previousWorth,
				NewWorth:        nextWorth,
				ClassTotalWorth: nextWorth,
				EffectiveDate:   effectiveDate,
				Note:            "seeded for growth-window test",
				CreatedAt:       createdAt,
			})
		}); err != nil {
			t.Fatalf("failed seeding history row %d: %v", i, err)
		}
		previousWorth = nextWorth
	}

	detailsReq, err := http.NewRequest(
		http.MethodGet,
		app.server.URL+"/assets/classes/"+createdClass.ID.String()+"?account_id="+acc.ID.String(),
		nil,
	)
	if err != nil {
		t.Fatalf("failed creating details request: %v", err)
	}
	detailsRes, err := http.DefaultClient.Do(detailsReq)
	if err != nil {
		t.Fatalf("GET /assets/classes/{class_id} failed: %v", err)
	}
	defer func() {
		if err := detailsRes.Body.Close(); err != nil {
			t.Fatalf("failed closing details response body: %v", err)
		}
	}()
	if detailsRes.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(detailsRes.Body)
		t.Fatalf("expected status 200, got %d body=%s", detailsRes.StatusCode, string(raw))
	}
	var details api.AssetClassDetailsResponse
	if err := json.NewDecoder(detailsRes.Body).Decode(&details); err != nil {
		t.Fatalf("failed decoding details response: %v", err)
	}
	if len(details.Growth) == 0 {
		t.Fatal("expected non-empty growth data")
	}

	lastPoint := details.Growth[len(details.Growth)-1]
	expectedDate := startDate.AddDate(0, 0, historyRows-1).Format("2006-01-02")
	expectedWorth := strconv.Itoa(1000 + historyRows - 1)
	if lastPoint.Date != expectedDate {
		t.Fatalf("expected latest growth date %s, got %s", expectedDate, lastPoint.Date)
	}
	if lastPoint.TotalWorth != expectedWorth {
		t.Fatalf("expected latest growth worth %s, got %s", expectedWorth, lastPoint.TotalWorth)
	}
}

func TestIntegration_AssetSnapshotsStoreUpsertUniquenessAndRangeOrdering(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	acc, err := account.NewAccount("asset-snapshots-store-account", nil)
	if err != nil {
		t.Fatalf("failed creating account seed: %v", err)
	}
	if err := storage.NewSQLXAccountStore(app.db).Create(ctx, acc); err != nil {
		t.Fatalf("failed storing account seed: %v", err)
	}

	assetStore := storage.NewSQLXAssetStore(app.db)
	if err := assetStore.EnsureAccount(ctx, assets.NewAccount(acc.ID)); err != nil {
		t.Fatalf("failed ensuring assets account projection: %v", err)
	}

	dayOne := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)
	dayTwo := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	firstWorth, err := money.NewPrice(100)
	if err != nil {
		t.Fatalf("failed creating first snapshot worth: %v", err)
	}
	replacementWorth, err := money.NewPrice(150)
	if err != nil {
		t.Fatalf("failed creating replacement snapshot worth: %v", err)
	}
	secondWorth, err := money.NewPrice(180)
	if err != nil {
		t.Fatalf("failed creating second snapshot worth: %v", err)
	}
	now := time.Now().UTC()

	if err := assetStore.UpsertSnapshots(ctx, []*assets.Snapshot{
		{
			ID:         uuid.New(),
			AccountID:  acc.ID,
			OccurredAt: dayOne,
			TotalWorth: firstWorth,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         uuid.New(),
			AccountID:  acc.ID,
			OccurredAt: dayOne,
			TotalWorth: replacementWorth,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         uuid.New(),
			AccountID:  acc.ID,
			OccurredAt: dayTwo,
			TotalWorth: secondWorth,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}); err != nil {
		t.Fatalf("failed upserting snapshots: %v", err)
	}

	from := dayOne
	to := dayTwo
	rows, err := assetStore.ListSnapshotsForAccount(ctx, acc.ID, &from, &to)
	if err != nil {
		t.Fatalf("failed listing snapshots: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 unique snapshot days, got %d", len(rows))
	}
	if rows[0].OccurredAt.Format("2006-01-02") != "2026-03-08" || rows[0].TotalWorth != replacementWorth {
		t.Fatalf("expected first row to be replaced day-one snapshot with worth %d, got date=%s worth=%d", replacementWorth, rows[0].OccurredAt.Format("2006-01-02"), rows[0].TotalWorth)
	}
	if rows[1].OccurredAt.Format("2006-01-02") != "2026-03-09" || rows[1].TotalWorth != secondWorth {
		t.Fatalf("expected second row day-two snapshot worth %d, got date=%s worth=%d", secondWorth, rows[1].OccurredAt.Format("2006-01-02"), rows[1].TotalWorth)
	}
}

func waitForAssetSnapshotsLatestWorth(t *testing.T, baseURL string, accountID uuid.UUID, expectedWorth string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for {
		req, err := http.NewRequest(
			http.MethodGet,
			baseURL+"/assets/snapshots?account_id="+accountID.String(),
			nil,
		)
		if err != nil {
			t.Fatalf("failed creating asset snapshots request: %v", err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET /assets/snapshots failed: %v", err)
		}
		var points []api.AssetGrowthPointResponse
		decodeErr := json.NewDecoder(res.Body).Decode(&points)
		closeErr := res.Body.Close()
		if closeErr != nil {
			t.Fatalf("failed closing asset snapshots response body: %v", closeErr)
		}
		if res.StatusCode == http.StatusOK && decodeErr == nil && len(points) > 0 {
			last := points[len(points)-1]
			if last.TotalWorth == expectedWorth {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for snapshots latest worth=%s (status=%d points=%d decodeErr=%v)", expectedWorth, res.StatusCode, len(points), decodeErr)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
