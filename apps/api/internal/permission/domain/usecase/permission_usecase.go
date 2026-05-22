package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/permission/data/repositories"
	"github.com/gilabs/gims/api/internal/permission/domain/dto"
	"github.com/gilabs/gims/api/internal/permission/domain/mapper"
	userRepo "github.com/gilabs/gims/api/internal/user/data/repositories"
	"gorm.io/gorm"
)

var (
	ErrPermissionNotFound = errors.New("permission not found")
	ErrUserNotFound       = errors.New("user not found")
)

type PermissionUsecase interface {
	List(ctx context.Context) ([]dto.PermissionResponse, error)
	GetByID(ctx context.Context, id string) (*dto.PermissionResponse, error)
	GetUserPermissions(ctx context.Context, userID string) (*dto.GetUserPermissionsResponse, error)
	GetMenuCategories(ctx context.Context) ([]dto.MenuCategoryResponse, error)
}

type permissionUsecase struct {
	permissionRepo repositories.PermissionRepository
	userRepo       userRepo.UserRepository
}

func NewPermissionUsecase(permissionRepo repositories.PermissionRepository, userRepo userRepo.UserRepository) PermissionUsecase {
	return &permissionUsecase{
		permissionRepo: permissionRepo,
		userRepo:       userRepo,
	}
}

func (u *permissionUsecase) List(ctx context.Context) ([]dto.PermissionResponse, error) {
	permissions, err := u.permissionRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PermissionResponse, len(permissions))
	for i, p := range permissions {
		responses[i] = *mapper.ToPermissionResponse(&p)
	}

	return responses, nil
}

func (u *permissionUsecase) GetByID(ctx context.Context, id string) (*dto.PermissionResponse, error) {
	p, err := u.permissionRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, err
	}
	return mapper.ToPermissionResponse(p), nil
}

func (u *permissionUsecase) GetUserPermissions(ctx context.Context, userID string) (*dto.GetUserPermissionsResponse, error) {
	// Check if user exists
	_, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return u.permissionRepo.GetUserPermissions(ctx, userID)
}

func (u *permissionUsecase) GetMenuCategories(ctx context.Context) ([]dto.MenuCategoryResponse, error) {
	menus, err := u.permissionRepo.GetRootMenusWithChildren(ctx)
	if err != nil {
		return nil, err
	}

	categories := make([]dto.MenuCategoryResponse, len(menus))
	for i, menu := range menus {
		categories[i] = *mapper.ToMenuCategoryResponse(&menu)
	}

	return categories, nil
}
