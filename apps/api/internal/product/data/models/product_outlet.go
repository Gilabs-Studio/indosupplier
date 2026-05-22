package models

import (
	"time"

	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductOutlet is the junction table linking a product to specific outlets where it is available.
type ProductOutlet struct {
	ID        string             `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  string             `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	ProductID string             `gorm:"type:uuid;not null;index" json:"product_id"`
	OutletID  string             `gorm:"type:uuid;not null;index" json:"outlet_id"`
	Outlet    *orgModels.Outlet  `gorm:"foreignKey:OutletID" json:"outlet,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	DeletedAt gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (ProductOutlet) TableName() string { return "product_outlets" }

func (p *ProductOutlet) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
