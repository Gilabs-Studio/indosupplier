package dto

// CreateJobPositionRequest represents create job position request
type CreateJobPositionRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
	IsActive    *bool  `json:"is_active"`
}

// UpdateJobPositionRequest represents update job position request
type UpdateJobPositionRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
	IsActive    *bool  `json:"is_active"`
}

// ListJobPositionsRequest represents list job positions request
type ListJobPositionsRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search" binding:"omitempty,max=100"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=name created_at updated_at"`
	SortDir string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// JobPositionResponse represents job position response
type JobPositionResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
