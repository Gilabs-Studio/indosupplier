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

// LeadStatusUsecase defines the interface for lead status business logic
type LeadStatusUsecase interface {
	Create(ctx context.Context, req dto.CreateLeadStatusRequest) (dto.LeadStatusResponse, error)
	GetByID(ctx context.Context, id string) (dto.LeadStatusResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.LeadStatusResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateLeadStatusRequest) (dto.LeadStatusResponse, error)
	Delete(ctx context.Context, id string) error
}

type leadStatusUsecase struct {
	repo repositories.LeadStatusRepository
}

// NewLeadStatusUsecase creates a new lead status usecase
func NewLeadStatusUsecase(repo repositories.LeadStatusRepository) LeadStatusUsecase {
	return &leadStatusUsecase{repo: repo}
}

func (u *leadStatusUsecase) Create(ctx context.Context, req dto.CreateLeadStatusRequest) (dto.LeadStatusResponse, error) {
	nextOrder, err := u.nextLeadStatusOrder(ctx)
	if err != nil {
		return dto.LeadStatusResponse{}, err
	}

	statusID := uuid.New().String()
	// Extract tenant_id from context
	tenantID, _ := ctx.Value("tenant_id").(string)
	status := &models.LeadStatus{
		ID:          statusID,
		TenantID:    tenantID,
		Name:        req.Name,
		Code:        generateLeadStatusCode(req.Name, statusID),
		Description: req.Description,
		Score:       req.Score,
		Color:       req.Color,
		Order:       nextOrder,
		IsActive:    true,
	}

	if err := u.repo.Create(ctx, status); err != nil {
		return dto.LeadStatusResponse{}, err
	}

	return mapper.ToLeadStatusResponse(status), nil
}

func (u *leadStatusUsecase) GetByID(ctx context.Context, id string) (dto.LeadStatusResponse, error) {
	status, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.LeadStatusResponse{}, err
	}
	return mapper.ToLeadStatusResponse(status), nil
}

func (u *leadStatusUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.LeadStatusResponse, int64, error) {
	statuses, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToLeadStatusResponseList(statuses), total, nil
}

func (u *leadStatusUsecase) Update(ctx context.Context, id string, req dto.UpdateLeadStatusRequest) (dto.LeadStatusResponse, error) {
	status, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.LeadStatusResponse{}, errors.New("lead status not found")
	}

	applyLeadStatusUpdate(status, req)

	if err := u.repo.Update(ctx, status); err != nil {
		return dto.LeadStatusResponse{}, err
	}

	return mapper.ToLeadStatusResponse(status), nil
}
func applyLeadStatusUpdate(status *models.LeadStatus, req dto.UpdateLeadStatusRequest) {
	if req.Name != "" {
		status.Name = req.Name
	}
	if req.Description != "" {
		status.Description = req.Description
	}
	if req.Score != nil {
		status.Score = *req.Score
	}
	if req.Color != "" {
		status.Color = req.Color
	}
	if req.IsActive != nil {
		status.IsActive = *req.IsActive
	}
}

func (u *leadStatusUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("lead status not found")
	}
	return u.repo.Delete(ctx, id)
}

func (u *leadStatusUsecase) nextLeadStatusOrder(ctx context.Context) (int, error) {
	maxOrder, err := u.repo.GetMaxOrder(ctx)
	if err != nil {
		return 0, err
	}
	return maxOrder + 1, nil
}

func generateLeadStatusCode(name, statusID string) string {
	base := normalizeCodeBase(name, "STATUS")
	return fmt.Sprintf("%s-%s", base, strings.Split(statusID, "-")[0])
}
