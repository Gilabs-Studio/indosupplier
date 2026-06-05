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

func ToUserResponse(u *models.User, accountContexts ...*dto.AccountContext) *dto.UserResponse {
	resp := &dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		Status:    u.Status,
		Capabilities: dto.AccountCapabilitiesResponse{
			Buyer: true,
		},
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	if len(accountContexts) == 0 || accountContexts[0] == nil {
		return resp
	}

	accountCtx := accountContexts[0]
	if accountCtx.BuyerProfileID != "" {
		resp.BuyerProfile = &dto.AccountProfileRefResponse{ID: accountCtx.BuyerProfileID}
	}
	if accountCtx.SupplierProfileID != "" {
		resp.Capabilities.Supplier = true
		resp.SupplierProfile = &dto.AccountProfileRefResponse{
			ID:     accountCtx.SupplierProfileID,
			Status: accountCtx.SupplierStatus,
		}
	}

	return resp
}
