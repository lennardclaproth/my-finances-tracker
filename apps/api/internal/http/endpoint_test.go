package http

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type endpointJSONRequest struct {
	Name string `json:"name"`
}

type endpointQueryRequest struct {
	Limit int `query:"limit"`
}

type endpointMultipartRequest struct {
	File     multipart.File `multipart:"file"`
	Filename string         `multipart:"filename"`
}

func TestEndpoint_DecodeJSONMalformed_Returns400(t *testing.T) {
	t.Parallel()

	h := Endpoint(
		func(r *http.Request) (endpointJSONRequest, error) { return DecodeJSON[endpointJSONRequest](r) },
		&captureLogger{},
		func(ctx context.Context, req endpointJSONRequest) (int, struct{}, error) {
			return http.StatusOK, struct{}{}, nil
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(`{"name"`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if payload["body"] == "" {
		t.Fatalf("expected body validation message, got %+v", payload)
	}
}

func TestEndpoint_DecodeJSONUnknownField_Returns400(t *testing.T) {
	t.Parallel()

	h := Endpoint(
		func(r *http.Request) (endpointJSONRequest, error) { return DecodeJSON[endpointJSONRequest](r) },
		&captureLogger{},
		func(ctx context.Context, req endpointJSONRequest) (int, struct{}, error) {
			return http.StatusOK, struct{}{}, nil
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(`{"name":"foo","extra":"bar"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if payload["extra"] == "" {
		t.Fatalf("expected unknown field error for extra, got %+v", payload)
	}
}

func TestEndpoint_DecodeQueryInvalidInt_Returns400(t *testing.T) {
	t.Parallel()

	h := Endpoint(
		func(r *http.Request) (endpointQueryRequest, error) { return DecodeQuery[endpointQueryRequest](r) },
		&captureLogger{},
		func(ctx context.Context, req endpointQueryRequest) (int, struct{}, error) {
			return http.StatusOK, struct{}{}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/cashflow/transactions?limit=abc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if payload["limit"] == "" {
		t.Fatalf("expected limit validation message, got %+v", payload)
	}
}

func TestEndpoint_DecodeMultipartTooLarge_Returns413(t *testing.T) {
	t.Parallel()

	h := Endpoint(
		func(r *http.Request) (endpointMultipartRequest, error) {
			return DecodeMultipartFile[endpointMultipartRequest](r, MultipartFileDecoderOptions{
				FieldName: "file",
				MaxBytes:  4,
				MaxMemory: 4,
			})
		},
		&captureLogger{},
		func(ctx context.Context, req endpointMultipartRequest) (int, struct{}, error) {
			defer func() {
				if err := req.File.Close(); err != nil {
					t.Fatalf("failed closing multipart file: %v", err)
				}
			}()
			return http.StatusOK, struct{}{}, nil
		},
	)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "sample.csv")
	if err != nil {
		t.Fatalf("failed creating multipart file: %v", err)
	}
	if _, err := part.Write([]byte("123456")); err != nil {
		t.Fatalf("failed writing multipart payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed closing multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/import/csv", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if payload["file"] == "" {
		t.Fatalf("expected file size validation message, got %+v", payload)
	}
}
