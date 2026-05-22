package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/domain/dto"
	"github.com/gilabs/gims/api/internal/core/domain/mapper"
	"github.com/google/uuid"
)

// LeaveTypeUsecase defines the interface for leave type business logic
type LeaveTypeUsecase interface {
	Create(ctx context.Context, req dto.CreateLeaveTypeRequest) (dto.LeaveTypeResponse, error)
	GetByID(ctx context.Context, id string) (dto.LeaveTypeResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.LeaveTypeResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateLeaveTypeRequest) (dto.LeaveTypeResponse, error)
	Delete(ctx context.Context, id string) error
}

type leaveTypeUsecase struct {
	repo repositories.LeaveTypeRepository
}

// NewLeaveTypeUsecase creates a new LeaveTypeUsecase
func NewLeaveTypeUsecase(repo repositories.LeaveTypeRepository) LeaveTypeUsecase {
	return &leaveTypeUsecase{repo: repo}
}

func (u *leaveTypeUsecase) Create(ctx context.Context, req dto.CreateLeaveTypeRequest) (dto.LeaveTypeResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	isPaid := true
	if req.IsPaid != nil {
		isPaid = *req.IsPaid
	}

	id := uuid.New().String()
	slug := "GEN"
	if len(req.Name) >= 3 {
		slug = req.Name[:3]
	} else {
		slug = req.Name
	}
	code := fmt.Sprintf("LT-%s-%s", slug, id[:4])

	leaveType := &models.LeaveType{
		ID:          id,
		Code:        code,
		Name:        req.Name,
		Description: req.Description,
		MaxDays:     req.MaxDays,
		IsPaid:      isPaid,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, leaveType); err != nil {
		return dto.LeaveTypeResponse{}, err
	}

	return mapper.ToLeaveTypeResponse(leaveType), nil
}

func (u *leaveTypeUsecase) GetByID(ctx context.Context, id string) (dto.LeaveTypeResponse, error) {
	leaveType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.LeaveTypeResponse{}, err
	}
	return mapper.ToLeaveTypeResponse(leaveType), nil
}

func (u *leaveTypeUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.LeaveTypeResponse, int64, error) {
	leaveTypes, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToLeaveTypeResponseList(leaveTypes), total, nil
}

func (u *leaveTypeUsecase) Update(ctx context.Context, id string, req dto.UpdateLeaveTypeRequest) (dto.LeaveTypeResponse, error) {
	leaveType, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.LeaveTypeResponse{}, err
	}

	if req.Code != "" {
		leaveType.Code = req.Code
	}
	if req.Name != "" {
		leaveType.Name = req.Name
	}
	if req.Description != "" {
		leaveType.Description = req.Description
	}
	if req.MaxDays != nil {
		leaveType.MaxDays = *req.MaxDays
	}
	if req.IsPaid != nil {
		leaveType.IsPaid = *req.IsPaid
	}
	if req.IsActive != nil {
		leaveType.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, leaveType); err != nil {
		return dto.LeaveTypeResponse{}, err
	}

	return mapper.ToLeaveTypeResponse(leaveType), nil
}

func (u *leaveTypeUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("leave type not found")
	}
	return u.repo.Delete(ctx, id)
}
