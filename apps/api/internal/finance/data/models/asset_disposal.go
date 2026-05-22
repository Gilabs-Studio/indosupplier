package models

import (
	"time"

	"github.com/google/uuid"
)

// AssetDisposal model
type AssetDisposal struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	TenantID            *uuid.UUID `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	AssetID             uuid.UUID  `gorm:"type:uuid;not null;index" json:"asset_id"`
	SourceTransactionID *uuid.UUID `gorm:"type:uuid;index" json:"source_transaction_id,omitempty"`

	DisposalType DisposalType `gorm:"type:varchar(50)" json:"disposal_type"`
	DisposalDate time.Time    `gorm:"type:date;not null;index" json:"disposal_date"`

	// Financial Details
	DisposalValue                     *float64 `gorm:"type:numeric(18,2)" json:"disposal_value"`
	AccumulatedDepreciationAtDisposal *float64 `gorm:"type:numeric(18,2)" json:"accumulated_depreciation_at_disposal"`
	GainOrLoss                        *float64 `gorm:"type:numeric(18,2)" json:"gain_or_loss"`

	// Transaction Details
	ReferenceType   *string `gorm:"type:varchar(50)" json:"reference_type"`
	ReferenceNumber *string `gorm:"type:varchar(100)" json:"reference_number"`
	Description     *string `gorm:"type:text" json:"description"`

	// Audit
	ApprovedBy *uuid.UUID `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt *time.Time `gorm:"type:timestamp" json:"approved_at"`
	CreatedBy  *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relations
	Asset *Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

// TableName specifies the table name
func (AssetDisposal) TableName() string {
	return "asset_disposals"
}
