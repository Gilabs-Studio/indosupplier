# Event-Driven Architecture Standards

Standarisasi arsitektur event-driven untuk API backend. Arsitektur ini memungkinkan decoupling antar domain dan mendukung integrasi dengan message broker seperti Kafka.

---

## 1. Overview

### Arsitektur

```
┌──────────────┐     ┌────────────────┐     ┌─────────────────┐
│   Usecase    │────▶│ EventPublisher │────▶│  Message Broker │
│              │     │   (Interface)  │     │  (Kafka/NoOp)   │
└──────────────┘     └────────────────┘     └─────────────────┘
                                                     │
                                                     ▼
                                            ┌─────────────────┐
                                            │    Consumers    │
                                            │ (Audit, Cache,  │
                                            │  Notification)  │
                                            └─────────────────┘
```

### Prinsip

- **Fire-and-Forget**: Events dipublish async untuk performance
- **Eventual Consistency**: Side-effects dihandle oleh consumers
- **Decoupling**: Domain tidak perlu tahu siapa yang consume events
- **Extensibility**: Mudah menambah consumer baru tanpa ubah usecase

---

## 2. Project Structure

```
internal/core/
├── events/                         # Domain events
│   ├── user_events.go             # User domain events
│   ├── role_events.go             # Role domain events
│   └── auth_events.go             # Auth domain events
└── infrastructure/
    └── events/                     # Event infrastructure
        ├── types.go               # Base event types & interfaces
        ├── publisher.go           # EventPublisher interface
        └── noop_publisher.go      # NoOp implementation (dev/test)
```

---

## 3. Event Structure

### EventMetadata

Setiap event memiliki metadata standar:

```go
type EventMetadata struct {
    EventID       string    `json:"event_id"`        // UUID unik per event
    EventType     EventType `json:"event_type"`      // Tipe event (e.g., "user.created")
    Timestamp     time.Time `json:"timestamp"`       // Waktu event dibuat
    CorrelationID string    `json:"correlation_id"`  // Untuk tracing
    ActorID       string    `json:"actor_id"`        // User yang trigger event
    ActorEmail    string    `json:"actor_email"`     // Email actor
    IPAddress     string    `json:"ip_address"`      // IP address
    UserAgent     string    `json:"user_agent"`      // User agent
}
```

### Event Interface

```go
type Event interface {
    GetType() EventType
    GetMetadata() EventMetadata
    GetPayload() interface{}
}
```

### Event Naming Convention

Format: `<domain>.<action>`

| Domain | Events |
|--------|--------|
| `user` | `user.created`, `user.updated`, `user.deleted` |
| `role` | `role.created`, `role.updated`, `role.deleted`, `role.permissions_assigned` |
| `auth` | `auth.user_logged_in`, `auth.user_logged_out`, `auth.token_refreshed` |

---

## 4. EventPublisher Interface

```go
type EventPublisher interface {
    // Sync publish - returns error if failed
    Publish(ctx context.Context, event Event) error
    
    // Async publish - fire-and-forget
    PublishAsync(ctx context.Context, event Event)
}
```

### Implementations

| Implementation | Use Case |
|---------------|----------|
| `NoOpEventPublisher` | Development, testing (logs to stdout) |
| `KafkaEventPublisher` | Production (publish ke Kafka) |

---

## 5. Creating New Events

### Step 1: Define Event Type Constant

```go
// internal/core/infrastructure/events/types.go
const (
    EventTypeOrderCreated EventType = "order.created"
    EventTypeOrderUpdated EventType = "order.updated"
)
```

### Step 2: Create Event Payload & Event

```go
// internal/core/events/order_events.go
package events

type OrderCreatedPayload struct {
    OrderID   string    `json:"order_id"`
    UserID    string    `json:"user_id"`
    Total     float64   `json:"total"`
    CreatedAt time.Time `json:"created_at"`
}

type OrderCreatedEvent struct {
    infraEvents.BaseEvent
}

func NewOrderCreatedEvent(ctx context.Context, payload OrderCreatedPayload) *OrderCreatedEvent {
    return &OrderCreatedEvent{
        BaseEvent: infraEvents.BaseEvent{
            Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeOrderCreated),
            Payload:  payload,
        },
    }
}
```

### Step 3: Publish from Usecase

```go
// internal/order/domain/usecase/order_usecase.go
func (u *orderUsecase) Create(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
    // ... create order logic ...
    
    // Publish event (async, fire-and-forget)
    u.eventPublisher.PublishAsync(ctx, events.NewOrderCreatedEvent(ctx, events.OrderCreatedPayload{
        OrderID:   order.ID,
        UserID:    order.UserID,
        Total:     order.Total,
        CreatedAt: order.CreatedAt,
    }))
    
    return resp, nil
}
```

---

## 6. Kafka Integration (Future)

Ketika Kafka sudah disetup, ganti `NoOpEventPublisher` dengan `KafkaEventPublisher`:

```go
// main.go
// Before (NoOp)
eventPublisher := events.NewNoOpEventPublisher(true)

// After (Kafka)
eventPublisher := kafka.NewKafkaEventPublisher(kafkaConfig)
```

### Kafka Topics

| Event Type | Topic |
|------------|-------|
| `user.*` | `crm.user.events` |
| `role.*` | `crm.role.events` |
| `auth.*` | `crm.auth.events` |

---

## 7. Best Practices

### DO ✅

- Publish events SETELAH operasi database berhasil
- Gunakan `PublishAsync` untuk non-critical events
- Include semua data yang diperlukan consumer di payload
- Gunakan event untuk audit, cache invalidation, notifications

### DON'T ❌

- Jangan publish event SEBELUM operasi berhasil
- Jangan rely on event untuk business logic utama
- Jangan include sensitive data (password, tokens) di payload
- Jangan block request untuk menunggu event publish

---

## 8. Testing

### Unit Test dengan Mock

```go
type mockEventPublisher struct {
    publishedEvents []events.Event
}

func (m *mockEventPublisher) Publish(ctx context.Context, event events.Event) error {
    m.publishedEvents = append(m.publishedEvents, event)
    return nil
}

func (m *mockEventPublisher) PublishAsync(ctx context.Context, event events.Event) {
    m.publishedEvents = append(m.publishedEvents, event)
}
```

---

## Related Documentation

- [API Folder Structure](./api-folder-structure.md)
- [API Response Standards](./api-response-standards.md)
- [API Enterprise Scenarios](./api-enterprise-scenarios.md)
