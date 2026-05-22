package dto

import "time"

// === PaymentTerms DTOs ===

type CreatePaymentTermsRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	Days        int    `json:"days" binding:"min=0"`
	IsActive    *bool  `json:"is_active"`
}

type UpdatePaymentTermsRequest struct {
	Code        string `json:"code" binding:"omitempty,min=2,max=20"`
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	Days        *int   `json:"days" binding:"omitempty,min=0"`
	IsActive    *bool  `json:"is_active"`
}

type PaymentTermsResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Days        int       `json:"days"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
