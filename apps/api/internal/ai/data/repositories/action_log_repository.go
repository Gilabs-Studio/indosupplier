package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/ai/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"gorm.io/gorm"
)

// ActionLogRepository defines the interface for AI action log data access
type ActionLogRepository interface {
	Create(ctx context.Context, log *models.AIActionLog) error
	FindByID(ctx context.Context, id string) (*models.AIActionLog, error)
	FindBySessionID(ctx context.Context, sessionID string) ([]models.AIActionLog, error)
	FindPendingBySessionID(ctx context.Context, sessionID string) (*models.AIActionLog, error)
	Update(ctx context.Context, log *models.AIActionLog) error
	FindAll(ctx context.Context, page, perPage int, userID, intent, status string) ([]models.AIActionLog, int64, error)
}

type actionLogRepository struct {
	db *gorm.DB
}

// NewActionLogRepository creates a new action log repository
func NewActionLogRepository(db *gorm.DB) ActionLogRepository {
	return &actionLogRepository{db: db}
}

func (r *actionLogRepository) Create(ctx context.Context, log *models.AIActionLog) error {
	db := database.GetDB(ctx, r.db)
	return db.Create(log).Error
}

func (r *actionLogRepository) FindByID(ctx context.Context, id string) (*models.AIActionLog, error) {
	db := database.GetDB(ctx, r.db)
	var actionLog models.AIActionLog
	if err := db.Where("id = ?", id).First(&actionLog).Error; err != nil {
		return nil, err
	}
	return &actionLog, nil
}

func (r *actionLogRepository) FindBySessionID(ctx context.Context, sessionID string) ([]models.AIActionLog, error) {
	db := database.GetDB(ctx, r.db)
	var logs []models.AIActionLog
	if err := db.Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *actionLogRepository) FindPendingBySessionID(ctx context.Context, sessionID string) (*models.AIActionLog, error) {
	db := database.GetDB(ctx, r.db)
	var actionLog models.AIActionLog
	if err := db.Where("session_id = ? AND status = ?", sessionID, models.ActionStatusPendingConfirmation).
		Order("created_at DESC").
		First(&actionLog).Error; err != nil {
		return nil, err
	}
	return &actionLog, nil
}

func (r *actionLogRepository) Update(ctx context.Context, log *models.AIActionLog) error {
	db := database.GetDB(ctx, r.db)
	return db.Save(log).Error
}

func (r *actionLogRepository) FindAll(ctx context.Context, page, perPage int, userID, intent, status string) ([]models.AIActionLog, int64, error) {
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

	query := db.Model(&models.AIActionLog{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if intent != "" {
		query = query.Where("intent = ?", intent)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []models.AIActionLog
	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
