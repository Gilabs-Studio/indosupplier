package handler

import (
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
)

type InventorySettingsHandler struct {
	uc usecase.InventorySettingsUsecase
}

func NewInventorySettingsHandler(uc usecase.InventorySettingsUsecase) *InventorySettingsHandler {
	return &InventorySettingsHandler{uc: uc}
}

func (h *InventorySettingsHandler) Get(c *gin.Context) {
	companyID := strings.TrimSpace(c.Query("company_id"))
	if companyID == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id is required", nil)
		return
	}
	res, err := h.uc.GetByCompanyID(c.Request.Context(), companyID)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusInternalServerError, response.ErrCodeInternalServerError)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *InventorySettingsHandler) Upsert(c *gin.Context) {
	var req dto.UpdateInventorySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, err.Error(), nil)
		return
	}
	res, err := h.uc.Upsert(c.Request.Context(), &req)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *InventorySettingsHandler) GetAverageCost(c *gin.Context) {
	companyID := strings.TrimSpace(c.Query("company_id"))
	productID := strings.TrimSpace(c.Param("product_id"))
	if companyID == "" || productID == "" {
		response.StandardErrorResponse(c, http.StatusBadRequest, response.ErrCodeValidationError, "company_id and product_id are required", nil)
		return
	}
	res, err := h.uc.GetAverageCost(c.Request.Context(), companyID, productID)
	if err != nil {
		writeFinanceStandardizedError(c, err, http.StatusBadRequest, response.ErrCodeBadRequest)
		return
	}
	response.SuccessResponse(c, res, nil)
}
