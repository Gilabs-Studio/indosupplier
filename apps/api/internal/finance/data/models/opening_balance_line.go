package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OpeningBalanceLine struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	CompanyID    string `gorm:"type:uuid;not null;index:idx_ob_line_company_fy,priority:1" json:"company_id"`
	FiscalYearID string `gorm:"type:uuid;not null;index:idx_ob_line_company_fy,priority:2" json:"fiscal_year_id"`
	AccountID    string `gorm:"type:uuid;not null;index" json:"account_id"`

	DebitAmount  float64 `gorm:"type:decimal(18,2);not null;default:0" json:"debit_amount"`
	CreditAmount float64 `gorm:"type:decimal(18,2);not null;default:0" json:"credit_amount"`
	Description  string  `gorm:"type:text" json:"description,omitempty"`

	ProductID      *string  `gorm:"type:uuid;index" json:"product_id,omitempty"`
	ProductQty     *float64 `gorm:"type:decimal(18,4)" json:"product_qty,omitempty"`
	ProductAvgCost *float64 `gorm:"type:decimal(18,6)" json:"product_avg_cost,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OpeningBalanceLine) TableName() string {
	return "opening_balance_lines"
}

func (o *OpeningBalanceLine) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}
