package dto

import "time"

type CreateTransactionRequest struct {
	SupplierProfileID string  `json:"supplier_profile_id" binding:"required,uuid"`
	RFQID             *string `json:"rfq_id" binding:"omitempty,uuid"`
	ProductName       string  `json:"product_name" binding:"required"`
	QuantityValue     float64 `json:"quantity_value" binding:"required,gt=0"`
	QuantityUnit      string  `json:"quantity_unit" binding:"required"`
	PricePerUnit      float64 `json:"price_per_unit" binding:"required,gt=0"`
	DeliveryAddress   string  `json:"delivery_address" binding:"required"`
	Notes             string  `json:"notes"`
}

type ListTransactionsRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Status  string `form:"status"`
}

type TransactionResponse struct {
	ID                string     `json:"id"`
	PONumber          string     `json:"po_number"`
	BuyerProfileID    string     `json:"buyer_profile_id"`
	SupplierProfileID string     `json:"supplier_profile_id"`
	SupplierName      string     `json:"supplier_name"`
	RFQID             *string    `json:"rfq_id,omitempty"`
	ProductName       string     `json:"product_name"`
	QuantityValue     float64    `json:"quantity_value"`
	QuantityUnit      string     `json:"quantity_unit"`
	PricePerUnit      float64    `json:"price_per_unit"`
	TotalAmount       float64    `json:"total_amount"`
	Status            string     `json:"status"`
	PaymentStatus     string     `json:"payment_status"`
	DeliveryAddress   string     `json:"delivery_address"`
	Notes             string     `json:"notes"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
