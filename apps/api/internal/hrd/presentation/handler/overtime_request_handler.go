package handler

import (
	"strconv"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/usecase"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// OvertimeRequestHandler handles overtime request HTTP requests
type OvertimeRequestHandler struct {
	overtimeUC   usecase.OvertimeRequestUsecase
	employeeRepo orgRepos.EmployeeRepository
}

// NewOvertimeRequestHandler creates a new OvertimeRequestHandler
func NewOvertimeRequestHandler(overtimeUC usecase.OvertimeRequestUsecase, employeeRepo orgRepos.EmployeeRepository) *OvertimeRequestHandler {
	return &OvertimeRequestHandler{
		overtimeUC:   overtimeUC,
		employeeRepo: employeeRepo,
	}
}

// GetMyRequests handles get my overtime requests (self-service)
func (h *OvertimeRequestHandler) GetMyRequests(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Lookup employee by user_id to get the correct employee_id
	employee, err := h.employeeRepo.FindByUserID(c.Request.Context(), userID.(string))
	if err != nil {
		errors.UnauthorizedResponse(c, "Employee not found for this user")
		return
	}

	var req dto.ListOvertimeRequestsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	// Force filter by current employee
	req.EmployeeID = employee.ID

	requests, pagination, err := h.overtimeUC.List(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      pagination.Total,
			TotalPages: pagination.TotalPages,
			HasNext:    pagination.Page < pagination.TotalPages,
			HasPrev:    pagination.Page > 1,
		},
		Filters: map[string]interface{}{
			"employee_id": employee.ID,
		},
	}

	if req.Status != "" {
		meta.Filters["status"] = req.Status
	}
	if req.RequestType != "" {
		meta.Filters["request_type"] = req.RequestType
	}

	response.SuccessResponse(c, requests, meta)
}

// List handles list overtime requests
func (h *OvertimeRequestHandler) List(c *gin.Context) {
	var req dto.ListOvertimeRequestsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidQueryParamResponse(c)
		return
	}

	requests, pagination, err := h.overtimeUC.List(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       pagination.Page,
			PerPage:    pagination.PerPage,
			Total:      pagination.Total,
			TotalPages: pagination.TotalPages,
			HasNext:    pagination.Page < pagination.TotalPages,
			HasPrev:    pagination.Page > 1,
		},
		Filters: map[string]interface{}{},
	}

	if req.EmployeeID != "" {
		meta.Filters["employee_id"] = req.EmployeeID
	}
	if req.Status != "" {
		meta.Filters["status"] = req.Status
	}
	if req.RequestType != "" {
		meta.Filters["request_type"] = req.RequestType
	}

	response.SuccessResponse(c, requests, meta)
}

// GetByID handles get overtime request by ID
func (h *OvertimeRequestHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	request, err := h.overtimeUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrOvertimeRequestNotFound {
			errors.ErrorResponse(c, "OVERTIME_REQUEST_NOT_FOUND", map[string]interface{}{
				"request_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, request, nil)
}

// GetPending handles get pending overtime requests for manager
func (h *OvertimeRequestHandler) GetPending(c *gin.Context) {
	// Get manager ID from context
	managerID, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	requests, err := h.overtimeUC.GetPendingForManager(c.Request.Context(), managerID.(string))
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, requests, nil)
}

// Create handles create overtime request
func (h *OvertimeRequestHandler) Create(c *gin.Context) {
	var req dto.CreateOvertimeRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	// Use employee_id from request body if provided (for admin/HR creating for others)
	// Otherwise use the current user's employee_id from context
	var targetEmployeeID string
	if req.EmployeeID != "" {
		targetEmployeeID = req.EmployeeID
	} else {
		// Get current user's employee_id
		if id, exists := c.Get("employee_id"); exists {
			targetEmployeeID = id.(string)
		} else {
			// Lookup employee by user_id
			userID, exists := c.Get("user_id")
			if !exists {
				errors.UnauthorizedResponse(c, "User not authenticated")
				return
			}
			employee, err := h.employeeRepo.FindByUserID(c.Request.Context(), userID.(string))
			if err != nil {
				errors.UnauthorizedResponse(c, "Employee not found for this user")
				return
			}
			targetEmployeeID = employee.ID
		}
	}

	request, err := h.overtimeUC.Create(c.Request.Context(), &req, targetEmployeeID)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, request, nil)
}

// Update handles update overtime request
func (h *OvertimeRequestHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateOvertimeRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	request, err := h.overtimeUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrOvertimeRequestNotFound {
			errors.ErrorResponse(c, "OVERTIME_REQUEST_NOT_FOUND", map[string]interface{}{
				"request_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrCannotModifyApprovedRequest {
			errors.ErrorResponse(c, "CANNOT_MODIFY_REQUEST", map[string]interface{}{
				"request_id": id,
				"message":    "Cannot modify an already processed request",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, request, nil)
}

// Approve handles approve overtime request
func (h *OvertimeRequestHandler) Approve(c *gin.Context) {
	id := c.Param("id")

	var approverID string

	// Try to get employee_id from context first
	if empID, exists := c.Get("employee_id"); exists && empID != nil && empID != "" {
		approverID = empID.(string)
	} else {
		// Fallback: lookup employee by user_id
		userID, exists := c.Get("user_id")
		if !exists || userID == nil || userID == "" {
			errors.UnauthorizedResponse(c, "User not authenticated")
			return
		}

		// Lookup employee by user_id
		employee, err := h.employeeRepo.FindByUserID(c.Request.Context(), userID.(string))
		if err != nil {
			errors.UnauthorizedResponse(c, "Employee not found for this user")
			return
		}
		approverID = employee.ID
	}

	var req dto.ApproveOvertimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	request, err := h.overtimeUC.Approve(c.Request.Context(), id, &req, approverID)
	if err != nil {
		if err == usecase.ErrOvertimeRequestNotFound {
			errors.ErrorResponse(c, "OVERTIME_REQUEST_NOT_FOUND", map[string]interface{}{
				"request_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrOvertimeAlreadyProcessed {
			errors.ErrorResponse(c, "REQUEST_ALREADY_PROCESSED", map[string]interface{}{
				"request_id": id,
				"message":    "This request has already been processed",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, request, nil)
}

// Reject handles reject overtime request
func (h *OvertimeRequestHandler) Reject(c *gin.Context) {
	id := c.Param("id")

	var rejecterID string

	// Try to get employee_id from context first
	if empID, exists := c.Get("employee_id"); exists && empID != nil && empID != "" {
		rejecterID = empID.(string)
	} else {
		// Fallback: lookup employee by user_id
		userID, exists := c.Get("user_id")
		if !exists || userID == nil || userID == "" {
			errors.UnauthorizedResponse(c, "User not authenticated")
			return
		}

		// Lookup employee by user_id
		employee, err := h.employeeRepo.FindByUserID(c.Request.Context(), userID.(string))
		if err != nil {
			errors.UnauthorizedResponse(c, "Employee not found for this user")
			return
		}
		rejecterID = employee.ID
	}

	var req dto.RejectOvertimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	request, err := h.overtimeUC.Reject(c.Request.Context(), id, &req, rejecterID)
	if err != nil {
		if err == usecase.ErrOvertimeRequestNotFound {
			errors.ErrorResponse(c, "OVERTIME_REQUEST_NOT_FOUND", map[string]interface{}{
				"request_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrOvertimeAlreadyProcessed {
			errors.ErrorResponse(c, "REQUEST_ALREADY_PROCESSED", map[string]interface{}{
				"request_id": id,
				"message":    "This request has already been processed",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, request, nil)
}

// Cancel handles cancel overtime request
func (h *OvertimeRequestHandler) Cancel(c *gin.Context) {
	id := c.Param("id")

	// Get employee ID from context
	var employeeID string
	if id, exists := c.Get("employee_id"); exists {
		employeeID = id.(string)
	} else {
		// Lookup employee by user_id
		userID, exists := c.Get("user_id")
		if !exists {
			errors.UnauthorizedResponse(c, "User not authenticated")
			return
		}
		employee, err := h.employeeRepo.FindByUserID(c.Request.Context(), userID.(string))
		if err != nil {
			errors.UnauthorizedResponse(c, "Employee not found for this user")
			return
		}
		employeeID = employee.ID
	}

	err := h.overtimeUC.Cancel(c.Request.Context(), id, employeeID)
	if err != nil {
		if err == usecase.ErrOvertimeRequestNotFound {
			errors.ErrorResponse(c, "OVERTIME_REQUEST_NOT_FOUND", map[string]interface{}{
				"request_id": id,
			}, nil)
			return
		}
		if err == usecase.ErrOvertimeAlreadyProcessed {
			errors.ErrorResponse(c, "REQUEST_ALREADY_PROCESSED", map[string]interface{}{
				"request_id": id,
				"message":    "Cannot cancel a processed request",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]interface{}{
		"message": "Overtime request cancelled successfully",
	}, nil)
}

// Delete handles delete overtime request
func (h *OvertimeRequestHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.overtimeUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrOvertimeRequestNotFound {
			errors.ErrorResponse(c, "OVERTIME_REQUEST_NOT_FOUND", map[string]interface{}{
				"request_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]interface{}{
		"message": "Overtime request deleted successfully",
	}, nil)
}

// GetMonthlySummary handles get monthly overtime summary
func (h *OvertimeRequestHandler) GetMonthlySummary(c *gin.Context) {
	// Get employee ID from context or query
	employeeID := c.Query("employee_id")
	if employeeID == "" {
		// Try to get from context first
		if id, exists := c.Get("employee_id"); exists {
			employeeID = id.(string)
		} else {
			// Lookup employee by user_id
			userID, exists := c.Get("user_id")
			if !exists {
				errors.UnauthorizedResponse(c, "User not authenticated")
				return
			}
			employee, err := h.employeeRepo.FindByUserID(c.Request.Context(), userID.(string))
			if err != nil {
				errors.UnauthorizedResponse(c, "Employee not found for this user")
				return
			}
			employeeID = employee.ID
		}
	}

	// Parse year and month from query params, default to current month
	year := apptime.Now().Year()
	month := int(apptime.Now().Month())

	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y >= 2000 && y <= 2100 {
			year = y
		}
	}
	if monthStr := c.Query("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	summary, err := h.overtimeUC.GetEmployeeMonthlySummary(c.Request.Context(), employeeID, year, month)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, summary, nil)
}

// GetPendingNotifications handles get pending overtime notifications for polling
func (h *OvertimeRequestHandler) GetPendingNotifications(c *gin.Context) {
	notifications, err := h.overtimeUC.GetUnnotifiedPendingRequests(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, notifications, nil)
}
