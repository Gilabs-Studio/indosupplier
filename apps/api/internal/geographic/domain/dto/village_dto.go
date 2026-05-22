package dto

// CreateVillageRequest represents create village request
type CreateVillageRequest struct {
	DistrictID string `json:"district_id" binding:"required,uuid"`
	Name       string `json:"name" binding:"required,min=2,max=100"`
	Code       string `json:"code" binding:"required,min=2,max=20"`
	PostalCode string `json:"postal_code" binding:"omitempty,max=10"`
	Type       string `json:"type" binding:"omitempty,oneof=village kelurahan"`
	IsActive   *bool  `json:"is_active"`
}

// UpdateVillageRequest represents update village request
type UpdateVillageRequest struct {
	DistrictID string `json:"district_id" binding:"omitempty,uuid"`
	Name       string `json:"name" binding:"omitempty,min=2,max=100"`
	Code       string `json:"code" binding:"omitempty,min=2,max=20"`
	PostalCode string `json:"postal_code" binding:"omitempty,max=10"`
	Type       string `json:"type" binding:"omitempty,oneof=village kelurahan"`
	IsActive   *bool  `json:"is_active"`
}

// ListVillagesRequest represents list villages request
type ListVillagesRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PerPage    int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search     string `form:"search" binding:"omitempty,max=100"`
	DistrictID string `form:"district_id" binding:"omitempty,uuid"`
	Type       string `form:"type" binding:"omitempty,oneof=village kelurahan"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=name code type postal_code created_at updated_at"`
	SortDir    string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// VillageResponse represents village response
type VillageResponse struct {
	ID         string            `json:"id"`
	DistrictID string            `json:"district_id"`
	District   *DistrictResponse `json:"district,omitempty"`
	Name       string            `json:"name"`
	Code       string            `json:"code"`
	PostalCode string            `json:"postal_code"`
	Type       string            `json:"type"`
	IsActive   bool              `json:"is_active"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
}
