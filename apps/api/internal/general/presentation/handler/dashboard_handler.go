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
)

// DashboardHandler handles HTTP requests for the dashboard overview endpoint.
type DashboardHandler struct {
	uc usecase.DashboardUsecase
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(uc usecase.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

// GetOverview handles GET /general/dashboard/overview
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	var req dto.DashboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		errors.InvalidQueryParamResponse(c)
		return
	}

	// Populate RBAC context for scope-based filtering
	if uid, ok := c.Get("user_id"); ok {
		req.UserID, _ = uid.(string)
	}

	if req.Scope != "" {
		if !req.Scope.IsValid() {
			errors.InvalidQueryParamResponse(c)
			return
		}

		// Scoped requests: cache per scope
		cacheKey := fmt.Sprintf("dashboard:scoped:%s:%s:%s:%s", req.UserID, req.Scope, req.StartDate, req.EndDate)
		if cached, err := getCachedResponse(c.Request.Context(), cacheKey); cached != nil && err == nil {
			response.SuccessResponse(c, cached, nil)
			return
		}

		result, err := h.uc.GetOverviewByScope(c.Request.Context(), req)
		if err != nil {
			errors.InternalServerErrorResponse(c, err.Error())
			return
		}

		setCachedResponse(c.Request.Context(), cacheKey, result, 60*time.Second)
		response.SuccessResponse(c, result, nil)
		return
	}

	// Full overview: cache for 60 seconds
	cacheKey := fmt.Sprintf("dashboard:overview:%s:%s:%s", req.UserID, req.StartDate, req.EndDate)
	if cached, err := getCachedResponse(c.Request.Context(), cacheKey); cached != nil && err == nil {
		response.SuccessResponse(c, cached, nil)
		return
	}

	result, err := h.uc.GetOverview(c.Request.Context(), req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	setCachedResponse(c.Request.Context(), cacheKey, result, 60*time.Second)
	response.SuccessResponse(c, result, nil)
}

// getCachedResponse retrieves a cached JSON response from Redis.
func getCachedResponse(ctx context.Context, key string) (interface{}, error) {
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

// setCachedResponse stores a JSON response in Redis with TTL.
func setCachedResponse(ctx context.Context, key string, value interface{}, ttl time.Duration) {
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

// GetLayout handles GET /general/dashboard/layout — fetches the current user's saved layout.
func (h *DashboardHandler) GetLayout(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "missing user context")
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		errors.UnauthorizedResponse(c, "invalid user context")
		return
	}

	dashboardType := c.DefaultQuery("type", "general")

	result, err := h.uc.GetLayout(c.Request.Context(), userID, dashboardType)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	if result == nil {
		// No saved layout yet — frontend should use the default
		response.SuccessResponse(c, nil, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// SaveLayout handles PUT /general/dashboard/layout — persists the current user's layout.
func (h *DashboardHandler) SaveLayout(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "missing user context")
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		errors.UnauthorizedResponse(c, "invalid user context")
		return
	}

	var req dto.DashboardLayoutSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.InvalidRequestBodyResponse(c)
		return
	}

	if err := h.uc.SaveLayout(c.Request.Context(), userID, req); err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, nil, nil)
}
