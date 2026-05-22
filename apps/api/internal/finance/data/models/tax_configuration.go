package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaxType string

const (
	TaxTypeVAT            TaxType = "vat"
	TaxTypeIncomeTax      TaxType = "income_tax"
	TaxTypeWithholdingTax TaxType = "withholding_tax"
)

type TaxConfiguration struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	CompanyID string  `gorm:"type:uuid;not null;index:idx_tax_config_company_code,priority:1" json:"company_id"`
	TaxCode   string  `gorm:"type:varchar(50);not null;index:idx_tax_config_company_code,priority:2,unique" json:"tax_code"`
	TaxName   string  `gorm:"type:varchar(200);not null" json:"tax_name"`
	TaxType   TaxType `gorm:"type:varchar(30);not null;index" json:"tax_type"`
	Rate      float64 `gorm:"type:decimal(5,2);not null;default:0" json:"rate"`

	IsInclusive bool   `gorm:"default:false" json:"is_inclusive"`
	AccountID   string `gorm:"type:uuid;not null;index" json:"account_id"`
	IsActive    bool   `gorm:"default:true;index" json:"is_active"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Account ChartOfAccount `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

func (TaxConfiguration) TableName() string {
	return "tax_configurations"
}

func (t *TaxConfiguration) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
