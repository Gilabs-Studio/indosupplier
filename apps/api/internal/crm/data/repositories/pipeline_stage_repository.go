package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PipelineStageRepository defines the interface for pipeline stage data access
type PipelineStageRepository interface {
	Create(ctx context.Context, stage *models.PipelineStage) error
	FindByID(ctx context.Context, id string) (*models.PipelineStage, error)
	List(ctx context.Context, params ListParams) ([]models.PipelineStage, int64, error)
	Update(ctx context.Context, stage *models.PipelineStage) error
	Delete(ctx context.Context, id string) error
	FindByOrder(ctx context.Context, order int) (*models.PipelineStage, error)
	FindWonStage(ctx context.Context) (*models.PipelineStage, error)
	FindLostStage(ctx context.Context) (*models.PipelineStage, error)
}

type pipelineStageRepository struct {
	db *gorm.DB
}

// NewPipelineStageRepository creates a new pipeline stage repository
func NewPipelineStageRepository(db *gorm.DB) PipelineStageRepository {
	return &pipelineStageRepository{db: db}
}

func (r *pipelineStageRepository) Create(ctx context.Context, stage *models.PipelineStage) error {
	return database.GetDB(ctx, r.db).Create(stage).Error
}

func (r *pipelineStageRepository) FindByID(ctx context.Context, id string) (*models.PipelineStage, error) {
	var stage models.PipelineStage
	err := database.GetDB(ctx, r.db).First(&stage, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &stage, nil
}

func (r *pipelineStageRepository) List(ctx context.Context, params ListParams) ([]models.PipelineStage, int64, error) {
	var stages []models.PipelineStage
	var total int64

	query := database.GetDB(ctx, r.db).Model(&models.PipelineStage{})

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search, search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.SortBy != "" {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: params.SortBy}, Desc: params.SortDir == "desc"})
	} else {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}, Desc: false})
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Find(&stages).Error; err != nil {
		return nil, 0, err
	}

	return stages, total, nil
}

func (r *pipelineStageRepository) Update(ctx context.Context, stage *models.PipelineStage) error {
	return database.GetDB(ctx, r.db).Save(stage).Error
}

func (r *pipelineStageRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.PipelineStage{}, "id = ?", id).Error
}

func (r *pipelineStageRepository) FindByOrder(ctx context.Context, order int) (*models.PipelineStage, error) {
	var stage models.PipelineStage
	err := database.GetDB(ctx, r.db).Where("\"order\" = ?", order).First(&stage).Error
	if err != nil {
		return nil, err
	}
	return &stage, nil
}

func (r *pipelineStageRepository) FindWonStage(ctx context.Context) (*models.PipelineStage, error) {
	var stage models.PipelineStage
	err := database.GetDB(ctx, r.db).Where("is_won = ?", true).First(&stage).Error
	if err != nil {
		return nil, err
	}
	return &stage, nil
}

func (r *pipelineStageRepository) FindLostStage(ctx context.Context) (*models.PipelineStage, error) {
	var stage models.PipelineStage
	err := database.GetDB(ctx, r.db).Where("is_lost = ?", true).First(&stage).Error
	if err != nil {
		return nil, err
	}
	return &stage, nil
}
