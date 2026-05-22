package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToAreaCaptureResponse converts AreaCapture model to AreaCaptureResponse DTO
func ToAreaCaptureResponse(m *models.AreaCapture) *dto.AreaCaptureResponse {
	if m == nil {
		return nil
	}

	return &dto.AreaCaptureResponse{
		ID:            m.ID,
		VisitReportID: m.VisitReportID,
		CaptureType:   m.CaptureType,
		Latitude:      m.Latitude,
		Longitude:     m.Longitude,
		Address:       m.Address,
		Accuracy:      m.Accuracy,
		AreaID:        m.AreaID,
		CapturedAt:    m.CapturedAt.Format(time.RFC3339),
		CapturedBy:    m.CapturedBy,
		CreatedAt:     m.CreatedAt.Format(time.RFC3339),
	}
}

// ToAreaCaptureResponses converts slice of AreaCapture models to slice of DTOs
func ToAreaCaptureResponses(captures []models.AreaCapture) []dto.AreaCaptureResponse {
	responses := make([]dto.AreaCaptureResponse, len(captures))
	for i, c := range captures {
		responses[i] = *ToAreaCaptureResponse(&c)
	}
	return responses
}

// AreaCaptureFromCreateRequest creates AreaCapture model from CreateAreaCaptureRequest
func AreaCaptureFromCreateRequest(req *dto.CreateAreaCaptureRequest) *models.AreaCapture {
	capturedAt := apptime.Now()
	if req.CapturedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, req.CapturedAt); err == nil {
			capturedAt = parsed
		}
	}

	return &models.AreaCapture{
		VisitReportID: req.VisitReportID,
		CaptureType:   req.CaptureType,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		Address:       req.Address,
		Accuracy:      req.Accuracy,
		AreaID:        req.AreaID,
		CapturedAt:    capturedAt,
		CapturedBy:    req.CapturedBy,
	}
}
