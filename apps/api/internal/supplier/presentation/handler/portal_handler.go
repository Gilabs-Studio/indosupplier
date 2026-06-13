package handler

import (
	stderrors "errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/usecase"
)

type SupplierPortalHandler struct {
	portalUC usecase.PortalUsecase
}

func NewSupplierPortalHandler(portalUC usecase.PortalUsecase) *SupplierPortalHandler {
	return &SupplierPortalHandler{portalUC: portalUC}
}

func (h *SupplierPortalHandler) getAuthenticatedUserID(c *gin.Context) (string, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		errors.UnauthorizedResponse(c, "unauthorized")
		return "", false
	}
	userID, ok := userIDVal.(string)
	if !ok {
		errors.InternalServerErrorResponse(c, "invalid user ID structure in context")
		return "", false
	}
	return userID, true
}

func (h *SupplierPortalHandler) GetProfile(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	profile, err := h.portalUC.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrProfileNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, profile, nil)
}

func (h *SupplierPortalHandler) UpdateProfile(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	profile, err := h.portalUC.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrProfileNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, profile, &response.Meta{UpdatedBy: userID})
}

func (h *SupplierPortalHandler) GetBillingOverview(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	overview, err := h.portalUC.GetBillingOverview(c.Request.Context(), userID)
	if err != nil {
		if stderrors.Is(err, usecase.ErrProfileNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, overview, nil)
}

func (h *SupplierPortalHandler) UpgradeSubscriptionPlan(c *gin.Context) {
	userID, ok := h.getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var req dto.UpgradePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	overview, err := h.portalUC.UpgradePlan(c.Request.Context(), userID, &req)
	if err != nil {
		if stderrors.Is(err, usecase.ErrProfileNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"user_id": userID}, nil)
			return
		}
		if stderrors.Is(err, usecase.ErrPlanNotFound) {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{"plan_id": req.PlanID}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, overview, &response.Meta{UpdatedBy: userID})
}
