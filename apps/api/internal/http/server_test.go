package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_newHTTPServer_SetsTimeouts(t *testing.T) {
	t.Parallel()

	srv := &Server{
		addr: ":8080",
		log:  &captureLogger{},
	}

	httpSrv := srv.newHTTPServer(http.NewServeMux())
	if httpSrv.ReadTimeout != readTimeout {
		t.Fatalf("expected read timeout %s, got %s", readTimeout, httpSrv.ReadTimeout)
	}
	if httpSrv.ReadHeaderTimeout != readHeaderTimeout {
		t.Fatalf("expected read header timeout %s, got %s", readHeaderTimeout, httpSrv.ReadHeaderTimeout)
	}
	if httpSrv.WriteTimeout != writeTimeout {
		t.Fatalf("expected write timeout %s, got %s", writeTimeout, httpSrv.WriteTimeout)
	}
	if httpSrv.IdleTimeout != idleTimeout {
		t.Fatalf("expected idle timeout %s, got %s", idleTimeout, httpSrv.IdleTimeout)
	}
}

func TestShouldIgnoreAPMServerRequest_IgnoresWebsocketRoutePattern(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/ws/accounts/64a3d50f-c71a-4015-9ee1-45572147ce56", nil)
	req.Pattern = "GET /ws/accounts/{account_id}"

	if !shouldIgnoreAPMServerRequest(req) {
		t.Fatalf("expected websocket route pattern to be ignored by APM")
	}
}

func TestShouldIgnoreAPMServerRequest_IgnoresWebsocketPathFallback(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/ws/accounts/64a3d50f-c71a-4015-9ee1-45572147ce56", nil)

	if !shouldIgnoreAPMServerRequest(req) {
		t.Fatalf("expected websocket path to be ignored by APM")
	}
}

func TestShouldIgnoreAPMServerRequest_DoesNotIgnoreOtherRoutes(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Pattern = "GET /health"

	if shouldIgnoreAPMServerRequest(req) {
		t.Fatalf("expected non-websocket route to be traced by APM")
	}
}
