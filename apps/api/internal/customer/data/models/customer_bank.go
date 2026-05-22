package models

import (
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	supplierModels "github.com/gilabs/gims/api/internal/supplier/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomerBank represents a bank account associated with a customer
type CustomerBank struct {
	ID            string               `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	CustomerID    string               `gorm:"type:uuid;not null;index" json:"customer_id"`
	BankID        string               `gorm:"type:uuid;not null;index" json:"bank_id"`
	Bank          *supplierModels.Bank `gorm:"foreignKey:BankID" json:"bank,omitempty"`
	CurrencyID    *string              `gorm:"type:uuid;index" json:"currency_id"`
	Currency      *coreModels.Currency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
	AccountNumber string               `gorm:"type:varchar(50);not null" json:"account_number"`
	AccountName   string               `gorm:"type:varchar(100);not null" json:"account_name"`
	Branch        string               `gorm:"type:varchar(100)" json:"branch"`
	IsPrimary     bool                 `gorm:"default:false" json:"is_primary"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	DeletedAt     gorm.DeletedAt       `gorm:"index" json:"-"`
}

// TableName specifies the table name for CustomerBank
func (CustomerBank) TableName() string {
	return "customer_banks"
}

// BeforeCreate hook to generate UUID
func (cb *CustomerBank) BeforeCreate(tx *gorm.DB) error {
	if cb.ID == "" {
		cb.ID = uuid.New().String()
	}
	return nil
}
