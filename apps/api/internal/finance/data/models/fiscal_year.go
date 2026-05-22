package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FiscalYearStatus string

const (
	FiscalYearStatusDraft  FiscalYearStatus = "draft"
	FiscalYearStatusActive FiscalYearStatus = "active"
	FiscalYearStatusLocked FiscalYearStatus = "locked"
)

type FiscalYear struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	CompanyID string           `gorm:"type:uuid;not null;index:idx_fiscal_year_company" json:"company_id"`
	Name      string           `gorm:"type:varchar(100);not null" json:"name"`
	StartDate time.Time        `gorm:"type:date;not null;index" json:"start_date"`
	EndDate   time.Time        `gorm:"type:date;not null;index" json:"end_date"`
	Status    FiscalYearStatus `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"`

	CreatedBy *string        `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FiscalYear) TableName() string {
	return "fiscal_years"
}

func (fy *FiscalYear) BeforeCreate(tx *gorm.DB) error {
	if fy.ID == "" {
		fy.ID = uuid.New().String()
	}
	return nil
}
