package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountType string

const (
	AccountTypeAsset        AccountType = "ASSET"
	AccountTypeLiability    AccountType = "LIABILITY"
	AccountTypeEquity       AccountType = "EQUITY"
	AccountTypeRevenue      AccountType = "REVENUE"
	AccountTypeExpense      AccountType = "EXPENSE"
	AccountTypeCashBank     AccountType = "CASH_BANK"
	AccountTypeCurrentAsset AccountType = "CURRENT_ASSET"
	AccountTypeFixedAsset   AccountType = "FIXED_ASSET"
	AccountTypeTradePayable AccountType = "TRADE_PAYABLE"
	AccountTypeCOGS         AccountType = "COST_OF_GOODS_SOLD"
	AccountTypeSalaryWages  AccountType = "SALARY_WAGES"
	AccountTypeOperational  AccountType = "OPERATIONAL"
)

type ChartOfAccount struct {
	ID       string      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string      `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code     string      `gorm:"type:varchar(50);not null;index" json:"code"`
	Name     string      `gorm:"type:varchar(200);not null;index" json:"name"`
	Type     AccountType `gorm:"type:varchar(20);not null;index" json:"type"`
	ParentID *string     `gorm:"type:uuid;index" json:"parent_id"`
	Parent   *ChartOfAccount
	Children []ChartOfAccount `gorm:"foreignKey:ParentID" json:"children,omitempty"`

	IsActive    bool `gorm:"default:true;index" json:"is_active"`
	IsPostable  bool `gorm:"default:true;index" json:"is_postable"`
	IsProtected bool `gorm:"default:false;index" json:"is_protected"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChartOfAccount) TableName() string {
	return "chart_of_accounts"
}

func (coa *ChartOfAccount) BeforeCreate(tx *gorm.DB) error {
	if coa.ID == "" {
		coa.ID = uuid.New().String()
	}
	return nil
}
