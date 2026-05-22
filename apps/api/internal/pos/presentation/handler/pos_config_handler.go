package handler

import (
	"errors"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/usecase"
	"github.com/gin-gonic/gin"
)

// POSConfigHandler handles POS configuration per outlet
type POSConfigHandler struct {
	uc usecase.POSConfigUsecase
}

// NewPOSConfigHandler creates the handler
func NewPOSConfigHandler(uc usecase.POSConfigUsecase) *POSConfigHandler {
	return &POSConfigHandler{uc: uc}
}

// GetByOutlet returns the POS config for the given outlet (auto-creates default if missing)
func (h *POSConfigHandler) GetByOutlet(c *gin.Context) {
	outletID := c.Param("outletID")
	cfg, err := h.uc.GetOrCreate(c.Request.Context(), outletID)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, cfg, nil)
}

// Upsert creates or updates the POS config for the given outlet
func (h *POSConfigHandler) Upsert(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	outletID := c.Param("outletID")
	var req dto.UpsertPOSConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	cfg, err := h.uc.Upsert(c.Request.Context(), outletID, &req, uc.userID)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, cfg, nil)
}

// UpdateReceiptWhatsAppTemplate updates only receipt WhatsApp template and is restricted to owner/admin scope.
func (h *POSConfigHandler) UpdateReceiptWhatsAppTemplate(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	outletID := c.Param("outletID")
	var req dto.UpdateReceiptWhatsAppTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	cfg, err := h.uc.UpdateReceiptWhatsAppTemplate(c.Request.Context(), outletID, &req, uc.userID, uc.isOwner)
	if err != nil {
		if errors.Is(err, usecase.ErrPOSConfigForbidden) {
			coreErrors.ForbiddenResponse(c, "pos.order.read", nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}

	response.SuccessResponse(c, cfg, nil)
}
