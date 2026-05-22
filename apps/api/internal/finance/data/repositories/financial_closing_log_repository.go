package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

type FinancialClosingLogRepository interface {
	Create(ctx context.Context, log *financeModels.FinancialClosingLog) error
}

type financialClosingLogRepository struct {
	db *gorm.DB
}

func NewFinancialClosingLogRepository(db *gorm.DB) FinancialClosingLogRepository {
	return &financialClosingLogRepository{db: db}
}

func (r *financialClosingLogRepository) Create(ctx context.Context, log *financeModels.FinancialClosingLog) error {
	return database.GetDB(ctx, r.db).Create(log).Error
}
