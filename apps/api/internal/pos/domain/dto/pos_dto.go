package dto

import "time"

// ─── Session DTOs ────────────────────────────────────────────────────────────

// OpenSessionRequest opens a new POS cashier shift
type OpenSessionRequest struct {
	OutletID    string  `json:"outlet_id" binding:"required,uuid"`
	WarehouseID string  `json:"warehouse_id" binding:"required,uuid"`
	OpeningCash float64 `json:"opening_cash"`
	Notes       *string `json:"notes"`
}

// CloseSessionRequest closes the current POS session
type CloseSessionRequest struct {
	ClosingCash float64 `json:"closing_cash"`
	Notes       *string `json:"notes"`
}

// POSSessionResponse returned to client
type POSSessionResponse struct {
	ID          string   `json:"id"`
	Code        string   `json:"code"`
	OutletID    string   `json:"outlet_id"`
	WarehouseID string   `json:"warehouse_id"`
	CashierID   string   `json:"cashier_id"`
	OpeningCash float64  `json:"opening_cash"`
	ClosingCash *float64 `json:"closing_cash"`
	Status      string   `json:"status"`
	TotalSales  float64  `json:"total_sales"`
	TotalOrders int      `json:"total_orders"`
	OpenedAt    string   `json:"opened_at"`
	ClosedAt    *string  `json:"closed_at"`
	Notes       *string  `json:"notes"`
	CreatedAt   string   `json:"created_at"`
}

// ─── Order DTOs ──────────────────────────────────────────────────────────────

// CreateOrderRequest creates a new POS order
type CreateOrderRequest struct {
	OutletID     string  `json:"outlet_id" binding:"required,uuid"`
	OrderType    string  `json:"order_type" binding:"required"`
	TableID      *string `json:"table_id"`
	TableLabel   *string `json:"table_label"`
	CustomerID   *string `json:"customer_id"`
	CustomerName *string `json:"customer_name"`
	GuestCount   int     `json:"guest_count"`
	Notes        *string `json:"notes"`
}

// ConfirmOrderRequest locks the order for checkout (validates stock first)
type ConfirmOrderRequest struct {
	Notes *string `json:"notes"`
}

// VoidOrderRequest voids the order with a reason
type VoidOrderRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// AssignTableRequest assigns or re-assigns a table to an order
type AssignTableRequest struct {
	TableID    string `json:"table_id" binding:"required"`
	TableLabel string `json:"table_label"`
}

// AddOrderItemRequest adds a product line to an order
type AddOrderItemRequest struct {
	ProductID string  `json:"product_id" binding:"required,uuid"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Notes     *string `json:"notes"`
}

// UpdateOrderItemRequest modifies quantity/notes of an existing line
type UpdateOrderItemRequest struct {
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Notes    *string `json:"notes"`
}

// POSOrderItemResponse represents a single item in an order response
type POSOrderItemResponse struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductCode string  `json:"product_code"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Discount    float64 `json:"discount"`
	Subtotal    float64 `json:"subtotal"`
	Notes       *string `json:"notes"`
	Status      string  `json:"status"`
}

// POSOrderResponse returned to client for order operations
type POSOrderResponse struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"tenant_id,omitempty"`
	OrderNumber       string                 `json:"order_number"`
	SessionID         *string                `json:"session_id,omitempty"`
	OutletID          string                 `json:"outlet_id"`
	OrderType         string                 `json:"order_type"`
	TableID           *string                `json:"table_id"`
	TableLabel        *string                `json:"table_label"`
	CustomerID        *string                `json:"customer_id"`
	CustomerName      *string                `json:"customer_name"`
	GuestCount        int                    `json:"guest_count"`
	Subtotal          float64                `json:"subtotal"`
	DiscountAmount    float64                `json:"discount_amount"`
	TaxAmount         float64                `json:"tax_amount"`
	ServiceCharge     float64                `json:"service_charge"`
	TotalAmount       float64                `json:"total_amount"`
	Status            string                 `json:"status"`
	OrderSource       string                 `json:"order_source,omitempty"`
	VoidReason        *string                `json:"void_reason,omitempty"`
	Notes             *string                `json:"notes"`
	SalesOrderID      *string                `json:"sales_order_id"`
	CustomerInvoiceID *string                `json:"customer_invoice_id"`
	LoyaltyMemberID   *string                `json:"loyalty_member_id,omitempty"`
	LoyaltyRewardID   *string                `json:"loyalty_reward_id,omitempty"`
	Items             []POSOrderItemResponse `json:"items"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// ─── Catalog DTOs ─────────────────────────────────────────────────────────────

// POSCatalogItem represents a product available for sale in POS
type POSCatalogItem struct {
	ProductID   string  `json:"product_id"`
	ProductCode string  `json:"product_code"`
	ProductName string  `json:"product_name"`
	ProductKind string  `json:"product_kind"`
	Price       float64 `json:"price"`
	Stock       float64 `json:"stock"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
	IsAvailable bool    `json:"is_available"`
}

// ─── POS Config DTOs ──────────────────────────────────────────────────────────

// UpsertPOSConfigRequest creates or updates POS outlet configuration
type UpsertPOSConfigRequest struct {
	TaxRate                 *float64 `json:"tax_rate"`
	ServiceChargeRate       *float64 `json:"service_charge_rate"`
	AllowDiscount           *bool    `json:"allow_discount"`
	MaxDiscountPercent      *float64 `json:"max_discount_percent"`
	PrintReceiptAuto        *bool    `json:"print_receipt_auto"`
	ReceiptFooter           *string  `json:"receipt_footer"`
	ReceiptWhatsAppTemplate *string  `json:"receipt_whatsapp_template"`
	Currency                *string  `json:"currency"`
}

// UpdateReceiptWhatsAppTemplateRequest updates receipt WhatsApp template for one outlet.
type UpdateReceiptWhatsAppTemplateRequest struct {
	ReceiptWhatsAppTemplate *string `json:"receipt_whatsapp_template"`
}

// POSConfigResponse returned to client
type POSConfigResponse struct {
	ID                      string  `json:"id"`
	OutletID                string  `json:"outlet_id"`
	TaxRate                 float64 `json:"tax_rate"`
	ServiceChargeRate       float64 `json:"service_charge_rate"`
	AllowDiscount           bool    `json:"allow_discount"`
	MaxDiscountPercent      float64 `json:"max_discount_percent"`
	PrintReceiptAuto        bool    `json:"print_receipt_auto"`
	ReceiptFooter           *string `json:"receipt_footer"`
	ReceiptWhatsAppTemplate *string `json:"receipt_whatsapp_template"`
	Currency                string  `json:"currency"`
}

// ─── Xendit Config DTOs ───────────────────────────────────────────────────────

// ConnectXenditRequest saves Xendit credentials and marks the account as connected
type ConnectXenditRequest struct {
	SecretKey       string `json:"secret_key" binding:"required"`
	XenditAccountID string `json:"xendit_account_id"` // XenPlatform sub-account ID
	BusinessName    string `json:"business_name"`
	Environment     string `json:"environment" binding:"required,oneof=sandbox production"`
	WebhookToken    string `json:"webhook_token"`
}

// UpdateXenditConfigRequest updates non-credential settings (environment, active state)
type UpdateXenditConfigRequest struct {
	Environment  string `json:"environment" binding:"omitempty,oneof=sandbox production"`
	BusinessName string `json:"business_name"`
	IsActive     *bool  `json:"is_active"`
}

// XenditConfigResponse returned to client — secret key and webhook token are never exposed
type XenditConfigResponse struct {
	ID               string `json:"id"`
	CompanyID        string `json:"company_id"`
	XenditAccountID  string `json:"xendit_account_id"`
	BusinessName     string `json:"business_name"`
	Environment      string `json:"environment"`
	ConnectionStatus string `json:"connection_status"`
	IsActive         bool   `json:"is_active"`
	UpdatedAt        string `json:"updated_at"`
}

// XenditConnectionStatusResponse is a lightweight check payload for cashiers
type XenditConnectionStatusResponse struct {
	IsConnected bool   `json:"is_connected"`
	Status      string `json:"status"`
}

// TestXenditConnectionRequest validates credentials before saving config.
type TestXenditConnectionRequest struct {
	SecretKey       string `json:"secret_key" binding:"required"`
	XenditAccountID string `json:"xendit_account_id"`
}

// TestXenditConnectionResponse indicates whether Xendit API is reachable with provided credentials.
type TestXenditConnectionResponse struct {
	Reachable bool   `json:"reachable"`
	Message   string `json:"message"`
}
