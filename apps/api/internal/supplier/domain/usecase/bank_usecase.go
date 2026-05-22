package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/supplier/data/models"
	"github.com/gilabs/gims/api/internal/supplier/data/repositories"
	"github.com/gilabs/gims/api/internal/supplier/domain/dto"
	"github.com/gilabs/gims/api/internal/supplier/domain/mapper"
	"github.com/google/uuid"
)

// BankUsecase defines the interface for bank business logic
type BankUsecase interface {
	Create(ctx context.Context, req dto.CreateBankRequest) (dto.BankResponse, error)
	GetByID(ctx context.Context, id string) (dto.BankResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.BankResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateBankRequest) (dto.BankResponse, error)
	Delete(ctx context.Context, id string) error
}

type bankUsecase struct {
	repo repositories.BankRepository
}

// NewBankUsecase creates a new BankUsecase
func NewBankUsecase(repo repositories.BankRepository) BankUsecase {
	return &bankUsecase{repo: repo}
}

func (u *bankUsecase) Create(ctx context.Context, req dto.CreateBankRequest) (dto.BankResponse, error) {
	// Check if code already exists
	existing, _ := u.repo.FindByCode(ctx, req.Code)
	if existing != nil {
		return dto.BankResponse{}, errors.New("bank code already exists")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	bank := &models.Bank{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Code:      req.Code,
		SwiftCode: req.SwiftCode,
		IsActive:  isActive,
	}

	if err := u.repo.Create(ctx, bank); err != nil {
		return dto.BankResponse{}, err
	}

	return mapper.ToBankResponse(bank), nil
}

func (u *bankUsecase) GetByID(ctx context.Context, id string) (dto.BankResponse, error) {
	bank, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.BankResponse{}, err
	}
	return mapper.ToBankResponse(bank), nil
}

func (u *bankUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.BankResponse, int64, error) {
	banks, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToBankResponseList(banks), total, nil
}

func (u *bankUsecase) Update(ctx context.Context, id string, req dto.UpdateBankRequest) (dto.BankResponse, error) {
	bank, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.BankResponse{}, err
	}

	if req.Name != "" {
		bank.Name = req.Name
	}
	if req.Code != "" {
		// Check if new code conflicts with existing
		existing, _ := u.repo.FindByCode(ctx, req.Code)
		if existing != nil && existing.ID != id {
			return dto.BankResponse{}, errors.New("bank code already exists")
		}
		bank.Code = req.Code
	}
	if req.SwiftCode != "" {
		bank.SwiftCode = req.SwiftCode
	}
	if req.IsActive != nil {
		bank.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, bank); err != nil {
		return dto.BankResponse{}, err
	}

	return mapper.ToBankResponse(bank), nil
}

func (u *bankUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("bank not found")
	}
	return u.repo.Delete(ctx, id)
}
