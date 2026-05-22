package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FinancialClosingSnapshot struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	PeriodEndDate           time.Time `gorm:"type:date;not null;index" json:"period_end_date"`
	NetProfit               float64   `gorm:"not null" json:"net_profit"`
	RetainedEarningsBalance float64   `gorm:"not null" json:"retained_earnings_balance"`
	SnapshotJSON            string    `gorm:"type:jsonb" json:"snapshot_json"`
	Notes                   string    `gorm:"type:text" json:"notes"`

	CreatedBy *string `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FinancialClosingSnapshot) TableName() string {
	return "financial_closing_snapshots"
}

func (s *FinancialClosingSnapshot) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
