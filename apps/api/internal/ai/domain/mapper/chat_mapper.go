package mapper

import (
	"encoding/json"
	"time"

	"github.com/gilabs/gims/api/internal/ai/data/models"
	"github.com/gilabs/gims/api/internal/ai/domain/dto"
	"github.com/gilabs/gims/api/internal/core/apptime"
)

// ChatMapper handles mapping between AI models and DTOs
type ChatMapper struct{}

// NewChatMapper creates a new ChatMapper
func NewChatMapper() *ChatMapper {
	return &ChatMapper{}
}

// ToMessageResponse converts a chat message model to a DTO response
func (m *ChatMapper) ToMessageResponse(msg *models.AIChatMessage) dto.MessageResponse {
	resp := dto.MessageResponse{
		ID:         msg.ID,
		Role:       string(msg.Role),
		Content:    msg.Content,
		DurationMs: msg.DurationMs,
		CreatedAt:  msg.CreatedAt.In(apptime.Location()).Format(time.RFC3339),
	}

	if msg.Intent != nil {
		resp.Intent = msg.Intent
	}

	return resp
}

// ToSessionListResponse converts a session model to a list response DTO
func (m *ChatMapper) ToSessionListResponse(session *models.AIChatSession) dto.SessionListResponse {
	resp := dto.SessionListResponse{
		ID:           session.ID,
		Title:        session.Title,
		Status:       string(session.Status),
		MessageCount: session.MessageCount,
		CreatedAt:    session.CreatedAt.In(apptime.Location()).Format(time.RFC3339),
	}
	if session.LastActivity != nil {
		resp.LastActivity = session.LastActivity.In(apptime.Location()).Format(time.RFC3339)
	}
	return resp
}

// ToSessionListResponses converts multiple sessions to list response DTOs
func (m *ChatMapper) ToSessionListResponses(sessions []models.AIChatSession) []dto.SessionListResponse {
	responses := make([]dto.SessionListResponse, 0, len(sessions))
	for i := range sessions {
		responses = append(responses, m.ToSessionListResponse(&sessions[i]))
	}
	return responses
}

// ToSessionDetailResponse converts a session with messages into a detail response
func (m *ChatMapper) ToSessionDetailResponse(session *models.AIChatSession) dto.SessionDetailResponse {
	messages := make([]dto.MessageResponse, 0, len(session.Messages))
	for i := range session.Messages {
		messages = append(messages, m.ToMessageResponse(&session.Messages[i]))
	}

	actions := make([]dto.ActionLogResponse, 0, len(session.Actions))
	for i := range session.Actions {
		actions = append(actions, m.ToActionLogResponse(&session.Actions[i]))
	}

	resp := dto.SessionDetailResponse{
		ID:           session.ID,
		Title:        session.Title,
		Status:       string(session.Status),
		MessageCount: session.MessageCount,
		Messages:     messages,
		Actions:      actions,
		CreatedAt:    session.CreatedAt.In(apptime.Location()).Format(time.RFC3339),
	}

	if session.LastActivity != nil {
		resp.LastActivity = session.LastActivity.In(apptime.Location()).Format(time.RFC3339)
	}

	if resp.MessageCount < len(messages) {
		resp.MessageCount = len(messages)
	}

	return resp
}

// ToActionLogResponse converts an action log model to a DTO response
func (m *ChatMapper) ToActionLogResponse(log *models.AIActionLog) dto.ActionLogResponse {
	resp := dto.ActionLogResponse{
		ID:             log.ID,
		Intent:         log.Intent,
		Action:         string(log.Action),
		EntityType:     log.EntityType,
		Status:         string(log.Status),
		ErrorMessage:   log.ErrorMessage,
		PermissionUsed: log.PermissionUsed,
		DurationMs:     log.DurationMs,
		CreatedAt:      log.CreatedAt.In(apptime.Location()).Format(time.RFC3339),
	}

	if log.EntityID != nil {
		resp.EntityID = *log.EntityID
	}

	if log.RequestPayload != nil {
		var payload interface{}
		if err := json.Unmarshal([]byte(*log.RequestPayload), &payload); err == nil {
			resp.RequestPayload = payload
		}
	}
	if log.ResponsePayload != nil {
		var payload interface{}
		if err := json.Unmarshal([]byte(*log.ResponsePayload), &payload); err == nil {
			resp.ResponsePayload = payload
		}
	}

	return resp
}

// ToActionLogResponses converts multiple action logs to response DTOs
func (m *ChatMapper) ToActionLogResponses(logs []models.AIActionLog) []dto.ActionLogResponse {
	responses := make([]dto.ActionLogResponse, 0, len(logs))
	for i := range logs {
		responses = append(responses, m.ToActionLogResponse(&logs[i]))
	}
	return responses
}

// ToActionPreview converts an action log to a preview DTO
func (m *ChatMapper) ToActionPreview(log *models.AIActionLog) *dto.ActionPreview {
	preview := &dto.ActionPreview{
		ID:     log.ID,
		Intent: log.Intent,
		Status: string(log.Status),
	}

	if log.EntityType != "" {
		preview.EntityType = log.EntityType
	}
	if log.EntityID != nil {
		preview.EntityID = *log.EntityID
	}
	if log.RequestPayload != nil {
		var payload interface{}
		if err := json.Unmarshal([]byte(*log.RequestPayload), &payload); err == nil {
			preview.PayloadPreview = payload
		}
	}
	preview.DurationMs = log.DurationMs

	return preview
}

// ToIntentRegistryResponse converts an intent registry model to a DTO
func (m *ChatMapper) ToIntentRegistryResponse(intent *models.AIIntentRegistry) dto.IntentRegistryResponse {
	return dto.IntentRegistryResponse{
		ID:                   intent.ID,
		IntentCode:           intent.IntentCode,
		DisplayName:          intent.DisplayName,
		Description:          intent.Description,
		Module:               intent.Module,
		ActionType:           intent.ActionType,
		RequiredPermission:   intent.RequiredPermission,
		RequiresConfirmation: intent.RequiresConfirmation,
		IsActive:             intent.IsActive,
	}
}

// ToIntentRegistryResponses converts multiple intents
func (m *ChatMapper) ToIntentRegistryResponses(intents []models.AIIntentRegistry) []dto.IntentRegistryResponse {
	responses := make([]dto.IntentRegistryResponse, 0, len(intents))
	for i := range intents {
		responses = append(responses, m.ToIntentRegistryResponse(&intents[i]))
	}
	return responses
}
