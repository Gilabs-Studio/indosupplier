package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountingPeriodStatus string

const (
	AccountingPeriodStatusOpen   AccountingPeriodStatus = "open"
	AccountingPeriodStatusClosed AccountingPeriodStatus = "closed"
)

type AccountingPeriod struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	PeriodName string                 `gorm:"type:varchar(50);not null;uniqueIndex" json:"period_name"`
	StartDate  time.Time              `gorm:"type:date;not null;index" json:"start_date"`
	EndDate    time.Time              `gorm:"type:date;not null;index" json:"end_date"`
	Status     AccountingPeriodStatus `gorm:"type:varchar(20);not null;default:'open';index" json:"status"`
	LockedAt   *time.Time             `json:"locked_at"`
	LockedBy   *string                `gorm:"type:uuid" json:"locked_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AccountingPeriod) TableName() string {
	return "accounting_periods"
}

func (p *AccountingPeriod) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
