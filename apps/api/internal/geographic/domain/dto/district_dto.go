package dto

// CreateDistrictRequest represents create district request
type CreateDistrictRequest struct {
	CityID   string `json:"city_id" binding:"required,uuid"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Code     string `json:"code" binding:"required,min=2,max=20"`
	IsActive *bool  `json:"is_active"`
}

// UpdateDistrictRequest represents update district request
type UpdateDistrictRequest struct {
	CityID   string `json:"city_id" binding:"omitempty,uuid"`
	Name     string `json:"name" binding:"omitempty,min=2,max=100"`
	Code     string `json:"code" binding:"omitempty,min=2,max=20"`
	IsActive *bool  `json:"is_active"`
}

// ListDistrictsRequest represents list districts request
type ListDistrictsRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search" binding:"omitempty,max=100"`
	CityID  string `form:"city_id" binding:"omitempty,uuid"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=name code created_at updated_at"`
	SortDir string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// DistrictResponse represents district response
type DistrictResponse struct {
	ID        string        `json:"id"`
	CityID    string        `json:"city_id"`
	City      *CityResponse `json:"city,omitempty"`
	Name      string        `json:"name"`
	Code      string        `json:"code"`
	IsActive  bool          `json:"is_active"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}
