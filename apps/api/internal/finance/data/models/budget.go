package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BudgetStatus string

const (
	BudgetStatusDraft    BudgetStatus = "draft"
	BudgetStatusApproved BudgetStatus = "approved"
)

type Budget struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	Name        string `gorm:"type:varchar(255);not null;index" json:"name"`
	Description string `gorm:"type:text" json:"description"`

	StartDate time.Time `gorm:"type:date;not null;index" json:"start_date"`
	EndDate   time.Time `gorm:"type:date;not null;index" json:"end_date"`

	TotalAmount float64      `gorm:"type:numeric(18,2);not null" json:"total_amount"`
	Status      BudgetStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`

	ApprovedAt *time.Time `json:"approved_at"`
	ApprovedBy *string    `gorm:"type:uuid" json:"approved_by"`

	CreatedBy *string `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Items []BudgetItem `gorm:"foreignKey:BudgetID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

func (Budget) TableName() string {
	return "budgets"
}

func (b *Budget) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
