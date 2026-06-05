package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	coreErrors "github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/response"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/domain/mapper"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/domain/usecase"
)

type SystemAdminHandler struct {
	uc   usecase.SystemAdminUsecase
	repo repositories.SystemAdminRepository
}

const adminCookiePath = "/api/v1/sysadmin"

func NewSystemAdminHandler(uc usecase.SystemAdminUsecase, repo repositories.SystemAdminRepository) *SystemAdminHandler {
	return &SystemAdminHandler{
		uc:   uc,
		repo: repo,
	}
}

func (h *SystemAdminHandler) Login(c *gin.Context) {
	var req dto.SysadminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	res, err := h.uc.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			coreErrors.ErrorResponse(c, "INVALID_CREDENTIALS", nil, nil)
			return
		}
		if errors.Is(err, usecase.ErrAdminInactive) {
			coreErrors.ErrorResponse(c, "ACCOUNT_DISABLED", nil, nil)
			return
		}
		coreErrors.ErrorResponse(c, "INTERNAL_SERVER_ERROR", nil, nil)
		return
	}

	// Scope admin session cookies to sysadmin routes so they are not sent to user-facing APIs.
	c.SetCookie("indosupplier_admin_token", res.Token, int(res.ExpiresIn), adminCookiePath, "", false, true)

	response.SuccessResponse(c, res, nil)
}

func (h *SystemAdminHandler) Me(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		coreErrors.UnauthorizedResponse(c, "unauthenticated")
		return
	}

	admin, err := h.repo.FindByID(c.Request.Context(), adminID.(string))
	if err != nil {
		coreErrors.NotFoundResponse(c, "admin", adminID.(string))
		return
	}

	res := mapper.ToSysadminResponse(admin)
	response.SuccessResponse(c, res, nil)
}

func (h *SystemAdminHandler) Logout(c *gin.Context) {
	c.SetCookie("indosupplier_admin_token", "", -1, adminCookiePath, "", false, true)
	response.SuccessResponse(c, gin.H{"message": "Logged out successfully"}, nil)
}
