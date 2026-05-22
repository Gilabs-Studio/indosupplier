package handler

import (
	goErrors "errors"
	"log"
	"strconv"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/role/data/repositories"
	"github.com/gilabs/gims/api/internal/role/domain/dto"
	"github.com/gilabs/gims/api/internal/role/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RoleHandler struct {
	roleUC usecase.RoleUsecase
}

func NewRoleHandler(roleUC usecase.RoleUsecase) *RoleHandler {
	return &RoleHandler{
		roleUC: roleUC,
	}
}

// List handles list roles request
func (h *RoleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	roles, total, err := h.roleUC.List(c.Request.Context(), page, limit, search)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	pagination := response.NewPaginationMeta(page, limit, int(total))
	meta := &response.Meta{
		Pagination: pagination,
	}

	response.SuccessResponse(c, roles, meta)
}

// GetByID handles get role by ID request
func (h *RoleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	role, err := h.roleUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrRoleNotFound {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
				"resource": "role",
				"role_id":  id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, role, nil)
}

// Create handles create role request
func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	createdRole, err := h.roleUC.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrRoleAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "role",
				"field":    "code",
				"value":    req.Code,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			meta.CreatedBy = id
		}
	}

	response.SuccessResponseCreated(c, createdRole, meta)
}

// Update handles update role request
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	updatedRole, err := h.roleUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrRoleNotFound {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
				"resource": "role",
				"role_id":  id,
			}, nil)
			return
		}
		if err == usecase.ErrRoleAlreadyExists {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
				"resource": "role",
				"field":    "code",
			}, nil)
			return
		}
		if err == usecase.ErrRoleProtected {
			errors.ErrorResponse(c, "ROLE_PROTECTED", map[string]interface{}{
				"resource": "role",
				"role_id":  id,
				"message":  "This role is protected and cannot be modified",
			}, nil)
			return
		}
		if err == usecase.ErrLastAdminCannotDisable {
			errors.ErrorResponse(c, "LAST_ADMIN_CANNOT_DISABLE", map[string]interface{}{
				"resource": "role",
				"role_id":  id,
				"message":  "Cannot disable the last admin role",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{}
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			meta.UpdatedBy = id
		}
	}

	response.SuccessResponse(c, updatedRole, meta)
}

// Delete handles delete role request
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.roleUC.Delete(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrRoleNotFound {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
				"resource": "role",
				"role_id":  id,
			}, nil)
			return
		}
		if err == usecase.ErrRoleProtected {
			errors.ErrorResponse(c, "ROLE_PROTECTED", map[string]interface{}{
				"resource": "role",
				"role_id":  id,
				"message":  "This role is protected and cannot be deleted",
			}, nil)
			return
		}
		if err == usecase.ErrRoleInUse {
			errors.ErrorResponse(c, "ROLE_IN_USE", map[string]interface{}{
				"resource": "role",
				"role_id":  id,
				"message":  "This role is in use by users and cannot be deleted",
			}, nil)
			return
		}
		if err == usecase.ErrLastAdminCannotDelete {
			errors.ErrorResponse(c, "LAST_ADMIN_CANNOT_DELETE", map[string]interface{}{
				"resource": "role",
				"role_id":  id,
				"message":  "Cannot delete the last admin role",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	// Get user ID for meta
	meta := &response.Meta{}
	if userIDVal, exists := c.Get("user_id"); exists {
		if id, ok := userIDVal.(string); ok {
			meta.DeletedBy = id
		}
	}

	response.SuccessResponseDeleted(c, "role", id, meta)
}

// AssignPermissions handles assign permissions to role request.
// Supports legacy format (permission_ids), scope-aware format (assignments), and differential mode (diff).
// Query parameter ?tolerant=true enables tolerant mode: drops unauthorized permissions instead of rejecting entire request.
// Differential mode (mode=diff): Only validates added permissions against plan, leaving existing unchanged.
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id := c.Param("id")
	tenantID := c.GetString("tenant_id")
	tolerantMode := c.Query("tolerant") == "true"
	var req dto.AssignPermissionsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	// DIFF MODE: Only validate added permissions, leave unchanged untouched
	if req.Mode == "diff" && req.Diff != nil {
		log.Printf("[RoleHandler] assign_permissions mode=diff tenant_id=%s role_id=%s added=%v removed=%v added_with_scope=%d", tenantID, id, req.Diff.Added, req.Diff.Removed, len(req.Diff.AddedWithScope))
		err := h.roleUC.UpdatePermissionsDiff(c.Request.Context(), id, req.Diff)
		if err != nil {
			if err == usecase.ErrRoleNotFound {
				errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
					"resource": "role",
					"role_id":  id,
				}, nil)
				return
			}
			if err == usecase.ErrPermissionNotAllowedForPlan {
				errors.ErrorResponse(c, "MODULE_NOT_ENTITLED", map[string]interface{}{
					"reason": "One or more added permissions are not included in the tenant's active subscription plan",
				}, nil)
				return
			}
			if goErrors.Is(err, repositories.ErrInvalidPermissionIDs) {
				errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
					"reason": "One or more permission IDs are invalid or no longer exist. Please refresh and try again.",
				}, nil)
				return
			}
			errors.InternalServerErrorResponse(c, err.Error())
			return
		}

		// Fetch and return updated role
		role, err := h.roleUC.GetByID(c.Request.Context(), id)
		if err == nil {
			// return role as the data payload, with no meta
			response.SuccessResponse(c, role, nil)
			return
		}
		// return empty object as data when unable to fetch
		response.SuccessResponse(c, gin.H{}, nil)
		return
	}

	// Use scope-aware assignments if provided, otherwise fall back to legacy format
	if len(req.Assignments) > 0 {
		permissionIDs := make([]string, 0, len(req.Assignments))
		for _, assignment := range req.Assignments {
			permissionIDs = append(permissionIDs, assignment.PermissionID)
		}
		log.Printf("[RoleHandler] assign_permissions mode=scope tenant_id=%s role_id=%s permission_ids=%v", tenantID, id, permissionIDs)

		// In tolerant mode, filter out unauthorized permissions before assignment
		if tolerantMode {
			permIDs := make([]string, 0, len(req.Assignments))
			for _, a := range req.Assignments {
				permIDs = append(permIDs, a.PermissionID)
			}
			allowed, _, err := h.roleUC.FilterAllowedPermissions(c.Request.Context(), permIDs)
			if err == nil {
				// Filter assignments to only include allowed permissions
				filteredAssignments := make([]dto.PermissionAssignment, 0, len(allowed))
				allowedSet := make(map[string]struct{}, len(allowed))
				for _, id := range allowed {
					allowedSet[id] = struct{}{}
				}
				for _, a := range req.Assignments {
					if _, ok := allowedSet[a.PermissionID]; ok {
						filteredAssignments = append(filteredAssignments, a)
					}
				}
				req.Assignments = filteredAssignments
			}
		}

		err := h.roleUC.AssignPermissionsWithScope(c.Request.Context(), id, req.Assignments)
		if err != nil {
			if err == usecase.ErrRoleNotFound {
				errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
					"resource": "role",
					"role_id":  id,
				}, nil)
				return
			}
			if err == usecase.ErrPermissionNotAllowedForPlan {
				// In tolerant mode, this shouldn't happen after filtering, but handle it
				if tolerantMode {
					errors.ErrorResponse(c, "PARTIAL_ASSIGNMENT", map[string]interface{}{
						"reason": "Some permissions could not be assigned due to plan restrictions",
					}, nil)
				} else {
					errors.ErrorResponse(c, "MODULE_NOT_ENTITLED", map[string]interface{}{
						"reason": "One or more selected permissions are not included in the tenant's active subscription plan",
					}, nil)
				}
				return
			}
			if goErrors.Is(err, repositories.ErrInvalidPermissionIDs) {
				errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
					"reason": "One or more permission IDs are invalid or no longer exist. Please refresh and try again.",
				}, nil)
				return
			}
			errors.InternalServerErrorResponse(c, err.Error())
			return
		}
	} else {
		log.Printf("[RoleHandler] assign_permissions mode=legacy tenant_id=%s role_id=%s permission_ids=%v", tenantID, id, req.PermissionIDs)
		// Legacy format (permission_ids only)
		if tolerantMode {
			allowed, _, err := h.roleUC.FilterAllowedPermissions(c.Request.Context(), req.PermissionIDs)
			if err == nil {
				req.PermissionIDs = allowed
			}
		}

		err := h.roleUC.AssignPermissions(c.Request.Context(), id, req.PermissionIDs)
		if err != nil {
			if err == usecase.ErrRoleNotFound {
				errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
					"resource": "role",
					"role_id":  id,
				}, nil)
				return
			}
			if err == usecase.ErrPermissionNotAllowedForPlan {
				if tolerantMode {
					errors.ErrorResponse(c, "PARTIAL_ASSIGNMENT", map[string]interface{}{
						"reason": "Some permissions could not be assigned due to plan restrictions",
					}, nil)
				} else {
					errors.ErrorResponse(c, "MODULE_NOT_ENTITLED", map[string]interface{}{
						"reason": "One or more selected permissions are not included in the tenant's active subscription plan",
					}, nil)
				}
				return
			}
			if goErrors.Is(err, repositories.ErrInvalidPermissionIDs) {
				errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
					"reason": "One or more permission IDs are invalid or no longer exist. Please refresh and try again.",
				}, nil)
				return
			}
			errors.InternalServerErrorResponse(c, err.Error())
			return
		}
	}

	// Return updated role with scope information
	updatedRole, err := h.roleUC.GetByID(c.Request.Context(), id)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, updatedRole, nil)
}

// GetMenuAccess handles get role menu-access request.
func (h *RoleHandler) GetMenuAccess(c *gin.Context) {
	id := c.Param("id")

	menuAccess, err := h.roleUC.GetMenuAccess(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrRoleNotFound {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
				"resource": "role",
				"role_id":  id,
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{"assignments": menuAccess}, nil)
}

// UpdateMenuAccess handles replace role menu-access request.
func (h *RoleHandler) UpdateMenuAccess(c *gin.Context) {
	id := c.Param("id")
	var req dto.RoleMenuAccessRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	if err := h.roleUC.UpdateMenuAccess(c.Request.Context(), id, req.Assignments); err != nil {
		if err == usecase.ErrRoleNotFound {
			errors.ErrorResponse(c, "NOT_FOUND", map[string]interface{}{
				"resource": "role",
				"role_id":  id,
			}, nil)
			return
		}
		if err == usecase.ErrPermissionNotAllowedForPlan {
			errors.ErrorResponse(c, "MODULE_NOT_ENTITLED", map[string]interface{}{
				"reason": "One or more selected menus are not included in the tenant's active subscription plan",
			}, nil)
			return
		}
		if goErrors.Is(err, repositories.ErrInvalidPermissionIDs) {
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
				"reason": "One or more menu IDs are invalid or no longer exist. Please refresh and try again.",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	menuAccess, err := h.roleUC.GetMenuAccess(c.Request.Context(), id)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, gin.H{"assignments": menuAccess}, nil)
}

// ValidateUserRole handles validate user role request
func (h *RoleHandler) ValidateUserRole(c *gin.Context) {
	userID := c.Param("user_id")
	roleID := c.Query("role_id")

	if roleID == "" {
		errors.ErrorResponse(c, "INVALID_REQUEST", map[string]interface{}{
			"message": "role_id is required",
		}, nil)
		return
	}

	isValid, err := h.roleUC.ValidateUserRole(c.Request.Context(), userID, roleID)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, map[string]interface{}{
		"user_id":  userID,
		"role_id":  roleID,
		"is_valid": isValid,
	}, nil)
}
