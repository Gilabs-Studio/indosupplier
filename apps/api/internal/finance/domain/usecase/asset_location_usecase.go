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
	ErrAssetLocationNotFound = errors.New("asset location not found")
)

type AssetLocationUsecase interface {
	Create(ctx context.Context, req *dto.CreateAssetLocationRequest) (*dto.AssetLocationResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateAssetLocationRequest) (*dto.AssetLocationResponse, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.AssetLocationResponse, error)
	List(ctx context.Context, req *dto.ListAssetLocationsRequest) ([]dto.AssetLocationResponse, int64, error)
}

type assetLocationUsecase struct {
	db     *gorm.DB
	repo   repositories.AssetLocationRepository
	mapper *mapper.AssetLocationMapper
}

func NewAssetLocationUsecase(db *gorm.DB, repo repositories.AssetLocationRepository, mapper *mapper.AssetLocationMapper) AssetLocationUsecase {
	return &assetLocationUsecase{db: db, repo: repo, mapper: mapper}
}

func (uc *assetLocationUsecase) Create(ctx context.Context, req *dto.CreateAssetLocationRequest) (*dto.AssetLocationResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return nil, errors.New("user not authenticated")
	}

	item := &financeModels.AssetLocation{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
	if err := uc.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}

	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *assetLocationUsecase) Update(ctx context.Context, id string, req *dto.UpdateAssetLocationRequest) (*dto.AssetLocationResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}

	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAssetLocationNotFound
		}
		return nil, err
	}

	if err := uc.db.WithContext(ctx).Model(&financeModels.AssetLocation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        strings.TrimSpace(req.Name),
		"description": strings.TrimSpace(req.Description),
	}).Error; err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := uc.mapper.ToResponse(full)
	return &res, nil
}

func (uc *assetLocationUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAssetLocationNotFound
		}
		return err
	}
	return uc.db.WithContext(ctx).Delete(&financeModels.AssetLocation{}, "id = ?", id).Error
}

func (uc *assetLocationUsecase) GetByID(ctx context.Context, id string) (*dto.AssetLocationResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAssetLocationNotFound
		}
		return nil, err
	}
	res := uc.mapper.ToResponse(item)
	return &res, nil
}

func (uc *assetLocationUsecase) List(ctx context.Context, req *dto.ListAssetLocationsRequest) ([]dto.AssetLocationResponse, int64, error) {
	if req == nil {
		req = &dto.ListAssetLocationsRequest{}
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

	items, total, err := uc.repo.List(ctx, repositories.AssetLocationListParams{
		Search:  req.Search,
		SortBy:  req.SortBy,
		SortDir: req.SortDir,
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	res := make([]dto.AssetLocationResponse, 0, len(items))
	for i := range items {
		mapped := uc.mapper.ToResponse(&items[i])
		res = append(res, mapped)
	}
	return res, total, nil
}
