package memorybus

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"go.elastic.co/apm/v2"
)

type BackpressurePolicy int

const (
	BackpressureError BackpressurePolicy = iota
	BackpressureDrop
	BackpressureBlock
)

type Option func(*MemoryBus)

func WithWorkers(n int) Option {
	return func(b *MemoryBus) {
		if n > 0 {
			b.workers = n
		}
	}
}

func WithQueueSize(n int) Option {
	return func(mb *MemoryBus) {
		if n > 0 {
			mb.queueSize = n
		}
	}
}

func WithBackpressure(p BackpressurePolicy) Option {
	return func(mb *MemoryBus) { mb.backpressure = p }
}

type MemoryBus struct {
	mu sync.RWMutex

	closed atomic.Bool
	wg     sync.WaitGroup

	subs   map[string]map[uint64]*subscriber
	nextID uint64

	ready chan *subscriber

	workers      int
	queueSize    int
	backpressure BackpressurePolicy
}

var (
	ErrBusClosed          = errors.New("bus is closed")
	ErrInvalidType        = errors.New("event type required")
	ErrHandlerNil         = errors.New("handler required")
	ErrBackpressure       = errors.New("subscriber mailbox full")
	ErrTopicCannotBeEmpty = errors.New("topic cannot be empty")
)

// Ensure MemoryBus satisfies the eventbus.Bus interface.
var _ eventbus.Bus = (*MemoryBus)(nil)

func NewMemoryBus(opts ...Option) *MemoryBus {
	b := &MemoryBus{
		subs:    make(map[string]map[uint64]*subscriber),
		workers: runtime.GOMAXPROCS(0),
	}

	for _, opt := range opts {
		opt(b)
	}

	b.ready = make(chan *subscriber, b.workers*8)

	b.wg.Add(b.workers)
	for i := 0; i < b.workers; i++ {
		go b.worker()
	}

	return b
}

// Publish implements eventbus.Bus by building an envelope from the topic,
// payload, and options and dispatching it to the topic's subscribers.
func (b *MemoryBus) Publish(ctx context.Context, topic string, payload any, opts ...eventbus.PublishOption) error {
	return b.dispatch(ctx, eventbus.NewEnvelope(ctx, topic, payload, opts...))
}

// dispatch routes an already-built envelope to the subscribers of its topic,
// honoring the configured backpressure policy.
func (b *MemoryBus) dispatch(ctx context.Context, envelope eventbus.Envelope) error {
	if b.closed.Load() {
		return ErrBusClosed
	}
	if envelope.Topic == "" {
		return ErrTopicCannotBeEmpty
	}
	if envelope.OccurredAt.IsZero() {
		envelope.OccurredAt = time.Now()
	}
	// acquire lock and check if any are subscribed to the topic
	// return nil if none are subscribed to the topic
	b.mu.RLock()
	m := b.subs[envelope.Topic]
	if len(m) == 0 {
		b.mu.RUnlock()
		return nil
	}
	// create a new list and add the pointers of the subscribers
	// to this list. We do this so that we can iterate over this
	// list instead of keeping a potentially long lock on the subscriptions
	// on the bus.
	list := make([]*subscriber, 0, len(m))
	for _, s := range m {
		list = append(list, s)
	}
	b.mu.RUnlock()
	// populate the queues of the subscribers according to the
	// backpressure setting and schedule the subscriber
	for _, s := range list {
		if s.closed.Load() {
			continue
		}
		switch b.backpressure {
		case BackpressureDrop:
			select {
			case s.queue <- envelope:
				b.schedule(s)
			default:
				// drop for this subscriber
			}
		case BackpressureBlock:
			select {
			case <-ctx.Done():
				return ctx.Err()
			case s.queue <- envelope:
				b.schedule(s)
			}
		default: // BackpressureError
			select {
			case s.queue <- envelope:
				b.schedule(s)
			default:
				return ErrBackpressure
			}
		}
	}
	return nil
}

func (b *MemoryBus) Subscribe(topic string, h eventbus.HandlerFunc) (eventbus.Subscription, error) {
	// should return error when the bus is closed
	if b.closed.Load() {
		return nil, ErrBusClosed
	}
	// return invalid type if type is empty
	if topic == "" {
		return nil, ErrTopicCannotBeEmpty
	}
	// handler nees to be filled
	if h == nil {
		return nil, ErrHandlerNil
	}
	// assign id to subscriber
	id := atomic.AddUint64(&b.nextID, 1)
	// create subscriber
	s := &subscriber{
		id:      id,
		topic:   topic,
		handler: h,
		queue:   make(chan eventbus.Envelope, b.queueSize),
	}
	// add subscriber to routing table
	b.mu.Lock()
	if b.subs[topic] == nil {
		b.subs[topic] = make(map[uint64]*subscriber)
	}
	b.subs[topic][id] = s
	b.mu.Unlock()
	return subscription{bus: b, sub: s}, nil
}

func (b *MemoryBus) Close() error {
	if b.closed.Swap(true) {
		return nil
	}
	// Unsubscribe everything
	b.mu.Lock()
	for _, m := range b.subs {
		for _, s := range m {
			s.closed.Store(true)
		}
	}
	b.subs = make(map[string]map[uint64]*subscriber)
	b.mu.Unlock()
	// Stop workers: close ready channel after all in-flight scheduling settles.
	close(b.ready)

	b.wg.Wait()
	return nil
}

func (b *MemoryBus) worker() {
	// when worker crashes or stop make sure to update the wait group
	// so that the initiator knows this worker is done.
	defer b.wg.Done()
	// loop over
	for s := range b.ready {
		if s == nil || s.closed.Load() {
			continue
		}

		var envelope eventbus.Envelope
		var ok bool
		select {
		case envelope, ok = <-s.queue: // consume from queue
			// could not consume unschedule and continue
			if !ok {
				s.scheduled.Store(0)
				continue
			}
		default:
			// nothing to consume, unschedule and continue
			s.scheduled.Store(0)
			continue
		}
		// Carry the message metadata so that events published by the handler
		// inherit this message's correlation chain and are marked as caused by it.
		hctx := eventbus.ContextWithMetadata(context.Background(), eventbus.MetadataFromEnvelope(envelope))
		// TODO: implement a timeout to prevent goroutine exhaustion
		if err := s.handler(hctx, envelope); err != nil {
			apm.CaptureError(context.Background(), err).Send()
		}
		if len(s.queue) > 0 && !s.closed.Load() {
			// keep scheduled=1, requeue
			b.ready <- s
		} else {
			s.scheduled.Store(0)
		}
	}
}

func (b *MemoryBus) schedule(s *subscriber) {
	// cannot schedule bc subscriber is closed
	if s.closed.Load() {
		return
	}
	// if subscriber is not scheduled, schedule it by changing the value of scheduled to 1
	if s.scheduled.CompareAndSwap(0, 1) {
		// we only want to recover when scheduling a subscriber fails, hence the defer is
		// in this if statement.
		defer func() {
			if recovered := recover(); recovered != nil {
				apm.CaptureError(context.Background(), fmt.Errorf("memory bus schedule panic: %v", recovered)).Send()
			}
		}()
		// add subscriber to ready queue to be handled
		b.ready <- s
	}
}

type subscriber struct {
	id      uint64
	topic   string
	handler eventbus.HandlerFunc

	queue chan eventbus.Envelope

	// scheduled = 0/1 indicates whether subscriber is currently enqueued in the scheduling pool
	scheduled atomic.Int32

	// closed indicates if the subscriber is unsubscribed; used to ignore future envelope deliveries
	closed atomic.Bool
}

type subscription struct {
	bus *MemoryBus
	sub *subscriber
}

func (s subscription) Close() error {
	if s.bus.closed.Load() {
		// bus is closing anyway; still mark subscriber closed
		s.sub.closed.Store(true)
		return nil
	}
	// Remove from routing table
	s.bus.mu.Lock()
	m := s.bus.subs[s.sub.topic]
	if m != nil {
		delete(m, s.sub.id)
		if len(m) == 0 {
			delete(s.bus.subs, s.sub.topic)
		}
	}
	s.bus.mu.Unlock()
	// Mark closed so publishers/workers ignore it
	s.sub.closed.Store(true)
	// Optional: drain mailbox to help GC; not necessary for correctness.
	for {
		select {
		case <-s.sub.queue:
		default:
			return nil
		}
	}
}
