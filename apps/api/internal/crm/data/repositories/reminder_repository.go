package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"gorm.io/gorm"
)

// ReminderRepository defines data access for task reminders
type ReminderRepository interface {
	Create(ctx context.Context, reminder *models.Reminder) error
	FindByID(ctx context.Context, id string) (*models.Reminder, error)
	FindByTaskID(ctx context.Context, taskID string) ([]models.Reminder, error)
	Update(ctx context.Context, reminder *models.Reminder) error
	Delete(ctx context.Context, id string) error
}

type reminderRepository struct {
	db *gorm.DB
}

// NewReminderRepository creates a new reminder repository
func NewReminderRepository(db *gorm.DB) ReminderRepository {
	return &reminderRepository{db: db}
}

func (r *reminderRepository) Create(ctx context.Context, reminder *models.Reminder) error {
	return database.GetDB(ctx, r.db).Create(reminder).Error
}

func (r *reminderRepository) FindByID(ctx context.Context, id string) (*models.Reminder, error) {
	var reminder models.Reminder
	err := database.GetDB(ctx, r.db).First(&reminder, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &reminder, nil
}

func (r *reminderRepository) FindByTaskID(ctx context.Context, taskID string) ([]models.Reminder, error) {
	var reminders []models.Reminder
	err := database.GetDB(ctx, r.db).
		Where("task_id = ?", taskID).
		Order("remind_at ASC").
		Find(&reminders).Error
	return reminders, err
}

func (r *reminderRepository) Update(ctx context.Context, reminder *models.Reminder) error {
	return database.GetDB(ctx, r.db).Save(reminder).Error
}

func (r *reminderRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.Reminder{}).Error
}
