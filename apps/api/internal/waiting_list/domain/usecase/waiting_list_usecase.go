package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/gilabs/indosupplier/api/internal/waiting_list/data/models"
	"github.com/gilabs/indosupplier/api/internal/waiting_list/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/waiting_list/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/waiting_list/domain/mapper"
)

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered in waiting list")
	ErrEntryNotFound          = errors.New("waiting list entry not found")
)

type WaitingListUsecase interface {
	Join(ctx context.Context, req dto.JoinWaitingListRequest) (dto.WaitingListResponse, error)
	List(ctx context.Context, limit, offset int, status string) ([]dto.WaitingListResponse, int64, error)
	UpdateStatus(ctx context.Context, id string, req dto.UpdateWaitingListStatusRequest) (dto.WaitingListResponse, error)
	Delete(ctx context.Context, id string) error
}

type waitingListUsecase struct {
	repo repositories.WaitingListRepository
}

func NewWaitingListUsecase(repo repositories.WaitingListRepository) WaitingListUsecase {
	return &waitingListUsecase{repo: repo}
}

func (u *waitingListUsecase) Join(ctx context.Context, req dto.JoinWaitingListRequest) (dto.WaitingListResponse, error) {
	// Check if already registered
	existing, err := u.repo.FindByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return dto.WaitingListResponse{}, ErrEmailAlreadyRegistered
	}

	w := &models.WaitingList{
		ID:          uuid.New().String(),
		Email:       req.Email,
		Name:        req.Name,
		CompanyName: req.CompanyName,
		CompanyType: req.CompanyType,
		Phone:       req.Phone,
		Notes:       req.Notes,
		Status:      "pending",
	}

	if err := u.repo.Create(ctx, w); err != nil {
		return dto.WaitingListResponse{}, err
	}

	return mapper.ToWaitingListResponse(w), nil
}

func (u *waitingListUsecase) List(ctx context.Context, limit, offset int, status string) ([]dto.WaitingListResponse, int64, error) {
	items, total, err := u.repo.List(ctx, limit, offset, status)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToWaitingListResponses(items), total, nil
}

func (u *waitingListUsecase) UpdateStatus(ctx context.Context, id string, req dto.UpdateWaitingListStatusRequest) (dto.WaitingListResponse, error) {
	w, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.WaitingListResponse{}, ErrEntryNotFound
	}

	w.Status = req.Status

	if err := u.repo.Update(ctx, w); err != nil {
		return dto.WaitingListResponse{}, err
	}

	return mapper.ToWaitingListResponse(w), nil
}

func (u *waitingListUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return ErrEntryNotFound
	}
	return u.repo.Delete(ctx, id)
}
