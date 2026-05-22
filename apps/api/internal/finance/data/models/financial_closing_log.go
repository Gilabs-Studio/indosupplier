package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FinancialClosingLogAction string

const (
	FinancialClosingLogActionValidate FinancialClosingLogAction = "validate"
	FinancialClosingLogActionClose    FinancialClosingLogAction = "close"
	FinancialClosingLogActionReopen   FinancialClosingLogAction = "reopen"
)

type FinancialClosingLog struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	PeriodEndDate time.Time                 `gorm:"type:date;not null;index" json:"period_end_date"`
	Action        FinancialClosingLogAction `gorm:"type:varchar(20);not null;index" json:"action"`
	Reason        string                    `gorm:"type:text" json:"reason"`
	CreatedBy     *string                   `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FinancialClosingLog) TableName() string {
	return "financial_closing_logs"
}

func (l *FinancialClosingLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}
