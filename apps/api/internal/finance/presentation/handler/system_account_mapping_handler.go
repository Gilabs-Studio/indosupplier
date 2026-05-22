package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SystemAccountMappingHandler struct {
	uc usecase.SystemAccountMappingUsecase
}

func mappingKeyParam(c *gin.Context) string {
	key := strings.TrimSpace(c.Param("key"))
	if key != "" {
		return key
	}
	return strings.TrimSpace(c.Param("id"))
}

func NewSystemAccountMappingHandler(uc usecase.SystemAccountMappingUsecase) *SystemAccountMappingHandler {
	return &SystemAccountMappingHandler{uc: uc}
}

func (h *SystemAccountMappingHandler) List(c *gin.Context) {
	companyID, ok := parseMappingCompanyID(c)
	if !ok {
		return
	}

	mappings, err := h.uc.List(c.Request.Context(), companyID)
	if err != nil {
		response.StandardErrorResponse(c, http.StatusInternalServerError, response.ErrCodeInternalServerError, "failed to list system account mappings", map[string]interface{}{"cause": err.Error()})
		return
	}

	response.SuccessResponse(c, mappings, nil)
}

func (h *SystemAccountMappingHandler) GetByKey(c *gin.Context) {
	key := mappingKeyParam(c)
	if key == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "key is required", map[string]interface{}{"key": "is required"})
		return
	}

	companyID, ok := parseMappingCompanyID(c)
	if !ok {
		return
	}

	mapping, err := h.uc.GetByKey(c.Request.Context(), key, companyID)
	if err != nil {
		h.handleMappingError(c, err)
		return
	}

	response.SuccessResponse(c, mapping, nil)
}

func (h *SystemAccountMappingHandler) Upsert(c *gin.Context) {
	key := mappingKeyParam(c)
	if key == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "key is required", map[string]interface{}{"key": "is required"})
		return
	}

	companyID, ok := parseMappingCompanyID(c)
	if !ok {
		return
	}

	var req dto.UpsertSystemAccountMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "invalid request body", map[string]interface{}{"cause": err.Error()})
		return
	}

	mapping, err := h.uc.Upsert(c.Request.Context(), key, req.COACode, req.Label, companyID)
	if err != nil {
		h.handleMappingError(c, err)
		return
	}

	response.SuccessResponse(c, mapping, nil)
}

func (h *SystemAccountMappingHandler) Delete(c *gin.Context) {
	key := mappingKeyParam(c)
	if key == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "key is required", map[string]interface{}{"key": "is required"})
		return
	}

	companyID, ok := parseMappingCompanyID(c)
	if !ok {
		return
	}

	if err := h.uc.Delete(c.Request.Context(), key, companyID); err != nil {
		h.handleMappingError(c, err)
		return
	}

	response.SuccessResponse(c, gin.H{"deleted": true, "key": key, "company_id": companyID}, nil)
}

func (h *SystemAccountMappingHandler) handleMappingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrMappingValidation):
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
	case errors.Is(err, usecase.ErrAccountNotPostable):
		response.StandardErrorResponse(c, http.StatusUnprocessableEntity, response.ErrCodeAccountNotPostable, err.Error(), nil)
	case errors.Is(err, usecase.ErrAccountInactive):
		response.StandardErrorResponse(c, http.StatusUnprocessableEntity, response.ErrCodeAccountInactive, err.Error(), nil)
	case errors.Is(err, usecase.ErrMappingNotConfigured):
		response.StandardErrorResponse(c, http.StatusUnprocessableEntity, response.ErrCodeMappingNotConfigured, err.Error(), nil)
	default:
		response.StandardErrorResponse(c, http.StatusInternalServerError, response.ErrCodeInternalServerError, "system account mapping operation failed", map[string]interface{}{"cause": err.Error()})
	}
}

func parseMappingCompanyID(c *gin.Context) (*string, bool) {
	companyID := strings.TrimSpace(c.Query("company_id"))
	if companyID == "" {
		return nil, true
	}

	if _, err := uuid.Parse(companyID); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id must be a valid UUID", map[string]interface{}{"company_id": "invalid UUID"})
		return nil, false
	}

	return &companyID, true
}
