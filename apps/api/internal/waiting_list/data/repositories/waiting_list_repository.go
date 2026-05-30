package repositories

import (
	"context"

	"gorm.io/gorm"
	"github.com/gilabs/indosupplier/api/internal/waiting_list/data/models"
)

type WaitingListRepository interface {
	Create(ctx context.Context, w *models.WaitingList) error
	FindByEmail(ctx context.Context, email string) (*models.WaitingList, error)
	FindByID(ctx context.Context, id string) (*models.WaitingList, error)
	List(ctx context.Context, limit, offset int, status string) ([]models.WaitingList, int64, error)
	Update(ctx context.Context, w *models.WaitingList) error
	Delete(ctx context.Context, id string) error
}

type waitingListRepository struct {
	db *gorm.DB
}

func NewWaitingListRepository(db *gorm.DB) WaitingListRepository {
	return &waitingListRepository{db: db}
}

func (r *waitingListRepository) Create(ctx context.Context, w *models.WaitingList) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *waitingListRepository) FindByEmail(ctx context.Context, email string) (*models.WaitingList, error) {
	var w models.WaitingList
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *waitingListRepository) FindByID(ctx context.Context, id string) (*models.WaitingList, error) {
	var w models.WaitingList
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *waitingListRepository) List(ctx context.Context, limit, offset int, status string) ([]models.WaitingList, int64, error) {
	var items []models.WaitingList
	var total int64

	query := r.db.WithContext(ctx).Model(&models.WaitingList{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *waitingListRepository) Update(ctx context.Context, w *models.WaitingList) error {
	return r.db.WithContext(ctx).Save(w).Error
}

func (r *waitingListRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.WaitingList{}, "id = ?", id).Error
}
