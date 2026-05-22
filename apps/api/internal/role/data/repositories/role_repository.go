package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/role/data/models"
	"gorm.io/gorm"
)

// ErrInvalidPermissionIDs indicates one or more permission IDs do not exist in the database
var ErrInvalidPermissionIDs = errors.New("one or more permission IDs are invalid or no longer exist")

type RoleRepository interface {
	FindByID(ctx context.Context, id string) (*models.Role, error)
	FindByCode(ctx context.Context, code string) (*models.Role, error)
	List(ctx context.Context, page, limit int, search string) ([]models.Role, int64, error)
	Create(ctx context.Context, ro *models.Role) error
	Update(ctx context.Context, ro *models.Role) error
	Delete(ctx context.Context, id string) error
	AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error
	AssignPermissionsWithScope(ctx context.Context, roleID string, assignments []models.RolePermission) error
	GetPermissions(ctx context.Context, roleID string) ([]string, error)
	CountUsersByRoleID(ctx context.Context, roleID string) (int64, error)
	CountAdmins(ctx context.Context) (int64, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *roleRepository) FindByID(ctx context.Context, id string) (*models.Role, error) {
	var ro models.Role
	err := r.getDB(ctx).
		Preload("Permissions").
		Preload("RolePermissions").
		Preload("RolePermissions.Permission").
		Preload("RolePermissions.Permission.Menu").
		Preload("RolePermissions.Permission.Menu.Parent").
		Where("id = ?", id).First(&ro).Error
	if err != nil {
		return nil, err
	}
	return &ro, nil
}

func (r *roleRepository) FindByCode(ctx context.Context, code string) (*models.Role, error) {
	var ro models.Role
	err := r.getDB(ctx).
		Preload("Permissions").
		Preload("RolePermissions").
		Preload("RolePermissions.Permission").
		Where("code = ?", code).First(&ro).Error
	if err != nil {
		return nil, err
	}
	return &ro, nil
}

func (r *roleRepository) List(ctx context.Context, page, limit int, search string) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	query := r.getDB(ctx).Model(&models.Role{})

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Permissions").
		Preload("RolePermissions").
		Preload("RolePermissions.Permission").
		Preload("RolePermissions.Permission.Menu").
		Preload("RolePermissions.Permission.Menu.Parent").
		Order("status DESC, updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&roles).Error

	if err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

func (r *roleRepository) Create(ctx context.Context, ro *models.Role) error {
	return r.getDB(ctx).Create(ro).Error
}

func (r *roleRepository) Update(ctx context.Context, ro *models.Role) error {
	return r.getDB(ctx).Save(ro).Error
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.Role{}).Error
}

func (r *roleRepository) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Validate all permission IDs exist before making changes
		if len(permissionIDs) > 0 {
			var validCount int64
			// permissions is a global table; avoid tenant-scoped DB here.
			if err := r.db.WithContext(ctx).Table("permissions").Where("id IN ? AND deleted_at IS NULL", permissionIDs).Count(&validCount).Error; err != nil {
				return err
			}
			if int(validCount) != len(permissionIDs) {
				return ErrInvalidPermissionIDs
			}
		}

		// Clear existing permissions
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}

		// Assign new permissions with default ALL scope
		for _, permID := range permissionIDs {
			if err := tx.Exec(
				"INSERT INTO role_permissions (role_id, permission_id, scope) VALUES (?, ?, 'ALL')",
				roleID, permID,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *roleRepository) AssignPermissionsWithScope(ctx context.Context, roleID string, assignments []models.RolePermission) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Validate all permission IDs exist before making changes
		if len(assignments) > 0 {
			permIDs := make([]string, 0, len(assignments))
			for _, a := range assignments {
				permIDs = append(permIDs, a.PermissionID)
			}

			var existingIDs []string
			// permissions is a global table; avoid tenant-scoped DB here.
			if err := r.db.WithContext(ctx).Table("permissions").Where("id IN ? AND deleted_at IS NULL", permIDs).Pluck("id", &existingIDs).Error; err != nil {
				return err
			}

			existingSet := make(map[string]bool, len(existingIDs))
			for _, id := range existingIDs {
				existingSet[id] = true
			}

			missingIDs := make([]string, 0)
			for _, id := range permIDs {
				if !existingSet[id] {
					missingIDs = append(missingIDs, id)
				}
			}
			if len(missingIDs) > 0 {
				tenantID := ""
				if v, ok := ctx.Value("tenant_id").(string); ok {
					tenantID = v
				}
				sort.Strings(missingIDs)
				log.Printf("[RoleRepository] AssignPermissionsWithScope invalid_permission_ids tenant_id=%s role_id=%s permission_ids=%v", tenantID, roleID, missingIDs)
				return fmt.Errorf("%w: %v", ErrInvalidPermissionIDs, missingIDs)
			}
		}

		// Clear existing permissions
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}

		// Assign new permissions with their respective scopes
		for _, a := range assignments {
			scope := a.Scope
			if scope == "" {
				scope = "ALL"
			}
			if err := tx.Exec(
				"INSERT INTO role_permissions (role_id, permission_id, scope) VALUES (?, ?, ?)",
				roleID, a.PermissionID, scope,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *roleRepository) GetPermissions(ctx context.Context, roleID string) ([]string, error) {
	var permissionIDs []string
	err := r.getDB(ctx).Table("role_permissions").
		Where("role_id = ?", roleID).
		Pluck("permission_id", &permissionIDs).Error
	if err != nil {
		return nil, err
	}
	return permissionIDs, nil
}

func (r *roleRepository) CountUsersByRoleID(ctx context.Context, roleID string) (int64, error) {
	var count int64
	err := r.getDB(ctx).Table("users").Where("role_id = ? AND deleted_at IS NULL", roleID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *roleRepository) CountAdmins(ctx context.Context) (int64, error) {
	var count int64
	// Count users with admin role (code = "admin")
	err := r.getDB(ctx).Table("users").
		Joins("JOIN roles ON users.role_id = roles.id").
		Where("roles.code = ? AND users.deleted_at IS NULL AND roles.deleted_at IS NULL", "admin").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
