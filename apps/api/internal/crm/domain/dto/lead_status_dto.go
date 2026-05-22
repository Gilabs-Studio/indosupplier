package dto

import "time"

// CreateLeadStatusRequest represents the request to create a lead status
type CreateLeadStatusRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	Score       int    `json:"score" binding:"min=0,max=100"`
	Color       string `json:"color" binding:"max=20"`
}

// UpdateLeadStatusRequest represents the request to update a lead status
type UpdateLeadStatusRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	Score       *int   `json:"score" binding:"omitempty,min=0,max=100"`
	Color       string `json:"color" binding:"max=20"`
	IsActive    *bool  `json:"is_active"`
}

// LeadStatusResponse represents the response for a lead status
type LeadStatusResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Score       int       `json:"score"`
	Color       string    `json:"color"`
	Order       int       `json:"order"`
	IsActive    bool      `json:"is_active"`
	IsDefault   bool      `json:"is_default"`
	IsConverted bool      `json:"is_converted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
