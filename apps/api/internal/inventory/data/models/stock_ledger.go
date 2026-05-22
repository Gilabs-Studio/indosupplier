package models

import (
	"time"

	productModels "github.com/gilabs/gims/api/internal/product/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StockLedger is an immutable moving-average log per stock-changing transaction.
type StockLedger struct {
	ID              string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID        string    `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	ProductID       string    `gorm:"type:uuid;not null;index" json:"product_id"`
	TransactionID   string    `gorm:"type:varchar(255);not null;index" json:"transaction_id"`
	TransactionType string    `gorm:"type:varchar(30);not null;index" json:"transaction_type"`
	Qty             float64   `gorm:"type:decimal(15,3);not null" json:"qty"`
	UnitCost        float64   `gorm:"type:decimal(18,4);not null;default:0" json:"unit_cost"`
	AverageCost     float64   `gorm:"type:decimal(18,4);not null;default:0" json:"average_cost"`
	StockValue      float64   `gorm:"type:decimal(18,4);not null;default:0" json:"stock_value"`
	RunningQty      float64   `gorm:"type:decimal(15,3);not null;default:0" json:"running_qty"`
	CreatedAt       time.Time `json:"created_at"`

	Product *productModels.Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (StockLedger) TableName() string {
	return "stock_ledgers"
}

func (sl *StockLedger) BeforeCreate(tx *gorm.DB) error {
	if sl.ID == "" {
		sl.ID = uuid.New().String()
	}

	if sl.TenantID == "" {
		if tid, ok := tx.Statement.Context.Value("tenant_id").(string); ok {
			sl.TenantID = tid
		}
	}
	return nil
}
