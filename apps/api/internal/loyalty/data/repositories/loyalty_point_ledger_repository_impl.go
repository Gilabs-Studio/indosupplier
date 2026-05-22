package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/loyalty/data/models"
	"gorm.io/gorm"
)

type loyaltyPointLedgerRepo struct {
	db *gorm.DB
}

func NewLoyaltyPointLedgerRepository(db *gorm.DB) LoyaltyPointLedgerRepository {
	return &loyaltyPointLedgerRepo{db: db}
}

func (r *loyaltyPointLedgerRepo) Create(ctx context.Context, entry *models.LoyaltyPointLedger) error {
	return database.GetDB(ctx, r.db).Create(entry).Error
}

func (r *loyaltyPointLedgerRepo) ListByMember(ctx context.Context, params LedgerListParams) ([]models.LoyaltyPointLedger, int64, error) {
	var entries []models.LoyaltyPointLedger
	var total int64

	q := database.GetDB(ctx, r.db).Model(&models.LoyaltyPointLedger{}).Where("member_id = ?", params.MemberID)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PerPage
	err := q.Order("created_at DESC").Offset(offset).Limit(params.PerPage).Find(&entries).Error
	return entries, total, err
}

func (r *loyaltyPointLedgerRepo) FindByTransaction(ctx context.Context, memberID, transactionID, entryType string) (*models.LoyaltyPointLedger, error) {
	var entry models.LoyaltyPointLedger
	err := database.GetDB(ctx, r.db).
		Where("member_id = ? AND transaction_id = ? AND entry_type = ?", memberID, transactionID, entryType).
		First(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &entry, err
}
