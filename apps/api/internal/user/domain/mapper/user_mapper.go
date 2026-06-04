package mapper

import (
	"github.com/gilabs/indosupplier/api/internal/user/data/models"
	"github.com/gilabs/indosupplier/api/internal/user/domain/dto"
)

func ToAvailableUserResponse(u *models.User) dto.AvailableUserResponse {
	return dto.AvailableUserResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
	}
}

func ToUserResponse(u *models.User) *dto.UserResponse {
	resp := &dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	return resp
}
