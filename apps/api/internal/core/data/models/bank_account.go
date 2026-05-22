package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BankAccount represents a company bank account (cash/bank) used for payments.
type BankAccount struct {
	ID               string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID         string         `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	CompanyID        string         `gorm:"type:uuid;not null;index:idx_company_code,unique" json:"company_id"`
	Code             string         `gorm:"type:varchar(50);not null;index:idx_company_code,unique" json:"code"`
	Name             string         `gorm:"type:varchar(150);not null;index" json:"name"`
	AccountType      string         `gorm:"type:varchar(50);not null;default:'operational'" json:"account_type"` // operational|suspense|transit
	BankID           *string        `gorm:"type:uuid;index" json:"bank_id,omitempty"`
	AccountNumber    string         `gorm:"type:varchar(50);not null;uniqueIndex:idx_company_account,unique" json:"account_number"`
	AccountHolder    string         `gorm:"type:varchar(150);not null" json:"account_holder"`
	CurrencyID       *string        `gorm:"type:uuid;index" json:"currency_id"`
	CurrencyDetail   *Currency      `gorm:"foreignKey:CurrencyID" json:"currency_detail,omitempty"`
	Currency         string         `gorm:"type:varchar(10);not null;default:'IDR';index" json:"currency"`
	ChartOfAccountID *string        `gorm:"type:uuid;index" json:"chart_of_account_id"`
	VillageID        *string        `gorm:"type:uuid;index" json:"village_id"`
	BankAddress      string         `gorm:"type:varchar(255)" json:"bank_address"`
	BankPhone        string         `gorm:"type:varchar(20)" json:"bank_phone"`
	OpeningBalance   float64        `gorm:"type:decimal(20,4);default:0" json:"opening_balance"` // documentation only
	IsActive         bool           `gorm:"default:true;index" json:"is_active"`
	CountryCode      string         `gorm:"type:varchar(2)" json:"country_code"`
	BankBranchCode   string         `gorm:"type:varchar(20)" json:"bank_branch_code"`
	CreatedBy        string         `gorm:"type:uuid" json:"created_by"`
	UpdatedBy        string         `gorm:"type:uuid" json:"updated_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BankAccount) TableName() string {
	return "bank_accounts"
}

func (b *BankAccount) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}
