package mapper

import (
	"github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
)

// ToXenditConfigResponse maps a XenditConfig to its DTO response (secret key and webhook token are excluded)
func ToXenditConfigResponse(c *models.XenditConfig) *dto.XenditConfigResponse {
	return &dto.XenditConfigResponse{
		ID:               c.ID,
		CompanyID:        c.CompanyID,
		XenditAccountID:  c.XenditAccountID,
		BusinessName:     c.BusinessName,
		Environment:      string(c.Environment),
		ConnectionStatus: string(c.ConnectionStatus),
		IsActive:         c.IsActive,
		UpdatedAt:        c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
