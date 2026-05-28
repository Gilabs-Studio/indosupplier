package usecase

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/auth/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/events"
	jwtManager "github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	refreshTokenModels "github.com/gilabs/indosupplier/api/internal/refresh_token/data/models"
	refreshTokenRepo "github.com/gilabs/indosupplier/api/internal/refresh_token/data/repositories"
	userRepo "github.com/gilabs/indosupplier/api/internal/user/data/repositories"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserInactive        = errors.New("user is inactive")
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid")
	ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
)

type AuthUsecase interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authUsecase struct {
	db               *gorm.DB
	userRepo         userRepo.UserRepository
	refreshTokenRepo refreshTokenRepo.RefreshTokenRepository
	jwtManager       *jwtManager.JWTManager
	eventPublisher   events.EventPublisher
}

func NewAuthUsecase(
	db *gorm.DB,
	userRepository userRepo.UserRepository,
	refreshTokenRepository refreshTokenRepo.RefreshTokenRepository,
	jwt *jwtManager.JWTManager,
	eventPublisher events.EventPublisher,
) AuthUsecase {
	return &authUsecase{
		db:               db,
		userRepo:         userRepository,
		refreshTokenRepo: refreshTokenRepository,
		jwtManager:       jwt,
		eventPublisher:   eventPublisher,
	}
}

func (u *authUsecase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := u.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := u.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	tokenID, err := u.jwtManager.ExtractRefreshTokenID(refreshToken)
	if err != nil {
		return nil, err
	}

	refreshTokenEntity := &refreshTokenModels.RefreshToken{
		UserID:    user.ID,
		TokenID:   tokenID,
		ExpiresAt: apptime.Now().Add(u.jwtManager.RefreshTokenTTL()),
		Revoked:   false,
	}

	if err := u.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		User: &dto.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(u.jwtManager.AccessTokenTTL().Seconds()),
	}, nil
}

func (u *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	userID, tokenID, err := u.jwtManager.ValidateRefreshTokenWithID(refreshToken)
	if err != nil {
		return nil, ErrRefreshTokenInvalid
	}

	tokenEntity, err := u.refreshTokenRepo.FindByTokenID(ctx, tokenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefreshTokenInvalid
		}
		return nil, err
	}

	if tokenEntity.Revoked {
		return nil, ErrRefreshTokenRevoked
	}

	if tokenEntity.IsExpired() {
		return nil, ErrRefreshTokenExpired
	}

	if tokenEntity.UserID != userID {
		return nil, ErrRefreshTokenInvalid
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	newAccessToken, err := u.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := u.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	newTokenID, err := u.jwtManager.ExtractRefreshTokenID(newRefreshToken)
	if err != nil {
		return nil, err
	}

	tokenEntity.Revoked = true
	if err := u.refreshTokenRepo.Revoke(ctx, tokenEntity.TokenID); err != nil {
		return nil, err
	}

	if err := u.refreshTokenRepo.Create(ctx, &refreshTokenModels.RefreshToken{
		UserID:    user.ID,
		TokenID:   newTokenID,
		ExpiresAt: apptime.Now().Add(u.jwtManager.RefreshTokenTTL()),
		Revoked:   false,
	}); err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		User: &dto.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int(u.jwtManager.AccessTokenTTL().Seconds()),
	}, nil
}

func (u *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	_, tokenID, err := u.jwtManager.ValidateRefreshTokenWithID(refreshToken)
	if err != nil {
		return nil
	}

	tokenEntity, err := u.refreshTokenRepo.FindByTokenID(ctx, tokenID)
	if err != nil {
		return nil
	}

	tokenEntity.Revoked = true
	return u.refreshTokenRepo.Revoke(ctx, tokenEntity.TokenID)
}
