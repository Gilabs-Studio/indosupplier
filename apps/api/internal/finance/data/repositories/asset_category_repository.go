package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"strings"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AssetCategoryListParams struct {
	Search  string
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
}

type AssetCategoryRepository interface {
	FindByID(ctx context.Context, id string) (*financeModels.AssetCategory, error)
	List(ctx context.Context, params AssetCategoryListParams) ([]financeModels.AssetCategory, int64, error)
}

type assetCategoryRepository struct {
	db *gorm.DB
}

func NewAssetCategoryRepository(db *gorm.DB) AssetCategoryRepository {
	return &assetCategoryRepository{db: db}
}

func (r *assetCategoryRepository) FindByID(ctx context.Context, id string) (*financeModels.AssetCategory, error) {
	var item financeModels.AssetCategory
	if err := database.GetDB(ctx, r.db).First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var assetCategoryAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "asset_categories", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "asset_categories", Name: "updated_at"},
	},
	"name": {
		Column: clause.Column{Table: "asset_categories", Name: "name"},
	},
}

func (r *assetCategoryRepository) List(ctx context.Context, params AssetCategoryListParams) ([]financeModels.AssetCategory, int64, error) {
	var items []financeModels.AssetCategory
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.AssetCategory{})
	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("asset_categories.name ILIKE ?", like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := assetCategoryAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = assetCategoryAllowedSort["created_at"]
	}
	sortDir := strings.ToLower(strings.TrimSpace(params.SortDir))
	if sortDir == "asc" {
		sortCol.Desc = false
	} else {
		sortCol.Desc = true
	}
	q = q.Order(sortCol)

	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}

	if err := q.Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
