package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetRevaluation represents the final record of asset revaluation (IMMUTABLE after approval)
type AssetRevaluation struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_revaluation_tenant_asset" json:"tenant_id"`
	AssetID      uuid.UUID  `gorm:"type:uuid;not null;index:idx_revaluation_tenant_asset" json:"asset_id"`
	Asset        *Asset     `gorm:"foreignKey:AssetID;constraint:OnDelete:CASCADE" json:"asset,omitempty"`

	// Revaluation Details
	RevaluationDate   time.Time `gorm:"type:date;not null" json:"revaluation_date"`
	OldGrossValue     float64   `gorm:"type:numeric(18,2);not null;default:0" json:"old_gross_value"`
	NewGrossValue     float64   `gorm:"type:numeric(18,2);not null;default:0" json:"new_gross_value"`
	Adjustment        float64   `gorm:"type:numeric(18,2);not null;default:0" json:"adjustment"`
	RevaluationReason string    `gorm:"type:text" json:"revaluation_reason"`

	// Accounting Information
	SourceTransactionID *uuid.UUID `gorm:"type:uuid;index" json:"source_transaction_id,omitempty"`
	JournalEntryID      *uuid.UUID `gorm:"type:uuid;index" json:"journal_entry_id,omitempty"`

	// Workflow
	ApprovedBy *uuid.UUID `gorm:"type:uuid" json:"approved_by,omitempty"`
	ApprovedAt *time.Time `gorm:"type:timestamptz" json:"approved_at,omitempty"`

	// Audit
	CreatedBy uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
}

func (AssetRevaluation) TableName() string {
	return "asset_revaluations"
}

func (r *AssetRevaluation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.TenantID == uuid.Nil {
		return gorm.ErrInvalidData
	}
	return nil
}
