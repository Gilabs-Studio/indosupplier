package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CompanyStatus represents the approval status of a company
type CompanyStatus string

const (
	CompanyStatusDraft    CompanyStatus = "draft"
	CompanyStatusPending  CompanyStatus = "pending"
	CompanyStatusApproved CompanyStatus = "approved"
	CompanyStatusRejected CompanyStatus = "rejected"
)

// Company represents a company entity with approval workflow
type Company struct {
	ID         string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name       string         `gorm:"type:varchar(200);not null;index" json:"name"`
	Address    string         `gorm:"type:text" json:"address"`
	Email      string         `gorm:"type:varchar(100)" json:"email"`
	Phone      string         `gorm:"type:varchar(20)" json:"phone"`
	NPWP       string         `gorm:"type:varchar(30)" json:"npwp"`
	ProvinceID *string        `gorm:"type:uuid;index" json:"province_id"`
	Province   *models.Province `gorm:"foreignKey:ProvinceID" json:"province,omitempty"`
	CityID     *string        `gorm:"type:uuid;index" json:"city_id"`
	City       *models.City     `gorm:"foreignKey:CityID" json:"city,omitempty"`
	DistrictID *string        `gorm:"type:uuid;index" json:"district_id"`
	District   *models.District `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	NIB        string         `gorm:"type:varchar(30)" json:"nib"`
	VillageID   *string         `gorm:"type:uuid;index" json:"village_id"`
	VillageName *string         `gorm:"type:varchar(255)" json:"village_name,omitempty"`
	Latitude    *float64        `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude   *float64        `gorm:"type:decimal(11,8)" json:"longitude"`
	Village     *models.Village `gorm:"foreignKey:VillageID" json:"village,omitempty"`
	// IANA timezone for company-local time calculations (e.g. "Asia/Jakarta", "Asia/Makassar")
	Timezone   string         `gorm:"type:varchar(50);default:'Asia/Jakarta';not null" json:"timezone"`
	// Approval workflow
	Status     CompanyStatus  `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	IsApproved bool           `gorm:"default:false;index" json:"is_approved"`
	CreatedBy  *string        `gorm:"type:uuid" json:"created_by"`
	ApprovedBy *string        `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt *time.Time     `json:"approved_at"`
	IsActive   bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	// OutletCount is populated via subquery in list queries — not stored in DB
	OutletCount int64 `gorm:"column:outlet_count;->" json:"-"`
}

// TableName specifies the table name for Company
func (Company) TableName() string {
	return "companies"
}

// BeforeCreate hook to generate UUID
func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
