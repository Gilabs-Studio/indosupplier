package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	"github.com/gilabs/indosupplier/api/internal/rfq/data/models"
)

type RFQRepository interface {
	Create(ctx context.Context, rfq *models.RFQ, attachment *models.RFQAttachment, recipients []models.RFQRecipient) error
	FindByID(ctx context.Context, id string) (*models.RFQ, *models.RFQAttachment, int, error)
	List(ctx context.Context, buyerProfileID string, status string, page, perPage int) ([]models.RFQ, []int, int64, error)
	GetBids(ctx context.Context, rfqID string) ([]models.RFQRecipient, error)
	AcceptBid(ctx context.Context, rfqID string, bidID string) error
	ListForSupplier(ctx context.Context, supplierProfileID string, page, perPage int) ([]models.RFQRecipient, int64, error)
	FindForSupplier(ctx context.Context, supplierProfileID string, rfqID string) (*models.RFQRecipient, error)
	SubmitProposal(ctx context.Context, recipient *models.RFQRecipient, message *models.RFQMessage) error
}

type rfqRepository struct {
	db *gorm.DB
}

func NewRFQRepository(db *gorm.DB) RFQRepository {
	return &rfqRepository{db: db}
}

func (r *rfqRepository) Create(ctx context.Context, rfq *models.RFQ, attachment *models.RFQAttachment, recipients []models.RFQRecipient) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rfq).Error; err != nil {
			return err
		}
		if attachment != nil {
			attachment.RFQID = rfq.ID
			if err := tx.Create(attachment).Error; err != nil {
				return err
			}
		}
		for i := range recipients {
			recipients[i].RFQID = rfq.ID
			if err := tx.Create(&recipients[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *rfqRepository) FindByID(ctx context.Context, id string) (*models.RFQ, *models.RFQAttachment, int, error) {
	var rfq models.RFQ
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rfq).Error; err != nil {
		return nil, nil, 0, err
	}

	var attachment models.RFQAttachment
	err := r.db.WithContext(ctx).Where("rfq_id = ?", rfq.ID).First(&attachment).Error
	var attachPtr *models.RFQAttachment
	if err == nil {
		attachPtr = &attachment
	}

	var replies int64
	r.db.WithContext(ctx).
		Model(&models.RFQRecipient{}).
		Where("rfq_id = ? AND (status = ? OR status = ?)", rfq.ID, "responded", "processing").
		Count(&replies)

	return &rfq, attachPtr, int(replies), nil
}

func (r *rfqRepository) List(ctx context.Context, buyerProfileID string, status string, page, perPage int) ([]models.RFQ, []int, int64, error) {
	var rfqList []models.RFQ
	var total int64

	q := r.db.WithContext(ctx).Model(&models.RFQ{}).Where("buyer_profile_id = ?", buyerProfileID)

	// In transaction_usecase list, completed status gets filtered.
	// RFQ status is calculated dynamically.
	// Waiting: ClosedAt IS NULL, replies count = 0
	// Received (Offers Received): ClosedAt IS NULL, replies count > 0
	// Completed: ClosedAt IS NOT NULL
	if status == "waiting" {
		// Filter RFQs with no replies yet
		q = q.Where("closed_at IS NULL").
			Where("id NOT IN (?)", r.db.Model(&models.RFQRecipient{}).Select("rfq_id").Where("status = ?", "responded"))
	} else if status == "received" {
		// Filter RFQs that have replies
		q = q.Where("closed_at IS NULL").
			Where("id IN (?)", r.db.Model(&models.RFQRecipient{}).Select("rfq_id").Where("status = ?", "responded"))
	} else if status == "completed" {
		q = q.Where("closed_at IS NOT NULL")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	page, perPage = utils.NormalizePagination(page, perPage, 0)
	offset := utils.PaginationOffset(page, perPage)
	if err := q.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&rfqList).Error; err != nil {
		return nil, nil, 0, err
	}

	repliesCounts := make([]int, len(rfqList))
	for i, rfq := range rfqList {
		var replies int64
		r.db.WithContext(ctx).
			Model(&models.RFQRecipient{}).
			Where("rfq_id = ? AND (status = ? OR status = ?)", rfq.ID, "responded", "processing").
			Count(&replies)
		repliesCounts[i] = int(replies)
	}

	return rfqList, repliesCounts, total, nil
}

func (r *rfqRepository) GetBids(ctx context.Context, rfqID string) ([]models.RFQRecipient, error) {
	var recipients []models.RFQRecipient
	// Only return recipients that responded or prepared a quote
	err := r.db.WithContext(ctx).
		Where("rfq_id = ? AND (status = ? OR status = ?)", rfqID, "responded", "processing").
		Find(&recipients).Error
	return recipients, err
}

func (r *rfqRepository) AcceptBid(ctx context.Context, rfqID string, bidID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verify recipient exists and belongs to this RFQ
		var recipient models.RFQRecipient
		if err := tx.Where("id = ? AND rfq_id = ?", bidID, rfqID).First(&recipient).Error; err != nil {
			return err
		}

		// Update recipient status
		if err := tx.Model(&recipient).Update("status", "accepted").Error; err != nil {
			return err
		}

		// Close/complete RFQ
		now := apptime.Now()
		if err := tx.Model(&models.RFQ{}).Where("id = ?", rfqID).Updates(map[string]interface{}{
			"closed_at": &now,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *rfqRepository) ListForSupplier(ctx context.Context, supplierProfileID string, page, perPage int) ([]models.RFQRecipient, int64, error) {
	var recipients []models.RFQRecipient
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.RFQRecipient{}).
		Where("supplier_profile_id = ?", supplierProfileID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, perPage = utils.NormalizePagination(page, perPage, 10)
	offset := utils.PaginationOffset(page, perPage)

	err := query.
		Joins("JOIN rfqs ON rfqs.id = rfq_recipients.rfq_id").
		Order("rfqs.created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&recipients).Error
	return recipients, total, err
}

func (r *rfqRepository) FindForSupplier(ctx context.Context, supplierProfileID string, rfqID string) (*models.RFQRecipient, error) {
	var recipient models.RFQRecipient
	err := r.db.WithContext(ctx).
		Where("supplier_profile_id = ? AND rfq_id = ?", supplierProfileID, rfqID).
		First(&recipient).Error
	if err != nil {
		return nil, err
	}
	return &recipient, nil
}

func (r *rfqRepository) SubmitProposal(ctx context.Context, recipient *models.RFQRecipient, message *models.RFQMessage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := apptime.Now()
		updates := map[string]interface{}{
			"status":       "responded",
			"responded_at": &now,
		}
		if err := tx.Model(&models.RFQRecipient{}).
			Where("id = ?", recipient.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		return tx.Create(message).Error
	})
}
