// Package usecase contains the rewritten AI chat business logic using
// Claude Code-inspired context engineering: modular system prompt, tool registry,
// multi-turn conversation engine with tool loop, and SSE streaming.
//
// This file (chat_engine_usecase.go) replaces the core message processing flow
// while maintaining backward compatibility with existing endpoints.
package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/ai/data/models"
	"github.com/gilabs/gims/api/internal/ai/data/repositories"
	"github.com/gilabs/gims/api/internal/ai/domain/dto"
	aiContext "github.com/gilabs/gims/api/internal/ai/domain/usecase/context"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase/engine"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase/tools"
	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/cerebras"
	"github.com/gilabs/gims/api/internal/core/utils"
)

// ChatEngineUsecase implements the new AI chat interface using the conversation engine.
type ChatEngineUsecase struct {
	sessionRepo      repositories.ChatSessionRepository
	messageRepo      repositories.ChatMessageRepository
	actionRepo       repositories.ActionLogRepository
	cerebrasClient   *cerebras.Client
	contextBuilder   *aiContext.Builder
	engine           *engine.Engine
	toolRegistry     *tools.Registry
	entityResolver   *EntityResolver
	requestValidator *RequestValidator
}

// NewChatEngineUsecase creates the engine-based usecase.
func NewChatEngineUsecase(
	sessionRepo repositories.ChatSessionRepository,
	messageRepo repositories.ChatMessageRepository,
	actionRepo repositories.ActionLogRepository,
	cerebrasClient *cerebras.Client,
	toolRegistry *tools.Registry,
	contextBuilder *aiContext.Builder,
	entityResolver *EntityResolver,
) *ChatEngineUsecase {
	var requestValidator *RequestValidator
	if entityResolver != nil && entityResolver.db != nil {
		requestValidator = NewRequestValidator(entityResolver.db, entityResolver)
	}

	return &ChatEngineUsecase{
		sessionRepo:      sessionRepo,
		messageRepo:      messageRepo,
		actionRepo:       actionRepo,
		cerebrasClient:   cerebrasClient,
		contextBuilder:   contextBuilder,
		engine:           engine.NewEngine(cerebrasClient, toolRegistry, contextBuilder),
		toolRegistry:     toolRegistry,
		entityResolver:   entityResolver,
		requestValidator: requestValidator,
	}
}

// SendMessage processes a user message through the conversation engine.
func (u *ChatEngineUsecase) SendMessage(ctx context.Context, req *dto.SendMessageRequest, userID string, userPermissions map[string]bool, isAdmin bool) (*dto.ChatResponse, error) {
	start := apptime.Now()

	if !u.cerebrasClient.IsConfigured() {
		return nil, fmt.Errorf("AI_SERVICE_NOT_CONFIGURED: Cerebras AI is not configured")
	}

	// Get or create session
	session, err := u.getOrCreateSession(ctx, req.SessionID, userID)
	if err != nil {
		return nil, err
	}

	// IDOR prevention
	if session.UserID != userID {
		return nil, fmt.Errorf("FORBIDDEN: you do not have access to this session")
	}

	// Check for pending action confirmation/cancellation
	if req.SessionID != nil && *req.SessionID != "" {
		pendingAction, pendingErr := u.actionRepo.FindPendingBySessionID(ctx, session.ID)
		hasPendingAction := pendingErr == nil && pendingAction != nil
		if hasPendingAction {
			if isAffirmativeMessage(req.Message) {
				return u.ConfirmAction(ctx, &dto.ConfirmActionRequest{
					ActionID:  pendingAction.ID,
					Confirmed: true,
				}, userID, userPermissions, isAdmin)
			}
			if isNegativeMessage(req.Message) {
				return u.ConfirmAction(ctx, &dto.ConfirmActionRequest{
					ActionID:  pendingAction.ID,
					Confirmed: false,
				}, userID, userPermissions, isAdmin)
			}
			// Cancel stale pending action and process as new request
			pendingAction.Status = models.ActionStatusCancelled
			_ = u.actionRepo.Update(ctx, pendingAction)
		}

		if !hasPendingAction && isStandaloneConfirmationReply(req.Message) {
			return u.handleOrphanConfirmationSync(ctx, session, req.Message, req.Model, userID, userPermissions, isAdmin, start)
		}
	}

	// Save user message
	userMsg := &models.AIChatMessage{
		SessionID: session.ID,
		Role:      models.MessageRoleUser,
		Content:   req.Message,
	}
	if err := u.messageRepo.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to save user message: %w", err)
	}

	// Load conversation history
	history, err := u.messageRepo.FindBySessionID(ctx, session.ID, engine.MaxContextMessages)
	if err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to load conversation history: %w", err)
	}

	// Convert to engine message format
	engineHistory := u.toEngineMessages(history)

	// Build user context for the context builder
	userCtx := u.buildUserContext(userID, userPermissions, isAdmin, session.ID)

	// Load recent actions for context continuity
	recentActions, _ := u.actionRepo.FindBySessionID(ctx, session.ID)
	// Limit to last 5 for context
	if len(recentActions) > 5 {
		recentActions = recentActions[len(recentActions)-5:]
	}
	if len(recentActions) > 0 {
		for _, a := range recentActions {
			userCtx.RecentActions = append(userCtx.RecentActions, aiContext.RecentAction{
				Intent:    a.Intent,
				Module:    a.EntityType,
				Status:    string(a.Status),
				CreatedAt: a.CreatedAt,
			})
		}
	}

	// Select model
	model := u.cerebrasClient.GetDefaultModel()
	if req.Model != "" {
		model = req.Model
	}

	// Process through the conversation engine
	result, err := u.engine.ProcessMessage(ctx, req.Message, engineHistory, userCtx, model)
	if err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: %w", err)
	}

	// Handle tool calls requiring confirmation
	if result.RequiresConfirmation && result.PendingToolCall != nil {
		return u.createPendingToolAction(ctx, session, result, userID, start)
	}

	// Log executed tool calls
	for _, tcr := range result.ToolCallResults {
		u.logToolCall(ctx, session.ID, userID, tcr)
	}

	// Save assistant response
	return u.saveEngineResponse(ctx, session, result, start)
}

// SendMessageStream processes a message with SSE streaming.
func (u *ChatEngineUsecase) SendMessageStream(ctx context.Context, req *dto.SendMessageRequest, userID string, userPermissions map[string]bool, isAdmin bool, eventChan chan<- tools.StreamEvent) error {
	if !u.cerebrasClient.IsConfigured() {
		eventChan <- tools.StreamEvent{
			Type:    tools.EventError,
			Content: "AI service is not configured",
		}
		return fmt.Errorf("AI_SERVICE_NOT_CONFIGURED")
	}

	session, err := u.getOrCreateSession(ctx, req.SessionID, userID)
	if err != nil {
		return err
	}
	if session.UserID != userID {
		return fmt.Errorf("FORBIDDEN: you do not have access to this session")
	}

	// Intercept pending action confirmation/cancellation — mirrors the sync SendMessage path.
	// Without this check, user replies like "ya"/"yes" would re-enter the engine and trigger
	// a new NeedsConfirmation loop instead of executing the already-queued action.
	if req.SessionID != nil && *req.SessionID != "" {
		pendingAction, pendingErr := u.actionRepo.FindPendingBySessionID(ctx, session.ID)
		hasPendingAction := pendingErr == nil && pendingAction != nil
		if hasPendingAction {
			model := u.cerebrasClient.GetDefaultModel()
			if req.Model != "" {
				model = req.Model
			}
			eventChan <- tools.StreamEvent{
				Type: tools.EventMessageStart,
				Data: map[string]string{"session_id": session.ID},
			}
			if isAffirmativeMessage(req.Message) {
				return u.confirmActionStream(ctx, pendingAction, session, userID, userPermissions, isAdmin, req.Message, model, eventChan)
			}
			if isNegativeMessage(req.Message) {
				return u.cancelActionStream(ctx, pendingAction, session, req.Message, eventChan)
			}
			// Unrecognised reply — cancel stale action and continue as a new request
			pendingAction.Status = models.ActionStatusCancelled
			_ = u.actionRepo.Update(ctx, pendingAction)
		}

		if !hasPendingAction && isStandaloneConfirmationReply(req.Message) {
			eventChan <- tools.StreamEvent{
				Type: tools.EventMessageStart,
				Data: map[string]string{"session_id": session.ID},
			}
			return u.handleOrphanConfirmationStream(ctx, session, req.Message, req.Model, userID, userPermissions, isAdmin, eventChan)
		}
	}

	// Save user message
	userMsg := &models.AIChatMessage{
		SessionID: session.ID,
		Role:      models.MessageRoleUser,
		Content:   req.Message,
	}
	if err := u.messageRepo.Create(ctx, userMsg); err != nil {
		return fmt.Errorf("AI_CHAT_FAILED: failed to save user message: %w", err)
	}

	// Load history and context
	history, _ := u.messageRepo.FindBySessionID(ctx, session.ID, engine.MaxContextMessages)
	engineHistory := u.toEngineMessages(history)
	userCtx := u.buildUserContext(userID, userPermissions, isAdmin, session.ID)

	model := u.cerebrasClient.GetDefaultModel()
	if req.Model != "" {
		model = req.Model
	}

	// Send session_id as first event
	eventChan <- tools.StreamEvent{
		Type: tools.EventMessageStart,
		Data: map[string]string{"session_id": session.ID},
	}

	// Process with streaming — now returns the accumulated result for DB persistence
	start := apptime.Now()
	result, processErr := u.engine.ProcessMessageStream(ctx, req.Message, engineHistory, userCtx, model, eventChan)
	if processErr != nil {
		if isModelNotAvailableError(processErr) {
			fallbackModel := u.pickAlternativeModel(model)
			if fallbackModel != "" {
				log.Printf(
					"[AI] selected model unavailable, retrying with fallback model session_id=%s user_id=%s from=%s to=%s err=%v",
					session.ID,
					userID,
					model,
					fallbackModel,
					processErr,
				)
				model = fallbackModel
				result, processErr = u.engine.ProcessMessageStream(ctx, req.Message, engineHistory, userCtx, model, eventChan)
			}
		}

		if processErr == nil {
			// Continue with normal persistence path below if retry succeeded.
			goto persistResult
		}

		log.Printf(
			"[AI] stream failed, attempting fallback session_id=%s user_id=%s model=%s err=%v",
			session.ID,
			userID,
			model,
			processErr,
		)

		fallbackResult, fallbackErr := u.engine.ProcessMessage(ctx, req.Message, engineHistory, userCtx, model)
		if fallbackErr != nil {
			log.Printf(
				"[AI] fallback failed session_id=%s user_id=%s model=%s stream_err=%v fallback_err=%v",
				session.ID,
				userID,
				model,
				processErr,
				fallbackErr,
			)
			interruptionMessage := buildStreamInterruptionMessage(processErr, model)
			if _, saveErr := u.saveSimpleResponse(ctx, session, interruptionMessage, start); saveErr != nil {
				log.Printf("[AI] warning: failed to save interruption response: %v", saveErr)
			}
			return processErr
		}

		log.Printf(
			"[AI] fallback succeeded session_id=%s user_id=%s model=%s",
			session.ID,
			userID,
			model,
		)

		if fallbackResult != nil && fallbackResult.Response != "" {
			eventChan <- tools.StreamEvent{Type: tools.EventContentDelta, Content: fallbackResult.Response}
		}

		durationMs := time.Since(start).Milliseconds()
		if fallbackResult != nil {
			if fallbackResult.RequiresConfirmation && fallbackResult.PendingToolCall != nil {
				eventChan <- tools.StreamEvent{
					Type: tools.EventMessageEnd,
					Data: map[string]interface{}{
						"requires_confirmation": true,
						"pending_tool_call":     fallbackResult.PendingToolCall,
						"duration_ms":           durationMs,
						"turn_count":            fallbackResult.TurnCount,
						"fallback":              true,
					},
				}
				if _, saveErr := u.createPendingToolAction(ctx, session, fallbackResult, userID, start); saveErr != nil {
					log.Printf("[AI] warning: failed to create pending tool action after fallback: %v", saveErr)
				}
				return nil
			}

			eventChan <- tools.StreamEvent{
				Type: tools.EventMessageEnd,
				Data: map[string]interface{}{
					"duration_ms": durationMs,
					"turn_count":  fallbackResult.TurnCount,
					"fallback":    true,
				},
			}

			if fallbackResult.Response != "" {
				if _, saveErr := u.saveEngineResponse(ctx, session, fallbackResult, start); saveErr != nil {
					log.Printf("[AI] warning: failed to save fallback response: %v", saveErr)
				}
			}
			return nil
		}

		return processErr
	}

persistResult:

	// Persist the assistant response so it appears in subsequent session reads
	if result != nil {
		if result.RequiresConfirmation && result.PendingToolCall != nil {
			// Create an action log record so the next "ya"/"yes" can be intercepted above
			if _, saveErr := u.createPendingToolAction(ctx, session, result, userID, start); saveErr != nil {
				log.Printf("[AI] warning: failed to create pending tool action: %v", saveErr)
			}
		} else if result.Response != "" {
			if _, saveErr := u.saveEngineResponse(ctx, session, result, start); saveErr != nil {
				log.Printf("[AI] warning: failed to save engine response: %v", saveErr)
			}
		}
	}

	return nil
}

func isModelNotAvailableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "model_not_found") ||
		strings.Contains(errMsg, "does not exist or you do not have access to it")
}

func (u *ChatEngineUsecase) pickAlternativeModel(currentModel string) string {
	available := u.cerebrasClient.AvailableModels()
	for _, model := range available {
		if strings.TrimSpace(model.ID) == "" {
			continue
		}
		if model.ID != currentModel {
			return model.ID
		}
	}
	return ""
}

func buildStreamInterruptionMessage(streamErr error, model string) string {
	if isModelNotAvailableError(streamErr) {
		return fmt.Sprintf("Model '%s' tidak tersedia. Silakan pilih model lain lalu kirim ulang permintaan Anda.", model)
	}

	if streamErr != nil {
		if strings.Contains(streamErr.Error(), "context_length_exceeded") {
			return "Respons terhenti karena konteks percakapan terlalu panjang. Silakan mulai chat baru atau ringkas permintaan Anda."
		}
	}

	return "Respons sebelumnya terputus sebelum selesai. Silakan kirim ulang permintaan terakhir Anda agar saya lanjutkan."
}

// ConfirmAction processes user confirmation of a pending tool action.
func (u *ChatEngineUsecase) ConfirmAction(ctx context.Context, req *dto.ConfirmActionRequest, userID string, userPermissions map[string]bool, isAdmin bool) (*dto.ChatResponse, error) {
	start := apptime.Now()

	action, err := u.actionRepo.FindByID(ctx, req.ActionID)
	if err != nil {
		return nil, fmt.Errorf("AI_ACTION_NOT_FOUND: action not found: %w", err)
	}

	if action.UserID != userID {
		return nil, fmt.Errorf("FORBIDDEN: you do not have access to this action")
	}
	if action.Status != models.ActionStatusPendingConfirmation {
		return nil, fmt.Errorf("AI_ACTION_INVALID_STATE: action is not pending confirmation (current: %s)", action.Status)
	}

	session, err := u.sessionRepo.FindByID(ctx, action.SessionID)
	if err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: session not found: %w", err)
	}

	if !req.Confirmed {
		action.Status = models.ActionStatusCancelled
		_ = u.actionRepo.Update(ctx, action)
		return u.saveSimpleResponse(ctx, session, "Tindakan dibatalkan. Ada yang bisa saya bantu lagi?", start)
	}

	// Execute the confirmed tool
	tool := u.toolRegistry.Get(strings.ToLower(action.Intent))
	if tool == nil {
		action.Status = models.ActionStatusFailed
		action.ErrorMessage = "Tool no longer available"
		_ = u.actionRepo.Update(ctx, action)
		return nil, fmt.Errorf("AI_CHAT_FAILED: tool '%s' not found in registry", action.Intent)
	}

	var storedParams map[string]interface{}
	if action.RequestPayload != nil {
		_ = json.Unmarshal([]byte(*action.RequestPayload), &storedParams)
	}

	if validationMsg := u.validateActionPayloadAgainstFacts(ctx, session.ID, action.Intent, string(action.Action), storedParams); validationMsg != "" {
		action.Status = models.ActionStatusCancelled
		action.ErrorMessage = validationMsg
		_ = u.actionRepo.Update(ctx, action)
		return u.saveSimpleResponse(ctx, session, validationMsg, start)
	}

	execCtx := &tools.ExecutionContext{
		UserID:          userID,
		UserPermissions: userPermissions,
		IsAdmin:         isAdmin,
		SessionID:       session.ID,
	}

	toolResult, execErr := tool.Execute(ctx, storedParams, execCtx)

	if execErr != nil || (toolResult != nil && !toolResult.Success) {
		action.Status = models.ActionStatusFailed
		if execErr != nil {
			action.ErrorMessage = execErr.Error()
		} else if toolResult != nil {
			action.ErrorMessage = toolResult.ErrorMessage
		}
		_ = u.actionRepo.Update(ctx, action)

		errMsg := "Tindakan gagal."
		if toolResult != nil && toolResult.ErrorMessage != "" {
			errMsg = fmt.Sprintf("Maaf, tindakan gagal: %s", toolResult.ErrorMessage)
		}
		return u.saveSimpleResponse(ctx, session, errMsg, start)
	}

	// Success
	action.Status = models.ActionStatusSuccess
	if toolResult.EntityID != "" {
		action.EntityID = &toolResult.EntityID
	}
	responseJSON, _ := json.Marshal(toolResult.Data)
	respStr := string(responseJSON)
	action.ResponsePayload = &respStr
	action.DurationMs = int(toolResult.DurationMs)
	_ = u.actionRepo.Update(ctx, action)

	assistantContent := fmt.Sprintf("Tindakan berhasil dilakukan. %s", toolResult.Message)
	return u.saveSimpleResponse(ctx, session, assistantContent, start)
}

// ListSessions returns paginated sessions
func (u *ChatEngineUsecase) ListSessions(ctx context.Context, req *dto.ListSessionsRequest, userID string) ([]dto.SessionListResponse, *utils.PaginationResult, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}

	sessions, total, err := u.sessionRepo.FindByUserID(ctx, userID, req.Page, req.PerPage, req.Status, req.Search)
	if err != nil {
		return nil, nil, fmt.Errorf("AI_CHAT_FAILED: failed to list sessions: %w", err)
	}

	responses := make([]dto.SessionListResponse, len(sessions))
	for i, s := range sessions {
		responses[i] = dto.SessionListResponse{
			ID:           s.ID,
			Title:        s.Title,
			Status:       string(s.Status),
			LastActivity: formatTimePtr(s.LastActivity),
			MessageCount: s.MessageCount,
			CreatedAt:    s.CreatedAt.Format(time.RFC3339),
		}
	}

	pagination := &utils.PaginationResult{
		Page:       req.Page,
		PerPage:    req.PerPage,
		Total:      int(total),
		TotalPages: int((total + int64(req.PerPage) - 1) / int64(req.PerPage)),
	}

	return responses, pagination, nil
}

// GetSessionDetail returns a session with messages
func (u *ChatEngineUsecase) GetSessionDetail(ctx context.Context, sessionID string, userID string) (*dto.SessionDetailResponse, error) {
	session, err := u.sessionRepo.FindByIDWithMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("AI_SESSION_NOT_FOUND: session not found: %w", err)
	}

	if session.UserID != userID {
		return nil, fmt.Errorf("FORBIDDEN: you do not have access to this session")
	}

	messages := make([]dto.MessageResponse, len(session.Messages))
	for i, m := range session.Messages {
		messages[i] = dto.MessageResponse{
			ID:         m.ID,
			Role:       string(m.Role),
			Content:    m.Content,
			Intent:     m.Intent,
			Model:      m.Model,
			DurationMs: m.DurationMs,
			CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		}
	}

	// Check for pending action
	var pendingAction *dto.ActionPreview
	if pending, err := u.actionRepo.FindPendingBySessionID(ctx, session.ID); err == nil && pending != nil {
		pendingAction = &dto.ActionPreview{
			ID:     pending.ID,
			Intent: pending.Intent,
			Status: string(pending.Status),
		}
		if pending.RequestPayload != nil {
			var preview interface{}
			if err := json.Unmarshal([]byte(*pending.RequestPayload), &preview); err == nil {
				pendingAction.PayloadPreview = preview
			}
		}
	}

	resp := &dto.SessionDetailResponse{
		ID:            session.ID,
		Title:         session.Title,
		Status:        string(session.Status),
		LastActivity:  formatTimePtr(session.LastActivity),
		MessageCount:  session.MessageCount,
		Messages:      messages,
		PendingAction: pendingAction,
		CreatedAt:     session.CreatedAt.Format(time.RFC3339),
	}

	return resp, nil
}

// DeleteSession soft deletes a session
func (u *ChatEngineUsecase) DeleteSession(ctx context.Context, sessionID string, userID string) error {
	session, err := u.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("AI_SESSION_NOT_FOUND: session not found: %w", err)
	}
	if session.UserID != userID {
		return fmt.Errorf("FORBIDDEN: you do not have access to this session")
	}
	return u.sessionRepo.Delete(ctx, sessionID)
}

// ListActions returns paginated action logs (admin)
func (u *ChatEngineUsecase) ListActions(ctx context.Context, req *dto.ListActionsRequest) ([]dto.ActionLogResponse, *utils.PaginationResult, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 20
	}

	actions, total, err := u.actionRepo.FindAll(ctx, req.Page, req.PerPage, req.UserID, req.Intent, req.Status)
	if err != nil {
		return nil, nil, fmt.Errorf("AI_CHAT_FAILED: failed to list actions: %w", err)
	}

	responses := make([]dto.ActionLogResponse, len(actions))
	for i, a := range actions {
		responses[i] = dto.ActionLogResponse{
			ID:         a.ID,
			SessionID:  a.SessionID,
			UserID:     a.UserID,
			Intent:     a.Intent,
			EntityType: a.EntityType,
			Action:     string(a.Action),
			Status:     string(a.Status),
			DurationMs: a.DurationMs,
			CreatedAt:  a.CreatedAt.Format(time.RFC3339),
		}
		if a.EntityID != nil {
			responses[i].EntityID = *a.EntityID
		}
		if a.ErrorMessage != "" {
			responses[i].ErrorMessage = a.ErrorMessage
		}
	}

	pagination := &utils.PaginationResult{
		Page:       req.Page,
		PerPage:    req.PerPage,
		Total:      int(total),
		TotalPages: int((total + int64(req.PerPage) - 1) / int64(req.PerPage)),
	}

	return responses, pagination, nil
}

// GetIntentRegistry returns all active intents
func (u *ChatEngineUsecase) GetIntentRegistry(ctx context.Context) ([]dto.IntentRegistryResponse, error) {
	return nil, fmt.Errorf("use the tool registry instead")
}

// --- Private helpers ---

func (u *ChatEngineUsecase) getOrCreateSession(ctx context.Context, sessionID *string, userID string) (*models.AIChatSession, error) {
	if sessionID != nil && *sessionID != "" {
		session, err := u.sessionRepo.FindByID(ctx, *sessionID)
		if err != nil {
			return nil, fmt.Errorf("AI_SESSION_NOT_FOUND: session not found: %w", err)
		}
		_ = u.sessionRepo.UpdateLastActivity(ctx, session.ID)
		return session, nil
	}

	now := apptime.Now()
	session := &models.AIChatSession{
		UserID:       userID,
		Title:        "New Chat",
		Status:       "ACTIVE",
		LastActivity: &now,
	}
	if err := u.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to create session: %w", err)
	}
	return session, nil
}

func (u *ChatEngineUsecase) toEngineMessages(messages []models.AIChatMessage) []engine.ConversationMessage {
	result := make([]engine.ConversationMessage, len(messages))
	for i, m := range messages {
		result[i] = engine.ConversationMessage{
			Role:    string(m.Role),
			Content: m.Content,
		}
	}
	return result
}

func (u *ChatEngineUsecase) buildUserContext(userID string, permissions map[string]bool, isAdmin bool, sessionID string) *aiContext.UserContext {
	return &aiContext.UserContext{
		UserID:      userID,
		IsAdmin:     isAdmin,
		Permissions: permissions,
		Locale:      "id", // Default to Indonesian
	}
}

func (u *ChatEngineUsecase) saveEngineResponse(ctx context.Context, session *models.AIChatSession, result *engine.EngineResult, start time.Time) (*dto.ChatResponse, error) {
	durationMs := time.Since(start).Milliseconds()

	// Build metadata from engine result
	metadata := map[string]interface{}{
		"turn_count":       result.TurnCount,
		"tool_calls_count": len(result.ToolCallResults),
	}
	metadataJSON, _ := json.Marshal(metadata)
	metaStr := string(metadataJSON)

	assistantMsg := &models.AIChatMessage{
		SessionID:  session.ID,
		Role:       models.MessageRoleAssistant,
		Content:    result.Response,
		Intent:     &metaStr,
		DurationMs: int(durationMs),
	}
	if err := u.messageRepo.Create(ctx, assistantMsg); err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to save assistant message: %w", err)
	}

	_ = u.sessionRepo.IncrementMessageCount(ctx, session.ID)
	_ = u.sessionRepo.UpdateLastActivity(ctx, session.ID)

	// Auto-generate title from first exchange
	if session.Title == "New Chat" && result.Response != "" {
		title := result.Response
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		session.Title = title
		_ = u.sessionRepo.UpdateTitle(ctx, session.ID, session.Title)
	}

	return &dto.ChatResponse{
		SessionID: session.ID,
		Message: dto.MessageResponse{
			ID:         assistantMsg.ID,
			Role:       string(assistantMsg.Role),
			Content:    assistantMsg.Content,
			DurationMs: assistantMsg.DurationMs,
			CreatedAt:  assistantMsg.CreatedAt.Format(time.RFC3339),
		},
		TokenUsage: &dto.TokenUsageResponse{TotalTokens: 0},
	}, nil
}

func (u *ChatEngineUsecase) saveSimpleResponse(ctx context.Context, session *models.AIChatSession, content string, start time.Time) (*dto.ChatResponse, error) {
	durationMs := time.Since(start).Milliseconds()

	assistantMsg := &models.AIChatMessage{
		SessionID:  session.ID,
		Role:       models.MessageRoleAssistant,
		Content:    content,
		DurationMs: int(durationMs),
	}
	if err := u.messageRepo.Create(ctx, assistantMsg); err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to save assistant message: %w", err)
	}

	_ = u.sessionRepo.IncrementMessageCount(ctx, session.ID)
	_ = u.sessionRepo.UpdateLastActivity(ctx, session.ID)

	return &dto.ChatResponse{
		SessionID: session.ID,
		Message: dto.MessageResponse{
			ID:         assistantMsg.ID,
			Role:       string(assistantMsg.Role),
			Content:    assistantMsg.Content,
			DurationMs: assistantMsg.DurationMs,
			CreatedAt:  assistantMsg.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// confirmActionStream executes a confirmed pending action and streams the formatted response.
// Called from SendMessageStream when the user sends an affirmative reply ("ya", "yes", etc.)
// while there is a pending action awaiting approval.
func (u *ChatEngineUsecase) confirmActionStream(
	ctx context.Context,
	action *models.AIActionLog,
	session *models.AIChatSession,
	userID string,
	userPermissions map[string]bool,
	isAdmin bool,
	userMessage string,
	model string,
	eventChan chan<- tools.StreamEvent,
) error {
	start := apptime.Now()

	// Persist the user's confirmation message for history continuity
	userMsg := &models.AIChatMessage{
		SessionID: session.ID,
		Role:      models.MessageRoleUser,
		Content:   userMessage,
	}
	_ = u.messageRepo.Create(ctx, userMsg)

	tool := u.toolRegistry.Get(strings.ToLower(action.Intent))
	if tool == nil {
		action.Status = models.ActionStatusFailed
		action.ErrorMessage = "Tool no longer available"
		_ = u.actionRepo.Update(ctx, action)
		resp := fmt.Sprintf("Maaf, tindakan '%s' tidak dapat dijalankan karena tool tidak tersedia.", action.Intent)
		eventChan <- tools.StreamEvent{Type: tools.EventContentDelta, Content: resp}
		eventChan <- tools.StreamEvent{Type: tools.EventMessageEnd, Data: map[string]interface{}{"duration_ms": time.Since(start).Milliseconds()}}
		result := &engine.EngineResult{Response: resp, TurnCount: 1, TotalDurationMs: time.Since(start).Milliseconds()}
		_, _ = u.saveEngineResponse(ctx, session, result, start)
		return nil
	}

	// Decode stored parameters
	var storedParams map[string]interface{}
	if action.RequestPayload != nil {
		_ = json.Unmarshal([]byte(*action.RequestPayload), &storedParams)
	}

	if validationMsg := u.validateActionPayloadAgainstFacts(ctx, session.ID, action.Intent, string(action.Action), storedParams); validationMsg != "" {
		action.Status = models.ActionStatusCancelled
		action.ErrorMessage = validationMsg
		_ = u.actionRepo.Update(ctx, action)
		eventChan <- tools.StreamEvent{Type: tools.EventContentDelta, Content: validationMsg}
		eventChan <- tools.StreamEvent{Type: tools.EventMessageEnd, Data: map[string]interface{}{"duration_ms": time.Since(start).Milliseconds()}}
		result := &engine.EngineResult{Response: validationMsg, TurnCount: 1, TotalDurationMs: time.Since(start).Milliseconds()}
		_, _ = u.saveEngineResponse(ctx, session, result, start)
		return nil
	}

	execCtx := &tools.ExecutionContext{
		UserID:          userID,
		UserPermissions: userPermissions,
		IsAdmin:         isAdmin,
		SessionID:       session.ID,
	}

	// Emit tool_call event so the UI shows a status card
	eventChan <- tools.StreamEvent{
		Type: tools.EventToolCall,
		Data: map[string]interface{}{"name": strings.ToLower(action.Intent), "parameters": storedParams},
	}

	toolResult, execErr := tool.Execute(ctx, storedParams, execCtx)

	// Emit tool_result event
	var errStr string
	if execErr != nil {
		errStr = execErr.Error()
	}
	eventChan <- tools.StreamEvent{
		Type: tools.EventToolResult,
		Data: map[string]interface{}{
			"call":   map[string]interface{}{"name": action.Intent, "parameters": storedParams},
			"result": toolResult,
			"error":  errStr,
		},
	}

	// Update action status in DB
	if execErr != nil || toolResult == nil || !toolResult.Success {
		action.Status = models.ActionStatusFailed
		if execErr != nil {
			action.ErrorMessage = execErr.Error()
		} else if toolResult != nil {
			action.ErrorMessage = toolResult.ErrorMessage
		}
		_ = u.actionRepo.Update(ctx, action)
	} else {
		action.Status = models.ActionStatusSuccess
		_ = u.actionRepo.Update(ctx, action)
	}

	// Use Cerebras to format the tool result as a natural language response
	availableTools := u.toolRegistry.FilterByPermissions(userPermissions, isAdmin)
	userCtx := u.buildUserContext(userID, userPermissions, isAdmin, session.ID)
	systemPrompt := u.contextBuilder.BuildFlatSystemPrompt(availableTools, userCtx, u.toolRegistry)

	toolResultJSON, _ := json.Marshal(toolResult)
	presentationPrompt := fmt.Sprintf(
		"Tool '%s' telah dijalankan. Berikut hasilnya:\n\n%s\n\nSajikan hasil ini kepada pengguna secara jelas dan terformat. Sertakan tautan navigasi yang relevan jika data ditampilkan.",
		strings.ToLower(action.Intent), string(toolResultJSON),
	)

	messages := []cerebras.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: presentationPrompt},
	}

	var fullResponse strings.Builder
	emittedLen := 0
	_, streamErr := u.cerebrasClient.ChatStream(&cerebras.ChatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.3,
		MaxTokens:   1024,
	}, func(chunk string) error {
		fullResponse.WriteString(chunk)
		current := fullResponse.String()
		if len(current) > emittedLen {
			eventChan <- tools.StreamEvent{Type: tools.EventContentDelta, Content: current[emittedLen:]}
			emittedLen = len(current)
		}
		return nil
	})

	responseText := strings.TrimSpace(fullResponse.String())
	if streamErr != nil || responseText == "" {
		if execErr != nil {
			responseText = fmt.Sprintf("Tindakan gagal: %v", execErr)
		} else {
			responseText = fmt.Sprintf("Tindakan '%s' berhasil dijalankan.", strings.ToLower(action.Intent))
		}
		eventChan <- tools.StreamEvent{Type: tools.EventContentDelta, Content: responseText}
	}

	durationMs := time.Since(start).Milliseconds()
	eventChan <- tools.StreamEvent{
		Type: tools.EventMessageEnd,
		Data: map[string]interface{}{"duration_ms": durationMs},
	}

	result := &engine.EngineResult{Response: responseText, TurnCount: 1, TotalDurationMs: durationMs}
	_, _ = u.saveEngineResponse(ctx, session, result, start)
	return nil
}

// cancelActionStream cancels a pending action and streams a brief acknowledgement.
func (u *ChatEngineUsecase) cancelActionStream(
	ctx context.Context,
	action *models.AIActionLog,
	session *models.AIChatSession,
	userMessage string,
	eventChan chan<- tools.StreamEvent,
) error {
	start := apptime.Now()

	_ = u.messageRepo.Create(ctx, &models.AIChatMessage{
		SessionID: session.ID,
		Role:      models.MessageRoleUser,
		Content:   userMessage,
	})

	action.Status = models.ActionStatusCancelled
	_ = u.actionRepo.Update(ctx, action)

	resp := "Tindakan dibatalkan. Ada yang bisa saya bantu lagi?"
	eventChan <- tools.StreamEvent{Type: tools.EventContentDelta, Content: resp}
	eventChan <- tools.StreamEvent{
		Type: tools.EventMessageEnd,
		Data: map[string]interface{}{"duration_ms": time.Since(start).Milliseconds()},
	}

	result := &engine.EngineResult{Response: resp, TurnCount: 1, TotalDurationMs: time.Since(start).Milliseconds()}
	_, _ = u.saveEngineResponse(ctx, session, result, start)
	return nil
}

func (u *ChatEngineUsecase) createPendingToolAction(ctx context.Context, session *models.AIChatSession, result *engine.EngineResult, userID string, start time.Time) (*dto.ChatResponse, error) {
	tc := result.PendingToolCall
	if tc == nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: pending tool call is nil")
	}

	tool := u.toolRegistry.Get(tc.Name)
	actionType := ""
	entityType := ""
	if tool != nil {
		actionType = tool.Category()
		entityType = tool.Module()
	}

	if validationMsg := u.validateActionPayloadAgainstFacts(ctx, session.ID, tc.Name, actionType, tc.Parameters); validationMsg != "" {
		return u.saveSimpleResponse(ctx, session, validationMsg, start)
	}

	requestJSON, _ := json.Marshal(tc.Parameters)
	reqPayload := string(requestJSON)

	actionLog := &models.AIActionLog{
		SessionID:      session.ID,
		UserID:         userID,
		Intent:         strings.ToUpper(tc.Name),
		EntityType:     entityType,
		Action:         models.ActionType(strings.ToUpper(actionType)),
		RequestPayload: &reqPayload,
		Status:         models.ActionStatusPendingConfirmation,
	}

	if err := u.actionRepo.Create(ctx, actionLog); err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to create pending action: %w", err)
	}

	// Save the confirmation message
	assistantContent := result.Response
	durationMs := time.Since(start).Milliseconds()

	assistantMsg := &models.AIChatMessage{
		SessionID:  session.ID,
		Role:       models.MessageRoleAssistant,
		Content:    assistantContent,
		DurationMs: int(durationMs),
	}
	if err := u.messageRepo.Create(ctx, assistantMsg); err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to save assistant message: %w", err)
	}

	_ = u.sessionRepo.IncrementMessageCount(ctx, session.ID)
	_ = u.sessionRepo.UpdateLastActivity(ctx, session.ID)

	return &dto.ChatResponse{
		SessionID: session.ID,
		Message: dto.MessageResponse{
			ID:         assistantMsg.ID,
			Role:       string(assistantMsg.Role),
			Content:    assistantContent,
			DurationMs: assistantMsg.DurationMs,
			CreatedAt:  assistantMsg.CreatedAt.Format(time.RFC3339),
		},
		Action: &dto.ActionPreview{
			ID:     actionLog.ID,
			Intent: actionLog.Intent,
			Status: string(actionLog.Status),
		},
		RequiresConfirmation: true,
	}, nil
}

func (u *ChatEngineUsecase) logToolCall(ctx context.Context, sessionID, userID string, tcr tools.ToolCallResult) {
	requestJSON, _ := json.Marshal(tcr.Call.Parameters)
	reqStr := string(requestJSON)

	actionLog := &models.AIActionLog{
		SessionID:      sessionID,
		UserID:         userID,
		Intent:         strings.ToUpper(tcr.Call.Name),
		Action:         models.ActionType("QUERY"),
		RequestPayload: &reqStr,
	}

	if tcr.Result != nil {
		actionLog.EntityType = tcr.Result.EntityType
		if tcr.Result.EntityID != "" {
			actionLog.EntityID = &tcr.Result.EntityID
		}
		actionLog.Action = models.ActionType(tcr.Result.Action)
		actionLog.DurationMs = int(tcr.Result.DurationMs)

		if tcr.Result.Success {
			actionLog.Status = models.ActionStatusSuccess
			responseJSON, _ := json.Marshal(tcr.Result.Data)
			respStr := string(responseJSON)
			actionLog.ResponsePayload = &respStr
		} else {
			actionLog.Status = models.ActionStatusFailed
			actionLog.ErrorMessage = tcr.Result.ErrorMessage
		}
	} else if tcr.Error != "" {
		actionLog.Status = models.ActionStatusFailed
		actionLog.ErrorMessage = tcr.Error
	}

	_ = u.actionRepo.Create(ctx, actionLog)
}

func (u *ChatEngineUsecase) validateActionPayloadAgainstFacts(ctx context.Context, sessionID string, intentCode string, actionType string, params map[string]interface{}) string {
	normalizedIntent := strings.ToUpper(strings.TrimSpace(intentCode))
	normalizedAction := strings.ToUpper(strings.TrimSpace(actionType))
	if normalizedAction == "" {
		switch {
		case strings.HasPrefix(normalizedIntent, "CREATE_"):
			normalizedAction = "CREATE"
		case strings.HasPrefix(normalizedIntent, "UPDATE_"):
			normalizedAction = "UPDATE"
		case strings.HasPrefix(normalizedIntent, "DELETE_"):
			normalizedAction = "DELETE"
		}
	}

	if params == nil {
		params = map[string]interface{}{}
	}

	if u.requestValidator != nil {
		validation := u.requestValidator.Validate(ctx, &IntentResult{
			IntentCode: normalizedIntent,
			ActionType: normalizedAction,
			Parameters: params,
		}, params)
		if validation != nil && !validation.Valid {
			return buildValidationGuardMessage(normalizedIntent, validation, nil)
		}
	}

	if normalizedIntent == "CREATE_SALES_ORDER" {
		evidence := u.loadUserFactEvidence(ctx, sessionID)
		ungrounded := collectUngroundedSalesOrderFields(params, evidence)
		if len(ungrounded) > 0 {
			return buildValidationGuardMessage(normalizedIntent, nil, ungrounded)
		}
	}

	return ""
}

func (u *ChatEngineUsecase) loadUserFactEvidence(ctx context.Context, sessionID string) string {
	messages, err := u.messageRepo.FindBySessionID(ctx, sessionID, 100)
	if err != nil || len(messages) == 0 {
		return ""
	}

	parts := make([]string, 0, len(messages))
	for _, msg := range messages {
		if msg.Role != models.MessageRoleUser {
			continue
		}
		content := strings.TrimSpace(msg.Content)
		if content == "" || isAffirmativeMessage(content) || isNegativeMessage(content) {
			continue
		}
		normalized := normalizeFactText(content)
		if normalized != "" {
			parts = append(parts, normalized)
		}
	}

	return strings.Join(parts, " ")
}

func collectUngroundedSalesOrderFields(params map[string]interface{}, evidence string) []string {
	evidence = normalizeFactText(evidence)
	stringFields := []string{
		"customer_name",
		"sales_rep_name",
		"payment_terms_name",
		"business_unit_name",
	}

	ungrounded := collectUngroundedNamedFields(params, evidence, stringFields)
	ungrounded = append(ungrounded, collectUngroundedProductItems(params["items"], evidence)...)

	return uniqueStrings(ungrounded)
}

func collectUngroundedNamedFields(params map[string]interface{}, evidence string, fields []string) []string {
	ungrounded := make([]string, 0)

	for _, field := range fields {
		value := strings.TrimSpace(getStringParam(params, field))
		if value == "" {
			continue
		}
		if !isValueGroundedInEvidence(evidence, value) {
			ungrounded = append(ungrounded, fmt.Sprintf("%s=%q", field, value))
		}
	}

	return ungrounded
}

func collectUngroundedProductItems(rawItems interface{}, evidence string) []string {
	ungrounded := make([]string, 0)
	items := toInterfaceSlice(rawItems)
	for idx, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		productName := strings.TrimSpace(getStringParam(itemMap, "product_name"))
		if productName == "" {
			continue
		}
		if !isValueGroundedInEvidence(evidence, productName) {
			ungrounded = append(ungrounded, fmt.Sprintf("items[%d].product_name=%q", idx, productName))
		}
	}

	return ungrounded
}

func toInterfaceSlice(value interface{}) []interface{} {
	if value == nil {
		return nil
	}

	if items, ok := value.([]interface{}); ok {
		return items
	}

	if strValue, ok := value.(string); ok {
		trimmed := strings.TrimSpace(strValue)
		if trimmed == "" {
			return nil
		}
		var parsed []interface{}
		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			return parsed
		}
	}

	return nil
}

func normalizeFactText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	var b strings.Builder
	lastSpace := true
	for _, ch := range value {
		isAlphaNumeric := (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
		if isAlphaNumeric {
			b.WriteRune(ch)
			lastSpace = false
			continue
		}
		if !lastSpace {
			b.WriteRune(' ')
			lastSpace = true
		}
	}

	return strings.TrimSpace(b.String())
}

func isValueGroundedInEvidence(evidence string, value string) bool {
	normalizedValue := normalizeFactText(value)
	if normalizedValue == "" {
		return true
	}
	if evidence == "" {
		return false
	}
	if strings.Contains(evidence, normalizedValue) {
		return true
	}

	tokens := strings.Fields(normalizedValue)
	if len(tokens) <= 1 {
		return strings.Contains(evidence, normalizedValue)
	}

	significant := 0
	matched := 0
	for _, token := range tokens {
		if len(token) < 3 {
			continue
		}
		significant++
		if strings.Contains(evidence, token) {
			matched++
		}
	}

	if significant == 0 {
		return strings.Contains(evidence, normalizedValue)
	}

	requiredMatches := 1
	if significant >= 3 {
		requiredMatches = 2
	}

	return matched >= requiredMatches
}

func buildValidationGuardMessage(intentCode string, validation *ValidationResult, ungrounded []string) string {
	normalizedIntent := strings.ToLower(intentCode)
	appendTemplate := strings.EqualFold(intentCode, "CREATE_SALES_ORDER")

	if len(ungrounded) > 0 {
		message := fmt.Sprintf(
			"Saya belum bisa mengeksekusi **%s** karena ada nilai yang tidak ditemukan di percakapan.\n\nPeriksa nilai berikut:\n%s\n\nMohon kirim data faktual dari Anda. Jika belum ada nilainya, isi null.",
			normalizedIntent,
			toMarkdownBulletList(ungrounded),
		)
		if appendTemplate {
			message += "\n\n" + salesOrderInputTemplateMarkdown()
		}
		return message
	}

	if validation != nil && len(validation.Errors) > 0 {
		messages := make([]string, 0, len(validation.Errors))
		seen := make(map[string]bool)
		for _, issue := range validation.Errors {
			msg := strings.TrimSpace(issue.Message)
			if msg == "" || seen[msg] {
				continue
			}
			seen[msg] = true
			messages = append(messages, msg)
		}

		if len(messages) > 0 {
			message := fmt.Sprintf(
				"Saya belum bisa mengeksekusi **%s** karena data belum lengkap/valid.\n\nLengkapi field berikut:\n%s\n\nMohon kirim nilai yang faktual. Jika belum ada nilainya, isi null.",
				normalizedIntent,
				toMarkdownBulletList(messages),
			)
			if appendTemplate {
				message += "\n\n" + salesOrderInputTemplateMarkdown()
			}
			return message
		}
	}

	return fmt.Sprintf(
		"Saya belum bisa mengeksekusi **%s** karena data belum cukup jelas. Mohon kirim data faktual yang lengkap. Jika ada nilai yang belum tersedia, isi null.",
		normalizedIntent,
	)
}

func isStandaloneConfirmationReply(message string) bool {
	if !(isAffirmativeMessage(message) || isNegativeMessage(message)) {
		return false
	}

	wordCount := len(strings.Fields(strings.TrimSpace(message)))
	return wordCount > 0 && wordCount <= 3
}

func (u *ChatEngineUsecase) handleOrphanConfirmationSync(
	ctx context.Context,
	session *models.AIChatSession,
	userMessage string,
	model string,
	userID string,
	userPermissions map[string]bool,
	isAdmin bool,
	start time.Time,
) (*dto.ChatResponse, error) {
	userMsg := &models.AIChatMessage{
		SessionID: session.ID,
		Role:      models.MessageRoleUser,
		Content:   userMessage,
	}
	if err := u.messageRepo.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to save user message: %w", err)
	}

	if recovered, err := u.tryRecoverOrphanConfirmation(ctx, session, model, userID, userPermissions, isAdmin, start); recovered != nil || err != nil {
		return recovered, err
	}

	response := orphanConfirmationMessage()
	return u.saveSimpleResponse(ctx, session, response, start)
}

func (u *ChatEngineUsecase) handleOrphanConfirmationStream(
	ctx context.Context,
	session *models.AIChatSession,
	userMessage string,
	model string,
	userID string,
	userPermissions map[string]bool,
	isAdmin bool,
	eventChan chan<- tools.StreamEvent,
) error {
	start := apptime.Now()

	userMsg := &models.AIChatMessage{
		SessionID: session.ID,
		Role:      models.MessageRoleUser,
		Content:   userMessage,
	}
	_ = u.messageRepo.Create(ctx, userMsg)

	if recovered, err := u.tryRecoverOrphanConfirmation(ctx, session, model, userID, userPermissions, isAdmin, start); err != nil {
		log.Printf("[AI] orphan confirmation recovery failed session_id=%s user_id=%s err=%v", session.ID, userID, err)
	} else if recovered != nil {
		eventChan <- tools.StreamEvent{Type: tools.EventContentDelta, Content: recovered.Message.Content}
		eventChan <- tools.StreamEvent{
			Type: tools.EventMessageEnd,
			Data: map[string]interface{}{"duration_ms": time.Since(start).Milliseconds(), "recovered": true},
		}
		return nil
	}

	response := orphanConfirmationMessage()
	eventChan <- tools.StreamEvent{Type: tools.EventContentDelta, Content: response}
	eventChan <- tools.StreamEvent{
		Type: tools.EventMessageEnd,
		Data: map[string]interface{}{"duration_ms": time.Since(start).Milliseconds()},
	}

	_, _ = u.saveSimpleResponse(ctx, session, response, start)
	return nil
}

func orphanConfirmationMessage() string {
	return "Saat ini belum ada tindakan yang menunggu konfirmasi.\n\nKemungkinan respons sebelumnya terpotong sehingga action belum terbentuk.\n\nSilakan kirim ulang detail yang ingin dieksekusi dalam satu pesan. Jika ada nilai yang belum tersedia, isi null."
}

func (u *ChatEngineUsecase) tryRecoverOrphanConfirmation(
	ctx context.Context,
	session *models.AIChatSession,
	model string,
	userID string,
	userPermissions map[string]bool,
	isAdmin bool,
	start time.Time,
) (*dto.ChatResponse, error) {
	history, err := u.messageRepo.FindBySessionID(ctx, session.ID, 8)
	if err != nil || len(history) == 0 {
		return nil, nil
	}

	if !hasRecentConfirmationPrompt(history) {
		return nil, nil
	}

	engineHistory := u.toEngineMessages(history)
	userCtx := u.buildUserContext(userID, userPermissions, isAdmin, session.ID)
	selectedModel := model
	if strings.TrimSpace(selectedModel) == "" {
		selectedModel = u.cerebrasClient.GetDefaultModel()
	}

	recoveryMessage := buildOrphanRecoveryMessage(history)
	result, processErr := u.engine.ProcessMessage(ctx, recoveryMessage, engineHistory, userCtx, selectedModel)
	if processErr != nil || result == nil {
		if processErr != nil {
			log.Printf("[AI] orphan confirmation recovery process failed session_id=%s user_id=%s err=%v", session.ID, userID, processErr)
		}
		return nil, nil
	}

	if result.RequiresConfirmation && result.PendingToolCall != nil {
		if _, err := u.createPendingToolAction(ctx, session, result, userID, start); err != nil {
			log.Printf("[AI] orphan confirmation recovery failed to create pending action session_id=%s user_id=%s err=%v", session.ID, userID, err)
			return nil, nil
		}

		pendingAction, pendingErr := u.actionRepo.FindPendingBySessionID(ctx, session.ID)
		if pendingErr != nil || pendingAction == nil {
			return nil, nil
		}

		return u.ConfirmAction(ctx, &dto.ConfirmActionRequest{
			ActionID:  pendingAction.ID,
			Confirmed: true,
		}, userID, userPermissions, isAdmin)
	}

	if strings.TrimSpace(result.Response) == "" {
		return nil, nil
	}

	return u.saveEngineResponse(ctx, session, result, start)
}

func hasRecentConfirmationPrompt(messages []models.AIChatMessage) bool {
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role != models.MessageRoleAssistant {
			continue
		}

		content := strings.ToLower(strings.TrimSpace(msg.Content))
		if content == "" {
			continue
		}

		if strings.Contains(content, "konfirmasi") ||
			strings.Contains(content, "apakah data di atas sudah benar") ||
			strings.Contains(content, "siap untuk dibuatkan") ||
			strings.Contains(content, "sebelum saya proses") {
			return true
		}

		break
	}

	return false
}

func buildOrphanRecoveryMessage(messages []models.AIChatMessage) string {
	lastDetail := ""
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role != models.MessageRoleUser {
			continue
		}

		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}

		if isStandaloneConfirmationReply(content) {
			continue
		}

		lastDetail = content
		break
	}

	if lastDetail == "" {
		return "Konfirmasi pengguna sudah diberikan (ya). Lanjutkan aksi terakhir yang sudah diringkas. Jika data sudah cukup, hasilkan tool_call yang valid dan lanjutkan alur eksekusi."
	}

	return fmt.Sprintf(
		"Konfirmasi pengguna sudah diberikan (ya). Gunakan ulang detail terakhir ini sebagai sumber fakta utama dan lanjutkan eksekusi tanpa meminta ulang field yang sudah ada:\n\n%s\n\nJika data sudah cukup, hasilkan tool_call valid untuk aksi terakhir dan lanjutkan.",
		lastDetail,
	)
}

func toMarkdownBulletList(items []string) string {
	if len(items) == 0 {
		return "- (tidak ada detail)"
	}

	lines := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		lines = append(lines, "- "+trimmed)
	}

	if len(lines) == 0 {
		return "- (tidak ada detail)"
	}

	return strings.Join(lines, "\n")
}

func salesOrderInputTemplateMarkdown() string {
	return "Template input Sales Order (isi null jika belum ada):\n```json\n{\n  \"customer_name\": null,\n  \"order_date\": null,\n  \"payment_terms_name\": null,\n  \"business_unit_name\": null,\n  \"sales_rep_name\": null,\n  \"items\": [\n    {\"product_name\": null, \"quantity\": null, \"price\": null, \"discount\": 0}\n  ],\n  \"notes\": null\n}\n```"
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
