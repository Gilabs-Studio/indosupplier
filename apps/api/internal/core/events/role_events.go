package events

import (
	"context"
	"time"

	infraEvents "github.com/gilabs/gims/api/internal/core/infrastructure/events"
)

// RoleCreatedPayload contains the data for a role created event
type RoleCreatedPayload struct {
	RoleID      string    `json:"role_id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// RoleCreatedEvent represents a role creation event
type RoleCreatedEvent struct {
	infraEvents.BaseEvent
}

// NewRoleCreatedEvent creates a new RoleCreatedEvent
func NewRoleCreatedEvent(ctx context.Context, payload RoleCreatedPayload) *RoleCreatedEvent {
	return &RoleCreatedEvent{
		BaseEvent: infraEvents.BaseEvent{
			Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeRoleCreated),
			Payload:  payload,
		},
	}
}

// RoleUpdatedPayload contains the data for a role updated event
type RoleUpdatedPayload struct {
	RoleID        string                 `json:"role_id"`
	Name          string                 `json:"name,omitempty"`
	Code          string                 `json:"code,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Status        string                 `json:"status,omitempty"`
	ChangedFields map[string]interface{} `json:"changed_fields,omitempty"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// RoleUpdatedEvent represents a role update event
type RoleUpdatedEvent struct {
	infraEvents.BaseEvent
}

// NewRoleUpdatedEvent creates a new RoleUpdatedEvent
func NewRoleUpdatedEvent(ctx context.Context, payload RoleUpdatedPayload) *RoleUpdatedEvent {
	return &RoleUpdatedEvent{
		BaseEvent: infraEvents.BaseEvent{
			Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeRoleUpdated),
			Payload:  payload,
		},
	}
}

// RoleDeletedPayload contains the data for a role deleted event
type RoleDeletedPayload struct {
	RoleID    string    `json:"role_id"`
	Code      string    `json:"code"`
	DeletedAt time.Time `json:"deleted_at"`
}

// RoleDeletedEvent represents a role deletion event
type RoleDeletedEvent struct {
	infraEvents.BaseEvent
}

// NewRoleDeletedEvent creates a new RoleDeletedEvent
func NewRoleDeletedEvent(ctx context.Context, payload RoleDeletedPayload) *RoleDeletedEvent {
	return &RoleDeletedEvent{
		BaseEvent: infraEvents.BaseEvent{
			Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeRoleDeleted),
			Payload:  payload,
		},
	}
}

// RolePermissionsAssignedPayload contains the data for permissions assigned event
type RolePermissionsAssignedPayload struct {
	RoleID        string    `json:"role_id"`
	PermissionIDs []string  `json:"permission_ids"`
	AssignedAt    time.Time `json:"assigned_at"`
}

// RolePermissionsAssignedEvent represents a role permissions assignment event
type RolePermissionsAssignedEvent struct {
	infraEvents.BaseEvent
}

// NewRolePermissionsAssignedEvent creates a new RolePermissionsAssignedEvent
func NewRolePermissionsAssignedEvent(ctx context.Context, payload RolePermissionsAssignedPayload) *RolePermissionsAssignedEvent {
	return &RolePermissionsAssignedEvent{
		BaseEvent: infraEvents.BaseEvent{
			Metadata: infraEvents.NewEventMetadataWithContext(ctx, infraEvents.EventTypeRolePermissionsAssigned),
			Payload:  payload,
		},
	}
}
