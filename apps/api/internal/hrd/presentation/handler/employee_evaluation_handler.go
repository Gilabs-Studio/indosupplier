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
	"github.com/google/uuid"
)

type EmployeeEvaluationHandler struct {
	usecase usecase.EmployeeEvaluationUsecase
}

const invalidEmployeeEvaluationRequestBodyMessage = "Invalid request body"

// NewEmployeeEvaluationHandler creates a new instance of EmployeeEvaluationHandler
func NewEmployeeEvaluationHandler(usecase usecase.EmployeeEvaluationUsecase) *EmployeeEvaluationHandler {
	return &EmployeeEvaluationHandler{
		usecase: usecase,
	}
}

// GetAll retrieves all employee evaluations with pagination and filters
// GET /hrd/employee-evaluations?page=1&per_page=20&employee_id=uuid&evaluation_group_id=uuid&evaluation_type=SELF
func (h *EmployeeEvaluationHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	search := c.Query("search")
	employeeID := c.Query("employee_id")
	evaluationGroupID := c.Query("evaluation_group_id")
	evaluationType := c.Query("evaluation_type")

	evaluations, meta, err := h.usecase.GetAll(c.Request.Context(), page, perPage, search, employeeID, evaluationGroupID, evaluationType)
	if err != nil {
		handleEmployeeEvaluationError(c, err)
		return
	}

	response.SuccessResponse(c, evaluations, &response.Meta{Pagination: meta})
}

// GetByID retrieves an employee evaluation by ID (with details)
// GET /hrd/employee-evaluations/:id
func (h *EmployeeEvaluationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	evaluation, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		handleEmployeeEvaluationError(c, err)
		return
	}

	response.SuccessResponse(c, evaluation, nil)
}

// GetFormData retrieves form dropdown data
// GET /hrd/employee-evaluations/form-data
func (h *EmployeeEvaluationHandler) GetFormData(c *gin.Context) {
	formData, err := h.usecase.GetFormData(c.Request.Context())
	if err != nil {
		handleEmployeeEvaluationError(c, err)
		return
	}

	response.SuccessResponse(c, formData, nil)
}

// Create creates a new employee evaluation
// POST /hrd/employee-evaluations
func (h *EmployeeEvaluationHandler) Create(c *gin.Context) {
	var req dto.CreateEmployeeEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", invalidEmployeeEvaluationRequestBodyMessage, err.Error(), nil)
		return
	}

	evaluation, err := h.usecase.Create(c.Request.Context(), &req)
	if err != nil {
		handleEmployeeEvaluationError(c, err)
		return
	}

	response.SuccessResponse(c, evaluation, nil)
}

// Update updates an existing employee evaluation
// PUT /hrd/employee-evaluations/:id
func (h *EmployeeEvaluationHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateEmployeeEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", invalidEmployeeEvaluationRequestBodyMessage, err.Error(), nil)
		return
	}

	evaluation, err := h.usecase.Update(c.Request.Context(), id, &req)
	if err != nil {
		handleEmployeeEvaluationError(c, err)
		return
	}

	response.SuccessResponse(c, evaluation, nil)
}

// Delete performs soft delete on an employee evaluation
// DELETE /hrd/employee-evaluations/:id
func (h *EmployeeEvaluationHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		handleEmployeeEvaluationError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{"message": "Employee evaluation deleted successfully"}, nil)
}

// AuditTrail retrieves paginated audit trail rows for an employee evaluation.
// GET /hrd/employee-evaluations/:id/audit-trail
func (h *EmployeeEvaluationHandler) AuditTrail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "ID is required"}, nil)
		return
	}
	if _, err := uuid.Parse(id); err != nil {
		errors.ErrorResponse(c, "INVALID_PATH_PARAM", map[string]interface{}{"message": "Invalid ID format"}, nil)
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

	items, total, err := h.usecase.ListAuditTrail(c.Request.Context(), id, page, perPage)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
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

// handleEmployeeEvaluationError handles errors and returns appropriate HTTP responses
func handleEmployeeEvaluationError(c *gin.Context, err error) {
	switch err.Error() {
	case "employee not found":
		errors.ErrorResponse(c, "EMPLOYEE_NOT_FOUND", map[string]interface{}{"message": "Employee not found"}, nil)
	case "evaluator not found":
		errors.ErrorResponse(c, "EVALUATOR_NOT_FOUND", map[string]interface{}{"message": "Evaluator not found"}, nil)
	case "employee evaluation not found":
		errors.ErrorResponse(c, "EMPLOYEE_EVALUATION_NOT_FOUND", map[string]interface{}{"message": "Employee evaluation not found"}, nil)
	case "evaluation group not found":
		errors.ErrorResponse(c, "EVALUATION_GROUP_NOT_FOUND", map[string]interface{}{"message": "Evaluation group not found"}, nil)
	case "evaluation group is not active":
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": "Evaluation group is not active"}, nil)
	case "period_end must be after period_start":
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": "Period end must be after period start"}, nil)
	case "invalid period_start format, must be YYYY-MM-DD",
		"invalid period_end format, must be YYYY-MM-DD":
		errors.ErrorResponse(c, "INVALID_DATE_FORMAT", map[string]interface{}{"message": err.Error()}, nil)
	default:
		// Handle dynamic error messages
		if strings.Contains(err.Error(), "does not belong to") {
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": err.Error()}, nil)
		} else {
			errors.ErrorResponse(c, "INTERNAL_ERROR", map[string]interface{}{"message": "An unexpected error occurred"}, nil)
		}
	}
}
