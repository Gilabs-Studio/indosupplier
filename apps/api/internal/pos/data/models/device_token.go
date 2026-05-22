package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type POSDeviceToken struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  string         `gorm:"column:tenant_id;type:uuid;not null;index:idx_pos_device_tokens_scope" json:"tenant_id"`
	UserID    string         `gorm:"type:uuid;not null;index" json:"user_id"`
	OutletID  string         `gorm:"type:uuid;not null;index:idx_pos_device_tokens_scope" json:"outlet_id"`
	Platform  string         `gorm:"type:varchar(20);not null;index" json:"platform"`
	Token     string         `gorm:"type:text;not null;uniqueIndex:uidx_pos_device_tokens_token" json:"token"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (POSDeviceToken) TableName() string {
	return "pos_device_tokens"
}

func (t *POSDeviceToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
