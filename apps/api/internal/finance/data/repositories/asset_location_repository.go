package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"strings"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AssetLocationListParams struct {
	Search  string
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
}

type AssetLocationRepository interface {
	FindByID(ctx context.Context, id string) (*financeModels.AssetLocation, error)
	List(ctx context.Context, params AssetLocationListParams) ([]financeModels.AssetLocation, int64, error)
}

type assetLocationRepository struct {
	db *gorm.DB
}

func NewAssetLocationRepository(db *gorm.DB) AssetLocationRepository {
	return &assetLocationRepository{db: db}
}

func (r *assetLocationRepository) FindByID(ctx context.Context, id string) (*financeModels.AssetLocation, error) {
	var item financeModels.AssetLocation
	if err := database.GetDB(ctx, r.db).First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var assetLocationAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "asset_locations", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "asset_locations", Name: "updated_at"},
	},
	"name": {
		Column: clause.Column{Table: "asset_locations", Name: "name"},
	},
}

func (r *assetLocationRepository) List(ctx context.Context, params AssetLocationListParams) ([]financeModels.AssetLocation, int64, error) {
	var items []financeModels.AssetLocation
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.AssetLocation{})
	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("asset_locations.name ILIKE ? OR asset_locations.description ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := assetLocationAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = assetLocationAllowedSort["created_at"]
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
