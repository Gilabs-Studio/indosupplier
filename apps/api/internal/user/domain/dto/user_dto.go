package dto

import "time"

type AccountCapabilitiesResponse struct {
	Buyer    bool `json:"buyer"`
	Supplier bool `json:"supplier"`
}

type AccountProfileRefResponse struct {
	ID     string `json:"id"`
	Status string `json:"status,omitempty"`
}

type UserResponse struct {
	ID              string                      `json:"id"`
	Email           string                      `json:"email"`
	Name            string                      `json:"name"`
	AvatarURL       string                      `json:"avatar_url"`
	Status          string                      `json:"status"`
	Capabilities    AccountCapabilitiesResponse `json:"capabilities"`
	BuyerProfile    *AccountProfileRefResponse  `json:"buyer_profile,omitempty"`
	SupplierProfile *AccountProfileRefResponse  `json:"supplier_profile,omitempty"`
	CreatedAt       time.Time                   `json:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,min=3"`
	Status   string `json:"status" binding:"omitempty,oneof=active inactive"`
}

type UpdateUserRequest struct {
	Email  string `json:"email" binding:"omitempty,email"`
	Name   string `json:"name" binding:"omitempty,min=3"`
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
}

type AccountContext struct {
	BuyerProfileID    string
	SupplierProfileID string
	SupplierStatus    string
}
