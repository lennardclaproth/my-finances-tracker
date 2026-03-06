package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/jobs"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

func TestIntegration_UploadDailiesFile_AcceptsAndProcessesAsync(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()
	seedBrandNewDayManualProvider(t, app.db)

	listing, err := marketdata.NewListing(
		"BND.AS",
		"BrandNewDay Listing",
		marketdata.SourceBrandNewDay,
	)
	if err != nil {
		t.Fatalf("failed creating listing: %v", err)
	}
	if err := storage.NewSQLXListingStore(app.db).Create(ctx, listing); err != nil {
		t.Fatalf("failed storing listing: %v", err)
	}

	var reqBody bytes.Buffer
	writer := multipart.NewWriter(&reqBody)
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="brandnewday.csv"`)
	partHeader.Set("Content-Type", "text/csv")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("failed creating multipart part: %v", err)
	}
	_, _ = part.Write([]byte("Date\tNAV\tAsk\tBid\tDividend\n28/02/2026\t38,286191\t38,286191\t38,286191\t-\n27/02/2026\t38,286065\t38,286065\t38,286065\t-\n"))
	if err := writer.WriteField("listing_id", listing.ID.String()); err != nil {
		t.Fatalf("failed writing listing_id field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed closing multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/marketdata/dailies/upload", &reqBody)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /marketdata/dailies/upload failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 202, got %d body=%s", res.StatusCode, string(body))
	}

	var accepted api.DailyUploadAcceptedResponse
	if err := json.NewDecoder(res.Body).Decode(&accepted); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if accepted.UploadID == uuid.Nil {
		t.Fatalf("expected non-nil upload id")
	}
	if accepted.Status != "PENDING" {
		t.Fatalf("expected PENDING status, got %s", accepted.Status)
	}

	uploadJob := jobs.NewDailyUploadJob(
		storage.NewSQLXDailyUploadStore(app.db),
		storage.NewSQLXListingStore(app.db),
		storage.NewSQLXDailyStore(app.db),
		storage.NewDisk(""),
		nilLogger{},
		20*time.Millisecond,
		256,
	)

	jobCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- uploadJob.Start(jobCtx)
	}()
	if err := uploadJob.Enqueue(jobCtx, accepted.UploadID); err != nil {
		t.Fatalf("failed to enqueue upload: %v", err)
	}

	deadline := time.Now().Add(1500 * time.Millisecond)
	for time.Now().Before(deadline) {
		statusReq, err := http.NewRequest(http.MethodGet, app.server.URL+"/marketdata/dailies/uploads/"+accepted.UploadID.String(), nil)
		if err != nil {
			t.Fatalf("failed creating status request: %v", err)
		}
		statusRes, err := http.DefaultClient.Do(statusReq)
		if err != nil {
			t.Fatalf("GET /marketdata/dailies/uploads/{id} failed: %v", err)
		}
		if statusRes.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(statusRes.Body)
			statusRes.Body.Close()
			t.Fatalf("expected 200 for status endpoint, got %d body=%s", statusRes.StatusCode, string(body))
		}
		var status api.DailyUploadStatusResponse
		if err := json.NewDecoder(statusRes.Body).Decode(&status); err != nil {
			statusRes.Body.Close()
			t.Fatalf("failed decoding status response: %v", err)
		}
		statusRes.Body.Close()

		if status.Status == string(marketdata.DailyUploadStatusSucceeded) || status.Status == string(marketdata.DailyUploadStatusPartial) {
			if status.InsertedRows == 0 {
				t.Fatalf("expected inserted_rows > 0, got %d", status.InsertedRows)
			}
			cancel()
			select {
			case <-time.After(time.Second):
				t.Fatalf("daily upload job did not stop after cancel")
			case <-done:
			}
			table := qualifiedTable(app.backend, storage.SchemaMarketData, storage.TableHistories)
			var count int
			query := app.db.Rebind(fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE listing_id = ?", table))
			if err := app.db.GetContext(ctx, &count, query, listing.ID); err != nil {
				t.Fatalf("failed counting persisted dailies: %v", err)
			}
			if count == 0 {
				t.Fatalf("expected persisted dailies for listing")
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for daily upload to process")
}

func TestIntegration_UploadDailiesFile_UnsupportedSourceReturns422(t *testing.T) {
	app := newIntegrationApp(t)
	ctx := context.Background()

	listing, err := marketdata.NewListing(
		"NO-PARSER.AS",
		"No Parser Listing",
		marketdata.SourceAlphaVantage,
	)
	if err != nil {
		t.Fatalf("failed creating listing: %v", err)
	}
	if err := storage.NewSQLXListingStore(app.db).Create(ctx, listing); err != nil {
		t.Fatalf("failed storing listing: %v", err)
	}

	var reqBody bytes.Buffer
	writer := multipart.NewWriter(&reqBody)
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="daily.csv"`)
	partHeader.Set("Content-Type", "text/csv")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("failed creating multipart part: %v", err)
	}
	_, _ = part.Write([]byte("Date\tNAV\tAsk\tBid\tDividend\n28/02/2026\t38,286191\t38,286191\t38,286191\t-\n"))
	if err := writer.WriteField("listing_id", listing.ID.String()); err != nil {
		t.Fatalf("failed writing listing_id field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed closing multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, app.server.URL+"/marketdata/dailies/upload", &reqBody)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /marketdata/dailies/upload failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("expected 422, got %d body=%s", res.StatusCode, string(body))
	}
}

type nilLogger struct{}

func (n nilLogger) Info(ctx context.Context, msg string, fields ...any) {}

func (n nilLogger) Error(ctx context.Context, msg string, err error, fields ...any) {}
