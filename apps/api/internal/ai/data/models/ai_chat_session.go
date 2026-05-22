package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionStatus represents the status of an AI chat session
type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "ACTIVE"
	SessionStatusClosed  SessionStatus = "CLOSED"
	SessionStatusExpired SessionStatus = "EXPIRED"
)

// AIChatSession represents an AI chat conversation session
type AIChatSession struct {
	ID           string         `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	UserID       string         `gorm:"type:uuid;not null;index:idx_ai_chat_sessions_user_id" json:"user_id"`
	Title        string         `gorm:"type:varchar(255)" json:"title"`
	Status       SessionStatus  `gorm:"type:varchar(20);default:ACTIVE;index:idx_ai_chat_sessions_status" json:"status"`
	LastActivity *time.Time     `gorm:"type:timestamptz" json:"last_activity"`
	MessageCount int            `gorm:"default:0" json:"message_count"`
	Metadata     *string        `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;not null;index:idx_ai_chat_sessions_created,sort:desc" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;not null" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Messages []AIChatMessage `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
	Actions  []AIActionLog   `gorm:"foreignKey:SessionID" json:"actions,omitempty"`
}

// TableName returns the table name for AIChatSession
func (AIChatSession) TableName() string {
	return "ai_chat_sessions"
}

// BeforeCreate generates UUID before inserting
func (s *AIChatSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.LastActivity == nil {
		now := apptime.Now()
		s.LastActivity = &now
	}
	return nil
}
