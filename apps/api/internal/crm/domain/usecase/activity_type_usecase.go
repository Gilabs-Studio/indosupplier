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

// ActivityTypeUsecase defines the interface for activity type business logic
type ActivityTypeUsecase interface {
	Create(ctx context.Context, req dto.CreateActivityTypeRequest) (dto.ActivityTypeResponse, error)
	GetByID(ctx context.Context, id string) (dto.ActivityTypeResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.ActivityTypeResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateActivityTypeRequest) (dto.ActivityTypeResponse, error)
	Delete(ctx context.Context, id string) error
}

type activityTypeUsecase struct {
	repo repositories.ActivityTypeRepository
}

// NewActivityTypeUsecase creates a new activity type usecase
func NewActivityTypeUsecase(repo repositories.ActivityTypeRepository) ActivityTypeUsecase {
	return &activityTypeUsecase{repo: repo}
}

func (u *activityTypeUsecase) Create(ctx context.Context, req dto.CreateActivityTypeRequest) (dto.ActivityTypeResponse, error) {
	nextOrder, err := u.nextActivityTypeOrder(ctx)
	if err != nil {
		return dto.ActivityTypeResponse{}, err
	}

	activityTypeID := uuid.New().String()

	// Extract tenant_id from context
	tenantID, _ := ctx.Value("tenant_id").(string)

	actType := &models.ActivityType{
		ID:          activityTypeID,
		TenantID:    tenantID,
		Name:        req.Name,
		Code:        generateActivityTypeCode(req.Name, activityTypeID),
		Description: req.Description,
		Icon:        req.Icon,
		BadgeColor:  req.BadgeColor,
		Order:       nextOrder,
		IsActive:    true,
	}

	if err := u.repo.Create(ctx, actType); err != nil {
		return dto.ActivityTypeResponse{}, err
	}

	return mapper.ToActivityTypeResponse(actType), nil
}

func (u *activityTypeUsecase) GetByID(ctx context.Context, id string) (dto.ActivityTypeResponse, error) {
	actType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ActivityTypeResponse{}, err
	}
	return mapper.ToActivityTypeResponse(actType), nil
}

func (u *activityTypeUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.ActivityTypeResponse, int64, error) {
	actTypes, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToActivityTypeResponseList(actTypes), total, nil
}

func (u *activityTypeUsecase) Update(ctx context.Context, id string, req dto.UpdateActivityTypeRequest) (dto.ActivityTypeResponse, error) {
	actType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ActivityTypeResponse{}, errors.New("activity type not found")
	}

	if req.Name != "" {
		actType.Name = req.Name
	}
	if req.Description != "" {
		actType.Description = req.Description
	}
	if req.Icon != "" {
		actType.Icon = req.Icon
	}
	if req.BadgeColor != "" {
		actType.BadgeColor = req.BadgeColor
	}
	if req.IsActive != nil {
		actType.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, actType); err != nil {
		return dto.ActivityTypeResponse{}, err
	}

	return mapper.ToActivityTypeResponse(actType), nil
}

func (u *activityTypeUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("activity type not found")
	}
	return u.repo.Delete(ctx, id)
}

func (u *activityTypeUsecase) nextActivityTypeOrder(ctx context.Context) (int, error) {
	maxOrder, err := u.repo.GetMaxOrder(ctx)
	if err != nil {
		return 0, err
	}
	return maxOrder + 1, nil
}

func generateActivityTypeCode(name, activityTypeID string) string {
	base := normalizeCodeBase(name, "ACTIVITY")
	return fmt.Sprintf("%s-%s", base, strings.Split(activityTypeID, "-")[0])
}
