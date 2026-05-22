package usecase

import (
	"context"
	"errors"
	"net/url"

	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/events"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	infraEvents "github.com/gilabs/gims/api/internal/core/infrastructure/events"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/utils"
	roleRepositories "github.com/gilabs/gims/api/internal/role/data/repositories"
	tenantModels "github.com/gilabs/gims/api/internal/tenant/data/models"
	tenantRepositories "github.com/gilabs/gims/api/internal/tenant/data/repositories"
	"github.com/gilabs/gims/api/internal/user/data/models"
	"github.com/gilabs/gims/api/internal/user/data/repositories"
	"github.com/gilabs/gims/api/internal/user/domain/dto"
	"github.com/gilabs/gims/api/internal/user/domain/mapper"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound                = errors.New("user not found")
	ErrUserAlreadyExists           = errors.New("user already exists")
	ErrRoleNotFound                = errors.New("role not found")
	ErrUserLimitReached            = errors.New("user limit reached")
	ErrOwnerMutationForbidden = errors.New("owner user cannot be modified")
	ErrDeleteAccountForbidden = errors.New("only tenant owner can request account deletion")
)

// UserLimitResponse carries the current and maximum allowed user count for a tenant.
type UserLimitResponse struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

type subscriptionFinder interface {
	FindActiveByTenantID(ctx context.Context, tenantID string) (*tenantModels.TenantSubscription, error)
}

type UserUsecase interface {
	List(ctx context.Context, req *dto.ListUsersRequest) ([]dto.UserResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.UserResponse, error)
	// GetAvailable returns active users not yet linked to an employee.
	GetAvailable(ctx context.Context, search string, excludeEmployeeID string) ([]dto.AvailableUserResponse, error)
	// GetLimit returns the current user count and active subscription seat limit for the tenant.
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
	userRepo         repositories.UserRepository
	roleRepo         roleRepositories.RoleRepository
	tenantRepo       tenantRepositories.TenantRepository
	subscriptionRepo subscriptionFinder
	auditService     audit.AuditService
	eventPublisher   infraEvents.EventPublisher
	redis            *redis.Client
}

func userCacheScope(ctx context.Context) string {
	if middleware.IsSystemAdmin(ctx) {
		return "system_admin"
	}
	if tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx)); tenantID != "" {
		return tenantID
	}
	return "public"
}

func userListCacheKey(ctx context.Context, req *dto.ListUsersRequest) string {
	scope := userCacheScope(ctx)
	return fmt.Sprintf("users:list:scope:%s:page:%d:perPage:%d:search:%s:status:%s:roleID:%s",
		scope, req.Page, req.PerPage, req.Search, req.Status, req.RoleID)
}

func userByIDCacheKey(ctx context.Context, id string) string {
	scope := userCacheScope(ctx)
	return fmt.Sprintf("users:id:scope:%s:id:%s", scope, id)
}

func NewUserUsecase(
	userRepo repositories.UserRepository,
	roleRepo roleRepositories.RoleRepository,
	tenantRepo tenantRepositories.TenantRepository,
	subscriptionRepo subscriptionFinder,
	auditService audit.AuditService,
	eventPublisher infraEvents.EventPublisher,
	redis *redis.Client,
) UserUsecase {
	return &userUsecase{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		tenantRepo:       tenantRepo,
		subscriptionRepo: subscriptionRepo,
		auditService:     auditService,
		eventPublisher:   eventPublisher,
		redis:            redis,
	}
}

func (u *userUsecase) List(ctx context.Context, req *dto.ListUsersRequest) ([]dto.UserResponse, *utils.PaginationResult, error) {
	// Generate tenant-scoped cache key to prevent cross-tenant data leakage.
	cacheKey := userListCacheKey(ctx, req)

	// Try to get from cache
	val, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedResult struct {
			Users      []dto.UserResponse      `json:"users"`
			Pagination *utils.PaginationResult `json:"pagination"`
		}
		if err := json.Unmarshal([]byte(val), &cachedResult); err == nil {
			return cachedResult.Users, cachedResult.Pagination, nil
		}
	}

	// Fetch from DB
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

	// Cache result (TTL 5 minutes)
	cacheData := struct {
		Users      []dto.UserResponse      `json:"users"`
		Pagination *utils.PaginationResult `json:"pagination"`
	}{
		Users:      responses,
		Pagination: pagination,
	}

	if data, err := json.Marshal(cacheData); err == nil {
		u.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return responses, pagination, nil
}

func (u *userUsecase) GetByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	cacheKey := userByIDCacheKey(ctx, id)

	// Try to get from cache
	val, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedUser dto.UserResponse
		if err := json.Unmarshal([]byte(val), &cachedUser); err == nil {
			return &cachedUser, nil
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

	// Cache result (TTL 15 minutes)
	if data, err := json.Marshal(resp); err == nil {
		u.redis.Set(ctx, cacheKey, data, 15*time.Minute)
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

// GetLimit resolves the active tenant's seat limit from the active subscription and the current user count.
// System admins bypass tenant lookup and receive max=0 (unlimited).
func (u *userUsecase) GetLimit(ctx context.Context) (*UserLimitResponse, error) {
	current, err := u.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	if middleware.IsSystemAdmin(ctx) {
		return &UserLimitResponse{Current: int(current), Max: 0}, nil
	}

	tenantID := middleware.TenantFromContext(ctx)
	max := 0
	if u.subscriptionRepo != nil {
		if sub, err := u.subscriptionRepo.FindActiveByTenantID(ctx, tenantID); err == nil && sub != nil {
			max = sub.SeatLimit
			if sub.UserCount > max {
				max = sub.UserCount
			}
		}
	}

	tenant, err := u.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if tenant.MaxUsers > max {
		max = tenant.MaxUsers
	}

	if max <= 0 {
		max = 1
	}

	return &UserLimitResponse{Current: int(current), Max: max}, nil
}

func (u *userUsecase) Create(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Enforce per-tenant user limit (skip for system admins).
	// Fail-safe: if the limit cannot be determined, deny the create to prevent limit bypass.
	if !middleware.IsSystemAdmin(ctx) {
		limit, err := u.GetLimit(ctx)
		if err != nil {
			return nil, err
		}
		if limit.Max > 0 && limit.Current >= limit.Max {
			return nil, ErrUserLimitReached
		}
	}

	// Check if role exists
	_, err := u.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	// Check if email already exists
	_, err = u.userRepo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Set default status
	status := req.Status
	if status == "" {
		status = "active"
	}

	// Generate avatar URL using dicebear lorelei
	avatarURL := "https://api.dicebear.com/7.x/lorelei/svg?seed=" + url.QueryEscape(req.Email)

	// Create user
	usr := &models.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		Name:      req.Name,
		AvatarURL: avatarURL,
		RoleID:    req.RoleID,
		Status:    status,
	}

	if err := u.userRepo.Create(ctx, usr); err != nil {
		return nil, err
	}

	// Invalidate list cache
	pattern := "users:list:*"
	iter := u.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		u.redis.Del(ctx, iter.Val())
	}

	// Audit Log
	u.auditService.Log(ctx, "user.create", usr.ID, map[string]interface{}{
		"email": req.Email,
		"name":  req.Name,
		"role":  req.RoleID,
	})

	// Reload with role
	createdUser, err := u.userRepo.FindByID(ctx, usr.ID)
	if err != nil {
		return nil, err
	}

	// Publish event (async, fire-and-forget)
	u.eventPublisher.PublishAsync(ctx, events.NewUserCreatedEvent(ctx, events.UserCreatedPayload{
		UserID:    usr.ID,
		Email:     usr.Email,
		Name:      usr.Name,
		RoleID:    usr.RoleID,
		Status:    usr.Status,
		CreatedAt: usr.CreatedAt,
	}))

	return mapper.ToUserResponse(createdUser), nil
}

func (u *userUsecase) ensureOwnerMutationAllowed(ctx context.Context, targetUser *models.User) error {
	if middleware.IsSystemAdmin(ctx) {
		return nil
	}

	actorID, _ := ctx.Value("user_id").(string)
	if actorID == "" || targetUser == nil {
		return nil
	}

	if actorID == targetUser.ID {
		if targetUser.Role != nil && targetUser.Role.IsProtected {
			return ErrOwnerMutationForbidden
		}
		return nil
	}

	actorUser, err := u.userRepo.FindByID(ctx, actorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if actorUser.Role == nil || !actorUser.Role.IsProtected {
		return nil
	}

	if targetUser.Role != nil && targetUser.Role.IsProtected {
		return ErrOwnerMutationForbidden
	}

	return nil
}

func (u *userUsecase) Update(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// Find user
	usr, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := u.ensureOwnerMutationAllowed(ctx, usr); err != nil {
		return nil, err
	}

	// Update fields
	if req.Email != "" {
		// Check if email already exists (excluding current user)
		existingUser, err := u.userRepo.FindByEmail(ctx, req.Email)
		if err == nil && existingUser.ID != id {
			return nil, ErrUserAlreadyExists
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		usr.Email = req.Email
	}

	if req.Name != "" {
		usr.Name = req.Name
	}

	if req.RoleID != "" {
		// Check if role exists
		_, err := u.roleRepo.FindByID(ctx, req.RoleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrRoleNotFound
			}
			return nil, err
		}
		usr.RoleID = req.RoleID
		// Clear preloaded role to prevent GORM from reverting role_id during Save()
		// This happens because FindByID preloads Role, and GORM may restore the association
		usr.Role = nil
	}

	if req.Status != "" {
		usr.Status = req.Status
	}

	if err := u.userRepo.Update(ctx, usr); err != nil {
		return nil, err
	}

	// Invalidate cache (new tenant-scoped key + legacy key for safety)
	u.redis.Del(ctx, userByIDCacheKey(ctx, id), fmt.Sprintf("users:id:%s", id))

	// Invalidate list cache
	pattern := "users:list:*"
	iter := u.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		u.redis.Del(ctx, iter.Val())
	}

	// Audit Log
	u.auditService.Log(ctx, "user.update", id, map[string]interface{}{
		"updates": req,
	})

	// Reload with role
	updatedUser, err := u.userRepo.FindByID(ctx, usr.ID)
	if err != nil {
		return nil, err
	}

	// Publish event (async, fire-and-forget)
	u.eventPublisher.PublishAsync(ctx, events.NewUserUpdatedEvent(ctx, events.UserUpdatedPayload{
		UserID:    usr.ID,
		Email:     usr.Email,
		Name:      usr.Name,
		RoleID:    usr.RoleID,
		Status:    usr.Status,
		UpdatedAt: usr.UpdatedAt,
	}))

	return mapper.ToUserResponse(updatedUser), nil
}

func (u *userUsecase) Delete(ctx context.Context, id string) error {
	// Check if user exists
	usr, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if err := u.ensureOwnerMutationAllowed(ctx, usr); err != nil {
		return err
	}

	if err := u.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache (new tenant-scoped key + legacy key for safety)
	u.redis.Del(ctx, userByIDCacheKey(ctx, id), fmt.Sprintf("users:id:%s", id))

	// Invalidate list cache
	pattern := "users:list:*"
	iter := u.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		u.redis.Del(ctx, iter.Val())
	}

	// Audit Log
	u.auditService.Log(ctx, "user.delete", id, nil)

	// Publish event (async, fire-and-forget)
	u.eventPublisher.PublishAsync(ctx, events.NewUserDeletedEvent(ctx, events.UserDeletedPayload{
		UserID:    id,
		DeletedAt: apptime.Now(),
	}))

	return nil
}

func (u *userUsecase) RequestAccountDeletion(ctx context.Context, id string) (*dto.TenantDeletionScheduleResponse, error) {
	usr, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if usr.Role == nil || !usr.Role.IsProtected {
		return nil, ErrDeleteAccountForbidden
	}

	tenant, err := u.tenantRepo.FindByID(ctx, usr.TenantID)
	if err != nil {
		return nil, err
	}

	if tenant.OwnerUserID == nil || strings.TrimSpace(*tenant.OwnerUserID) != strings.TrimSpace(usr.ID) {
		return nil, ErrDeleteAccountForbidden
	}

	now := apptime.Now()
	scheduledAt := now.AddDate(0, 0, 30)

	if strings.EqualFold(strings.TrimSpace(tenant.Status), "pending_deletion") && tenant.DeletionScheduledAt != nil {
		requestedAt := now
		if tenant.DeletionRequestedAt != nil {
			requestedAt = *tenant.DeletionRequestedAt
		}
		return &dto.TenantDeletionScheduleResponse{
			TenantID:            tenant.ID,
			DeletionRequestedAt: requestedAt.Format(time.RFC3339),
			DeletionScheduledAt: tenant.DeletionScheduledAt.Format(time.RFC3339),
			GracePeriodDays:     30,
		}, nil
	}

	previousStatus := strings.TrimSpace(tenant.Status)
	if previousStatus == "" {
		previousStatus = "active"
	}

	tenant.Status = "pending_deletion"
	tenant.DeletionRequestedAt = &now
	tenant.DeletionScheduledAt = &scheduledAt
	tenant.DeletionRequestedBy = &usr.ID
	tenant.DeletionPreviousStatus = &previousStatus
	tenant.DeletionRecoveredAt = nil

	if err := u.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	// Immediately disable tenant users while retaining data during the grace period.
	_ = database.DB.WithContext(ctx).
		Table("users").
		Where("tenant_id = ? AND deleted_at IS NULL", tenant.ID).
		Updates(map[string]interface{}{"status": "inactive", "updated_at": now}).Error

	// Revoke active refresh tokens so existing sessions are not silently extended.
	_ = database.DB.WithContext(ctx).
		Table("refresh_tokens").
		Where("tenant_id = ? AND revoked = false", tenant.ID).
		Updates(map[string]interface{}{"revoked": true, "updated_at": now}).Error

	u.auditService.Log(ctx, "tenant.deletion.requested", tenant.ID, map[string]interface{}{
		"requested_by": usr.ID,
		"scheduled_at": scheduledAt.Format(time.RFC3339),
	})

	return &dto.TenantDeletionScheduleResponse{
		TenantID:            tenant.ID,
		DeletionRequestedAt: now.Format(time.RFC3339),
		DeletionScheduledAt: scheduledAt.Format(time.RFC3339),
		GracePeriodDays:     30,
	}, nil
}

func (u *userUsecase) UpdateProfile(ctx context.Context, id string, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	// Find user
	usr, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Update fields
	if req.Email != "" {
		// Check if email already exists (excluding current user)
		existingUser, err := u.userRepo.FindByEmail(ctx, req.Email)
		if err == nil && existingUser.ID != id {
			return nil, ErrUserAlreadyExists
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		// If email changed, regenerate avatar
		if usr.Email != req.Email {
			usr.Email = req.Email
			// Regenerate avatar URL using dicebear lorelei
			usr.AvatarURL = "https://api.dicebear.com/7.x/lorelei/svg?seed=" + url.QueryEscape(req.Email)
		}
	}

	if req.Name != "" {
		usr.Name = req.Name
	}

	if err := u.userRepo.Update(ctx, usr); err != nil {
		return nil, err
	}

	// Invalidate cache (new tenant-scoped key + legacy key for safety)
	u.redis.Del(ctx, userByIDCacheKey(ctx, id), fmt.Sprintf("users:id:%s", id))

	// Invalidate list cache
	pattern := "users:list:*"
	iter := u.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		u.redis.Del(ctx, iter.Val())
	}

	// Audit Log
	u.auditService.Log(ctx, "user.profile_update", id, map[string]interface{}{
		"updates": req,
	})

	// Reload with role
	updatedUser, err := u.userRepo.FindByID(ctx, usr.ID)
	if err != nil {
		return nil, err
	}

	// Publish event (async, fire-and-forget)
	u.eventPublisher.PublishAsync(ctx, events.NewUserUpdatedEvent(ctx, events.UserUpdatedPayload{
		UserID:    usr.ID,
		Email:     usr.Email,
		Name:      usr.Name,
		RoleID:    usr.RoleID,
		Status:    usr.Status,
		UpdatedAt: usr.UpdatedAt,
	}))

	return mapper.ToUserResponse(updatedUser), nil
}

func (u *userUsecase) ChangePassword(ctx context.Context, id string, req *dto.ChangePasswordRequest) error {
	// Find user
	usr, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	usr.Password = string(hashedPassword)
	if err := u.userRepo.Update(ctx, usr); err != nil {
		return err
	}

	// Audit Log
	u.auditService.Log(ctx, "user.change_password", id, nil)

	return nil
}

func (u *userUsecase) UpdateAvatar(ctx context.Context, id string, avatarURL string) error {
	// Find user
	usr, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Update avatar URL
	usr.AvatarURL = avatarURL
	if err := u.userRepo.Update(ctx, usr); err != nil {
		return err
	}

	// Invalidate cache (new tenant-scoped key + legacy key for safety)
	u.redis.Del(ctx, userByIDCacheKey(ctx, id), fmt.Sprintf("users:id:%s", id))

	// Invalidate list cache
	pattern := "users:list:*"
	iter := u.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		u.redis.Del(ctx, iter.Val())
	}

	// Audit Log
	u.auditService.Log(ctx, "user.update_avatar", id, map[string]interface{}{
		"avatar_url": avatarURL,
	})

	return nil
}
