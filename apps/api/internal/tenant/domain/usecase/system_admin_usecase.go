package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/gims/api/internal/tenant/data/repositories"
	"github.com/gilabs/gims/api/internal/tenant/domain/dto"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrSysAdminInvalidCredentials = errors.New("invalid credentials")
	ErrSysAdminNotFound           = errors.New("system admin not found")
	ErrSysAdminDisabled           = errors.New("system admin account is disabled")
	ErrSysAdminEmailAlreadyTaken  = errors.New("system admin email already taken")

	systemAdminDummyHash string
)

func init() {
	hash, err := bcrypt.GenerateFromPassword([]byte("system-admin-invalid-password"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	systemAdminDummyHash = string(hash)
}

// SystemAdminUsecase handles system admin authentication
type SystemAdminUsecase interface {
	Login(ctx context.Context, req *dto.SystemAdminLoginRequest) (*dto.SystemAdminLoginResponse, error)
	UpdateProfile(ctx context.Context, adminID string, req *dto.SystemAdminUpdateProfileRequest) (*dto.SystemAdminResponse, error)
	ChangePassword(ctx context.Context, adminID string, req *dto.SystemAdminChangePasswordRequest) error
}

type systemAdminUsecase struct {
	repo       repositories.SystemAdminRepository
	jwtManager *jwt.JWTManager
}

// NewSystemAdminUsecase creates a new SystemAdminUsecase
func NewSystemAdminUsecase(repo repositories.SystemAdminRepository, jwtManager *jwt.JWTManager) SystemAdminUsecase {
	return &systemAdminUsecase{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (u *systemAdminUsecase) Login(ctx context.Context, req *dto.SystemAdminLoginRequest) (*dto.SystemAdminLoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	admin, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Preserve constant-time behavior even when account does not exist.
			_ = bcrypt.CompareHashAndPassword([]byte(systemAdminDummyHash), []byte(req.Password))
			return nil, ErrSysAdminInvalidCredentials
		}
		return nil, err
	}

	if admin.Status != "active" {
		return nil, ErrSysAdminDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, ErrSysAdminInvalidCredentials
	}

	// Generate token with role="system_admin" — tenantID is empty since system admins
	// operate at the platform level and are not scoped to any single tenant.
	accessToken, err := u.jwtManager.GenerateAccessToken(admin.ID, admin.Email, "system_admin", "")
	if err != nil {
		return nil, err
	}

	expiresIn := int(u.jwtManager.AccessTokenTTL().Seconds())

	return &dto.SystemAdminLoginResponse{
		Admin: &dto.SystemAdminResponse{
			ID:       admin.ID,
			Email:    admin.Email,
			Name:     admin.Name,
			Username: admin.Name,
			Status:   admin.Status,
		},
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
		// Tokens set in cookies by handler, not exposed in body
	}, nil
}

func (u *systemAdminUsecase) UpdateProfile(ctx context.Context, adminID string, req *dto.SystemAdminUpdateProfileRequest) (*dto.SystemAdminResponse, error) {
	admin, err := u.repo.FindByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSysAdminNotFound
		}
		return nil, err
	}

	nextEmail := strings.ToLower(strings.TrimSpace(req.Email))
	nextName := strings.TrimSpace(req.Username)

	if nextEmail != admin.Email {
		existing, findErr := u.repo.FindByEmail(ctx, nextEmail)
		if findErr == nil && existing != nil && existing.ID != admin.ID {
			return nil, ErrSysAdminEmailAlreadyTaken
		}
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil, findErr
		}
	}

	admin.Email = nextEmail
	admin.Name = nextName

	if err := u.repo.Update(ctx, admin); err != nil {
		return nil, err
	}

	return &dto.SystemAdminResponse{
		ID:       admin.ID,
		Email:    admin.Email,
		Name:     admin.Name,
		Username: admin.Name,
		Status:   admin.Status,
	}, nil
}

func (u *systemAdminUsecase) ChangePassword(ctx context.Context, adminID string, req *dto.SystemAdminChangePasswordRequest) error {
	admin, err := u.repo.FindByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSysAdminNotFound
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.CurrentPassword)); err != nil {
		return ErrSysAdminInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin.Password = string(hashedPassword)
	if err := u.repo.Update(ctx, admin); err != nil {
		return err
	}

	return nil
}
