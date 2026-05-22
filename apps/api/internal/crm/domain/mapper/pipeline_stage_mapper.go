package mapper

import (
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToPipelineStageResponse converts a PipelineStage model to a response DTO
func ToPipelineStageResponse(m *models.PipelineStage) dto.PipelineStageResponse {
	return dto.PipelineStageResponse{
		ID:          m.ID,
		Name:        m.Name,
		Code:        m.Code,
		Order:       m.Order,
		Color:       m.Color,
		Probability: m.Probability,
		IsWon:       m.IsWon,
		IsLost:      m.IsLost,
		IsActive:    m.IsActive,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToPipelineStageResponseList converts a slice of PipelineStage models to response DTOs
func ToPipelineStageResponseList(stages []models.PipelineStage) []dto.PipelineStageResponse {
	responses := make([]dto.PipelineStageResponse, len(stages))
	for i, m := range stages {
		responses[i] = ToPipelineStageResponse(&m)
	}
	return responses
}
