package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/ai/data/models"
	"github.com/gilabs/gims/api/internal/ai/data/repositories"
	"github.com/gilabs/gims/api/internal/ai/domain/dto"
	"github.com/gilabs/gims/api/internal/ai/domain/mapper"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase/prompts"
	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/cerebras"
	"github.com/gilabs/gims/api/internal/core/utils"
)

// AIChatUsecase defines the AI chat business logic interface
type AIChatUsecase interface {
	SendMessage(ctx context.Context, req *dto.SendMessageRequest, userID string, userPermissions map[string]bool, isAdmin bool) (*dto.ChatResponse, error)
	ConfirmAction(ctx context.Context, req *dto.ConfirmActionRequest, userID string, userPermissions map[string]bool, isAdmin bool) (*dto.ChatResponse, error)
	ListSessions(ctx context.Context, req *dto.ListSessionsRequest, userID string) ([]dto.SessionListResponse, *utils.PaginationResult, error)
	GetSessionDetail(ctx context.Context, sessionID string, userID string) (*dto.SessionDetailResponse, error)
	DeleteSession(ctx context.Context, sessionID string, userID string) error
	ListActions(ctx context.Context, req *dto.ListActionsRequest) ([]dto.ActionLogResponse, *utils.PaginationResult, error)
	GetIntentRegistry(ctx context.Context) ([]dto.IntentRegistryResponse, error)
}

type aiChatUsecase struct {
	sessionRepo      repositories.ChatSessionRepository
	messageRepo      repositories.ChatMessageRepository
	actionRepo       repositories.ActionLogRepository
	intentRepo       repositories.IntentRegistryRepository
	cerebrasClient   *cerebras.Client
	chatMapper       *mapper.ChatMapper
	intentResolver   *IntentResolver
	paramExtractor   *ParameterExtractor
	requestValidator *RequestValidator
	permValidator    *PermissionValidator
	entityResolver   *EntityResolver
	actionExecutor   *ActionExecutor
}

// NewAIChatUsecase creates a new AI chat usecase instance
func NewAIChatUsecase(
	sessionRepo repositories.ChatSessionRepository,
	messageRepo repositories.ChatMessageRepository,
	actionRepo repositories.ActionLogRepository,
	intentRepo repositories.IntentRegistryRepository,
	cerebrasClient *cerebras.Client,
	chatMapper *mapper.ChatMapper,
	intentResolver *IntentResolver,
	paramExtractor *ParameterExtractor,
	requestValidator *RequestValidator,
	permValidator *PermissionValidator,
	entityResolver *EntityResolver,
	actionExecutor *ActionExecutor,
) AIChatUsecase {
	return &aiChatUsecase{
		sessionRepo:      sessionRepo,
		messageRepo:      messageRepo,
		actionRepo:       actionRepo,
		intentRepo:       intentRepo,
		cerebrasClient:   cerebrasClient,
		chatMapper:       chatMapper,
		intentResolver:   intentResolver,
		paramExtractor:   paramExtractor,
		requestValidator: requestValidator,
		permValidator:    permValidator,
		entityResolver:   entityResolver,
		actionExecutor:   actionExecutor,
	}
}

// affirmativePatterns matches user messages that confirm a pending action
var affirmativePatterns = []string{
	"ya", "yes", "ok", "oke", "iya", "setuju", "confirm", "lanjutkan",
	"proceed", "benar", "betul", "yap", "yup", "sure", "tentu", "boleh",
	"acc", "deal", "gas", "siap", "lanjut", "go", "do it", "jalankan",
}

// negativePatterns matches user messages that cancel a pending action
var negativePatterns = []string{
	"tidak", "no", "batal", "cancel", "jangan", "nope", "nah", "gak",
	"enggak", "nggak", "batalkan", "stop", "abort", "nevermind", "skip",
}

// isAffirmativeMessage checks if the user message matches an affirmative confirmation.
// Supports both exact matches ("ya", "ok") and word-boundary checks for multi-word
// messages like "saya setuju" or "iya lanjutkan".
func isAffirmativeMessage(msg string) bool {
	lower := strings.TrimSpace(strings.ToLower(msg))
	// Exact match first for single-word messages
	for _, p := range affirmativePatterns {
		if lower == p {
			return true
		}
	}
	// Word-boundary match for multi-word messages (e.g. "saya setuju", "ok lanjutkan")
	words := strings.Fields(lower)
	for _, w := range words {
		for _, p := range affirmativePatterns {
			if w == p {
				// Ensure no negative word is also present (e.g. "tidak setuju")
				hasNegative := false
				for _, neg := range negativePatterns {
					for _, ww := range words {
						if ww == neg {
							hasNegative = true
							break
						}
					}
					if hasNegative {
						break
					}
				}
				if !hasNegative {
					return true
				}
			}
		}
	}
	return false
}

// isNegativeMessage checks if the user message matches a negative/cancellation response.
// Uses word-boundary matching to handle multi-word messages like "tidak jadi" or "saya batal".
func isNegativeMessage(msg string) bool {
	lower := strings.TrimSpace(strings.ToLower(msg))
	// Exact match first
	for _, p := range negativePatterns {
		if lower == p {
			return true
		}
	}
	// Word-boundary match for multi-word messages
	words := strings.Fields(lower)
	for _, w := range words {
		for _, p := range negativePatterns {
			if w == p {
				return true
			}
		}
	}
	return false
}

// SendMessage processes a user message through the agentic AI pipeline
func (u *aiChatUsecase) SendMessage(ctx context.Context, req *dto.SendMessageRequest, userID string, userPermissions map[string]bool, isAdmin bool) (*dto.ChatResponse, error) {
	start := apptime.Now()

	if !u.cerebrasClient.IsConfigured() {
		return nil, fmt.Errorf("AI_SERVICE_NOT_CONFIGURED: Cerebras AI is not configured")
	}

	// Get or create session
	session, err := u.getOrCreateSession(ctx, req.SessionID, userID)
	if err != nil {
		return nil, err
	}

	// Validate session ownership for IDOR prevention
	if session.UserID != userID {
		return nil, fmt.Errorf("FORBIDDEN: you do not have access to this session")
	}

	// Check for pending action in session — intercept confirmation/cancellation messages
	// before running the full intent classification pipeline to prevent confirmation loops
	if req.SessionID != nil && *req.SessionID != "" {
		pendingAction, pendingErr := u.actionRepo.FindPendingBySessionID(ctx, session.ID)
		if pendingErr == nil && pendingAction != nil {
			if isAffirmativeMessage(req.Message) {
				// Auto-confirm: delegate to ConfirmAction flow
				return u.ConfirmAction(ctx, &dto.ConfirmActionRequest{
					ActionID:  pendingAction.ID,
					Confirmed: true,
				}, userID, userPermissions, isAdmin)
			}
			if isNegativeMessage(req.Message) {
				// Auto-cancel: delegate to ConfirmAction flow
				return u.ConfirmAction(ctx, &dto.ConfirmActionRequest{
					ActionID:  pendingAction.ID,
					Confirmed: false,
				}, userID, userPermissions, isAdmin)
			}
			// Message is neither affirmative nor negative — cancel the stale pending action
			// and process the message as a new request
			pendingAction.Status = models.ActionStatusCancelled
			_ = u.actionRepo.Update(ctx, pendingAction)
		}
	}

	// Save user message
	userMessage := &models.AIChatMessage{
		SessionID: session.ID,
		Role:      models.MessageRoleUser,
		Content:   req.Message,
	}
	if err := u.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to save user message: %w", err)
	}

	// Get conversation history for context
	history, err := u.messageRepo.FindBySessionID(ctx, session.ID, 10)
	if err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to load conversation history: %w", err)
	}

	// Build conversation context for LLM
	conversationHistory := u.buildConversationHistory(history)

	if intent := buildAutonomousHolidayCreateIntent(req.Message, history); intent != nil {
		permResult, permErr := u.permValidator.Validate(ctx, intent.IntentCode, userPermissions, isAdmin)
		if permErr != nil {
			assistantContent := fmt.Sprintf("Gagal mengeksekusi aksi holiday otomatis: %v", permErr)
			return u.saveAssistantResponse(ctx, session, assistantContent, intent, start)
		}
		if !permResult.Allowed {
			assistantContent := fmt.Sprintf("Maaf, Anda tidak memiliki izin untuk tindakan ini. %s", permResult.Reason)
			return u.saveAssistantResponse(ctx, session, assistantContent, intent, start)
		}

		validation := u.requestValidator.Validate(ctx, intent, intent.Parameters)
		if !validation.Valid {
			assistantContent := u.buildMissingFieldsMessage(ctx, intent, validation)
			return u.saveAssistantResponse(ctx, session, assistantContent, intent, start)
		}

		return u.executeAndRespond(ctx, session, intent, validation.ResolvedEntities, userID, permResult.RequiredPermission, conversationHistory, start)
	}

	if isCurrentTimeQuestion(req.Message) {
		now := apptime.Now()
		assistantContent := fmt.Sprintf("Waktu saat ini: **%s** (WIB). Tahun sekarang: **%d**.", now.Format("02 Jan 2006 15:04:05"), now.Year())
		return u.saveAssistantResponse(ctx, session, assistantContent, &IntentResult{
			IntentCode: "GENERAL_CHAT",
			Confidence: 1,
			Parameters: map[string]interface{}{"source": "apptime_now"},
			Module:     "general",
			ActionType: "QUERY",
			IsQuery:    true,
			RawMessage: req.Message,
		}, start)
	}

	// Step 1: Extract intent from user message (uses cheap model for classification)
	intent, err := u.intentResolver.Resolve(ctx, req.Message, conversationHistory)
	if err != nil {
		// Fall back to general chat on intent resolution failure
		return u.handleGeneralChat(ctx, session, req.Message, conversationHistory, req.Model, start)
	}

	// Deterministic fallback for critical business actions when classifier confidence is low.
	intent = u.applyDeterministicIntentFallback(req.Message, intent)
	intent = u.applyContextualCreateContinuationFallback(req.Message, history, intent)

	// Step 2: Handle general chat directly without agentic flow
	if intent.IntentCode == "GENERAL_CHAT" || intent.Confidence < 0.6 {
		return u.handleGeneralChat(ctx, session, req.Message, conversationHistory, req.Model, start)
	}

	// Step 3: Validate permissions
	permResult, err := u.permValidator.Validate(ctx, intent.IntentCode, userPermissions, isAdmin)
	if err != nil {
		assistantContent := fmt.Sprintf("Aksi tidak bisa diproses saat ini: %v", err)
		return u.saveAssistantResponse(ctx, session, assistantContent, intent, start)
	}

	if !permResult.Allowed {
		assistantContent := fmt.Sprintf("Maaf, Anda tidak memiliki izin untuk tindakan ini. %s", permResult.Reason)
		return u.saveAssistantResponse(ctx, session, assistantContent, intent, start)
	}

	// Step 4 (Layer 2): Extract structured parameters using focused LLM prompt
	extractedParams, err := u.paramExtractor.Extract(ctx, intent, req.Message, conversationHistory)
	if err == nil && len(extractedParams) > 0 {
		intent.Parameters = extractedParams
	}

	// Step 4.5: Post-extraction processing
	// a) Strip fabricated entity values when user asks for "random" data
	u.sanitizeIfRandomRequest(req.Message, intent.Parameters)
	// b) Normalize parameter names to match validation expectations
	u.normalizeParamNames(intent)
	// c) Carry forward parameters from a previous same-intent exchange
	u.mergeWithPreviousIntentParams(history, intent)
	// d) Re-normalize after merge (previous params may have old field names)
	u.normalizeParamNames(intent)

	// Step 5 (Layer 3): Backend validation — no LLM, validates required fields and resolves entities
	validation := u.requestValidator.Validate(ctx, intent, intent.Parameters)
	resolvedEntities := validation.ResolvedEntities

	// If validation has blocking errors for CREATE actions, ask the user naturally
	if !validation.Valid && intent.ActionType == "CREATE" {
		assistantContent := u.buildMissingFieldsMessage(ctx, intent, validation)
		return u.saveAssistantResponse(ctx, session, assistantContent, intent, start)
	}

	// Step 6: Check if action requires confirmation
	needsConfirmation, _ := u.permValidator.NeedsConfirmation(ctx, intent.IntentCode)
	if intent.IntentCode == "CREATE_HOLIDAY" && isIndonesiaBulkHolidayMode(intent.Parameters) {
		needsConfirmation = false
	}

	if needsConfirmation && !intent.IsQuery {
		// Create pending action log for confirmation
		actionPreview := u.actionExecutor.BuildActionPreview(intent, resolvedEntities)
		return u.createPendingAction(ctx, session, intent, resolvedEntities, actionPreview, userID, permResult.RequiredPermission, start)
	}

	// Step 7: Execute action directly (queries or actions not requiring confirmation)
	return u.executeAndRespond(ctx, session, intent, resolvedEntities, userID, permResult.RequiredPermission, conversationHistory, start)
}

func (u *aiChatUsecase) applyContextualCreateContinuationFallback(message string, history []models.AIChatMessage, intent *IntentResult) *IntentResult {
	if intent == nil {
		return intent
	}

	if intent.IntentCode != "GENERAL_CHAT" && intent.Confidence >= 0.6 {
		return intent
	}

	previous := u.findLatestAssistantCreateIntent(history)
	if previous == nil {
		return intent
	}

	lower := strings.ToLower(strings.TrimSpace(message))
	isDetailFollowUp := strings.Contains(lower, ":") || strings.Contains(lower, "payment") || strings.Contains(lower, "terms") || strings.Contains(lower, "bisnis unit") || strings.Contains(lower, "business unit") || strings.Contains(lower, "items") || strings.Contains(lower, "qty") || strings.Contains(lower, "harga") || strings.Contains(lower, "price")
	if !isDetailFollowUp && !isAffirmativeMessage(message) {
		return intent
	}

	return &IntentResult{
		IntentCode: previous.IntentCode,
		Confidence: 0.82,
		Parameters: map[string]interface{}{},
		Module:     previous.Module,
		ActionType: previous.ActionType,
		IsQuery:    false,
		RawMessage: message,
	}
}

func (u *aiChatUsecase) findLatestAssistantCreateIntent(history []models.AIChatMessage) *IntentResult {
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		if msg.Role != models.MessageRoleAssistant || msg.Intent == nil {
			continue
		}

		var prev IntentResult
		if err := json.Unmarshal([]byte(*msg.Intent), &prev); err != nil {
			continue
		}

		if strings.ToUpper(strings.TrimSpace(prev.ActionType)) != "CREATE" {
			continue
		}

		switch prev.IntentCode {
		case "CREATE_SALES_ORDER", "CREATE_SALES_QUOTATION", "CREATE_PURCHASE_ORDER", "CREATE_LEAVE_REQUEST":
			return &prev
		}
	}

	return nil
}

// ConfirmAction processes a user's confirmation of a pending action
func (u *aiChatUsecase) ConfirmAction(ctx context.Context, req *dto.ConfirmActionRequest, userID string, userPermissions map[string]bool, isAdmin bool) (*dto.ChatResponse, error) {
	start := apptime.Now()

	// Find the pending action
	action, err := u.actionRepo.FindByID(ctx, req.ActionID)
	if err != nil {
		return nil, fmt.Errorf("AI_ACTION_NOT_FOUND: action not found: %w", err)
	}

	// Validate ownership (IDOR prevention)
	if action.UserID != userID {
		return nil, fmt.Errorf("FORBIDDEN: you do not have access to this action")
	}

	if action.Status != models.ActionStatusPendingConfirmation {
		return nil, fmt.Errorf("AI_ACTION_INVALID_STATE: action is not pending confirmation (current: %s)", action.Status)
	}

	// Find session
	session, err := u.sessionRepo.FindByID(ctx, action.SessionID)
	if err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: session not found: %w", err)
	}

	if !req.Confirmed {
		// User cancelled
		action.Status = models.ActionStatusCancelled
		if err := u.actionRepo.Update(ctx, action); err != nil {
			return nil, fmt.Errorf("AI_CHAT_FAILED: failed to update action: %w", err)
		}

		assistantContent := "Tindakan dibatalkan. Ada yang bisa saya bantu lagi?"
		return u.saveAssistantResponse(ctx, session, assistantContent, nil, start)
	}

	// Re-validate permissions before execution
	intentCode := action.Intent
	permResult, err := u.permValidator.Validate(ctx, intentCode, userPermissions, isAdmin)
	if err != nil || !permResult.Allowed {
		action.Status = models.ActionStatusFailed
		action.ErrorMessage = "Permission denied on confirmation"
		_ = u.actionRepo.Update(ctx, action)
		return nil, fmt.Errorf("FORBIDDEN: permission denied for this action")
	}

	// Parse stored request payload to reconstruct intent
	var storedPayload map[string]interface{}
	if action.RequestPayload != nil {
		if err := json.Unmarshal([]byte(*action.RequestPayload), &storedPayload); err != nil {
			return nil, fmt.Errorf("AI_CHAT_FAILED: failed to parse action payload: %w", err)
		}
	}

	intent := &IntentResult{
		IntentCode: action.Intent,
		Parameters: storedPayload,
		Module:     action.EntityType,
		ActionType: string(action.Action),
	}

	// Resolve entities again for execution
	resolvedEntities, _ := u.entityResolver.ResolveEntitiesFromParameters(ctx, intent.Parameters)

	// Execute the confirmed action
	result := u.actionExecutor.Execute(ctx, intent, resolvedEntities, userID)

	// Update action log
	if result.Success {
		action.Status = models.ActionStatusSuccess
		if result.EntityID != "" {
			action.EntityID = &result.EntityID
		}
		responseJSON, _ := json.Marshal(result.Data)
		respStr := string(responseJSON)
		action.ResponsePayload = &respStr
	} else {
		action.Status = models.ActionStatusFailed
		action.ErrorMessage = result.ErrorMessage
	}
	action.DurationMs = int(result.DurationMs)
	_ = u.actionRepo.Update(ctx, action)

	// Generate natural language response
	var assistantContent string
	if result.Success {
		assistantContent = fmt.Sprintf("Tindakan berhasil dilakukan. %s", result.Message)
	} else {
		assistantContent = fmt.Sprintf("Maaf, tindakan gagal: %s", result.ErrorMessage)
	}

	return u.saveAssistantResponse(ctx, session, assistantContent, intent, start)
}

// ListSessions returns paginated sessions for a user
func (u *aiChatUsecase) ListSessions(ctx context.Context, req *dto.ListSessionsRequest, userID string) ([]dto.SessionListResponse, *utils.PaginationResult, error) {
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

	responses := u.chatMapper.ToSessionListResponses(sessions)
	pagination := &utils.PaginationResult{
		Page:       req.Page,
		PerPage:    req.PerPage,
		Total:      int(total),
		TotalPages: int((total + int64(req.PerPage) - 1) / int64(req.PerPage)),
	}

	return responses, pagination, nil
}

// GetSessionDetail returns a session with its messages and actions
func (u *aiChatUsecase) GetSessionDetail(ctx context.Context, sessionID string, userID string) (*dto.SessionDetailResponse, error) {
	session, err := u.sessionRepo.FindByIDWithMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("AI_SESSION_NOT_FOUND: session not found: %w", err)
	}

	// IDOR prevention
	if session.UserID != userID {
		return nil, fmt.Errorf("FORBIDDEN: you do not have access to this session")
	}

	response := u.chatMapper.ToSessionDetailResponse(session)
	return &response, nil
}

// DeleteSession soft deletes a session owned by the user
func (u *aiChatUsecase) DeleteSession(ctx context.Context, sessionID string, userID string) error {
	session, err := u.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("AI_SESSION_NOT_FOUND: session not found: %w", err)
	}

	if session.UserID != userID {
		return fmt.Errorf("FORBIDDEN: you do not have access to this session")
	}

	return u.sessionRepo.Delete(ctx, sessionID)
}

// ListActions returns paginated action logs (admin endpoint)
func (u *aiChatUsecase) ListActions(ctx context.Context, req *dto.ListActionsRequest) ([]dto.ActionLogResponse, *utils.PaginationResult, error) {
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

	responses := u.chatMapper.ToActionLogResponses(actions)
	pagination := &utils.PaginationResult{
		Page:       req.Page,
		PerPage:    req.PerPage,
		Total:      int(total),
		TotalPages: int((total + int64(req.PerPage) - 1) / int64(req.PerPage)),
	}

	return responses, pagination, nil
}

// GetIntentRegistry returns all active intents
func (u *aiChatUsecase) GetIntentRegistry(ctx context.Context) ([]dto.IntentRegistryResponse, error) {
	intents, err := u.intentRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to fetch intent registry: %w", err)
	}

	return u.chatMapper.ToIntentRegistryResponses(intents), nil
}

// --- Private helper methods ---

func (u *aiChatUsecase) getOrCreateSession(ctx context.Context, sessionID *string, userID string) (*models.AIChatSession, error) {
	if sessionID != nil && *sessionID != "" {
		session, err := u.sessionRepo.FindByID(ctx, *sessionID)
		if err != nil {
			return nil, fmt.Errorf("AI_SESSION_NOT_FOUND: session not found: %w", err)
		}
		_ = u.sessionRepo.UpdateLastActivity(ctx, session.ID)
		return session, nil
	}

	// Create new session
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

func (u *aiChatUsecase) buildConversationHistory(messages []models.AIChatMessage) []cerebras.ChatMessage {
	history := make([]cerebras.ChatMessage, 0, len(messages))
	for _, msg := range messages {
		history = append(history, cerebras.ChatMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}
	return history
}

func (u *aiChatUsecase) handleGeneralChat(ctx context.Context, session *models.AIChatSession, userMessage string, conversationHistory []cerebras.ChatMessage, model string, start time.Time) (*dto.ChatResponse, error) {
	messages := make([]cerebras.ChatMessage, 0, len(conversationHistory)+3)
	messages = append(messages, cerebras.ChatMessage{
		Role:    "system",
		Content: prompts.GeneralChatSystemPrompt,
	})
	messages = append(messages, cerebras.ChatMessage{
		Role:    "system",
		Content: currentDatetimeContextMessage(),
	})
	messages = append(messages, conversationHistory...)
	messages = append(messages, cerebras.ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	resp, err := u.cerebrasClient.Chat(&cerebras.ChatRequest{
		Messages:    messages,
		Model:       model,
		MaxTokens:   1024,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: %w", err)
	}

	return u.saveAssistantResponse(ctx, session, resp.Message.Content, nil, start)
}

func (u *aiChatUsecase) saveAssistantResponse(ctx context.Context, session *models.AIChatSession, content string, intent *IntentResult, start time.Time) (*dto.ChatResponse, error) {
	durationMs := time.Since(start).Milliseconds()

	// Build intent JSON for storage
	var intentJSON string
	if intent != nil {
		data, _ := json.Marshal(intent)
		intentJSON = string(data)
	}

	// Save assistant message
	var intentPtr *string
	if intentJSON != "" {
		intentPtr = &intentJSON
	}
	assistantMsg := &models.AIChatMessage{
		SessionID:  session.ID,
		Role:       models.MessageRoleAssistant,
		Content:    content,
		Intent:     intentPtr,
		DurationMs: int(durationMs),
	}
	if err := u.messageRepo.Create(ctx, assistantMsg); err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to save assistant message: %w", err)
	}

	// Update session counters
	_ = u.sessionRepo.IncrementMessageCount(ctx, session.ID)
	_ = u.sessionRepo.UpdateLastActivity(ctx, session.ID)

	// Update session title from first meaningful exchange
	if session.Title == "New Chat" && content != "" {
		titleContent := content
		if len(titleContent) > 50 {
			titleContent = titleContent[:50] + "..."
		}
		session.Title = titleContent
		_ = u.sessionRepo.UpdateTitle(ctx, session.ID, session.Title)
	}

	return &dto.ChatResponse{
		SessionID: session.ID,
		Message:   u.chatMapper.ToMessageResponse(assistantMsg),
		TokenUsage: &dto.TokenUsageResponse{
			TotalTokens: 0,
		},
	}, nil
}

func (u *aiChatUsecase) createPendingAction(ctx context.Context, session *models.AIChatSession, intent *IntentResult, resolvedEntities map[string]*ResolvedEntity, actionPreview map[string]interface{}, userID string, permUsed string, start time.Time) (*dto.ChatResponse, error) {
	// Serialize request payload
	requestJSON, _ := json.Marshal(intent.Parameters)
	reqPayload := string(requestJSON)

	actionLog := &models.AIActionLog{
		SessionID:      session.ID,
		UserID:         userID,
		Intent:         intent.IntentCode,
		EntityType:     intent.Module,
		Action:         models.ActionType(intent.ActionType),
		RequestPayload: &reqPayload,
		Status:         models.ActionStatusPendingConfirmation,
		PermissionUsed: permUsed,
	}

	if err := u.actionRepo.Create(ctx, actionLog); err != nil {
		return nil, fmt.Errorf("AI_CHAT_FAILED: failed to create action log: %w", err)
	}

	// Build confirmation message
	preview := u.chatMapper.ToActionPreview(actionLog)
	if params, ok := actionPreview["parameters"].(map[string]interface{}); ok {
		preview.PayloadPreview = params
	}
	if entities, ok := actionPreview["resolved_entities"]; ok {
		if entityMap, ok := entities.(map[string]string); ok {
			previewMap := make(map[string]interface{})
			if existing, ok := preview.PayloadPreview.(map[string]interface{}); ok {
				for k, v := range existing {
					previewMap[k] = v
				}
			}
			for k, v := range entityMap {
				previewMap["resolved_"+k] = v
			}
			preview.PayloadPreview = previewMap
		}
	}

	assistantContent := u.buildConfirmationMessage(intent, resolvedEntities)

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
		SessionID:            session.ID,
		Message:              u.chatMapper.ToMessageResponse(assistantMsg),
		Action:               preview,
		RequiresConfirmation: true,
	}, nil
}

func (u *aiChatUsecase) executeAndRespond(ctx context.Context, session *models.AIChatSession, intent *IntentResult, resolvedEntities map[string]*ResolvedEntity, userID string, permUsed string, conversationHistory []cerebras.ChatMessage, start time.Time) (*dto.ChatResponse, error) {
	// Execute the action
	result := u.actionExecutor.Execute(ctx, intent, resolvedEntities, userID)

	// For not-yet-implemented intents, return the guidance directly without LLM roundtrip
	if result.ErrorCode == "NOT_IMPLEMENTED" {
		return u.saveAssistantResponse(ctx, session, result.Message, intent, start)
	}

	// Log the action
	requestJSON, _ := json.Marshal(intent.Parameters)
	responseJSON, _ := json.Marshal(result.Data)
	reqStr := string(requestJSON)
	respStr := string(responseJSON)

	actionLog := &models.AIActionLog{
		SessionID:       session.ID,
		UserID:          userID,
		Intent:          intent.IntentCode,
		EntityType:      result.EntityType,
		Action:          models.ActionType(result.Action),
		RequestPayload:  &reqStr,
		ResponsePayload: &respStr,
		PermissionUsed:  permUsed,
		DurationMs:      int(result.DurationMs),
	}

	if result.EntityID != "" {
		actionLog.EntityID = &result.EntityID
	}

	if result.Success {
		actionLog.Status = models.ActionStatusSuccess
	} else {
		actionLog.Status = models.ActionStatusFailed
		actionLog.ErrorMessage = result.ErrorMessage
	}

	_ = u.actionRepo.Create(ctx, actionLog)

	if intent.IntentCode == "CREATE_HOLIDAY" && isIndonesiaBulkHolidayMode(intent.Parameters) {
		assistantContent := formatHolidayAutoCreateResult(result)
		return u.saveAssistantResponse(ctx, session, assistantContent, intent, start)
	}

	// Generate natural language response using LLM
	assistantContent, err := u.generateActionResponse(result, intent)
	if err != nil {
		// Fallback to structured response if LLM fails
		if result.Success {
			assistantContent = result.Message
		} else {
			assistantContent = fmt.Sprintf("Terjadi kesalahan: %s", result.ErrorMessage)
		}
	}

	return u.saveAssistantResponse(ctx, session, assistantContent, intent, start)
}

func formatHolidayAutoCreateResult(result *ActionResult) string {
	if result == nil {
		return "Terjadi kesalahan saat membuat holiday otomatis."
	}

	if !result.Success {
		if result.ErrorMessage != "" {
			return fmt.Sprintf("Gagal membuat holiday otomatis: %s", result.ErrorMessage)
		}
		return "Gagal membuat holiday otomatis."
	}

	createdCount := 0
	year := apptime.Now().Year()
	createdCount, year = readHolidayResultMetrics(result.Data, year)

	if createdCount == 0 {
		return fmt.Sprintf("Tidak ada data holiday baru yang perlu dibuat untuk tahun %d. Semua data sudah ada.", year)
	}

	return fmt.Sprintf("Berhasil membuat %d data holiday Indonesia untuk tahun %d ke modul Holiday.", createdCount, year)
}

func readHolidayResultMetrics(data interface{}, defaultYear int) (int, int) {
	createdCount := 0
	year := defaultYear

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return createdCount, year
	}

	if n, ok := toInt(dataMap["created_count"]); ok {
		createdCount = n
	}
	if n, ok := toInt(dataMap["year"]); ok {
		year = n
	}

	return createdCount, year
}

func toInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int32:
		return int(val), true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case float32:
		return int(val), true
	case json.Number:
		n, err := val.Int64()
		if err != nil {
			return 0, false
		}
		return int(n), true
	default:
		return 0, false
	}
}

func (u *aiChatUsecase) generateActionResponse(result *ActionResult, intent *IntentResult) (string, error) {
	// Build a targeted JSON summary — avoid sending huge payloads that confuse the LLM
	resultJSON, _ := json.Marshal(result)

	// Truncate very large JSON to prevent LLM confusion
	jsonStr := string(resultJSON)
	const maxJSONLen = 4000
	if len(jsonStr) > maxJSONLen {
		jsonStr = jsonStr[:maxJSONLen] + "...TRUNCATED]"
	}

	messages := []cerebras.ChatMessage{
		{
			Role:    "system",
			Content: prompts.ActionResponseSystemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(prompts.ActionResponseUserTemplate, intent.IntentCode, jsonStr),
		},
	}

	resp, err := u.cerebrasClient.Chat(&cerebras.ChatRequest{
		Messages:    messages,
		MaxTokens:   1500,
		Temperature: 0.3,
	})
	if err != nil {
		return "", err
	}

	return resp.Message.Content, nil
}

func (u *aiChatUsecase) buildConfirmationMessage(intent *IntentResult, resolvedEntities map[string]*ResolvedEntity) string {
	switch intent.IntentCode {
	case "CREATE_HOLIDAY":
		return u.buildHolidayConfirmationMessage(intent)
	case "CREATE_SALES_QUOTATION":
		return u.buildSalesQuotationConfirmationMessage(intent, resolvedEntities)
	case "CREATE_SALES_ORDER":
		return u.buildSalesOrderConfirmationMessage(intent, resolvedEntities)
	case "CREATE_SALES_TARGET":
		return u.buildSalesTargetConfirmationMessage(intent, resolvedEntities)
	case "CREATE_LEAVE_REQUEST":
		return u.buildLeaveRequestConfirmationMessage(intent)
	default:
		return u.buildGenericConfirmationMessage(intent, resolvedEntities)
	}
}

func (u *aiChatUsecase) buildGenericConfirmationMessage(intent *IntentResult, resolvedEntities map[string]*ResolvedEntity) string {
	displayName := u.getIntentDisplayName(intent.IntentCode)
	actionLabel := confirmationActionLabel(intent.ActionType)
	details := collectConfirmationDetails(intent.Parameters, resolvedEntities)

	if len(details) == 0 {
		return fmt.Sprintf("Saya akan %s **%s**.\n\nApakah Anda setuju?", actionLabel, displayName)
	}

	return fmt.Sprintf("Saya akan %s **%s** dengan detail:\n%s\n\nApakah Anda setuju?", actionLabel, displayName, strings.Join(details, "\n"))
}

func confirmationActionLabel(actionType string) string {
	switch strings.ToUpper(strings.TrimSpace(actionType)) {
	case "CREATE":
		return "membuat"
	case "UPDATE":
		return "memperbarui"
	case "DELETE":
		return "menghapus"
	default:
		return "menjalankan"
	}
}

func collectConfirmationDetails(params map[string]interface{}, resolvedEntities map[string]*ResolvedEntity) []string {
	if params == nil {
		params = map[string]interface{}{}
	}

	var details []string
	appendIf := func(label, val string) {
		value := strings.TrimSpace(val)
		if value != "" {
			details = append(details, fmt.Sprintf("- **%s**: %s", label, value))
		}
	}

	appendIf("Customer", getStringParam(params, "customer_name"))
	appendIf("Supplier", getStringParam(params, "supplier_name"))
	appendIf("Karyawan", getStringParam(params, "employee_name"))
	appendIf("Company", getStringParam(params, "company_name"))
	appendIf("Division", getStringParam(params, "division_name"))
	appendIf("Bank", getStringParam(params, "bank_name"))
	appendIf("Area", getStringParam(params, "area_name"))
	appendIf("Produk", getStringParam(params, "product_name"))
	appendIf("Gudang", getStringParam(params, "warehouse_name"))
	appendIf("Business Type", getStringParam(params, "business_type_name"))
	appendIf("Supplier Type", getStringParam(params, "supplier_type_name"))
	appendIf("Customer Type", getStringParam(params, "customer_type_name"))
	appendIf("Tanggal", firstNonEmptyString(getStringParam(params, "quotation_date"), getStringParam(params, "order_date"), getStringParam(params, "date")))
	appendIf("Mulai", firstNonEmptyString(getStringParam(params, "start_date"), getStringParam(params, "date_from")))
	appendIf("Selesai", firstNonEmptyString(getStringParam(params, "end_date"), getStringParam(params, "date_to")))
	appendIf("Status", getStringParam(params, "status"))
	appendIf("Pencarian", getStringParam(params, "search"))

	if year := getIntParam(params, "year"); year != 0 {
		details = append(details, fmt.Sprintf("- **Tahun**: %d", year))
	}

	if amount := firstPositiveAmount(params, "total_target", "amount", "total"); amount > 0 {
		details = append(details, fmt.Sprintf("- **Nilai**: %.0f", amount))
	}

	if items, ok := params["items"].([]interface{}); ok && len(items) > 0 {
		details = append(details, fmt.Sprintf("- **Jumlah Item**: %d", len(items)))
	}

	resolvedLabels := map[string]string{
		"customer":  "Customer (resolved)",
		"supplier":  "Supplier (resolved)",
		"employee":  "Karyawan (resolved)",
		"company":   "Company (resolved)",
		"division":  "Division (resolved)",
		"bank":      "Bank (resolved)",
		"area":      "Area (resolved)",
		"product":   "Produk (resolved)",
		"warehouse": "Gudang (resolved)",
		"business_type": "Business Type (resolved)",
		"supplier_type": "Supplier Type (resolved)",
		"customer_type": "Customer Type (resolved)",
	}
	for key, label := range resolvedLabels {
		if entity, ok := resolvedEntities[key]; ok && entity != nil {
			if name := strings.TrimSpace(entity.DisplayName); name != "" {
				details = append(details, fmt.Sprintf("- **%s**: %s", label, name))
			}
		}
	}

	return uniqueStrings(details)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		v := strings.TrimSpace(value)
		if v != "" {
			return v
		}
	}
	return ""
}

func firstPositiveAmount(params map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		if value := getFloatParam(params, key); value > 0 {
			return value
		}
	}
	return 0
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (u *aiChatUsecase) buildHolidayConfirmationMessage(intent *IntentResult) string {
	name := getStringParam(intent.Parameters, "name")
	date := getStringParam(intent.Parameters, "date")
	return fmt.Sprintf("Saya akan membuat hari libur:\n- **Nama**: %s\n- **Tanggal**: %s\n- **Tipe**: %s\n\nApakah Anda setuju?",
		name, date, getStringParam(intent.Parameters, "type"))
}

func (u *aiChatUsecase) buildSalesOrderConfirmationMessage(intent *IntentResult, resolvedEntities map[string]*ResolvedEntity) string {
	customer := getStringParam(intent.Parameters, "customer_name")
	if c, ok := resolvedEntities["customer"]; ok && c.DisplayName != "" {
		customer = c.DisplayName
	}
	date := getStringParam(intent.Parameters, "order_date")
	if date == "" {
		date = apptime.Now().Format("2006-01-02")
	}

	details := []string{
		fmt.Sprintf("- **Customer**: %s", customer),
		fmt.Sprintf("- **Tanggal**: %s", date),
	}
	if pt := getStringParam(intent.Parameters, "payment_terms_name"); pt != "" {
		details = append(details, fmt.Sprintf("- **Syarat Pembayaran**: %s", pt))
	}
	if bu := getStringParam(intent.Parameters, "business_unit_name"); bu != "" {
		details = append(details, fmt.Sprintf("- **Unit Bisnis**: %s", bu))
	}
	if sr := firstNonEmptyString(
		getStringParam(intent.Parameters, "sales_rep_name"),
		getStringParam(intent.Parameters, "sales_rep"),
		getStringParam(intent.Parameters, "sales_name"),
	); sr != "" {
		details = append(details, fmt.Sprintf("- **Sales Rep**: %s", sr))
	}
	if items, ok := intent.Parameters["items"].([]interface{}); ok && len(items) > 0 {
		details = append(details, fmt.Sprintf("- **Jumlah Item**: %d produk", len(items)))
	}

	return fmt.Sprintf("Saya akan membuat Sales Order:\n%s\n\nApakah Anda setuju?", strings.Join(details, "\n"))
}

func (u *aiChatUsecase) buildLeaveRequestConfirmationMessage(intent *IntentResult) string {
	startDate := getStringParam(intent.Parameters, "start_date")
	endDate := getStringParam(intent.Parameters, "end_date")
	reason := getStringParam(intent.Parameters, "reason")

	details := make([]string, 0, 3)
	if startDate != "" {
		details = append(details, fmt.Sprintf("- **Mulai**: %s", startDate))
	}
	if endDate != "" {
		details = append(details, fmt.Sprintf("- **Selesai**: %s", endDate))
	}
	if reason != "" {
		details = append(details, fmt.Sprintf("- **Alasan**: %s", reason))
	}

	if len(details) == 0 {
		return "Saya akan membuat pengajuan cuti dengan detail yang Anda berikan.\n\nApakah Anda setuju?"
	}

	return fmt.Sprintf("Saya akan membuat pengajuan cuti dengan detail:\n%s\n\nApakah Anda setuju?", strings.Join(details, "\n"))
}

func (u *aiChatUsecase) buildSalesQuotationConfirmationMessage(intent *IntentResult, resolvedEntities map[string]*ResolvedEntity) string {
	customer := getStringParam(intent.Parameters, "customer_name")
	if c, ok := resolvedEntities["customer"]; ok && c.DisplayName != "" {
		customer = c.DisplayName
	}
	date := getStringParam(intent.Parameters, "quotation_date")
	if date == "" {
		date = apptime.Now().Format("2006-01-02")
	}

	details := []string{
		fmt.Sprintf("- **Customer**: %s", customer),
		fmt.Sprintf("- **Tanggal**: %s", date),
	}
	if pt := getStringParam(intent.Parameters, "payment_terms_name"); pt != "" {
		details = append(details, fmt.Sprintf("- **Syarat Pembayaran**: %s", pt))
	}
	if bu := getStringParam(intent.Parameters, "business_unit_name"); bu != "" {
		details = append(details, fmt.Sprintf("- **Unit Bisnis**: %s", bu))
	}
	if items, ok := intent.Parameters["items"].([]interface{}); ok && len(items) > 0 {
		details = append(details, fmt.Sprintf("- **Jumlah Item**: %d produk", len(items)))
	}

	return fmt.Sprintf("Saya akan membuat Sales Quotation:\n%s\n\nApakah Anda setuju?", strings.Join(details, "\n"))
}

func (u *aiChatUsecase) buildSalesTargetConfirmationMessage(intent *IntentResult, resolvedEntities map[string]*ResolvedEntity) string {
	year := getIntParam(intent.Parameters, "year")
	if year == 0 {
		year = apptime.Now().Year()
	}
	area := getStringParam(intent.Parameters, "area_name")
	if a, ok := resolvedEntities["area"]; ok && a.DisplayName != "" {
		area = a.DisplayName
	}
	if area == "" {
		area = "Semua Area"
	}
	total := getFloatParam(intent.Parameters, "total_target")
	if total <= 0 {
		total = 180000000.0
	}

	return fmt.Sprintf("Saya akan membuat Sales Target tahunan dengan detail:\n- **Area**: %s\n- **Tahun**: %d\n- **Total Target**: %.0f\n\nApakah Anda setuju?", area, year, total)
}

// buildMissingFieldsMessage generates a natural conversational follow-up
// when a CREATE action has incomplete parameters, including real DB options
func (u *aiChatUsecase) buildMissingFieldsMessage(ctx context.Context, intent *IntentResult, validation *ValidationResult) string {
	// Determine the display name for the intent
	displayName := u.getIntentDisplayName(intent.IntentCode)

	// Build a detailed summary of already-known parameters
	knownSummary := u.buildKnownParamsSummary(intent)

	// Fetch real form data options from database
	formOptions := u.entityResolver.FetchFormDataOptions(ctx, intent.IntentCode)

	// Build numbered missing-field guidance with real options
	var missingItems []string
	for i, ve := range validation.Errors {
		missingItems = append(missingItems, fmt.Sprintf("%d. %s", i+1, u.formatValidationGuidance(ve, formOptions)))
	}
	missingList := strings.Join(missingItems, "\n\n")

	return fmt.Sprintf(prompts.MissingFieldsAssistantTemplate, displayName, knownSummary, missingList)
}

// buildKnownParamsSummary builds a structured summary of parameters already provided
func (u *aiChatUsecase) buildKnownParamsSummary(intent *IntentResult) string {
	var parts []string

	switch intent.IntentCode {
	case "CREATE_SALES_QUOTATION", "CREATE_SALES_ORDER":
		if cn := getStringParam(intent.Parameters, "customer_name"); cn != "" {
			parts = append(parts, fmt.Sprintf("- **Customer**: %s", cn))
		}
		if pt := getStringParam(intent.Parameters, "payment_terms_name"); pt != "" {
			parts = append(parts, fmt.Sprintf("- **Syarat Pembayaran**: %s", pt))
		}
		if bu := getStringParam(intent.Parameters, "business_unit_name"); bu != "" {
			parts = append(parts, fmt.Sprintf("- **Unit Bisnis**: %s", bu))
		}
		if sr := getStringParam(intent.Parameters, "sales_rep_name"); sr != "" {
			parts = append(parts, fmt.Sprintf("- **Sales Rep**: %s", sr))
		}
		if qd := getStringParam(intent.Parameters, "quotation_date"); qd != "" {
			parts = append(parts, fmt.Sprintf("- **Tanggal**: %s", qd))
		}
	case "CREATE_HOLIDAY":
		if n := getStringParam(intent.Parameters, "name"); n != "" {
			parts = append(parts, fmt.Sprintf("- **Nama**: %s", n))
		}
		if d := getStringParam(intent.Parameters, "date"); d != "" {
			parts = append(parts, fmt.Sprintf("- **Tanggal**: %s", d))
		}
	case "CREATE_LEAVE_REQUEST":
		if sd := getStringParam(intent.Parameters, "start_date"); sd != "" {
			parts = append(parts, fmt.Sprintf("- **Mulai**: %s", sd))
		}
		if ed := getStringParam(intent.Parameters, "end_date"); ed != "" {
			parts = append(parts, fmt.Sprintf("- **Selesai**: %s", ed))
		}
	case "CREATE_PURCHASE_ORDER":
		if sn := getStringParam(intent.Parameters, "supplier_name"); sn != "" {
			parts = append(parts, fmt.Sprintf("- **Supplier**: %s", sn))
		}
	case "CREATE_SALES_TARGET":
		if area := getStringParam(intent.Parameters, "area_name"); area != "" {
			parts = append(parts, fmt.Sprintf("- **Area**: %s", area))
		}
		if year := getIntParam(intent.Parameters, "year"); year != 0 {
			parts = append(parts, fmt.Sprintf("- **Tahun**: %d", year))
		}
		if total := getFloatParam(intent.Parameters, "total_target"); total > 0 {
			parts = append(parts, fmt.Sprintf("- **Total Target**: %.0f", total))
		}
	}

	if len(parts) > 0 {
		return "\n\nSaya sudah mencatat:\n" + strings.Join(parts, "\n")
	}
	return ""
}

// getIntentDisplayName returns a human-readable name for an intent code
func (u *aiChatUsecase) getIntentDisplayName(intentCode string) string {
	names := map[string]string{
		"LIST_HOLIDAYS":                "Daftar Hari Libur",
		"LIST_LEAVE_REQUESTS":          "Daftar Pengajuan Cuti",
		"LIST_SALES_ORDERS":            "Daftar Sales Order",
		"LIST_SALES_QUOTATIONS":        "Daftar Sales Quotation",
		"LIST_PURCHASE_ORDERS":         "Daftar Purchase Order",
		"LIST_PURCHASE_REQUISITIONS":   "Daftar Purchase Requisition",
		"LIST_INVENTORY":               "Daftar Inventory",
		"LIST_STOCK_MOVEMENTS":         "Daftar Pergerakan Stok",
		"LIST_BANK_ACCOUNTS":           "Daftar Bank Account",
		"LIST_PAYMENTS":                "Daftar Payment",
		"LIST_SUPPLIERS":               "Daftar Supplier",
		"LIST_PRODUCTS":                "Daftar Produk",
		"LIST_WAREHOUSES":              "Daftar Gudang",
		"QUERY_STOCK":                  "Cek Stok",
		"CREATE_SALES_QUOTATION":      "Sales Quotation",
		"CREATE_SALES_ORDER":          "Sales Order",
		"CREATE_SALES_TARGET":         "Sales Target",
		"CREATE_PURCHASE_ORDER":       "Purchase Order",
		"CREATE_HOLIDAY":              "Hari Libur",
		"CREATE_LEAVE_REQUEST":        "Pengajuan Cuti",
		"CREATE_STOCK_OPNAME":         "Stock Opname",
		"CREATE_JOURNAL":              "Jurnal",
		"CREATE_PRODUCT":              "Produk",
		"CREATE_DELIVERY_ORDER":       "Delivery Order",
		"CREATE_PURCHASE_REQUISITION": "Purchase Requisition",
	}
	if name, ok := names[intentCode]; ok {
		return name
	}
	// Fallback: clean up the intent code
	clean := strings.TrimPrefix(intentCode, "CREATE_")
	clean = strings.TrimPrefix(clean, "LIST_")
	clean = strings.TrimPrefix(clean, "QUERY_")
	clean = strings.TrimPrefix(clean, "APPROVE_")
	clean = strings.TrimPrefix(clean, "REJECT_")
	clean = strings.ReplaceAll(clean, "_", " ")
	return clean
}

// formatValidationGuidance converts a ValidationError to a friendly guidance message
// enriched with real form data options from the database
func (u *aiChatUsecase) formatValidationGuidance(ve ValidationError, formOpts *FormDataOptions) string {
	switch {
	case ve.Field == "sales_quotation" && strings.Contains(ve.Message, "items"):
		guidance := "**Daftar Produk** - Produk apa saja yang ingin dimasukkan? Sebutkan nama produk, jumlah (qty), dan harganya."
		if formOpts != nil && len(formOpts.Products) > 0 {
			guidance += "\n   > Produk yang tersedia:"
			limit := 5
			if len(formOpts.Products) < limit {
				limit = len(formOpts.Products)
			}
			for _, p := range formOpts.Products[:limit] {
				if p.SKU != "" {
					guidance += fmt.Sprintf("\n   > - %s (%s)", p.Name, p.SKU)
				} else {
					guidance += fmt.Sprintf("\n   > - %s", p.Name)
				}
			}
			if len(formOpts.Products) > limit {
				guidance += fmt.Sprintf("\n   > - ...dan %d produk lainnya", len(formOpts.Products)-limit)
			}
		}
		guidance += "\n   > Contoh: \"Amoxicillin 500mg qty 100 harga 15000\""
		return guidance

	case ve.Field == "sales_quotation" && strings.Contains(ve.Message, "payment_terms"):
		guidance := "**Syarat Pembayaran** - Mau pakai syarat pembayaran apa?"
		if formOpts != nil && len(formOpts.PaymentTerms) > 0 {
			guidance += "\n   > Pilihan yang tersedia:"
			for _, pt := range formOpts.PaymentTerms {
				guidance += fmt.Sprintf("\n   > - %s", pt.Name)
			}
		} else {
			guidance += "\n   > Contoh: Net 30, COD, atau Net 60"
		}
		return guidance

	case ve.Field == "sales_quotation" && strings.Contains(ve.Message, "business_unit"):
		guidance := "**Unit Bisnis** - Quotation ini untuk unit bisnis/cabang mana?"
		if formOpts != nil && len(formOpts.BusinessUnits) > 0 {
			guidance += "\n   > Pilihan yang tersedia:"
			for _, bu := range formOpts.BusinessUnits {
				guidance += fmt.Sprintf("\n   > - %s", bu.Name)
			}
		}
		return guidance

	case ve.Field == "customer_name" && ve.Code == "REQUIRED":
		guidance := "**Customer** - Siapa customer untuk quotation ini?"
		if formOpts != nil && len(formOpts.Customers) > 0 {
			guidance += "\n   > Customer yang pernah tercatat:"
			limit := 5
			if len(formOpts.Customers) < limit {
				limit = len(formOpts.Customers)
			}
			for _, c := range formOpts.Customers[:limit] {
				guidance += fmt.Sprintf("\n   > - %s", c)
			}
			if len(formOpts.Customers) > limit {
				guidance += fmt.Sprintf("\n   > - ...dan %d customer lainnya", len(formOpts.Customers)-limit)
			}
		}
		return guidance

	case ve.Field == "sales_rep" && ve.Code == "REQUIRED":
		return "**Sales Rep** - Siapa sales representative untuk transaksi ini? Contoh: 'sales rep: Andi Saputra'."

	case ve.Field == "sales_rep_name" && ve.Code == "ENTITY_NOT_FOUND":
		return fmt.Sprintf("**Sales Rep** - %s", ve.Message)

	case ve.Field == "area_name" && ve.Code == "REQUIRED":
		return "**Area** - Mohon sebutkan area target (contoh: Bali, Jakarta, Surabaya)."

	case ve.Field == "area_name" && ve.Code == "ENTITY_NOT_FOUND":
		return fmt.Sprintf("**Area** - %s", ve.Message)

	case ve.Code == "REQUIRED":
		return fmt.Sprintf("**%s** - Mohon isi informasi ini.", ve.Field)
	case ve.Code == "INVALID_FORMAT":
		return fmt.Sprintf("**%s** - Format tidak sesuai. %s", ve.Field, ve.Message)
	case ve.Code == "INVALID_RANGE":
		return fmt.Sprintf("**%s** - %s", ve.Field, ve.Message)
	default:
		return ve.Message
	}
}

// sanitizeIfRandomRequest removes hallucinated entity params when user asks for "random" data
func (u *aiChatUsecase) sanitizeIfRandomRequest(message string, params map[string]interface{}) {
	lower := strings.ToLower(message)
	isRandom := strings.Contains(lower, "random") || strings.Contains(lower, "acak") ||
		strings.Contains(lower, "dummy") || strings.Contains(lower, "sample")
	if !isRandom {
		return
	}

	// Remove entity params that aren't literally mentioned in the user's message
	entityKeys := []string{"customer_name", "supplier_name", "employee_name", "product_name", "notes"}
	for _, key := range entityKeys {
		if val, ok := params[key].(string); ok && val != "" {
			if !strings.Contains(lower, strings.ToLower(val)) {
				delete(params, key)
			}
		}
	}

	// Remove fabricated items array if product names aren't in the message
	if items, ok := params["items"].([]interface{}); ok && len(items) > 0 {
		allFabricated := true
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				pn, _ := itemMap["product_name"].(string)
				if pn != "" && strings.Contains(lower, strings.ToLower(pn)) {
					allFabricated = false
					break
				}
			}
		}
		if allFabricated {
			delete(params, "items")
		}
	}
}

// normalizeParamNames standardizes parameter names from LLM variations to expected field names
func (u *aiChatUsecase) normalizeParamNames(intent *IntentResult) {
	params := intent.Parameters

	if intent.IntentCode == "CREATE_SALES_TARGET" {
		// Extract direct amount revisions from raw user text (e.g., "200jt", "200 juta", "Rp 200.000.000").
		if amount, ok := extractTargetAmountFromText(intent.RawMessage); ok {
			params["total_target"] = amount
		}

		// area / wilayah / region / lokasi -> area_name
		if _, has := params["area_name"]; !has {
			for _, key := range []string{"area", "wilayah", "region", "lokasi"} {
				if val, ok := params[key].(string); ok && val != "" {
					params["area_name"] = strings.TrimSpace(strings.TrimSuffix(strings.ToLower(val), " full"))
					delete(params, key)
					break
				}
			}
		}

		// tahun -> year
		if _, has := params["year"]; !has {
			if val, ok := params["tahun"]; ok {
				params["year"] = val
				delete(params, "tahun")
			}
		}

		// target / nominal_target / target_tahunan -> total_target
		if _, has := params["total_target"]; !has {
			for _, key := range []string{"target", "nominal_target", "target_tahunan", "annual_target"} {
				if val, ok := params[key]; ok {
					params["total_target"] = val
					delete(params, key)
					break
				}
			}
		}

		// Capture loose profile hints for executor defaulting logic.
		if _, has := params["target_profile"]; !has {
			for _, key := range []string{"profile", "kategori_target", "jenis_target"} {
				if val, ok := params[key].(string); ok && val != "" {
					params["target_profile"] = val
					delete(params, key)
					break
				}
			}
		}
	}

	if intent.IntentCode != "CREATE_SALES_QUOTATION" && intent.IntentCode != "CREATE_SALES_ORDER" {
		return
	}

	// payment_terms / syarat_pembayaran → payment_terms_name
	if _, has := params["payment_terms_name"]; !has {
		for _, key := range []string{"payment_terms", "syarat_pembayaran", "pembayaran"} {
			if val, ok := params[key].(string); ok && val != "" {
				params["payment_terms_name"] = val
				delete(params, key)
				break
			}
		}
	}

	// business_unit / unit_bisnis → business_unit_name
	if _, has := params["business_unit_name"]; !has {
		for _, key := range []string{"business_unit", "unit_bisnis", "cabang"} {
			if val, ok := params[key].(string); ok && val != "" {
				params["business_unit_name"] = val
				delete(params, key)
				break
			}
		}
	}

	if _, has := params["sales_rep_name"]; !has {
		for _, key := range []string{"sales_rep", "sales", "sales_name", "salesperson", "employee_name"} {
			if val, ok := params[key].(string); ok && strings.TrimSpace(val) != "" {
				params["sales_rep_name"] = strings.TrimSpace(val)
				delete(params, key)
				break
			}
		}
	}

	if intent.IntentCode == "CREATE_SALES_ORDER" {
		u.hydrateSalesOrderParamsFromRawMessage(intent)
	}
}

func (u *aiChatUsecase) hydrateSalesOrderParamsFromRawMessage(intent *IntentResult) {
	if intent == nil || intent.Parameters == nil {
		return
	}

	message := strings.TrimSpace(intent.RawMessage)
	if message == "" {
		return
	}

	params := intent.Parameters

	if strings.TrimSpace(getStringParam(params, "payment_terms_name")) == "" {
		re := regexp.MustCompile(`(?i)(payment\s*terms?|syarat\s*pembayaran)\s*[:=-]\s*([^\n,;]+)`)
		if m := re.FindStringSubmatch(message); len(m) > 2 {
			params["payment_terms_name"] = strings.TrimSpace(m[2])
		}
	}

	if strings.TrimSpace(getStringParam(params, "business_unit_name")) == "" {
		re := regexp.MustCompile(`(?i)(business|bisnis)\s*unit\s*[:=-]\s*([^\n,;]+)`)
		if m := re.FindStringSubmatch(message); len(m) > 2 {
			params["business_unit_name"] = strings.TrimSpace(m[2])
		}
	}

	if strings.TrimSpace(getStringParam(params, "sales_rep_name")) == "" {
		re := regexp.MustCompile(`(?i)(sales\s*rep|salesperson|sales|nama\s*sales)\s*[:=-]\s*([^\n,;]+)`)
		if m := re.FindStringSubmatch(message); len(m) > 2 {
			params["sales_rep_name"] = strings.TrimSpace(m[2])
		}
	}

	if _, hasItems := params["items"]; !hasItems {
		if item := parseSingleSalesOrderItemFromMessage(message); item != nil {
			params["items"] = []interface{}{item}
		}
	}
}

func parseSingleSalesOrderItemFromMessage(message string) map[string]interface{} {
	reItem := regexp.MustCompile(`(?i)items?\s*[:=-]\s*([a-zA-Z0-9\s\-/\.]+?)\s+(\d+)\s+(?:dengan\s+)?harga\s+([\w\.,]+)`)
	matches := reItem.FindStringSubmatch(message)
	if len(matches) < 4 {
		return nil
	}

	productName := strings.TrimSpace(matches[1])
	quantity, err := strconv.Atoi(strings.TrimSpace(matches[2]))
	if err != nil || quantity <= 0 {
		return nil
	}

	price, ok := parseAmountString(strings.TrimSpace(matches[3]))
	if !ok || price <= 0 {
		return nil
	}

	return map[string]interface{}{
		"product_name": productName,
		"quantity":     float64(quantity),
		"price":        price,
		"discount":     float64(0),
	}
}

func (u *aiChatUsecase) applyDeterministicIntentFallback(message string, intent *IntentResult) *IntentResult {
	if intent == nil {
		return intent
	}

	lower := strings.ToLower(strings.TrimSpace(message))
	isSalesOrder := isLikelySalesOrderCreateMessage(lower)
	isSalesQuotation := isLikelySalesQuotationCreateMessage(lower)
	isPurchaseOrder := isLikelyPurchaseOrderCreateMessage(lower)
	isLeaveRequest := isLikelyLeaveRequestCreateMessage(lower)
	isSalesTarget := isLikelySalesTargetCreateMessage(lower)
	isHoliday := isLikelyHolidayCreateMessage(lower)
	if !isSalesOrder && !isSalesQuotation && !isPurchaseOrder && !isLeaveRequest && !isSalesTarget && !isHoliday {
		return intent
	}

	if intent.IntentCode != "GENERAL_CHAT" && intent.Confidence >= 0.6 {
		return intent
	}

	if isHoliday {
		return buildHolidayFallbackIntent(message)
	}

	if isSalesOrder {
		return buildSalesOrderFallbackIntent(message)
	}

	if isSalesQuotation {
		return buildSalesQuotationFallbackIntent(message)
	}

	if isPurchaseOrder {
		return buildPurchaseOrderFallbackIntent(message)
	}

	if isLeaveRequest {
		return buildLeaveRequestFallbackIntent(message)
	}

	return buildSalesTargetFallbackIntent(message, lower)
}

func buildSalesOrderFallbackIntent(message string) *IntentResult {
	params := map[string]interface{}{}
	if customer := extractEntityNameAfterKeyword(message, "customer"); customer != "" {
		params["customer_name"] = customer
	}

	return &IntentResult{
		IntentCode: "CREATE_SALES_ORDER",
		Confidence: 0.85,
		Parameters: params,
		Module:     "sales",
		ActionType: "CREATE",
		IsQuery:    false,
		RawMessage: message,
	}
}

func buildSalesQuotationFallbackIntent(message string) *IntentResult {
	params := map[string]interface{}{}
	if customer := extractEntityNameAfterKeyword(message, "customer"); customer != "" {
		params["customer_name"] = customer
	}

	return &IntentResult{
		IntentCode: "CREATE_SALES_QUOTATION",
		Confidence: 0.85,
		Parameters: params,
		Module:     "sales",
		ActionType: "CREATE",
		IsQuery:    false,
		RawMessage: message,
	}
}

func buildPurchaseOrderFallbackIntent(message string) *IntentResult {
	params := map[string]interface{}{}
	if supplier := extractEntityNameAfterKeyword(message, "supplier"); supplier != "" {
		params["supplier_name"] = supplier
	}

	return &IntentResult{
		IntentCode: "CREATE_PURCHASE_ORDER",
		Confidence: 0.85,
		Parameters: params,
		Module:     "purchase",
		ActionType: "CREATE",
		IsQuery:    false,
		RawMessage: message,
	}
}

func buildLeaveRequestFallbackIntent(message string) *IntentResult {
	return &IntentResult{
		IntentCode: "CREATE_LEAVE_REQUEST",
		Confidence: 0.8,
		Parameters: map[string]interface{}{},
		Module:     "hrd",
		ActionType: "CREATE",
		IsQuery:    false,
		RawMessage: message,
	}
}

func buildHolidayFallbackIntent(message string) *IntentResult {
	params := make(map[string]interface{})
	if year, ok := extractYearOrCurrent(message); ok {
		params["year"] = year
	}

	lower := strings.ToLower(message)
	if strings.Contains(lower, "indonesia") || strings.Contains(lower, "nasional") {
		params["country_code"] = "ID"
		params["holiday_source"] = "PUBLIC_API"
	}

	return &IntentResult{
		IntentCode: "CREATE_HOLIDAY",
		Confidence: 0.8,
		Parameters: params,
		Module:     "hrd",
		ActionType: "CREATE",
		IsQuery:    false,
		RawMessage: message,
	}
}

func buildSalesTargetFallbackIntent(message, lowerMessage string) *IntentResult {
	params := make(map[string]interface{})
	if area := extractAreaName(message); area != "" {
		params["area_name"] = area
	}

	if year, ok := extractYearOrCurrent(message); ok {
		params["year"] = year
	}

	if strings.Contains(lowerMessage, "normal") || strings.Contains(lowerMessage, "farmasi") {
		params["target_profile"] = "normal_sales_farmasi"
		params["total_target"] = 180000000
	}

	return &IntentResult{
		IntentCode: "CREATE_SALES_TARGET",
		Confidence: 0.8,
		Parameters: params,
		Module:     "sales",
		ActionType: "CREATE",
		IsQuery:    false,
		RawMessage: message,
	}
}

func extractAreaName(message string) string {
	areaRegex := regexp.MustCompile(`(?i)area\s+([a-zA-Z\s]+)`)
	matches := areaRegex.FindStringSubmatch(message)
	if len(matches) <= 1 {
		return ""
	}

	areaRaw := strings.TrimSpace(matches[1])
	areaRaw = strings.TrimSpace(strings.ReplaceAll(strings.ToLower(areaRaw), " full", ""))
	for _, stop := range []string{" untuk", " dengan", " target", " normal", " sales", " farmasi"} {
		if idx := strings.Index(areaRaw, stop); idx > 0 {
			areaRaw = strings.TrimSpace(areaRaw[:idx])
		}
	}

	return areaRaw
}

func currentDatetimeContextMessage() string {
	now := apptime.Now()
	return fmt.Sprintf("CURRENT_DATETIME_WIB: %s (year=%d). Gunakan konteks ini saat pengguna bertanya 'sekarang', 'today', atau 'tahun ini'.", now.Format(time.RFC3339), now.Year())
}

func isLikelySalesTargetCreateMessage(lowerMessage string) bool {
	hasCreateVerb := strings.Contains(lowerMessage, "buat") || strings.Contains(lowerMessage, "create") || strings.Contains(lowerMessage, "tambahkan")
	hasTarget := strings.Contains(lowerMessage, "target") && (strings.Contains(lowerMessage, "sales") || strings.Contains(lowerMessage, "penjualan"))
	hasSalesTargetPhrase := strings.Contains(lowerMessage, "sales target") || strings.Contains(lowerMessage, "target sales")

	return (hasCreateVerb && hasTarget) || hasSalesTargetPhrase
}

func isLikelySalesOrderCreateMessage(lowerMessage string) bool {
	hasCreateVerb := strings.Contains(lowerMessage, "buat") || strings.Contains(lowerMessage, "create") || strings.Contains(lowerMessage, "tambahkan")
	hasSalesOrderPhrase := strings.Contains(lowerMessage, "sales order") || strings.Contains(lowerMessage, "order penjualan") || strings.Contains(lowerMessage, "so ") || strings.HasSuffix(lowerMessage, " so")
	return hasCreateVerb && hasSalesOrderPhrase
}

func isLikelySalesQuotationCreateMessage(lowerMessage string) bool {
	hasCreateVerb := strings.Contains(lowerMessage, "buat") || strings.Contains(lowerMessage, "create") || strings.Contains(lowerMessage, "tambahkan")
	hasQuotationPhrase := strings.Contains(lowerMessage, "sales quotation") || strings.Contains(lowerMessage, "quotation") || strings.Contains(lowerMessage, "penawaran")
	return hasCreateVerb && hasQuotationPhrase
}

func isLikelyPurchaseOrderCreateMessage(lowerMessage string) bool {
	hasCreateVerb := strings.Contains(lowerMessage, "buat") || strings.Contains(lowerMessage, "create") || strings.Contains(lowerMessage, "tambahkan")
	hasPurchaseOrderPhrase := strings.Contains(lowerMessage, "purchase order") || strings.Contains(lowerMessage, "order pembelian") || strings.Contains(lowerMessage, "po ") || strings.HasSuffix(lowerMessage, " po")
	return hasCreateVerb && hasPurchaseOrderPhrase
}

func isLikelyLeaveRequestCreateMessage(lowerMessage string) bool {
	hasCreateVerb := strings.Contains(lowerMessage, "buat") || strings.Contains(lowerMessage, "create") || strings.Contains(lowerMessage, "ajukan") || strings.Contains(lowerMessage, "submit")
	hasLeavePhrase := strings.Contains(lowerMessage, "cuti") || strings.Contains(lowerMessage, "leave request")
	return hasCreateVerb && hasLeavePhrase
}

func extractEntityNameAfterKeyword(message, keyword string) string {
	pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(keyword) + `\s+([a-zA-Z0-9\s\.-]+)`)
	matches := pattern.FindStringSubmatch(message)
	if len(matches) <= 1 {
		return ""
	}

	name := strings.TrimSpace(matches[1])
	lower := strings.ToLower(name)
	for _, stop := range []string{" dengan", " untuk", " tanggal", " pada", " dan", " yg", " yang"} {
		if idx := strings.Index(lower, stop); idx > 0 {
			name = strings.TrimSpace(name[:idx])
			break
		}
	}

	return strings.TrimSpace(name)
}

func isLikelyHolidayCreateMessage(lowerMessage string) bool {
	hasCreateVerb := strings.Contains(lowerMessage, "buat") || strings.Contains(lowerMessage, "create") || strings.Contains(lowerMessage, "tambahkan") || strings.Contains(lowerMessage, "generate")
	hasHolidayTerm := strings.Contains(lowerMessage, "holiday") || strings.Contains(lowerMessage, "hari libur") || strings.Contains(lowerMessage, "libur")
	return hasCreateVerb && hasHolidayTerm
}

func extractYearOrCurrent(message string) (int, bool) {
	yearRegex := regexp.MustCompile(`\b(20[2-9][0-9]|2100)\b`)
	if year := yearRegex.FindString(message); year != "" {
		if parsed, err := strconv.Atoi(year); err == nil {
			return parsed, true
		}
	}

	lower := strings.ToLower(message)
	if strings.Contains(lower, "tahun ini") || strings.Contains(lower, "this year") {
		return apptime.Now().Year(), true
	}

	return 0, false
}

func isCurrentTimeQuestion(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return false
	}

	actionKeywords := []string{
		"buat", "buatkan", "create", "tambahkan", "generate", "input", "simpan", "insert", "tambah",
	}
	for _, kw := range actionKeywords {
		if strings.Contains(lower, kw) {
			return false
		}
	}

	keywords := []string{
		"jam berapa", "waktu sekarang", "waktu saat ini", "sekarang jam", "tanggal hari ini", "hari ini tanggal",
		"tahun berapa", "tahun sekarang", "tahun ini", "what time", "current time", "current date", "what year",
	}

	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	return false
}

func buildAutonomousHolidayCreateIntent(message string, history []models.AIChatMessage) *IntentResult {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return nil
	}

	directCreateCommand := isLikelyHolidayCreateMessage(lower)
	followUpCreateCommand := isCreateFollowUpCommand(lower) && hasRecentHolidayContext(history)
	if !directCreateCommand && !followUpCreateCommand {
		return nil
	}

	year := apptime.Now().Year()
	if parsedYear, ok := extractYearOrCurrent(message); ok {
		year = parsedYear
	}

	params := map[string]interface{}{
		"year":           year,
		"country_code":   "ID",
		"holiday_source": "PUBLIC_API",
	}

	return &IntentResult{
		IntentCode: "CREATE_HOLIDAY",
		Confidence: 1,
		Parameters: params,
		Module:     "hrd",
		ActionType: "CREATE",
		IsQuery:    false,
		RawMessage: message,
	}
}

func isCreateFollowUpCommand(lowerMessage string) bool {
	followUps := []string{
		"ya buat", "ya lanjut", "lanjutkan", "buat untuk saya", "buatkan untuk saya", "ok buat", "oke buat", "proceed",
	}
	for _, phrase := range followUps {
		if strings.Contains(lowerMessage, phrase) {
			return true
		}
	}
	return false
}

func hasRecentHolidayContext(history []models.AIChatMessage) bool {
	if len(history) == 0 {
		return false
	}

	start := len(history) - 6
	if start < 0 {
		start = 0
	}

	for i := len(history) - 1; i >= start; i-- {
		content := strings.ToLower(history[i].Content)
		if strings.Contains(content, "holiday") || strings.Contains(content, "hari libur") || strings.Contains(content, "indonesia") {
			return true
		}
	}

	return false
}

func extractTargetAmountFromText(message string) (float64, bool) {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return 0, false
	}

	// Strong pattern for shorthand and rupiah expression
	amountPattern := regexp.MustCompile(`(?i)(rp\s*)?\d+[\d\.,]*\s*(jt|juta|m)?`)
	candidate := amountPattern.FindString(msg)
	if candidate == "" {
		return 0, false
	}

	value, ok := parseAmountString(candidate)
	if !ok || value <= 0 {
		return 0, false
	}

	return value, true
}

// mergeWithPreviousIntentParams carries forward parameters from a previous same-intent
// assistant message, enabling multi-turn conversations for CREATE actions
func (u *aiChatUsecase) mergeWithPreviousIntentParams(history []models.AIChatMessage, intent *IntentResult) {
	if intent.ActionType != "CREATE" {
		return
	}

	// Search backwards for the last assistant message with the same intent
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		if msg.Role != models.MessageRoleAssistant || msg.Intent == nil {
			continue
		}
		var prevIntent IntentResult
		if err := json.Unmarshal([]byte(*msg.Intent), &prevIntent); err != nil {
			continue
		}
		if prevIntent.IntentCode != intent.IntentCode {
			break // Only merge with the immediately preceding same-intent exchange
		}

		// Merge: previous params as base, current params override (skip empty)
		merged := make(map[string]interface{})
		for k, v := range prevIntent.Parameters {
			if !isEmptyParamValue(v) {
				merged[k] = v
			}
		}
		for k, v := range intent.Parameters {
			if !isEmptyParamValue(v) {
				merged[k] = v
			}
		}
		intent.Parameters = merged
		return
	}
}

// isEmptyParamValue checks if a parameter value is effectively empty
func isEmptyParamValue(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case []interface{}:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	}
	return false
}
