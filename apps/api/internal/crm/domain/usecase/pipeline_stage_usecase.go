package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/mapper"
	"github.com/google/uuid"
)

// PipelineStageUsecase defines the interface for pipeline stage business logic
type PipelineStageUsecase interface {
	Create(ctx context.Context, req dto.CreatePipelineStageRequest) (dto.PipelineStageResponse, error)
	GetByID(ctx context.Context, id string) (dto.PipelineStageResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.PipelineStageResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdatePipelineStageRequest) (dto.PipelineStageResponse, error)
	Delete(ctx context.Context, id string) error
}

type pipelineStageUsecase struct {
	repo      repositories.PipelineStageRepository
	dealRepo  repositories.DealRepository
}

// NewPipelineStageUsecase creates a new pipeline stage usecase
func NewPipelineStageUsecase(repo repositories.PipelineStageRepository, dealRepo repositories.DealRepository) PipelineStageUsecase {
	return &pipelineStageUsecase{repo: repo, dealRepo: dealRepo}
}

func (u *pipelineStageUsecase) Create(ctx context.Context, req dto.CreatePipelineStageRequest) (dto.PipelineStageResponse, error) {
	// Validate: only one won stage allowed
	if req.IsWon != nil && *req.IsWon {
		existing, err := u.repo.FindWonStage(ctx)
		if err == nil && existing != nil {
			return dto.PipelineStageResponse{}, errors.New("only one pipeline stage can be marked as won")
		}
	}

	// Validate: only one lost stage allowed
	if req.IsLost != nil && *req.IsLost {
		existing, err := u.repo.FindLostStage(ctx)
		if err == nil && existing != nil {
			return dto.PipelineStageResponse{}, errors.New("only one pipeline stage can be marked as lost")
		}
	}

	// Validate: cannot be both won and lost
	if req.IsWon != nil && *req.IsWon && req.IsLost != nil && *req.IsLost {
		return dto.PipelineStageResponse{}, errors.New("a pipeline stage cannot be both won and lost")
	}

	isWon := false
	if req.IsWon != nil {
		isWon = *req.IsWon
	}

	isLost := false
	if req.IsLost != nil {
		isLost = *req.IsLost
	}

	// Extract tenant_id from context
	tenantID, _ := ctx.Value("tenant_id").(string)

	stageID := uuid.New().String()
	stage := &models.PipelineStage{
		ID:          stageID,
		TenantID:    tenantID,
		Name:        req.Name,
		Code:        generatePipelineStageCode(req.Name, stageID),
		Order:       req.Probability,
		Color:       req.Color,
		Probability: req.Probability,
		IsWon:       isWon,
		IsLost:      isLost,
		IsActive:    true,
		Description: req.Description,
	}

	if err := u.repo.Create(ctx, stage); err != nil {
		return dto.PipelineStageResponse{}, err
	}

	return mapper.ToPipelineStageResponse(stage), nil
}

func (u *pipelineStageUsecase) GetByID(ctx context.Context, id string) (dto.PipelineStageResponse, error) {
	stage, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.PipelineStageResponse{}, err
	}
	return mapper.ToPipelineStageResponse(stage), nil
}

func (u *pipelineStageUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.PipelineStageResponse, int64, error) {
	stages, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToPipelineStageResponseList(stages), total, nil
}

func (u *pipelineStageUsecase) Update(ctx context.Context, id string, req dto.UpdatePipelineStageRequest) (dto.PipelineStageResponse, error) {
	stage, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.PipelineStageResponse{}, errors.New("pipeline stage not found")
	}

	if err := u.validateWonLostStageUpdate(ctx, id, req); err != nil {
		return dto.PipelineStageResponse{}, err
	}

	applyPipelineStageUpdate(stage, req)

	// Post-update validation: cannot be both won and lost
	if stage.IsWon && stage.IsLost {
		return dto.PipelineStageResponse{}, errors.New("a pipeline stage cannot be both won and lost")
	}

	if err := u.repo.Update(ctx, stage); err != nil {
		return dto.PipelineStageResponse{}, err
	}

	if req.Probability != nil {
		if err := u.dealRepo.UpdateProbabilityByStageID(ctx, id, *req.Probability); err != nil {
			return dto.PipelineStageResponse{}, err
		}
	}

	return mapper.ToPipelineStageResponse(stage), nil
}

func (u *pipelineStageUsecase) validateWonLostStageUpdate(ctx context.Context, id string, req dto.UpdatePipelineStageRequest) error {
	if req.IsWon != nil && *req.IsWon {
		existing, err := u.repo.FindWonStage(ctx)
		if err == nil && existing != nil && existing.ID != id {
			return errors.New("only one pipeline stage can be marked as won")
		}
	}

	if req.IsLost != nil && *req.IsLost {
		existing, err := u.repo.FindLostStage(ctx)
		if err == nil && existing != nil && existing.ID != id {
			return errors.New("only one pipeline stage can be marked as lost")
		}
	}

	return nil
}

func applyPipelineStageUpdate(stage *models.PipelineStage, req dto.UpdatePipelineStageRequest) {
	if req.Name != "" {
		stage.Name = req.Name
	}
	if req.Color != "" {
		stage.Color = req.Color
	}
	if req.Probability != nil {
		stage.Probability = *req.Probability
		stage.Order = *req.Probability
	}
	if req.IsWon != nil {
		stage.IsWon = *req.IsWon
	}
	if req.IsLost != nil {
		stage.IsLost = *req.IsLost
	}
	if req.IsActive != nil {
		stage.IsActive = *req.IsActive
	}
	if req.Description != "" {
		stage.Description = req.Description
	}
}

func (u *pipelineStageUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("pipeline stage not found")
	}
	return u.repo.Delete(ctx, id)
}

func generatePipelineStageCode(name, stageID string) string {
	normalized := strings.ToUpper(strings.TrimSpace(name))

	b := strings.Builder{}
	prevUnderscore := false
	for _, ch := range normalized {
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			b.WriteRune(ch)
			prevUnderscore = false
			continue
		}

		if b.Len() > 0 && !prevUnderscore {
			b.WriteByte('_')
			prevUnderscore = true
		}
	}

	base := strings.Trim(b.String(), "_")
	if base == "" {
		base = "STAGE"
	}

	suffix := strings.Split(stageID, "-")[0]
	maxBaseLen := 50 - 1 - len(suffix)
	if maxBaseLen < 1 {
		maxBaseLen = 1
	}
	if len(base) > maxBaseLen {
		base = base[:maxBaseLen]
	}

	return fmt.Sprintf("%s-%s", base, suffix)
}
