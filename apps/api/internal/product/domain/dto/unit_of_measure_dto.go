package dto

import "time"

// === UnitOfMeasure DTOs ===

type CreateUnitOfMeasureRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Symbol      string `json:"symbol" binding:"required,min=1,max=20"`
	Description string `json:"description" binding:"max=500"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateUnitOfMeasureRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=100"`
	Symbol      string `json:"symbol" binding:"omitempty,min=1,max=20"`
	Description string `json:"description" binding:"max=500"`
	IsActive    *bool  `json:"is_active"`
}

type UnitOfMeasureResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Symbol      string    `json:"symbol"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
