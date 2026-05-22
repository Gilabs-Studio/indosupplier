package dto

import "time"

// === LeaveType DTOs ===

type CreateLeaveTypeRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	MaxDays     int    `json:"max_days" binding:"min=0"`
	IsPaid      *bool  `json:"is_paid"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateLeaveTypeRequest struct {
	Code        string `json:"code" binding:"omitempty,min=2,max=20"`
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	MaxDays     *int   `json:"max_days" binding:"omitempty,min=0"`
	IsPaid      *bool  `json:"is_paid"`
	IsActive    *bool  `json:"is_active"`
}

type LeaveTypeResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MaxDays     int       `json:"max_days"`
	IsPaid      bool      `json:"is_paid"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
