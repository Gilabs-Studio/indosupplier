package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

type FinancialClosingSnapshotRepository interface {
	Create(ctx context.Context, snapshot *financeModels.FinancialClosingSnapshot) error
	FindByPeriodEndDate(ctx context.Context, periodEndDate string) (*financeModels.FinancialClosingSnapshot, error)
}

type financialClosingSnapshotRepository struct {
	db *gorm.DB
}

func NewFinancialClosingSnapshotRepository(db *gorm.DB) FinancialClosingSnapshotRepository {
	return &financialClosingSnapshotRepository{db: db}
}

func (r *financialClosingSnapshotRepository) Create(ctx context.Context, snapshot *financeModels.FinancialClosingSnapshot) error {
	return database.GetDB(ctx, r.db).Create(snapshot).Error
}

func (r *financialClosingSnapshotRepository) FindByPeriodEndDate(ctx context.Context, periodEndDate string) (*financeModels.FinancialClosingSnapshot, error) {
	var item financeModels.FinancialClosingSnapshot
	if err := database.GetDB(ctx, r.db).Where("period_end_date = ?", periodEndDate).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
