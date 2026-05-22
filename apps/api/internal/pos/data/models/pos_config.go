package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// POSConfig holds per-outlet POS configuration and business rules
type POSConfig struct {
	ID                      string         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID                string         `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	OutletID                string         `gorm:"type:uuid;not null;uniqueIndex" json:"outlet_id"`
	TaxRate                 float64        `gorm:"type:decimal(5,2);default:11.00" json:"tax_rate"`
	ServiceChargeRate       float64        `gorm:"type:decimal(5,2);default:0" json:"service_charge_rate"`
	AllowDiscount           bool           `gorm:"default:true" json:"allow_discount"`
	MaxDiscountPercent      float64        `gorm:"type:decimal(5,2);default:100" json:"max_discount_percent"`
	PrintReceiptAuto        bool           `gorm:"default:false" json:"print_receipt_auto"`
	ReceiptFooter           *string        `gorm:"type:text" json:"receipt_footer"`
	ReceiptWhatsAppTemplate *string        `gorm:"type:text" json:"receipt_whatsapp_template"`
	Currency                string         `gorm:"type:varchar(10);default:'IDR'" json:"currency"`
	UpdatedBy               *string        `gorm:"type:uuid" json:"updated_by"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`
}

func (POSConfig) TableName() string {
	return "pos_configs"
}

func (p *POSConfig) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
