package usecase

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/buyer/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/buyer/domain/dto"
)

var ErrProfileDocumentNotFound = errors.New("buyer document not found")

type ProfileUsecase interface {
	GetProfile(ctx context.Context, userID string) (*dto.BuyerProfileResponse, error)
	UpdatePersonal(ctx context.Context, userID string, req *dto.UpdateBuyerPersonalRequest) (*dto.BuyerProfileResponse, error)
	UpdateCompany(ctx context.Context, userID string, req *dto.UpdateBuyerCompanyRequest) (*dto.BuyerProfileResponse, error)
	ListDocuments(ctx context.Context, userID string) ([]dto.BuyerDocumentResponse, error)
	UploadDocument(ctx context.Context, userID string, req *dto.UploadBuyerDocumentRequest) (*dto.BuyerDocumentResponse, error)
}

type profileUsecase struct {
	profileRepo repositories.ProfileRepository
}

func NewProfileUsecase(profileRepo repositories.ProfileRepository) ProfileUsecase {
	return &profileUsecase{profileRepo: profileRepo}
}

func (u *profileUsecase) getProfile(ctx context.Context, userID string) (*models.BuyerProfile, error) {
	profile, err := u.profileRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBuyerProfileNotFound
		}
		return nil, err
	}
	return profile, nil
}

func (u *profileUsecase) toProfileResponse(ctx context.Context, profile *models.BuyerProfile) *dto.BuyerProfileResponse {
	email, _ := u.profileRepo.GetUserEmail(ctx, profile.UserID)
	return &dto.BuyerProfileResponse{
		ID:                  profile.ID,
		UserID:              profile.UserID,
		Email:               email,
		FullName:            profile.FullName,
		CompanyName:         profile.CompanyName,
		CountryCode:         profile.CountryCode,
		Industry:            profile.Industry,
		PurchaseFrequency:   profile.PurchaseFrequency,
		Phone:               profile.Phone,
		Website:             profile.Website,
		Address:             profile.Address,
		ProfileCompleteness: profile.ProfileCompleteness,
		CompanyVerifiedAt:   profile.CompanyVerifiedAt,
		CreatedAt:           profile.CreatedAt,
		UpdatedAt:           profile.UpdatedAt,
	}
}

func calculateBuyerProfileCompleteness(profile *models.BuyerProfile) int {
	fields := []string{
		profile.FullName,
		profile.CompanyName,
		profile.CountryCode,
		profile.Industry,
		profile.PurchaseFrequency,
		profile.Phone,
		profile.Website,
		profile.Address,
	}
	completed := 0
	for _, field := range fields {
		if strings.TrimSpace(field) != "" {
			completed++
		}
	}
	return int(float64(completed) / float64(len(fields)) * 100)
}

func (u *profileUsecase) GetProfile(ctx context.Context, userID string) (*dto.BuyerProfileResponse, error) {
	profile, err := u.getProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.toProfileResponse(ctx, profile), nil
}

func (u *profileUsecase) UpdatePersonal(ctx context.Context, userID string, req *dto.UpdateBuyerPersonalRequest) (*dto.BuyerProfileResponse, error) {
	profile, err := u.getProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile.FullName = strings.TrimSpace(req.FullName)
	profile.Phone = strings.TrimSpace(req.Phone)
	profile.ProfileCompleteness = calculateBuyerProfileCompleteness(profile)
	if err := u.profileRepo.UpdateProfile(ctx, profile); err != nil {
		return nil, err
	}
	return u.toProfileResponse(ctx, profile), nil
}

func (u *profileUsecase) UpdateCompany(ctx context.Context, userID string, req *dto.UpdateBuyerCompanyRequest) (*dto.BuyerProfileResponse, error) {
	profile, err := u.getProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile.CompanyName = strings.TrimSpace(req.CompanyName)
	profile.Industry = strings.TrimSpace(req.Industry)
	profile.Website = strings.TrimSpace(req.Website)
	profile.Address = strings.TrimSpace(req.Address)
	profile.ProfileCompleteness = calculateBuyerProfileCompleteness(profile)
	if err := u.profileRepo.UpdateProfile(ctx, profile); err != nil {
		return nil, err
	}
	return u.toProfileResponse(ctx, profile), nil
}

func toBuyerDocumentResponse(document models.BuyerDocument) dto.BuyerDocumentResponse {
	return dto.BuyerDocumentResponse{
		ID:             document.ID,
		BuyerProfileID: document.BuyerProfileID,
		DocumentType:   document.DocumentType,
		DocumentNumber: document.DocumentNumber,
		FileURL:        document.FileURL,
		Status:         document.Status,
		ReviewedAt:     document.ReviewedAt,
		ReviewReason:   document.ReviewReason,
		CreatedAt:      document.CreatedAt,
		UpdatedAt:      document.UpdatedAt,
	}
}

func (u *profileUsecase) ListDocuments(ctx context.Context, userID string) ([]dto.BuyerDocumentResponse, error) {
	profile, err := u.getProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	documents, err := u.profileRepo.ListDocuments(ctx, profile.ID)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.BuyerDocumentResponse, 0, len(documents))
	for _, document := range documents {
		responses = append(responses, toBuyerDocumentResponse(document))
	}
	return responses, nil
}

func (u *profileUsecase) UploadDocument(ctx context.Context, userID string, req *dto.UploadBuyerDocumentRequest) (*dto.BuyerDocumentResponse, error) {
	profile, err := u.getProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	document := &models.BuyerDocument{
		BuyerProfileID: profile.ID,
		DocumentType:   strings.TrimSpace(req.DocumentType),
		DocumentNumber: strings.TrimSpace(req.DocumentNumber),
		FileURL:        strings.TrimSpace(req.FileURL),
		Status:         "pending",
	}
	if err := u.profileRepo.CreateDocument(ctx, document); err != nil {
		return nil, err
	}
	response := toBuyerDocumentResponse(*document)
	return &response, nil
}
