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

type NonTradePayableHandler struct {
	uc usecase.NonTradePayableUsecase
}

func NewNonTradePayableHandler(uc usecase.NonTradePayableUsecase) *NonTradePayableHandler {
	return &NonTradePayableHandler{uc: uc}
}

func (h *NonTradePayableHandler) Create(c *gin.Context) {
	var req dto.CreateNonTradePayableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Create(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "NON_TRADE_PAYABLE_CREATE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponseCreated(c, res, nil)
}

func (h *NonTradePayableHandler) Update(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.UpdateNonTradePayableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "NON_TRADE_PAYABLE_UPDATE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *NonTradePayableHandler) Delete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "NON_TRADE_PAYABLE_DELETE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponseDeleted(c, "non_trade_payable", id, nil)
}

func (h *NonTradePayableHandler) Submit(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.Submit(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "NON_TRADE_PAYABLE_SUBMIT_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *NonTradePayableHandler) Post(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.Post(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "NON_TRADE_PAYABLE_POST_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *NonTradePayableHandler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "NON_TRADE_PAYABLE_NOT_FOUND", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *NonTradePayableHandler) List(c *gin.Context) {
	var req dto.ListNonTradePayablesRequest
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
		response.ErrorResponse(c, http.StatusInternalServerError, "NON_TRADE_PAYABLE_LIST_FAILED", err.Error(), nil, nil)
		return
	}
	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, items, meta)
}

func (h *NonTradePayableHandler) Approve(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.Approve(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "NON_TRADE_PAYABLE_APPROVE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *NonTradePayableHandler) Reject(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.Reject(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "NON_TRADE_PAYABLE_REJECT_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *NonTradePayableHandler) Cancel(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.Cancel(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "NON_TRADE_PAYABLE_CANCEL_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *NonTradePayableHandler) Pay(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.PayNonTradePayableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.Pay(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "NON_TRADE_PAYABLE_PAY_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *NonTradePayableHandler) PostPayment(c *gin.Context) {
	h.Pay(c)
}
