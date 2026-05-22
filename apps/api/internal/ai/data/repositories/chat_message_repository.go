package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/ai/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"gorm.io/gorm"
)

// ChatMessageRepository defines the interface for AI chat message data access
type ChatMessageRepository interface {
	Create(ctx context.Context, message *models.AIChatMessage) error
	FindBySessionID(ctx context.Context, sessionID string, limit int) ([]models.AIChatMessage, error)
	FindByID(ctx context.Context, id string) (*models.AIChatMessage, error)
}

type chatMessageRepository struct {
	db *gorm.DB
}

// NewChatMessageRepository creates a new chat message repository
func NewChatMessageRepository(db *gorm.DB) ChatMessageRepository {
	return &chatMessageRepository{db: db}
}

func (r *chatMessageRepository) Create(ctx context.Context, message *models.AIChatMessage) error {
	db := database.GetDB(ctx, r.db)
	return db.Create(message).Error
}

func (r *chatMessageRepository) FindBySessionID(ctx context.Context, sessionID string, limit int) ([]models.AIChatMessage, error) {
	db := database.GetDB(ctx, r.db)

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var messages []models.AIChatMessage
	if err := db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *chatMessageRepository) FindByID(ctx context.Context, id string) (*models.AIChatMessage, error) {
	db := database.GetDB(ctx, r.db)
	var message models.AIChatMessage
	if err := db.Where("id = ?", id).First(&message).Error; err != nil {
		return nil, err
	}
	return &message, nil
}
