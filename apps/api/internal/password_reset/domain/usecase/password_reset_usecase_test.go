package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	infraEvents "github.com/gilabs/gims/api/internal/core/infrastructure/events"
	notificationModels "github.com/gilabs/gims/api/internal/notification/data/models"
	notificationRepo "github.com/gilabs/gims/api/internal/notification/data/repositories"
	passwordResetModels "github.com/gilabs/gims/api/internal/password_reset/data/models"
	resetDto "github.com/gilabs/gims/api/internal/password_reset/domain/dto"
	roleModels "github.com/gilabs/gims/api/internal/role/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	userDto "github.com/gilabs/gims/api/internal/user/domain/dto"
)

type fakePasswordResetRepo struct {
	created []*passwordResetModels.PasswordResetRequest
}

func (f *fakePasswordResetRepo) Create(ctx context.Context, req *passwordResetModels.PasswordResetRequest) error {
	f.created = append(f.created, req)
	return nil
}

func (f *fakePasswordResetRepo) FindByToken(ctx context.Context, token string) (*passwordResetModels.PasswordResetRequest, error) {
	return nil, gorm.ErrRecordNotFound
}

func (f *fakePasswordResetRepo) FindByUserID(ctx context.Context, userID string) (*passwordResetModels.PasswordResetRequest, error) {
	return nil, gorm.ErrRecordNotFound
}

func (f *fakePasswordResetRepo) FindPendingByUserID(ctx context.Context, userID string) (*passwordResetModels.PasswordResetRequest, error) {
	return nil, gorm.ErrRecordNotFound
}

func (f *fakePasswordResetRepo) Update(ctx context.Context, req *passwordResetModels.PasswordResetRequest) error {
	return nil
}

func (f *fakePasswordResetRepo) DeleteByUserID(ctx context.Context, userID string) error {
	return nil
}

func (f *fakePasswordResetRepo) DeleteExpired(ctx context.Context) error {
	return nil
}

type fakeUserRepo struct {
	user    *userModels.User
	users   []userModels.User
	updated []*userModels.User
}

func (f *fakeUserRepo) FindByID(ctx context.Context, id string) (*userModels.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*userModels.User, error) {
	if f.user == nil || f.user.Email != email {
		return nil, gorm.ErrRecordNotFound
	}
	return f.user, nil
}

func (f *fakeUserRepo) List(ctx context.Context, req *userDto.ListUsersRequest) ([]userModels.User, int64, error) {
	if req != nil && req.RoleID != "" {
		filtered := make([]userModels.User, 0, len(f.users))
		for _, item := range f.users {
			if item.RoleID == req.RoleID {
				filtered = append(filtered, item)
			}
		}
		return filtered, int64(len(filtered)), nil
	}
	return f.users, int64(len(f.users)), nil
}

func (f *fakeUserRepo) FindAvailable(ctx context.Context, search string, excludeEmployeeID string) ([]userModels.User, error) {
	return nil, nil
}

func (f *fakeUserRepo) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

func (f *fakeUserRepo) Create(ctx context.Context, u *userModels.User) error {
	return nil
}

func (f *fakeUserRepo) Update(ctx context.Context, u *userModels.User) error {
	f.updated = append(f.updated, u)
	return nil
}

func (f *fakeUserRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type fakeRoleRepo struct {
	role *roleModels.Role
}

func (f *fakeRoleRepo) FindByID(ctx context.Context, id string) (*roleModels.Role, error) {
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRoleRepo) FindByCode(ctx context.Context, code string) (*roleModels.Role, error) {
	if f.role == nil || f.role.Code != code {
		return nil, gorm.ErrRecordNotFound
	}
	return f.role, nil
}

func (f *fakeRoleRepo) List(ctx context.Context, page, limit int, search string) ([]roleModels.Role, int64, error) {
	return nil, 0, nil
}

func (f *fakeRoleRepo) Create(ctx context.Context, ro *roleModels.Role) error {
	return nil
}

func (f *fakeRoleRepo) Update(ctx context.Context, ro *roleModels.Role) error {
	return nil
}

func (f *fakeRoleRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (f *fakeRoleRepo) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	return nil
}

func (f *fakeRoleRepo) AssignPermissionsWithScope(ctx context.Context, roleID string, assignments []roleModels.RolePermission) error {
	return nil
}

func (f *fakeRoleRepo) GetPermissions(ctx context.Context, roleID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleRepo) CountUsersByRoleID(ctx context.Context, roleID string) (int64, error) {
	return 0, nil
}

func (f *fakeRoleRepo) CountAdmins(ctx context.Context) (int64, error) {
	return 0, nil
}

type fakeNotificationRepo struct {
	created []notificationModels.Notification
}

func (f *fakeNotificationRepo) List(ctx context.Context, params notificationRepo.ListParams) ([]notificationModels.Notification, int64, error) {
	return nil, 0, nil
}

func (f *fakeNotificationRepo) CountUnread(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}

func (f *fakeNotificationRepo) MarkAsRead(ctx context.Context, userID, id string, readAt time.Time) (*notificationModels.Notification, error) {
	return nil, nil
}

func (f *fakeNotificationRepo) MarkAllAsRead(ctx context.Context, userID string, readAt time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeNotificationRepo) CreateBulk(ctx context.Context, notifications []notificationModels.Notification) error {
	f.created = append(f.created, notifications...)
	return nil
}

type noopAuditService struct{}

func (noopAuditService) Log(ctx context.Context, action string, targetID string, metadata map[string]interface{}) {}
func (noopAuditService) LogWithReason(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{}) {}
func (noopAuditService) LogWithChanges(ctx context.Context, action string, targetID string, metadata map[string]interface{}, changes interface{}) {}
func (noopAuditService) LogWithChangesFull(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{}, changes interface{}) {}

type noopEventPublisher struct{}

func (noopEventPublisher) Publish(ctx context.Context, event infraEvents.Event) error { return nil }
func (noopEventPublisher) PublishAsync(ctx context.Context, event infraEvents.Event) {}

func TestForgotPasswordSetsTenantIDOnNotifications(t *testing.T) {
	ctx := context.Background()
	requestingUser := &userModels.User{
		ID:       uuid.NewString(),
		TenantID: uuid.NewString(),
		Email:    "admin@example.com",
		Name:     "Tenant Admin",
		RoleID:   uuid.NewString(),
	}
	adminRole := &roleModels.Role{ID: uuid.NewString(), Code: "admin"}
	adminUser := userModels.User{
		ID:     uuid.NewString(),
		Email:  "tenant-admin@example.com",
		Name:   "Tenant Admin Recipient",
		RoleID: adminRole.ID,
	}

	passwordResetRepo := &fakePasswordResetRepo{}
	userRepo := &fakeUserRepo{user: requestingUser, users: []userModels.User{adminUser}}
	roleRepo := &fakeRoleRepo{role: adminRole}
	notificationRepo := &fakeNotificationRepo{}

	uc := NewPasswordResetUsecase(
		passwordResetRepo,
		userRepo,
		roleRepo,
		notificationRepo,
		noopAuditService{},
		noopEventPublisher{},
		(*redis.Client)(nil),
	)

	resp, err := uc.ForgotPassword(ctx, &resetDto.ForgotPasswordRequest{Email: requestingUser.Email})
	if err != nil {
		t.Fatalf("ForgotPassword returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if len(notificationRepo.created) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notificationRepo.created))
	}
	created := notificationRepo.created[0]
	if created.TenantID != requestingUser.TenantID {
		t.Fatalf("expected tenant_id %s, got %s", requestingUser.TenantID, created.TenantID)
	}
	if created.UserID != adminUser.ID {
		t.Fatalf("expected admin recipient %s, got %s", adminUser.ID, created.UserID)
	}
	if created.EntityType != "password_reset_request" {
		t.Fatalf("expected entity_type password_reset_request, got %s", created.EntityType)
	}
	if created.EntityID != requestingUser.ID {
		t.Fatalf("expected entity_id %s, got %s", requestingUser.ID, created.EntityID)
	}
	if created.Message == "" {
		t.Fatal("expected token-bearing message, got empty message")
	}
	if len(passwordResetRepo.created) != 1 {
		t.Fatalf("expected 1 password reset record, got %d", len(passwordResetRepo.created))
	}
}