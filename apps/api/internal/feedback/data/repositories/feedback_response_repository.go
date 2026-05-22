package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/feedback/data/models"
	"gorm.io/gorm"
)

// FeedbackResponseFilter holds query filter parameters for listing responses.
type FeedbackResponseFilter struct {
	OutletID  string
	OutletIDs []string
	FormID    string
	StartDate string // YYYY-MM-DD
	EndDate   string // YYYY-MM-DD
	Search    string
	Page      int
	PerPage   int
}

// FeedbackResponseRepository defines data access for submitted feedback.
type FeedbackResponseRepository interface {
	FindByID(ctx context.Context, id string) (*models.FeedbackResponse, error)
	Create(ctx context.Context, resp *models.FeedbackResponse) error
	List(ctx context.Context, filter FeedbackResponseFilter) ([]models.FeedbackResponse, int64, error)
	ExistsByTokenID(ctx context.Context, tokenID string) (bool, error)
}

type feedbackResponseRepository struct{ db *gorm.DB }

// NewFeedbackResponseRepository creates a new FeedbackResponseRepository implementation.
func NewFeedbackResponseRepository(db *gorm.DB) FeedbackResponseRepository {
	return &feedbackResponseRepository{db: db}
}

func (r *feedbackResponseRepository) scopedQuery(ctx context.Context) *gorm.DB {
	db := r.db.Session(&gorm.Session{NewDB: true}).WithContext(ctx)
	if isSystemAdmin, _ := ctx.Value("is_system_admin").(bool); isSystemAdmin {
		return db.Table("feedback_responses")
	}
	if tenantID, _ := ctx.Value("tenant_id").(string); tenantID != "" {
		return db.Table("feedback_responses").Where("feedback_responses.tenant_id = ?", tenantID)
	}
	return db.Table("feedback_responses")
}

func (r *feedbackResponseRepository) FindByID(ctx context.Context, id string) (*models.FeedbackResponse, error) {
	var resp models.FeedbackResponse
	if err := r.scopedQuery(ctx).
		Select("feedback_responses.*, sales_orders.id AS sales_order_id").
		Joins("LEFT JOIN sales_orders ON sales_orders.source_pos_order_id = feedback_responses.pos_order_id AND sales_orders.tenant_id = feedback_responses.tenant_id AND sales_orders.deleted_at IS NULL").
		Preload("Form").
		Preload("Token").
		Where("feedback_responses.id = ?", id).First(&resp).Error; err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *feedbackResponseRepository) Create(ctx context.Context, resp *models.FeedbackResponse) error {
	return database.GetDB(ctx, r.db).Create(resp).Error
}

func (r *feedbackResponseRepository) List(ctx context.Context, f FeedbackResponseFilter) ([]models.FeedbackResponse, int64, error) {
	var responses []models.FeedbackResponse
	var total int64

	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}
	offset := (f.Page - 1) * f.PerPage

	q := r.scopedQuery(ctx).
		Joins("LEFT JOIN sales_orders ON sales_orders.source_pos_order_id = feedback_responses.pos_order_id AND sales_orders.tenant_id = feedback_responses.tenant_id AND sales_orders.deleted_at IS NULL")

	if len(f.OutletIDs) > 0 {
		q = q.Where("feedback_responses.outlet_id IN ?", f.OutletIDs)
	} else if f.OutletID != "" {
		q = q.Where("feedback_responses.outlet_id = ?", f.OutletID)
	}
	if f.FormID != "" {
		q = q.Where("feedback_responses.form_id = ?", f.FormID)
	}
	if f.Search != "" {
		q = q.Where("feedback_responses.customer_name ILIKE ?", "%"+f.Search+"%")
	}
	if f.StartDate != "" {
		q = q.Where("feedback_responses.submitted_at >= ?", f.StartDate+" 00:00:00")
	}
	if f.EndDate != "" {
		q = q.Where("feedback_responses.submitted_at <= ?", f.EndDate+" 23:59:59")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Select("feedback_responses.*, sales_orders.id AS sales_order_id").
		Preload("Form").
		Order("feedback_responses.submitted_at DESC").
		Offset(offset).Limit(f.PerPage).
		Find(&responses).Error; err != nil {
		return nil, 0, err
	}
	return responses, total, nil
}

func (r *feedbackResponseRepository) ExistsByTokenID(ctx context.Context, tokenID string) (bool, error) {
	var count int64
	if err := r.scopedQuery(ctx).
		Where("token_id = ?", tokenID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
