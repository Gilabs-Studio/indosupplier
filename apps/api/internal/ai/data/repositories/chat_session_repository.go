package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/ai/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"gorm.io/gorm"
)

// ChatSessionRepository defines the interface for AI chat session data access
type ChatSessionRepository interface {
	Create(ctx context.Context, session *models.AIChatSession) error
	FindByID(ctx context.Context, id string) (*models.AIChatSession, error)
	FindByIDWithMessages(ctx context.Context, id string) (*models.AIChatSession, error)
	FindByUserID(ctx context.Context, userID string, page, perPage int, status, search string) ([]models.AIChatSession, int64, error)
	Update(ctx context.Context, session *models.AIChatSession) error
	UpdateTitle(ctx context.Context, id string, title string) error
	Delete(ctx context.Context, id string) error
	IncrementMessageCount(ctx context.Context, id string) error
	UpdateLastActivity(ctx context.Context, id string) error
}

type chatSessionRepository struct {
	db *gorm.DB
}

// NewChatSessionRepository creates a new chat session repository
func NewChatSessionRepository(db *gorm.DB) ChatSessionRepository {
	return &chatSessionRepository{db: db}
}

func (r *chatSessionRepository) Create(ctx context.Context, session *models.AIChatSession) error {
	db := database.GetDB(ctx, r.db)
	return db.Create(session).Error
}

func (r *chatSessionRepository) FindByID(ctx context.Context, id string) (*models.AIChatSession, error) {
	var session models.AIChatSession
	db := database.GetDB(ctx, r.db)
	if err := db.Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *chatSessionRepository) FindByIDWithMessages(ctx context.Context, id string) (*models.AIChatSession, error) {
	var session models.AIChatSession
	db := database.GetDB(ctx, r.db)
	if err := db.
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Actions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *chatSessionRepository) FindByUserID(ctx context.Context, userID string, page, perPage int, status, search string) ([]models.AIChatSession, int64, error) {
	db := database.GetDB(ctx, r.db)

	if perPage > 100 {
		perPage = 100
	}
	if perPage < 1 {
		perPage = 20
	}
	if page < 1 {
		page = 1
	}

	query := db.Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("title ILIKE ?", search + "%")
	}

	var total int64
	if err := query.Model(&models.AIChatSession{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var sessions []models.AIChatSession
	offset := (page - 1) * perPage
	if err := query.Order("last_activity DESC").Offset(offset).Limit(perPage).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

func (r *chatSessionRepository) Update(ctx context.Context, session *models.AIChatSession) error {
	db := database.GetDB(ctx, r.db)
	return db.Save(session).Error
}

func (r *chatSessionRepository) UpdateTitle(ctx context.Context, id string, title string) error {
	db := database.GetDB(ctx, r.db)
	return db.Model(&models.AIChatSession{}).Where("id = ?", id).
		UpdateColumn("title", title).Error
}

func (r *chatSessionRepository) Delete(ctx context.Context, id string) error {
	db := database.GetDB(ctx, r.db)
	return db.Where("id = ?", id).Delete(&models.AIChatSession{}).Error
}

func (r *chatSessionRepository) IncrementMessageCount(ctx context.Context, id string) error {
	db := database.GetDB(ctx, r.db)
	return db.Model(&models.AIChatSession{}).Where("id = ?", id).
		UpdateColumn("message_count", gorm.Expr("message_count + 1")).Error
}

func (r *chatSessionRepository) UpdateLastActivity(ctx context.Context, id string) error {
	db := database.GetDB(ctx, r.db)
	return db.Model(&models.AIChatSession{}).Where("id = ?", id).
		UpdateColumn("last_activity", gorm.Expr("NOW()")).Error
}
