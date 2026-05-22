package models

import (
	"time"

	"gorm.io/gorm"
)

// TenantPaymentMethod stores a Xendit card token reference for a tenant.
// Raw card data is never persisted — only the opaque token ID returned by Xendit.
// This table enables the payment-method management UI (list, set default, remove).
type TenantPaymentMethod struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID        string         `gorm:"type:uuid;not null;index"                       json:"tenant_id"`
	XenditTokenID   string         `gorm:"type:varchar(255);not null;uniqueIndex"          json:"xendit_token_id"`
	MaskedCardNumber string        `gorm:"type:varchar(32)"                               json:"masked_card_number"`
	CardBrand       string         `gorm:"type:varchar(32)"                               json:"card_brand"`
	CardHolderName  string         `gorm:"type:varchar(255)"                              json:"card_holder_name"`
	ExpiryMonth     int            `gorm:"type:int"                                       json:"expiry_month"`
	ExpiryYear      int            `gorm:"type:int"                                       json:"expiry_year"`
	IsDefault       bool           `gorm:"type:boolean;not null;default:false"            json:"is_default"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index"                                          json:"-"`
}

// TableName specifies the table name.
func (TenantPaymentMethod) TableName() string {
	return "tenant_payment_methods"
}
