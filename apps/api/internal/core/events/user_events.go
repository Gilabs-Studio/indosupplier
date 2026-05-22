package events

import (
	"context"
	"time"

	infraEvents "github.com/gilabs/gims/api/internal/core/infrastructure/events"
)

// UserCreatedPayload contains the data for a user created event
type UserCreatedPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	RoleID    string `json:"role_id"`
	Status    string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	infraEvents.BaseEvent
}

// NewUserCreatedEvent creates a new UserCreatedEvent
func NewUserCreatedEvent(ctx context.Context, payload UserCreatedPayload) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseEvent: infraEvents.BaseEvent{
			Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeUserCreated),
			Payload:  payload,
		},
	}
}

// UserUpdatedPayload contains the data for a user updated event
type UserUpdatedPayload struct {
	UserID       string                 `json:"user_id"`
	Email        string                 `json:"email,omitempty"`
	Name         string                 `json:"name,omitempty"`
	RoleID       string                 `json:"role_id,omitempty"`
	Status       string                 `json:"status,omitempty"`
	ChangedFields map[string]interface{} `json:"changed_fields,omitempty"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// UserUpdatedEvent represents a user update event
type UserUpdatedEvent struct {
	infraEvents.BaseEvent
}

// NewUserUpdatedEvent creates a new UserUpdatedEvent
func NewUserUpdatedEvent(ctx context.Context, payload UserUpdatedPayload) *UserUpdatedEvent {
	return &UserUpdatedEvent{
		BaseEvent: infraEvents.BaseEvent{
			Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeUserUpdated),
			Payload:  payload,
		},
	}
}

// UserDeletedPayload contains the data for a user deleted event
type UserDeletedPayload struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	DeletedAt time.Time `json:"deleted_at"`
}

// UserDeletedEvent represents a user deletion event
type UserDeletedEvent struct {
	infraEvents.BaseEvent
}

// NewUserDeletedEvent creates a new UserDeletedEvent
func NewUserDeletedEvent(ctx context.Context, payload UserDeletedPayload) *UserDeletedEvent {
	return &UserDeletedEvent{
		BaseEvent: infraEvents.BaseEvent{
			Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeUserDeleted),
			Payload:  payload,
		},
	}
}
