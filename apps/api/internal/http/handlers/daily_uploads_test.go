package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

type fakeDailyUploadHandlerStore struct {
	createFn func(ctx context.Context, upload *marketdata.DailyUpload) error
	fetchFn  func(ctx context.Context, id uuid.UUID) (*marketdata.DailyUpload, error)
}

func (s *fakeDailyUploadHandlerStore) Create(ctx context.Context, upload *marketdata.DailyUpload) error {
	if s.createFn != nil {
		return s.createFn(ctx, upload)
	}
	return nil
}

func (s *fakeDailyUploadHandlerStore) FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.DailyUpload, error) {
	if s.fetchFn != nil {
		return s.fetchFn(ctx, id)
	}
	return nil, marketdata.ErrDailyUploadNotFound
}

type fakeDailyUploadListingHandlerStore struct {
	listing *marketdata.Listing
}

func (s *fakeDailyUploadListingHandlerStore) FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
	return s.listing, nil
}

type fakeDailyUploadProviderHandlerStore struct {
	getByNameFn func(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error)
}

func (s *fakeDailyUploadProviderHandlerStore) GetByName(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error) {
	if s.getByNameFn != nil {
		return s.getByNameFn(ctx, name)
	}
	return nil, marketdata.ErrProviderNotFound
}

type fakeDailyUploadFileWriter struct{}

func (w *fakeDailyUploadFileWriter) WriteCsv(r io.Reader) (string, error) {
	_, _ = io.ReadAll(r)
	return "test.csv", nil
}

type fakeDailyUploadFileRemover struct{}

func (r *fakeDailyUploadFileRemover) Remove(path string) error { return nil }

type fakeDailyUploadEnqueuer struct{}

func (e *fakeDailyUploadEnqueuer) Enqueue(ctx context.Context, uploadID uuid.UUID) error { return nil }

func TestGetDailyUploadStatus_InvalidUUIDReturns400(t *testing.T) {
	h := GetDailyUploadStatus(&testLogger{}, &fakeDailyUploadHandlerStore{})
	req := httptest.NewRequest(http.MethodGet, "/marketdata/dailies/uploads/not-a-uuid", nil)
	req.SetPathValue("upload_id", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetDailyUploadStatus_NotFoundReturns404(t *testing.T) {
	h := GetDailyUploadStatus(&testLogger{}, &fakeDailyUploadHandlerStore{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*marketdata.DailyUpload, error) {
			return nil, marketdata.ErrDailyUploadNotFound
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/marketdata/dailies/uploads/"+uuid.New().String(), nil)
	req.SetPathValue("upload_id", uuid.New().String())
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUploadDailiesFile_BadFileExtensionReturns400(t *testing.T) {
	listingID := uuid.New()
	h := UploadDailiesFile(
		&testLogger{},
		&fakeDailyUploadHandlerStore{},
		&fakeDailyUploadListingHandlerStore{
			listing: &marketdata.Listing{
				ID:     listingID,
				Symbol: "BND.AS",
				Source: marketdata.SourceBrandNewDay,
				Active: true,
			},
		},
		&fakeDailyUploadProviderHandlerStore{
			getByNameFn: func(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error) {
				return marketdata.NewManualProvider(marketdata.ProviderBrandNewDay)
			},
		},
		&fakeDailyUploadFileWriter{},
		&fakeDailyUploadFileRemover{},
		&fakeDailyUploadEnqueuer{},
	)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="daily.pdf"`)
	partHeader.Set("Content-Type", "text/plain")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("failed creating part: %v", err)
	}
	_, _ = part.Write([]byte("Date\tNAV\tAsk\tBid\tDividend\n28/02/2026\t1\t1\t1\t-\n"))
	if err := writer.WriteField("listing_id", listingID.String()); err != nil {
		t.Fatalf("failed writing listing_id: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed closing writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/marketdata/dailies/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUploadDailiesFile_NonManualProviderReturns422(t *testing.T) {
	listingID := uuid.New()
	h := UploadDailiesFile(
		&testLogger{},
		&fakeDailyUploadHandlerStore{},
		&fakeDailyUploadListingHandlerStore{
			listing: &marketdata.Listing{
				ID:     listingID,
				Symbol: "AAA.AS",
				Source: marketdata.SourceAlphaVantage,
				Active: true,
			},
		},
		&fakeDailyUploadProviderHandlerStore{
			getByNameFn: func(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error) {
				return marketdata.NewAPIProviderWithAPIKey(name, "https://api.test", "key-1")
			},
		},
		&fakeDailyUploadFileWriter{},
		&fakeDailyUploadFileRemover{},
		&fakeDailyUploadEnqueuer{},
	)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="daily.csv"`)
	partHeader.Set("Content-Type", "text/csv")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("failed creating part: %v", err)
	}
	_, _ = part.Write([]byte("Date\tNAV\tAsk\tBid\tDividend\n28/02/2026\t1\t1\t1\t-\n"))
	if err := writer.WriteField("listing_id", listingID.String()); err != nil {
		t.Fatalf("failed writing listing_id: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed closing writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/marketdata/dailies/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUploadDailiesFile_MissingProviderReturns422(t *testing.T) {
	listingID := uuid.New()
	h := UploadDailiesFile(
		&testLogger{},
		&fakeDailyUploadHandlerStore{},
		&fakeDailyUploadListingHandlerStore{
			listing: &marketdata.Listing{
				ID:     listingID,
				Symbol: "BND.AS",
				Source: marketdata.SourceBrandNewDay,
				Active: true,
			},
		},
		&fakeDailyUploadProviderHandlerStore{},
		&fakeDailyUploadFileWriter{},
		&fakeDailyUploadFileRemover{},
		&fakeDailyUploadEnqueuer{},
	)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="daily.csv"`)
	partHeader.Set("Content-Type", "text/csv")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("failed creating part: %v", err)
	}
	_, _ = part.Write([]byte("Date\tNAV\tAsk\tBid\tDividend\n28/02/2026\t1\t1\t1\t-\n"))
	if err := writer.WriteField("listing_id", listingID.String()); err != nil {
		t.Fatalf("failed writing listing_id: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed closing writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/marketdata/dailies/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetDailyUploadStatus_SuccessPayload(t *testing.T) {
	uploadID := uuid.New()
	listingID := uuid.New()
	now := time.Now().UTC()
	h := GetDailyUploadStatus(&testLogger{}, &fakeDailyUploadHandlerStore{
		fetchFn: func(ctx context.Context, id uuid.UUID) (*marketdata.DailyUpload, error) {
			return &marketdata.DailyUpload{
				ID:            uploadID,
				ListingID:     listingID,
				Source:        marketdata.SourceBrandNewDay,
				Status:        marketdata.DailyUploadStatusSucceeded,
				TotalRows:     1,
				InsertedRows:  1,
				DuplicateRows: 0,
				ErrorRows:     0,
				RowErrors:     []marketdata.DailyUploadRowError{},
				CreatedAt:     now,
				UpdatedAt:     now,
			}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/marketdata/dailies/uploads/"+uploadID.String(), nil)
	req.SetPathValue("upload_id", uploadID.String())
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}
	if payload["status"] != "SUCCEEDED" {
		t.Fatalf("expected status SUCCEEDED, got %v", payload["status"])
	}
}
