package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

type ListJournalTemplateParams struct {
	CompanyID string
	Search    string
	Limit     int
	Offset    int
}

type JournalTemplateRepository interface {
	Create(ctx context.Context, item *financeModels.JournalTemplate) error
	FindByID(ctx context.Context, id string) (*financeModels.JournalTemplate, error)
	List(ctx context.Context, params ListJournalTemplateParams) ([]financeModels.JournalTemplate, int64, error)
	TouchLastUsed(ctx context.Context, id string) error
}

type journalTemplateRepository struct {
	db *gorm.DB
}

func NewJournalTemplateRepository(db *gorm.DB) JournalTemplateRepository {
	return &journalTemplateRepository{db: db}
}

func (r *journalTemplateRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *journalTemplateRepository) Create(ctx context.Context, item *financeModels.JournalTemplate) error {
	return r.getDB(ctx).Create(item).Error
}

func (r *journalTemplateRepository) FindByID(ctx context.Context, id string) (*financeModels.JournalTemplate, error) {
	var item financeModels.JournalTemplate
	err := r.getDB(ctx).First(&item, "id = ?", strings.TrimSpace(id)).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *journalTemplateRepository) List(ctx context.Context, params ListJournalTemplateParams) ([]financeModels.JournalTemplate, int64, error) {
	q := r.getDB(ctx).Model(&financeModels.JournalTemplate{})
	if strings.TrimSpace(params.CompanyID) != "" {
		q = q.Where("company_id = ?", strings.TrimSpace(params.CompanyID))
	}
	if strings.TrimSpace(params.Search) != "" {
		like := "%" + strings.TrimSpace(params.Search) + "%"
		q = q.Where("template_name ILIKE ? OR description ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q = q.Order("updated_at DESC")
	if params.Limit > 0 {
		q = q.Limit(params.Limit)
	}
	if params.Offset > 0 {
		q = q.Offset(params.Offset)
	}

	var items []financeModels.JournalTemplate
	if err := q.Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *journalTemplateRepository) TouchLastUsed(ctx context.Context, id string) error {
	now := apptime.Now()
	return r.getDB(ctx).
		Model(&financeModels.JournalTemplate{}).
		Where("id = ?", strings.TrimSpace(id)).
		Updates(map[string]interface{}{"last_used_at": now, "updated_at": now}).Error
}
