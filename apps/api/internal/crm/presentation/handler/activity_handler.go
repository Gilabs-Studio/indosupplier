package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActivityHandler handles HTTP requests for CRM activities
type ActivityHandler struct {
	uc usecase.ActivityUsecase
	db *gorm.DB
}

// NewActivityHandler creates a new activity handler
func NewActivityHandler(uc usecase.ActivityUsecase, db *gorm.DB) *ActivityHandler {
	return &ActivityHandler{uc: uc, db: db}
}

// Create handles POST request to create an activity
func (h *ActivityHandler) Create(c *gin.Context) {
	var req dto.CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	createdBy := ""
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			createdBy = id
		}
	}

	employeeID, resolveErr := h.resolveActivityEmployeeID(c, req, createdBy)
	if resolveErr != nil {
		errors.ErrorResponse(c, "INVALID_FORMAT", map[string]interface{}{
			"field":   "employee_id",
			"message": resolveErr.Error(),
		}, nil)
		return
	}

	result, err := h.uc.Create(c.Request.Context(), req, employeeID)
	if err != nil {
		handleActivityError(c, err)
		return
	}

	meta := &response.Meta{}
	if createdBy != "" {
		meta.CreatedBy = createdBy
	}

	response.SuccessResponseCreated(c, result, meta)
}

func (h *ActivityHandler) resolveActivityEmployeeID(c *gin.Context, req dto.CreateActivityRequest, authUserID string) (string, error) {
	candidate := ""

	if req.EmployeeID != nil && strings.TrimSpace(*req.EmployeeID) != "" {
		candidate = strings.TrimSpace(*req.EmployeeID)
	} else if req.UserID != nil && strings.TrimSpace(*req.UserID) != "" {
		candidate = strings.TrimSpace(*req.UserID)
	} else {
		candidate = strings.TrimSpace(authUserID)
	}

	if candidate == "" {
		return "", fmt.Errorf("employee_id or user_id is required")
	}

	var employeeID string

	// If candidate looks like a UUID, try ID lookup first.
	if _, err := uuid.Parse(candidate); err == nil {
		if err := h.db.WithContext(c.Request.Context()).
			Table("employees").
			Select("id").
			Where("id = ? AND deleted_at IS NULL", candidate).
			Scan(&employeeID).Error; err != nil {
			return "", err
		}
		if employeeID != "" {
			return employeeID, nil
		}
	}

	// Try lookup by linked user_id (candidate may be a user id)
	if err := h.db.WithContext(c.Request.Context()).
		Table("employees").
		Select("id").
		Where("user_id = ? AND deleted_at IS NULL", candidate).
		Scan(&employeeID).Error; err != nil {
		return "", err
	}
	if employeeID != "" {
		return employeeID, nil
	}

// Fallback: try lookup by employee_code (frontend may send codes like EMP-123)
	if err := h.db.WithContext(c.Request.Context()).
		Table("employees").
		Select("id").
		Where("employee_code = ? AND deleted_at IS NULL", candidate).
		Scan(&employeeID).Error; err != nil {
		return "", err
	}
	if employeeID != "" {
		return employeeID, nil
	}

	// No employee found. If we have an authenticated user id, fall back to that
	// so clients (users without linked employee rows) can still create activities.
	// This writes the user UUID into the employee_id column; it's acceptable
	// as a pragmatic fallback because some tenants don't create employee records
	// for every user. Prefer linking employees in the long run.
	if strings.TrimSpace(authUserID) != "" {
		fmt.Printf("[WARN] activity: no employee profile found for candidate %s, falling back to auth user %s\n", candidate, authUserID)
		return strings.TrimSpace(authUserID), nil
	}

	return "", fmt.Errorf("no employee profile found for the given employee_id/user_id")
}

// GetByID handles GET request to get an activity by ID
func (h *ActivityHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_ID", map[string]interface{}{
			"message": "ID is required",
		}, nil)
		return
	}

	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.ErrorResponse(c, "CRM_ACTIVITY_NOT_FOUND", map[string]interface{}{
			"activity_id": id,
		}, nil)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// List handles GET request to list activities with filtering and pagination
func (h *ActivityHandler) List(c *gin.Context) {
	params := repositories.ActivityListParams{
		Search:         c.Query("search"),
		SortBy:         c.DefaultQuery("sort_by", "timestamp"),
		SortDir:        c.DefaultQuery("sort_dir", "desc"),
		Type:           c.Query("type"),
		ActivityTypeID: c.Query("activity_type_id"),
		CustomerID:     c.Query("customer_id"),
		ContactID:      c.Query("contact_id"),
		DealID:         c.Query("deal_id"),
		LeadID:         c.Query("lead_id"),
		EmployeeID:     c.Query("employee_id"),
		DateFrom:       c.Query("date_from"),
		DateTo:         c.Query("date_to"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}
	params.Limit = perPage
	params.Offset = (page - 1) * perPage

	results, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      int(total),
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		Filters: map[string]interface{}{},
	}

	if params.Type != "" {
		meta.Filters["type"] = params.Type
	}
	if params.EmployeeID != "" {
		meta.Filters["employee_id"] = params.EmployeeID
	}
	if params.CustomerID != "" {
		meta.Filters["customer_id"] = params.CustomerID
	}

	response.SuccessResponse(c, results, meta)
}

// Timeline handles GET request to get activity timeline for an entity
func (h *ActivityHandler) Timeline(c *gin.Context) {
	params := repositories.ActivityListParams{
		SortBy:     "timestamp",
		SortDir:    "desc",
		CustomerID: c.Query("customer_id"),
		ContactID:  c.Query("contact_id"),
		DealID:     c.Query("deal_id"),
		LeadID:     c.Query("lead_id"),
		EmployeeID: c.Query("employee_id"),
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if perPage > 100 {
		perPage = 100
	}
	params.Limit = perPage
	params.Offset = (page - 1) * perPage

	results, total, err := h.uc.Timeline(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      int(total),
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}

	response.SuccessResponse(c, results, meta)
}

// MyActivities handles GET request to list current user's activities
func (h *ActivityHandler) MyActivities(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", map[string]interface{}{
			"message": "User not authenticated",
		}, nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		errors.ErrorResponse(c, "UNAUTHORIZED", map[string]interface{}{
			"message": "Invalid user ID",
		}, nil)
		return
	}

	// Resolve employee_id from user_id
	var employeeID string
	if err := h.db.WithContext(c.Request.Context()).
		Table("employees").
		Select("id").
		Where("user_id = ? AND deleted_at IS NULL", userIDStr).
		Row().Scan(&employeeID); err != nil {
		errors.ErrorResponse(c, "CRM_ACTIVITY_NOT_FOUND", map[string]interface{}{
			"message": "No employee profile linked to current user",
		}, nil)
		return
	}

	params := repositories.ActivityListParams{
		SortBy:         "timestamp",
		SortDir:        "desc",
		EmployeeID:     employeeID,
		ActivityTypeID: c.Query("activity_type_id"),
		DateFrom:       c.Query("date_from"),
		DateTo:         c.Query("date_to"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}
	params.Limit = perPage
	params.Offset = (page - 1) * perPage

	results, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      int(total),
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}

	response.SuccessResponse(c, results, meta)
}

// handleActivityError maps business errors to appropriate HTTP responses
func handleActivityError(c *gin.Context, err error) {
	switch err.Error() {
	case "activity not found":
		errors.ErrorResponse(c, "CRM_ACTIVITY_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "activity type not found":
		errors.ErrorResponse(c, "CRM_ACTIVITY_TYPE_NOT_FOUND", map[string]interface{}{
			"message": err.Error(),
		}, nil)
	case "invalid timestamp format, use ISO 8601":
		errors.ErrorResponse(c, "INVALID_FORMAT", map[string]interface{}{
			"message": err.Error(),
			"field":   "timestamp",
		}, nil)
	default:
		errors.InternalServerErrorResponse(c, err.Error())
	}
}
