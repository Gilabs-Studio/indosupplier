// Package registration bridges the existing ActionExecutor and IntentRegistry
// into the new Tool system. It reads intent definitions from the database and
// creates Tool instances that delegate execution to the existing action handlers.
//
// This is the migration bridge: all 80+ existing actions become tools without
// rewriting their business logic. New tools can be implemented as native Tool
// implementations over time.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/ai/data/models"
	"github.com/gilabs/gims/api/internal/ai/data/repositories"
)

// ActionExecutorFunc is the signature for the legacy action executor dispatch.
type ActionExecutorFunc func(ctx context.Context, intentCode string, params map[string]interface{}, userID string, resolvedEntities map[string]*ResolvedEntity) *ToolResult

// EntityResolverFunc resolves entity names to IDs from parameters.
type EntityResolverFunc func(ctx context.Context, params map[string]interface{}) (map[string]*ResolvedEntity, error)

// RegisterFromIntentRegistry reads all active intents from the database and registers
// them as AdaptedTool instances in the registry. This is called once at startup.
func RegisterFromIntentRegistry(
	ctx context.Context,
	registry *Registry,
	intentRepo repositories.IntentRegistryRepository,
	executorFn ActionExecutorFunc,
	resolverFn EntityResolverFunc,
) error {
	intents, err := intentRepo.FindActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to load intent registry for tool registration: %w", err)
	}

	for _, intent := range intents {
		tool := intentToTool(intent, executorFn, resolverFn)
		if tool != nil {
			registry.Register(tool)
		}
	}

	return nil
}

// intentToTool converts an AIIntentRegistry entry to an AdaptedTool.
func intentToTool(intent models.AIIntentRegistry, executorFn ActionExecutorFunc, resolverFn EntityResolverFunc) *AdaptedTool {
	// Parse parameter schema from JSONB
	params := parseParameterSchema(intent.ParameterSchema)

	// Determine category from action type
	category := strings.ToLower(intent.ActionType)
	isReadOperation := category == "query" || category == "list" || category == "read" || category == "get"
	concurrencySafe := isReadOperation

	// Read operations never require confirmation regardless of DB setting — only
	// write operations (create, update, delete, approve, execute) may require it.
	needsConfirmation := intent.RequiresConfirmation && !isReadOperation

	// Build the tool name: convert INTENT_CODE to lowercase with underscores
	toolName := strings.ToLower(intent.IntentCode)

	return NewAdaptedTool(AdaptedToolConfig{
		Name:              toolName,
		Description:       buildToolDescription(intent),
		Module:            intent.Module,
		Category:          category,
		Permission:        intent.RequiredPermission,
		NeedsConfirmation: needsConfirmation,
		ConcurrencySafe:   concurrencySafe,
		Params:            params,
		ExecuteFn: func(ctx context.Context, callParams map[string]interface{}, execCtx *ExecutionContext) (*ToolResult, error) {
			// Resolve entities from natural language names to database IDs
			var resolved map[string]*ResolvedEntity
			if resolverFn != nil {
				resolved, _ = resolverFn(ctx, callParams)
			}

			// Delegate to the existing action executor
			result := executorFn(ctx, intent.IntentCode, callParams, execCtx.UserID, resolved)
			if result == nil {
				return nil, fmt.Errorf("action executor returned nil for %s", intent.IntentCode)
			}

			return result, nil
		},
	})
}

// buildToolDescription creates a clear description for the LLM.
func buildToolDescription(intent models.AIIntentRegistry) string {
	if intent.Description != "" {
		return intent.Description
	}
	return intent.DisplayName
}

// parseParameterSchema converts the JSONB parameter schema into ToolParameter definitions.
func parseParameterSchema(schemaStr *string) []ToolParameter {
	if schemaStr == nil || *schemaStr == "" {
		return nil
	}

	// Try parsing as a JSON object with "properties" (JSON Schema format)
	var jsonSchema struct {
		Properties map[string]struct {
			Type        string   `json:"type"`
			Description string   `json:"description"`
			Enum        []string `json:"enum"`
		} `json:"properties"`
		Required []string `json:"required"`
	}
	if err := json.Unmarshal([]byte(*schemaStr), &jsonSchema); err == nil && len(jsonSchema.Properties) > 0 {
		requiredSet := make(map[string]bool)
		for _, r := range jsonSchema.Required {
			requiredSet[r] = true
		}

		params := make([]ToolParameter, 0, len(jsonSchema.Properties))
		for name, prop := range jsonSchema.Properties {
			params = append(params, ToolParameter{
				Name:        name,
				Type:        prop.Type,
				Description: prop.Description,
				Required:    requiredSet[name],
				Enum:        prop.Enum,
			})
		}
		return params
	}

	// Try parsing as a simple key-value map (legacy format: {"param_name": "description"})
	var simpleSchema map[string]string
	if err := json.Unmarshal([]byte(*schemaStr), &simpleSchema); err == nil {
		params := make([]ToolParameter, 0, len(simpleSchema))
		for name, desc := range simpleSchema {
			pType := inferParameterType(name)
			params = append(params, ToolParameter{
				Name:        name,
				Type:        pType,
				Description: desc,
			})
		}
		return params
	}

	return nil
}

// inferParameterType guesses the parameter type from its name.
func inferParameterType(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, "_id"):
		return "uuid"
	case strings.HasSuffix(lower, "_date") || lower == "date" || lower == "start_date" || lower == "end_date":
		return "date"
	case strings.Contains(lower, "amount") || strings.Contains(lower, "price") || strings.Contains(lower, "qty") || strings.Contains(lower, "quantity") || strings.Contains(lower, "count"):
		return "number"
	case lower == "low_stock" || strings.HasPrefix(lower, "is_") || strings.HasPrefix(lower, "has_"):
		return "boolean"
	default:
		return "string"
	}
}
