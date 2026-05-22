package models

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetStatus string

const (
	AssetStatusBorrowed AssetStatus = "BORROWED"
	AssetStatusReturned AssetStatus = "RETURNED"
)

type AssetCondition string

const (
	AssetConditionNew     AssetCondition = "NEW"
	AssetConditionGood    AssetCondition = "GOOD"
	AssetConditionFair    AssetCondition = "FAIR"
	AssetConditionPoor    AssetCondition = "POOR"
	AssetConditionDamaged AssetCondition = "DAMAGED"
)

type EmployeeAsset struct {
	ID              string          `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	EmployeeID      string          `gorm:"type:uuid;not null;index:idx_employee_asset_employee" json:"employee_id"`
	AssetID         *string         `gorm:"type:uuid;index:idx_employee_assets_asset_id" json:"asset_id,omitempty"`
	AssetName       string          `gorm:"type:varchar(200);not null" json:"asset_name"`
	AssetCode       string          `gorm:"type:varchar(100);not null;uniqueIndex" json:"asset_code"`
	AssetCategory   string          `gorm:"type:varchar(100);not null" json:"asset_category"`
	BorrowDate      time.Time       `gorm:"type:date;not null" json:"borrow_date"`
	ReturnDate      *time.Time      `gorm:"type:date" json:"return_date"`
	BorrowCondition AssetCondition  `gorm:"type:varchar(50);not null" json:"borrow_condition"`
	ReturnCondition *AssetCondition `gorm:"type:varchar(50)" json:"return_condition"`
	AssetImage      string          `gorm:"type:varchar(255)" json:"asset_image"`
	Notes           *string         `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`
}

func (e *EmployeeAsset) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

func (EmployeeAsset) TableName() string {
	return "employee_assets"
}

func (ea *EmployeeAsset) IsReturned() bool {
	return ea.ReturnDate != nil
}

func (ea *EmployeeAsset) GetStatus() AssetStatus {
	if ea.IsReturned() {
		return AssetStatusReturned
	}
	return AssetStatusBorrowed
}

func (ea *EmployeeAsset) DaysBorrowed() int {
	endDate := apptime.Now()
	if ea.IsReturned() {
		endDate = *ea.ReturnDate
	}
	duration := endDate.Sub(ea.BorrowDate)
	return int(duration.Hours() / 24)
}
