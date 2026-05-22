package dto

import "time"

// === Bank DTOs ===

type CreateBankRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	Code      string `json:"code" binding:"required,min=2,max=20"`
	SwiftCode string `json:"swift_code" binding:"max=20"`
	IsActive  *bool  `json:"is_active"`
}

type UpdateBankRequest struct {
	Name      string `json:"name" binding:"omitempty,min=2,max=100"`
	Code      string `json:"code" binding:"omitempty,min=2,max=20"`
	SwiftCode string `json:"swift_code" binding:"max=20"`
	IsActive  *bool  `json:"is_active"`
}

type BankResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	SwiftCode string    `json:"swift_code"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
