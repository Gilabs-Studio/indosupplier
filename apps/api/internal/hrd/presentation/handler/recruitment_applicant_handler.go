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

// RecruitmentApplicantHandler handles HTTP requests for recruitment applicants
type RecruitmentApplicantHandler struct {
	usecase usecase.RecruitmentApplicantUsecase
}

// NewRecruitmentApplicantHandler creates a new handler instance
func NewRecruitmentApplicantHandler(usecase usecase.RecruitmentApplicantUsecase) *RecruitmentApplicantHandler {
	return &RecruitmentApplicantHandler{usecase: usecase}
}

// GetAll handles GET /applicants
func (h *RecruitmentApplicantHandler) GetAll(c *gin.Context) {
	var params dto.ListApplicantsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", err.Error(), nil)
		return
	}

	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}

	applicants, meta, err := h.usecase.GetAll(c.Request.Context(), params)
	if err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, applicants, &response.Meta{Pagination: meta})
}

// GetByID handles GET /applicants/:id
func (h *RecruitmentApplicantHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	applicant, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, applicant, nil)
}

// Create handles POST /applicants
func (h *RecruitmentApplicantHandler) Create(c *gin.Context) {
	var reqDTO dto.CreateRecruitmentApplicantDTO
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
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// Update handles PUT /applicants/:id
func (h *RecruitmentApplicantHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var reqDTO dto.UpdateRecruitmentApplicantDTO
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
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Delete handles DELETE /applicants/:id
func (h *RecruitmentApplicantHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{"message": "Applicant deleted successfully"}, nil)
}

// MoveStage handles POST /applicants/:id/move-stage
func (h *RecruitmentApplicantHandler) MoveStage(c *gin.Context) {
	id := c.Param("id")

	var reqDTO dto.MoveApplicantStageDTO
	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}
	if reqDTO.TargetStageID() == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", "to_stage_id or stage_id is required", nil)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.usecase.MoveStage(c.Request.Context(), id, &reqDTO, userID.(string))
	if err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// GetByStage handles GET /applicants/by-stage
func (h *RecruitmentApplicantHandler) GetByStage(c *gin.Context) {
	var params dto.ListApplicantsByStageParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", err.Error(), nil)
		return
	}

	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}

	byStage, err := h.usecase.GetByStage(c.Request.Context(), params)
	if err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, byStage, nil)
}

// GetStages handles GET /applicant-stages
func (h *RecruitmentApplicantHandler) GetStages(c *gin.Context) {
	stages, err := h.usecase.GetStages(c.Request.Context())
	if err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, stages, nil)
}

// GetActivities handles GET /applicants/:id/activities
func (h *RecruitmentApplicantHandler) GetActivities(c *gin.Context) {
	id := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	activities, meta, err := h.usecase.GetActivities(c.Request.Context(), id, page, perPage)
	if err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, activities, &response.Meta{Pagination: meta})
}

// AddActivity handles POST /applicants/:id/activities
func (h *RecruitmentApplicantHandler) AddActivity(c *gin.Context) {
	id := c.Param("id")

	var reqDTO dto.CreateApplicantActivityDTO
	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.usecase.AddActivity(c.Request.Context(), id, &reqDTO, userID.(string))
	if err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// GetByRecruitmentRequest handles GET /recruitment-requests/:id/applicants
func (h *RecruitmentApplicantHandler) GetByRecruitmentRequest(c *gin.Context) {
	recruitmentRequestID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	applicants, meta, err := h.usecase.GetByRecruitmentRequest(c.Request.Context(), recruitmentRequestID, page, perPage)
	if err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, applicants, &response.Meta{Pagination: meta})
}

// CanConvertToEmployee handles GET /applicants/:id/can-convert
func (h *RecruitmentApplicantHandler) CanConvertToEmployee(c *gin.Context) {
	id := c.Param("id")

	canConvert, reason, err := h.usecase.CanConvertToEmployee(c.Request.Context(), id)
	if err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{
		"can_convert": canConvert,
		"reason":      reason,
	}, nil)
}

// ConvertToEmployee handles POST /applicants/:id/convert-to-employee
func (h *RecruitmentApplicantHandler) ConvertToEmployee(c *gin.Context) {
	id := c.Param("id")

	var reqDTO dto.ConvertApplicantToEmployeeDTO
	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errors.ErrorResponse(c, "UNAUTHORIZED", nil, nil)
		return
	}

	result, err := h.usecase.ConvertToEmployee(c.Request.Context(), id, &reqDTO, userID.(string))
	if err != nil {
		handleApplicantError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// handleApplicantError maps usecase errors to appropriate HTTP responses
func handleApplicantError(c *gin.Context, err error) {
	msg := err.Error()

	switch {
	case msg == "applicant not found":
		errors.ErrorResponse(c, "APPLICANT_NOT_FOUND", map[string]any{"message": msg}, nil)
	case msg == "recruitment request not found":
		errors.ErrorResponse(c, "RECRUITMENT_REQUEST_NOT_FOUND", map[string]any{"message": msg}, nil)
	case msg == "stage not found":
		errors.ErrorResponse(c, "STAGE_NOT_FOUND", map[string]any{"message": msg}, nil)
	case msg == "target stage not found":
		errors.ErrorResponse(c, "TARGET_STAGE_NOT_FOUND", map[string]any{"message": msg}, nil)
	case msg == "invalid applicant source":
		errors.ErrorResponse(c, "INVALID_SOURCE", map[string]any{"message": msg}, nil)
	case msg == "applicant already converted to employee":
		errors.ErrorResponse(c, "ALREADY_CONVERTED", map[string]any{"message": msg}, nil)
	case msg == "applicant must be in hired stage to convert":
		errors.ErrorResponse(c, "NOT_IN_HIRED_STAGE", map[string]any{"message": msg}, nil)
	case msg == "applicant must be in hired stage":
		errors.ErrorResponse(c, "NOT_IN_HIRED_STAGE", map[string]any{"message": msg}, nil)
	case strings.Contains(msg, "cannot move applicant from terminal stage"):
		errors.ErrorResponse(c, "CANNOT_MOVE_FROM_TERMINAL", map[string]any{"message": msg}, nil)
	case strings.Contains(msg, "cannot delete stage with existing applicants"):
		errors.ErrorResponse(c, "STAGE_IN_USE", map[string]any{"message": msg}, nil)
	case strings.Contains(msg, "failed to create employee"):
		errors.ErrorResponse(c, "EMPLOYEE_CREATION_FAILED", map[string]any{"message": msg}, nil)
	case strings.Contains(msg, "failed to link applicant to employee"):
		errors.ErrorResponse(c, "LINK_FAILED", map[string]any{"message": msg}, nil)
	default:
		errors.ErrorResponse(c, "INTERNAL_ERROR", map[string]any{"message": "An unexpected error occurred"}, nil)
	}
}
