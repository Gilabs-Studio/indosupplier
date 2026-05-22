package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
)

type BankReconciliationHandler struct {
	uc usecase.BankReconciliationUsecase
}

func NewBankReconciliationHandler(uc usecase.BankReconciliationUsecase) *BankReconciliationHandler {
	return &BankReconciliationHandler{uc: uc}
}

func (h *BankReconciliationHandler) Import(c *gin.Context) {
	var req dto.ImportBankStatementRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "file is required", nil, nil)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "PARSE_ERROR")
		return
	}

	res, err := h.uc.ImportStatement(c.Request.Context(), &req, header.Filename, content)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "BANK_RECONCILIATION_IMPORT_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *BankReconciliationHandler) List(c *gin.Context) {
	var req dto.ListBankReconciliationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
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
		writeFinanceStandardizedError(c, err, http.StatusInternalServerError, "BANK_RECONCILIATION_LIST_FAILED")
		return
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, items, meta)
}

func (h *BankReconciliationHandler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusNotFound, "BANK_RECONCILIATION_NOT_FOUND")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *BankReconciliationHandler) AutoMatch(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.AutoMatch(c.Request.Context(), id)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "BANK_RECONCILIATION_AUTO_MATCH_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *BankReconciliationHandler) MatchLine(c *gin.Context) {
	reconciliationID := strings.TrimSpace(c.Param("id"))
	lineID := strings.TrimSpace(c.Param("line_id"))
	var req dto.MatchBankStatementLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.MatchLine(c.Request.Context(), reconciliationID, lineID, &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "BANK_RECONCILIATION_MATCH_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *BankReconciliationHandler) ExcludeLine(c *gin.Context) {
	reconciliationID := strings.TrimSpace(c.Param("id"))
	lineID := strings.TrimSpace(c.Param("line_id"))
	var req dto.ExcludeBankStatementLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.ExcludeLine(c.Request.Context(), reconciliationID, lineID, &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "BANK_RECONCILIATION_EXCLUDE_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *BankReconciliationHandler) Confirm(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.Confirm(c.Request.Context(), id)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "BANK_RECONCILIATION_CONFIRM_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *BankReconciliationHandler) Lock(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.Lock(c.Request.Context(), id)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "BANK_RECONCILIATION_LOCK_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *BankReconciliationHandler) GetFormData(c *gin.Context) {
	companyID := strings.TrimSpace(c.Query("company_id"))
	res, err := h.uc.GetFormData(c.Request.Context(), companyID)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusInternalServerError, "BANK_RECONCILIATION_FORM_DATA_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}
