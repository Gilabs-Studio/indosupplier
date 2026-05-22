package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// City represents a city/regency entity
type City struct {
	ID         string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProvinceID string         `gorm:"type:uuid;not null;index" json:"province_id"`
	Province   *Province      `gorm:"foreignKey:ProvinceID" json:"province,omitempty"`
	Name       string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Code       string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"`
	Type       string         `gorm:"type:varchar(20);default:'city'" json:"type"` // city, regency (kabupaten)
	Geometry   *string        `gorm:"type:jsonb" json:"geometry,omitempty"` // GeoJSON geometry (MultiPolygon/Polygon)
	IsActive   bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	Districts  []District     `gorm:"foreignKey:CityID" json:"districts,omitempty"`
}

// TableName specifies the table name for City
func (City) TableName() string {
	return "cities"
}

// BeforeCreate hook to generate UUID
func (c *City) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
