package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
	"github.com/gin-gonic/gin"
)

type FinanceSettingsHandler struct {
	settingsService financesettings.SettingsService
}

func NewFinanceSettingsHandler(settingsService financesettings.SettingsService) *FinanceSettingsHandler {
	return &FinanceSettingsHandler{
		settingsService: settingsService,
	}
}

func (h *FinanceSettingsHandler) GetAll(c *gin.Context) {
	settings, err := h.settingsService.GetAll(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve finance settings", err.Error(), nil, nil)
		return
	}

	var res []dto.FinanceSettingResponse
	for _, s := range settings {
		res = append(res, dto.FinanceSettingResponse{
			ID:          s.ID,
			SettingKey:  s.SettingKey,
			Value:       s.Value,
			Description: s.Description,
			Category:    s.Category,
		})
	}

	response.SuccessResponse(c, res, nil)
}

func (h *FinanceSettingsHandler) BatchUpsert(c *gin.Context) {
	var req dto.BatchUpsertFinanceSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error(), nil, nil)
		return
	}

	ctx := c.Request.Context()
	for _, config := range req.Settings {
		if err := h.settingsService.Upsert(ctx, config.SettingKey, config.Value, config.Description, config.Category); err != nil {
			log.Printf("Failed to upsert setting %s: %v", config.SettingKey, err)
			response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update settings", err.Error(), nil, nil)
			return
		}
	}

	// Reload settings after upsert
	settings, err := h.settingsService.GetAll(ctx)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve finance settings", err.Error(), nil, nil)
		return
	}

	var res []dto.FinanceSettingResponse
	for _, s := range settings {
		res = append(res, dto.FinanceSettingResponse{
			ID:          s.ID,
			SettingKey:  s.SettingKey,
			Value:       s.Value,
			Description: s.Description,
			Category:    s.Category,
		})
	}

	response.SuccessResponse(c, res, nil)
}

func defaultAgingBucketConfig() []dto.AgingBucketDefinition {
	zero := 0
	one := 1
	thirty := 30
	thirtyOne := 31
	sixty := 60
	sixtyOne := 61
	ninety := 90
	ninetyOne := 91

	return []dto.AgingBucketDefinition{
		{Key: "current", Label: "Current", MaxDays: &zero},
		{Key: "days_1_30", Label: "1-30 Days", MinDays: &one, MaxDays: &thirty},
		{Key: "days_31_60", Label: "31-60 Days", MinDays: &thirtyOne, MaxDays: &sixty},
		{Key: "days_61_90", Label: "61-90 Days", MinDays: &sixtyOne, MaxDays: &ninety},
		{Key: "over_90", Label: ">90 Days", MinDays: &ninetyOne},
	}
}

func normalizeAgingBucketKey(raw string) string {
	key := strings.TrimSpace(strings.ToLower(raw))
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, " ", "_")
	return key
}

func agingBucketConfigSettingKeys(reportType string) []string {
	primary := models.AgingBucketConfigSettingKey(reportType)
	if primary == models.SettingAgingBucketConfig {
		return []string{primary}
	}
	return []string{primary, models.SettingAgingBucketConfig}
}

func (h *FinanceSettingsHandler) GetAgingBucketConfig(c *gin.Context) {
	reportType := c.Query("report_type")
	raw := ""
	for _, key := range agingBucketConfigSettingKeys(reportType) {
		value, err := h.settingsService.GetValue(c.Request.Context(), key)
		if err != nil || strings.TrimSpace(value) == "" {
			continue
		}
		raw = value
		break
	}
	if strings.TrimSpace(raw) == "" {
		response.SuccessResponse(c, gin.H{"buckets": defaultAgingBucketConfig()}, nil)
		return
	}

	buckets := make([]dto.AgingBucketDefinition, 0)
	if err := json.Unmarshal([]byte(raw), &buckets); err != nil || len(buckets) == 0 {
		response.SuccessResponse(c, gin.H{"buckets": defaultAgingBucketConfig()}, nil)
		return
	}

	response.SuccessResponse(c, gin.H{"buckets": buckets}, nil)
}

type resolvedAgingBucketRange struct {
	key string
	min int
	max int
}

func sanitizeAgingBucketConfig(input []dto.AgingBucketDefinition) ([]dto.AgingBucketDefinition, error) {
	if len(input) == 0 {
		return nil, errors.New("bucket list cannot be empty")
	}
	if len(input) > 20 {
		return nil, errors.New("bucket count exceeds maximum allowed")
	}

	sanitized := make([]dto.AgingBucketDefinition, 0, len(input))
	seenKeys := make(map[string]struct{}, len(input))
	resolvedRanges := make([]resolvedAgingBucketRange, 0, len(input))
	hasCurrentCoverage := false
	hasPositiveCoverage := false

	for idx, bucket := range input {
		key := normalizeAgingBucketKey(bucket.Key)
		if key == "" {
			key = normalizeAgingBucketKey(bucket.Label)
		}
		if key == "" {
			key = fmt.Sprintf("bucket_%d", idx+1)
		}

		if _, exists := seenKeys[key]; exists {
			return nil, fmt.Errorf("duplicate bucket key '%s'", key)
		}
		seenKeys[key] = struct{}{}

		label := strings.TrimSpace(bucket.Label)
		if label == "" {
			label = key
		}

		if bucket.MinDays != nil && *bucket.MinDays < 0 {
			return nil, fmt.Errorf("bucket '%s' has negative min_days", key)
		}
		if bucket.MaxDays != nil && *bucket.MaxDays < 0 {
			return nil, fmt.Errorf("bucket '%s' has negative max_days", key)
		}
		if bucket.MinDays != nil && bucket.MaxDays != nil && *bucket.MinDays > *bucket.MaxDays {
			return nil, fmt.Errorf("bucket '%s' has min_days greater than max_days", key)
		}

		resolvedMin := -1 << 30
		resolvedMax := 1 << 30
		if bucket.MinDays != nil {
			resolvedMin = *bucket.MinDays
		}
		if bucket.MaxDays != nil {
			resolvedMax = *bucket.MaxDays
		}
		if bucket.MinDays == nil && bucket.MaxDays == nil {
			return nil, fmt.Errorf("bucket '%s' must define min_days or max_days", key)
		}

		if resolvedMin <= 0 && resolvedMax >= 0 {
			hasCurrentCoverage = true
		}
		if resolvedMax > 0 {
			hasPositiveCoverage = true
		}

		resolvedRanges = append(resolvedRanges, resolvedAgingBucketRange{key: key, min: resolvedMin, max: resolvedMax})

		sanitized = append(sanitized, dto.AgingBucketDefinition{
			Key:     key,
			Label:   label,
			MinDays: bucket.MinDays,
			MaxDays: bucket.MaxDays,
		})
	}

	for i := 0; i < len(resolvedRanges); i++ {
		for j := i + 1; j < len(resolvedRanges); j++ {
			left := resolvedRanges[i]
			right := resolvedRanges[j]
			if left.min <= right.max && right.min <= left.max {
				return nil, fmt.Errorf("bucket ranges overlap between '%s' and '%s'", left.key, right.key)
			}
		}
	}

	if !hasCurrentCoverage {
		return nil, errors.New("bucket configuration must include current range (day 0)")
	}
	if !hasPositiveCoverage {
		return nil, errors.New("bucket configuration must include overdue range (>0 day)")
	}

	return sanitized, nil
}

func (h *FinanceSettingsHandler) UpsertAgingBucketConfig(c *gin.Context) {
	var req dto.UpdateAgingBucketConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error(), nil, nil)
		return
	}

	buckets, err := sanitizeAgingBucketConfig(req.Buckets)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	raw, err := json.Marshal(buckets)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to encode aging bucket settings", err.Error(), nil, nil)
		return
	}

	if err := h.settingsService.Upsert(
		c.Request.Context(),
		models.AgingBucketConfigSettingKey(req.ReportType),
		string(raw),
		fmt.Sprintf("Aging bucket configuration for %s aging report", strings.ToUpper(req.ReportType)),
		"report",
	); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Failed to update aging bucket settings", err.Error(), nil, nil)
		return
	}

	response.SuccessResponse(c, gin.H{"buckets": buckets}, nil)
}
