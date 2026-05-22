package dto

import "time"

// ProcessPaymentRequest initiates payment for a POS order
type ProcessPaymentRequest struct {
	Method          string  `json:"method" binding:"required,oneof=CASH CARD QRIS TRANSFER DIGITAL"`
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	ReferenceNumber *string `json:"reference_number"`
	Notes           *string `json:"notes"`
	// CustomerName optionally captures the customer name on receipt (F&B use-case)
	CustomerName *string `json:"customer_name"`
	// CustomerID links the payment to an existing customer master record when selected.
	CustomerID *string `json:"customer_id"`
	// CustomerPhone optionally captures the customer phone for SO/invoice records
	CustomerPhone *string `json:"customer_phone"`
	// CustomerEmail optionally captures the customer email for SO/invoice records
	CustomerEmail *string `json:"customer_email"`
	// LoyaltyMemberID links the order to an existing loyalty member so points are
	// awarded automatically after payment.
	LoyaltyMemberID *string `json:"loyalty_member_id"`
	// LoyaltyRewardID stores a selected reward choice that will be redeemed
	// atomically when payment is finalized.
	LoyaltyRewardID *string `json:"loyalty_reward_id"`
	// ChannelCode specifies the Xendit payment channel (e.g. QRIS, BCA, DANA, OVO)
	ChannelCode *string `json:"channel_code"`
}

// CancelPendingPaymentRequest carries an optional cashier reason when cancelling
// a pending payment.
type CancelPendingPaymentRequest struct {
	Reason *string `json:"reason" binding:"omitempty,max=500"`
}

// POSPaymentResponse returned to client for payment operations
type POSPaymentResponse struct {
	ID              string     `json:"id"`
	OrderID         string     `json:"order_id"`
	Method          string     `json:"method"`
	Status          string     `json:"status"`
	Amount          float64    `json:"amount"`
	TenderAmount    float64    `json:"tender_amount"`
	ChangeAmount    float64    `json:"change_amount"`
	ReferenceNumber *string    `json:"reference_number"`
	TransactionID   *string    `json:"transaction_id"`
	PaymentType     *string    `json:"payment_type"`
	VaNumber        *string    `json:"va_number"`
	QrCode          *string    `json:"qr_code"`
	PaymentURL      *string    `json:"payment_url"`
	ChannelCode     *string    `json:"channel_code"`
	ExpiresAt       *time.Time `json:"expires_at"`
	PaidAt          *time.Time `json:"paid_at"`
	Notes           *string    `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
}

// XenditWebhookPayload is the Xendit server-to-server notification body for invoice events
type XenditWebhookPayload struct {
	// ExternalID matches the ExternalOrderID stored in POSPayment
	ExternalID string `json:"external_id"`
	// Status: PAID, EXPIRED, or other Xendit invoice statuses
	Status         string  `json:"status"`
	ID             string  `json:"id"` // Xendit invoice ID
	PaymentMethod  string  `json:"payment_method"`
	PaymentChannel string  `json:"payment_channel"`
	PaidAmount     float64 `json:"paid_amount"`
	PaidAt         string  `json:"paid_at"`
}
