package mapper

import (
	"github.com/gilabs/gims/api/internal/password_reset/data/models"
	"github.com/gilabs/gims/api/internal/password_reset/domain/dto"
)

// ToPasswordResetRequestResponse converts a model to a response DTO
func ToPasswordResetRequestResponse(req *models.PasswordResetRequest) *dto.PasswordResetRequestResponse {
	if req == nil {
		return nil
	}

	return &dto.PasswordResetRequestResponse{
		ID:        req.ID,
		UserID:    req.UserID,
		Status:    req.Status,
		ExpiresAt: req.ExpiresAt,
		UsedAt:    req.UsedAt,
		CreatedAt: req.CreatedAt,
	}
}

// ToForgotPasswordResponse converts to forgot password response
func ToForgotPasswordResponse(email string) *dto.ForgotPasswordResponse {
	return &dto.ForgotPasswordResponse{
		Message: "Password reset link has been sent to your email",
		Email:   email,
	}
}

// ToResetPasswordResponse converts to reset password response
func ToResetPasswordResponse(email string) *dto.ResetPasswordResponse {
	return &dto.ResetPasswordResponse{
		Message: "Password has been reset successfully",
		Email:   email,
	}
}
