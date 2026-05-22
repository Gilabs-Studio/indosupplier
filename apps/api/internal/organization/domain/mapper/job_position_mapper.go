package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
)

// ToJobPositionResponse converts JobPosition model to JobPositionResponse DTO
func ToJobPositionResponse(m *models.JobPosition) *dto.JobPositionResponse {
	if m == nil {
		return nil
	}
	return &dto.JobPositionResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
	}
}

// ToJobPositionResponses converts slice of JobPosition models to slice of JobPositionResponse DTOs
func ToJobPositionResponses(models []models.JobPosition) []dto.JobPositionResponse {
	responses := make([]dto.JobPositionResponse, len(models))
	for i, m := range models {
		responses[i] = *ToJobPositionResponse(&m)
	}
	return responses
}

// JobPositionFromCreateRequest creates JobPosition model from CreateJobPositionRequest
func JobPositionFromCreateRequest(req *dto.CreateJobPositionRequest) *models.JobPosition {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &models.JobPosition{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	}
}
