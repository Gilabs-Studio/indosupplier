package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Province represents a province/state entity
type Province struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CountryID string         `gorm:"type:uuid;not null;index" json:"country_id"`
	Country   *Country       `gorm:"foreignKey:CountryID" json:"country,omitempty"`
	Name      string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Code      string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"`
	Geometry  *string        `gorm:"type:jsonb" json:"geometry,omitempty"` // GeoJSON geometry (MultiPolygon/Polygon)
	IsActive  bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Cities    []City         `gorm:"foreignKey:ProvinceID" json:"cities,omitempty"`
}

// TableName specifies the table name for Province
func (Province) TableName() string {
	return "provinces"
}

// BeforeCreate hook to generate UUID
func (p *Province) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
