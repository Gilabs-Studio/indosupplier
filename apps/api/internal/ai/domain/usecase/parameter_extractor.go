package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/ai/data/repositories"
	"github.com/gilabs/gims/api/internal/ai/domain/usecase/prompts"
	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/cerebras"
)

// Layer 2: Parameter Extractor — uses a focused prompt to extract structured parameters
// Only invoked for non-GENERAL_CHAT intents to minimize LLM cost

// ParameterExtractor extracts structured parameters for a classified intent
type ParameterExtractor struct {
	cerebrasClient *cerebras.Client
	intentRepo     repositories.IntentRegistryRepository
}

// NewParameterExtractor creates a new ParameterExtractor
func NewParameterExtractor(client *cerebras.Client, intentRepo repositories.IntentRegistryRepository) *ParameterExtractor {
	return &ParameterExtractor{
		cerebrasClient: client,
		intentRepo:     intentRepo,
	}
}

// Extract refines parameter extraction for a given intent using a targeted LLM prompt
func (p *ParameterExtractor) Extract(ctx context.Context, intent *IntentResult, userMessage string, conversationHistory []cerebras.ChatMessage) (map[string]interface{}, error) {
	// Fetch the schema for the specific intent from the registry
	schema, err := p.getParameterSchema(ctx, intent.IntentCode)
	if err != nil || schema == "" {
		// If no schema defined, return whatever the classifier already extracted
		return intent.Parameters, nil
	}

	// Build a focused extraction prompt
	systemPrompt := fmt.Sprintf(prompts.ParameterExtractionTemplate, intent.IntentCode, schema)

	messages := make([]cerebras.ChatMessage, 0, 5)
	messages = append(messages, cerebras.ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})
	messages = append(messages, cerebras.ChatMessage{
		Role:    "system",
		Content: fmt.Sprintf("CURRENT_DATETIME_WIB: %s (year=%d). Use this when user says 'today', 'sekarang', 'tahun ini'.", apptime.Now().Format("2006-01-02T15:04:05-07:00"), apptime.Now().Year()),
	})

	// Include last 6 conversation messages for better multi-turn context
	historyLimit := 6
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

	resp, err := p.cerebrasClient.Chat(&cerebras.ChatRequest{
		Messages:    messages,
		Model:       intentClassifierModel, // Use the same cheap model for parameter extraction
		MaxTokens:   256,
		Temperature: 0.05,
	})
	if err != nil {
		// Fallback to intent resolver's parameters on failure
		return intent.Parameters, nil
	}

	params := p.parseParameterResponse(resp.Message.Content)
	if len(params) == 0 {
		return intent.Parameters, nil
	}

	// Merge: extracted params override, but keep any intent resolver params as fallback
	merged := make(map[string]interface{})
	for k, v := range intent.Parameters {
		merged[k] = v
	}
	for k, v := range params {
		merged[k] = v
	}

	return merged, nil
}

// getParameterSchema fetches the parameter schema for an intent from the registry
func (p *ParameterExtractor) getParameterSchema(ctx context.Context, intentCode string) (string, error) {
	intent, err := p.intentRepo.FindByIntentCode(ctx, intentCode)
	if err != nil {
		return "", err
	}
	if intent.ParameterSchema != nil {
		return *intent.ParameterSchema, nil
	}
	return "", nil
}

// parseParameterResponse parses the LLM response into a parameter map
func (p *ParameterExtractor) parseParameterResponse(response string) map[string]interface{} {
	cleaned := strings.TrimSpace(response)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	startIdx := strings.Index(cleaned, "{")
	endIdx := strings.LastIndex(cleaned, "}")
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned[startIdx:endIdx+1]), &params); err != nil {
		return nil
	}

	return params
}
