package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/auth/domain/dto"
	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/events"
	jwtManager "github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	refreshTokenModels "github.com/gilabs/indosupplier/api/internal/refresh_token/data/models"
	refreshTokenRepo "github.com/gilabs/indosupplier/api/internal/refresh_token/data/repositories"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	userModels "github.com/gilabs/indosupplier/api/internal/user/data/models"
	userRepo "github.com/gilabs/indosupplier/api/internal/user/data/repositories"
)

func normalizeSlug(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return ""
	}

	var builder strings.Builder
	lastDash := false
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}

		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserInactive        = errors.New("user is inactive")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid")
	ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
)

type AuthUsecase interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error)
	BecomeSupplier(ctx context.Context, userID string, req *dto.SupplierOnboardingRequest) (*dto.UserResponse, error)
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

func (u *authUsecase) toAuthUserResponse(ctx context.Context, userID, name, email string) *dto.UserResponse {
	resp := &dto.UserResponse{
		ID:    userID,
		Name:  name,
		Email: email,
		Capabilities: dto.AccountCapabilitiesResponse{
			Buyer: true,
		},
	}

	accountCtx, err := u.userRepo.FindAccountContext(ctx, userID)
	if err != nil || accountCtx == nil {
		return resp
	}

	if accountCtx.BuyerProfileID != "" {
		resp.BuyerProfile = &dto.AccountProfileRefResponse{ID: accountCtx.BuyerProfileID}
	}
	if accountCtx.SupplierProfileID != "" {
		resp.Capabilities.Supplier = true
		resp.SupplierProfile = &dto.AccountProfileRefResponse{
			ID:     accountCtx.SupplierProfileID,
			Status: accountCtx.SupplierStatus,
		}
	}

	return resp
}

func (u *authUsecase) issueLoginResponse(ctx context.Context, userID, name, email string) (*dto.LoginResponse, error) {
	accessToken, err := u.jwtManager.GenerateAccessToken(userID, email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := u.jwtManager.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	tokenID, err := u.jwtManager.ExtractRefreshTokenID(refreshToken)
	if err != nil {
		return nil, err
	}

	refreshTokenEntity := &refreshTokenModels.RefreshToken{
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: apptime.Now().Add(u.jwtManager.RefreshTokenTTL()),
		Revoked:   false,
	}

	if err := u.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		User:         u.toAuthUserResponse(ctx, userID, name, email),
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(u.jwtManager.AccessTokenTTL().Seconds()),
	}, nil
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

	return u.issueLoginResponse(ctx, user.ID, user.Name, user.Email)
}

func (u *authUsecase) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)

	if _, err := u.userRepo.FindByEmail(ctx, email); err == nil {
		return nil, ErrUserAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	companyName := strings.TrimSpace(req.CompanyName)
	if companyName == "" {
		companyName = name
	}
	industry := strings.TrimSpace(req.Industry)
	if industry == "" {
		industry = "General Procurement"
	}

	user := &userModels.User{
		Email:     email,
		Password:  string(hashedPassword),
		Name:      name,
		AvatarURL: "https://api.dicebear.com/7.x/lorelei/svg?seed=" + url.QueryEscape(email),
		Status:    "active",
	}

	if err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505") {
				return ErrUserAlreadyExists
			}
			return err
		}

		buyerProfile := &buyerModels.BuyerProfile{
			UserID:              user.ID,
			FullName:            name,
			CompanyName:         companyName,
			CountryCode:         "ID",
			Industry:            industry,
			PurchaseFrequency:   "monthly",
			ProfileCompleteness: 35,
		}

		return tx.Create(buyerProfile).Error
	}); err != nil {
		return nil, err
	}

	return u.issueLoginResponse(ctx, user.ID, user.Name, user.Email)
}

func (u *authUsecase) BecomeSupplier(ctx context.Context, userID string, req *dto.SupplierOnboardingRequest) (*dto.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrUserInactive
	}

	var supplierProfileID string
	if err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing supplierModels.SupplierProfile
		if err := tx.Where("user_id = ?", userID).First(&existing).Error; err == nil {
			supplierProfileID = existing.ID
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var primaryCategoryID string
		primaryCategoryName := strings.TrimSpace(req.PrimaryCategory)
		subcategoryName := strings.TrimSpace(req.Subcategory)

		findOrCreateCategory := func(name string, parentID *string) (string, error) {
			slug := normalizeSlug(name)
			if slug == "" {
				return "", nil
			}

			query := tx.Where("slug = ?", slug)
			if parentID == nil {
				query = query.Where("parent_id IS NULL")
			} else {
				query = query.Where("parent_id = ?", *parentID)
			}

			var existingCategory supplierModels.Category
			if err := query.First(&existingCategory).Error; err == nil {
				return existingCategory.ID, nil
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return "", err
			}

			category := &supplierModels.Category{
				ParentID:    parentID,
				Slug:        slug,
				Name:        strings.TrimSpace(name),
				Description: fmt.Sprintf("Category created during supplier registration for %s.", strings.TrimSpace(name)),
				IsActive:    true,
			}
			if err := tx.Create(category).Error; err != nil {
				// On duplicate key, re-query to get the existing category
				if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key") {
					var refetch supplierModels.Category
					if refetchErr := query.First(&refetch).Error; refetchErr == nil {
						return refetch.ID, nil
					}
				}
				return "", err
			}
			return category.ID, nil
		}

		if primaryCategoryName != "" {
			createdCategoryID, err := findOrCreateCategory(primaryCategoryName, nil)
			if err != nil {
				return err
			}
			primaryCategoryID = createdCategoryID
		}

		supplierProfile := &supplierModels.SupplierProfile{
			UserID:                 userID,
			CompanyName:            strings.TrimSpace(req.CompanyName),
			CompanyType:            strings.TrimSpace(req.CompanyType),
			TaxStatus:              strings.TrimSpace(req.TaxStatus),
			NPWP:                   strings.TrimSpace(req.NPWP),
			CountryCode:            "ID",
			ProvinceID:             strings.TrimSpace(req.ProvinceID),
			CityID:                 strings.TrimSpace(req.CityID),
			Address:                strings.TrimSpace(req.Address),
			BusinessHours:          strings.TrimSpace(req.BusinessHours),
			Timezone:               strings.TrimSpace(req.Timezone),
			Description:            strings.TrimSpace(req.Description),
			Phone:                  strings.TrimSpace(req.Phone),
			WhatsApp:               strings.TrimSpace(req.WhatsApp),
			Email:                  strings.TrimSpace(req.Email),
			Website:                strings.TrimSpace(req.Website),
			VerificationLevel:      1,
			IsPremiumVerified:      false,
			ResponseRate:           0,
			AvgResponseTimeMinutes: 0,
			StarRating:             0,
			ReviewCount:            0,
			ProfileCompleteness:    60,
			Status:                 "active",
		}
		if supplierProfile.BusinessHours == "" {
			supplierProfile.BusinessHours = "Monday-Friday 08:00-17:00"
		}
		if supplierProfile.Timezone == "" {
			supplierProfile.Timezone = "Asia/Jakarta"
		}
		if supplierProfile.Email == "" {
			supplierProfile.Email = user.Email
		}
		if err := tx.Create(supplierProfile).Error; err != nil {
			return err
		}
		supplierProfileID = supplierProfile.ID

		if primaryCategoryID != "" {
			if err := tx.Create(&supplierModels.SupplierCategory{
				SupplierProfileID: supplierProfile.ID,
				CategoryID:        primaryCategoryID,
				IsPrimary:         true,
			}).Error; err != nil {
				return err
			}

			if subcategoryName != "" {
				subCategoryID, err := findOrCreateCategory(subcategoryName, &primaryCategoryID)
				if err != nil {
					return err
				}
				if subCategoryID != "" {
					if err := tx.Create(&supplierModels.SupplierCategory{
						SupplierProfileID: supplierProfile.ID,
						CategoryID:        subCategoryID,
						IsPrimary:         false,
					}).Error; err != nil {
						return err
					}
				}
			}
		}

		productName := strings.TrimSpace(req.FirstProductName)
		if productName == "" {
			return nil
		}

		return tx.Create(&supplierModels.SupplierProduct{
			SupplierProfileID: supplierProfile.ID,
			CategoryID:        primaryCategoryID,
			Name:              productName,
			Description:       strings.TrimSpace(req.FirstProductPrice),
			Currency:          "IDR",
			SortOrder:         1,
		}).Error
	}); err != nil {
		return nil, err
	}

	resp := u.toAuthUserResponse(ctx, user.ID, user.Name, user.Email)
	if resp.SupplierProfile == nil && supplierProfileID != "" {
		resp.Capabilities.Supplier = true
		resp.SupplierProfile = &dto.AccountProfileRefResponse{ID: supplierProfileID, Status: "active"}
	}
	return resp, nil
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
		User:         u.toAuthUserResponse(ctx, user.ID, user.Name, user.Email),
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
