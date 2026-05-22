package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetDepreciationStatus string

const (
	AssetDepreciationStatusPending  AssetDepreciationStatus = "pending"
	AssetDepreciationStatusApproved AssetDepreciationStatus = "approved"
)

type AssetDepreciation struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	AssetID string `gorm:"type:uuid;not null;index;uniqueIndex:idx_asset_period" json:"asset_id"`
	Asset   *Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`

	Period           string             `gorm:"type:varchar(7);not null;index;uniqueIndex:idx_asset_period" json:"period"` // YYYY-MM
	DepreciationDate time.Time          `gorm:"type:date;not null;index" json:"depreciation_date"`
	Method           DepreciationMethod `gorm:"type:varchar(10);not null" json:"method"`

	Amount      float64 `gorm:"type:numeric(18,2);not null" json:"amount"`
	Accumulated float64 `gorm:"type:numeric(18,2);not null" json:"accumulated"`
	BookValue   float64 `gorm:"type:numeric(18,2);not null" json:"book_value"`

	Status AssetDepreciationStatus `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`

	JournalEntryID *string `gorm:"type:uuid;index" json:"journal_entry_id"`

	CreatedBy *string   `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

func (AssetDepreciation) TableName() string {
	return "asset_depreciations"
}

func (d *AssetDepreciation) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}
