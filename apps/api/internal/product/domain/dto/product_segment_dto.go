package dto

import "time"

// === ProductSegment DTOs ===

type CreateProductSegmentRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateProductSegmentRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	IsActive    *bool  `json:"is_active"`
}

type ProductSegmentResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
