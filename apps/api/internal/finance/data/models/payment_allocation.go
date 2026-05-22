package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentAllocation struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	PaymentID string   `gorm:"type:uuid;not null;index" json:"payment_id"`
	Payment   *Payment `gorm:"foreignKey:PaymentID" json:"-"`

	ChartOfAccountID           string          `gorm:"type:uuid;not null;index" json:"chart_of_account_id"`
	ChartOfAccount             *ChartOfAccount `gorm:"foreignKey:ChartOfAccountID" json:"chart_of_account,omitempty"`
	ChartOfAccountCodeSnapshot string          `gorm:"type:varchar(50)" json:"chart_of_account_code_snapshot,omitempty"`
	ChartOfAccountNameSnapshot string          `gorm:"type:varchar(200)" json:"chart_of_account_name_snapshot,omitempty"`
	ChartOfAccountTypeSnapshot string          `gorm:"type:varchar(20)" json:"chart_of_account_type_snapshot,omitempty"`

	ReferenceType *string `gorm:"type:varchar(50);index" json:"reference_type"`
	ReferenceID   *string `gorm:"type:uuid;index" json:"reference_id"`

	Amount float64 `gorm:"type:numeric(18,2);not null" json:"amount"`
	Memo   string  `gorm:"type:text" json:"memo"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PaymentAllocation) TableName() string {
	return "payment_allocations"
}

func (pa *PaymentAllocation) BeforeCreate(tx *gorm.DB) error {
	if pa.ID == "" {
		pa.ID = uuid.New().String()
	}
	return nil
}
