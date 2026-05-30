package usecase

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/sysadmin/domain/mapper"
	jwtManager "github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAdminInactive      = errors.New("admin account is inactive")
)

type SystemAdminUsecase interface {
	Login(ctx context.Context, req dto.SysadminLoginRequest) (dto.SysadminLoginResponse, error)
}

type systemAdminUsecase struct {
	repo       repositories.SystemAdminRepository
	jwtManager *jwtManager.JWTManager
}

func NewSystemAdminUsecase(repo repositories.SystemAdminRepository, jwt *jwtManager.JWTManager) SystemAdminUsecase {
	return &systemAdminUsecase{
		repo:       repo,
		jwtManager: jwt,
	}
}

func (u *systemAdminUsecase) Login(ctx context.Context, req dto.SysadminLoginRequest) (dto.SysadminLoginResponse, error) {
	sa, err := u.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return dto.SysadminLoginResponse{}, ErrInvalidCredentials
	}

	if sa.Status != "active" {
		return dto.SysadminLoginResponse{}, ErrAdminInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(sa.Password), []byte(req.Password)); err != nil {
		return dto.SysadminLoginResponse{}, ErrInvalidCredentials
	}

	accessToken, err := u.jwtManager.GenerateAccessToken(sa.ID, sa.Email)
	if err != nil {
		return dto.SysadminLoginResponse{}, err
	}

	// We can also generate a refresh token if needed, or simply reuse the access token
	refreshToken, err := u.jwtManager.GenerateRefreshToken(sa.ID)
	if err != nil {
		return dto.SysadminLoginResponse{}, err
	}

	return dto.SysadminLoginResponse{
		Admin:        mapper.ToSysadminResponse(sa),
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(u.jwtManager.AccessTokenTTL().Seconds()),
	}, nil
}
