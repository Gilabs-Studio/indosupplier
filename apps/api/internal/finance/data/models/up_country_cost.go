package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UpCountryCostStatus string

const (
	UpCountryCostStatusDraft           UpCountryCostStatus = "draft"
	UpCountryCostStatusSubmitted       UpCountryCostStatus = "submitted"
	UpCountryCostStatusManagerApproved UpCountryCostStatus = "manager_approved"
	UpCountryCostStatusFinanceApproved UpCountryCostStatus = "finance_approved"
	UpCountryCostStatusPaid            UpCountryCostStatus = "paid"
	UpCountryCostStatusRejected        UpCountryCostStatus = "rejected"
)

type UpCountryCost struct {
	ID        string              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  string              `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code      string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Purpose   string              `gorm:"type:varchar(255)" json:"purpose"`
	Location  string              `gorm:"type:varchar(255)" json:"location"`
	StartDate time.Time           `gorm:"type:date;not null" json:"start_date"`
	EndDate   time.Time           `gorm:"type:date;not null" json:"end_date"`
	Status    UpCountryCostStatus `gorm:"type:varchar(30);default:'draft';index" json:"status"`
	Notes     string              `gorm:"type:text" json:"notes"`

	Employees []UpCountryCostEmployee `gorm:"foreignKey:UpCountryCostID;constraint:OnDelete:CASCADE" json:"employees,omitempty"`
	Items     []UpCountryCostItem     `gorm:"foreignKey:UpCountryCostID;constraint:OnDelete:CASCADE" json:"items,omitempty"`

	// Submission tracking
	SubmittedAt *time.Time `json:"submitted_at"`
	SubmittedBy *string    `gorm:"type:uuid" json:"submitted_by"`

	// Manager approval tracking
	ManagerApprovedAt *time.Time `json:"manager_approved_at"`
	ManagerApprovedBy *string    `gorm:"type:uuid" json:"manager_approved_by"`
	ManagerComment    string     `gorm:"type:text" json:"manager_comment"`

	// Finance approval tracking
	FinanceApprovedAt *time.Time `json:"finance_approved_at"`
	FinanceApprovedBy *string    `gorm:"type:uuid" json:"finance_approved_by"`

	// Payment tracking
	PaidAt *time.Time `json:"paid_at"`
	PaidBy *string    `gorm:"type:uuid" json:"paid_by"`

	// Legacy fields (kept for backward compatibility)
	ApprovedAt *time.Time `json:"approved_at"`
	ApprovedBy *string    `gorm:"type:uuid" json:"approved_by"`

	CreatedBy *string        `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UpCountryCost) TableName() string {
	return "up_country_costs"
}

func (u *UpCountryCost) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

type UpCountryCostEmployee struct {
	ID              string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UpCountryCostID string `gorm:"type:uuid;not null;index" json:"up_country_cost_id"`
	EmployeeID      string `gorm:"type:uuid;not null;index" json:"employee_id"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UpCountryCostEmployee) TableName() string {
	return "up_country_cost_employees"
}

func (u *UpCountryCostEmployee) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

type CostType string

const (
	CostTypeTransport     CostType = "transport"
	CostTypeAccommodation CostType = "accommodation"
	CostTypeMeal          CostType = "meal"
	CostTypeFuel          CostType = "fuel"
	CostTypeOther         CostType = "other"
)

type UpCountryCostItem struct {
	ID              string     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UpCountryCostID string     `gorm:"type:uuid;not null;index" json:"up_country_cost_id"`
	CostType        CostType   `gorm:"type:varchar(50);not null" json:"cost_type"`
	Description     string     `gorm:"type:text" json:"description"`
	Amount          float64    `gorm:"type:numeric(18,2);default:0" json:"amount"`
	ExpenseDate     *time.Time `gorm:"type:date" json:"expense_date"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UpCountryCostItem) TableName() string {
	return "up_country_cost_items"
}

func (u *UpCountryCostItem) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
