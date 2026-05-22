package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// JobPositionRepository defines the interface for job position data access
type JobPositionRepository interface {
	FindByID(ctx context.Context, id string) (*models.JobPosition, error)
	FindByName(ctx context.Context, name string) (*models.JobPosition, error)
	List(ctx context.Context, req *dto.ListJobPositionsRequest) ([]models.JobPosition, int64, error)
	Create(ctx context.Context, j *models.JobPosition) error
	Update(ctx context.Context, j *models.JobPosition) error
	Delete(ctx context.Context, id string) error
}

type jobPositionRepository struct {
	db *gorm.DB
}

// NewJobPositionRepository creates a new JobPositionRepository
func NewJobPositionRepository(db *gorm.DB) JobPositionRepository {
	return &jobPositionRepository{db: db}
}

func (r *jobPositionRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *jobPositionRepository) FindByID(ctx context.Context, id string) (*models.JobPosition, error) {
	var jobPosition models.JobPosition
	err := r.getDB(ctx).Where("id = ?", id).First(&jobPosition).Error
	if err != nil {
		return nil, err
	}
	return &jobPosition, nil
}

func (r *jobPositionRepository) FindByName(ctx context.Context, name string) (*models.JobPosition, error) {
	var jobPosition models.JobPosition
	err := r.getDB(ctx).Where("LOWER(TRIM(name)) = LOWER(TRIM(?))", name).First(&jobPosition).Error
	if err != nil {
		return nil, err
	}
	return &jobPosition, nil
}

func (r *jobPositionRepository) List(ctx context.Context, req *dto.ListJobPositionsRequest) ([]models.JobPosition, int64, error) {
	var jobPositions []models.JobPosition
	var total int64

	query := r.getDB(ctx).Model(&models.JobPosition{})

	// Apply search filter
	if searchTerm := strings.TrimSpace(req.Search); searchTerm != "" {
		search := "%" + searchTerm + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
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

	// Apply sorting
	sortBy := "updated_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}

	err := query.Order("is_active DESC").
		Order(clause.OrderByColumn{
			Column: clause.Column{Name: sortBy},
			Desc:   req.SortDir != "asc",
		}).
		Offset(offset).
		Limit(perPage).
		Find(&jobPositions).Error
	if err != nil {
		return nil, 0, err
	}

	return jobPositions, total, nil
}

func (r *jobPositionRepository) Create(ctx context.Context, j *models.JobPosition) error {
	return r.getDB(ctx).Create(j).Error
}

func (r *jobPositionRepository) Update(ctx context.Context, j *models.JobPosition) error {
	return r.getDB(ctx).Save(j).Error
}

func (r *jobPositionRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&models.JobPosition{}).Error
}
