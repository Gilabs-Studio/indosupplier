package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserWarehouse represents the many-to-many assignment between users and warehouses (POS outlets)
type UserWarehouse struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    string         `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	UserID      string         `gorm:"type:uuid;not null;uniqueIndex:idx_user_warehouse" json:"user_id"`
	WarehouseID string         `gorm:"type:uuid;not null;uniqueIndex:idx_user_warehouse" json:"warehouse_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for UserWarehouse
func (UserWarehouse) TableName() string {
	return "user_warehouses"
}

// BeforeCreate hook to generate UUID
func (uw *UserWarehouse) BeforeCreate(tx *gorm.DB) error {
	if uw.ID == "" {
		uw.ID = uuid.New().String()
	}
	return nil
}
