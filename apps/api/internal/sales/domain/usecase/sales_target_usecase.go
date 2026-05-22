package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	salesRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"github.com/gilabs/gims/api/internal/sales/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrSalesTargetNotFound = errors.New("sales target not found")
	ErrSalesTargetConflict = errors.New("sales target for selected year and employee already exists")
)

// SalesTargetUsecase defines the interface for sales target business logic
type SalesTargetUsecase interface {
	List(ctx context.Context, req *dto.ListSalesTargetsRequest) ([]dto.SalesTargetResponse, *utils.PaginationResult, error)
	ListAvailableEmployees(ctx context.Context, req *dto.ListAvailableSalesTargetEmployeesRequest) ([]dto.EmployeeResponse, error)
	GetByID(ctx context.Context, id string) (*dto.SalesTargetResponse, error)
	Create(ctx context.Context, req *dto.CreateSalesTargetRequest) (*dto.SalesTargetResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateSalesTargetRequest) (*dto.SalesTargetResponse, error)
	Delete(ctx context.Context, id string) error
}

type salesTargetUsecase struct {
	db         *gorm.DB
	targetRepo salesRepos.SalesTargetRepository
}

// NewSalesTargetUsecase creates a new SalesTargetUsecase
func NewSalesTargetUsecase(db *gorm.DB, targetRepo salesRepos.SalesTargetRepository) SalesTargetUsecase {
	return &salesTargetUsecase{db: db, targetRepo: targetRepo}
}

func (u *salesTargetUsecase) List(ctx context.Context, req *dto.ListSalesTargetsRequest) ([]dto.SalesTargetResponse, *utils.PaginationResult, error) {
	targets, total, err := u.targetRepo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]dto.SalesTargetResponse, len(targets))
	for i := range targets {
		responses[i] = mapper.ToSalesTargetResponse(&targets[i])
	}

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

func (u *salesTargetUsecase) ListAvailableEmployees(ctx context.Context, req *dto.ListAvailableSalesTargetEmployeesRequest) ([]dto.EmployeeResponse, error) {
	return u.targetRepo.ListAvailableEmployeesByYear(ctx, req.Year, req.IncludeEmployeeID)
}

func (u *salesTargetUsecase) GetByID(ctx context.Context, id string) (*dto.SalesTargetResponse, error) {
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.SalesTarget{}, id, security.SalesTargetScopeQueryOptions()) {
		return nil, ErrSalesTargetNotFound
	}

	target, err := u.targetRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSalesTargetNotFound
		}
		return nil, err
	}

	response := mapper.ToSalesTargetResponse(target)
	return &response, nil
}

func (u *salesTargetUsecase) Create(ctx context.Context, req *dto.CreateSalesTargetRequest) (*dto.SalesTargetResponse, error) {
	exists, err := u.targetRepo.ExistsByYearAndEmployee(ctx, req.Year, req.EmployeeID, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSalesTargetConflict
	}

	target := mapper.ToSalesTargetModel(req)
	if err := u.targetRepo.Create(ctx, target); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") || strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return nil, ErrSalesTargetConflict
		}
		return nil, err
	}

	created, err := u.targetRepo.FindByID(ctx, target.ID)
	if err != nil {
		return nil, err
	}

	response := mapper.ToSalesTargetResponse(created)
	return &response, nil
}

func (u *salesTargetUsecase) Update(ctx context.Context, id string, req *dto.UpdateSalesTargetRequest) (*dto.SalesTargetResponse, error) {
	target, err := u.targetRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSalesTargetNotFound
		}
		return nil, err
	}

	mapper.UpdateSalesTargetModel(target, req)

	if err := u.targetRepo.Update(ctx, target); err != nil {
		return nil, err
	}

	updated, err := u.targetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := mapper.ToSalesTargetResponse(updated)
	return &response, nil
}

func (u *salesTargetUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.targetRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSalesTargetNotFound
		}
		return err
	}

	return u.targetRepo.Delete(ctx, id)
}
