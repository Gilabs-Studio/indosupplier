package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/events"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/audit"
	infraEvents "github.com/gilabs/indosupplier/api/internal/core/infrastructure/events"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	"github.com/gilabs/indosupplier/api/internal/user/data/models"
	"github.com/gilabs/indosupplier/api/internal/user/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/user/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/user/domain/mapper"
	"github.com/redis/go-redis/v9"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrRoleNotFound           = errors.New("role not found")
	ErrUserLimitReached       = errors.New("user limit reached")
	ErrOwnerMutationForbidden = errors.New("owner user cannot be modified")
	ErrDeleteAccountForbidden = errors.New("only tenant owner can request account deletion")
)

type UserLimitResponse struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

type UserUsecase interface {
	List(ctx context.Context, req *dto.ListUsersRequest) ([]dto.UserResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.UserResponse, error)
	GetAvailable(ctx context.Context, search string, excludeEmployeeID string) ([]dto.AvailableUserResponse, error)
	GetLimit(ctx context.Context) (*UserLimitResponse, error)
	Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdateProfile(ctx context.Context, id string, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, id string, req *dto.ChangePasswordRequest) error
	UpdateAvatar(ctx context.Context, id string, avatarURL string) error
	Delete(ctx context.Context, id string) error
	RequestAccountDeletion(ctx context.Context, id string) (*dto.TenantDeletionScheduleResponse, error)
}

type userUsecase struct {
	userRepo       repositories.UserRepository
	auditService   audit.AuditService
	eventPublisher infraEvents.EventPublisher
	redis          *redis.Client
}

func NewUserUsecase(
	userRepo repositories.UserRepository,
	auditService audit.AuditService,
	eventPublisher infraEvents.EventPublisher,
	redis *redis.Client,
) UserUsecase {
	return &userUsecase{userRepo: userRepo, auditService: auditService, eventPublisher: eventPublisher, redis: redis}
}

func userListCacheKey(req *dto.ListUsersRequest) string {
	return fmt.Sprintf("users:list:page:%d:per_page:%d:search:%s:status:%s:role_id:%s", req.Page, req.PerPage, req.Search, req.Status, req.RoleID)
}

func userByIDCacheKey(id string) string {
	return fmt.Sprintf("users:id:%s", id)
}

func (u *userUsecase) List(ctx context.Context, req *dto.ListUsersRequest) ([]dto.UserResponse, *utils.PaginationResult, error) {
	cacheKey := userListCacheKey(req)
	if u.redis != nil {
		if val, err := u.redis.Get(ctx, cacheKey).Result(); err == nil {
			var cached struct {
				Users      []dto.UserResponse      `json:"users"`
				Pagination *utils.PaginationResult `json:"pagination"`
			}
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return cached.Users, cached.Pagination, nil
			}
		}
	}

	users, total, err := u.userRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]dto.UserResponse, len(users))
	for i, usr := range users {
		responses[i] = *mapper.ToUserResponse(&usr)
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

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	if u.redis != nil {
		payload, _ := json.Marshal(struct {
			Users      []dto.UserResponse      `json:"users"`
			Pagination *utils.PaginationResult `json:"pagination"`
		}{Users: responses, Pagination: pagination})
		u.redis.Set(ctx, cacheKey, payload, 5*time.Minute)
	}

	return responses, pagination, nil
}

func (u *userUsecase) GetByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	cacheKey := userByIDCacheKey(id)
	if u.redis != nil {
		if val, err := u.redis.Get(ctx, cacheKey).Result(); err == nil {
			var cached dto.UserResponse
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return &cached, nil
			}
		}
	}

	usr, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	resp := mapper.ToUserResponse(usr)
	if u.redis != nil {
		payload, _ := json.Marshal(resp)
		u.redis.Set(ctx, cacheKey, payload, 15*time.Minute)
	}
	return resp, nil
}

func (u *userUsecase) GetAvailable(ctx context.Context, search string, excludeEmployeeID string) ([]dto.AvailableUserResponse, error) {
	users, err := u.userRepo.FindAvailable(ctx, search, excludeEmployeeID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.AvailableUserResponse, len(users))
	for i, usr := range users {
		responses[i] = mapper.ToAvailableUserResponse(&usr)
	}
	return responses, nil
}

func (u *userUsecase) GetLimit(ctx context.Context) (*UserLimitResponse, error) {
	current, err := u.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	return &UserLimitResponse{Current: int(current), Max: 0}, nil
}

func (u *userUsecase) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	if _, err := u.userRepo.FindByEmail(ctx, req.Email); err == nil {
		return nil, ErrUserAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	usr := &models.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		Name:      req.Name,
		AvatarURL: "https://api.dicebear.com/7.x/lorelei/svg?seed=" + url.QueryEscape(req.Email),
		RoleID:    req.RoleID,
		Status:    status,
	}

	if err := u.userRepo.Create(ctx, usr); err != nil {
		if err.Error() == "user already exists" {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	u.invalidateCaches(ctx, usr.ID)

	u.auditService.Log(ctx, "user.create", usr.ID, map[string]interface{}{"email": req.Email, "name": req.Name, "role": req.RoleID})
	u.eventPublisher.PublishAsync(ctx, events.NewUserCreatedEvent(ctx, events.UserCreatedPayload{
		UserID:    usr.ID,
		Email:     usr.Email,
		Name:      usr.Name,
		RoleID:    usr.RoleID,
		Status:    usr.Status,
		CreatedAt: usr.CreatedAt,
	}))

	created, err := u.userRepo.FindByID(ctx, usr.ID)
	if err != nil {
		return nil, err
	}

	return mapper.ToUserResponse(created), nil
}

func (u *userUsecase) Update(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	usr, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if req.Email != "" {
		if req.Email != usr.Email {
			if _, err := u.userRepo.FindByEmail(ctx, req.Email); err == nil {
				return nil, ErrUserAlreadyExists
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		}
		usr.Email = req.Email
	}
	if req.Name != "" {
		usr.Name = req.Name
	}
	if req.RoleID != "" {
		usr.RoleID = req.RoleID
	}
	if req.Status != "" {
		usr.Status = req.Status
	}

	if err := u.userRepo.Update(ctx, usr); err != nil {
		if err.Error() == "user already exists" {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	u.invalidateCaches(ctx, usr.ID)
	return mapper.ToUserResponse(usr), nil
}

func (u *userUsecase) UpdateProfile(ctx context.Context, id string, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	return u.Update(ctx, id, &dto.UpdateUserRequest{Email: req.Email, Name: req.Name})
}

func (u *userUsecase) ChangePassword(ctx context.Context, id string, req *dto.ChangePasswordRequest) error {
	usr, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("old password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	usr.Password = string(hashedPassword)
	if err := u.userRepo.Update(ctx, usr); err != nil {
		return err
	}

	u.invalidateCaches(ctx, usr.ID)
	return nil
}

func (u *userUsecase) UpdateAvatar(ctx context.Context, id string, avatarURL string) error {
	usr, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	usr.AvatarURL = avatarURL
	if err := u.userRepo.Update(ctx, usr); err != nil {
		return err
	}
	u.invalidateCaches(ctx, usr.ID)
	return nil
}

func (u *userUsecase) Delete(ctx context.Context, id string) error {
	if _, err := u.userRepo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if err := u.userRepo.Delete(ctx, id); err != nil {
		return err
	}
	u.invalidateCaches(ctx, id)
	return nil
}

func (u *userUsecase) RequestAccountDeletion(ctx context.Context, id string) (*dto.TenantDeletionScheduleResponse, error) {
	return nil, ErrDeleteAccountForbidden
}

func (u *userUsecase) invalidateCaches(ctx context.Context, userID string) {
	if u.redis == nil {
		return
	}

	u.redis.Del(ctx, userByIDCacheKey(userID))

	iter := u.redis.Scan(ctx, 0, "users:list:*", 0).Iterator()
	for iter.Next(ctx) {
		u.redis.Del(ctx, iter.Val())
	}
}
