package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"gorm.io/gorm"
)

type OpeningBalanceRepository interface {
	ListLines(ctx context.Context, companyID, fiscalYearID string) ([]financeModels.OpeningBalanceLine, error)
	ReplaceLines(ctx context.Context, companyID, fiscalYearID string, lines []financeModels.OpeningBalanceLine) error
	DeleteLines(ctx context.Context, companyID, fiscalYearID string) error
	HasPostedOpeningJournal(ctx context.Context, companyID, fiscalYearID string) (bool, error)
	GetPostedOpeningJournalID(ctx context.Context, companyID, fiscalYearID string) (*string, error)
	HasPostedOperationalJournalInRange(ctx context.Context, startDate, endDate string) (bool, error)
	GetDB(ctx context.Context) *gorm.DB
}

type openingBalanceRepository struct {
	db *gorm.DB
}

func NewOpeningBalanceRepository(db *gorm.DB) OpeningBalanceRepository {
	return &openingBalanceRepository{db: db}
}

func (r *openingBalanceRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok && tx != nil {
		return tx
	}
	return database.GetDB(ctx, r.db)
}

func (r *openingBalanceRepository) GetDB(ctx context.Context) *gorm.DB {
	return r.getDB(ctx)
}

func (r *openingBalanceRepository) ListLines(ctx context.Context, companyID, fiscalYearID string) ([]financeModels.OpeningBalanceLine, error) {
	items := make([]financeModels.OpeningBalanceLine, 0)
	err := r.getDB(ctx).
		Where("company_id = ? AND fiscal_year_id = ?", companyID, fiscalYearID).
		Order("updated_at desc").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *openingBalanceRepository) ReplaceLines(ctx context.Context, companyID, fiscalYearID string, lines []financeModels.OpeningBalanceLine) error {
	db := r.getDB(ctx)
	if err := db.Where("company_id = ? AND fiscal_year_id = ?", companyID, fiscalYearID).Delete(&financeModels.OpeningBalanceLine{}).Error; err != nil {
		return err
	}
	if len(lines) == 0 {
		return nil
	}
	return db.Create(&lines).Error
}

func (r *openingBalanceRepository) DeleteLines(ctx context.Context, companyID, fiscalYearID string) error {
	return r.getDB(ctx).Where("company_id = ? AND fiscal_year_id = ?", companyID, fiscalYearID).Delete(&financeModels.OpeningBalanceLine{}).Error
}

func (r *openingBalanceRepository) HasPostedOpeningJournal(ctx context.Context, companyID, fiscalYearID string) (bool, error) {
	refID := companyID + ":" + fiscalYearID
	var count int64
	err := r.getDB(ctx).
		Model(&financeModels.JournalEntry{}).
		Where("reference_type = ? AND reference_id = ? AND status = ?", string(financeModels.RefOpeningBalance), refID, financeModels.JournalStatusPosted).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *openingBalanceRepository) GetPostedOpeningJournalID(ctx context.Context, companyID, fiscalYearID string) (*string, error) {
	refID := companyID + ":" + fiscalYearID
	var entry financeModels.JournalEntry
	err := r.getDB(ctx).
		Select("id").
		Model(&financeModels.JournalEntry{}).
		Where("reference_type = ? AND reference_id = ? AND status = ?", string(financeModels.RefOpeningBalance), refID, financeModels.JournalStatusPosted).
		Order("created_at desc").
		Take(&entry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	id := entry.ID
	return &id, nil
}

func (r *openingBalanceRepository) HasPostedOperationalJournalInRange(ctx context.Context, startDate, endDate string) (bool, error) {
	var count int64
	err := r.getDB(ctx).
		Model(&financeModels.JournalEntry{}).
		Where("status = ?", financeModels.JournalStatusPosted).
		Where("journal_type <> ?", financeModels.JournalTypeOpeningBalance).
		Where("entry_date >= ? AND entry_date <= ?", startDate, endDate).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
