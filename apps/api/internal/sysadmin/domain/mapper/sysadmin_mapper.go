package mapper

import (
	"github.com/gilabs/indosupplier/api/internal/sysadmin/data/models"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/domain/dto"
)

func permissionsForSet(permissionSet string) []string {
	switch permissionSet {
	case "super_admin":
		return []string{"system:*"}
	case "content_admin":
		return []string{"content:read", "content:write", "verification:read"}
	case "ads_admin":
		return []string{"ads:read", "ads:write", "auction:read"}
	case "cs_admin":
		return []string{"support:read", "support:write", "buyer:read", "supplier:read"}
	case "finance_admin":
		return []string{"finance:read", "finance:write", "subscription:read", "subscription:write"}
	case "moderator":
		return []string{"moderation:read", "moderation:write", "review:read", "review:write"}
	default:
		return []string{"system:read"}
	}
}

func ToSysadminResponse(sa *models.SystemAdmin) dto.SysadminResponse {
	return dto.SysadminResponse{
		ID:            sa.ID,
		Email:         sa.Email,
		Name:          sa.Name,
		PermissionSet: sa.PermissionSet,
		Permissions:   permissionsForSet(sa.PermissionSet),
		Status:        sa.Status,
	}
}
