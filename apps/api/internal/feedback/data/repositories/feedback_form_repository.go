package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/feedback/data/models"
	"gorm.io/gorm"
)

// FeedbackFormRepository defines data access for feedback forms.
type FeedbackFormRepository interface {
	FindByID(ctx context.Context, id string) (*models.FeedbackForm, error)
	FindActiveByOutletID(ctx context.Context, outletID string) (*models.FeedbackForm, error)
	FindByOutletID(ctx context.Context, outletID string) ([]models.FeedbackForm, error)
	NextVersionForOutlet(ctx context.Context, outletID string) (int, error)
	List(ctx context.Context, page, perPage int, outletIDs []string) ([]models.FeedbackForm, int64, error)
	Create(ctx context.Context, form *models.FeedbackForm) error
	Update(ctx context.Context, form *models.FeedbackForm) error
	DeactivateAllForOutlet(ctx context.Context, outletID string) error
	Delete(ctx context.Context, id string) error
}

type feedbackFormRepository struct{ db *gorm.DB }

const (
	feedbackFormDeletedClause = "deleted_at IS NULL"
	feedbackFormOutletClause  = "outlet_id = ? AND deleted_at IS NULL"
	feedbackFormOrderClause   = "created_at DESC"
)

// NewFeedbackFormRepository creates a new FeedbackFormRepository implementation.
func NewFeedbackFormRepository(db *gorm.DB) FeedbackFormRepository {
	return &feedbackFormRepository{db: db}
}

func (r *feedbackFormRepository) dbForContext(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *feedbackFormRepository) FindByID(ctx context.Context, id string) (*models.FeedbackForm, error) {
	var form models.FeedbackForm
	if err := r.dbForContext(ctx).Where("id = ? AND "+feedbackFormDeletedClause, id).First(&form).Error; err != nil {
		return nil, err
	}
	return &form, nil
}

func (r *feedbackFormRepository) FindActiveByOutletID(ctx context.Context, outletID string) (*models.FeedbackForm, error) {
	var form models.FeedbackForm
	if err := r.dbForContext(ctx).
		Where("outlet_id = ? AND is_active = true AND "+feedbackFormDeletedClause, outletID).
		Order(feedbackFormOrderClause).
		First(&form).Error; err != nil {
		return nil, err
	}
	return &form, nil
}

func (r *feedbackFormRepository) FindByOutletID(ctx context.Context, outletID string) ([]models.FeedbackForm, error) {
	var forms []models.FeedbackForm
	if err := r.dbForContext(ctx).
		Where(feedbackFormOutletClause, outletID).
		Order(feedbackFormOrderClause).
		Find(&forms).Error; err != nil {
		return nil, err
	}
	return forms, nil
}

func (r *feedbackFormRepository) NextVersionForOutlet(ctx context.Context, outletID string) (int, error) {
	var nextVersion int
	if err := r.dbForContext(ctx).
		Model(&models.FeedbackForm{}).
		Where(feedbackFormOutletClause, outletID).
		Select("COALESCE(MAX(version), 0) + 1").
		Scan(&nextVersion).Error; err != nil {
		return 0, err
	}
	if nextVersion < 1 {
		nextVersion = 1
	}
	return nextVersion, nil
}

func (r *feedbackFormRepository) List(ctx context.Context, page, perPage int, outletIDs []string) ([]models.FeedbackForm, int64, error) {
	var forms []models.FeedbackForm
	var total int64
	offset := (page - 1) * perPage

	q := r.dbForContext(ctx).Model(&models.FeedbackForm{}).Where(feedbackFormDeletedClause)
	if len(outletIDs) > 0 {
		q = q.Where("outlet_id IN ?", outletIDs)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&forms).Error; err != nil {
		return nil, 0, err
	}
	return forms, total, nil
}

func (r *feedbackFormRepository) Create(ctx context.Context, form *models.FeedbackForm) error {
	return r.dbForContext(ctx).Create(form).Error
}

func (r *feedbackFormRepository) Update(ctx context.Context, form *models.FeedbackForm) error {
	return r.dbForContext(ctx).Save(form).Error
}

func (r *feedbackFormRepository) DeactivateAllForOutlet(ctx context.Context, outletID string) error {
	return r.dbForContext(ctx).Model(&models.FeedbackForm{}).
		Where(feedbackFormOutletClause, outletID).
		Update("is_active", false).Error
}

func (r *feedbackFormRepository) Delete(ctx context.Context, id string) error {
	return r.dbForContext(ctx).Where("id = ?", id).Delete(&models.FeedbackForm{}).Error
}
