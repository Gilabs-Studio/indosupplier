package mapper

import (
	geographicDto "github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
)

// OutletMapper handles conversion between Outlet model and DTOs
type OutletMapper struct{}

// NewOutletMapper creates a new OutletMapper
func NewOutletMapper() *OutletMapper {
	return &OutletMapper{}
}

// ToResponse converts an Outlet model to OutletResponse DTO
func (m *OutletMapper) ToResponse(outlet *models.Outlet) *dto.OutletResponse {
	if outlet == nil {
		return nil
	}

	resp := &dto.OutletResponse{
		ID:          outlet.ID,
		TenantID:    outlet.TenantID,
		Code:        outlet.Code,
		Name:        outlet.Name,
		Description: outlet.Description,
		Phone:       outlet.Phone,
		Email:       outlet.Email,
		Address:     outlet.Address,
		ProvinceID:  outlet.ProvinceID,
		CityID:      outlet.CityID,
		DistrictID:  outlet.DistrictID,
		VillageID:   outlet.VillageID,
		Latitude:    outlet.Latitude,
		Longitude:   outlet.Longitude,
		ManagerID:   outlet.ManagerID,
		CompanyID:   outlet.CompanyID,
		WarehouseID: outlet.WarehouseID,
		IsActive:    outlet.IsActive,
		CreatedAt:   outlet.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   outlet.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Map geographic relations
	if outlet.Province != nil {
		resp.Province = &geographicDto.ProvinceResponse{ID: outlet.Province.ID, Name: outlet.Province.Name}
	}
	if outlet.City != nil {
		resp.City = &geographicDto.CityResponse{ID: outlet.City.ID, Name: outlet.City.Name}
	}
	if outlet.District != nil {
		resp.District = &geographicDto.DistrictResponse{ID: outlet.District.ID, Name: outlet.District.Name}
	}

	// Map nested village hierarchy
	if outlet.Village != nil {
		resp.Village = &geographicDto.VillageResponse{
			ID:   outlet.Village.ID,
			Name: outlet.Village.Name,
		}
		if outlet.Village.District != nil {
			resp.Village.District = &geographicDto.DistrictResponse{
				ID:   outlet.Village.District.ID,
				Name: outlet.Village.District.Name,
			}
			if outlet.Village.District.City != nil {
				resp.Village.District.City = &geographicDto.CityResponse{
					ID:   outlet.Village.District.City.ID,
					Name: outlet.Village.District.City.Name,
				}
				if outlet.Village.District.City.Province != nil {
					resp.Village.District.City.Province = &geographicDto.ProvinceResponse{
						ID:   outlet.Village.District.City.Province.ID,
						Name: outlet.Village.District.City.Province.Name,
					}
				}
			}
		}
	}

	// Map manager
	if outlet.Manager != nil {
		resp.Manager = &dto.ManagerResponse{
			ID:           outlet.Manager.ID,
			EmployeeCode: outlet.Manager.EmployeeCode,
			Name:         outlet.Manager.Name,
		}
	}

	// Map company
	if outlet.Company != nil {
		resp.Company = &dto.CompanySimpleResponse{
			ID:   outlet.Company.ID,
			Name: outlet.Company.Name,
		}
	}

	return resp
}

// ToResponseList converts a list of Outlet models to OutletResponse DTOs
func (m *OutletMapper) ToResponseList(outlets []*models.Outlet) []*dto.OutletResponse {
	responses := make([]*dto.OutletResponse, len(outlets))
	for i, outlet := range outlets {
		responses[i] = m.ToResponse(outlet)
	}
	return responses
}

// FromCreateRequest converts CreateOutletRequest to Outlet model
func (m *OutletMapper) FromCreateRequest(req dto.CreateOutletRequest) *models.Outlet {
	outlet := &models.Outlet{
		Name:        req.Name,
		Description: req.Description,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		ProvinceID:  req.ProvinceID,
		CityID:      req.CityID,
		DistrictID:  req.DistrictID,
		VillageID:   req.VillageID,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		ManagerID:   req.ManagerID,
		CompanyID:   req.CompanyID,
		IsActive:    true,
	}

	if req.IsActive != nil {
		outlet.IsActive = *req.IsActive
	}

	return outlet
}

// ApplyUpdateRequest applies UpdateOutletRequest to an existing Outlet model
func (m *OutletMapper) ApplyUpdateRequest(outlet *models.Outlet, req dto.UpdateOutletRequest) {
	// Nil out preloaded associations to prevent GORM from reverting FK columns
	outlet.Province = nil
	outlet.City = nil
	outlet.District = nil
	outlet.Village = nil
	outlet.Manager = nil
	outlet.Company = nil

	if req.Name != nil {
		outlet.Name = *req.Name
	}
	if req.Description != nil {
		outlet.Description = *req.Description
	}
	if req.Phone != nil {
		outlet.Phone = *req.Phone
	}
	if req.Email != nil {
		outlet.Email = *req.Email
	}
	if req.Address != nil {
		outlet.Address = *req.Address
	}

	// Geographic FK fields
	if req.ProvinceID != nil {
		outlet.ProvinceID = req.ProvinceID
	}
	if req.CityID != nil {
		outlet.CityID = req.CityID
	}
	if req.DistrictID != nil {
		outlet.DistrictID = req.DistrictID
	}
	if req.VillageID != nil {
		outlet.VillageID = req.VillageID
	}

	if req.Latitude != nil {
		outlet.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		outlet.Longitude = req.Longitude
	}
	if req.ManagerID != nil {
		outlet.ManagerID = req.ManagerID
	}
	if req.CompanyID != nil {
		outlet.CompanyID = req.CompanyID
	}

	if req.IsActive != nil {
		outlet.IsActive = *req.IsActive
	}
}
