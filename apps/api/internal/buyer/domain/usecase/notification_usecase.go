package usecase

import (
	"context"
	"strings"

	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/buyer/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/buyer/domain/dto"
	trustModels "github.com/gilabs/indosupplier/api/internal/trust/data/models"
)

type NotificationUsecase interface {
	ListNotifications(ctx context.Context, userID string) ([]dto.BuyerNotificationResponse, error)
	MarkAllRead(ctx context.Context, userID string) error
}

type notificationUsecase struct {
	db               *gorm.DB
	notificationRepo repositories.NotificationRepository
}

func NewNotificationUsecase(db *gorm.DB, notificationRepo repositories.NotificationRepository) NotificationUsecase {
	return &notificationUsecase{
		db:               db,
		notificationRepo: notificationRepo,
	}
}

func (u *notificationUsecase) getRecipientIDs(ctx context.Context, userID string) []string {
	recipientIDs := []string{userID}
	var buyer buyerModels.BuyerProfile
	if err := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&buyer).Error; err == nil && buyer.ID != "" {
		recipientIDs = append(recipientIDs, buyer.ID)
	}
	return recipientIDs
}

func mapNotificationType(rawType string) string {
	switch strings.ToLower(strings.TrimSpace(rawType)) {
	case "quote", "rfq", "rfq_bid", "bid":
		return "quote"
	case "message", "chat":
		return "message"
	case "alert", "warning", "danger":
		return "alert"
	default:
		return "system"
	}
}

func mapBuyerNotification(n trustModels.Notification) dto.BuyerNotificationResponse {
	return dto.BuyerNotificationResponse{
		ID:     n.ID,
		Title:  n.Title,
		Desc:   n.Body,
		Date:   n.CreatedAt.Format("2006-01-02 15:04"),
		Unread: !n.IsRead,
		Type:   mapNotificationType(n.Type),
	}
}

func (u *notificationUsecase) ListNotifications(ctx context.Context, userID string) ([]dto.BuyerNotificationResponse, error) {
	recipientIDs := u.getRecipientIDs(ctx, userID)
	notifs, err := u.notificationRepo.ListByRecipients(ctx, recipientIDs)
	if err != nil {
		return nil, err
	}

	result := make([]dto.BuyerNotificationResponse, 0, len(notifs))
	for _, n := range notifs {
		result = append(result, mapBuyerNotification(n))
	}
	return result, nil
}

func (u *notificationUsecase) MarkAllRead(ctx context.Context, userID string) error {
	recipientIDs := u.getRecipientIDs(ctx, userID)
	return u.notificationRepo.MarkAllRead(ctx, recipientIDs)
}
