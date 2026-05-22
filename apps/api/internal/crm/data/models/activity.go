package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Activity represents an immutable log of CRM interactions
type Activity struct {
	ID             string  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Type           string  `gorm:"type:varchar(30);not null;index" json:"type"` // visit, call, email, task, deal, lead
	ActivityTypeID *string `gorm:"type:uuid;index" json:"activity_type_id"`
	ActivityType   *ActivityType `gorm:"foreignKey:ActivityTypeID" json:"activity_type,omitempty"`

	// Polymorphic references (all optional)
	CustomerID    *string `gorm:"type:uuid;index" json:"customer_id"`
	ContactID     *string `gorm:"type:uuid;index" json:"contact_id"`
	DealID        *string `gorm:"type:uuid;index" json:"deal_id"`
	LeadID        *string `gorm:"type:uuid;index" json:"lead_id"`
	VisitReportID *string `gorm:"type:uuid;index" json:"visit_report_id"`

	// Who performed the activity
	EmployeeID       string             `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee         *orgModels.Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`

	// Content
	Description string    `gorm:"type:text;not null" json:"description"`
	Timestamp   time.Time `gorm:"not null;index" json:"timestamp"`
	Metadata    *string   `gorm:"type:jsonb" json:"metadata"`

	// Timestamps (no soft delete — activities are immutable)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
}

// TableName returns the database table name
func (Activity) TableName() string {
	return "crm_activities"
}

// BeforeCreate generates a UUID if not set
func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Timestamp.IsZero() {
		a.Timestamp = apptime.Now()
	}
	return nil
}
