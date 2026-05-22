package dto

import "time"

type RoleResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type UserResponse struct {
	ID        string        `json:"id"`
	Email     string        `json:"email"`
	Name      string        `json:"name"`
	AvatarURL string        `json:"avatar_url"`
	RoleID    string        `json:"role_id"`
	Role      *RoleResponse `json:"role,omitempty"`
	Status    string        `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,min=3"`
	RoleID   string `json:"role_id" binding:"omitempty"`
	Status   string `json:"status" binding:"omitempty,oneof=active inactive"`
}

type UpdateUserRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Name   string `json:"name" binding:"omitempty,min=3"`
	RoleID string `json:"role_id" binding:"omitempty"`
	Status string `json:"status" binding:"omitempty,oneof=active inactive"`
}

type UpdateProfileRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required,min=3"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type AvailableUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type ListUsersRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search" binding:"omitempty"`
	Status  string `form:"status" binding:"omitempty,oneof=active inactive"`
	RoleID  string `form:"role_id" binding:"omitempty"`
}

type TenantDeletionScheduleResponse struct {
	TenantID            string `json:"tenant_id"`
	DeletionRequestedAt string `json:"deletion_requested_at"`
	DeletionScheduledAt string `json:"deletion_scheduled_at"`
	GracePeriodDays     int    `json:"grace_period_days"`
}
