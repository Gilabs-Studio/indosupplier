package mapper

import (
	roleMapper "github.com/gilabs/gims/api/internal/role/domain/mapper"
	"github.com/gilabs/gims/api/internal/user/data/models"
	"github.com/gilabs/gims/api/internal/user/domain/dto"
)

// ToAvailableUserResponse converts User to the lightweight AvailableUserResponse.
func ToAvailableUserResponse(u *models.User) dto.AvailableUserResponse {
	return dto.AvailableUserResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
	}
}

// ToUserResponse converts User to UserResponse
func ToUserResponse(u *models.User) *dto.UserResponse {
	resp := &dto.UserResponse{
		ID:                   u.ID,
		Email:                u.Email,
		Name:                 u.Name,
		AvatarURL:            u.AvatarURL,
		RoleID:               u.RoleID,
		Status:               u.Status,
		PasswordResetPending: u.PasswordResetPending,
		CreatedAt:            u.CreatedAt,
		UpdatedAt:            u.UpdatedAt,
	}
	if u.Role != nil {
		resp.Role = roleMapper.ToRoleResponse(u.Role)
	}
	return resp
}
