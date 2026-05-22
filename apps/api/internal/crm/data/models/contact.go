package models

import (
	"time"

	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Contact represents a person associated with a Customer (child of Customer)
type Contact struct {
	ID            string                   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	CustomerID    string                   `gorm:"type:uuid;not null;index" json:"customer_id"`
	Customer      *customerModels.Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	ContactRoleID *string                  `gorm:"type:uuid;index" json:"contact_role_id"`
	ContactRole   *ContactRole             `gorm:"foreignKey:ContactRoleID" json:"contact_role,omitempty"`
	Name          string                   `gorm:"type:varchar(200);not null;index" json:"name"`
	Phone         string                   `gorm:"type:varchar(30)" json:"phone"`
	Email         string                   `gorm:"type:varchar(100)" json:"email"`
	Position      string                   `gorm:"type:varchar(100)" json:"position"`
	Notes         string                   `gorm:"type:text" json:"notes"`
	IsActive      bool                     `gorm:"default:true;index" json:"is_active"`
	CreatedBy     *string                  `gorm:"type:uuid" json:"created_by"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `gorm:"index" json:"updated_at"`
	DeletedAt     gorm.DeletedAt           `gorm:"index" json:"-"`
}

// TableName specifies the table name for Contact
func (Contact) TableName() string {
	return "crm_contacts"
}

// BeforeCreate hook to generate UUID
func (c *Contact) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
