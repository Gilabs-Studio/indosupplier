package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Country represents a country entity
type Country struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Code      string         `gorm:"type:varchar(10);uniqueIndex;not null" json:"code"` // ISO code e.g., "ID", "US"
	PhoneCode string         `gorm:"type:varchar(10)" json:"phone_code"`                // e.g., "+62"
	IsActive  bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Provinces []Province     `gorm:"foreignKey:CountryID" json:"provinces,omitempty"`
}

// TableName specifies the table name for Country
func (Country) TableName() string {
	return "countries"
}

// BeforeCreate hook to generate UUID
func (c *Country) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
