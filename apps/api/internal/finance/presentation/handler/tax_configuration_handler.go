package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
)

type TaxConfigurationHandler struct {
	uc usecase.TaxConfigurationUsecase
}

func NewTaxConfigurationHandler(uc usecase.TaxConfigurationUsecase) *TaxConfigurationHandler {
	return &TaxConfigurationHandler{uc: uc}
}

func (h *TaxConfigurationHandler) List(c *gin.Context) {
	var req dto.ListTaxConfigurationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	req.Page = page
	req.PerPage = perPage

	items, total, err := h.uc.List(c.Request.Context(), &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusInternalServerError, response.ErrCodeInternalServerError)
		return
	}
	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, items, meta)
}

func (h *TaxConfigurationHandler) Create(c *gin.Context) {
	var req dto.CreateTaxConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponseCreated(c, res, nil)
}

func (h *TaxConfigurationHandler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	companyID := strings.TrimSpace(c.Query("company_id"))
	if companyID == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id is required", nil)
		return
	}
	res, err := h.uc.GetByID(c.Request.Context(), id, companyID)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *TaxConfigurationHandler) Update(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	companyID := strings.TrimSpace(c.Query("company_id"))
	if companyID == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id is required", nil)
		return
	}
	var req dto.UpdateTaxConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
		return
	}
	res, err := h.uc.Update(c.Request.Context(), id, companyID, &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *TaxConfigurationHandler) ToggleStatus(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.ToggleTaxConfigurationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
		return
	}
	res, err := h.uc.ToggleStatus(c.Request.Context(), id, &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}
