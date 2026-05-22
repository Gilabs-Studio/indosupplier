package handler

import (
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	resetDto "github.com/gilabs/gims/api/internal/password_reset/domain/dto"
	"github.com/gilabs/gims/api/internal/password_reset/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// PasswordResetHandler handles password reset requests.
type PasswordResetHandler struct {
	usecase usecase.PasswordResetUsecase
}

func NewPasswordResetHandler(uc usecase.PasswordResetUsecase) *PasswordResetHandler {
	return &PasswordResetHandler{usecase: uc}
}

func (h *PasswordResetHandler) ForgotPassword(c *gin.Context) {
	var req resetDto.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.usecase.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	var req resetDto.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.usecase.ResetPassword(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrInvalidResetToken:
			errors.ErrorResponse(c, "INVALID_RESET_TOKEN", map[string]interface{}{"reason": err.Error()}, nil)
			return
		case usecase.ErrResetTokenExpired:
			errors.ErrorResponse(c, "RESET_TOKEN_EXPIRED", map[string]interface{}{"reason": err.Error()}, nil)
			return
		case usecase.ErrResetTokenAlreadyUsed:
			errors.ErrorResponse(c, "RESET_TOKEN_ALREADY_USED", map[string]interface{}{"reason": err.Error()}, nil)
			return
		case usecase.ErrUserNotFound:
			errors.ErrorResponse(c, "USER_NOT_FOUND", nil, nil)
			return
		default:
			errors.InternalServerErrorResponse(c, err.Error())
			return
		}
	}

	response.SuccessResponse(c, result, nil)
}

func (h *PasswordResetHandler) ValidateResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		errors.ErrorResponse(c, "MISSING_TOKEN", map[string]interface{}{"field": "token"}, nil)
		return
	}

	err := h.usecase.ValidateResetToken(c.Request.Context(), token)
	if err != nil {
		switch err {
		case usecase.ErrInvalidResetToken:
			errors.ErrorResponse(c, "INVALID_RESET_TOKEN", map[string]interface{}{"reason": err.Error()}, nil)
			return
		case usecase.ErrResetTokenExpired:
			errors.ErrorResponse(c, "RESET_TOKEN_EXPIRED", map[string]interface{}{"reason": err.Error()}, nil)
			return
		case usecase.ErrResetTokenAlreadyUsed:
			errors.ErrorResponse(c, "RESET_TOKEN_ALREADY_USED", map[string]interface{}{"reason": err.Error()}, nil)
			return
		default:
			errors.InternalServerErrorResponse(c, err.Error())
			return
		}
	}

	response.SuccessResponse(c, map[string]string{"message": "Token is valid"}, nil)
}
