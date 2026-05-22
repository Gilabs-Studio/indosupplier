package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/user/data/models"
	"github.com/gilabs/gims/api/internal/user/domain/dto"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, req *dto.ListUsersRequest) ([]models.User, int64, error)
	// FindAvailable returns users not yet linked to any employee (excluding the given employeeID if non-empty).
	FindAvailable(ctx context.Context, search string, excludeEmployeeID string) ([]models.User, error)
	// Count returns the total number of users belonging to the current tenant.
	Count(ctx context.Context) (int64, error)
	Create(ctx context.Context, u *models.User) error
	Update(ctx context.Context, u *models.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.getDB(ctx).
		Preload("Role").
		Preload("Role.RolePermissions").
		Preload("Role.RolePermissions.Permission").
		Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.getDB(ctx).
		Preload("Role").
		Preload("Role.RolePermissions").
		Preload("Role.RolePermissions.Permission").
		Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) List(ctx context.Context, req *dto.ListUsersRequest) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Scope to current tenant so users only see members of their own organisation
	query := r.getDB(ctx).Model(&models.User{}).Preload("Role").Preload("Role.Permissions")

	// Apply filters
	if req.Search != "" {
		search := "%" + req.Search + "%" // Prefix search for better performance with GIN index
		// Search across name, email, and role name
		query = query.Where(
			"users.name ILIKE ? OR users.email ILIKE ? OR EXISTS (SELECT 1 FROM roles WHERE roles.id = users.role_id AND roles.name ILIKE ?)",
			search, search, search,
		)
	}

	if req.Status != "" {
		query = query.Where("users.status = ?", req.Status)
	}

	if req.RoleID != "" {
		query = query.Where("users.role_id = ?", req.RoleID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	// Fetch data - Order by updated_at DESC so recently updated items appear first
	err := query.Order("users.updated_at DESC").Offset(offset).Limit(perPage).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Create(ctx context.Context, u *models.User) error {
	err := r.getDB(ctx).Create(u).Error
	if err != nil {
		if isDuplicateEmailError(err) {
			return errors.New("user already exists")
		}
		return err
	}
	return nil
}

// isDuplicateEmailError detects PostgreSQL unique-constraint violations on the users email column.
// Covers both the legacy constraint name and the index name used by newer migrations.
func isDuplicateEmailError(err error) bool {
	msg := err.Error()
	if !strings.Contains(msg, "duplicate key value violates unique constraint") {
		return false
	}
	return strings.Contains(msg, "idx_users_email") ||
		strings.Contains(msg, "users_email_key") ||
		strings.Contains(msg, "23505")
}

func (r *userRepository) Update(ctx context.Context, u *models.User) error {
	err := r.getDB(ctx).Save(u).Error
	if err != nil {
		if isDuplicateEmailError(err) {
			return errors.New("user already exists")
		}
		return err
	}
	return nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	err := r.getDB(ctx).Model(&models.User{}).Count(&total).Error
	return total, err
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.User{}).Error
}

// FindAvailable returns active users not yet linked to any employee.
// If excludeEmployeeID is non-empty, the user already linked to that employee is still included.
func (r *userRepository) FindAvailable(ctx context.Context, search string, excludeEmployeeID string) ([]models.User, error) {
	var users []models.User

	query := r.getDB(ctx).Model(&models.User{}).
		Preload("Role").
		Where("users.status = ?", "active")

	// Exclude users that are already linked to an employee (except the one being edited)
	subQuery := "users.id NOT IN (SELECT user_id FROM employees WHERE user_id IS NOT NULL AND deleted_at IS NULL"
	if excludeEmployeeID != "" {
		subQuery += " AND id != ?"
		query = query.Where(subQuery+")", excludeEmployeeID)
	} else {
		query = query.Where(subQuery + ")")
	}

	if search != "" {
		prefix := search + "%"
		query = query.Where("users.name ILIKE ? OR users.email ILIKE ?", prefix, prefix)
	}

	err := query.Order("users.name ASC").Limit(50).Find(&users).Error
	return users, err
}
