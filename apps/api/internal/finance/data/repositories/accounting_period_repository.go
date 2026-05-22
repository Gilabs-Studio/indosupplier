package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

type AccountingPeriodRepository interface {
	FindByDate(ctx context.Context, date time.Time) (*financeModels.AccountingPeriod, error)
	Create(ctx context.Context, period *financeModels.AccountingPeriod) error
	UpdateStatus(ctx context.Context, id string, status financeModels.AccountingPeriodStatus, lockedAt *time.Time, lockedBy *string) error
	FindLatestClosed(ctx context.Context) (*financeModels.AccountingPeriod, error)
	// IsClosed checks whether the given period name is closed for the tenant
	IsClosed(ctx context.Context, tenantID string, periodName string) (bool, error)
}

type accountingPeriodRepository struct {
	db *gorm.DB
}

func NewAccountingPeriodRepository(db *gorm.DB) AccountingPeriodRepository {
	return &accountingPeriodRepository{db: db}
}

func (r *accountingPeriodRepository) FindByDate(ctx context.Context, date time.Time) (*financeModels.AccountingPeriod, error) {
	var period financeModels.AccountingPeriod
	if err := database.GetDB(ctx, r.db).
		Where("start_date <= ? AND end_date >= ?", date, date).
		First(&period).Error; err != nil {
		return nil, err
	}
	return &period, nil
}

func (r *accountingPeriodRepository) Create(ctx context.Context, period *financeModels.AccountingPeriod) error {
	return database.GetDB(ctx, r.db).Create(period).Error
}

func (r *accountingPeriodRepository) UpdateStatus(ctx context.Context, id string, status financeModels.AccountingPeriodStatus, lockedAt *time.Time, lockedBy *string) error {
	updates := map[string]interface{}{"status": status}
	if lockedAt != nil {
		updates["locked_at"] = lockedAt
	}
	if lockedBy != nil {
		updates["locked_by"] = lockedBy
	}
	return database.GetDB(ctx, r.db).Model(&financeModels.AccountingPeriod{}).Where("id = ?", id).Updates(updates).Error
}

func (r *accountingPeriodRepository) FindLatestClosed(ctx context.Context) (*financeModels.AccountingPeriod, error) {
	var period financeModels.AccountingPeriod
	if err := database.GetDB(ctx, r.db).
		Where("status = ?", financeModels.AccountingPeriodStatusClosed).
		Order("end_date desc").
		First(&period).Error; err != nil {
		return nil, err
	}
	return &period, nil
}

// IsClosed checks whether the given period name (e.g. "2026-05") is closed for the tenant
func (r *accountingPeriodRepository) IsClosed(ctx context.Context, tenantID string, periodName string) (bool, error) {
	var period financeModels.AccountingPeriod
	db := database.GetDB(ctx, r.db).Where("tenant_id = ? AND period_name = ?", tenantID, periodName)
	if err := db.First(&period).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return period.Status == financeModels.AccountingPeriodStatusClosed, nil
}
