package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageRole represents the role of a chat message sender
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// AIChatMessage represents a single message in an AI chat session
type AIChatMessage struct {
	ID         string      `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	SessionID  string      `gorm:"type:uuid;not null;index:idx_ai_chat_messages_session" json:"session_id"`
	Role       MessageRole `gorm:"type:varchar(20);not null" json:"role"`
	Content    string      `gorm:"type:text;not null" json:"content"`
	Intent     *string     `gorm:"type:jsonb" json:"intent,omitempty"`
	TokenUsage *string     `gorm:"type:jsonb" json:"token_usage,omitempty"`
	Model      string      `gorm:"type:varchar(100)" json:"model,omitempty"`
	DurationMs int         `gorm:"column:duration_ms" json:"duration_ms,omitempty"`
	CreatedAt  time.Time   `gorm:"type:timestamptz;not null" json:"created_at"`

	// Relations
	Session   *AIChatSession `gorm:"foreignKey:SessionID" json:"-"`
	ActionLog *AIActionLog   `gorm:"foreignKey:MessageID" json:"action_log,omitempty"`
}

// TableName returns the table name for AIChatMessage
func (AIChatMessage) TableName() string {
	return "ai_chat_messages"
}

// BeforeCreate generates UUID before inserting
func (m *AIChatMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
