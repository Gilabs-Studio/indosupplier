package dto

// CreateCountryRequest represents create country request
type CreateCountryRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	Code      string `json:"code" binding:"required,min=2,max=10"`
	PhoneCode string `json:"phone_code" binding:"omitempty,max=10"`
	IsActive  *bool  `json:"is_active"`
}

// UpdateCountryRequest represents update country request
type UpdateCountryRequest struct {
	Name      string `json:"name" binding:"omitempty,min=2,max=100"`
	Code      string `json:"code" binding:"omitempty,min=2,max=10"`
	PhoneCode string `json:"phone_code" binding:"omitempty,max=10"`
	IsActive  *bool  `json:"is_active"`
}

// ListCountriesRequest represents list countries request
type ListCountriesRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search" binding:"omitempty,max=100"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=name code created_at updated_at"`
	SortDir string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// CountryResponse represents country response
type CountryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	PhoneCode string `json:"phone_code"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
