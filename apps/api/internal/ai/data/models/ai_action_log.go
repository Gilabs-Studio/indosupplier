package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActionStatus represents the status of an AI action
type ActionStatus string

const (
	ActionStatusSuccess             ActionStatus = "SUCCESS"
	ActionStatusFailed              ActionStatus = "FAILED"
	ActionStatusPendingConfirmation ActionStatus = "PENDING_CONFIRMATION"
	ActionStatusCancelled           ActionStatus = "CANCELLED"
)

// ActionType represents the type of action performed
type ActionType string

const (
	ActionTypeCreate  ActionType = "CREATE"
	ActionTypeUpdate  ActionType = "UPDATE"
	ActionTypeDelete  ActionType = "DELETE"
	ActionTypeQuery   ActionType = "QUERY"
	ActionTypeApprove ActionType = "APPROVE"
	ActionTypeReport  ActionType = "REPORT"
)

// AIActionLog records every action performed by the AI assistant
type AIActionLog struct {
	ID              string       `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	SessionID       string       `gorm:"type:uuid;not null;index:idx_ai_action_logs_session" json:"session_id"`
	MessageID       *string      `gorm:"type:uuid;index" json:"message_id,omitempty"`
	UserID          string       `gorm:"type:uuid;not null;index:idx_ai_action_logs_user" json:"user_id"`
	Intent          string       `gorm:"type:varchar(100);not null;index:idx_ai_action_logs_intent" json:"intent"`
	EntityType      string       `gorm:"type:varchar(100)" json:"entity_type,omitempty"`
	EntityID        *string      `gorm:"type:uuid" json:"entity_id,omitempty"`
	Action          ActionType   `gorm:"type:varchar(20);not null" json:"action"`
	RequestPayload  *string      `gorm:"type:jsonb" json:"request_payload,omitempty"`
	ResponsePayload *string      `gorm:"type:jsonb" json:"response_payload,omitempty"`
	Status          ActionStatus `gorm:"type:varchar(20);not null;index:idx_ai_action_logs_status" json:"status"`
	ErrorMessage    string       `gorm:"type:text" json:"error_message,omitempty"`
	PermissionUsed  string       `gorm:"type:varchar(100)" json:"permission_used,omitempty"`
	DurationMs      int          `gorm:"column:duration_ms" json:"duration_ms,omitempty"`
	IPAddress       string       `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent       string       `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt       time.Time    `gorm:"type:timestamptz;not null;index:idx_ai_action_logs_created,sort:desc" json:"created_at"`

	// Relations
	Session *AIChatSession `gorm:"foreignKey:SessionID" json:"-"`
	Message *AIChatMessage `gorm:"foreignKey:MessageID" json:"-"`
}

// TableName returns the table name for AIActionLog
func (AIActionLog) TableName() string {
	return "ai_action_logs"
}

// BeforeCreate generates UUID before inserting
func (a *AIActionLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}
