package handler

import (
	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/usecase"
	"github.com/gin-gonic/gin"
)

type POSDeviceTokenHandler struct {
	uc usecase.POSDeviceTokenUsecase
}

func NewPOSDeviceTokenHandler(uc usecase.POSDeviceTokenUsecase) *POSDeviceTokenHandler {
	return &POSDeviceTokenHandler{uc: uc}
}

func (h *POSDeviceTokenHandler) Register(c *gin.Context) {
	var req dto.RegisterPOSDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}
	result, err := h.uc.Register(c.Request.Context(), c.GetString("user_id"), &req)
	if err != nil {
		if err == usecase.ErrPOSDeviceTokenForbidden {
			coreErrors.ForbiddenResponse(c, "pos.order.read", nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, result, nil)
}
