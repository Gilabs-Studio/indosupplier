package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SystemAccountMappingRepository interface {
	GetByKey(ctx context.Context, key string, companyID *string) (string, error)
	GetMappingByKey(ctx context.Context, key string, companyID *string) (*financeModels.SystemAccountMapping, error)
	GetExactMappingByKey(ctx context.Context, key string, companyID *string) (*financeModels.SystemAccountMapping, error)
	ListMappings(ctx context.Context, companyID *string) ([]financeModels.SystemAccountMapping, error)
	DeleteByKey(ctx context.Context, key string, companyID *string) error
	Upsert(ctx context.Context, mapping *financeModels.SystemAccountMapping) error
}

type systemAccountMappingRepository struct {
	db *gorm.DB
}

func NewSystemAccountMappingRepository(db *gorm.DB) SystemAccountMappingRepository {
	return &systemAccountMappingRepository{db: db}
}

func (r *systemAccountMappingRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *systemAccountMappingRepository) GetByKey(ctx context.Context, key string, companyID *string) (string, error) {
	m, err := r.GetMappingByKey(ctx, key, companyID)
	if err != nil {
		return "", err
	}

	return m.COACode, nil
}

func (r *systemAccountMappingRepository) GetMappingByKey(ctx context.Context, key string, companyID *string) (*financeModels.SystemAccountMapping, error) {
	var m financeModels.SystemAccountMapping
	db := r.getDB(ctx).Where("key = ?", key)

	if normalizedCompanyID, ok := normalizeMappingCompanyID(companyID); ok {
		if err := db.
			Where("company_id = ? OR company_id IS NULL", normalizedCompanyID).
			Order(clause.Expr{SQL: "CASE WHEN company_id = ? THEN 0 ELSE 1 END", Vars: []interface{}{normalizedCompanyID}}).
			First(&m).Error; err != nil {
			return nil, err
		}
		return &m, nil
	}

	if err := db.Where("company_id IS NULL").First(&m).Error; err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *systemAccountMappingRepository) GetExactMappingByKey(ctx context.Context, key string, companyID *string) (*financeModels.SystemAccountMapping, error) {
	var m financeModels.SystemAccountMapping
	query := r.getDB(ctx).Where("key = ?", key)

	if normalizedCompanyID, ok := normalizeMappingCompanyID(companyID); ok {
		query = query.Where("company_id = ?", normalizedCompanyID)
	} else {
		query = query.Where("company_id IS NULL")
	}

	if err := query.First(&m).Error; err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *systemAccountMappingRepository) ListMappings(ctx context.Context, companyID *string) ([]financeModels.SystemAccountMapping, error) {
	rows := make([]financeModels.SystemAccountMapping, 0)
	query := r.getDB(ctx).Model(&financeModels.SystemAccountMapping{})

	if normalizedCompanyID, ok := normalizeMappingCompanyID(companyID); ok {
		if err := query.
			Where("company_id = ? OR company_id IS NULL", normalizedCompanyID).
			Order("key ASC").
			Order(clause.Expr{SQL: "CASE WHEN company_id = ? THEN 0 ELSE 1 END", Vars: []interface{}{normalizedCompanyID}}).
			Find(&rows).Error; err != nil {
			return nil, err
		}

		// Collapse global/company rows by key while keeping company override first.
		result := make([]financeModels.SystemAccountMapping, 0, len(rows))
		seen := make(map[string]struct{}, len(rows))
		for _, row := range rows {
			if _, ok := seen[row.Key]; ok {
				continue
			}
			seen[row.Key] = struct{}{}
			result = append(result, row)
		}

		return result, nil
	}

	if err := query.
		Where("company_id IS NULL").
		Order("key ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *systemAccountMappingRepository) DeleteByKey(ctx context.Context, key string, companyID *string) error {
	query := r.getDB(ctx).Where("key = ?", key)
	if normalizedCompanyID, ok := normalizeMappingCompanyID(companyID); ok {
		query = query.Where("company_id = ?", normalizedCompanyID)
	} else {
		query = query.Where("company_id IS NULL")
	}

	result := query.Delete(&financeModels.SystemAccountMapping{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *systemAccountMappingRepository) Upsert(ctx context.Context, mapping *financeModels.SystemAccountMapping) error {
	var existing financeModels.SystemAccountMapping
	query := r.getDB(ctx).Where("key = ?", mapping.Key)
	if normalizedCompanyID, ok := normalizeMappingCompanyID(mapping.CompanyID); ok {
		query = query.Where("company_id = ?", normalizedCompanyID)
		mapping.CompanyID = &normalizedCompanyID
	} else {
		query = query.Where("company_id IS NULL")
		mapping.CompanyID = nil
	}

	err := query.First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.getDB(ctx).Create(mapping).Error
	} else if err != nil {
		return err
	}

	existing.COACode = mapping.COACode
	existing.Label = mapping.Label
	return r.getDB(ctx).Save(&existing).Error
}

func normalizeMappingCompanyID(companyID *string) (string, bool) {
	if companyID == nil {
		return "", false
	}

	trimmed := strings.TrimSpace(*companyID)
	if trimmed == "" {
		return "", false
	}

	return trimmed, true
}
