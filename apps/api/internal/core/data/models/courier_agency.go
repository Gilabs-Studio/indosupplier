package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CourierAgency represents a courier/shipping agency
type CourierAgency struct {
	ID          string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    string         `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code        string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"code"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Phone       string         `gorm:"type:varchar(20)" json:"phone"`
	Address     string         `gorm:"type:text" json:"address"`
	TrackingURL string         `gorm:"type:varchar(255)" json:"tracking_url"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for CourierAgency
func (CourierAgency) TableName() string {
	return "courier_agencies"
}

// BeforeCreate hook to generate UUID
func (c *CourierAgency) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
