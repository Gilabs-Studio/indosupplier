package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DepreciationMethod string

const (
	DepreciationMethodStraightLine      DepreciationMethod = "SL"
	DepreciationMethodDecliningBalance  DepreciationMethod = "DB"
	DepreciationMethodSumOfYearsDigits  DepreciationMethod = "SYD"
	DepreciationMethodUnitsOfProduction DepreciationMethod = "UOP"
	DepreciationMethodNone              DepreciationMethod = "NONE"
)

type AssetCategoryType string

const (
	AssetCategoryTypeFixed      AssetCategoryType = "FIXED"
	AssetCategoryTypeCurrent    AssetCategoryType = "CURRENT"
	AssetCategoryTypeIntangible AssetCategoryType = "INTANGIBLE"
	AssetCategoryTypeOther      AssetCategoryType = "OTHER"
)

type AssetCategory struct {
	ID       string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`

	Name string            `gorm:"type:varchar(150);not null;uniqueIndex" json:"name"`
	Type AssetCategoryType `gorm:"type:varchar(20);not null;default:'FIXED'" json:"type"`

	DepreciationMethod DepreciationMethod `gorm:"type:varchar(10);not null" json:"depreciation_method"`
	UsefulLifeMonths   int                `gorm:"not null" json:"useful_life_months"`
	DepreciationRate   float64            `gorm:"type:numeric(8,4);default:0" json:"depreciation_rate"`
	IsDepreciable      bool               `gorm:"default:true" json:"is_depreciable"`

	AssetAccountID                   string  `gorm:"type:uuid;not null;index" json:"asset_account_id"`
	AccumulatedDepreciationAccountID string  `gorm:"type:uuid;not null;index" json:"accumulated_depreciation_account_id"`
	DepreciationExpenseAccountID     string  `gorm:"type:uuid;not null;index" json:"depreciation_expense_account_id"`
	DisposalGainAccountID            *string `gorm:"type:uuid;index" json:"disposal_gain_account_id,omitempty"`
	DisposalLossAccountID            *string `gorm:"type:uuid;index" json:"disposal_loss_account_id,omitempty"`

	IsActive bool `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AssetCategory) TableName() string {
	return "asset_categories"
}

func (c *AssetCategory) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}
