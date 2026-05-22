package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductSegment represents a market segment for products
type ProductSegment struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for ProductSegment
func (ProductSegment) TableName() string {
	return "product_segments"
}

// BeforeCreate hook to generate UUID
func (p *ProductSegment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
