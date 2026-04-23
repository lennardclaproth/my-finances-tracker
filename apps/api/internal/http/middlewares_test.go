package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
	"go.elastic.co/apm/v2"
)

type captureLogger struct {
	mu     sync.Mutex
	fields []any
}

func (l *captureLogger) Debug(_ context.Context, _ string, fields ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fields = append([]any{}, fields...)
}

func (l *captureLogger) Info(_ context.Context, _ string, fields ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fields = append([]any{}, fields...)
}

func (l *captureLogger) Warn(_ context.Context, _ string, fields ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fields = append([]any{}, fields...)
}

func (l *captureLogger) Error(_ context.Context, _ string, _ error, fields ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fields = append([]any{}, fields...)
}

func (l *captureLogger) asMap() map[string]any {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make(map[string]any, len(l.fields)/2)
	for i := 0; i+1 < len(l.fields); i += 2 {
		key, ok := l.fields[i].(string)
		if !ok {
			continue
		}
		out[key] = l.fields[i+1]
	}
	return out
}

func (l *captureLogger) keyCount(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	count := 0
	for i := 0; i+1 < len(l.fields); i += 2 {
		k, ok := l.fields[i].(string)
		if !ok {
			continue
		}
		if k == key {
			count++
		}
	}
	return count
}

func TestWithRequestIdentifiers_GeneratesAndEchoesHeaders(t *testing.T) {
	t.Parallel()

	var requestIDFromContext string
	var correlationIDFromContext string
	handler := WithRequestIdentifiers()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestIDFromContext = observability.RequestIDFromContext(r.Context())
		correlationIDFromContext = observability.CorrelationIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if requestIDFromContext == "" {
		t.Fatalf("expected request id in context")
	}
	if correlationIDFromContext == "" {
		t.Fatalf("expected correlation id in context")
	}
	if got := res.Header().Get(observability.HeaderRequestID); got != requestIDFromContext {
		t.Fatalf("unexpected response request id: %q", got)
	}
	if got := res.Header().Get(observability.HeaderCorrelationID); got != correlationIDFromContext {
		t.Fatalf("unexpected response correlation id: %q", got)
	}
}

func TestWithRequestIdentifiers_PreservesInboundHeaders(t *testing.T) {
	t.Parallel()

	handler := WithRequestIdentifiers()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(observability.HeaderRequestID, "req-inbound")
	req.Header.Set(observability.HeaderCorrelationID, "corr-inbound")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if got := res.Header().Get(observability.HeaderRequestID); got != "req-inbound" {
		t.Fatalf("expected inbound request id to be preserved, got %q", got)
	}
	if got := res.Header().Get(observability.HeaderCorrelationID); got != "corr-inbound" {
		t.Fatalf("expected inbound correlation id to be preserved, got %q", got)
	}
}

func TestWithRequestLogging_IncludesTraceAndCorrelationFields(t *testing.T) {
	t.Parallel()

	logger := &captureLogger{}
	handler := WithRequestLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tx := apm.DefaultTracer().StartTransaction("http.test", "request")
	defer tx.End()

	ctx := apm.ContextWithTransaction(context.Background(), tx)
	ctx = observability.ContextWithRequestID(ctx, "req-123")
	ctx = observability.ContextWithCorrelationID(ctx, "corr-456")

	req := httptest.NewRequest(http.MethodGet, "/portfolio", nil).WithContext(ctx)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	fields := logger.asMap()
	if fields["request_id"] != "req-123" {
		t.Fatalf("expected request_id in logs, got %#v", fields["request_id"])
	}
	if fields["correlation_id"] != "corr-456" {
		t.Fatalf("expected correlation_id in logs, got %#v", fields["correlation_id"])
	}
	if _, ok := fields["trace.id"]; !ok {
		t.Fatalf("expected trace.id in logs")
	}
	if _, ok := fields["transaction.id"]; !ok {
		t.Fatalf("expected transaction.id in logs")
	}
	if got := fields["operation"]; got != "http.get.portfolio" {
		t.Fatalf("expected operation http.get.portfolio, got %#v", got)
	}
	if got := logger.keyCount("request_id"); got != 1 {
		t.Fatalf("expected request_id to appear exactly once, got %d", got)
	}
	if got := logger.keyCount("correlation_id"); got != 1 {
		t.Fatalf("expected correlation_id to appear exactly once, got %d", got)
	}
	if got := logger.keyCount("trace.id"); got != 1 {
		t.Fatalf("expected trace.id to appear exactly once, got %d", got)
	}
	if got := logger.keyCount("transaction.id"); got != 1 {
		t.Fatalf("expected transaction.id to appear exactly once, got %d", got)
	}
}

func TestWithRequestLogging_PreservesHijackerForWebsocketUpgrade(t *testing.T) {
	t.Parallel()

	logger := &captureLogger{}
	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool { return true },
	}
	upgradeErr := make(chan error, 1)

	handler := WithRequestLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			upgradeErr <- err
			return
		}
		if closeErr := conn.Close(); closeErr != nil {
			upgradeErr <- closeErr
			return
		}
		upgradeErr <- nil
	}))

	server := httptest.NewServer(WithRequestIdentifiers()(handler))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("websocket dial failed (status=%d): %v", status, err)
	}
	if closeErr := conn.Close(); closeErr != nil {
		t.Fatalf("failed closing websocket client connection: %v", closeErr)
	}

	if err := <-upgradeErr; err != nil {
		t.Fatalf("expected websocket upgrade to succeed through middleware wrapper: %v", err)
	}
}
