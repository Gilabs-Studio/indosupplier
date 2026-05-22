package dto

// CreateCityRequest represents create city request
type CreateCityRequest struct {
	ProvinceID string `json:"province_id" binding:"required,uuid"`
	Name       string `json:"name" binding:"required,min=2,max=100"`
	Code       string `json:"code" binding:"required,min=2,max=20"`
	Type       string `json:"type" binding:"omitempty,oneof=city regency"`
	IsActive   *bool  `json:"is_active"`
}

// UpdateCityRequest represents update city request
type UpdateCityRequest struct {
	ProvinceID string `json:"province_id" binding:"omitempty,uuid"`
	Name       string `json:"name" binding:"omitempty,min=2,max=100"`
	Code       string `json:"code" binding:"omitempty,min=2,max=20"`
	Type       string `json:"type" binding:"omitempty,oneof=city regency"`
	IsActive   *bool  `json:"is_active"`
}

// ListCitiesRequest represents list cities request
type ListCitiesRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PerPage    int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search     string `form:"search" binding:"omitempty,max=100"`
	ProvinceID string `form:"province_id" binding:"omitempty,uuid"`
	Type       string `form:"type" binding:"omitempty,oneof=city regency"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=name code type created_at updated_at"`
	SortDir    string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// CityResponse represents city response
type CityResponse struct {
	ID         string            `json:"id"`
	ProvinceID string            `json:"province_id"`
	Province   *ProvinceResponse `json:"province,omitempty"`
	Name       string            `json:"name"`
	Code       string            `json:"code"`
	Type       string            `json:"type"`
	IsActive   bool              `json:"is_active"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
}
