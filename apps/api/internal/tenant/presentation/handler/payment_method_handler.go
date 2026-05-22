package handler

import (
	"errors"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	tenantUsecase "github.com/gilabs/gims/api/internal/tenant/domain/usecase"
	"github.com/gin-gonic/gin"
)

// PaymentMethodHandler manages saved Xendit card tokens for billing auto-renewal.
type PaymentMethodHandler struct {
	uc tenantUsecase.PaymentMethodUsecase
}

// NewPaymentMethodHandler creates a PaymentMethodHandler.
func NewPaymentMethodHandler(uc tenantUsecase.PaymentMethodUsecase) *PaymentMethodHandler {
	return &PaymentMethodHandler{uc: uc}
}

type addPaymentMethodRequest struct {
	XenditTokenID string `json:"xendit_token_id" binding:"required,max=255"`
}

// ListPaymentMethods returns all saved payment methods for the calling tenant.
//
// GET /api/v1/billing/payment-methods
func (h *PaymentMethodHandler) ListPaymentMethods(c *gin.Context) {
	tenantID, ok := c.Get("tenant_id")
	if !ok || tenantID == "" {
		coreErrors.UnauthorizedResponse(c, missingTenantContextMessage)
		return
	}

	methods, err := h.uc.List(c.Request.Context(), tenantID.(string))
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "failed to list payment methods")
		return
	}

	response.SuccessResponse(c, methods, nil)
}

// AddPaymentMethod stores a Xendit card token reference after verifying it with Xendit.
//
// POST /api/v1/billing/payment-methods
func (h *PaymentMethodHandler) AddPaymentMethod(c *gin.Context) {
	tenantID, ok := c.Get("tenant_id")
	if !ok || tenantID == "" {
		coreErrors.UnauthorizedResponse(c, missingTenantContextMessage)
		return
	}
	if !canModifyBilling(c) {
		coreErrors.ForbiddenResponse(c, "billing.change", nil)
		return
	}

	var req addPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.InvalidRequestBodyResponse(c)
		return
	}

	method, err := h.uc.Add(c.Request.Context(), tenantID.(string), req.XenditTokenID)
	if err != nil {
		switch {
		case errors.Is(err, tenantUsecase.ErrXenditTokenInvalid):
			coreErrors.ErrorResponse(c, "INVALID_CARD_TOKEN", nil, nil)
		case errors.Is(err, tenantUsecase.ErrXenditTokenAlreadySaved):
			coreErrors.ErrorResponse(c, "CARD_ALREADY_SAVED", nil, nil)
		default:
			coreErrors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	response.SuccessResponse(c, method, nil)
}

// SetDefaultPaymentMethod marks a saved payment method as the default for auto-renewal.
//
// PATCH /api/v1/billing/payment-methods/:id/default
func (h *PaymentMethodHandler) SetDefaultPaymentMethod(c *gin.Context) {
	tenantID, ok := c.Get("tenant_id")
	if !ok || tenantID == "" {
		coreErrors.UnauthorizedResponse(c, missingTenantContextMessage)
		return
	}
	if !canModifyBilling(c) {
		coreErrors.ForbiddenResponse(c, "billing.change", nil)
		return
	}

	methodID := c.Param("id")
	if err := h.uc.SetDefault(c.Request.Context(), methodID, tenantID.(string)); err != nil {
		if errors.Is(err, tenantUsecase.ErrPaymentMethodNotFound) {
			coreErrors.NotFoundResponse(c, "payment_method", methodID)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{"status": "updated"}, nil)
}

// RemovePaymentMethod soft-deletes a saved payment method.
//
// DELETE /api/v1/billing/payment-methods/:id
func (h *PaymentMethodHandler) RemovePaymentMethod(c *gin.Context) {
	tenantID, ok := c.Get("tenant_id")
	if !ok || tenantID == "" {
		coreErrors.UnauthorizedResponse(c, missingTenantContextMessage)
		return
	}
	if !canModifyBilling(c) {
		coreErrors.ForbiddenResponse(c, "billing.change", nil)
		return
	}

	methodID := c.Param("id")
	if err := h.uc.Remove(c.Request.Context(), methodID, tenantID.(string)); err != nil {
		if errors.Is(err, tenantUsecase.ErrPaymentMethodNotFound) {
			coreErrors.NotFoundResponse(c, "payment_method", methodID)
			return
		}
		coreErrors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{"status": "removed"}, nil)
}
