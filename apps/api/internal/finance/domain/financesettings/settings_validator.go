package financesettings

import (
	"context"
	"fmt"
	"strings"
)

// RequiredCOASetting defines a COA setting that must be configured for the system to work.
type RequiredCOASetting struct {
	Key         string
	Description string
	Category    string // inventory, fx, depreciation, etc
}

// ValidatorConfig holds the configuration for settings validation.
type ValidatorConfig struct {
	// RequiredSettings list of settings that MUST be configured
	RequiredSettings []RequiredCOASetting

	// FailFast: if true, fail on FIRST missing setting; if false, collect all errors
	FailFast bool
}

// SettingsValidator checks that all required finance settings are properly configured.
type SettingsValidator interface {
	// ValidateRequiredSettings checks that all required COA settings exist and are non-empty.
	// Returns error only if validation fails; nil if all OK.
	ValidateRequiredSettings(ctx context.Context, config ValidatorConfig) error

	// ValidateValueExists checks if a single setting key is configured and non-empty.
	ValidateValueExists(ctx context.Context, settingKey string) error
}

type settingsValidator struct {
	service SettingsService
}

// NewSettingsValidator creates a new validator.
func NewSettingsValidator(service SettingsService) SettingsValidator {
	return &settingsValidator{service: service}
}

// DefaultRequiredSettings returns the minimal set of COA settings required for valuation operations.
// These MUST be seeded before system can perform valuation runs.
func DefaultRequiredSettings() []RequiredCOASetting {
	return []RequiredCOASetting{
		// Inventory
		{Key: "coa.inventory_asset", Description: "Inventory Asset", Category: "inventory"},
		{Key: "coa.inventory_loss", Description: "Inventory Loss", Category: "inventory"},
		{Key: "coa.inventory_revaluation_reserve", Description: "Inventory Revaluation Reserve (Equity)", Category: "inventory"},
		// FX
		{Key: "coa.fx_gain", Description: "FX Gain", Category: "fx"},
		{Key: "coa.fx_loss", Description: "FX Loss", Category: "fx"},
		{Key: "coa.fx_remeasurement", Description: "FX Remeasurement", Category: "fx"},
		// Depreciation
		{Key: "coa.depreciation_expense", Description: "Depreciation Expense", Category: "depreciation"},
		{Key: "coa.depreciation_accumulated", Description: "Accumulated Depreciation", Category: "depreciation"},
		{Key: "coa.depreciation_gain", Description: "Depreciation Reversal Gain", Category: "depreciation"},
	}
}

// ValidateRequiredSettings checks that all required COA settings exist and are non-empty.
func (v *settingsValidator) ValidateRequiredSettings(ctx context.Context, config ValidatorConfig) error {
	var errors []string

	for _, required := range config.RequiredSettings {
		if err := v.ValidateValueExists(ctx, required.Key); err != nil {
			errorMsg := fmt.Sprintf(
				"[%s] %s: %v",
				required.Category,
				required.Key,
				err,
			)
			errors = append(errors, errorMsg)

			if config.FailFast {
				return fmt.Errorf("finance settings validation failed:\n%s\n\nAction: Run finance settings seeder or configure COA mappings via admin panel", errorMsg)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf(
			"finance settings validation failed (%d errors):\n\n"+
				strings.Join(errors, "\n\n")+
				"\n\nAction: Configure all required COA mappings in admin panel or run seeder",
			len(errors),
		)
	}

	return nil
}

// ValidateValueExists validates that a single setting is configured and non-empty.
func (v *settingsValidator) ValidateValueExists(ctx context.Context, settingKey string) error {
	if settingKey == "" {
		return fmt.Errorf("setting key cannot be empty")
	}

	value, err := v.service.GetValue(ctx, settingKey)
	if err != nil {
		return fmt.Errorf("setting not found")
	}

	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("setting is configured but empty (value must be COA code)")
	}

	return nil
}
