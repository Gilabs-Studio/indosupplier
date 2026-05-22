package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/usecase"
	"github.com/gin-gonic/gin"
)

// RecruitmentRequestHandler handles HTTP requests for recruitment requests
type RecruitmentRequestHandler struct {
	usecase usecase.RecruitmentRequestUsecase
}

// NewRecruitmentRequestHandler creates a new handler instance
func NewRecruitmentRequestHandler(usecase usecase.RecruitmentRequestUsecase) *RecruitmentRequestHandler {
	return &RecruitmentRequestHandler{usecase: usecase}
}

// GetAll handles GET /recruitment-requests
func (h *RecruitmentRequestHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	var divisionID *string
	if d := c.Query("division_id"); d != "" {
		divisionID = &d
	}

	var positionID *string
	if p := c.Query("position_id"); p != "" {
		positionID = &p
	}

	var priority *string
	if p := c.Query("priority"); p != "" {
		priority = &p
	}

	requests, meta, err := h.usecase.GetAll(c.Request.Context(), page, perPage, search, status, divisionID, positionID, priority)
	if err != nil {
		handleRecruitmentError(c, err)
		return
	}

	response.SuccessResponse(c, requests, &response.Meta{Pagination: meta})
}

// GetByID handles GET /recruitment-requests/:id
func (h *RecruitmentRequestHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	req, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		handleRecruitmentError(c, err)
		return
	}

	response.SuccessResponse(c, req, nil)
}

// Create handles POST /recruitment-requests
func (h *RecruitmentRequestHandler) Create(c *gin.Context) {
	var reqDTO dto.CreateRecruitmentRequestDTO
	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.usecase.Create(c.Request.Context(), &reqDTO, userID.(string))
	if err != nil {
		handleRecruitmentError(c, err)
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// Update handles PUT /recruitment-requests/:id
func (h *RecruitmentRequestHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var reqDTO dto.UpdateRecruitmentRequestDTO
	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.usecase.Update(c.Request.Context(), id, &reqDTO, userID.(string))
	if err != nil {
		handleRecruitmentError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Delete handles DELETE /recruitment-requests/:id
func (h *RecruitmentRequestHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		handleRecruitmentError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{"message": "Recruitment request deleted successfully"}, nil)
}

// UpdateStatus handles POST /recruitment-requests/:id/status
func (h *RecruitmentRequestHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var reqDTO dto.UpdateRecruitmentStatusDTO
	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.usecase.UpdateStatus(c.Request.Context(), id, &reqDTO, userID.(string))
	if err != nil {
		handleRecruitmentError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Submit handles POST /recruitment-requests/:id/submit (DRAFT/REJECTED -> PENDING)
func (h *RecruitmentRequestHandler) Submit(c *gin.Context) {
	h.updateStatusAction(c, "PENDING")
}

// Approve handles POST /recruitment-requests/:id/approve (PENDING -> APPROVED)
func (h *RecruitmentRequestHandler) Approve(c *gin.Context) {
	h.updateStatusAction(c, "APPROVED")
}

// Reject handles POST /recruitment-requests/:id/reject (PENDING -> REJECTED)
func (h *RecruitmentRequestHandler) Reject(c *gin.Context) {
	id := c.Param("id")

	var reqDTO dto.RejectRecruitmentRequestDTO
	_ = c.ShouldBindJSON(&reqDTO) // Optional body

	userID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	statusDTO := &dto.UpdateRecruitmentStatusDTO{Status: "REJECTED", Notes: reqDTO.Notes}
	result, err := h.usecase.UpdateStatus(c.Request.Context(), id, statusDTO, userID.(string))
	if err != nil {
		handleRecruitmentError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Open handles POST /recruitment-requests/:id/open (APPROVED -> OPEN)
func (h *RecruitmentRequestHandler) Open(c *gin.Context) {
	h.updateStatusAction(c, "OPEN")
}

// Close handles POST /recruitment-requests/:id/close (OPEN -> CLOSED)
func (h *RecruitmentRequestHandler) Close(c *gin.Context) {
	h.updateStatusAction(c, "CLOSED")
}

// Cancel handles POST /recruitment-requests/:id/cancel (DRAFT/PENDING -> CANCELLED)
func (h *RecruitmentRequestHandler) Cancel(c *gin.Context) {
	h.updateStatusAction(c, "CANCELLED")
}

// updateStatusAction is a helper for simple status transition endpoints
func (h *RecruitmentRequestHandler) updateStatusAction(c *gin.Context, status string) {
	id := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	statusDTO := &dto.UpdateRecruitmentStatusDTO{Status: status}
	result, err := h.usecase.UpdateStatus(c.Request.Context(), id, statusDTO, userID.(string))
	if err != nil {
		handleRecruitmentError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// UpdateFilledCount handles PUT /recruitment-requests/:id/filled-count
func (h *RecruitmentRequestHandler) UpdateFilledCount(c *gin.Context) {
	id := c.Param("id")

	var reqDTO dto.UpdateFilledCountDTO
	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	result, err := h.usecase.UpdateFilledCount(c.Request.Context(), id, &reqDTO)
	if err != nil {
		handleRecruitmentError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetFormData handles GET /recruitment-requests/form-data
func (h *RecruitmentRequestHandler) GetFormData(c *gin.Context) {
	formData, err := h.usecase.GetFormData(c.Request.Context())
	if err != nil {
		handleRecruitmentError(c, err)
		return
	}

	response.SuccessResponse(c, formData, nil)
}

// handleRecruitmentError maps usecase errors to appropriate HTTP responses
func handleRecruitmentError(c *gin.Context, err error) {
	msg := err.Error()

	switch {
	case msg == "recruitment request not found":
		errors.ErrorResponse(c, "RECRUITMENT_REQUEST_NOT_FOUND", map[string]interface{}{"message": msg}, nil)
	case msg == "division not found":
		errors.ErrorResponse(c, "DIVISION_NOT_FOUND", map[string]interface{}{"message": msg}, nil)
	case msg == "position not found":
		errors.ErrorResponse(c, "POSITION_NOT_FOUND", map[string]interface{}{"message": msg}, nil)
	case msg == "employee not found for current user":
		errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{"message": msg}, nil)
	case strings.HasPrefix(msg, "RECRUITMENT_NOT_EDITABLE"):
		errors.ErrorResponse(c, "RECRUITMENT_NOT_EDITABLE", map[string]interface{}{"message": msg}, nil)
	case strings.HasPrefix(msg, "INVALID_STATUS_TRANSITION"):
		errors.ErrorResponse(c, "INVALID_STATUS_TRANSITION", map[string]interface{}{"message": msg}, nil)
	case strings.HasPrefix(msg, "INVALID_SALARY_RANGE"):
		errors.ErrorResponse(c, "INVALID_SALARY_RANGE", map[string]interface{}{"message": msg}, nil)
	case strings.HasPrefix(msg, "INVALID_DATE_FORMAT"):
		errors.ErrorResponse(c, "INVALID_DATE_FORMAT", map[string]interface{}{"message": msg}, nil)
	case strings.HasPrefix(msg, "RECRUITMENT_NOT_OPEN"):
		errors.ErrorResponse(c, "RECRUITMENT_NOT_OPEN", map[string]interface{}{"message": msg}, nil)
	case strings.HasPrefix(msg, "FILLED_EXCEEDS_REQUIRED"):
		errors.ErrorResponse(c, "FILLED_EXCEEDS_REQUIRED", map[string]interface{}{"message": msg}, nil)
	default:
		errors.ErrorResponse(c, "INTERNAL_ERROR", map[string]interface{}{"message": "An unexpected error occurred"}, nil)
	}
}
