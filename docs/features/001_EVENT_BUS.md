# EventBus Feature Specification

## Overview

The goal is to build an asynchronous, strongly-typed event bus in Go that initially works in-memory but is designed from the beginning to support external brokers such as:

- Kafka
- RabbitMQ
- AWS SQS
- NATS

The system should support:

- Strongly typed handlers using Go generics
- Middleware
- Context propagation
- Correlation/tracing metadata
- Ordered topic queues
- Concurrent handler execution
- Global worker pool
- Broker-friendly abstractions

---

# Design Goals

## Primary Goals

- Async event publishing
- Strong typing at the API boundaries
- Internal envelope abstraction
- Ordered dequeue behavior per topic
- Concurrent handler execution
- Middleware support
- Broker portability
- Clean and idiomatic Go API

## Non-Goals (For Now)

- Dead letter queues
- Retries
- Wildcard topic subscriptions
- Persistent queues
- Distributed delivery guarantees
- Exactly-once delivery

---

# Public API

## Publishing

```go
err := bus.Publish(
    ctx,
    "order.created",
    OrderCreated{
        OrderID: "123",
    },
    eventbus.WithCorrelationID("corr-id"),
    eventbus.WithMetadata("tenant", "acme"),
)
```

---

## Subscribing

```go
sub, err := eventbus.Subscribe(
    bus,
    "order.created",
    func(
        ctx context.Context,
        event OrderCreated,
        meta eventbus.Metadata,
    ) error {
        return nil
    },
)

defer sub.Close()
```

---

# Core Concepts

## Envelope

The envelope is created internally by the bus.

Handlers do not directly receive the envelope.

```go
type Envelope[T any] struct {
    EventID       string
    Topic         string
    Data          T
    CorrelationID string
    CausationID   string
    TraceID       string
    OccurredAt    time.Time
    Metadata      map[string]string
}
```

---

## Metadata

Metadata combines tracing and arbitrary metadata values.

```go
type Metadata struct {
    EventID       string
    CorrelationID string
    CausationID   string
    TraceID       string
    OccurredAt    time.Time
    Values        map[string]string
}
```

---

## Handler

```go
type Handler[T any] func(
    ctx context.Context,
    event T,
    meta Metadata,
) error
```

---

## Subscription

Subscriptions should be closable.

```go
type Subscription interface {
    Close() error
}
```

---

# Processing Model

## Topic Queues

Each topic owns an ordered queue.

Example:

```text
order.created
payment.completed
invoice.generated
```

Each topic queue preserves dequeue order.

---

## Worker Pool

A global worker pool processes events.

Workers pull events from all topic queues.

Workers are not dedicated to a single topic.

---

## Event Ordering

The system guarantees:

- dequeue order per topic

The system does NOT guarantee:

- completion order per topic

This is valid behavior:

```text
Event #1 dequeued first
Event #2 dequeued second

Event #2 finishes before Event #1
```

This avoids unnecessary backpressure.

---

# Handler Execution

For each event:

- all handlers of the topic are discovered
- handlers execute concurrently

Example:

```text
OrderCreated
 ├── SendEmail
 ├── UpdateProjection
 └── PublishAnalytics
```

All handlers may run simultaneously.

---

# Error Handling

## Current Behavior

- handler errors do NOT stop other handlers
- handler errors are ignored by default

---

## Panic Recovery

Panics should be recovered automatically.

Recovered panics should become handler errors.

The process should never crash due to a handler panic.

---

# Middleware

Middleware should wrap handlers.

## Middleware Interface

```go
type Middleware func(next HandlerFunc) HandlerFunc
```

---

## Example Use Cases

- logging
- tracing
- metrics
- panic recovery
- retries
- authorization
- tenant propagation

---

# Backpressure

Publishing supports configurable backpressure strategies.

## Options

```go
WithBackpressure(BlockWhenFull)
WithBackpressure(FailWhenFull)
WithBackpressure(DropWhenFull)
```

---

# Topic Matching

Currently supported:

- exact topic matching only

Not yet supported:

```text
order.*
```

---

# Broker Readiness

The public API should not expose in-memory-only concepts.

The architecture should allow:

```text
InMemoryBus
KafkaBus
RabbitMQBus
SQSBus
NATSBus
```

without changing application code.

---

# Suggested Package Structure

```text
eventbus/
  bus.go
  envelope.go
  metadata.go
  options.go
  subscription.go
  handler.go
  middleware.go

  inmemory/
    bus.go
    topic_queue.go
    worker_pool.go

transport/
    messaging/
        
```

---

# Implementation Milestones

## Phase 1

- Define envelope
- Define metadata
- Define subscription model
- Define publish options

## Phase 2

- Implement in-memory bus
- Implement topic queues
- Implement async publish

## Phase 3

- Implement worker pool
- Implement concurrent handler execution

## Phase 4

- Implement middleware chain
- Implement panic recovery middleware

## Phase 5

- Implement backpressure strategies
- Add cancellation support

## Phase 6

- Add tests
    - ordering
    - concurrency
    - unsubscribe
    - panic recovery
    - queue overflow
    - cancellation
