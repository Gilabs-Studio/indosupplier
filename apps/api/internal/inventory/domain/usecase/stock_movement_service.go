package usecase

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/inventory/data/models"
	"github.com/gilabs/gims/api/internal/inventory/domain/dto"
	"github.com/gilabs/gims/api/internal/inventory/domain/repository"
)

type StockMovementService interface {
	GetMovements(ctx context.Context, req *dto.GetStockMovementsRequest) ([]models.StockMovement, *response.PaginationMeta, error)
}

type stockMovementService struct {
	repo repository.StockMovementRepository
}

func NewStockMovementService(repo repository.StockMovementRepository) StockMovementService {
	return &stockMovementService{
		repo: repo,
	}
}

func (s *stockMovementService) GetMovements(ctx context.Context, req *dto.GetStockMovementsRequest) ([]models.StockMovement, *response.PaginationMeta, error) {
	movements, total, err := s.repo.FindAll(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	meta := response.NewPaginationMeta(req.Page, req.PerPage, int(total))
	return movements, meta, nil
}
