package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/gilabs/gims/api/internal/auth/domain/dto"
	"github.com/gilabs/gims/api/internal/auth/domain/usecase"
	authMocks "github.com/gilabs/gims/api/internal/auth/domain/usecase/mocks"
	infraEvents "github.com/gilabs/gims/api/internal/core/infrastructure/events"
	jwtManager "github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/role/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
	userMocks "github.com/gilabs/gims/api/internal/user/domain/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAuthUsecase_Login(t *testing.T) {
	// Setup generic DB for transaction support
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// JWT Manager setup
	jwtOpts := jwtManager.Options{
		AccessSecretKey:  "access-secret",
		RefreshSecretKey: "refresh-secret",
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
	}
	jwtMgr := jwtManager.NewJWTManager(jwtOpts)

	type fields struct {
		userRepo         *userMocks.UserRepository
		refreshTokenRepo *authMocks.RefreshTokenRepository
		eventPublisher   *userMocks.EventPublisher
	}
	type args struct {
		ctx context.Context
		req *dto.LoginRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(f fields)
		want    *dto.LoginResponse
		wantErr bool
		err     error
	}{
		{
			name: "Success",
			fields: fields{
				userRepo:         userMocks.NewUserRepository(t),
				refreshTokenRepo: authMocks.NewRefreshTokenRepository(t),
				eventPublisher:   userMocks.NewEventPublisher(t),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.LoginRequest{
					Email:    "test@example.com",
					Password: "password123",
				},
			},
			mock: func(f fields) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 4)
				
				user := &userModels.User{
					ID:       "user-1",
					Email:    "test@example.com",
					Password: string(hash),
					Status:   "active",
					Name:     "Test User",
					Role: &models.Role{
						Code: "admin",
						Name: "Admin",
					},
				}

				f.userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
				f.refreshTokenRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				f.eventPublisher.On("PublishAsync", mock.Anything, mock.MatchedBy(func(e infraEvents.Event) bool {
					return e.GetType() == infraEvents.EventTypeUserLoggedIn
				})).Return()
			},
			want: &dto.LoginResponse{
				User: &dto.UserResponse{
					Email: "test@example.com",
					Role:  "admin",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Credentials",
			fields: fields{
				userRepo:         userMocks.NewUserRepository(t),
				refreshTokenRepo: authMocks.NewRefreshTokenRepository(t),
				eventPublisher:   userMocks.NewEventPublisher(t),
			},
			args: args{
				ctx: context.Background(),
				req: &dto.LoginRequest{
					Email:    "test@example.com",
					Password: "wrongpassword",
				},
			},
			mock: func(f fields) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 4)
				
				user := &userModels.User{
					ID:       "user-1",
					Email:    "test@example.com",
					Password: string(hash),
					Status:   "active",
				}

				f.userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			want:    nil,
			wantErr: true,
			err:     usecase.ErrInvalidCredentials,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// couponUC is not required for Login tests — pass nil (it is only used in RegisterTenant).
		u := usecase.NewAuthUsecase(db, tt.fields.userRepo, tt.fields.refreshTokenRepo, jwtMgr, tt.fields.eventPublisher, nil, nil, nil)

			if tt.mock != nil {
				tt.mock(tt.fields)
			}

			got, err := u.Login(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.err != nil {
					assert.Equal(t, tt.err, err)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.User.Email, got.User.Email)
				assert.NotEmpty(t, got.Token)
			}
		})
	}
}
