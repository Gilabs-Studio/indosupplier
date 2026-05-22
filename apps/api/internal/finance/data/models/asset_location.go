package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetLocation struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	CompanyID *string `gorm:"type:uuid;index" json:"company_id,omitempty"`
	Company   *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`

	Name        string   `gorm:"type:varchar(150);not null;uniqueIndex" json:"name"`
	Description string   `gorm:"type:text" json:"description"`
	Address     string   `gorm:"type:text" json:"address"`
	Latitude    *float64 `gorm:"type:numeric(10,7)" json:"latitude"`
	Longitude   *float64 `gorm:"type:numeric(10,7)" json:"longitude"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AssetLocation) TableName() string {
	return "asset_locations"
}

func (l *AssetLocation) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
