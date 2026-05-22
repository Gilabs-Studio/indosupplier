package dto

// CreateProvinceRequest represents create province request
type CreateProvinceRequest struct {
	CountryID string `json:"country_id" binding:"required,uuid"`
	Name      string `json:"name" binding:"required,min=2,max=100"`
	Code      string `json:"code" binding:"required,min=2,max=20"`
	IsActive  *bool  `json:"is_active"`
}

// UpdateProvinceRequest represents update province request
type UpdateProvinceRequest struct {
	CountryID string `json:"country_id" binding:"omitempty,uuid"`
	Name      string `json:"name" binding:"omitempty,min=2,max=100"`
	Code      string `json:"code" binding:"omitempty,min=2,max=20"`
	IsActive  *bool  `json:"is_active"`
}

// ListProvincesRequest represents list provinces request
type ListProvincesRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PerPage   int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search    string `form:"search" binding:"omitempty,max=100"`
	CountryID string `form:"country_id" binding:"omitempty,uuid"`
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=name code created_at updated_at"`
	SortDir   string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// ProvinceResponse represents province response
type ProvinceResponse struct {
	ID        string           `json:"id"`
	CountryID string           `json:"country_id"`
	Country   *CountryResponse `json:"country,omitempty"`
	Name      string           `json:"name"`
	Code      string           `json:"code"`
	IsActive  bool             `json:"is_active"`
	CreatedAt string           `json:"created_at"`
	UpdatedAt string           `json:"updated_at"`
}
