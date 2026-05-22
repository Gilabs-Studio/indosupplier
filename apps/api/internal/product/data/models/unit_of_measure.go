package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UnitOfMeasure represents a unit of measure for products
type UnitOfMeasure struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Symbol      string         `gorm:"type:varchar(20);not null" json:"symbol"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for UnitOfMeasure
func (UnitOfMeasure) TableName() string {
	return "units_of_measure"
}

// BeforeCreate hook to generate UUID
func (u *UnitOfMeasure) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
