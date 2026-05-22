package dto

import "time"

// UserResponseDTO represents user response
type UserResponseDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	RoleID    string    `json:"role_id"`
	Role      RoleDTO   `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type RoleDTO struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CreateUserRequestDTO represents create user payload
type CreateUserRequestDTO struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	RoleID   string `json:"role_id" binding:"required,uuid"`
}

// UpdateUserRequestDTO represents update user payload
type UpdateUserRequestDTO struct {
	Name   string `json:"name"`
	RoleID string `json:"role_id" binding:"omitempty,uuid"`
	Status string `json:"status" binding:"omitempty,oneof=active inactive suspended"`
}
