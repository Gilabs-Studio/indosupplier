package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/errors"
	redisInfra "github.com/gilabs/gims/api/internal/core/infrastructure/redis"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/general/domain/dto"
	"github.com/gilabs/gims/api/internal/general/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// OnboardingHandler handles tenant onboarding state endpoints.
type OnboardingHandler struct {
	uc usecase.OnboardingUsecase
}

func NewOnboardingHandler(uc usecase.OnboardingUsecase) *OnboardingHandler {
	return &OnboardingHandler{uc: uc}
}

// GetState returns the current onboarding state for the active tenant.
func (h *OnboardingHandler) GetState(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	cacheKey := fmt.Sprintf("onboarding:state:%v", tenantID)

	// Try cache first (5 min TTL — onboarding state rarely changes)
	if cached, err := getOnboardingCache(c.Request.Context(), cacheKey); cached != nil && err == nil {
		response.SuccessResponse(c, cached, nil)
		return
	}

	state, err := h.uc.GetState(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	setOnboardingCache(c.Request.Context(), cacheKey, state, 5*time.Minute)
	response.SuccessResponse(c, state, nil)
}

func getOnboardingCache(ctx context.Context, key string) (interface{}, error) {
	client := redisInfra.GetClient()
	if client == nil {
		return nil, fmt.Errorf("redis not available")
	}
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func setOnboardingCache(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	client := redisInfra.GetClient()
	if client == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = client.Set(ctx, key, data, ttl).Err()
}

func deleteOnboardingCache(ctx context.Context, key string) {
	client := redisInfra.GetClient()
	if client == nil {
		return
	}
	_ = client.Del(ctx, key).Err()
}

// SetBusinessType sets the business type for the active tenant's onboarding.
func (h *OnboardingHandler) SetBusinessType(c *gin.Context) {
	var req dto.SetBusinessTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	state, err := h.uc.SetBusinessType(c.Request.Context(), req.BusinessType)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	tenantID, _ := c.Get("tenant_id")
	cacheKey := fmt.Sprintf("onboarding:state:%v", tenantID)
	deleteOnboardingCache(c.Request.Context(), cacheKey)
	setOnboardingCache(c.Request.Context(), cacheKey, state, 5*time.Minute)

	response.SuccessResponse(c, state, nil)
}

// MarkComplete marks onboarding as completed for the active tenant.
func (h *OnboardingHandler) MarkComplete(c *gin.Context) {
	state, err := h.uc.MarkComplete(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	tenantID, _ := c.Get("tenant_id")
	cacheKey := fmt.Sprintf("onboarding:state:%v", tenantID)
	deleteOnboardingCache(c.Request.Context(), cacheKey)
	setOnboardingCache(c.Request.Context(), cacheKey, state, 5*time.Minute)

	response.SuccessResponse(c, state, nil)
}
