package usecase

import (
	"context"
	"errors"
	"strings"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrAssetCategoryNotFound = errors.New("asset category not found")
)

type AssetCategoryUsecase interface {
	Create(ctx context.Context, req *dto.CreateAssetCategoryRequest) (*dto.AssetCategoryResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateAssetCategoryRequest) (*dto.AssetCategoryResponse, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.AssetCategoryResponse, error)
	List(ctx context.Context, req *dto.ListAssetCategoriesRequest) ([]dto.AssetCategoryResponse, int64, error)
	GetFormData(ctx context.Context) (*dto.AssetCategoryFormDataResponse, error)
}

type assetCategoryUsecase struct {
	db      *gorm.DB
	coaRepo repositories.ChartOfAccountRepository
	repo    repositories.AssetCategoryRepository
	mapper  *mapper.AssetCategoryMapper
}

func NewAssetCategoryUsecase(db *gorm.DB, coaRepo repositories.ChartOfAccountRepository, repo repositories.AssetCategoryRepository, mapper *mapper.AssetCategoryMapper) AssetCategoryUsecase {
	return &assetCategoryUsecase{db: db, coaRepo: coaRepo, repo: repo, mapper: mapper}
}

func (uc *assetCategoryUsecase) Create(ctx context.Context, req *dto.CreateAssetCategoryRequest) (*dto.AssetCategoryResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	if _, err := uc.coaRepo.FindByID(ctx, req.AssetAccountID); err != nil {
		return nil, err
	}
	if _, err := uc.coaRepo.FindByID(ctx, req.AccumulatedDepreciationAccountID); err != nil {
		return nil, err
	}
	if _, err := uc.coaRepo.FindByID(ctx, req.DepreciationExpenseAccountID); err != nil {
		return nil, err
	}
	if req.DisposalGainAccountID != nil && strings.TrimSpace(*req.DisposalGainAccountID) != "" {
		if _, err := uc.coaRepo.FindByID(ctx, strings.TrimSpace(*req.DisposalGainAccountID)); err != nil {
			return nil, err
		}
	}
	if req.DisposalLossAccountID != nil && strings.TrimSpace(*req.DisposalLossAccountID) != "" {
		if _, err := uc.coaRepo.FindByID(ctx, strings.TrimSpace(*req.DisposalLossAccountID)); err != nil {
			return nil, err
		}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	item := &financeModels.AssetCategory{
		Name:                             strings.TrimSpace(req.Name),
		Type:                             req.Type,
		DepreciationMethod:               req.DepreciationMethod,
		UsefulLifeMonths:                 req.UsefulLifeMonths,
		DepreciationRate:                 req.DepreciationRate,
		AssetAccountID:                   strings.TrimSpace(req.AssetAccountID),
		AccumulatedDepreciationAccountID: strings.TrimSpace(req.AccumulatedDepreciationAccountID),
		DepreciationExpenseAccountID:     strings.TrimSpace(req.DepreciationExpenseAccountID),
		DisposalGainAccountID:            req.DisposalGainAccountID,
		DisposalLossAccountID:            req.DisposalLossAccountID,
		IsActive:                         isActive,
	}
	if err := uc.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}

	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *assetCategoryUsecase) Update(ctx context.Context, id string, req *dto.UpdateAssetCategoryRequest) (*dto.AssetCategoryResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}

	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAssetCategoryNotFound
		}
		return nil, err
	}

	if _, err := uc.coaRepo.FindByID(ctx, req.AssetAccountID); err != nil {
		return nil, err
	}
	if _, err := uc.coaRepo.FindByID(ctx, req.AccumulatedDepreciationAccountID); err != nil {
		return nil, err
	}
	if _, err := uc.coaRepo.FindByID(ctx, req.DepreciationExpenseAccountID); err != nil {
		return nil, err
	}
	if req.DisposalGainAccountID != nil && strings.TrimSpace(*req.DisposalGainAccountID) != "" {
		if _, err := uc.coaRepo.FindByID(ctx, strings.TrimSpace(*req.DisposalGainAccountID)); err != nil {
			return nil, err
		}
	}
	if req.DisposalLossAccountID != nil && strings.TrimSpace(*req.DisposalLossAccountID) != "" {
		if _, err := uc.coaRepo.FindByID(ctx, strings.TrimSpace(*req.DisposalLossAccountID)); err != nil {
			return nil, err
		}
	}

	updates := map[string]interface{}{
		"name":                                strings.TrimSpace(req.Name),
		"type":                                req.Type,
		"depreciation_method":                 req.DepreciationMethod,
		"useful_life_months":                  req.UsefulLifeMonths,
		"depreciation_rate":                   req.DepreciationRate,
		"asset_account_id":                    strings.TrimSpace(req.AssetAccountID),
		"accumulated_depreciation_account_id": strings.TrimSpace(req.AccumulatedDepreciationAccountID),
		"depreciation_expense_account_id":     strings.TrimSpace(req.DepreciationExpenseAccountID),
		"disposal_gain_account_id":            req.DisposalGainAccountID,
		"disposal_loss_account_id":            req.DisposalLossAccountID,
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := uc.db.WithContext(ctx).Model(&financeModels.AssetCategory{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := uc.mapper.ToResponse(full)
	return &res, nil
}

func (uc *assetCategoryUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAssetCategoryNotFound
		}
		return err
	}
	return uc.db.WithContext(ctx).Delete(&financeModels.AssetCategory{}, "id = ?", id).Error
}

func (uc *assetCategoryUsecase) GetByID(ctx context.Context, id string) (*dto.AssetCategoryResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAssetCategoryNotFound
		}
		return nil, err
	}
	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *assetCategoryUsecase) List(ctx context.Context, req *dto.ListAssetCategoriesRequest) ([]dto.AssetCategoryResponse, int64, error) {
	if req == nil {
		req = &dto.ListAssetCategoriesRequest{}
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	items, total, err := uc.repo.List(ctx, repositories.AssetCategoryListParams{
		Search:  req.Search,
		SortBy:  req.SortBy,
		SortDir: req.SortDir,
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	res := make([]dto.AssetCategoryResponse, 0, len(items))
	for i := range items {
		mapped := uc.mapper.ToResponse(&items[i])
		res = append(res, mapped)
	}
	return res, total, nil
}
