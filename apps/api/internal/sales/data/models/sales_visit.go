package models

import (
	"time"

	geoModels "github.com/gilabs/gims/api/internal/geographic/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SalesVisitStatus represents the status of a sales visit
type SalesVisitStatus string

const (
	SalesVisitStatusPlanned    SalesVisitStatus = "planned"
	SalesVisitStatusInProgress SalesVisitStatus = "in_progress"
	SalesVisitStatusCompleted  SalesVisitStatus = "completed"
	SalesVisitStatusCancelled  SalesVisitStatus = "cancelled"
)

// SalesVisit represents a sales visit to a customer
type SalesVisit struct {
	ID   string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`

	// Schedule
	VisitDate     time.Time  `gorm:"type:date;not null;index" json:"visit_date"`
	ScheduledTime *time.Time `gorm:"type:time" json:"scheduled_time"`
	ActualTime    *time.Time `gorm:"type:time" json:"actual_time"`

	// Sales Representative
	EmployeeID string                `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee   *orgModels.Employee   `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`

	// Customer
	CompanyID *string              `gorm:"type:uuid;index" json:"company_id"`
	Company   *orgModels.Company   `gorm:"foreignKey:CompanyID" json:"company,omitempty"`

	// Contact
	ContactPerson string `gorm:"type:varchar(200)" json:"contact_person"`
	ContactPhone  string `gorm:"type:varchar(20)" json:"contact_phone"`

	// Location
	Address   string             `gorm:"type:text" json:"address"`
	VillageID *string            `gorm:"type:uuid;index" json:"village_id"`
	Village   *geoModels.Village `gorm:"foreignKey:VillageID" json:"village,omitempty"`
	Latitude  *float64           `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude *float64           `gorm:"type:decimal(11,8)" json:"longitude"`

	// Visit Content
	Purpose string `gorm:"type:text" json:"purpose"`
	Notes   string `gorm:"type:text" json:"notes"`
	Result  string `gorm:"type:text" json:"result"`

	// Status
	Status SalesVisitStatus `gorm:"type:varchar(20);default:'planned';index" json:"status"`

	// Check-in timestamps
	CheckInAt  *time.Time `json:"check_in_at"`
	CheckOutAt *time.Time `json:"check_out_at"`

	// Audit fields
	CreatedBy   *string    `gorm:"type:uuid" json:"created_by"`
	CancelledBy *string    `gorm:"type:uuid" json:"cancelled_by"`
	CancelledAt *time.Time `json:"cancelled_at"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Details         []SalesVisitDetail          `gorm:"foreignKey:SalesVisitID;constraint:OnDelete:CASCADE" json:"details,omitempty"`
	ProgressHistory []SalesVisitProgressHistory `gorm:"foreignKey:SalesVisitID;constraint:OnDelete:CASCADE" json:"progress_history,omitempty"`
}

// TableName specifies the table name for SalesVisit
func (SalesVisit) TableName() string {
	return "sales_visits"
}

// BeforeCreate hook to generate UUID
func (sv *SalesVisit) BeforeCreate(tx *gorm.DB) error {
	if sv.ID == "" {
		sv.ID = uuid.New().String()
	}
	return nil
}

// SalesVisitDetail represents a product discussed during the visit
type SalesVisitDetail struct {
	ID           string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesVisitID string `gorm:"type:uuid;not null;index" json:"sales_visit_id"`
	SalesVisit   *SalesVisit `gorm:"foreignKey:SalesVisitID" json:"sales_visit,omitempty"`

	ProductID string                 `gorm:"type:uuid;not null;index" json:"product_id"`
	Product   *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	// Interest level (1-5)
	InterestLevel int    `gorm:"default:0" json:"interest_level"`
	Notes         string `gorm:"type:text" json:"notes"`

	// Quantity discussed (optional)
	Quantity *float64 `gorm:"type:decimal(15,3)" json:"quantity"`
	Price    *float64 `gorm:"type:decimal(15,2)" json:"price"`

	// Relations
	Answers []SalesVisitInterestAnswer `gorm:"foreignKey:SalesVisitDetailID;constraint:OnDelete:CASCADE" json:"answers,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for SalesVisitDetail
func (SalesVisitDetail) TableName() string {
	return "sales_visit_details"
}

// BeforeCreate hook to generate UUID
func (svd *SalesVisitDetail) BeforeCreate(tx *gorm.DB) error {
	if svd.ID == "" {
		svd.ID = uuid.New().String()
	}
	return nil
}

// SalesVisitProgressHistory tracks status/notes changes
type SalesVisitProgressHistory struct {
	ID           string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SalesVisitID string `gorm:"type:uuid;not null;index" json:"sales_visit_id"`
	SalesVisit   *SalesVisit `gorm:"foreignKey:SalesVisitID" json:"sales_visit,omitempty"`

	// Change tracking
	FromStatus SalesVisitStatus `gorm:"type:varchar(20)" json:"from_status"`
	ToStatus   SalesVisitStatus `gorm:"type:varchar(20)" json:"to_status"`
	Notes      string           `gorm:"type:text" json:"notes"`

	// Actor
	ChangedBy *string `gorm:"type:uuid" json:"changed_by"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for SalesVisitProgressHistory
func (SalesVisitProgressHistory) TableName() string {
	return "sales_visit_progress_history"
}

// BeforeCreate hook to generate UUID
func (svph *SalesVisitProgressHistory) BeforeCreate(tx *gorm.DB) error {
	if svph.ID == "" {
		svph.ID = uuid.New().String()
	}
	return nil
}
