package dto

// ─── Public Self-Order DTOs ───────────────────────────────────────────────────

// PublicTableInfoResponse is the landing payload for a QR-scanned table URL.
// It contains everything the customer needs to browse the menu and place an order.
type PublicTableInfoResponse struct {
	OutletID   string           `json:"outlet_id"`
	OutletName string           `json:"outlet_name"`
	TableLabel string           `json:"table_label"`
	Token      string           `json:"token"`
	Catalog    []POSCatalogItem `json:"catalog"`
	Config     PublicPOSConfig  `json:"config"`
}

// PublicPOSConfig exposes tax/service rates needed for cart total calculation.
type PublicPOSConfig struct {
	TaxRate           float64 `json:"tax_rate"`
	ServiceChargeRate float64 `json:"service_charge_rate"`
	Currency          string  `json:"currency"`
}

// PublicCartItem represents a single product line in the customer's self-order cart.
type PublicCartItem struct {
	ProductID string  `json:"product_id" binding:"required,uuid"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Notes     *string `json:"notes"`
}

// CreateCustomerOrderRequest is the body for POST /public/pos/tables/:token/orders.
type CreateCustomerOrderRequest struct {
	CustomerName string           `json:"customer_name" binding:"required,min=1,max=100"`
	Items        []PublicCartItem `json:"items" binding:"required,min=1,dive"`
	Notes        *string          `json:"notes"`
}

// CustomerOrderResponse is returned after a customer places an order and
// on subsequent status polling requests.
type CustomerOrderResponse struct {
	OrderID       string                 `json:"order_id"`
	OrderNumber   string                 `json:"order_number"`
	TableLabel    string                 `json:"table_label"`
	Status        string                 `json:"status"`
	PaymentStatus *string                `json:"payment_status,omitempty"`
	CancelReason  *string                `json:"cancel_reason,omitempty"`
	Subtotal      float64                `json:"subtotal"`
	TaxAmount     float64                `json:"tax_amount"`
	TotalAmount   float64                `json:"total_amount"`
	Items         []POSOrderItemResponse `json:"items"`
	CreatedAt     string                 `json:"created_at"`
}

// InitiateCustomerPaymentRequest is the body for
// POST /public/pos/tables/:token/orders/:orderId/pay/digital.
type InitiateCustomerPaymentRequest struct {
	// Method must be QRIS or DIGITAL.
	Method      string  `json:"method" binding:"required,oneof=QRIS DIGITAL"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ChannelCode *string `json:"channel_code"`
}
