package mapper

import (
	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
)

// ToLeaveTypeResponse converts LeaveType model to response DTO
func ToLeaveTypeResponse(m *models.LeaveType) dto.LeaveTypeResponse {
	return dto.LeaveTypeResponse{
		ID:          m.ID,
		Code:        m.Code,
		Name:        m.Name,
		Description: m.Description,
		MaxDays:     m.MaxDays,
		IsPaid:      m.IsPaid,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToLeaveTypeResponseList converts a slice of LeaveType models to response DTOs
func ToLeaveTypeResponseList(models []models.LeaveType) []dto.LeaveTypeResponse {
	responses := make([]dto.LeaveTypeResponse, len(models))
	for i, m := range models {
		responses[i] = ToLeaveTypeResponse(&m)
	}
	return responses
}
