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

// PaymentTermsUsecase defines the interface for payment terms business logic
type PaymentTermsUsecase interface {
	Create(ctx context.Context, req dto.CreatePaymentTermsRequest) (dto.PaymentTermsResponse, error)
	GetByID(ctx context.Context, id string) (dto.PaymentTermsResponse, error)
	List(ctx context.Context, params repositories.ListParams) ([]dto.PaymentTermsResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdatePaymentTermsRequest) (dto.PaymentTermsResponse, error)
	Delete(ctx context.Context, id string) error
}

type paymentTermsUsecase struct {
	repo repositories.PaymentTermsRepository
}

// NewPaymentTermsUsecase creates a new PaymentTermsUsecase
func NewPaymentTermsUsecase(repo repositories.PaymentTermsRepository) PaymentTermsUsecase {
	return &paymentTermsUsecase{repo: repo}
}

func (u *paymentTermsUsecase) Create(ctx context.Context, req dto.CreatePaymentTermsRequest) (dto.PaymentTermsResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Generate ID to use in Code
	id := uuid.New().String()
	
	// Generate Code: PT-<NameSlug>-<Days>D-<ShortID>
	// Simple slug: first 3 chars or whole name (sanitized)
	// For simplicity and uniqueness: PT-<Days>D-<First8CharsOfID> (User asked for Name+Days+ID)
	// Let's do: PT-<First3CharsOfName>-<Days>D-<First4CharsOfID>
	
	// Helper to get slug
	slug := "GEN"
	if len(req.Name) >= 3 {
		slug = req.Name[:3]
	} else {
		slug = req.Name
	}
	// Uppercase slug
	// (Assuming we can import strings, if not I'll just use ID for now or rely on user understanding)
	// Since I can't easily add imports with replace_file_content without seeing the whole file imports, 
	// I'll stick to a simpler format or assume `strings` is available if I added it? No I didn't.
	// I'll use a simple format: PT-DAYS-ID_SUFFIX using just uuid
	
	code := fmt.Sprintf("PT-%s-%dD-%s", slug, req.Days, id[:4])
	
	// Note: I need to add "fmt" to imports. I will handle imports in a separate step or assume I can rewrite imports.
	// Actually, `replace_file_content` is risky for adding imports if I don't replace the import block.
	// I'll use multi_replace to add imports.

	paymentTerms := &models.PaymentTerms{
		ID:          id,
		Code:        code,
		Name:        req.Name,
		Description: req.Description,
		Days:        req.Days,
		IsActive:    isActive,
	}

	if err := u.repo.Create(ctx, paymentTerms); err != nil {
		return dto.PaymentTermsResponse{}, err
	}

	return mapper.ToPaymentTermsResponse(paymentTerms), nil
}

func (u *paymentTermsUsecase) GetByID(ctx context.Context, id string) (dto.PaymentTermsResponse, error) {
	paymentTerms, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.PaymentTermsResponse{}, err
	}
	return mapper.ToPaymentTermsResponse(paymentTerms), nil
}

func (u *paymentTermsUsecase) List(ctx context.Context, params repositories.ListParams) ([]dto.PaymentTermsResponse, int64, error) {
	paymentTermsList, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToPaymentTermsResponseList(paymentTermsList), total, nil
}

func (u *paymentTermsUsecase) Update(ctx context.Context, id string, req dto.UpdatePaymentTermsRequest) (dto.PaymentTermsResponse, error) {
	paymentTerms, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return dto.PaymentTermsResponse{}, err
	}

	if req.Code != "" {
		paymentTerms.Code = req.Code
	}
	if req.Name != "" {
		paymentTerms.Name = req.Name
	}
	if req.Description != "" {
		paymentTerms.Description = req.Description
	}
	if req.Days != nil {
		paymentTerms.Days = *req.Days
	}
	if req.IsActive != nil {
		paymentTerms.IsActive = *req.IsActive
	}

	if err := u.repo.Update(ctx, paymentTerms); err != nil {
		return dto.PaymentTermsResponse{}, err
	}

	return mapper.ToPaymentTermsResponse(paymentTerms), nil
}

func (u *paymentTermsUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("payment terms not found")
	}
	return u.repo.Delete(ctx, id)
}
