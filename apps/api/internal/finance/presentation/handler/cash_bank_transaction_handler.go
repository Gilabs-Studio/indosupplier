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

type CashBankTransactionHandler struct {
	uc usecase.CashBankTransactionUsecase
}

func NewCashBankTransactionHandler(uc usecase.CashBankTransactionUsecase) *CashBankTransactionHandler {
	return &CashBankTransactionHandler{uc: uc}
}

func (h *CashBankTransactionHandler) Create(c *gin.Context) {
	var req dto.CreateCashBankTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "CASH_BANK_TRANSACTION_CREATE_FAILED")
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

func (h *CashBankTransactionHandler) List(c *gin.Context) {
	var req dto.ListCashBankTransactionsRequest
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
		writeFinanceStandardizedError(c, err, http.StatusInternalServerError, "CASH_BANK_TRANSACTION_LIST_FAILED")
		return
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, items, meta)
}

func (h *CashBankTransactionHandler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusNotFound, "CASH_BANK_TRANSACTION_NOT_FOUND")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *CashBankTransactionHandler) Reverse(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.ReverseCashBankTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.Reverse(c.Request.Context(), id, &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "CASH_BANK_TRANSACTION_REVERSE_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *CashBankTransactionHandler) GetFormData(c *gin.Context) {
	companyID := strings.TrimSpace(c.Query("company_id"))
	res, err := h.uc.GetFormData(c.Request.Context(), companyID)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusInternalServerError, "CASH_BANK_TRANSACTION_FORM_DATA_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}
