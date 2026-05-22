package dto

import "time"

// === CourierAgency DTOs ===

type CreateCourierAgencyRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	Phone       string `json:"phone" binding:"max=20"`
	Address     string `json:"address" binding:"max=500"`
	TrackingURL string `json:"tracking_url" binding:"max=255"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateCourierAgencyRequest struct {
	Code        string `json:"code" binding:"omitempty,min=2,max=20"`
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	Phone       string `json:"phone" binding:"max=20"`
	Address     string `json:"address" binding:"max=500"`
	TrackingURL string `json:"tracking_url" binding:"max=255"`
	IsActive    *bool  `json:"is_active"`
}

type CourierAgencyResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	TrackingURL string    `json:"tracking_url"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
