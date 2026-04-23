package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

const (
	defaultPingInterval      = 10 * time.Second
	defaultStaleCheckEvery   = 30 * time.Second
	defaultIdleNoUpdateAfter = 5 * time.Minute
	defaultWriteWait         = 5 * time.Second
	defaultSendQueueSize     = 32
	defaultMaxMissedPongs    = 3

	// CloseCodeIdleNoUpdates closes stale sockets with no data_changed updates.
	CloseCodeIdleNoUpdates = 4001
	// CloseReasonIdleNoUpdates is used with CloseCodeIdleNoUpdates.
	CloseReasonIdleNoUpdates = "idle_no_updates"
)

// DataChangedMessage is sent to websocket clients when account data has changed.
type DataChangedMessage struct {
	Type      string    `json:"type"`
	Event     string    `json:"event"`
	AccountID uuid.UUID `json:"account_id"`
	Timestamp time.Time `json:"timestamp"`
}

// Option configures Hub behavior.
type Option func(*Hub)

// WithPingInterval configures how often ping frames are sent while idle.
func WithPingInterval(d time.Duration) Option {
	return func(h *Hub) {
		if d > 0 {
			h.pingInterval = d
		}
	}
}

// WithStaleCheckEvery configures how often no-update staleness is evaluated.
func WithStaleCheckEvery(d time.Duration) Option {
	return func(h *Hub) {
		if d > 0 {
			h.staleCheckEvery = d
		}
	}
}

// WithIdleNoUpdateAfter configures the no-update timeout before closing a socket.
func WithIdleNoUpdateAfter(d time.Duration) Option {
	return func(h *Hub) {
		if d > 0 {
			h.idleNoUpdateAfter = d
		}
	}
}

// WithMaxMissedPongs configures how many missed pong responses are tolerated.
func WithMaxMissedPongs(n int) Option {
	return func(h *Hub) {
		if n > 0 {
			h.maxMissedPongs = n
		}
	}
}

// WithWriteWait configures the write deadline used for websocket writes.
func WithWriteWait(d time.Duration) Option {
	return func(h *Hub) {
		if d > 0 {
			h.writeWait = d
		}
	}
}

// WithSendQueueSize configures the buffered update queue size per websocket client.
func WithSendQueueSize(size int) Option {
	return func(h *Hub) {
		if size > 0 {
			h.sendQueueSize = size
		}
	}
}

// WithCheckOrigin configures the origin checker used during websocket upgrades.
func WithCheckOrigin(fn func(r *http.Request) bool) Option {
	return func(h *Hub) {
		if fn != nil {
			h.upgrader.CheckOrigin = fn
		}
	}
}

// Hub manages account-scoped websocket clients and realtime data-changed fanout.
type Hub struct {
	log logging.Logger

	pingInterval      time.Duration
	staleCheckEvery   time.Duration
	idleNoUpdateAfter time.Duration
	maxMissedPongs    int
	writeWait         time.Duration
	sendQueueSize     int
	now               func() time.Time

	upgrader websocket.Upgrader

	mu      sync.RWMutex
	clients map[uuid.UUID]map[*client]struct{}
}

// NewHub creates a Hub with sensible defaults and optional overrides.
func NewHub(log logging.Logger, opts ...Option) *Hub {
	h := &Hub{
		log:               log,
		pingInterval:      defaultPingInterval,
		staleCheckEvery:   defaultStaleCheckEvery,
		idleNoUpdateAfter: defaultIdleNoUpdateAfter,
		maxMissedPongs:    defaultMaxMissedPongs,
		writeWait:         defaultWriteWait,
		sendQueueSize:     defaultSendQueueSize,
		now:               time.Now,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[uuid.UUID]map[*client]struct{}),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Handler returns the websocket HTTP handler.
func (h *Hub) Handler() http.Handler {
	return http.HandlerFunc(h.ServeWS)
}

// ServeWS upgrades the request and attaches the connection to the account scope.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	accountIDRaw := r.PathValue("account_id")
	if strings.TrimSpace(accountIDRaw) == "" {
		const prefix = "/ws/accounts/"
		if strings.HasPrefix(r.URL.Path, prefix) {
			accountIDRaw = strings.TrimPrefix(r.URL.Path, prefix)
		}
	}
	accountID, err := uuid.Parse(accountIDRaw)
	if err != nil {
		h.log.Warn(r.Context(), "websocket rejected: invalid account id", "path", r.URL.Path, "error", err.Error())
		http.Error(w, "invalid account_id", http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Warn(r.Context(), "websocket upgrade failed", "account_id", accountIDRaw, "path", r.URL.Path, "error", err.Error())
		return
	}

	c := newClient(h, conn, accountID)
	h.addClient(c)
	h.log.Info(r.Context(), "websocket connected", "account_id", accountID.String())

	c.start()
	c.readLoop()
	c.close(websocket.CloseNormalClosure, "")
}

// NotifyDataChanged broadcasts a data_changed message to all clients for accountID.
// It returns the number of clients that accepted the message into their send queue.
func (h *Hub) NotifyDataChanged(ctx context.Context, accountID uuid.UUID, event string) int {
	msg := DataChangedMessage{
		Type:      "data_changed",
		Event:     event,
		AccountID: accountID,
		Timestamp: h.now().UTC(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		h.log.Error(ctx, "failed marshaling websocket payload", err, "account_id", accountID.String(), "event", event)
		return 0
	}

	targets := h.snapshot(accountID)
	sent := 0
	for _, c := range targets {
		if c.enqueueUpdate(payload) {
			sent++
		}
	}
	return sent
}

// Close closes all active websocket clients managed by the hub.
func (h *Hub) Close() error {
	h.mu.RLock()
	all := make([]*client, 0, 16)
	for _, group := range h.clients {
		for c := range group {
			all = append(all, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range all {
		c.close(websocket.CloseGoingAway, "server_shutdown")
	}
	return nil
}

func (h *Hub) addClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	group := h.clients[c.accountID]
	if group == nil {
		group = make(map[*client]struct{})
		h.clients[c.accountID] = group
	}
	group[c] = struct{}{}
}

func (h *Hub) removeClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	group := h.clients[c.accountID]
	if group == nil {
		return
	}
	delete(group, c)
	if len(group) == 0 {
		delete(h.clients, c.accountID)
	}
}

func (h *Hub) snapshot(accountID uuid.UUID) []*client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	group := h.clients[accountID]
	if len(group) == 0 {
		return nil
	}

	out := make([]*client, 0, len(group))
	for c := range group {
		out = append(out, c)
	}
	return out
}

type client struct {
	hub       *Hub
	conn      *websocket.Conn
	accountID uuid.UUID
	send      chan []byte
	done      chan struct{}

	mu               sync.Mutex
	lastUpdateSentAt time.Time
	missedPongs      int
	closeOnce        sync.Once
}

func newClient(h *Hub, conn *websocket.Conn, accountID uuid.UUID) *client {
	return &client{
		hub:              h,
		conn:             conn,
		accountID:        accountID,
		send:             make(chan []byte, h.sendQueueSize),
		done:             make(chan struct{}),
		lastUpdateSentAt: h.now(),
	}
}

func (c *client) start() {
	c.conn.SetPongHandler(func(_ string) error {
		c.mu.Lock()
		c.missedPongs = 0
		c.mu.Unlock()
		return nil
	})

	go c.writeLoop()
	go c.heartbeatLoop()
}

func (c *client) readLoop() {
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *client) enqueueUpdate(payload []byte) bool {
	select {
	case <-c.done:
		return false
	default:
	}

	select {
	case c.send <- payload:
		return true
	default:
		c.close(websocket.CloseTryAgainLater, "send_queue_full")
		return false
	}
}

func (c *client) writeLoop() {
	for {
		select {
		case <-c.done:
			return
		case payload := <-c.send:
			if err := c.conn.SetWriteDeadline(c.hub.now().Add(c.hub.writeWait)); err != nil {
				c.close(websocket.CloseGoingAway, "write_deadline_failed")
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				c.close(websocket.CloseGoingAway, "write_failed")
				return
			}
			c.mu.Lock()
			c.lastUpdateSentAt = c.hub.now()
			c.mu.Unlock()
		}
	}
}

func (c *client) heartbeatLoop() {
	pingTicker := time.NewTicker(c.hub.pingInterval)
	staleTicker := time.NewTicker(c.hub.staleCheckEvery)
	defer pingTicker.Stop()
	defer staleTicker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-pingTicker.C:
			now := c.hub.now()
			if c.idleFor(now) < c.hub.pingInterval {
				continue
			}
			if c.incrementMissedPongs() >= c.hub.maxMissedPongs {
				c.close(websocket.CloseNormalClosure, "pong_timeout")
				return
			}
			if err := c.conn.WriteControl(websocket.PingMessage, []byte("ping"), now.Add(c.hub.writeWait)); err != nil {
				c.close(websocket.CloseGoingAway, "ping_failed")
				return
			}
		case <-staleTicker.C:
			now := c.hub.now()
			if c.idleFor(now) >= c.hub.idleNoUpdateAfter {
				c.close(CloseCodeIdleNoUpdates, CloseReasonIdleNoUpdates)
				return
			}
		}
	}
}

func (c *client) idleFor(now time.Time) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return now.Sub(c.lastUpdateSentAt)
}

func (c *client) incrementMissedPongs() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.missedPongs++
	return c.missedPongs
}

func (c *client) close(code int, reason string) {
	c.closeOnce.Do(func() {
		close(c.done)
		c.hub.removeClient(c)
		if code != 0 {
			if err := c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), c.hub.now().Add(c.hub.writeWait)); err != nil {
				c.hub.log.Warn(context.Background(), "failed writing websocket close control message", "account_id", c.accountID.String(), "error", err.Error())
			}
		}
		if err := c.conn.Close(); err != nil {
			c.hub.log.Warn(context.Background(), "failed closing websocket connection", "account_id", c.accountID.String(), "error", err.Error())
		}
		c.hub.log.Info(context.Background(), "websocket disconnected", "account_id", c.accountID.String(), "code", code, "reason", reason)
	})
}
