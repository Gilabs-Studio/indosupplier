package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	NotificationTypeApprovalRequest = "approval_request"
	NotificationTypeInfo            = "info"
	NotificationTypeWarning         = "warning"
)

// Notification is a user-targeted in-app notification.
type Notification struct {
	ID         string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	UserID     string         `gorm:"type:uuid;not null;index:idx_notifications_user_created" json:"user_id"`
	Type       string         `gorm:"type:varchar(50);not null;index" json:"type"`
	Title      string         `gorm:"type:varchar(255);not null" json:"title"`
	Message    string         `gorm:"type:text;not null" json:"message"`
	EntityType string         `gorm:"type:varchar(100);not null;index" json:"entity_type"`
	EntityID   string         `gorm:"type:varchar(100);not null;index" json:"entity_id"`
	IsRead     bool           `gorm:"type:boolean;not null;default:false;index:idx_notifications_user_read" json:"is_read"`
	ReadAt     *time.Time     `json:"read_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}
