package dto

import "time"

// CreateLeadSourceRequest represents the request to create a lead source
type CreateLeadSourceRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateLeadSourceRequest represents the request to update a lead source
type UpdateLeadSourceRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	IsActive    *bool  `json:"is_active"`
}

// LeadSourceResponse represents the response for a lead source
type LeadSourceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Order       int       `json:"order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
