package events

import (
	"context"
	"time"

	infraEvents "github.com/gilabs/gims/api/internal/core/infrastructure/events"
)

// UserLoggedInPayload contains the data for a user login event
type UserLoggedInPayload struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	RoleCode  string    `json:"role_code"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	LoggedInAt time.Time `json:"logged_in_at"`
}

// UserLoggedInEvent represents a user login event
type UserLoggedInEvent struct {
	infraEvents.BaseEvent
}

// NewUserLoggedInEvent creates a new UserLoggedInEvent
func NewUserLoggedInEvent(ctx context.Context, payload UserLoggedInPayload) *UserLoggedInEvent {
	return &UserLoggedInEvent{
		BaseEvent: infraEvents.BaseEvent{
			Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeUserLoggedIn),
			Payload:  payload,
		},
	}
}

// UserLoggedOutPayload contains the data for a user logout event
type UserLoggedOutPayload struct {
	UserID      string    `json:"user_id"`
	LoggedOutAt time.Time `json:"logged_out_at"`
}

// UserLoggedOutEvent represents a user logout event
type UserLoggedOutEvent struct {
	infraEvents.BaseEvent
}

// NewUserLoggedOutEvent creates a new UserLoggedOutEvent
func NewUserLoggedOutEvent(ctx context.Context, payload UserLoggedOutPayload) *UserLoggedOutEvent {
	return &UserLoggedOutEvent{
		BaseEvent: infraEvents.BaseEvent{
			Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeUserLoggedOut),
			Payload:  payload,
		},
	}
}

// TokenRefreshedPayload contains the data for a token refresh event
type TokenRefreshedPayload struct {
	UserID       string    `json:"user_id"`
	OldTokenID   string    `json:"old_token_id,omitempty"`
	NewTokenID   string    `json:"new_token_id,omitempty"`
	RefreshedAt  time.Time `json:"refreshed_at"`
}

// TokenRefreshedEvent represents a token refresh event
type TokenRefreshedEvent struct {
	infraEvents.BaseEvent
}

// NewTokenRefreshedEvent creates a new TokenRefreshedEvent
func NewTokenRefreshedEvent(ctx context.Context, payload TokenRefreshedPayload) *TokenRefreshedEvent {
	return &TokenRefreshedEvent{
		BaseEvent: infraEvents.BaseEvent{
			Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeTokenRefreshed),
			Payload:  payload,
		},
	}
}
