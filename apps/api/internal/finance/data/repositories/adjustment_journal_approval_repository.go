package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

type AdjustmentJournalApprovalRepository interface {
	Create(ctx context.Context, item *financeModels.AdjustmentJournalApproval) error
	GetLatestByJournalID(ctx context.Context, journalID string) (*financeModels.AdjustmentJournalApproval, error)
	ListByJournalID(ctx context.Context, journalID string) ([]financeModels.AdjustmentJournalApproval, error)
}

type adjustmentJournalApprovalRepository struct {
	db *gorm.DB
}

func NewAdjustmentJournalApprovalRepository(db *gorm.DB) AdjustmentJournalApprovalRepository {
	return &adjustmentJournalApprovalRepository{db: db}
}

func (r *adjustmentJournalApprovalRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *adjustmentJournalApprovalRepository) Create(ctx context.Context, item *financeModels.AdjustmentJournalApproval) error {
	return r.getDB(ctx).Create(item).Error
}

func (r *adjustmentJournalApprovalRepository) GetLatestByJournalID(ctx context.Context, journalID string) (*financeModels.AdjustmentJournalApproval, error) {
	id := strings.TrimSpace(journalID)
	var item financeModels.AdjustmentJournalApproval
	err := r.getDB(ctx).
		Where("journal_id = ?", id).
		Order("created_at DESC").
		First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *adjustmentJournalApprovalRepository) ListByJournalID(ctx context.Context, journalID string) ([]financeModels.AdjustmentJournalApproval, error) {
	id := strings.TrimSpace(journalID)
	var items []financeModels.AdjustmentJournalApproval
	err := r.getDB(ctx).
		Where("journal_id = ?", id).
		Order("created_at ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
