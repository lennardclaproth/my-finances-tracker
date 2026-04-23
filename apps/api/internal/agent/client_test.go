package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCallAgent_ClassifiesHTTPFailures(t *testing.T) {
	tests := []struct {
		name            string
		status          int
		wantErr         bool
		wantClientErr   bool
		wantServerErr   bool
		wantUnknownKind bool
	}{
		{
			name:          "success status",
			status:        http.StatusOK,
			wantErr:       false,
			wantClientErr: false,
			wantServerErr: false,
		},
		{
			name:          "client status",
			status:        http.StatusBadRequest,
			wantErr:       true,
			wantClientErr: true,
			wantServerErr: false,
		},
		{
			name:          "server status",
			status:        http.StatusBadGateway,
			wantErr:       true,
			wantClientErr: false,
			wantServerErr: true,
		},
		{
			name:            "redirect treated as unknown",
			status:          http.StatusFound,
			wantErr:         true,
			wantUnknownKind: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				if _, err := w.Write([]byte("test-response")); err != nil {
					t.Errorf("failed writing test response: %v", err)
				}
			}))
			defer srv.Close()

			client := NewClient(srv.URL)
			err := client.CallAgent(context.Background(), uuid.New(), "test")

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr {
				return
			}

			if got := IsClientError(err); got != tt.wantClientErr {
				t.Fatalf("expected IsClientError=%v, got %v", tt.wantClientErr, got)
			}
			if got := IsServerError(err); got != tt.wantServerErr {
				t.Fatalf("expected IsServerError=%v, got %v", tt.wantServerErr, got)
			}
			if tt.wantUnknownKind {
				var callErr *CallError
				if !asCallError(err, &callErr) {
					t.Fatalf("expected CallError type, got %T", err)
				}
				if callErr.Kind != CallErrorUnknown {
					t.Fatalf("expected CallErrorUnknown kind, got %s", callErr.Kind)
				}
			}
		})
	}
}

func TestCallAgent_ClassifiesUnreachableFailures(t *testing.T) {
	client := NewClient(
		"http://127.0.0.1:65534",
		func(c *Client) { c.http.Timeout = 250 * time.Millisecond },
	)

	err := client.CallAgent(context.Background(), uuid.New(), "test")
	if err == nil {
		t.Fatalf("expected error for unreachable endpoint")
	}
	if !IsUnreachableError(err) {
		t.Fatalf("expected unreachable error, got %v", err)
	}
	if IsClientError(err) {
		t.Fatalf("unreachable errors must not be treated as client errors")
	}
	if IsServerError(err) {
		t.Fatalf("unreachable errors must not be treated as server errors")
	}
}

func asCallError(err error, target **CallError) bool {
	callErr, ok := err.(*CallError)
	if !ok {
		return false
	}
	*target = callErr
	return true
}
