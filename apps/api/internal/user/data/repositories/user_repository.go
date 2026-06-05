package repositories

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	"github.com/gilabs/indosupplier/api/internal/user/data/models"
	"github.com/gilabs/indosupplier/api/internal/user/domain/dto"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, req *dto.ListUsersRequest) ([]models.User, int64, error)
	FindAvailable(ctx context.Context, search string, excludeEmployeeID string) ([]models.User, error)
	FindAccountContext(ctx context.Context, userID string) (*dto.AccountContext, error)
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
	if err := r.getDB(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	if err := r.getDB(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) List(ctx context.Context, req *dto.ListUsersRequest) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.getDB(ctx).Model(&models.User{})

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("users.name ILIKE ? OR users.email ILIKE ?", search, search)
	}
	if req.Status != "" {
		query = query.Where("users.status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

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
	if err := query.Order("users.updated_at DESC").Offset(offset).Limit(perPage).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func isDuplicateEmailError(err error) bool {
	msg := err.Error()
	if !strings.Contains(msg, "duplicate key value violates unique constraint") {
		return false
	}
	return strings.Contains(msg, "idx_users_email") ||
		strings.Contains(msg, "users_email_key") ||
		strings.Contains(msg, "23505")
}

func (r *userRepository) Create(ctx context.Context, u *models.User) error {
	if err := r.getDB(ctx).Create(u).Error; err != nil {
		if isDuplicateEmailError(err) {
			return errors.New("user already exists")
		}
		return err
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, u *models.User) error {
	if err := r.getDB(ctx).Save(u).Error; err != nil {
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

func (r *userRepository) FindAvailable(ctx context.Context, search string, excludeEmployeeID string) ([]models.User, error) {
	var users []models.User

	query := r.getDB(ctx).Model(&models.User{}).
		Where("users.status = ?", "active")

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("users.name ILIKE ? OR users.email ILIKE ?", like, like)
	}

	err := query.Order("users.name ASC").Limit(50).Find(&users).Error
	return users, err
}

func (r *userRepository) FindAccountContext(ctx context.Context, userID string) (*dto.AccountContext, error) {
	accountCtx := &dto.AccountContext{}

	type buyerRow struct {
		ID string
	}
	var buyer buyerRow
	if err := r.getDB(ctx).
		Table("buyer_profiles").
		Select("id").
		Where("user_id = ?", userID).
		Limit(1).
		Scan(&buyer).Error; err != nil {
		return nil, err
	}
	accountCtx.BuyerProfileID = buyer.ID

	type supplierRow struct {
		ID     string
		Status string
	}
	var supplier supplierRow
	if err := r.getDB(ctx).
		Table("supplier_profiles").
		Select("id, status").
		Where("user_id = ?", userID).
		Limit(1).
		Scan(&supplier).Error; err != nil {
		return nil, err
	}
	accountCtx.SupplierProfileID = supplier.ID
	accountCtx.SupplierStatus = supplier.Status

	return accountCtx, nil
}
