package events

import (
	"context"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
)

// EventType represents the type of domain event
type EventType string

// Event types for different domains
const (
	// User events
	EventTypeUserCreated EventType = "user.created"
	EventTypeUserUpdated EventType = "user.updated"
	EventTypeUserDeleted EventType = "user.deleted"

	// Role events
	EventTypeRoleCreated             EventType = "role.created"
	EventTypeRoleUpdated             EventType = "role.updated"
	EventTypeRoleDeleted             EventType = "role.deleted"
	EventTypeRolePermissionsAssigned EventType = "role.permissions_assigned"

	// Auth events
	EventTypeUserLoggedIn    EventType = "auth.user_logged_in"
	EventTypeUserLoggedOut   EventType = "auth.user_logged_out"
	EventTypeTokenRefreshed  EventType = "auth.token_refreshed"
)

// EventMetadata contains common metadata for all events
type EventMetadata struct {
	EventID       string    `json:"event_id"`
	EventType     EventType `json:"event_type"`
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	ActorID       string    `json:"actor_id,omitempty"`
	ActorEmail    string    `json:"actor_email,omitempty"`
	IPAddress     string    `json:"ip_address,omitempty"`
	UserAgent     string    `json:"user_agent,omitempty"`
}

// Event is the base interface that all domain events must implement
type Event interface {
	// GetType returns the event type
	GetType() EventType
	// GetMetadata returns the event metadata
	GetMetadata() EventMetadata
	// GetPayload returns the event-specific payload
	GetPayload() interface{}
}

// BaseEvent provides a base implementation for common event fields
type BaseEvent struct {
	Metadata EventMetadata `json:"metadata"`
	Payload  interface{}   `json:"payload"`
}

// GetType returns the event type
func (e *BaseEvent) GetType() EventType {
	return e.Metadata.EventType
}

// GetMetadata returns the event metadata
func (e *BaseEvent) GetMetadata() EventMetadata {
	return e.Metadata
}

// GetPayload returns the event payload
func (e *BaseEvent) GetPayload() interface{} {
	return e.Payload
}

// NewEventMetadata creates a new EventMetadata with generated ID and timestamp
func NewEventMetadata(eventType EventType) EventMetadata {
	return EventMetadata{
		EventID:   uuid.NewString(),
		EventType: eventType,
		Timestamp: apptime.Now(),
	}
}

// NewEventMetadataWithContext creates EventMetadata extracting actor info from context
func NewEventMetadataWithContext(ctx context.Context, eventType EventType) EventMetadata {
	meta := NewEventMetadata(eventType)

	// Extract actor info from context (set by auth middleware)
	if v := ctx.Value("user_id"); v != nil {
		meta.ActorID = v.(string)
	}
	if v := ctx.Value("user_email"); v != nil {
		meta.ActorEmail = v.(string)
	}
	if v := ctx.Value("client_ip"); v != nil {
		meta.IPAddress = v.(string)
	}
	if v := ctx.Value("user_agent"); v != nil {
		meta.UserAgent = v.(string)
	}
	if v := ctx.Value("correlation_id"); v != nil {
		meta.CorrelationID = v.(string)
	}

	return meta
}
