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

type BankTransferHandler struct {
	uc usecase.BankTransferUsecase
}

func NewBankTransferHandler(uc usecase.BankTransferUsecase) *BankTransferHandler {
	return &BankTransferHandler{uc: uc}
}

func (h *BankTransferHandler) Create(c *gin.Context) {
	var req dto.CreateBankTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "BANK_TRANSFER_CREATE_FAILED")
		return
	}

	response.SuccessResponseCreated(c, res, nil)
}

func (h *BankTransferHandler) List(c *gin.Context) {
	var req dto.ListBankTransfersRequest
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
		writeFinanceStandardizedError(c, err, http.StatusInternalServerError, "BANK_TRANSFER_LIST_FAILED")
		return
	}

	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, items, meta)
}

func (h *BankTransferHandler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusNotFound, "BANK_TRANSFER_NOT_FOUND")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *BankTransferHandler) Complete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.Complete(c.Request.Context(), id)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "BANK_TRANSFER_COMPLETE_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}

func (h *BankTransferHandler) Cancel(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.CancelBankTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.Cancel(c.Request.Context(), id, &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, "BANK_TRANSFER_CANCEL_FAILED")
		return
	}

	response.SuccessResponse(c, res, nil)
}
