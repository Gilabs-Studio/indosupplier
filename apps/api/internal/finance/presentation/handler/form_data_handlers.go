package handler

// This file contains GetFormData handlers for all finance modules.
// Each handler delegates to the corresponding usecase GetFormData method.

import (
	"net/http"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) GetFormData(c *gin.Context) {
	res, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "PAYMENT_FORM_DATA_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *CashBankJournalHandler) GetFormData(c *gin.Context) {
	res, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "CASH_BANK_FORM_DATA_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *BudgetHandler) GetFormData(c *gin.Context) {
	res, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "BUDGET_FORM_DATA_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *NonTradePayableHandler) GetFormData(c *gin.Context) {
	res, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "NON_TRADE_PAYABLE_FORM_DATA_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) GetFormData(c *gin.Context) {
	res, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ASSET_FORM_DATA_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetCategoryHandler) GetFormData(c *gin.Context) {
	res, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ASSET_CATEGORY_FORM_DATA_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *UpCountryCostHandler) GetFormData(c *gin.Context) {
	res, err := h.uc.GetFormData(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "UP_COUNTRY_COST_FORM_DATA_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}
