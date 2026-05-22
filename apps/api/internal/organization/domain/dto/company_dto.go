package dto

import geographicDto "github.com/gilabs/gims/api/internal/geographic/domain/dto"

// CreateCompanyRequest represents create company request
type CreateCompanyRequest struct {
	Name       string  `json:"name" binding:"required,min=2,max=200"`
	Address    string  `json:"address" binding:"omitempty,max=500"`
	Email      string  `json:"email" binding:"omitempty,email,max=100"`
	Phone      string  `json:"phone" binding:"omitempty,max=20"`
	NPWP       string  `json:"npwp" binding:"omitempty,max=30"`
	NIB        string  `json:"nib" binding:"omitempty,max=30"`
	ProvinceID *string `json:"province_id" binding:"omitempty,uuid"`
	CityID     *string `json:"city_id" binding:"omitempty,uuid"`
	DistrictID *string `json:"district_id" binding:"omitempty,uuid"`
	VillageID  *string `json:"village_id" binding:"omitempty,uuid"`
	VillageName *string `json:"village_name" binding:"omitempty,max=255"`
	Latitude   *float64 `json:"latitude" binding:"omitempty"`
	Longitude  *float64 `json:"longitude" binding:"omitempty"`
	Timezone   string   `json:"timezone" binding:"omitempty,max=50"`
	IsActive   *bool    `json:"is_active"`
}

// UpdateCompanyRequest represents update company request
type UpdateCompanyRequest struct {
	Name       string  `json:"name" binding:"omitempty,min=2,max=200"`
	Address    string  `json:"address" binding:"omitempty,max=500"`
	Email      string  `json:"email" binding:"omitempty,email,max=100"`
	Phone      string  `json:"phone" binding:"omitempty,max=20"`
	NPWP       string  `json:"npwp" binding:"omitempty,max=30"`
	NIB        string  `json:"nib" binding:"omitempty,max=30"`
	ProvinceID *string `json:"province_id" binding:"omitempty,uuid"`
	CityID     *string `json:"city_id" binding:"omitempty,uuid"`
	DistrictID *string `json:"district_id" binding:"omitempty,uuid"`
	VillageID  *string `json:"village_id" binding:"omitempty,uuid"`
	VillageName *string `json:"village_name" binding:"omitempty,max=255"`
	Latitude   *float64 `json:"latitude" binding:"omitempty"`
	Longitude  *float64 `json:"longitude" binding:"omitempty"`
	Timezone   string   `json:"timezone" binding:"omitempty,max=50"`
	IsActive   *bool    `json:"is_active"`
}

// ListCompaniesRequest represents list companies request
type ListCompaniesRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PerPage   int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search    string `form:"search" binding:"omitempty,max=100"`
	Status    string `form:"status" binding:"omitempty,oneof=draft pending approved rejected"`
	IsActive  *bool  `form:"is_active" binding:"omitempty"`
	VillageID string `form:"village_id" binding:"omitempty,uuid"`
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=name status created_at updated_at"`
	SortDir   string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// ApproveCompanyRequest represents approve/reject company request
type ApproveCompanyRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

// SubmitForApprovalRequest represents submit for approval request
type SubmitForApprovalRequest struct {
	// No additional fields needed
}

// CompanyResponse represents company response
type CompanyResponse struct {
	ID         string                      `json:"id"`
	Name       string                      `json:"name"`
	Address    string                      `json:"address"`
	Email      string                      `json:"email"`
	Phone      string                      `json:"phone"`
	NPWP       string                      `json:"npwp"`
	NIB        string                      `json:"nib"`
	ProvinceID *string                     `json:"province_id"`
	Province   *geographicDto.ProvinceResponse `json:"province,omitempty"`
	CityID     *string                     `json:"city_id"`
	City       *geographicDto.CityResponse     `json:"city,omitempty"`
	DistrictID *string                     `json:"district_id"`
	District   *geographicDto.DistrictResponse `json:"district,omitempty"`
	VillageID  *string                     `json:"village_id"`
	VillageName *string                    `json:"village_name,omitempty"`
	Latitude   *float64                       `json:"latitude"`
	Longitude  *float64                       `json:"longitude"`
	Village    *geographicDto.VillageResponse `json:"village,omitempty"`
	Timezone   string                         `json:"timezone"`
	Status     string                      `json:"status"`
	IsApproved bool                        `json:"is_approved"`
	CreatedBy  *string                     `json:"created_by"`
	ApprovedBy *string                     `json:"approved_by"`
	ApprovedAt *string                     `json:"approved_at"`
	IsActive   bool                        `json:"is_active"`
	CreatedAt  string                      `json:"created_at"`
	UpdatedAt  string                      `json:"updated_at"`
	OutletCount int64                      `json:"outlet_count"`
}
