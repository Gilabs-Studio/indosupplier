package mapper

import (
	permissionDto "github.com/gilabs/gims/api/internal/permission/domain/dto"
	permissionMapper "github.com/gilabs/gims/api/internal/permission/domain/mapper"
	"github.com/gilabs/gims/api/internal/role/data/models"
	"github.com/gilabs/gims/api/internal/role/domain/dto"
)

// ToRoleResponse converts Role to RoleResponse
func ToRoleResponse(r *models.Role) *dto.RoleResponse {
	resp := &dto.RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		Status:      r.Status,
		IsProtected: r.IsProtected,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	// If RolePermissions (explicit junction with scope) is loaded, use it
	if len(r.RolePermissions) > 0 {
		resp.Permissions = make([]permissionDto.PermissionResponse, 0, len(r.RolePermissions))
		for _, rp := range r.RolePermissions {
			if rp.Permission != nil {
				pResp := permissionMapper.ToPermissionWithScopeResponse(rp.Permission, rp.Scope)
				resp.Permissions = append(resp.Permissions, *pResp)
			}
		}
	} else if len(r.Permissions) > 0 {
		// Fallback: use many2many Permissions (no scope info)
		resp.Permissions = make([]permissionDto.PermissionResponse, len(r.Permissions))
		for i, p := range r.Permissions {
			resp.Permissions[i] = *permissionMapper.ToPermissionResponse(&p)
		}
	}

	return resp
}
