package notify

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type noopLogger struct{}

func (noopLogger) Debug(_ context.Context, _ string, _ ...any) {}

func (noopLogger) Info(_ context.Context, _ string, _ ...any) {}

func (noopLogger) Warn(_ context.Context, _ string, _ ...any) {}

func (noopLogger) Error(_ context.Context, _ string, _ error, _ ...any) {}

func TestHub_IdleNoUpdatesClosesWith4001(t *testing.T) {
	t.Parallel()

	hub := NewHub(
		noopLogger{},
		WithPingInterval(300*time.Millisecond),
		WithStaleCheckEvery(10*time.Millisecond),
		WithIdleNoUpdateAfter(80*time.Millisecond),
		WithWriteWait(100*time.Millisecond),
	)
	t.Cleanup(func() {
		if err := hub.Close(); err != nil {
			t.Errorf("failed cleanup close: %v", err)
		}
	})

	accountID := uuid.New()
	conn := mustDialWS(t, hub, accountID)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("failed closing resource: %v", err)
		}
	}()

	if err := conn.SetReadDeadline(time.Now().Add(800 * time.Millisecond)); err != nil {
		t.Fatalf("failed setting read deadline: %v", err)
	}
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			if websocket.IsCloseError(err, CloseCodeIdleNoUpdates) {
				return
			}
			t.Fatalf("expected close code %d, got %v", CloseCodeIdleNoUpdates, err)
		}
	}
}

func TestHub_UpdateResetsIdleTimer(t *testing.T) {
	t.Parallel()

	hub := NewHub(
		noopLogger{},
		WithPingInterval(300*time.Millisecond),
		WithStaleCheckEvery(10*time.Millisecond),
		WithIdleNoUpdateAfter(160*time.Millisecond),
		WithWriteWait(100*time.Millisecond),
	)
	t.Cleanup(func() {
		if err := hub.Close(); err != nil {
			t.Errorf("failed cleanup close: %v", err)
		}
	})

	accountID := uuid.New()
	conn := mustDialWS(t, hub, accountID)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("failed closing resource: %v", err)
		}
	}()

	time.Sleep(90 * time.Millisecond)
	if got := hub.NotifyDataChanged(context.Background(), accountID, EventImportCompleted); got != 1 {
		t.Fatalf("expected one notified client, got %d", got)
	}

	if err := conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
		t.Fatalf("failed setting read deadline: %v", err)
	}
	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected data_changed message, got error: %v", err)
	}
	if !strings.Contains(string(payload), EventImportCompleted) {
		t.Fatalf("expected %q in payload, got %s", EventImportCompleted, string(payload))
	}

	time.Sleep(90 * time.Millisecond)
	if got := len(hub.snapshot(accountID)); got != 1 {
		t.Fatalf("expected connection to remain open before idle timeout, clients=%d", got)
	}

	if err := conn.SetReadDeadline(time.Now().Add(700 * time.Millisecond)); err != nil {
		t.Fatalf("failed setting read deadline: %v", err)
	}
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			if websocket.IsCloseError(err, CloseCodeIdleNoUpdates) {
				return
			}
			t.Fatalf("expected close code %d after idle window, got %v", CloseCodeIdleNoUpdates, err)
		}
	}
}

func TestHub_MissedPongsClosesConnection(t *testing.T) {
	t.Parallel()

	hub := NewHub(
		noopLogger{},
		WithPingInterval(20*time.Millisecond),
		WithStaleCheckEvery(500*time.Millisecond),
		WithIdleNoUpdateAfter(5*time.Second),
		WithWriteWait(100*time.Millisecond),
		WithMaxMissedPongs(3),
	)
	t.Cleanup(func() {
		if err := hub.Close(); err != nil {
			t.Errorf("failed cleanup close: %v", err)
		}
	})

	accountID := uuid.New()
	conn := mustDialWS(t, hub, accountID)
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("failed closing resource: %v", err)
		}
	}()

	// Disable automatic pong replies for this test.
	conn.SetPingHandler(func(_ string) error { return nil })

	if err := conn.SetReadDeadline(time.Now().Add(600 * time.Millisecond)); err != nil {
		t.Fatalf("failed setting read deadline: %v", err)
	}
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			if websocket.IsCloseError(err, CloseCodeIdleNoUpdates) {
				t.Fatalf("expected pong-timeout close, got idle close error: %v", err)
			}
			return
		}
	}
}

func mustDialWS(t *testing.T, hub *Hub, accountID uuid.UUID) *websocket.Conn {
	t.Helper()

	server := httptest.NewServer(hub.Handler())
	t.Cleanup(server.Close)

	wsURL := fmt.Sprintf("ws%s/ws/accounts/%s", strings.TrimPrefix(server.URL, "http"), accountID.String())
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if resp != nil {
			t.Fatalf("dial failed: %v status=%d", err, resp.StatusCode)
		}
		t.Fatalf("dial failed: %v", err)
	}
	deadline := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(hub.snapshot(accountID)) > 0 {
			return conn
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("websocket client was not registered for account %s", accountID)
	return conn
}
