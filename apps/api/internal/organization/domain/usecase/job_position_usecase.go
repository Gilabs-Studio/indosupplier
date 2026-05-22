package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrJobPositionNotFound      = errors.New("job position not found")
	ErrJobPositionAlreadyExists = errors.New("job position with this name already exists")
)

// JobPositionUsecase defines the interface for job position business logic
type JobPositionUsecase interface {
	List(ctx context.Context, req *dto.ListJobPositionsRequest) ([]dto.JobPositionResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.JobPositionResponse, error)
	Create(ctx context.Context, req *dto.CreateJobPositionRequest) (*dto.JobPositionResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateJobPositionRequest) (*dto.JobPositionResponse, error)
	Delete(ctx context.Context, id string) error
}

type jobPositionUsecase struct {
	jobPositionRepo repositories.JobPositionRepository
}

// NewJobPositionUsecase creates a new JobPositionUsecase
func NewJobPositionUsecase(jobPositionRepo repositories.JobPositionRepository) JobPositionUsecase {
	return &jobPositionUsecase{jobPositionRepo: jobPositionRepo}
}

func (u *jobPositionUsecase) List(ctx context.Context, req *dto.ListJobPositionsRequest) ([]dto.JobPositionResponse, *utils.PaginationResult, error) {
	jobPositions, total, err := u.jobPositionRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := mapper.ToJobPositionResponses(jobPositions)

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

	pagination := &utils.PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	}

	return responses, pagination, nil
}

func (u *jobPositionUsecase) GetByID(ctx context.Context, id string) (*dto.JobPositionResponse, error) {
	jobPosition, err := u.jobPositionRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrJobPositionNotFound
		}
		return nil, err
	}

	return mapper.ToJobPositionResponse(jobPosition), nil
}

func (u *jobPositionUsecase) Create(ctx context.Context, req *dto.CreateJobPositionRequest) (*dto.JobPositionResponse, error) {
	existing, err := u.jobPositionRepo.FindByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrJobPositionAlreadyExists
	}

	jobPosition := mapper.JobPositionFromCreateRequest(req)
	if err := u.jobPositionRepo.Create(ctx, jobPosition); err != nil {
		return nil, err
	}

	return mapper.ToJobPositionResponse(jobPosition), nil
}

func (u *jobPositionUsecase) Update(ctx context.Context, id string, req *dto.UpdateJobPositionRequest) (*dto.JobPositionResponse, error) {
	jobPosition, err := u.jobPositionRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrJobPositionNotFound
		}
		return nil, err
	}

	if req.Name != "" && req.Name != jobPosition.Name {
		existing, err := u.jobPositionRepo.FindByName(ctx, req.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrJobPositionAlreadyExists
		}
		jobPosition.Name = req.Name
	}

	if req.Description != "" {
		jobPosition.Description = req.Description
	}
	if req.IsActive != nil {
		jobPosition.IsActive = *req.IsActive
	}

	if err := u.jobPositionRepo.Update(ctx, jobPosition); err != nil {
		return nil, err
	}

	return mapper.ToJobPositionResponse(jobPosition), nil
}

func (u *jobPositionUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.jobPositionRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrJobPositionNotFound
		}
		return err
	}

	return u.jobPositionRepo.Delete(ctx, id)
}
