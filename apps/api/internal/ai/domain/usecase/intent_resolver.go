package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/ai/data/repositories"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase/prompts"
	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/cerebras"
)

// IntentResult represents the structured result of intent extraction from user message
type IntentResult struct {
	IntentCode string                 `json:"intent_code"`
	Confidence float64                `json:"confidence"`
	Parameters map[string]interface{} `json:"parameters"`
	Module     string                 `json:"module"`
	ActionType string                 `json:"action_type"`
	IsQuery    bool                   `json:"is_query"`
	RawMessage string                 `json:"raw_message"`
}

// Layer 1: Intent Classifier — uses cheapest model for fast classification only
const intentClassifierModel = "llama3.1-8b"

// IntentResolver extracts structured intents from natural language messages using LLM
type IntentResolver struct {
	cerebrasClient *cerebras.Client
	intentRepo     repositories.IntentRegistryRepository
	// Cached intent list to avoid DB query every message
	cachedIntentPrompt string
	cachedIntentCodes  map[string]struct{}
	cacheMu            sync.RWMutex
	cacheExpiry        time.Time
}

// NewIntentResolver creates a new IntentResolver
func NewIntentResolver(client *cerebras.Client, intentRepo repositories.IntentRegistryRepository) *IntentResolver {
	return &IntentResolver{
		cerebrasClient: client,
		intentRepo:     intentRepo,
	}
}

// Resolve classifies user intent using a lightweight model (Layer 1)
func (r *IntentResolver) Resolve(ctx context.Context, userMessage string, conversationHistory []cerebras.ChatMessage) (*IntentResult, error) {
	systemPrompt, intentCodes, err := r.getOrRefreshIntentPrompt(ctx)
	if err != nil {
		return nil, fmt.Errorf("AI_INTENT_RESOLUTION_FAILED: failed to build intent context: %w", err)
	}

	messages := make([]cerebras.ChatMessage, 0, len(conversationHistory)+2)
	messages = append(messages, cerebras.ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})
	messages = append(messages, cerebras.ChatMessage{
		Role:    "system",
		Content: fmt.Sprintf("CURRENT_DATETIME_WIB: %s (year=%d). Use this when user says 'today', 'sekarang', 'tahun ini'.", apptime.Now().Format(time.RFC3339), apptime.Now().Year()),
	})

	// Include recent conversation history for context (max 4 messages for speed)
	historyLimit := 4
	if len(conversationHistory) < historyLimit {
		historyLimit = len(conversationHistory)
	}
	if historyLimit > 0 {
		messages = append(messages, conversationHistory[len(conversationHistory)-historyLimit:]...)
	}

	messages = append(messages, cerebras.ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	// Use the cheapest/fastest model for intent classification
	resp, err := r.cerebrasClient.Chat(&cerebras.ChatRequest{
		Messages:    messages,
		Model:       intentClassifierModel,
		MaxTokens:   256,
		Temperature: 0.05,
	})
	if err != nil {
		return nil, fmt.Errorf("AI_INTENT_RESOLUTION_FAILED: LLM call failed: %w", err)
	}

	result, err := r.parseIntentResponse(resp.Message.Content, userMessage)
	if err != nil {
		return nil, fmt.Errorf("AI_INTENT_RESOLUTION_FAILED: failed to parse intent: %w", err)
	}

	result = sanitizeIntentResult(result, intentCodes)

	return result, nil
}

// getOrRefreshIntentPrompt returns cached prompt or builds a fresh one
func (r *IntentResolver) getOrRefreshIntentPrompt(ctx context.Context) (string, map[string]struct{}, error) {
	r.cacheMu.RLock()
	if r.cachedIntentPrompt != "" && len(r.cachedIntentCodes) > 0 && apptime.Now().Before(r.cacheExpiry) {
		prompt := r.cachedIntentPrompt
		codes := cloneIntentCodeSet(r.cachedIntentCodes)
		r.cacheMu.RUnlock()
		return prompt, codes, nil
	}
	r.cacheMu.RUnlock()

	// Fetch and rebuild
	activeIntents, err := r.intentRepo.FindActive(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch active intents: %w", err)
	}

	prompt := r.buildSystemPrompt(activeIntents)
	intentCodes := r.buildIntentCodeSet(activeIntents)

	r.cacheMu.Lock()
	r.cachedIntentPrompt = prompt
	r.cachedIntentCodes = cloneIntentCodeSet(intentCodes)
	r.cacheExpiry = apptime.Now().Add(5 * time.Minute)
	r.cacheMu.Unlock()

	return prompt, intentCodes, nil
}

// buildSystemPrompt constructs a concise classification prompt (Layer 1 only classifies, no param extraction)
func (r *IntentResolver) buildSystemPrompt(intents interface{}) string {
	intentList := r.formatIntentList(intents)
	return fmt.Sprintf(prompts.IntentClassifierTemplate, intentList)
}

// formatIntentList formats active intents into a compact list for the LLM prompt
func (r *IntentResolver) formatIntentList(intents interface{}) string {
	type intentInfo struct {
		IntentCode  string
		DisplayName string
		Description string
		Module      string
		ActionType  string
		Parameters  string
	}

	var intentInfos []intentInfo

	data, err := json.Marshal(intents)
	if err != nil {
		return "- GENERAL_CHAT: General conversation (module: general, action: QUERY)"
	}
	var rawIntents []map[string]interface{}
	if err := json.Unmarshal(data, &rawIntents); err != nil {
		return "- GENERAL_CHAT: General conversation (module: general, action: QUERY)"
	}

	for _, ri := range rawIntents {
		info := intentInfo{
			IntentCode:  fmt.Sprintf("%v", ri["intent_code"]),
			DisplayName: fmt.Sprintf("%v", ri["display_name"]),
			Description: fmt.Sprintf("%v", ri["description"]),
			Module:      fmt.Sprintf("%v", ri["module"]),
			ActionType:  fmt.Sprintf("%v", ri["action_type"]),
		}
		if ps, ok := ri["parameter_schema"]; ok && ps != nil {
			paramData, _ := json.Marshal(ps)
			info.Parameters = string(paramData)
		}
		intentInfos = append(intentInfos, info)
	}

	if len(intentInfos) == 0 {
		return "- GENERAL_CHAT: General conversation (module: general, action: QUERY)"
	}

	var sb strings.Builder
	for _, info := range intentInfos {
		sb.WriteString(fmt.Sprintf("- %s: %s (module: %s, action: %s)\n",
			info.IntentCode, info.DisplayName, info.Module, info.ActionType))
	}
	sb.WriteString("- GENERAL_CHAT: General conversation (module: general, action: QUERY)\n")

	return sb.String()
}

func (r *IntentResolver) buildIntentCodeSet(intents interface{}) map[string]struct{} {
	codes := map[string]struct{}{
		"GENERAL_CHAT": {},
	}

	data, err := json.Marshal(intents)
	if err != nil {
		return codes
	}

	var rawIntents []map[string]interface{}
	if err := json.Unmarshal(data, &rawIntents); err != nil {
		return codes
	}

	for _, ri := range rawIntents {
		code := strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", ri["intent_code"])))
		if code == "" || code == "<NIL>" {
			continue
		}
		codes[code] = struct{}{}
	}

	return codes
}

// parseIntentResponse parses the LLM JSON response into a structured IntentResult
func (r *IntentResolver) parseIntentResponse(llmResponse string, originalMessage string) (*IntentResult, error) {
	cleaned := strings.TrimSpace(llmResponse)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	startIdx := strings.Index(cleaned, "{")
	endIdx := strings.LastIndex(cleaned, "}")

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return &IntentResult{
			IntentCode: "GENERAL_CHAT",
			Confidence: 0.5,
			Parameters: map[string]interface{}{},
			Module:     "general",
			ActionType: "QUERY",
			IsQuery:    true,
			RawMessage: originalMessage,
		}, nil
	}

	jsonStr := cleaned[startIdx : endIdx+1]
	var result IntentResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &IntentResult{
			IntentCode: "GENERAL_CHAT",
			Confidence: 0.5,
			Parameters: map[string]interface{}{},
			Module:     "general",
			ActionType: "QUERY",
			IsQuery:    true,
			RawMessage: originalMessage,
		}, nil
	}

	if result.Parameters == nil {
		result.Parameters = map[string]interface{}{}
	}
	result.RawMessage = originalMessage

	if result.Confidence < 0 {
		result.Confidence = 0
	}
	if result.Confidence > 1 {
		result.Confidence = 1
	}

	return &result, nil
}

func sanitizeIntentResult(result *IntentResult, intentCodes map[string]struct{}) *IntentResult {
	if result == nil {
		return fallbackGeneralIntent("")
	}

	if result.Parameters == nil {
		result.Parameters = map[string]interface{}{}
	}

	normalized := strings.ToUpper(strings.TrimSpace(result.IntentCode))
	if normalized == "" {
		fallback := fallbackGeneralIntent(result.RawMessage)
		fallback.Parameters = result.Parameters
		return fallback
	}

	if _, ok := intentCodes[normalized]; ok {
		result.IntentCode = normalized
		return result
	}

	if recovered, ok := recoverIntentCode(normalized, intentCodes); ok {
		result.IntentCode = recovered
		if result.Confidence > 0.55 {
			result.Confidence = 0.55
		}
		return result
	}

	fallback := fallbackGeneralIntent(result.RawMessage)
	fallback.Parameters = result.Parameters
	if result.Confidence > 0 && result.Confidence < 0.4 {
		fallback.Confidence = result.Confidence
	}
	return fallback
}

func recoverIntentCode(intentCode string, intentCodes map[string]struct{}) (string, bool) {
	parts := strings.Split(intentCode, "_")
	for i := len(parts) - 1; i >= 2; i-- {
		candidate := strings.Join(parts[:i], "_")
		if _, ok := intentCodes[candidate]; ok {
			return candidate, true
		}
	}
	return "", false
}

func fallbackGeneralIntent(message string) *IntentResult {
	return &IntentResult{
		IntentCode: "GENERAL_CHAT",
		Confidence: 0.4,
		Parameters: map[string]interface{}{},
		Module:     "general",
		ActionType: "QUERY",
		IsQuery:    true,
		RawMessage: message,
	}
}

func cloneIntentCodeSet(in map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for k := range in {
		out[k] = struct{}{}
	}
	return out
}
