package models

import (
	"time"

	geographic "github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Outlet represents a physical outlet/store location that can be linked to a warehouse
type Outlet struct {
	ID          string               `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code        string               `gorm:"type:varchar(50);uniqueIndex" json:"code"`
	Name        string               `gorm:"type:varchar(200);not null;index" json:"name"`
	Description string               `gorm:"type:text" json:"description"`
	Phone       string               `gorm:"type:varchar(50)" json:"phone"`
	Email       string               `gorm:"type:varchar(100)" json:"email"`
	Address     string               `gorm:"type:text" json:"address"`

	// Geographic references
	ProvinceID *string              `gorm:"type:uuid;index" json:"province_id"`
	Province   *geographic.Province `gorm:"foreignKey:ProvinceID" json:"province,omitempty"`
	CityID     *string              `gorm:"type:uuid;index" json:"city_id"`
	City       *geographic.City     `gorm:"foreignKey:CityID" json:"city,omitempty"`
	DistrictID *string              `gorm:"type:uuid;index" json:"district_id"`
	District   *geographic.District `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	VillageID  *string              `gorm:"type:uuid;index" json:"village_id"`
	Village    *geographic.Village  `gorm:"foreignKey:VillageID" json:"village,omitempty"`

	// Location coordinates
	Latitude  *float64 `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude *float64 `gorm:"type:decimal(11,8)" json:"longitude"`

	// Manager assigned to this outlet
	ManagerID *string   `gorm:"type:uuid;index" json:"manager_id"`
	Manager   *Employee `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`

	// Company that owns this outlet
	CompanyID *string  `gorm:"type:uuid;index" json:"company_id"`
	Company   *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`

	// Linked warehouse (created on outlet creation if toggle is on)
	WarehouseID *string `gorm:"type:uuid;index" json:"warehouse_id"`

	IsActive  bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Outlet
func (Outlet) TableName() string {
	return "outlets"
}

// BeforeCreate hook to generate UUID
func (o *Outlet) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}
