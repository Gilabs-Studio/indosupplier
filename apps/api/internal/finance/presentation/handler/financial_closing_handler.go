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

type FinancialClosingHandler struct {
	uc usecase.FinancialClosingUsecase
}

func closingRouteID(c *gin.Context) string {
	id := strings.TrimSpace(c.Param("id"))
	if id != "" {
		return id
	}
	return strings.TrimSpace(c.Param("period_id"))
}

func NewFinancialClosingHandler(uc usecase.FinancialClosingUsecase) *FinancialClosingHandler {
	return &FinancialClosingHandler{uc: uc}
}

func (h *FinancialClosingHandler) Create(c *gin.Context) {
	var req dto.CreateFinancialClosingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "FINANCIAL_CLOSING_CREATE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponseCreated(c, res, nil)
}

func (h *FinancialClosingHandler) Approve(c *gin.Context) {
	id := closingRouteID(c)
	res, err := h.uc.Approve(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "FINANCIAL_CLOSING_APPROVE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *FinancialClosingHandler) GetByID(c *gin.Context) {
	id := closingRouteID(c)
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "FINANCIAL_CLOSING_NOT_FOUND", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *FinancialClosingHandler) List(c *gin.Context) {
	var req dto.ListFinancialClosingsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
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
	req.Page = page
	req.PerPage = perPage

	items, total, err := h.uc.List(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "FINANCIAL_CLOSING_LIST_FAILED", err.Error(), nil, nil)
		return
	}
	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, items, meta)
}

func (h *FinancialClosingHandler) GetAnalysis(c *gin.Context) {
	id := closingRouteID(c)
	res, err := h.uc.GetAnalysis(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "FINANCIAL_CLOSING_ANALYSIS_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *FinancialClosingHandler) Reopen(c *gin.Context) {
	id := closingRouteID(c)
	res, err := h.uc.Reopen(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "FINANCIAL_CLOSING_REOPEN_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *FinancialClosingHandler) YearEndClose(c *gin.Context) {
	var req dto.YearEndCloseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.YearEndClose(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "YEAR_END_CLOSE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}
func (h *FinancialClosingHandler) Delete(c *gin.Context) {
	id := closingRouteID(c)
	err := h.uc.Delete(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "FINANCIAL_CLOSING_DELETE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, nil, nil)
}
