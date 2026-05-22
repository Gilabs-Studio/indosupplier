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

// LeadSourceUsecase defines the interface for lead source business logic
type LeadSourceUsecase interface {
	Create(ctx context.Context, req dto.CreateLeadSourceRequest) (dto.LeadSourceResponse, error)
	GetByID(ctx context.Context, id string) (dto.LeadSourceResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.LeadSourceResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateLeadSourceRequest) (dto.LeadSourceResponse, error)
	Delete(ctx context.Context, id string) error
}

type leadSourceUsecase struct {
	repo repositories.LeadSourceRepository
}

// NewLeadSourceUsecase creates a new lead source usecase
func NewLeadSourceUsecase(repo repositories.LeadSourceRepository) LeadSourceUsecase {
	return &leadSourceUsecase{repo: repo}
}

func (u *leadSourceUsecase) Create(ctx context.Context, req dto.CreateLeadSourceRequest) (dto.LeadSourceResponse, error) {
	nextOrder, err := u.nextLeadSourceOrder(ctx)
	if err != nil {
		return dto.LeadSourceResponse{}, err
	}

	// Extract tenant_id from context
	tenantID, _ := ctx.Value("tenant_id").(string)
	sourceID := uuid.New().String()

	source := &models.LeadSource{
		ID:          sourceID,
		TenantID:    tenantID,
		Name:        req.Name,
		Code:        generateLeadSourceCode(req.Name, sourceID),
		Description: req.Description,
		Order:       nextOrder,
		IsActive:    true,
	}

	if err := u.repo.Create(ctx, source); err != nil {
		return dto.LeadSourceResponse{}, err
	}

	return mapper.ToLeadSourceResponse(source), nil
}

func (u *leadSourceUsecase) GetByID(ctx context.Context, id string) (dto.LeadSourceResponse, error) {
	source, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.LeadSourceResponse{}, err
	}
	return mapper.ToLeadSourceResponse(source), nil
}

func (u *leadSourceUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.LeadSourceResponse, int64, error) {
	sources, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToLeadSourceResponseList(sources), total, nil
}

func (u *leadSourceUsecase) Update(ctx context.Context, id string, req dto.UpdateLeadSourceRequest) (dto.LeadSourceResponse, error) {
	source, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.LeadSourceResponse{}, errors.New("lead source not found")
	}

	if req.Name != "" {
		source.Name = req.Name
	}
	if req.Description != "" {
		source.Description = req.Description
	}
	if req.IsActive != nil {
		source.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, source); err != nil {
		return dto.LeadSourceResponse{}, err
	}

	return mapper.ToLeadSourceResponse(source), nil
}

func (u *leadSourceUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("lead source not found")
	}
	return u.repo.Delete(ctx, id)
}

func (u *leadSourceUsecase) nextLeadSourceOrder(ctx context.Context) (int, error) {
	maxOrder, err := u.repo.GetMaxOrder(ctx)
	if err != nil {
		return 0, err
	}
	return maxOrder + 1, nil
}

func generateLeadSourceCode(name, sourceID string) string {
	base := normalizeCodeBase(name, "SOURCE")
	return fmt.Sprintf("%s-%s", base, strings.Split(sourceID, "-")[0])
}

func normalizeCodeBase(name, fallback string) string {
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
		return fallback
	}

	if len(base) > 41 {
		return base[:41]
	}

	return base
}
