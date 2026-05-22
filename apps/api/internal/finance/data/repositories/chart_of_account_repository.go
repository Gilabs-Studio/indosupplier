package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type chartOfAccountRepository struct {
	db *gorm.DB
}

func NewChartOfAccountRepository(db *gorm.DB) ChartOfAccountRepository {
	return &chartOfAccountRepository{db: db}
}

type ChartOfAccountListParams struct {
	Search   string
	Type     *financeModels.AccountType
	ParentID *string
	IsActive *bool
	SortBy   string
	SortDir  string
	Limit    int
	Offset   int
}

type ChartOfAccountRepository interface {
	Create(ctx context.Context, item *financeModels.ChartOfAccount) error
	FindByID(ctx context.Context, id string) (*financeModels.ChartOfAccount, error)
	FindAll(ctx context.Context, onlyActive bool) ([]financeModels.ChartOfAccount, error)
	List(ctx context.Context, params ChartOfAccountListParams) ([]financeModels.ChartOfAccount, int64, error)
	Update(ctx context.Context, item *financeModels.ChartOfAccount) error
	Delete(ctx context.Context, id string) error
	ExistsByCode(ctx context.Context, code string, excludeID *string) (bool, error)
	FindByCode(ctx context.Context, code string) (*financeModels.ChartOfAccount, error)
	GetByCode(ctx context.Context, code string) (*financeModels.ChartOfAccount, error)
	FindOpeningBalanceEquity(ctx context.Context) (*financeModels.ChartOfAccount, error)
	HasChildren(ctx context.Context, id string) (bool, error)
	IsUsedInJournal(ctx context.Context, id string) (bool, error)
	HasJournalLines(ctx context.Context, id string) (bool, error)
	UpdateIsPostable(ctx context.Context, id string, isPostable bool) error
	RecalculateAllIsPostable(ctx context.Context) error
	GetDB(ctx context.Context) *gorm.DB
}

func (r *chartOfAccountRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *chartOfAccountRepository) Create(ctx context.Context, item *financeModels.ChartOfAccount) error {
	return r.getDB(ctx).Create(item).Error
}

func (r *chartOfAccountRepository) FindByID(ctx context.Context, id string) (*financeModels.ChartOfAccount, error) {
	var item financeModels.ChartOfAccount
	q := r.getDB(ctx).Model(&financeModels.ChartOfAccount{})
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if err := q.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *chartOfAccountRepository) FindAll(ctx context.Context, onlyActive bool) ([]financeModels.ChartOfAccount, error) {
	var items []financeModels.ChartOfAccount
	q := r.getDB(ctx).Model(&financeModels.ChartOfAccount{})
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if onlyActive {
		q = q.Where("is_active = ?", true)
	}
	if err := q.Order("code asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

var coaAllowedSort = map[string]clause.OrderByColumn{
	"created_at": {
		Column: clause.Column{Table: "chart_of_accounts", Name: "created_at"},
	},
	"updated_at": {
		Column: clause.Column{Table: "chart_of_accounts", Name: "updated_at"},
	},
	"code": {
		Column: clause.Column{Table: "chart_of_accounts", Name: "code"},
	},
	"name": {
		Column: clause.Column{Table: "chart_of_accounts", Name: "name"},
	},
	"type": {
		Column: clause.Column{Table: "chart_of_accounts", Name: "type"},
	},
}

func (r *chartOfAccountRepository) List(ctx context.Context, params ChartOfAccountListParams) ([]financeModels.ChartOfAccount, int64, error) {
	var items []financeModels.ChartOfAccount
	var total int64

	q := r.getDB(ctx).Model(&financeModels.ChartOfAccount{})
	// Apply tenant + permission scope filtering for finance resources
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())

	if s := strings.TrimSpace(params.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("chart_of_accounts.code ILIKE ? OR chart_of_accounts.name ILIKE ?", like, like)
	}
	if params.Type != nil {
		q = q.Where("chart_of_accounts.type = ?", *params.Type)
	}
	if params.ParentID != nil {
		q = q.Where("chart_of_accounts.parent_id = ?", *params.ParentID)
	}
	if params.IsActive != nil {
		q = q.Where("chart_of_accounts.is_active = ?", *params.IsActive)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortCol := coaAllowedSort[params.SortBy]
	if sortCol.Column.Name == "" {
		sortCol = coaAllowedSort["code"]
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

func (r *chartOfAccountRepository) Update(ctx context.Context, item *financeModels.ChartOfAccount) error {
	return r.getDB(ctx).Save(item).Error
}

func (r *chartOfAccountRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&financeModels.ChartOfAccount{}, "id = ?", id).Error
}

func (r *chartOfAccountRepository) ExistsByCode(ctx context.Context, code string, excludeID *string) (bool, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return false, nil
	}

	q := r.getDB(ctx).Model(&financeModels.ChartOfAccount{}).Where("code = ?", code)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if excludeID != nil && strings.TrimSpace(*excludeID) != "" {
		q = q.Where("id <> ?", strings.TrimSpace(*excludeID))
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *chartOfAccountRepository) FindByCode(ctx context.Context, code string) (*financeModels.ChartOfAccount, error) {
	var item financeModels.ChartOfAccount
	q := r.getDB(ctx).Model(&financeModels.ChartOfAccount{}).Where("code = ?", code)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	if err := q.First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *chartOfAccountRepository) GetByCode(ctx context.Context, code string) (*financeModels.ChartOfAccount, error) {
	return r.FindByCode(ctx, code)
}

func (r *chartOfAccountRepository) FindOpeningBalanceEquity(ctx context.Context) (*financeModels.ChartOfAccount, error) {
	return r.FindByCode(ctx, "39999")
}

func (r *chartOfAccountRepository) HasChildren(ctx context.Context, id string) (bool, error) {
	var count int64
	q := r.getDB(ctx).Model(&financeModels.ChartOfAccount{}).Where("parent_id = ?", id)
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	err := q.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *chartOfAccountRepository) HasJournalLines(ctx context.Context, id string) (bool, error) {
	return r.IsUsedInJournal(ctx, id)
}

func (r *chartOfAccountRepository) IsUsedInJournal(ctx context.Context, id string) (bool, error) {
	var count int64
	q := r.getDB(ctx).Table("journal_lines").Where("chart_of_account_id = ?", id).Where("deleted_at IS NULL")
	q = security.ApplyScopeFilter(q, ctx, security.FinanceScopeQueryOptions())
	err := q.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *chartOfAccountRepository) UpdateIsPostable(ctx context.Context, id string, isPostable bool) error {
	return r.getDB(ctx).
		Model(&financeModels.ChartOfAccount{}).
		Where("id = ?", id).
		Update("is_postable", isPostable).Error
}

func (r *chartOfAccountRepository) RecalculateAllIsPostable(ctx context.Context) error {
	return r.getDB(ctx).Exec(`
		UPDATE chart_of_accounts AS coa
		SET is_postable = CASE
			WHEN coa.parent_id IS NULL THEN FALSE
			WHEN EXISTS (
				SELECT 1
				FROM chart_of_accounts AS child
				WHERE child.parent_id = coa.id
				  AND child.deleted_at IS NULL
			) THEN FALSE
			ELSE TRUE
		END
		WHERE coa.deleted_at IS NULL
	`).Error
}

func (r *chartOfAccountRepository) GetDB(ctx context.Context) *gorm.DB {
	return r.getDB(ctx)
}
