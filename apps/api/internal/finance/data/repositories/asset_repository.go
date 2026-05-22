package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AssetListParams struct {
	Search               string
	Status               *financeModels.AssetStatus
	CategoryID           *string
	TypeID               *string
	DepartmentID         *string
	LocationID           *string
	StartDate            *time.Time
	EndDate              *time.Time
	WarrantyExpiringDays *int
	IsCapitalized        *bool
	IncludeDisposed      bool
	Limit                int
	Offset               int
	SortBy               string
	SortDir              string
}

type AssetRepository interface {
	FindByID(ctx context.Context, id string, withDetails bool) (*financeModels.Asset, error)
	List(ctx context.Context, params AssetListParams) ([]financeModels.Asset, int64, error)
	FindLastDepreciation(ctx context.Context, assetID string) (*financeModels.AssetDepreciation, error)
	GenerateCode(ctx context.Context) (string, error)
	ExistsByCode(ctx context.Context, code string, excludeID *string) (bool, error)
	UpdateStatus(ctx context.Context, assetID string, status financeModels.AssetStatus) error
}

type assetRepository struct {
	db *gorm.DB
}

func NewAssetRepository(db *gorm.DB) AssetRepository {
	return &assetRepository{db: db}
}

func (r *assetRepository) FindByID(ctx context.Context, id string, withDetails bool) (*financeModels.Asset, error) {
	var item financeModels.Asset
	q := database.GetDB(ctx, r.db)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	q = q.Preload("Category").Preload("Location")
	if withDetails {
		q = q.Preload("Depreciations", func(db *gorm.DB) *gorm.DB {
			return db.Order("depreciation_date asc")
		}).Preload("Transactions", func(db *gorm.DB) *gorm.DB {
			return db.Order("transaction_date asc")
		})
	}
	if err := q.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

var assetAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "fixed_assets", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "fixed_assets", Name: "updated_at"},
	},
	"acquisition_date": {
		Column: clause.Column{Table: "fixed_assets", Name: "acquisition_date"},
	},
	"code": {
		Column: clause.Column{Table: "fixed_assets", Name: "code"},
	},
	"name": {
		Column: clause.Column{Table: "fixed_assets", Name: "name"},
	},
	"book_value": {
		Column: clause.Column{Table: "fixed_assets", Name: "book_value"},
	},
	"acquisition_cost": {
		Column: clause.Column{Table: "fixed_assets", Name: "acquisition_cost"},
	},
	"accumulated_depreciation": {
		Column: clause.Column{Table: "fixed_assets", Name: "accumulated_depreciation"},
	},
	"status": {
		Column: clause.Column{Table: "fixed_assets", Name: "status"},
	},
	"lifecycle_stage": {
		Column: clause.Column{Table: "fixed_assets", Name: "lifecycle_stage"},
	},
	"warranty_end": {
		Column: clause.Column{Table: "fixed_assets", Name: "warranty_end"},
	},
}

func (r *assetRepository) List(ctx context.Context, params AssetListParams) ([]financeModels.Asset, int64, error) {
	var items []financeModels.Asset
	var total int64

	q := database.GetDB(ctx, r.db).Model(&financeModels.Asset{}).
		Preload("AssignedToEmployee").
		Preload("Category").
		Preload("Location").
		Preload("Department")
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	// If the caller provided a department filter, apply it
	if params.DepartmentID != nil && strings.TrimSpace(*params.DepartmentID) != "" {
		q = q.Where("fixed_assets.department_id = ?", strings.TrimSpace(*params.DepartmentID))
	}

	// If the permission scope indicates DIVISION and a division/department context exists,
	// enforce department-level scoping to meet Department Head visibility requirement.
	if scope, _ := ctx.Value("permission_scope").(string); strings.ToUpper(scope) == "DIVISION" {
		if divID, _ := ctx.Value("scope_division_id").(string); strings.TrimSpace(divID) != "" {
			q = q.Where("fixed_assets.department_id = ?", strings.TrimSpace(divID))
		}
	}

	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("fixed_assets.code ILIKE ? OR fixed_assets.name ILIKE ?", like, like)
	}
	if params.Status != nil {
		q = q.Where("fixed_assets.status = ?", *params.Status)
	} else if !params.IncludeDisposed {
		q = q.Where("fixed_assets.status <> ?", financeModels.AssetStatusDisposed)
	}
	if params.CategoryID != nil && strings.TrimSpace(*params.CategoryID) != "" {
		q = q.Where("fixed_assets.category_id = ?", strings.TrimSpace(*params.CategoryID))
	}
	if params.TypeID != nil && strings.TrimSpace(*params.TypeID) != "" {
		q = q.Where("fixed_assets.asset_type_id = ?", strings.TrimSpace(*params.TypeID))
	}
	if params.LocationID != nil && strings.TrimSpace(*params.LocationID) != "" {
		q = q.Where("fixed_assets.location_id = ?", strings.TrimSpace(*params.LocationID))
	}
	if params.StartDate != nil {
		q = q.Where("fixed_assets.acquisition_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		q = q.Where("fixed_assets.acquisition_date <= ?", *params.EndDate)
	}
	if params.WarrantyExpiringDays != nil && *params.WarrantyExpiringDays > 0 {
		from := apptime.Now().Truncate(24 * time.Hour)
		to := from.AddDate(0, 0, *params.WarrantyExpiringDays)
		q = q.Where("fixed_assets.warranty_end IS NOT NULL AND fixed_assets.warranty_end BETWEEN ? AND ?", from, to)
	}
	if params.IsCapitalized != nil {
		q = q.Where("fixed_assets.is_capitalized = ?", *params.IsCapitalized)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := assetAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = assetAllowedSort["acquisition_date"]
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

func (r *assetRepository) FindLastDepreciation(ctx context.Context, assetID string) (*financeModels.AssetDepreciation, error) {
	var last financeModels.AssetDepreciation
	// ensure asset belongs to caller's scope
	if _, err := r.FindByID(ctx, assetID, false); err != nil {
		return nil, err
	}

	if err := database.GetDB(ctx, r.db).
		Where("asset_id = ?", assetID).
		Order("depreciation_date desc").
		First(&last).Error; err != nil {
		return nil, err
	}
	return &last, nil
}

func (r *assetRepository) GenerateCode(ctx context.Context) (string, error) {
	now := apptime.Now()
	prefix := "AST-" + now.Format("200601") + "-"

	var lastAsset financeModels.Asset
	q := database.GetDB(ctx, r.db).Unscoped().Model(&financeModels.Asset{}).
		Where("code LIKE ?", prefix+"%")
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	err := q.Order("code DESC").First(&lastAsset).Error

	nextNum := 1
	if err == nil {
		parts := strings.Split(lastAsset.Code, "-")
		if len(parts) == 3 {
			var lastNum int
			fmt.Sscanf(parts[2], "%d", &lastNum)
			nextNum = lastNum + 1
		}
	}

	return fmt.Sprintf("%s%04d", prefix, nextNum), nil
}

func (r *assetRepository) ExistsByCode(ctx context.Context, code string, excludeID *string) (bool, error) {
	q := database.GetDB(ctx, r.db).Model(&financeModels.Asset{}).Where("code = ?", strings.TrimSpace(code))
	if excludeID != nil && strings.TrimSpace(*excludeID) != "" {
		q = q.Where("id <> ?", strings.TrimSpace(*excludeID))
	}
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *assetRepository) UpdateStatus(ctx context.Context, assetID string, status financeModels.AssetStatus) error {
	q := database.GetDB(ctx, r.db).Model(&financeModels.Asset{}).
		Where("id = ?", assetID)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	return q.Update("status", status).Error
}
