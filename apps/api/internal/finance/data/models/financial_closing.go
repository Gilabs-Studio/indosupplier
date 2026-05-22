package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FinancialClosingStatus string

const (
	FinancialClosingStatusDraft    FinancialClosingStatus = "draft"
	FinancialClosingStatusApproved FinancialClosingStatus = "approved"
)

type FinancialClosing struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	PeriodEndDate time.Time              `gorm:"type:date;not null;index;uniqueIndex" json:"period_end_date"`
	Status        FinancialClosingStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	Notes         string                 `gorm:"type:text" json:"notes"`

	ApprovedAt *time.Time `json:"approved_at"`
	ApprovedBy *string    `gorm:"type:uuid" json:"approved_by"`

	CreatedBy *string `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FinancialClosing) TableName() string {
	return "financial_closings"
}

func (c *FinancialClosing) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
