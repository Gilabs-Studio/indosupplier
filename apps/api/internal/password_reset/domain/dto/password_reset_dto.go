package dto

import "time"

// ForgotPasswordRequest represents a forgot password request.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPasswordResponse represents a forgot password response.
type ForgotPasswordResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

// ResetPasswordRequest represents a reset password request.
type ResetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// ResetPasswordResponse represents a reset password response.
type ResetPasswordResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

// PasswordResetRequestResponse represents a password reset request response.
type PasswordResetRequestResponse struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Status    string     `json:"status"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}
