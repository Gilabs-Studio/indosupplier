package dto

import "time"

// CreateContactRoleRequest represents the request to create a contact role
type CreateContactRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	BadgeColor  string `json:"badge_color" binding:"max=20"`
}

// UpdateContactRoleRequest represents the request to update a contact role
type UpdateContactRoleRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	BadgeColor  string `json:"badge_color" binding:"max=20"`
	IsActive    *bool  `json:"is_active"`
}

// ContactRoleResponse represents the response for a contact role
type ContactRoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	BadgeColor  string    `json:"badge_color"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
