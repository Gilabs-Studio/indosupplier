package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetDepreciationSchedule model
type AssetDepreciationSchedule struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string    `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	AssetID  uuid.UUID `gorm:"type:uuid;not null;index" json:"asset_id"`

	PeriodStartDate time.Time `gorm:"type:date;not null" json:"period_start_date"`
	PeriodEndDate   time.Time `gorm:"type:date;not null" json:"period_end_date"`
	PeriodMonth     int       `gorm:"type:integer;not null" json:"period_month"`

	DepreciationAmount      float64 `gorm:"type:numeric(18,2);not null" json:"depreciation_amount"`
	AccumulatedDepreciation float64 `gorm:"type:numeric(18,2);not null" json:"accumulated_depreciation"`
	BookValue               float64 `gorm:"type:numeric(18,2);not null" json:"book_value"`

	// Accounting Integration
	JournalEntryID *uuid.UUID `gorm:"type:uuid" json:"journal_entry_id"`
	IsPosted       bool       `gorm:"type:boolean;default:false;index" json:"is_posted"`
	PostedAt       *time.Time `gorm:"type:timestamp" json:"posted_at"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Asset *Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

// TableName specifies the table name
func (AssetDepreciationSchedule) TableName() string {
	return "asset_depreciation_schedules"
}

func (s *AssetDepreciationSchedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
