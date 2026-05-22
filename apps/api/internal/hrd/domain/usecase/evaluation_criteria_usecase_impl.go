package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/mapper"
	"github.com/google/uuid"
)

type evaluationCriteriaUsecase struct {
	criteriaRepo repositories.EvaluationCriteriaRepository
	groupRepo    repositories.EvaluationGroupRepository
	auditService audit.AuditService
}

const errEvaluationCriteriaNotFound = "evaluation criteria not found"

// NewEvaluationCriteriaUsecase creates a new instance of EvaluationCriteriaUsecase
func NewEvaluationCriteriaUsecase(
	criteriaRepo repositories.EvaluationCriteriaRepository,
	groupRepo repositories.EvaluationGroupRepository,
	auditService audit.AuditService,
) EvaluationCriteriaUsecase {
	return &evaluationCriteriaUsecase{
		criteriaRepo: criteriaRepo,
		groupRepo:    groupRepo,
		auditService: auditService,
	}
}

func (u *evaluationCriteriaUsecase) GetByGroupID(ctx context.Context, groupID string) ([]*dto.EvaluationCriteriaResponse, error) {
	// Validate group exists
	group, err := u.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("evaluation group not found")
	}

	criteria, err := u.criteriaRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch evaluation criteria: %w", err)
	}

	return mapper.ToEvaluationCriteriaResponseList(criteria), nil
}

func (u *evaluationCriteriaUsecase) GetByID(ctx context.Context, id string) (*dto.EvaluationCriteriaResponse, error) {
	criteria, err := u.criteriaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if criteria == nil {
		return nil, errors.New(errEvaluationCriteriaNotFound)
	}

	return mapper.ToEvaluationCriteriaResponse(criteria), nil
}

func (u *evaluationCriteriaUsecase) Create(ctx context.Context, req *dto.CreateEvaluationCriteriaRequest) (*dto.EvaluationCriteriaResponse, error) {
	// Validate group exists
	group, err := u.groupRepo.FindByID(ctx, req.EvaluationGroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("evaluation group not found")
	}

	// Validate total weight does not exceed 100%
	currentWeight, err := u.criteriaRepo.GetTotalWeightByGroupID(ctx, req.EvaluationGroupID, "")
	if err != nil {
		return nil, err
	}
	if currentWeight+req.Weight > 100 {
		return nil, fmt.Errorf("total weight would exceed 100%%. Current: %.2f%%, Adding: %.2f%%", currentWeight, req.Weight)
	}

	id := uuid.New().String()
	criteria := mapper.ToEvaluationCriteriaModel(req, id)

	if err := u.criteriaRepo.Create(ctx, criteria); err != nil {
		return nil, fmt.Errorf("failed to create evaluation criteria: %w", err)
	}

	u.auditService.Log(ctx, "evaluation_criteria.create", criteria.EvaluationGroupID, map[string]interface{}{"after": criteria})

	return mapper.ToEvaluationCriteriaResponse(criteria), nil
}

func (u *evaluationCriteriaUsecase) Update(ctx context.Context, id string, req *dto.UpdateEvaluationCriteriaRequest) (*dto.EvaluationCriteriaResponse, error) {
	criteria, err := u.criteriaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if criteria == nil {
		return nil, errors.New(errEvaluationCriteriaNotFound)
	}
	before := *criteria

	// Validate weight if being updated
	if req.Weight != nil {
		// Get total weight excluding the current criteria
		otherWeight, err := u.criteriaRepo.GetTotalWeightByGroupID(ctx, criteria.EvaluationGroupID, id)
		if err != nil {
			return nil, err
		}
		if otherWeight+*req.Weight > 100 {
			return nil, fmt.Errorf("total weight would exceed 100%%. Other criteria: %.2f%%, New weight: %.2f%%", otherWeight, *req.Weight)
		}
	}

	mapper.UpdateEvaluationCriteriaModel(criteria, req)

	if err := u.criteriaRepo.Update(ctx, criteria); err != nil {
		return nil, fmt.Errorf("failed to update evaluation criteria: %w", err)
	}

	u.auditService.Log(ctx, "evaluation_criteria.update", criteria.EvaluationGroupID, map[string]interface{}{
		"before": before,
		"after":  criteria,
	})

	return mapper.ToEvaluationCriteriaResponse(criteria), nil
}

func (u *evaluationCriteriaUsecase) Delete(ctx context.Context, id string) error {
	criteria, err := u.criteriaRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if criteria == nil {
		return errors.New(errEvaluationCriteriaNotFound)
	}

	if err := u.criteriaRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete evaluation criteria: %w", err)
	}

	u.auditService.Log(ctx, "evaluation_criteria.delete", criteria.EvaluationGroupID, map[string]interface{}{"before": criteria})

	return nil
}
