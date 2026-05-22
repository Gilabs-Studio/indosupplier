package handler

import (
	"strings"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/permission/domain/usecase"
	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permissionUC usecase.PermissionUsecase
}

func NewPermissionHandler(permissionUC usecase.PermissionUsecase) *PermissionHandler {
	return &PermissionHandler{
		permissionUC: permissionUC,
	}
}

// List handles list permissions request
func (h *PermissionHandler) List(c *gin.Context) {
	permissions, err := h.permissionUC.List(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, permissions, nil)
}

// GetByID handles get permission by ID request
func (h *PermissionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	permission, err := h.permissionUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrPermissionNotFound {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
				"resource":      "permission",
				"permission_id": id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, permission, nil)
}

// GetUserPermissions handles get user permissions request
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("id")
	currentUserID, _ := c.Get("user_id")
	currentUserIDStr, _ := currentUserID.(string)

	if currentUserIDStr != userID && !hasPermissionCode(c, "user.read") {
		errors.ForbiddenResponse(c, "Missing permission: user.read", nil)
		return
	}

	permissions, err := h.permissionUC.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		if err == usecase.ErrUserNotFound {
			errors.ErrorResponse(c, "USER_NOT_FOUND", map[string]interface{}{
				"user_id": userID,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, permissions, nil)
}

func hasPermissionCode(c *gin.Context, permissionCode string) bool {
	perms, exists := c.Get("user_permissions")
	if !exists {
		return false
	}

	permMap, ok := perms.(map[string]bool)
	if !ok {
		return false
	}

	if permMap[permissionCode] {
		return true
	}

	for code := range permMap {
		if strings.EqualFold(code, permissionCode) {
			return true
		}
	}

	return false
}

// GetMenuCategories handles get menu categories request for dynamic grouping
func (h *PermissionHandler) GetMenuCategories(c *gin.Context) {
	categories, err := h.permissionUC.GetMenuCategories(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, categories, nil)
}
