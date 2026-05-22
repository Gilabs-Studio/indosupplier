package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PosSessionStatus represents the status of a POS session
type PosSessionStatus string

const (
	PosSessionStatusOpen   PosSessionStatus = "OPEN"
	PosSessionStatusClosed PosSessionStatus = "CLOSED"
)

// PosSession represents an open POS cashier shift for an outlet
type PosSession struct {
	ID          string           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code        string           `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	OutletID    string           `gorm:"type:uuid;not null;index" json:"outlet_id"`
	WarehouseID string           `gorm:"type:uuid;not null;index" json:"warehouse_id"`
	CashierID   string           `gorm:"type:uuid;not null;index" json:"cashier_id"`
	OpeningCash float64          `gorm:"type:decimal(15,2);default:0" json:"opening_cash"`
	ClosingCash *float64         `gorm:"type:decimal(15,2)" json:"closing_cash"`
	Status      PosSessionStatus `gorm:"type:varchar(20);not null;default:'OPEN';index" json:"status"`
	OpenedAt    time.Time        `gorm:"not null" json:"opened_at"`
	ClosedAt    *time.Time       `json:"closed_at"`
	Notes       *string          `gorm:"type:text" json:"notes"`
	TotalSales  float64          `gorm:"type:decimal(15,2);default:0" json:"total_sales"`
	TotalOrders int              `gorm:"default:0" json:"total_orders"`
	CreatedBy   string           `gorm:"type:uuid;not null;index" json:"created_by"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `gorm:"index" json:"updated_at"`
	DeletedAt   gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (PosSession) TableName() string {
	return "pos_sessions"
}

func (s *PosSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
