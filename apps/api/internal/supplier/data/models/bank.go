package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Bank represents a bank entity for supplier bank accounts
type Bank struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Code        string         `gorm:"type:varchar(20);index" json:"code"`
	SwiftCode   string         `gorm:"type:varchar(20)" json:"swift_code"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Bank
func (Bank) TableName() string {
	return "banks"
}

// BeforeCreate hook to generate UUID
func (b *Bank) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
