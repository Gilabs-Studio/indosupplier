package dto

import (
	"time"

	roleDto "github.com/gilabs/gims/api/internal/role/domain/dto"
)

// UserResponse represents user response DTO (without sensitive data)
type UserResponse struct {
	ID                   string                `json:"id"`
	Email                string                `json:"email"`
	Name                 string                `json:"name"`
	AvatarURL            string                `json:"avatar_url"`
	RoleID               string                `json:"role_id"`
	Role                 *roleDto.RoleResponse `json:"role,omitempty"`
	Status               string                `json:"status"`
	PasswordResetPending bool                  `json:"password_reset_pending"`
	CreatedAt            time.Time             `json:"created_at"`
	UpdatedAt            time.Time             `json:"updated_at"`
}

// CreateUserRequest represents create user request DTO
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,min=3"`
	RoleID   string `json:"role_id" binding:"required,uuid"`
	Status   string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// UpdateUserRequest represents update user request DTO
type UpdateUserRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Name   string `json:"name" binding:"omitempty,min=3"`
	RoleID string `json:"role_id" binding:"omitempty,uuid"`
	Status string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// UpdateProfileRequest represents update profile request DTO
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required,min=3"`
}

// ChangePasswordRequest represents change password request DTO
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// AvailableUserResponse is a lightweight DTO for the user-dropdown
// in the employee form (users not yet linked to any employee).
type AvailableUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ListUsersRequest represents list users query parameters
// Moved from entity.go to here as it is a request DTO
type ListUsersRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search" binding:"omitempty"`
	Status  string `form:"status" binding:"omitempty,oneof=active inactive"`
	RoleID  string `form:"role_id" binding:"omitempty,uuid"`
}

// TenantDeletionScheduleResponse describes a scheduled tenant data deletion window.
type TenantDeletionScheduleResponse struct {
	TenantID            string `json:"tenant_id"`
	DeletionRequestedAt string `json:"deletion_requested_at"`
	DeletionScheduledAt string `json:"deletion_scheduled_at"`
	GracePeriodDays     int    `json:"grace_period_days"`
}
