package events

import (
	"context"
	"encoding/json"
	"log"
)

// NoOpEventPublisher is a no-operation implementation of EventPublisher.
// It logs events to stdout instead of publishing to a message broker.
// Use this implementation when Kafka is not configured or in development.
type NoOpEventPublisher struct {
	// EnableLogging controls whether events are logged to stdout
	EnableLogging bool
}

// NewNoOpEventPublisher creates a new NoOpEventPublisher
func NewNoOpEventPublisher(enableLogging bool) *NoOpEventPublisher {
	return &NoOpEventPublisher{
		EnableLogging: enableLogging,
	}
}

// Publish logs the event if logging is enabled
func (p *NoOpEventPublisher) Publish(ctx context.Context, event Event) error {
	if p.EnableLogging {
		p.logEvent(event)
	}
	return nil
}

// PublishAsync logs the event asynchronously if logging is enabled
func (p *NoOpEventPublisher) PublishAsync(ctx context.Context, event Event) {
	if p.EnableLogging {
		go p.logEvent(event)
	}
}

// logEvent logs the event details to stdout
func (p *NoOpEventPublisher) logEvent(event Event) {
	metadata := event.GetMetadata()
	
	payload, _ := json.Marshal(event.GetPayload())
	
	log.Printf("[EVENT] type=%s id=%s actor=%s payload=%s",
		metadata.EventType,
		metadata.EventID,
		metadata.ActorID,
		string(payload),
	)
}
