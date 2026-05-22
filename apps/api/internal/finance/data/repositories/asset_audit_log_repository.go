package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"time"

	"github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetAuditLogRepository defines the interface for asset audit log data access
type AssetAuditLogRepository interface {
	// Create operations
	Create(ctx context.Context, log *models.AssetAuditLog) error
	CreateBatch(ctx context.Context, logs []*models.AssetAuditLog) error

	// Read operations
	GetByID(ctx context.Context, id string) (*models.AssetAuditLog, error)
	GetByAssetID(ctx context.Context, assetID string, limit int) ([]models.AssetAuditLog, error)
	GetByAssetIDAndAction(ctx context.Context, assetID string, action string) ([]models.AssetAuditLog, error)

	// Search and filter
	Search(ctx context.Context, params AuditLogSearchParams) ([]models.AssetAuditLog, int64, error)

	// Statistics
	GetRecentActivity(ctx context.Context, limit int) ([]models.AssetAuditLog, error)
	GetActivityCountByDate(ctx context.Context, startDate, endDate time.Time) (map[string]int64, error)
}

// AuditLogSearchParams defines parameters for searching audit logs
type AuditLogSearchParams struct {
	AssetID     string
	Action      string
	PerformedBy string
	StartDate   *time.Time
	EndDate     *time.Time
	Page        int
	PerPage     int
}

// assetAuditLogRepository implements AssetAuditLogRepository
type assetAuditLogRepository struct {
	db *gorm.DB
}

// NewAssetAuditLogRepository creates a new instance of AssetAuditLogRepository
func NewAssetAuditLogRepository(db *gorm.DB) AssetAuditLogRepository {
	return &assetAuditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *assetAuditLogRepository) Create(ctx context.Context, log *models.AssetAuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	log.PerformedAt = time.Now()

	return database.GetDB(ctx, r.db).Create(log).Error
}

// CreateBatch creates multiple audit log entries in a batch
func (r *assetAuditLogRepository) CreateBatch(ctx context.Context, logs []*models.AssetAuditLog) error {
	for _, log := range logs {
		if log.ID == uuid.Nil {
			log.ID = uuid.New()
		}
		log.PerformedAt = time.Now()
	}

	return database.GetDB(ctx, r.db).CreateInBatches(logs, 100).Error
}

// GetByID retrieves an audit log by ID
func (r *assetAuditLogRepository) GetByID(ctx context.Context, id string) (*models.AssetAuditLog, error) {
	var log models.AssetAuditLog
	err := database.GetDB(ctx, r.db).
		Where("id = ?", id).
		First(&log).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &log, nil
}

// GetByAssetID retrieves audit logs for a specific asset
func (r *assetAuditLogRepository) GetByAssetID(ctx context.Context, assetID string, limit int) ([]models.AssetAuditLog, error) {
	var logs []models.AssetAuditLog

	query := database.GetDB(ctx, r.db).
		Where("asset_id = ?", assetID).
		Order("performed_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetByAssetIDAndAction retrieves audit logs for a specific asset and action
func (r *assetAuditLogRepository) GetByAssetIDAndAction(ctx context.Context, assetID string, action string) ([]models.AssetAuditLog, error) {
	var logs []models.AssetAuditLog
	err := database.GetDB(ctx, r.db).
		Where("asset_id = ? AND action = ?", assetID, action).
		Order("performed_at DESC").
		Find(&logs).Error

	return logs, err
}

// Search searches audit logs with filters
func (r *assetAuditLogRepository) Search(ctx context.Context, params AuditLogSearchParams) ([]models.AssetAuditLog, int64, error) {
	var logs []models.AssetAuditLog
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.AssetAuditLog{})

	// Apply filters
	if params.AssetID != "" {
		query = query.Where("asset_id = ?", params.AssetID)
	}
	if params.Action != "" {
		query = query.Where("action = ?", params.Action)
	}
	if params.PerformedBy != "" {
		query = query.Where("performed_by = ?", params.PerformedBy)
	}
	if params.StartDate != nil {
		query = query.Where("performed_at >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		query = query.Where("performed_at <= ?", params.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := params.Page
	if page <= 0 {
		page = 1
	}
	perPage := params.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	err := query.
		Order("performed_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&logs).Error

	return logs, total, err
}

// GetRecentActivity retrieves recent audit log activity
func (r *assetAuditLogRepository) GetRecentActivity(ctx context.Context, limit int) ([]models.AssetAuditLog, error) {
	var logs []models.AssetAuditLog

	if limit <= 0 {
		limit = 50
	}

	err := database.GetDB(ctx, r.db).
		Order("performed_at DESC").
		Limit(limit).
		Find(&logs).Error

	return logs, err
}

// GetActivityCountByDate returns activity count grouped by date
func (r *assetAuditLogRepository) GetActivityCountByDate(ctx context.Context, startDate, endDate time.Time) (map[string]int64, error) {
	type result struct {
		Date  string
		Count int64
	}

	var results []result
	err := database.GetDB(ctx, r.db).
		Model(&models.AssetAuditLog{}).
		Select("DATE(performed_at) as date, COUNT(*) as count").
		Where("performed_at BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(performed_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	activityMap := make(map[string]int64)
	for _, r := range results {
		activityMap[r.Date] = r.Count
	}

	return activityMap, nil
}
