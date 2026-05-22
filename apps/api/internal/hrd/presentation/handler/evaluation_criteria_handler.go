package handler

import (
	"net/http"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/usecase"
	"github.com/gin-gonic/gin"
)

type EvaluationCriteriaHandler struct {
	usecase usecase.EvaluationCriteriaUsecase
}

// NewEvaluationCriteriaHandler creates a new instance of EvaluationCriteriaHandler
func NewEvaluationCriteriaHandler(usecase usecase.EvaluationCriteriaUsecase) *EvaluationCriteriaHandler {
	return &EvaluationCriteriaHandler{
		usecase: usecase,
	}
}

// GetByGroupID retrieves all criteria for a specific evaluation group
// GET /hrd/evaluation-criteria/group/:group_id
func (h *EvaluationCriteriaHandler) GetByGroupID(c *gin.Context) {
	groupID := c.Param("group_id")

	criteria, err := h.usecase.GetByGroupID(c.Request.Context(), groupID)
	if err != nil {
		handleEvaluationCriteriaError(c, err)
		return
	}

	response.SuccessResponse(c, criteria, nil)
}

// GetByID retrieves an evaluation criteria by ID
// GET /hrd/evaluation-criteria/:id
func (h *EvaluationCriteriaHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	criteria, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		handleEvaluationCriteriaError(c, err)
		return
	}

	response.SuccessResponse(c, criteria, nil)
}

// Create creates a new evaluation criteria
// POST /hrd/evaluation-criteria
func (h *EvaluationCriteriaHandler) Create(c *gin.Context) {
	var req dto.CreateEvaluationCriteriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	criteria, err := h.usecase.Create(c.Request.Context(), &req)
	if err != nil {
		handleEvaluationCriteriaError(c, err)
		return
	}

	response.SuccessResponse(c, criteria, nil)
}

// Update updates an existing evaluation criteria
// PUT /hrd/evaluation-criteria/:id
func (h *EvaluationCriteriaHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateEvaluationCriteriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error(), nil)
		return
	}

	criteria, err := h.usecase.Update(c.Request.Context(), id, &req)
	if err != nil {
		handleEvaluationCriteriaError(c, err)
		return
	}

	response.SuccessResponse(c, criteria, nil)
}

// Delete performs soft delete on an evaluation criteria
// DELETE /hrd/evaluation-criteria/:id
func (h *EvaluationCriteriaHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		handleEvaluationCriteriaError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{"message": "Evaluation criteria deleted successfully"}, nil)
}

// handleEvaluationCriteriaError handles errors and returns appropriate HTTP responses
func handleEvaluationCriteriaError(c *gin.Context, err error) {
	switch err.Error() {
	case "evaluation criteria not found":
		errors.ErrorResponse(c, "EVALUATION_CRITERIA_NOT_FOUND", map[string]interface{}{"message": "Evaluation criteria not found"}, nil)
	case "evaluation group not found":
		errors.ErrorResponse(c, "EVALUATION_GROUP_NOT_FOUND", map[string]interface{}{"message": "Evaluation group not found"}, nil)
	default:
		// Handle weight validation errors
		if len(err.Error()) > 20 && err.Error()[:20] == "total weight would e" {
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"message": err.Error()}, nil)
		} else {
			errors.ErrorResponse(c, "INTERNAL_ERROR", map[string]interface{}{"message": "An unexpected error occurred"}, nil)
		}
	}
}
