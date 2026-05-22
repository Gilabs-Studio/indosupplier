package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Area represents a geographical/sales territory with optional map polygon outline.
// Enhanced in Sprint 24 to merge CRM Brick concept — adds territory code, polygon, color, manager.
type Area struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Sprint 24: CRM Territory enhancement fields
	Code      string  `gorm:"type:varchar(50);index" json:"code"`                  // Territory code (e.g. JAWA-BARAT)
	Polygon   *string `gorm:"type:jsonb" json:"polygon"`                           // GeoJSON polygon coordinates stored as JSONB
	Color     string  `gorm:"type:varchar(50)" json:"color"`                       // Display color on map (hex or CSS token)
	ManagerID *string `gorm:"type:uuid;index" json:"manager_id"`                   // FK to employees (area manager)
	Province  string  `gorm:"type:text" json:"province"`                           // Province name (display)
	Regency   string  `gorm:"type:text" json:"regency"`                            // Regency/City name (display)
	District  string  `gorm:"type:text" json:"district"`                           // District name (display)

	// Relations
	Manager       *Employee      `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	EmployeeAreas []EmployeeArea `gorm:"foreignKey:AreaID" json:"employee_areas,omitempty"`
}

// TableName specifies the table name for Area
func (Area) TableName() string {
	return "areas"
}

// BeforeCreate hook to generate UUID
func (a *Area) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}
