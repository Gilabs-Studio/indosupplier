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

type FiscalYearHandler struct {
	uc usecase.FiscalYearUsecase
}

func NewFiscalYearHandler(uc usecase.FiscalYearUsecase) *FiscalYearHandler {
	return &FiscalYearHandler{uc: uc}
}

func (h *FiscalYearHandler) List(c *gin.Context) {
	var req dto.ListFiscalYearsRequest
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

func (h *FiscalYearHandler) Create(c *gin.Context) {
	var req dto.CreateFiscalYearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
		return
	}
	createdBy := getUserID(c)
	res, err := h.uc.Create(c.Request.Context(), &req, createdBy)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponseCreated(c, res, nil)
}

func (h *FiscalYearHandler) GetByID(c *gin.Context) {
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

func (h *FiscalYearHandler) Update(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	companyID := strings.TrimSpace(c.Query("company_id"))
	if companyID == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id is required", nil)
		return
	}
	var req dto.UpdateFiscalYearRequest
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

func (h *FiscalYearHandler) Delete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	companyID := strings.TrimSpace(c.Query("company_id"))
	if companyID == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id is required", nil)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id, companyID); err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}

	response.SuccessResponse(c, gin.H{"deleted": true, "id": id, "company_id": companyID}, nil)
}

func (h *FiscalYearHandler) Activate(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	companyID := strings.TrimSpace(c.Query("company_id"))
	if companyID == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id is required", nil)
		return
	}
	res, err := h.uc.Activate(c.Request.Context(), id, companyID)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *FiscalYearHandler) Lock(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	companyID := strings.TrimSpace(c.Query("company_id"))
	if companyID == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id is required", nil)
		return
	}
	res, err := h.uc.Lock(c.Request.Context(), id, companyID)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func getUserID(c *gin.Context) *string {
	if userIDVal, exists := c.Get("user_id"); exists {
		if userID, ok := userIDVal.(string); ok && strings.TrimSpace(userID) != "" {
			id := strings.TrimSpace(userID)
			return &id
		}
	}
	return nil
}
