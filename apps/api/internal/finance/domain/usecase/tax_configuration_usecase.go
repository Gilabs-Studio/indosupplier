package usecase

import (
	"context"
	"errors"
	"strings"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"gorm.io/gorm"
)

var (
	ErrTaxConfigurationNotFound        = errors.New("tax configuration not found")
	ErrTaxConfigurationCodeExists      = errors.New("tax code already exists in this company")
	ErrTaxConfigurationCompanyMismatch = errors.New("tax configuration does not belong to the provided company")
	ErrTaxConfigurationAccountInvalid  = errors.New("tax configuration account must exist, be active, and postable")
)

type TaxConfigurationUsecase interface {
	Create(ctx context.Context, req *dto.CreateTaxConfigurationRequest) (*dto.TaxConfigurationResponse, error)
	List(ctx context.Context, req *dto.ListTaxConfigurationsRequest) ([]dto.TaxConfigurationResponse, int64, error)
	GetByID(ctx context.Context, id, companyID string) (*dto.TaxConfigurationResponse, error)
	Update(ctx context.Context, id, companyID string, req *dto.UpdateTaxConfigurationRequest) (*dto.TaxConfigurationResponse, error)
	ToggleStatus(ctx context.Context, id string, req *dto.ToggleTaxConfigurationStatusRequest) (*dto.TaxConfigurationResponse, error)
}

type taxConfigurationUsecase struct {
	repo    repositories.TaxConfigurationRepository
	coaRepo repositories.ChartOfAccountRepository
}

func NewTaxConfigurationUsecase(repo repositories.TaxConfigurationRepository, coaRepo repositories.ChartOfAccountRepository) TaxConfigurationUsecase {
	return &taxConfigurationUsecase{repo: repo, coaRepo: coaRepo}
}

func (uc *taxConfigurationUsecase) validateAccount(ctx context.Context, accountID string) error {
	if uc.coaRepo == nil {
		return nil
	}

	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return ErrTaxConfigurationAccountInvalid
	}

	coa, err := uc.coaRepo.FindByID(ctx, accountID)
	if err != nil {
		return ErrTaxConfigurationAccountInvalid
	}
	if !coa.IsActive || !coa.IsPostable {
		return ErrTaxConfigurationAccountInvalid
	}
	return nil
}

func toTaxConfigurationResponse(item *financeModels.TaxConfiguration) *dto.TaxConfigurationResponse {
	return &dto.TaxConfigurationResponse{
		ID:          item.ID,
		CompanyID:   item.CompanyID,
		TaxCode:     item.TaxCode,
		TaxName:     item.TaxName,
		TaxType:     item.TaxType,
		Rate:        item.Rate,
		IsInclusive: item.IsInclusive,
		AccountID:   item.AccountID,
		IsActive:    item.IsActive,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func (uc *taxConfigurationUsecase) Create(ctx context.Context, req *dto.CreateTaxConfigurationRequest) (*dto.TaxConfigurationResponse, error) {
	taxCode := strings.ToUpper(strings.TrimSpace(req.TaxCode))
	if err := uc.validateAccount(ctx, req.AccountID); err != nil {
		return nil, err
	}
	exists, err := uc.repo.ExistsByTaxCode(ctx, req.CompanyID, taxCode, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTaxConfigurationCodeExists
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	item := &financeModels.TaxConfiguration{
		CompanyID:   req.CompanyID,
		TaxCode:     taxCode,
		TaxName:     strings.TrimSpace(req.TaxName),
		TaxType:     req.TaxType,
		Rate:        req.Rate,
		IsInclusive: req.IsInclusive,
		AccountID:   req.AccountID,
		IsActive:    isActive,
	}
	if err := uc.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return toTaxConfigurationResponse(item), nil
}

func (uc *taxConfigurationUsecase) List(ctx context.Context, req *dto.ListTaxConfigurationsRequest) ([]dto.TaxConfigurationResponse, int64, error) {
	items, total, err := uc.repo.List(ctx, repositories.TaxConfigurationListParams{
		CompanyID: req.CompanyID,
		TaxType:   req.TaxType,
		IsActive:  req.IsActive,
		Page:      req.Page,
		PerPage:   req.PerPage,
	})
	if err != nil {
		return nil, 0, err
	}
	res := make([]dto.TaxConfigurationResponse, 0, len(items))
	for i := range items {
		mapped := toTaxConfigurationResponse(&items[i])
		res = append(res, *mapped)
	}
	return res, total, nil
}

func (uc *taxConfigurationUsecase) GetByID(ctx context.Context, id, companyID string) (*dto.TaxConfigurationResponse, error) {
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaxConfigurationNotFound
		}
		return nil, err
	}
	if item.CompanyID != companyID {
		return nil, ErrTaxConfigurationCompanyMismatch
	}
	return toTaxConfigurationResponse(item), nil
}

func (uc *taxConfigurationUsecase) Update(ctx context.Context, id, companyID string, req *dto.UpdateTaxConfigurationRequest) (*dto.TaxConfigurationResponse, error) {
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaxConfigurationNotFound
		}
		return nil, err
	}
	if item.CompanyID != companyID {
		return nil, ErrTaxConfigurationCompanyMismatch
	}
	if err := uc.validateAccount(ctx, req.AccountID); err != nil {
		return nil, err
	}

	taxCode := strings.ToUpper(strings.TrimSpace(req.TaxCode))
	exists, err := uc.repo.ExistsByTaxCode(ctx, companyID, taxCode, &item.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTaxConfigurationCodeExists
	}

	item.TaxCode = taxCode
	item.TaxName = strings.TrimSpace(req.TaxName)
	item.TaxType = req.TaxType
	item.Rate = req.Rate
	item.IsInclusive = req.IsInclusive
	item.AccountID = req.AccountID
	if err := uc.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return toTaxConfigurationResponse(item), nil
}

func (uc *taxConfigurationUsecase) ToggleStatus(ctx context.Context, id string, req *dto.ToggleTaxConfigurationStatusRequest) (*dto.TaxConfigurationResponse, error) {
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaxConfigurationNotFound
		}
		return nil, err
	}
	if item.CompanyID != req.CompanyID {
		return nil, ErrTaxConfigurationCompanyMismatch
	}
	item.IsActive = req.IsActive
	if err := uc.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return toTaxConfigurationResponse(item), nil
}
