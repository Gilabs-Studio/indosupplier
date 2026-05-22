package usecase

import (
	"context"
	"errors"
	"fmt"

	securityInfra "github.com/gilabs/gims/api/internal/core/infrastructure/security"
	posModels "github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/mapper"
	"github.com/gilabs/gims/api/internal/pos/domain/provider"
	"gorm.io/gorm"
)

var ErrXenditConfigNotFound = errors.New("xendit config not found")
var ErrXenditConnectionValidation = errors.New("xendit connection validation failed")

// XenditConfigUsecase manages per-company Xendit payment gateway settings
type XenditConfigUsecase interface {
	Get(ctx context.Context, companyID string) (*dto.XenditConfigResponse, error)
	GetConnectionStatus(ctx context.Context, companyID string) (*dto.XenditConnectionStatusResponse, error)
	TestConnection(ctx context.Context, req *dto.TestXenditConnectionRequest) (*dto.TestXenditConnectionResponse, error)
	Connect(ctx context.Context, companyID string, req *dto.ConnectXenditRequest, updatedBy string) (*dto.XenditConfigResponse, error)
	Update(ctx context.Context, companyID string, req *dto.UpdateXenditConfigRequest, updatedBy string) (*dto.XenditConfigResponse, error)
	Disconnect(ctx context.Context, companyID string, updatedBy string) (*dto.XenditConfigResponse, error)
}

type xenditConfigUsecase struct {
	repo   repositories.XenditConfigRepository
	cipher *securityInfra.CredentialCipher
}

// NewXenditConfigUsecase creates the usecase
func NewXenditConfigUsecase(repo repositories.XenditConfigRepository, cipher *securityInfra.CredentialCipher) XenditConfigUsecase {
	return &xenditConfigUsecase{repo: repo, cipher: cipher}
}

func (u *xenditConfigUsecase) Get(ctx context.Context, companyID string) (*dto.XenditConfigResponse, error) {
	cfg, err := u.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrXenditConfigNotFound
		}
		return nil, err
	}
	return mapper.ToXenditConfigResponse(cfg), nil
}

func (u *xenditConfigUsecase) GetConnectionStatus(ctx context.Context, companyID string) (*dto.XenditConnectionStatusResponse, error) {
	cfg, err := u.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.XenditConnectionStatusResponse{
				IsConnected: false,
				Status:      string(posModels.XenditStatusNotConnected),
			}, nil
		}
		return nil, err
	}
	return &dto.XenditConnectionStatusResponse{
		IsConnected: cfg.IsConnected(),
		Status:      string(cfg.ConnectionStatus),
	}, nil
}

func (u *xenditConfigUsecase) TestConnection(ctx context.Context, req *dto.TestXenditConnectionRequest) (*dto.TestXenditConnectionResponse, error) {
	tester := provider.NewXenditProvider(req.SecretKey, req.XenditAccountID)
	if err := tester.TestConnection(ctx); err != nil {
		return &dto.TestXenditConnectionResponse{
			Reachable: false,
			Message:   err.Error(),
		}, errors.Join(ErrXenditConnectionValidation, err)
	}

	return &dto.TestXenditConnectionResponse{
		Reachable: true,
		Message:   "Connection successful",
	}, nil
}

func (u *xenditConfigUsecase) Connect(ctx context.Context, companyID string, req *dto.ConnectXenditRequest, updatedBy string) (*dto.XenditConfigResponse, error) {
	encryptedSecret, err := u.cipher.Encrypt(req.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt xendit secret key: %w", err)
	}

	encryptedWebhookToken, err := u.cipher.Encrypt(req.WebhookToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt xendit webhook token: %w", err)
	}

	testResp, err := u.TestConnection(ctx, &dto.TestXenditConnectionRequest{
		SecretKey:       req.SecretKey,
		XenditAccountID: req.XenditAccountID,
	})
	if err != nil || !testResp.Reachable {
		if testResp != nil && testResp.Message != "" {
			return nil, fmt.Errorf("%w: %s", ErrXenditConnectionValidation, testResp.Message)
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrXenditConnectionValidation, err)
		}
		return nil, ErrXenditConnectionValidation
	}

	cfg := &posModels.XenditConfig{
		CompanyID:        companyID,
		SecretKey:        encryptedSecret,
		XenditAccountID:  req.XenditAccountID,
		BusinessName:     req.BusinessName,
		Environment:      posModels.XenditEnvironment(req.Environment),
		WebhookToken:     encryptedWebhookToken,
		ConnectionStatus: posModels.XenditStatusConnected,
		IsActive:         true,
		UpdatedBy:        &updatedBy,
	}

	if err := u.repo.Upsert(ctx, cfg); err != nil {
		return nil, err
	}

	saved, err := u.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return mapper.ToXenditConfigResponse(saved), nil
}

func (u *xenditConfigUsecase) Update(ctx context.Context, companyID string, req *dto.UpdateXenditConfigRequest, updatedBy string) (*dto.XenditConfigResponse, error) {
	existing, err := u.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrXenditConfigNotFound
		}
		return nil, err
	}

	if req.Environment != "" {
		existing.Environment = posModels.XenditEnvironment(req.Environment)
	}
	if req.BusinessName != "" {
		existing.BusinessName = req.BusinessName
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	existing.UpdatedBy = &updatedBy

	if err := u.repo.Upsert(ctx, existing); err != nil {
		return nil, err
	}

	saved, err := u.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return mapper.ToXenditConfigResponse(saved), nil
}

func (u *xenditConfigUsecase) Disconnect(ctx context.Context, companyID string, updatedBy string) (*dto.XenditConfigResponse, error) {
	existing, err := u.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrXenditConfigNotFound
		}
		return nil, err
	}

	existing.ConnectionStatus = posModels.XenditStatusNotConnected
	existing.IsActive = false
	existing.SecretKey = ""
	existing.WebhookToken = ""
	existing.UpdatedBy = &updatedBy

	if err := u.repo.Upsert(ctx, existing); err != nil {
		return nil, err
	}

	saved, err := u.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return mapper.ToXenditConfigResponse(saved), nil
}
