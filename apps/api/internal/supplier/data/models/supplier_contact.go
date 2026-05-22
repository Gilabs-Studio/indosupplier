package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactRole shadow struct to avoid import cycle with CRM module
type ContactRole struct {
	ID         string `gorm:"primaryKey"`
	Name       string
	Code       string
	BadgeColor string
}

func (ContactRole) TableName() string {
	return "crm_contact_roles"
}

// SupplierContact represents a contact person for a supplier
type SupplierContact struct {
	ID            string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	SupplierID    string         `gorm:"type:uuid;not null;index" json:"supplier_id"`
	ContactRoleID *string        `gorm:"type:uuid;index" json:"contact_role_id"`
	ContactRole   *ContactRole   `gorm:"foreignKey:ContactRoleID" json:"contact_role,omitempty"`
	Name          string         `gorm:"type:varchar(200);not null;index" json:"name"`
	Phone         string         `gorm:"type:varchar(30)" json:"phone"`
	Email         string         `gorm:"type:varchar(100)" json:"email"`
	Position      string         `gorm:"type:varchar(100)" json:"position"`
	Notes         string         `gorm:"type:text" json:"notes"`
	IsPrimary     bool           `gorm:"default:false" json:"is_primary"`
	IsActive      bool           `gorm:"default:true;index" json:"is_active"`
	CreatedBy     *string        `gorm:"type:uuid" json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for SupplierContact
func (SupplierContact) TableName() string {
	return "supplier_contacts"
}

// BeforeCreate hook to generate UUID
func (s *SupplierContact) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
