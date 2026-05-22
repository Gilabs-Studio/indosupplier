package usecase_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	infraEvents "github.com/gilabs/gims/api/internal/core/infrastructure/events"
	"github.com/gilabs/gims/api/internal/role/data/models"
	"github.com/gilabs/gims/api/internal/role/domain/dto"
	"github.com/gilabs/gims/api/internal/role/domain/usecase"
	userMocks "github.com/gilabs/gims/api/internal/user/domain/usecase/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestRoleUsecase_GetByID(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	type fields struct {
		roleRepo       *userMocks.RoleRepository
		eventPublisher *userMocks.EventPublisher
		redis          *redis.Client
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
		want    *dto.RoleResponse
		wantErr bool
		err     error
	}{
		{
			name: "Success",
			fields: fields{
				roleRepo:       userMocks.NewRoleRepository(t),
				eventPublisher: userMocks.NewEventPublisher(t),
				redis:          rdb,
			},
			args: args{
				ctx: context.Background(),
				id:  "role-1",
			},
			mock: func(f fields) {
				role := &models.Role{
					ID:   "role-1",
					Name: "Admin",
					Code: "admin",
				}
				f.roleRepo.On("FindByID", mock.Anything, "role-1").Return(role, nil)
			},
			want: &dto.RoleResponse{
				ID:   "role-1",
				Name: "Admin",
				Code: "admin",
			},
			wantErr: false,
		},
		{
			name: "Not Found",
			fields: fields{
				roleRepo:       userMocks.NewRoleRepository(t),
				eventPublisher: userMocks.NewEventPublisher(t),
				redis:          rdb,
			},
			args: args{
				ctx: context.Background(),
				id:  "role-99",
			},
			mock: func(f fields) {
				f.roleRepo.On("FindByID", mock.Anything, "role-99").Return(nil, gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: true,
			err:     usecase.ErrRoleNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usecase.NewRoleUsecase(tt.fields.roleRepo, tt.fields.eventPublisher, tt.fields.redis, nil, nil, nil)

			if tt.mock != nil {
				tt.mock(tt.fields)
			}

			// Clear redis
			mr.FlushAll()

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
				assert.Equal(t, tt.want.Code, got.Code)
			}
		})
	}
}

func TestRoleUsecase_Create(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	type fields struct {
		roleRepo       *userMocks.RoleRepository
		eventPublisher *userMocks.EventPublisher
		redis          *redis.Client
	}
	type args struct {
		ctx context.Context
		req *dto.CreateRoleRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *dto.RoleResponse
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				roleRepo:       userMocks.NewRoleRepository(t),
				eventPublisher: userMocks.NewEventPublisher(t),
				redis:          rdb,
			},
			args: args{
				ctx: context.Background(),
				req: &dto.CreateRoleRequest{
					Name: "New Role",
					Code: "new_role",
				},
			},
			mock: func(f fields) {
				f.roleRepo.On("FindByCode", mock.Anything, "new_role").Return(nil, gorm.ErrRecordNotFound)
				f.roleRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				f.roleRepo.On("FindByID", mock.Anything, mock.Anything).Return(&models.Role{
					ID:   "new-id",
					Name: "New Role",
					Code: "new_role",
				}, nil)
				f.eventPublisher.On("PublishAsync", mock.Anything, mock.MatchedBy(func(e infraEvents.Event) bool {
					return e.GetType() == infraEvents.EventTypeRoleCreated
				})).Return()
			},
			want: &dto.RoleResponse{
				Name: "New Role",
				Code: "new_role",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usecase.NewRoleUsecase(tt.fields.roleRepo, tt.fields.eventPublisher, tt.fields.redis, nil, nil, nil)

			if tt.mock != nil {
				tt.mock(tt.fields)
			}

			got, err := u.Create(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Code, got.Code)
			}
		})
	}
}
