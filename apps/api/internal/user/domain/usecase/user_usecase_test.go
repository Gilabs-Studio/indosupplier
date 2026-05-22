package usecase_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	infraEvents "github.com/gilabs/gims/api/internal/core/infrastructure/events"
	"github.com/gilabs/gims/api/internal/role/data/models"
	tenantModels "github.com/gilabs/gims/api/internal/tenant/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	"github.com/gilabs/gims/api/internal/user/domain/dto"
	"github.com/gilabs/gims/api/internal/user/domain/usecase"
	"github.com/gilabs/gims/api/internal/user/domain/usecase/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type subscriptionFinderStub struct {
	sub *tenantModels.TenantSubscription
	err error
}

func (s *subscriptionFinderStub) FindActiveByTenantID(ctx context.Context, tenantID string) (*tenantModels.TenantSubscription, error) {
	return s.sub, s.err
}

func TestUserUsecase_GetByID(t *testing.T) {
	// Setup miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	type fields struct {
		userRepo         *mocks.UserRepository
		roleRepo         *mocks.RoleRepository
		tenantRepo       *mocks.TenantRepository
		subscriptionRepo *subscriptionFinderStub
		auditService     *mocks.AuditService
		eventPublisher   *mocks.EventPublisher
		redis            *redis.Client
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *dto.UserResponse
		wantErr bool
		err     error
	}{
		{
			name: "Success - From DB",
			fields: fields{
				userRepo:         mocks.NewUserRepository(t),
				roleRepo:         mocks.NewRoleRepository(t),
				tenantRepo:       mocks.NewTenantRepository(t),
				subscriptionRepo: &subscriptionFinderStub{},
				auditService:     mocks.NewAuditService(t),
				eventPublisher:   mocks.NewEventPublisher(t),
				redis:            rdb,
			},
			args: args{
				ctx: context.Background(),
				id:  "user-1",
			},
			mock: func(f fields) {
				user := &userModels.User{
					ID:    "user-1",
					Email: "test@example.com",
					Name:  "Test User",
				}
				f.userRepo.On("FindByID", mock.Anything, "user-1").Return(user, nil)
			},
			want: &dto.UserResponse{
				ID:    "user-1",
				Email: "test@example.com",
				Name:  "Test User",
			},
			wantErr: false,
		},
		{
			name: "Success - From Cache",
			fields: fields{
				userRepo:       mocks.NewUserRepository(t),
				roleRepo:       mocks.NewRoleRepository(t),
				tenantRepo:     mocks.NewTenantRepository(t),
				auditService:   mocks.NewAuditService(t),
				eventPublisher: mocks.NewEventPublisher(t),
				redis:          rdb,
			},
			args: args{
				ctx: context.Background(),
				id:  "user-cached",
			},
			mock: func(f fields) {
				// Pre-populate cache
				cachedUser := &dto.UserResponse{
					ID:    "user-cached",
					Email: "cached@example.com",
					Name:  "Cached User",
				}
				data, _ := json.Marshal(cachedUser)
				_ = f.redis.Set(context.Background(), "users:id:scope:public:id:user-cached", data, time.Minute).Err()
			},
			want: &dto.UserResponse{
				ID:    "user-cached",
				Email: "cached@example.com",
				Name:  "Cached User",
			},
			wantErr: false,
		},
		{
			name: "Not Found",
			fields: fields{
				userRepo:       mocks.NewUserRepository(t),
				roleRepo:       mocks.NewRoleRepository(t),
				tenantRepo:     mocks.NewTenantRepository(t),
				auditService:   mocks.NewAuditService(t),
				eventPublisher: mocks.NewEventPublisher(t),
				redis:          rdb,
			},
			args: args{
				ctx: context.Background(),
				id:  "user-99",
			},
			mock: func(f fields) {
				f.userRepo.On("FindByID", mock.Anything, "user-99").Return(nil, gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: true,
			err:     usecase.ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usecase.NewUserUsecase(tt.fields.userRepo, tt.fields.roleRepo, tt.fields.tenantRepo, tt.fields.subscriptionRepo, tt.fields.auditService, tt.fields.eventPublisher, tt.fields.redis)

			if tt.mock != nil {
				tt.mock(tt.fields)
			}

			// Clear redis for DB test
			if tt.name == "Success - From DB" {
				mr.FlushAll()
			}

			got, err := u.GetByID(tt.args.ctx, tt.args.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.err != nil {
					assert.Equal(t, tt.err, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Email, got.Email)
			}
		})
	}
}

func TestUserUsecaseGetLimit(t *testing.T) {
	type fields struct {
		userRepo         *mocks.UserRepository
		tenantRepo       *mocks.TenantRepository
		subscriptionRepo *subscriptionFinderStub
	}

	fieldsValue := fields{
		userRepo:         mocks.NewUserRepository(t),
		tenantRepo:       mocks.NewTenantRepository(t),
		subscriptionRepo: &subscriptionFinderStub{sub: &tenantModels.TenantSubscription{SeatLimit: 12, UserCount: 1}},
	}
	fieldsValue.userRepo.On("Count", mock.Anything).Return(int64(1), nil)
	fieldsValue.tenantRepo.On("FindByID", mock.Anything, "").Return(&tenantModels.Tenant{MaxUsers: 0}, nil)

	u := usecase.NewUserUsecase(fieldsValue.userRepo, mocks.NewRoleRepository(t), fieldsValue.tenantRepo, fieldsValue.subscriptionRepo, mocks.NewAuditService(t), mocks.NewEventPublisher(t), redis.NewClient(&redis.Options{Addr: "localhost:0"}))

	limit, err := u.GetLimit(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 1, limit.Current)
	assert.Equal(t, 12, limit.Max)
}

func TestUserUsecase_Create(t *testing.T) {
	// Setup miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	type fields struct {
		userRepo         *mocks.UserRepository
		roleRepo         *mocks.RoleRepository
		tenantRepo       *mocks.TenantRepository
		subscriptionRepo *subscriptionFinderStub
		auditService     *mocks.AuditService
		eventPublisher   *mocks.EventPublisher
		redis            *redis.Client
	}
	type args struct {
		ctx context.Context
		req *dto.CreateUserRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *dto.UserResponse
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				userRepo:         mocks.NewUserRepository(t),
				roleRepo:         mocks.NewRoleRepository(t),
				tenantRepo:       mocks.NewTenantRepository(t),
				subscriptionRepo: &subscriptionFinderStub{sub: &tenantModels.TenantSubscription{SeatLimit: 5, UserCount: 1}},
				auditService:     mocks.NewAuditService(t),
				eventPublisher:   mocks.NewEventPublisher(t),
				redis:            rdb,
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateUserRequest{
					Email:    "new@example.com",
					Password: "password123",
					Name:     "New User",
					RoleID:   "role-1",
				},
			},
			mock: func(f fields) {
				// Limit check: Count returns 0, tenantRepo returns safe default on empty tenant_id
				f.userRepo.On("Count", mock.Anything).Return(int64(0), nil)
				f.tenantRepo.On("FindByID", mock.Anything, "").Return(&tenantModels.Tenant{MaxUsers: 0}, nil)
				f.roleRepo.On("FindByID", mock.Anything, "role-1").Return(&models.Role{ID: "role-1"}, nil)
				f.userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, gorm.ErrRecordNotFound)
				f.userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *userModels.User) bool {
					return u.Email == "new@example.com" && u.Name == "New User"
				})).Return(nil)
				f.auditService.On("Log", mock.Anything, "user.create", mock.Anything, mock.Anything).Return()
				f.userRepo.On("FindByID", mock.Anything, mock.Anything).Return(&userModels.User{
					ID:     "new-id",
					Email:  "new@example.com",
					Name:   "New User",
					RoleID: "role-1",
					Status: "active",
				}, nil)
				f.eventPublisher.On("PublishAsync", mock.Anything, mock.MatchedBy(func(e infraEvents.Event) bool {
					return e.GetType() == infraEvents.EventTypeUserCreated
				})).Return()
			},
			want: &dto.UserResponse{
				Email: "new@example.com",
				Name:  "New User",
			},
			wantErr: false,
		},
		{
			name: "Role Not Found",
			fields: fields{
				userRepo:         mocks.NewUserRepository(t),
				roleRepo:         mocks.NewRoleRepository(t),
				tenantRepo:       mocks.NewTenantRepository(t),
				subscriptionRepo: &subscriptionFinderStub{sub: &tenantModels.TenantSubscription{SeatLimit: 5, UserCount: 1}},
				auditService:     mocks.NewAuditService(t),
				eventPublisher:   mocks.NewEventPublisher(t),
				redis:            rdb,
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateUserRequest{
					Email:    "new@example.com",
					RoleID:   "role-99",
					Password: "password",
				},
			},
			mock: func(f fields) {
				f.userRepo.On("Count", mock.Anything).Return(int64(0), nil)
				f.tenantRepo.On("FindByID", mock.Anything, "").Return(&tenantModels.Tenant{MaxUsers: 0}, nil)
				f.roleRepo.On("FindByID", mock.Anything, "role-99").Return(nil, gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usecase.NewUserUsecase(tt.fields.userRepo, tt.fields.roleRepo, tt.fields.tenantRepo, tt.fields.subscriptionRepo, tt.fields.auditService, tt.fields.eventPublisher, tt.fields.redis)

			if tt.mock != nil {
				tt.mock(tt.fields)
			}

			got, err := u.Create(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Email, got.Email)
				assert.Equal(t, tt.want.Name, got.Name)
			}
		})
	}
}
