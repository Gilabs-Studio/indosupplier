package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/pos/data/models"
	"gorm.io/gorm"
)

// PosSessionRepository defines data access for POS sessions
type PosSessionRepository interface {
	Create(ctx context.Context, session *models.PosSession) error
	GetByID(ctx context.Context, id string) (*models.PosSession, error)
	Update(ctx context.Context, session *models.PosSession) error
	FindActiveByOutlet(ctx context.Context, outletID string) (*models.PosSession, error)
	FindActiveByCashier(ctx context.Context, cashierID string) (*models.PosSession, error)
	List(ctx context.Context, outletID string, page, perPage int) ([]models.PosSession, int64, error)
	ListByParams(ctx context.Context, params POSSessionListParams) ([]models.PosSession, int64, error)
	GetNextSessionCode(ctx context.Context) (string, error)
}

// POSSessionListParams defines handler-level offset-based filter parameters
type POSSessionListParams struct {
	OutletID  string
	CashierID string
	Status    string
	Limit     int
	Offset    int
}

type posSessionRepository struct {
	db *gorm.DB
}

func NewPosSessionRepository(db *gorm.DB) PosSessionRepository {
	return &posSessionRepository{db: db}
}

func (r *posSessionRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *posSessionRepository) Create(ctx context.Context, session *models.PosSession) error {
	return r.getDB(ctx).Create(session).Error
}

func (r *posSessionRepository) GetByID(ctx context.Context, id string) (*models.PosSession, error) {
	var session models.PosSession
	if err := r.getDB(ctx).First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *posSessionRepository) Update(ctx context.Context, session *models.PosSession) error {
	return r.getDB(ctx).Save(session).Error
}

func (r *posSessionRepository) FindActiveByOutlet(ctx context.Context, outletID string) (*models.PosSession, error) {
	var session models.PosSession
	err := r.getDB(ctx).
		Where("outlet_id = ? AND status = ?", outletID, models.PosSessionStatusOpen).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *posSessionRepository) List(ctx context.Context, outletID string, page, perPage int) ([]models.PosSession, int64, error) {
	var sessions []models.PosSession
	var total int64

	query := r.getDB(ctx).Model(&models.PosSession{})

	// Apply dynamic scope filter
	query = security.ApplyScopeFilter(query, ctx, security.ScopeQueryOptions{
		OutletIDColumn: "outlet_id",
	})

	if outletID != "" {
		query = query.Where("outlet_id = ?", outletID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Order("opened_at DESC").Offset(offset).Limit(perPage).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

func (r *posSessionRepository) GetNextSessionCode(ctx context.Context) (string, error) {
	var count int64
	now := time.Now()
	prefix := fmt.Sprintf("POS-%s-", now.Format("20060102"))

	r.getDB(ctx).Model(&models.PosSession{}).
		Where("code LIKE ?", prefix+"%").
		Count(&count)

	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (r *posSessionRepository) FindActiveByCashier(ctx context.Context, cashierID string) (*models.PosSession, error) {
	var session models.PosSession
	err := r.getDB(ctx).
		Where("cashier_id = ? AND status = ?", cashierID, models.PosSessionStatusOpen).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *posSessionRepository) ListByParams(ctx context.Context, params POSSessionListParams) ([]models.PosSession, int64, error) {
	var sessions []models.PosSession
	var total int64

	query := r.getDB(ctx).Model(&models.PosSession{})

	// Apply dynamic scope filter
	query = security.ApplyScopeFilter(query, ctx, security.ScopeQueryOptions{
		OutletIDColumn: "outlet_id",
	})

	if params.OutletID != "" {
		query = query.Where("outlet_id = ?", params.OutletID)
	}
	if params.CashierID != "" {
		query = query.Where("cashier_id = ?", params.CashierID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := params.Limit
	if limit < 1 {
		limit = 20
	}

	if err := query.Order("opened_at DESC").Offset(params.Offset).Limit(limit).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}
