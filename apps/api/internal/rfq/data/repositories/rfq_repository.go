package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	ListRFQMessages(ctx context.Context, rfqID string, supplierProfileID string) ([]models.RFQMessage, error)
	CreateRFQMessage(ctx context.Context, message *models.RFQMessage) error
	AcceptBidBySupplier(ctx context.Context, rfqID string, supplierProfileID string) (*models.RFQRecipient, error)
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
		Where("rfq_id = ? AND status IN ?", rfq.ID, []string{"responded", "processing", "accepted"}).
		Count(&replies)

	return &rfq, attachPtr, int(replies), nil
}

func (r *rfqRepository) List(ctx context.Context, buyerProfileID string, status string, page, perPage int) ([]models.RFQ, []int, int64, error) {
	var rfqList []models.RFQ
	var total int64

	q := r.db.WithContext(ctx).Model(&models.RFQ{}).Where("buyer_profile_id = ?", buyerProfileID)

	// Filter based on tab status
	// Waiting: ClosedAt IS NULL, replies count = 0
	// Received (Offers Received): RFQs that have replies/proposals from suppliers
	// Completed: ClosedAt IS NOT NULL
	if status == "waiting" {
		// Filter RFQs with no replies yet
		q = q.Where("closed_at IS NULL").
			Where("id NOT IN (?)", r.db.Model(&models.RFQRecipient{}).Select("rfq_id").Where("status IN ?", []string{"responded", "processing", "accepted"}))
	} else if status == "received" {
		// Filter RFQs that have received offers/proposals and are still open (not closed)
		q = q.Where("closed_at IS NULL").
			Where("id IN (?)", r.db.Model(&models.RFQRecipient{}).Select("rfq_id").Where("status IN ?", []string{"responded", "processing", "accepted"}))
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
	if len(rfqList) > 0 {
		rfqIDs := make([]string, len(rfqList))
		for i, rfq := range rfqList {
			rfqIDs[i] = rfq.ID
		}

		type ReplyCountRow struct {
			RFQID string
			Count int
		}
		var rows []ReplyCountRow
		r.db.WithContext(ctx).
			Model(&models.RFQRecipient{}).
			Select("rfq_id, count(*) as count").
			Where("rfq_id IN ? AND status IN ?", rfqIDs, []string{"responded", "processing", "accepted"}).
			Group("rfq_id").
			Scan(&rows)

		countMap := make(map[string]int)
		for _, row := range rows {
			countMap[row.RFQID] = row.Count
		}

		for i, rfq := range rfqList {
			repliesCounts[i] = countMap[rfq.ID]
		}
	}

	return rfqList, repliesCounts, total, nil
}

func (r *rfqRepository) GetBids(ctx context.Context, rfqID string) ([]models.RFQRecipient, error) {
	var recipients []models.RFQRecipient
	// Return recipients that responded, are processing, or have been accepted
	err := r.db.WithContext(ctx).
		Where("rfq_id = ? AND status IN ?", rfqID, []string{"responded", "processing", "accepted"}).
		Find(&recipients).Error
	return recipients, err
}

func (r *rfqRepository) AcceptBid(ctx context.Context, rfqID string, bidID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Pessimistic lock on the RFQ to prevent concurrent bid acceptance
		var rfq models.RFQ
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", rfqID).First(&rfq).Error; err != nil {
			return err
		}
		if rfq.ClosedAt != nil {
			return errors.New("rfq is already closed")
		}

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
			"updated_at":   now,
		}
		if err := tx.Model(&models.RFQRecipient{}).
			Where("id = ?", recipient.ID).
			Updates(updates).Error; err != nil {
			return err
		}

		var existing models.RFQMessage
		if err := tx.Where("rfq_id = ? AND supplier_profile_id = ? AND message_type = 'offer'", message.RFQID, message.SupplierProfileID).First(&existing).Error; err == nil {
			return tx.Model(&existing).Updates(map[string]interface{}{
				"body":            message.Body,
				"price_formatted": message.PriceFormatted,
				"moq":             message.MOQ,
				"delivery_time":   message.DeliveryTime,
				"metadata":        message.Metadata,
				"updated_at":      now,
			}).Error
		}

		return tx.Create(message).Error
	})
}

func (r *rfqRepository) ListRFQMessages(ctx context.Context, rfqID string, supplierProfileID string) ([]models.RFQMessage, error) {
	var messages []models.RFQMessage
	err := r.db.WithContext(ctx).
		Where("rfq_id = ? AND supplier_profile_id = ?", rfqID, supplierProfileID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *rfqRepository) CreateRFQMessage(ctx context.Context, message *models.RFQMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *rfqRepository) AcceptBidBySupplier(ctx context.Context, rfqID string, supplierProfileID string) (*models.RFQRecipient, error) {
	var recipient models.RFQRecipient
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rfq models.RFQ
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", rfqID).First(&rfq).Error; err != nil {
			return err
		}
		if rfq.ClosedAt != nil {
			return errors.New("rfq is already closed")
		}

		if err := tx.Where("rfq_id = ? AND supplier_profile_id = ?", rfqID, supplierProfileID).First(&recipient).Error; err != nil {
			return err
		}

		now := apptime.Now()
		if err := tx.Model(&models.RFQRecipient{}).Where("id = ?", recipient.ID).Updates(map[string]interface{}{
			"status":     "accepted",
			"updated_at": now,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&rfq).Updates(map[string]interface{}{
			"closed_at":  now,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &recipient, nil
}

