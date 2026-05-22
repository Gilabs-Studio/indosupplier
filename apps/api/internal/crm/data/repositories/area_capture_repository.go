package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"gorm.io/gorm"
)

// AreaCaptureRepository defines the interface for area capture data access
type AreaCaptureRepository interface {
	Create(ctx context.Context, capture *models.AreaCapture) error
	FindByID(ctx context.Context, id string) (*models.AreaCapture, error)
	List(ctx context.Context, req *dto.ListAreaCapturesRequest) ([]models.AreaCapture, int64, error)
	GetHeatmapData(ctx context.Context, areaID string) ([]dto.HeatmapPoint, error)
	GetCoverageByArea(ctx context.Context) ([]dto.AreaCoverageResponse, error)
}

type areaCaptureRepository struct {
	db *gorm.DB
}

// NewAreaCaptureRepository creates a new AreaCaptureRepository
func NewAreaCaptureRepository(db *gorm.DB) AreaCaptureRepository {
	return &areaCaptureRepository{db: db}
}

func (r *areaCaptureRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *areaCaptureRepository) Create(ctx context.Context, capture *models.AreaCapture) error {
	return r.getDB(ctx).Create(capture).Error
}

func (r *areaCaptureRepository) FindByID(ctx context.Context, id string) (*models.AreaCapture, error) {
	var capture models.AreaCapture
	err := r.getDB(ctx).Where("id = ?", id).First(&capture).Error
	if err != nil {
		return nil, err
	}
	return &capture, nil
}

func (r *areaCaptureRepository) List(ctx context.Context, req *dto.ListAreaCapturesRequest) ([]models.AreaCapture, int64, error) {
	var captures []models.AreaCapture
	var total int64

	query := r.getDB(ctx).Model(&models.AreaCapture{})

	if req.EmployeeID != "" {
		query = query.Where("captured_by = ?", req.EmployeeID)
	}
	if req.AreaID != "" {
		query = query.Where("area_id = ?", req.AreaID)
	}
	if req.CaptureType != "" {
		query = query.Where("capture_type = ?", req.CaptureType)
	}
	if req.DateFrom != "" {
		query = query.Where("captured_at >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("captured_at <= ?", req.DateTo)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	err := query.Order("captured_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&captures).Error
	if err != nil {
		return nil, 0, err
	}

	return captures, total, nil
}

// GetHeatmapData aggregates capture points, grouping by rounded lat/lng for density
func (r *areaCaptureRepository) GetHeatmapData(ctx context.Context, areaID string) ([]dto.HeatmapPoint, error) {
	var points []dto.HeatmapPoint

	query := r.getDB(ctx).Model(&models.AreaCapture{}).
		Select("ROUND(latitude::numeric, 3) as latitude, ROUND(longitude::numeric, 3) as longitude, COUNT(*) as intensity").
		Group("ROUND(latitude::numeric, 3), ROUND(longitude::numeric, 3)")

	if areaID != "" {
		query = query.Where("area_id = ?", areaID)
	}

	if err := query.Find(&points).Error; err != nil {
		return nil, err
	}

	return points, nil
}

// GetCoverageByArea returns coverage stats per area
func (r *areaCaptureRepository) GetCoverageByArea(ctx context.Context) ([]dto.AreaCoverageResponse, error) {
	var results []dto.AreaCoverageResponse

	err := r.getDB(ctx).
		Table("crm_area_captures AS ac").
		Select(`
			ac.area_id,
			a.name AS area_name,
			a.code AS area_code,
			COUNT(*) AS total_visits,
			COUNT(DISTINCT CONCAT(ROUND(ac.latitude::numeric, 4), ',', ROUND(ac.longitude::numeric, 4))) AS unique_points,
			MAX(ac.captured_at)::text AS last_visit_at
		`).
		Joins("JOIN areas AS a ON a.id = ac.area_id").
		Where("ac.area_id IS NOT NULL AND ac.deleted_at IS NULL AND a.deleted_at IS NULL").
		Group("ac.area_id, a.name, a.code").
		Order("total_visits DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}
