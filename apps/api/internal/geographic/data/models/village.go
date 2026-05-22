package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Village represents a village/kelurahan entity
type Village struct {
	ID         string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DistrictID string         `gorm:"type:uuid;not null;index" json:"district_id"`
	District   *District      `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	Name       string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Code       string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"`
	PostalCode string         `gorm:"type:varchar(10)" json:"postal_code"`
	Type       string         `gorm:"type:varchar(20);default:'village'" json:"type"` // village (desa), kelurahan
	IsActive   bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Village
func (Village) TableName() string {
	return "villages"
}

// BeforeCreate hook to generate UUID
func (v *Village) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	return nil
}
