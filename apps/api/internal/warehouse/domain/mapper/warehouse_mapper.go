package mapper

import (
	geographicDto "github.com/gilabs/gims/api/internal/geographic/domain/dto"
	"github.com/gilabs/gims/api/internal/warehouse/data/models"
	"github.com/gilabs/gims/api/internal/warehouse/domain/dto"
)

// WarehouseMapper handles conversion between Warehouse model and DTOs
type WarehouseMapper struct{}

// NewWarehouseMapper creates a new WarehouseMapper
func NewWarehouseMapper() *WarehouseMapper {
	return &WarehouseMapper{}
}

// ToResponse converts a Warehouse model to WarehouseResponse DTO
func (m *WarehouseMapper) ToResponse(warehouse *models.Warehouse) *dto.WarehouseResponse {
	if warehouse == nil {
		return nil
	}

	response := &dto.WarehouseResponse{
		ID:          warehouse.ID,
		Code:        warehouse.Code,
		Name:        warehouse.Name,
		Description: warehouse.Description,
		Address:     warehouse.Address,
		ProvinceID:  warehouse.ProvinceID,
		CityID:      warehouse.CityID,
		DistrictID:  warehouse.DistrictID,
		VillageID:   warehouse.VillageID,
		VillageName: warehouse.VillageName,
		Latitude:    warehouse.Latitude,
		Longitude:   warehouse.Longitude,
		IsPosOutlet: warehouse.IsPosOutlet,
		OutletID:    warehouse.OutletID,
		IsActive:    warehouse.IsActive,
		HasStock:    warehouse.HasStock,
		CreatedAt:   warehouse.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   warehouse.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Map other geographic relations if loaded directly
	if warehouse.Province != nil {
		response.Province = &geographicDto.ProvinceResponse{ID: warehouse.Province.ID, Name: warehouse.Province.Name}
	}
	if warehouse.City != nil {
		response.City = &geographicDto.CityResponse{ID: warehouse.City.ID, Name: warehouse.City.Name}
	}
	if warehouse.District != nil {
		response.District = &geographicDto.DistrictResponse{ID: warehouse.District.ID, Name: warehouse.District.Name}
	}

	// Map nested village if present
	if warehouse.Village != nil {
		response.Village = &geographicDto.VillageResponse{
			ID:   warehouse.Village.ID,
			Name: warehouse.Village.Name,
		}

		// Map nested district
		if warehouse.Village.District != nil {
			response.Village.District = &geographicDto.DistrictResponse{
				ID:   warehouse.Village.District.ID,
				Name: warehouse.Village.District.Name,
			}

			// Map nested city
			if warehouse.Village.District.City != nil {
				response.Village.District.City = &geographicDto.CityResponse{
					ID:   warehouse.Village.District.City.ID,
					Name: warehouse.Village.District.City.Name,
				}

				// Map nested province
				if warehouse.Village.District.City.Province != nil {
					response.Village.District.City.Province = &geographicDto.ProvinceResponse{
						ID:   warehouse.Village.District.City.Province.ID,
						Name: warehouse.Village.District.City.Province.Name,
					}
				}
			}
		}
	}

	return response
}

// ToResponseList converts a list of Warehouse models to WarehouseResponse DTOs
func (m *WarehouseMapper) ToResponseList(warehouses []*models.Warehouse) []*dto.WarehouseResponse {
	responses := make([]*dto.WarehouseResponse, len(warehouses))
	for i, warehouse := range warehouses {
		responses[i] = m.ToResponse(warehouse)
	}
	return responses
}

// FromCreateRequest converts CreateWarehouseRequest to Warehouse model.
// Code is intentionally omitted here — the usecase assigns the auto-generated value.
func (m *WarehouseMapper) FromCreateRequest(req dto.CreateWarehouseRequest) *models.Warehouse {
	warehouse := &models.Warehouse{
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		ProvinceID:  req.ProvinceID,
		CityID:      req.CityID,
		DistrictID:  req.DistrictID,
		VillageID:   req.VillageID,
		VillageName: req.VillageName,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		IsActive:    true, // Default to active
	}

	if req.IsActive != nil {
		warehouse.IsActive = *req.IsActive
	}

	if req.IsPosOutlet != nil {
		warehouse.IsPosOutlet = *req.IsPosOutlet
	}

	warehouse.OutletID = req.OutletID

	return warehouse
}

// ApplyUpdateRequest applies UpdateWarehouseRequest to existing Warehouse model.
// Loaded geographic associations are explicitly nilled out so GORM's Save() writes the
// updated FK columns instead of reverting them to the preloaded association PKs.
func (m *WarehouseMapper) ApplyUpdateRequest(warehouse *models.Warehouse, req dto.UpdateWarehouseRequest) {
	// Nil out preloaded associations to prevent GORM from reverting FK columns
	warehouse.Province = nil
	warehouse.City = nil
	warehouse.District = nil
	warehouse.Village = nil

	if req.Code != nil {
		warehouse.Code = *req.Code
	}
	if req.Name != nil {
		warehouse.Name = *req.Name
	}
	if req.Description != nil {
		warehouse.Description = *req.Description
	}
	if req.Address != nil {
		warehouse.Address = *req.Address
	}

	// Keep partial updates safe: only update nullable relation fields when provided.
	if req.ProvinceID != nil {
		warehouse.ProvinceID = req.ProvinceID
	}
	if req.CityID != nil {
		warehouse.CityID = req.CityID
	}
	if req.DistrictID != nil {
		warehouse.DistrictID = req.DistrictID
	}
	if req.VillageID != nil {
		warehouse.VillageID = req.VillageID
	}
	if req.VillageName != nil {
		warehouse.VillageName = req.VillageName
	}

	if req.Latitude != nil {
		warehouse.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		warehouse.Longitude = req.Longitude
	}
	if req.IsActive != nil {
		warehouse.IsActive = *req.IsActive
	}
	if req.IsPosOutlet != nil {
		warehouse.IsPosOutlet = *req.IsPosOutlet
	}
	if req.OutletID != nil {
		warehouse.OutletID = req.OutletID
	}
}
