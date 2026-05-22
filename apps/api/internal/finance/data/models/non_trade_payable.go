package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NonTradePayableStatus string

const (
	NTPStatusDraft     NonTradePayableStatus = "draft"
	NTPStatusPosted    NonTradePayableStatus = "posted"
	NTPStatusPartial   NonTradePayableStatus = "partial"
	NTPStatusSubmitted NonTradePayableStatus = "submitted"
	NTPStatusApproved  NonTradePayableStatus = "approved"
	NTPStatusRejected  NonTradePayableStatus = "rejected"
	NTPStatusPaid      NonTradePayableStatus = "paid"
	NTPStatusCancelled NonTradePayableStatus = "cancelled"
)

type NonTradePayable struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	TransactionDate time.Time `gorm:"type:date;not null;index" json:"transaction_date"`
	Code            string    `gorm:"type:varchar(50);unique;not null;index" json:"code"`
	Description     string    `gorm:"type:text" json:"description"`

	ChartOfAccountID           string          `gorm:"type:uuid;not null;index" json:"chart_of_account_id"`
	ChartOfAccount             *ChartOfAccount `gorm:"foreignKey:ChartOfAccountID" json:"chart_of_account,omitempty"`
	ChartOfAccountCodeSnapshot string          `gorm:"type:varchar(50)" json:"chart_of_account_code_snapshot,omitempty"`
	ChartOfAccountNameSnapshot string          `gorm:"type:varchar(200)" json:"chart_of_account_name_snapshot,omitempty"`
	ChartOfAccountTypeSnapshot string          `gorm:"type:varchar(20)" json:"chart_of_account_type_snapshot,omitempty"`
	Amount                     float64         `gorm:"type:numeric(18,2);not null" json:"amount"`

	VendorName  string                `gorm:"type:varchar(200)" json:"vendor_name"`
	DueDate     *time.Time            `gorm:"type:date;index" json:"due_date"`
	ReferenceNo string                `gorm:"type:varchar(100);index" json:"reference_no"`
	Status      NonTradePayableStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`

	CreatedBy *string `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (NonTradePayable) TableName() string {
	return "non_trade_payables"
}

func (n *NonTradePayable) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}
