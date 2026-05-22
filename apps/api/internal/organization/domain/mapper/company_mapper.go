package mapper

import (
	"time"

	geographicDto "github.com/gilabs/gims/api/internal/geographic/domain/dto"
	geographicMapper "github.com/gilabs/gims/api/internal/geographic/domain/mapper"
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
)

// ToCompanyResponse converts Company model to CompanyResponse DTO
func ToCompanyResponse(m *models.Company) *dto.CompanyResponse {
	if m == nil {
		return nil
	}

	resp := &dto.CompanyResponse{
		ID:         m.ID,
		Name:       m.Name,
		Address:    m.Address,
		Email:      m.Email,
		Phone:      m.Phone,
		NPWP:       m.NPWP,
		NIB:        m.NIB,
		ProvinceID: m.ProvinceID,
		CityID:     m.CityID,
		DistrictID: m.DistrictID,
		VillageID:  m.VillageID,
		VillageName: m.VillageName,
		Latitude:   m.Latitude,
		Longitude:  m.Longitude,
		Timezone:   m.Timezone,
		Status:     string(m.Status),
		IsApproved: m.IsApproved,
		CreatedBy:  m.CreatedBy,
		ApprovedBy: m.ApprovedBy,
		IsActive:   m.IsActive,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  m.UpdatedAt.Format(time.RFC3339),
		OutletCount: m.OutletCount,
	}

	if m.ApprovedAt != nil {
		formatted := m.ApprovedAt.Format(time.RFC3339)
		resp.ApprovedAt = &formatted
	}

	// Map village if present
	if m.Village != nil {
		resp.Village = geographicMapper.ToVillageResponse(m.Village)
	}

	// Map other geographic relations if loaded directly
	if m.Province != nil {
		resp.Province = &geographicDto.ProvinceResponse{ID: m.Province.ID, Name: m.Province.Name}
	}
	if m.City != nil {
		resp.City = &geographicDto.CityResponse{ID: m.City.ID, Name: m.City.Name}
	}
	if m.District != nil {
		resp.District = &geographicDto.DistrictResponse{ID: m.District.ID, Name: m.District.Name}
	}

	return resp
}

// ToCompanyResponses converts slice of Company models to slice of CompanyResponse DTOs
func ToCompanyResponses(models []models.Company) []dto.CompanyResponse {
	responses := make([]dto.CompanyResponse, len(models))
	for i, m := range models {
		responses[i] = *ToCompanyResponse(&m)
	}
	return responses
}

// CompanyFromCreateRequest creates Company model from CreateCompanyRequest.
// Companies are immediately active — no approval workflow is required.
func CompanyFromCreateRequest(req *dto.CreateCompanyRequest, createdBy *string) *models.Company {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &models.Company{
		Name:       req.Name,
		Address:    req.Address,
		Email:      req.Email,
		Phone:      req.Phone,
		NPWP:       req.NPWP,
		NIB:        req.NIB,
		ProvinceID: req.ProvinceID,
		CityID:     req.CityID,
		DistrictID: req.DistrictID,
		VillageID:  req.VillageID,
		VillageName: req.VillageName,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Timezone:   req.Timezone,
		Status:     models.CompanyStatusApproved,
		IsApproved: true,
		CreatedBy:  createdBy,
		IsActive:   isActive,
	}
}
