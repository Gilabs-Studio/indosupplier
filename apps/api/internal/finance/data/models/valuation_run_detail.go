package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ValuationDirection marks whether a valuation item is gain or loss.
type ValuationDirection string

const (
	ValuationDirectionGain ValuationDirection = "gain"
	ValuationDirectionLoss ValuationDirection = "loss"
)

// ValuationRunDetail stores granular valuation item rows for auditability.
type ValuationRunDetail struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	ValuationRunID string        `gorm:"type:uuid;not null;index" json:"valuation_run_id"`
	ValuationRun   *ValuationRun `gorm:"foreignKey:ValuationRunID;constraint:OnDelete:CASCADE" json:"valuation_run,omitempty"`

	ReferenceID string  `gorm:"type:varchar(255);not null;index" json:"reference_id"`
	ProductID   *string `gorm:"type:uuid;index" json:"product_id,omitempty"`

	Qty         float64            `gorm:"type:numeric(18,6);not null;default:0" json:"qty"`
	BookValue   float64            `gorm:"type:numeric(18,2);not null;default:0" json:"book_value"`
	ActualValue float64            `gorm:"type:numeric(18,2);not null;default:0" json:"actual_value"`
	Delta       float64            `gorm:"type:numeric(18,2);not null;default:0" json:"delta"`
	Direction   ValuationDirection `gorm:"type:varchar(10);not null;index" json:"direction"`

	// Snapshots (immutable for audit trail)
	CostPriceSnapshot    *float64 `gorm:"type:numeric(18,6)" json:"cost_price_snapshot,omitempty"`
	CurrencyCodeSnapshot *string  `gorm:"type:varchar(3)" json:"currency_code_snapshot,omitempty"`
	ExchangeRateSnapshot *float64 `gorm:"type:numeric(15,8)" json:"exchange_rate_snapshot,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

func (ValuationRunDetail) TableName() string {
	return "valuation_run_details"
}

func (v *ValuationRunDetail) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	return nil
}
