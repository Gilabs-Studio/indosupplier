package mapper

import (
	"encoding/json"

	"github.com/gilabs/gims/api/internal/feedback/data/models"
	"github.com/gilabs/gims/api/internal/feedback/domain/dto"
)

// ToFeedbackFormResponse converts a FeedbackForm model to its DTO.
func ToFeedbackFormResponse(m *models.FeedbackForm) dto.FeedbackFormResponse {
	return dto.FeedbackFormResponse{
		ID:          m.ID,
		OutletID:    m.OutletID,
		Title:       m.Title,
		Description: m.Description,
		SchemaJSON:  json.RawMessage(m.SchemaJSON),
		Version:     m.Version,
		IsActive:    m.IsActive,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToFeedbackFormResponseList converts a slice of models.
func ToFeedbackFormResponseList(forms []models.FeedbackForm) []dto.FeedbackFormResponse {
	result := make([]dto.FeedbackFormResponse, 0, len(forms))
	for i := range forms {
		result = append(result, ToFeedbackFormResponse(&forms[i]))
	}
	return result
}

// ToFeedbackResponseItem converts a FeedbackResponse model to its admin list DTO.
func ToFeedbackResponseItem(m *models.FeedbackResponse) dto.FeedbackResponseItem {
	formTitle := ""
	var schemaJSON json.RawMessage
	if m.Form != nil {
		formTitle = m.Form.Title
		schemaJSON = json.RawMessage(m.Form.SchemaJSON)
	}
	return dto.FeedbackResponseItem{
		ID:           m.ID,
		FormID:       m.FormID,
		FormTitle:    formTitle,
		SchemaJSON:   schemaJSON,
		OutletID:     m.OutletID,
		SalesOrderID: m.SalesOrderID,
		PosOrderID:   m.PosOrderID,
		CustomerName: m.CustomerName,
		Answers:      json.RawMessage(m.Answers),
		AvgScore:     m.AvgScore,
		SubmittedAt:  m.SubmittedAt,
	}
}

// ToFeedbackResponseItemList converts a slice of response models.
func ToFeedbackResponseItemList(responses []models.FeedbackResponse) []dto.FeedbackResponseItem {
	result := make([]dto.FeedbackResponseItem, 0, len(responses))
	for i := range responses {
		result = append(result, ToFeedbackResponseItem(&responses[i]))
	}
	return result
}
