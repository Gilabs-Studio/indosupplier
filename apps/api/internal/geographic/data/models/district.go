package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// District represents a district/kecamatan entity
type District struct {
	ID        string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CityID    string         `gorm:"type:uuid;not null;index" json:"city_id"`
	City      *City          `gorm:"foreignKey:CityID" json:"city,omitempty"`
	Name      string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Code      string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"`
	Geometry  *string        `gorm:"type:jsonb" json:"geometry,omitempty"` // GeoJSON geometry (MultiPolygon/Polygon)
	IsActive  bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Villages  []Village      `gorm:"foreignKey:DistrictID" json:"villages,omitempty"`
}

// TableName specifies the table name for District
func (District) TableName() string {
	return "districts"
}

// BeforeCreate hook to generate UUID
func (d *District) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}
