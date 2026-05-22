package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/geographic/data/models"
	"github.com/gilabs/gims/api/internal/geographic/domain/dto"
)

// ToVillageResponse converts Village model to VillageResponse DTO
func ToVillageResponse(m *models.Village) *dto.VillageResponse {
	if m == nil {
		return nil
	}
	resp := &dto.VillageResponse{
		ID:         m.ID,
		DistrictID: m.DistrictID,
		Name:       m.Name,
		Code:       m.Code,
		PostalCode: m.PostalCode,
		Type:       m.Type,
		IsActive:   m.IsActive,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  m.UpdatedAt.Format(time.RFC3339),
	}
	if m.District != nil {
		resp.District = ToDistrictResponse(m.District)
	}
	return resp
}

// ToVillageResponses converts slice of Village models to slice of VillageResponse DTOs
func ToVillageResponses(models []models.Village) []dto.VillageResponse {
	responses := make([]dto.VillageResponse, len(models))
	for i, m := range models {
		responses[i] = *ToVillageResponse(&m)
	}
	return responses
}

// VillageFromCreateRequest creates Village model from CreateVillageRequest
func VillageFromCreateRequest(req *dto.CreateVillageRequest) *models.Village {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	villageType := "village"
	if req.Type != "" {
		villageType = req.Type
	}
	return &models.Village{
		DistrictID: req.DistrictID,
		Name:       req.Name,
		Code:       req.Code,
		PostalCode: req.PostalCode,
		Type:       villageType,
		IsActive:   isActive,
	}
}
