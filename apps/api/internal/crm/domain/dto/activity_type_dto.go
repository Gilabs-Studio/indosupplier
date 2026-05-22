package dto

import "time"

// CreateActivityTypeRequest represents the request to create an activity type
type CreateActivityTypeRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	Icon        string `json:"icon" binding:"max=50"`
	BadgeColor  string `json:"badge_color" binding:"max=20"`
}

// UpdateActivityTypeRequest represents the request to update an activity type
type UpdateActivityTypeRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	Icon        string `json:"icon" binding:"max=50"`
	BadgeColor  string `json:"badge_color" binding:"max=20"`
	IsActive    *bool  `json:"is_active"`
}

// ActivityTypeResponse represents the response for an activity type
type ActivityTypeResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	BadgeColor  string    `json:"badge_color"`
	Order       int       `json:"order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
