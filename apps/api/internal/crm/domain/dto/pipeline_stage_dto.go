package dto

import "time"

// CreatePipelineStageRequest represents the request to create a pipeline stage
type CreatePipelineStageRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Color       string `json:"color" binding:"max=20"`
	Probability int    `json:"probability" binding:"min=0,max=100"`
	IsWon       *bool  `json:"is_won"`
	IsLost      *bool  `json:"is_lost"`
	Description string `json:"description" binding:"max=500"`
}

// UpdatePipelineStageRequest represents the request to update a pipeline stage
type UpdatePipelineStageRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Color       string `json:"color" binding:"max=20"`
	Probability *int   `json:"probability" binding:"omitempty,min=0,max=100"`
	IsWon       *bool  `json:"is_won"`
	IsLost      *bool  `json:"is_lost"`
	IsActive    *bool  `json:"is_active"`
	Description string `json:"description" binding:"max=500"`
}

// PipelineStageResponse represents the response for a pipeline stage
type PipelineStageResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Order       int       `json:"order"`
	Color       string    `json:"color"`
	Probability int       `json:"probability"`
	IsWon       bool      `json:"is_won"`
	IsLost      bool      `json:"is_lost"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
