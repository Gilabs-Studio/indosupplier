package events

import "context"

// EventPublisher defines the interface for publishing domain events.
// This interface abstracts the underlying messaging system (Kafka, NATS, etc.)
// allowing for easy swapping of implementations.
type EventPublisher interface {
	// Publish publishes an event synchronously and returns an error if publishing fails.
	// Use this when you need to ensure the event is published before proceeding.
	Publish(ctx context.Context, event Event) error

	// PublishAsync publishes an event asynchronously (fire-and-forget).
	// Use this for non-critical events where you don't need delivery confirmation.
	PublishAsync(ctx context.Context, event Event)
}

// EventSubscriber defines the interface for subscribing to domain events.
// This will be implemented when Kafka consumer is added.
type EventSubscriber interface {
	// Subscribe registers a handler for a specific event type
	Subscribe(eventType EventType, handler EventHandler) error

	// Unsubscribe removes a handler for a specific event type
	Unsubscribe(eventType EventType) error

	// Start starts consuming events
	Start(ctx context.Context) error

	// Stop stops consuming events
	Stop() error
}

// EventHandler is a function type for handling events
type EventHandler func(ctx context.Context, event Event) error
