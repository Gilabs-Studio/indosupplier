package usecase_test

import (
	"context"
	"testing"

	permDto "github.com/gilabs/gims/api/internal/permission/domain/dto"
	"github.com/gilabs/gims/api/internal/permission/domain/usecase"
	permMocks "github.com/gilabs/gims/api/internal/permission/domain/usecase/mocks"
	"github.com/gilabs/gims/api/internal/user/data/models"
	userMocks "github.com/gilabs/gims/api/internal/user/domain/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestPermissionUsecase_GetUserPermissions(t *testing.T) {
	type fields struct {
		permissionRepo *permMocks.PermissionRepository
		userRepo       *userMocks.UserRepository
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *permDto.GetUserPermissionsResponse
		wantErr bool
		err     error
	}{
		{
			name: "Success",
			fields: fields{
				permissionRepo: permMocks.NewPermissionRepository(t),
				userRepo:       userMocks.NewUserRepository(t),
			},
			args: args{
				ctx:    context.Background(),
				userID: "user-1",
			},
			mock: func(f fields) {
				f.userRepo.On("FindByID", mock.Anything, "user-1").Return(&models.User{ID: "user-1"}, nil)
				f.permissionRepo.On("GetUserPermissions", mock.Anything, "user-1").Return(&permDto.GetUserPermissionsResponse{
					Menus: []permDto.MenuWithActionsResponse{
						{
							ID:   "menu-1",
							Name: "Users",
							Actions: []permDto.ActionResponse{
								{Code: "read:users", Access: true},
							},
						},
					},
				}, nil)
			},
			want: &permDto.GetUserPermissionsResponse{
				Menus: []permDto.MenuWithActionsResponse{
					{
						ID:   "menu-1",
						Name: "Users",
						Actions: []permDto.ActionResponse{
							{Code: "read:users", Access: true},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			fields: fields{
				permissionRepo: permMocks.NewPermissionRepository(t),
				userRepo:       userMocks.NewUserRepository(t),
			},
			args: args{
				ctx:    context.Background(),
				userID: "user-99",
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
			u := usecase.NewPermissionUsecase(tt.fields.permissionRepo, tt.fields.userRepo)

			if tt.mock != nil {
				tt.mock(tt.fields)
			}

			got, err := u.GetUserPermissions(tt.args.ctx, tt.args.userID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.err != nil {
					assert.Equal(t, tt.err, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got.Menus, len(tt.want.Menus))
				if len(got.Menus) > 0 {
					assert.Equal(t, tt.want.Menus[0].ID, got.Menus[0].ID)
				}
			}
		})
	}
}
