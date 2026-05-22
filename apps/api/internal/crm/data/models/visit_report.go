package models

import (
	"time"

	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	geoModels "github.com/gilabs/gims/api/internal/geographic/data/models"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VisitReportStatus represents the approval workflow status.
// Internal-only compatibility: API response may omit this field.
type VisitReportStatus string

const (
	VisitReportStatusDraft     VisitReportStatus = "draft"
	VisitReportStatusSubmitted VisitReportStatus = "submitted"
	VisitReportStatusApproved  VisitReportStatus = "approved"
	VisitReportStatusRejected  VisitReportStatus = "rejected"
)



// VisitReportOutcome categorizes the visit result
type VisitReportOutcome string

const (
	VisitReportOutcomePositive     VisitReportOutcome = "positive"
	VisitReportOutcomeNeutral      VisitReportOutcome = "neutral"
	VisitReportOutcomeNegative     VisitReportOutcome = "negative"
	VisitReportOutcomeVeryPositive VisitReportOutcome = "very_positive"
)

// VisitReport — Merged entity from ERP SalesVisit + CRM VisitReport
// Combines: interest survey (ERP) + approval workflow + GPS + photos (CRM)
type VisitReport struct {
	ID   string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index;uniqueIndex:uq_crm_visit_reports_tenant_code" json:"tenant_id,omitempty"`
	Code string `gorm:"type:varchar(50);not null;uniqueIndex:uq_crm_visit_reports_tenant_code" json:"code"`

	// Relations
	CustomerID   *string                  `gorm:"type:uuid;index" json:"customer_id"`
	Customer     *customerModels.Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	ContactID    *string                  `gorm:"type:uuid;index" json:"contact_id"`
	Contact      *Contact                 `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	DealID       *string                  `gorm:"type:uuid;index" json:"deal_id"`
	Deal         *Deal                    `gorm:"foreignKey:DealID" json:"deal,omitempty"`
	LeadID       *string                  `gorm:"type:uuid;index" json:"lead_id"`
	Lead         *Lead                    `gorm:"foreignKey:LeadID" json:"lead,omitempty"`
	TravelPlanID *string                  `gorm:"type:uuid;index" json:"travel_plan_id"`
	EmployeeID   string                   `gorm:"type:uuid;not null;index" json:"employee_id"`
	Employee     *orgModels.Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`

	// Visit timing
	VisitDate     time.Time  `gorm:"type:date;not null;index" json:"visit_date"`
	ScheduledTime *time.Time `gorm:"type:time" json:"scheduled_time"`
	ActualTime    *time.Time `gorm:"type:time" json:"actual_time"`

	// GPS Check-in/Check-out (from CRM)
	CheckInAt        *time.Time `json:"check_in_at"`
	CheckOutAt       *time.Time `json:"check_out_at"`
	CheckInLocation  *string    `gorm:"type:jsonb" json:"check_in_location"`
	CheckOutLocation *string    `gorm:"type:jsonb" json:"check_out_location"`

	// Location
	Address   string             `gorm:"type:text" json:"address"`
	VillageID *string            `gorm:"type:uuid;index" json:"village_id"`
	Village   *geoModels.Village `gorm:"foreignKey:VillageID" json:"village,omitempty"`
	Latitude  *float64           `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude *float64           `gorm:"type:decimal(11,8)" json:"longitude"`

	// Content
	Purpose       string `gorm:"type:text" json:"purpose"`
	Notes         string `gorm:"type:text" json:"notes"`
	Result        string `gorm:"type:text" json:"result"`
	Outcome       string `gorm:"type:varchar(20)" json:"outcome"`
	NextSteps     string `gorm:"type:text" json:"next_steps"`
	ContactPerson string `gorm:"type:varchar(200)" json:"contact_person"`
	ContactPhone  string `gorm:"type:varchar(20)" json:"contact_phone"`

	// Photos (from CRM) — JSONB array of photo URLs
	Photos *string `gorm:"type:jsonb" json:"photos"`

	// Approval workflow (kept for backward compatibility in existing domain logic)
	Status          VisitReportStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	ApprovedBy      *string           `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt      *time.Time        `json:"approved_at"`
	RejectedBy      *string           `gorm:"type:uuid" json:"rejected_by"`
	RejectedAt      *time.Time        `json:"rejected_at"`
	RejectionReason string            `gorm:"type:text" json:"rejection_reason"`

	// Metadata
	CreatedBy   *string    `gorm:"type:uuid" json:"created_by"`
	CancelledBy *string    `gorm:"type:uuid" json:"cancelled_by"`
	CancelledAt *time.Time `json:"cancelled_at"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	Details         []VisitReportDetail          `gorm:"foreignKey:VisitReportID;constraint:OnDelete:CASCADE" json:"details,omitempty"`
	ProgressHistory []VisitReportProgressHistory `gorm:"foreignKey:VisitReportID;constraint:OnDelete:CASCADE" json:"progress_history,omitempty"`
}

func (VisitReport) TableName() string {
	return "crm_visit_reports"
}

func (v *VisitReport) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	return nil
}

// VisitReportDetail — Product interest tracking per visit (from ERP SalesVisitDetail)
type VisitReportDetail struct {
	ID            string                 `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID      string                 `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	VisitReportID string                 `gorm:"type:uuid;not null;index" json:"visit_report_id"`
	VisitReport   *VisitReport           `gorm:"foreignKey:VisitReportID" json:"visit_report,omitempty"`
	ProductID     string                 `gorm:"type:uuid;not null;index" json:"product_id"`
	Product       *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	// Interest level (1-5 scale)
	InterestLevel int    `gorm:"default:0" json:"interest_level"`
	Notes         string `gorm:"type:text" json:"notes"`

	// Quantity and price discussed (optional)
	Quantity *float64 `gorm:"type:decimal(15,3)" json:"quantity"`
	Price    *float64 `gorm:"type:decimal(15,2)" json:"price"`

	// Survey answers
	Answers []VisitReportInterestAnswer `gorm:"foreignKey:VisitReportDetailID;constraint:OnDelete:CASCADE" json:"answers,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (VisitReportDetail) TableName() string {
	return "crm_visit_report_details"
}

func (d *VisitReportDetail) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// VisitReportProgressHistory tracks status changes.
type VisitReportProgressHistory struct {
	ID            string            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID      string            `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	VisitReportID string            `gorm:"type:uuid;not null;index" json:"visit_report_id"`
	VisitReport   *VisitReport      `gorm:"foreignKey:VisitReportID" json:"visit_report,omitempty"`
	FromStatus    VisitReportStatus `gorm:"type:varchar(20)" json:"from_status"`
	ToStatus      VisitReportStatus `gorm:"type:varchar(20);not null" json:"to_status"`
	Notes         string            `gorm:"type:text" json:"notes"`
	ChangedBy     *string           `gorm:"type:uuid" json:"changed_by"`
	CreatedAt     time.Time         `json:"created_at"`
}

func (VisitReportProgressHistory) TableName() string {
	return "crm_visit_report_progress_history"
}

func (h *VisitReportProgressHistory) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = uuid.New().String()
	}
	return nil
}

// VisitReportInterestAnswer — links to survey questions/options per visit detail
// Reuses existing SalesVisitInterestQuestion and SalesVisitInterestOption models
type VisitReportInterestAnswer struct {
	ID                  string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID            string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	VisitReportDetailID string `gorm:"type:uuid;not null;index" json:"visit_report_detail_id"`
	QuestionID          string `gorm:"type:uuid;not null;index" json:"question_id"`
	OptionID            string `gorm:"type:uuid;not null;index" json:"option_id"`

	// Denormalized for easier querying
	Score int `gorm:"default:0" json:"score"`

	// Relations — reuse ERP interest survey models
	Question *salesModels.SalesVisitInterestQuestion `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	Option   *salesModels.SalesVisitInterestOption   `gorm:"foreignKey:OptionID" json:"option,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
}

func (VisitReportInterestAnswer) TableName() string {
	return "crm_visit_report_interest_answers"
}

func (a *VisitReportInterestAnswer) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}
