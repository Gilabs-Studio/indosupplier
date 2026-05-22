package models

import (
	"time"

	geographic "github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Warehouse represents a warehouse/storage location entity
type Warehouse struct {
	ID          string              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code        string              `gorm:"type:varchar(50);uniqueIndex" json:"code"`
	Name        string              `gorm:"type:varchar(200);not null;index" json:"name"`
	Description string              `gorm:"type:text" json:"description"`
	Address     string              `gorm:"type:text" json:"address"`
	ProvinceID  *string             `gorm:"type:uuid;index" json:"province_id"`
	Province    *geographic.Province `gorm:"foreignKey:ProvinceID" json:"province,omitempty"`
	CityID      *string             `gorm:"type:uuid;index" json:"city_id"`
	City        *geographic.City     `gorm:"foreignKey:CityID" json:"city,omitempty"`
	DistrictID  *string             `gorm:"type:uuid;index" json:"district_id"`
	District    *geographic.District `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	VillageID   *string             `gorm:"type:uuid;index" json:"village_id"`
	Village     *geographic.Village `gorm:"foreignKey:VillageID" json:"village,omitempty"`
	// VillageName stores the kelurahan/desa name as free text (preferred over VillageID FK)
	VillageName *string             `gorm:"type:varchar(255)" json:"village_name,omitempty"`
	// Location coordinates
	Latitude    *float64       `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude   *float64       `gorm:"type:decimal(11,8)" json:"longitude"`
	// POS outlet flag: true marks this warehouse as a POS outlet
	IsPosOutlet bool           `gorm:"column:is_pos_outlet;default:false;index" json:"is_pos_outlet"`
	// OutletID links this warehouse to an outlet (nullable)
	OutletID    *string        `gorm:"type:uuid;index" json:"outlet_id"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	HasStock  bool           `gorm:"->;column:has_stock" json:"-"` // Not in DB table, just for query
}

// TableName specifies the table name for Warehouse
func (Warehouse) TableName() string {
	return "warehouses"
}

// BeforeCreate hook to generate UUID
func (w *Warehouse) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}
