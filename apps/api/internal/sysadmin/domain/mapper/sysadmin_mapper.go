package mapper

import (
	"github.com/gilabs/indosupplier/api/internal/sysadmin/data/models"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/domain/dto"
)

func ToSysadminResponse(sa *models.SystemAdmin) dto.SysadminResponse {
	return dto.SysadminResponse{
		ID:     sa.ID,
		Email:  sa.Email,
		Name:   sa.Name,
		Role:   sa.Role,
		Status: sa.Status,
	}
}
