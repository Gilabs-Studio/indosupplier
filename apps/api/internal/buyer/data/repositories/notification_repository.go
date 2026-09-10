package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	trustModels "github.com/gilabs/indosupplier/api/internal/trust/data/models"
)

type NotificationRepository interface {
	ListByRecipients(ctx context.Context, recipientIDs []string) ([]trustModels.Notification, error)
	MarkAllRead(ctx context.Context, recipientIDs []string) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *notificationRepository) ListByRecipients(ctx context.Context, recipientIDs []string) ([]trustModels.Notification, error) {
	var notifs []trustModels.Notification
	if len(recipientIDs) == 0 {
		return notifs, nil
	}
	err := r.getDB(ctx).
		Where("recipient_id IN ? AND channel = ?", recipientIDs, "in_app").
		Order("created_at DESC").
		Limit(50).
		Find(&notifs).Error
	return notifs, err
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, recipientIDs []string) error {
	if len(recipientIDs) == 0 {
		return nil
	}
	now := apptime.Now()
	return r.getDB(ctx).Model(&trustModels.Notification{}).
		Where("recipient_id IN ? AND is_read = ?", recipientIDs, false).
		Updates(map[string]interface{}{
			"is_read":    true,
			"read_at":    now,
			"updated_at": now,
		}).Error
}
