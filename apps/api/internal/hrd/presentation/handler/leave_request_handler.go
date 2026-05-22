package handler

import (
	stderrors "errors"
	"io"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LeaveRequestHandler handles HTTP requests for leave requests
type LeaveRequestHandler struct {
	leaveRequestUsecase usecase.LeaveRequestUsecase
}

const leaveRequestIDRequiredMessage = "Leave request ID is required"

// NewLeaveRequestHandler creates a new LeaveRequestHandler
func NewLeaveRequestHandler(leaveRequestUsecase usecase.LeaveRequestUsecase) *LeaveRequestHandler {
	return &LeaveRequestHandler{
		leaveRequestUsecase: leaveRequestUsecase,
	}
}

// Create handles POST /api/v1/hrd/leave-requests
func (h *LeaveRequestHandler) Create(c *gin.Context) {
	var req dto.CreateLeaveRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.Create(c.Request.Context(), &req, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// CreateSelf handles POST /api/v1/hrd/leave-requests/self
func (h *LeaveRequestHandler) CreateSelf(c *gin.Context) {
	var req dto.CreateMyLeaveRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.CreateSelf(c.Request.Context(), &req, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// List handles GET /api/v1/hrd/leave-requests
func (h *LeaveRequestHandler) List(c *gin.Context) {
	var filters dto.LeaveRequestListFilterDTO
	if err := c.ShouldBindQuery(&filters); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	results, total, err := h.leaveRequestUsecase.List(c.Request.Context(), &filters, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	perPage := filters.PerPage
	if perPage < 1 {
		perPage = 20
	}

	meta := response.NewPaginationMeta(page, perPage, int(total))

	response.SuccessResponse(c, results, &response.Meta{Pagination: meta})
}

// ListSelf handles GET /api/v1/hrd/leave-requests/self
func (h *LeaveRequestHandler) ListSelf(c *gin.Context) {
	var filters dto.LeaveRequestListFilterDTO
	if err := c.ShouldBindQuery(&filters); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	results, total, err := h.leaveRequestUsecase.ListSelf(c.Request.Context(), &filters, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	perPage := filters.PerPage
	if perPage < 1 {
		perPage = 20
	}

	meta := response.NewPaginationMeta(page, perPage, int(total))
	response.SuccessResponse(c, results, &response.Meta{Pagination: meta})
}

// GetByID handles GET /api/v1/hrd/leave-requests/:id
func (h *LeaveRequestHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": leaveRequestIDRequiredMessage}, nil)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.GetByID(c.Request.Context(), id, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetSelfByID handles GET /api/v1/hrd/leave-requests/self/:id
func (h *LeaveRequestHandler) GetSelfByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": leaveRequestIDRequiredMessage}, nil)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.GetSelfByID(c.Request.Context(), id, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Update handles PUT /api/v1/hrd/leave-requests/:id
func (h *LeaveRequestHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": leaveRequestIDRequiredMessage}, nil)
		return
	}

	var req dto.UpdateLeaveRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.Update(c.Request.Context(), id, &req, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// UpdateSelf handles PUT /api/v1/hrd/leave-requests/self/:id
func (h *LeaveRequestHandler) UpdateSelf(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": leaveRequestIDRequiredMessage}, nil)
		return
	}

	var req dto.UpdateLeaveRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.UpdateSelf(c.Request.Context(), id, &req, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Delete handles DELETE /api/v1/hrd/leave-requests/:id
func (h *LeaveRequestHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": leaveRequestIDRequiredMessage}, nil)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	err := h.leaveRequestUsecase.Delete(c.Request.Context(), id, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{"message": "Leave request deleted successfully"}, nil)
}

// GetBalance handles GET /api/v1/hrd/leave-requests/balance/:employee_id
func (h *LeaveRequestHandler) GetBalance(c *gin.Context) {
	employeeID := c.Param("employee_id")
	if employeeID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": "Employee ID is required"}, nil)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.CalculateBalance(c.Request.Context(), employeeID, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetMyBalance handles GET /api/v1/hrd/leave-requests/my-balance
func (h *LeaveRequestHandler) GetMyBalance(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.GetSelfBalance(c.Request.Context(), currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Approve handles POST /api/v1/hrd/leave-requests/:id/approve
func (h *LeaveRequestHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": leaveRequestIDRequiredMessage}, nil)
		return
	}

	var req dto.ApproveLeaveRequestDTO
	_ = c.ShouldBindJSON(&req) // Allow empty body

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.Approve(c.Request.Context(), id, &req, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetFormData handles GET /api/v1/hrd/leave-requests/form-data
func (h *LeaveRequestHandler) GetFormData(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.GetFormData(c.Request.Context(), currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetMyFormData handles GET /api/v1/hrd/leave-requests/my-form-data
func (h *LeaveRequestHandler) GetMyFormData(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.GetSelfFormData(c.Request.Context(), currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Reject handles POST /api/v1/hrd/leave-requests/:id/reject
func (h *LeaveRequestHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": leaveRequestIDRequiredMessage}, nil)
		return
	}

	var req dto.RejectLeaveRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleValidationError(c, err)
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.Reject(c.Request.Context(), id, &req, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Cancel handles POST /api/v1/hrd/leave-requests/:id/cancel
func (h *LeaveRequestHandler) Cancel(c *gin.Context) {
	id, ok := validateLeaveRequestIDParam(c)
	if !ok {
		return
	}

	var req dto.CancelLeaveRequestDTO
	if !bindOptionalCancelLeaveRequest(c, &req) {
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.Cancel(c.Request.Context(), id, &req, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// CancelSelf handles POST /api/v1/hrd/leave-requests/self/:id/cancel
func (h *LeaveRequestHandler) CancelSelf(c *gin.Context) {
	id, ok := validateLeaveRequestIDParam(c)
	if !ok {
		return
	}

	var req dto.CancelLeaveRequestDTO
	if !bindOptionalCancelLeaveRequest(c, &req) {
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.CancelSelf(c.Request.Context(), id, &req, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

func validateLeaveRequestIDParam(c *gin.Context) (string, bool) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": leaveRequestIDRequiredMessage}, nil)
		return "", false
	}

	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": "Invalid leave request ID"}, nil)
		return "", false
	}

	return id, true
}

func bindOptionalCancelLeaveRequest(c *gin.Context, req *dto.CancelLeaveRequestDTO) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		if stderrors.Is(err, io.EOF) {
			return true
		}

		errors.HandleValidationError(c, err)
		return false
	}

	return true
}

// Reapprove handles POST /api/v1/hrd/leave-requests/:id/reapprove
func (h *LeaveRequestHandler) Reapprove(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": "Leave request ID is required"}, nil)
		return
	}

	var req dto.ApproveLeaveRequestDTO
	_ = c.ShouldBindJSON(&req) // Allow empty body

	currentUserID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.leaveRequestUsecase.Reapprove(c.Request.Context(), id, &req, currentUserID.(string))
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// AuditTrail handles GET /api/v1/hrd/leave-requests/:id/audit-trail
func (h *LeaveRequestHandler) AuditTrail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": "Leave request ID is required"}, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	items, total, err := h.leaveRequestUsecase.ListAuditTrail(c.Request.Context(), id, page, perPage)
	if err != nil {
		handleUsecaseError(c, err)
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	meta.Pagination.TotalPages = totalPages
	meta.Pagination.HasNext = page < totalPages
	meta.Pagination.HasPrev = page > 1

	response.SuccessResponse(c, items, meta)
}

// handleUsecaseError maps usecase errors to appropriate HTTP responses
func handleUsecaseError(c *gin.Context, err error) {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "FORBIDDEN"):
		errors.ErrorResponse(c, "FORBIDDEN", nil, nil)
	case strings.Contains(errMsg, "INSUFFICIENT_LEAVE_BALANCE"):
		errors.ErrorResponse(c, "INSUFFICIENT_LEAVE_BALANCE", map[string]interface{}{"message": errMsg}, nil)
	case strings.Contains(errMsg, "OVERLAPPING_LEAVE_REQUEST"):
		errors.ErrorResponse(c, "OVERLAPPING_LEAVE_REQUEST", map[string]interface{}{"message": errMsg}, nil)
	case strings.Contains(errMsg, "INVALID_DATE_FORMAT"):
		errors.ErrorResponse(c, "INVALID_DATE_FORMAT", map[string]interface{}{"message": errMsg}, nil)
	case strings.Contains(errMsg, "INVALID_STATUS"):
		errors.ErrorResponse(c, "INVALID_STATUS", map[string]interface{}{"message": errMsg}, nil)
	case strings.Contains(errMsg, "INVALID_DATE"):
		errors.ErrorResponse(c, "INVALID_DATE", map[string]interface{}{"message": errMsg}, nil)
	case strings.Contains(errMsg, "VALIDATION_ERROR"):
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": errMsg}, nil)
	case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "NOT_FOUND"):
		errors.ErrorResponse(c, "LEAVE_REQUEST_NOT_FOUND", nil, nil)
	default:
		errors.ErrorResponse(c, "INTERNAL_ERROR", map[string]interface{}{"message": "An unexpected error occurred"}, nil)
	}
}
