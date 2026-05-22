package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Reminder represents a notification trigger for a task
type Reminder struct {
	ID           string     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	TaskID       string     `gorm:"type:uuid;not null;index" json:"task_id"`
	RemindAt     time.Time  `gorm:"not null;index" json:"remind_at"`
	ReminderType string     `gorm:"type:varchar(20);default:'in_app'" json:"reminder_type"` // in_app, email
	IsSent       bool       `gorm:"default:false" json:"is_sent"`
	SentAt       *time.Time `json:"sent_at"`
	Message      string     `gorm:"type:text" json:"message"`
	CreatedBy    *string    `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
}

// TableName returns the database table name
func (Reminder) TableName() string {
	return "crm_reminders"
}

// BeforeCreate generates a UUID if not set
func (r *Reminder) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
