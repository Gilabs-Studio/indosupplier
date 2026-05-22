package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog represents an audit log entry for sensitive actions
type AuditLog struct {
	ID             string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActorID        string         `gorm:"type:uuid;index;not null" json:"actor_id"` // User who performed the action
	PermissionCode string         `gorm:"type:varchar(100);index;not null" json:"permission_code"`
	TargetID       string         `gorm:"type:varchar(255);index;not null" json:"target_id"` // ID of the affected resource
	Action         string         `gorm:"type:varchar(50);not null" json:"action"`           // CREATE, UPDATE, DELETE, etc.
	IPAddress      string         `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent      string         `gorm:"type:text" json:"user_agent"`
	ResultStatus   string         `gorm:"type:varchar(50)" json:"result_status"`
	Reason         string         `gorm:"type:text" json:"reason,omitempty"`
	Metadata       string         `gorm:"type:jsonb" json:"metadata"` // Additional details (JSON)
	Changes        string         `gorm:"type:jsonb" json:"changes"`  // Array of FieldChange JSON
	CreatedAt      time.Time      `gorm:"index" json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate hook to generate UUID
func (l *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
