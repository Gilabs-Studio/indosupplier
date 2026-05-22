package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/ai/data/repositories"
)

// PermissionValidationResult holds the result of a permission check
type PermissionValidationResult struct {
	Allowed            bool   `json:"allowed"`
	RequiredPermission string `json:"required_permission"`
	Reason             string `json:"reason,omitempty"`
}

// PermissionValidator checks whether a user has the required permission for an AI intent
type PermissionValidator struct {
	intentRepo repositories.IntentRegistryRepository
}

// NewPermissionValidator creates a new PermissionValidator
func NewPermissionValidator(intentRepo repositories.IntentRegistryRepository) *PermissionValidator {
	return &PermissionValidator{
		intentRepo: intentRepo,
	}
}

// Validate checks if the user has permission to execute the resolved intent
func (v *PermissionValidator) Validate(ctx context.Context, intentCode string, userPermissions map[string]bool, isAdmin bool) (*PermissionValidationResult, error) {
	intentCode = strings.ToUpper(strings.TrimSpace(intentCode))

	// General chat does not require specific permissions
	if intentCode == "GENERAL_CHAT" {
		return &PermissionValidationResult{
			Allowed:            true,
			RequiredPermission: "",
			Reason:             "",
		}, nil
	}

	// Look up the intent in the registry
	intent, err := v.intentRepo.FindByIntentCode(ctx, intentCode)
	if err != nil {
		if fallback, ok := fallbackPermissionForIntent(intentCode); ok {
			if isAdmin {
				return &PermissionValidationResult{
					Allowed:            true,
					RequiredPermission: fallback,
					Reason:             "admin bypass (fallback policy)",
				}, nil
			}

			if fallback == "" || userPermissions[fallback] {
				return &PermissionValidationResult{
					Allowed:            true,
					RequiredPermission: fallback,
					Reason:             "allowed by fallback policy",
				}, nil
			}

			return &PermissionValidationResult{
				Allowed:            false,
				RequiredPermission: fallback,
				Reason:             fmt.Sprintf("You do not have the '%s' permission required for this action", fallback),
			}, nil
		}

		return nil, fmt.Errorf("AI_PERMISSION_CHECK_FAILED: intent not found in registry: %w", err)
	}

	// Admin users bypass permission checks, but only for registered intents.
	if isAdmin {
		return &PermissionValidationResult{
			Allowed:            true,
			RequiredPermission: "",
			Reason:             "admin bypass",
		}, nil
	}

	if !intent.IsActive {
		return &PermissionValidationResult{
			Allowed:            false,
			RequiredPermission: intent.RequiredPermission,
			Reason:             fmt.Sprintf("Intent '%s' is currently disabled", intent.DisplayName),
		}, nil
	}

	requiredPerm := intent.RequiredPermission
	if requiredPerm == "" {
		// No specific permission required for this intent
		return &PermissionValidationResult{
			Allowed:            true,
			RequiredPermission: "",
			Reason:             "",
		}, nil
	}

	// Check if user has the required permission
	if userPermissions[requiredPerm] {
		return &PermissionValidationResult{
			Allowed:            true,
			RequiredPermission: requiredPerm,
			Reason:             "",
		}, nil
	}

	return &PermissionValidationResult{
		Allowed:            false,
		RequiredPermission: requiredPerm,
		Reason:             fmt.Sprintf("You do not have the '%s' permission required for this action", requiredPerm),
	}, nil
}

// NeedsConfirmation checks whether the intent requires user confirmation before execution
func (v *PermissionValidator) NeedsConfirmation(ctx context.Context, intentCode string) (bool, error) {
	intentCode = strings.ToUpper(strings.TrimSpace(intentCode))

	if intentCode == "GENERAL_CHAT" {
		return false, nil
	}

	intent, err := v.intentRepo.FindByIntentCode(ctx, intentCode)
	if err != nil {
		// If intent is not found, require confirmation as a safety measure
		return true, nil
	}

	return intent.RequiresConfirmation, nil
}

func fallbackPermissionForIntent(intentCode string) (string, bool) {
	// Explicit rules for intents with non-standard or empty permissions.
	explicit := map[string]string{
		"CREATE_HOLIDAY":       "holiday.create",
		"LIST_HOLIDAYS":        "holiday.read",
		"CREATE_LEAVE_REQUEST": "",
		"LIST_LEAVE_REQUESTS":  "",
		"GENERATE_REPORT":      "report.generate",
	}

	if perm, ok := explicit[intentCode]; ok {
		return perm, true
	}

	return inferPermissionFromIntentCode(intentCode)
}

func inferPermissionFromIntentCode(intentCode string) (string, bool) {
	intentCode = strings.ToUpper(strings.TrimSpace(intentCode))
	if intentCode == "" {
		return "", false
	}

	infer := func(rawEntity, actionSuffix string) (string, bool) {
		entity := strings.TrimSpace(rawEntity)
		if entity == "" {
			return "", false
		}

		// LIST/QUERY intents are usually plural; normalize to singular permission resource.
		if strings.HasSuffix(entity, "S") {
			entity = strings.TrimSuffix(entity, "S")
		}

		entity = strings.ToLower(entity)
		if entity == "" {
			return "", false
		}

		return entity + "." + actionSuffix, true
	}

	switch {
	case strings.HasPrefix(intentCode, "CREATE_"):
		return infer(strings.TrimPrefix(intentCode, "CREATE_"), "create")
	case strings.HasPrefix(intentCode, "LIST_"):
		return infer(strings.TrimPrefix(intentCode, "LIST_"), "read")
	case strings.HasPrefix(intentCode, "QUERY_"):
		return infer(strings.TrimPrefix(intentCode, "QUERY_"), "read")
	case strings.HasPrefix(intentCode, "APPROVE_"):
		return infer(strings.TrimPrefix(intentCode, "APPROVE_"), "approve")
	case strings.HasPrefix(intentCode, "REJECT_"):
		return infer(strings.TrimPrefix(intentCode, "REJECT_"), "approve")
	default:
		return "", false
	}
}
