package usecase

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	"github.com/gilabs/indosupplier/api/internal/verification/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/verification/domain/dto"
	verificationModels "github.com/gilabs/indosupplier/api/internal/verification/data/models"
)

var (
	ErrSupplierProfileNotFound = errors.New("supplier profile not found")
	ErrVerificationIncomplete  = errors.New("verification data is incomplete")
	ErrVerificationNotFound    = errors.New("verification request not found")
)

type VerificationUsecase interface {
	GetVerificationData(ctx context.Context, userID string) (*dto.VerificationResponse, error)
	UpdateVerificationData(ctx context.Context, userID string, req *dto.UpdateVerificationRequest) (*dto.VerificationResponse, error)
	SubmitVerification(ctx context.Context, userID string) (*dto.VerificationResponse, error)
}

type verificationUsecase struct {
	repo repositories.VerificationRepository
}

func NewVerificationUsecase(repo repositories.VerificationRepository) VerificationUsecase {
	return &verificationUsecase{repo: repo}
}

func (u *verificationUsecase) getSupplierProfile(ctx context.Context, userID string) (*models.SupplierProfile, error) {
	profile, err := u.repo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierProfileNotFound
		}
		return nil, err
	}
	return profile, nil
}

func (u *verificationUsecase) getLatestRequest(ctx context.Context, supplierProfileID string) (*verificationModels.VerificationRequest, error) {
	request, err := u.repo.GetLatestRequestBySupplierProfileID(ctx, supplierProfileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return request, nil
}

func verificationStatus(profile *models.SupplierProfile, request *verificationModels.VerificationRequest) string {
	if profile.VerificationLevel >= 2 || profile.IsPremiumVerified {
		return "verified"
	}
	if request == nil {
		return "unverified"
	}
	if request.Status == "pending" {
		return "pending"
	}
	return "unverified"
}

func toVerificationResponse(profile *models.SupplierProfile, request *verificationModels.VerificationRequest) *dto.VerificationResponse {
	response := &dto.VerificationResponse{
		BusinessInfo: dto.VerificationBusinessInfo{
			LegalName:   profile.CompanyName,
			NIBNumber:   profile.NIB,
			Phone:       profile.Phone,
			Description: profile.Description,
			Address:     profile.Address,
			NPWPNumber:  profile.NPWP,
		},
		Status: verificationStatus(profile, request),
	}

	if request == nil {
		return response
	}

	response.BusinessInfo = dto.VerificationBusinessInfo{
		LegalName:       firstNonEmpty(request.LegalName, profile.CompanyName),
		NIBNumber:       firstNonEmpty(request.NIBNumber, profile.NIB),
		EstablishedDate: strings.TrimSpace(request.EstablishedDate),
		Industry:        strings.TrimSpace(request.Industry),
		Phone:           firstNonEmpty(request.BusinessPhone, profile.Phone),
		Description:     firstNonEmpty(request.BusinessDesc, profile.Description),
		Address:         firstNonEmpty(request.BusinessAddress, profile.Address),
		NIBFileURL:      strings.TrimSpace(request.NIBFileURL),
		NPWPNumber:      firstNonEmpty(request.NPWPNumber, profile.NPWP),
		NPWPFileURL:     strings.TrimSpace(request.NPWPFileURL),
		AktaFileURL:     strings.TrimSpace(request.AktaFileURL),
	}
	response.StakeholderInfo = dto.VerificationStakeholderInfo{
		DirectorName: strings.TrimSpace(request.DirectorName),
		DirectorNIK:  strings.TrimSpace(request.DirectorNIK),
	}
	response.BankAccountInfo = dto.VerificationBankAccountInfo{
		BankName:      strings.TrimSpace(request.BankName),
		AccountNumber: strings.TrimSpace(request.AccountNumber),
		AccountName:   strings.TrimSpace(request.AccountName),
	}

	return response
}

func firstNonEmpty(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return strings.TrimSpace(primary)
	}
	return strings.TrimSpace(fallback)
}

func syncProfileFromVerification(profile *models.SupplierProfile, business *dto.VerificationBusinessInfo) {
	if business == nil {
		return
	}
	if value := strings.TrimSpace(business.LegalName); value != "" {
		profile.CompanyName = value
	}
	if value := strings.TrimSpace(business.NIBNumber); value != "" {
		profile.NIB = value
	}
	if value := strings.TrimSpace(business.Phone); value != "" {
		profile.Phone = value
	}
	if value := strings.TrimSpace(business.Description); value != "" {
		profile.Description = value
	}
	if value := strings.TrimSpace(business.Address); value != "" {
		profile.Address = value
	}
	if value := strings.TrimSpace(business.NPWPNumber); value != "" {
		profile.NPWP = value
	}
	if value := strings.TrimSpace(business.EstablishedDate); value != "" {
		parts := strings.Split(value, "-")
		if len(parts) > 0 {
			profile.EstablishedYear = parts[0]
		}
	}
}

func (u *verificationUsecase) ensureRequest(ctx context.Context, supplierProfileID string) (*verificationModels.VerificationRequest, error) {
	request, err := u.getLatestRequest(ctx, supplierProfileID)
	if err != nil {
		return nil, err
	}
	if request != nil {
		return request, nil
	}

	request = &verificationModels.VerificationRequest{
		SupplierProfileID: supplierProfileID,
		RequestType:       "supplier_verification",
		Status:            "draft",
	}
	return request, nil
}

func (u *verificationUsecase) GetVerificationData(ctx context.Context, userID string) (*dto.VerificationResponse, error) {
	profile, err := u.getSupplierProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	request, err := u.getLatestRequest(ctx, profile.ID)
	if err != nil {
		return nil, err
	}
	return toVerificationResponse(profile, request), nil
}

func (u *verificationUsecase) UpdateVerificationData(ctx context.Context, userID string, req *dto.UpdateVerificationRequest) (*dto.VerificationResponse, error) {
	profile, err := u.getSupplierProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	request, err := u.ensureRequest(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	if req.BusinessInfo != nil {
		request.LegalName = strings.TrimSpace(req.BusinessInfo.LegalName)
		request.NIBNumber = strings.TrimSpace(req.BusinessInfo.NIBNumber)
		request.EstablishedDate = strings.TrimSpace(req.BusinessInfo.EstablishedDate)
		request.Industry = strings.TrimSpace(req.BusinessInfo.Industry)
		request.BusinessPhone = strings.TrimSpace(req.BusinessInfo.Phone)
		request.BusinessDesc = strings.TrimSpace(req.BusinessInfo.Description)
		request.BusinessAddress = strings.TrimSpace(req.BusinessInfo.Address)
		request.NPWPNumber = strings.TrimSpace(req.BusinessInfo.NPWPNumber)
		request.NIBFileURL = strings.TrimSpace(req.BusinessInfo.NIBFileURL)
		request.NPWPFileURL = strings.TrimSpace(req.BusinessInfo.NPWPFileURL)
		request.AktaFileURL = strings.TrimSpace(req.BusinessInfo.AktaFileURL)
		syncProfileFromVerification(profile, req.BusinessInfo)
	}
	if req.StakeholderInfo != nil {
		request.DirectorName = strings.TrimSpace(req.StakeholderInfo.DirectorName)
		request.DirectorNIK = strings.TrimSpace(req.StakeholderInfo.DirectorNIK)
	}
	if req.BankAccountInfo != nil {
		request.BankName = strings.TrimSpace(req.BankAccountInfo.BankName)
		request.AccountNumber = strings.TrimSpace(req.BankAccountInfo.AccountNumber)
		request.AccountName = strings.TrimSpace(req.BankAccountInfo.AccountName)
	}
	if request.Status == "" {
		request.Status = "draft"
	}

	if err := u.repo.UpdateProfile(ctx, profile); err != nil {
		return nil, err
	}
	if err := u.repo.SaveVerificationRequest(ctx, request); err != nil {
		return nil, err
	}
	return toVerificationResponse(profile, request), nil
}

func isVerificationReady(data *dto.VerificationResponse) bool {
	business := data.BusinessInfo
	stakeholder := data.StakeholderInfo
	bank := data.BankAccountInfo

	required := []string{
		business.LegalName,
		business.NIBNumber,
		business.EstablishedDate,
		business.Industry,
		business.Phone,
		business.Address,
		business.NIBFileURL,
		business.NPWPNumber,
		business.NPWPFileURL,
		business.AktaFileURL,
		stakeholder.DirectorName,
		stakeholder.DirectorNIK,
		bank.BankName,
		bank.AccountNumber,
		bank.AccountName,
	}

	for _, field := range required {
		if strings.TrimSpace(field) == "" {
			return false
		}
	}

	return true
}

func (u *verificationUsecase) SubmitVerification(ctx context.Context, userID string) (*dto.VerificationResponse, error) {
	profile, err := u.getSupplierProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	request, err := u.getLatestRequest(ctx, profile.ID)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, ErrVerificationNotFound
	}

	response := toVerificationResponse(profile, request)
	if !isVerificationReady(response) {
		return nil, ErrVerificationIncomplete
	}

	now := apptime.Now()
	request.Status = "pending"
	request.SubmittedAt = &now
	if err := u.repo.SaveVerificationRequest(ctx, request); err != nil {
		return nil, err
	}

	return toVerificationResponse(profile, request), nil
}
