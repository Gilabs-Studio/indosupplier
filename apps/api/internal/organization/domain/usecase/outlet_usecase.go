package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/middleware"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/gilabs/gims/api/internal/organization/domain/mapper"
	roleModels "github.com/gilabs/gims/api/internal/role/data/models"
	tenantModels "github.com/gilabs/gims/api/internal/tenant/data/models"
	warehouseModels "github.com/gilabs/gims/api/internal/warehouse/data/models"
	warehouseRepo "github.com/gilabs/gims/api/internal/warehouse/data/repositories"
	"gorm.io/gorm"
)

var (
	ErrOutletNotFound     = errors.New("OUTLET_NOT_FOUND")
	ErrOutletLimitReached = errors.New("OUTLET_LIMIT_REACHED")
)

// OutletLimitResponse carries current outlet count and max allowed by subscription.
type OutletLimitResponse struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

type subscriptionFinder interface {
	FindActiveByTenantID(ctx context.Context, tenantID string) (*tenantModels.TenantSubscription, error)
}

// OutletUsecase defines business logic for outlets
type OutletUsecase interface {
	Create(ctx context.Context, req dto.CreateOutletRequest) (*dto.OutletResponse, error)
	GetByID(ctx context.Context, id string) (*dto.OutletResponse, error)
	List(ctx context.Context, params repositories.OutletListParams) ([]*dto.OutletResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateOutletRequest) (*dto.OutletResponse, error)
	Delete(ctx context.Context, id string) error
	GetFormData(ctx context.Context) (*dto.OutletFormDataResponse, error)
	GetLimit(ctx context.Context) (*OutletLimitResponse, error)
}

type outletUsecase struct {
	db               *gorm.DB
	repo             repositories.OutletRepository
	warehouseRepo    warehouseRepo.WarehouseRepository
	employeeRepo     repositories.EmployeeRepository
	companyRepo      repositories.CompanyRepository
	subscriptionRepo subscriptionFinder
	mapper           *mapper.OutletMapper
}

func (uc *outletUsecase) syncLinkedWarehousesStatus(tx *gorm.DB, outletID string, warehouseID *string, isActive bool) error {
	warehouseQuery := tx.Model(&warehouseModels.Warehouse{}).Where("outlet_id = ?", outletID)
	if warehouseID != nil {
		warehouseQuery = warehouseQuery.Or("id = ?", *warehouseID)
	}

	if err := warehouseQuery.Update("is_active", isActive).Error; err != nil {
		if isActive {
			return fmt.Errorf("failed to activate linked warehouse: %w", err)
		}
		return fmt.Errorf("failed to deactivate linked warehouse: %w", err)
	}

	return nil
}

// NewOutletUsecase creates a new outlet usecase
func NewOutletUsecase(
	db *gorm.DB,
	repo repositories.OutletRepository,
	whRepo warehouseRepo.WarehouseRepository,
	employeeRepo repositories.EmployeeRepository,
	companyRepo repositories.CompanyRepository,
	subscriptionRepo subscriptionFinder,
) OutletUsecase {
	return &outletUsecase{
		db:               db,
		repo:             repo,
		warehouseRepo:    whRepo,
		employeeRepo:     employeeRepo,
		companyRepo:      companyRepo,
		subscriptionRepo: subscriptionRepo,
		mapper:           mapper.NewOutletMapper(),
	}
}

func (uc *outletUsecase) GetLimit(ctx context.Context) (*OutletLimitResponse, error) {
	var current int64
	if err := database.GetDB(ctx, uc.db).Model(&orgModels.Outlet{}).Count(&current).Error; err != nil {
		return nil, err
	}

	if middleware.IsSystemAdmin(ctx) {
		return &OutletLimitResponse{Current: int(current), Max: 0}, nil
	}

	tenantID := middleware.TenantFromContext(ctx)
	max := 1
	if uc.subscriptionRepo != nil {
		if sub, err := uc.subscriptionRepo.FindActiveByTenantID(ctx, tenantID); err == nil && sub != nil {
			if sub.OutletLimit > max {
				max = sub.OutletLimit
			}
		}
	}

	if max <= 0 {
		max = 1
	}

	return &OutletLimitResponse{Current: int(current), Max: max}, nil
}

// Create creates a new outlet and optionally a linked warehouse
func (uc *outletUsecase) Create(ctx context.Context, req dto.CreateOutletRequest) (*dto.OutletResponse, error) {
	limit, limitErr := uc.GetLimit(ctx)
	if limitErr != nil {
		return nil, limitErr
	}
	if limit.Max > 0 && limit.Current >= limit.Max {
		return nil, ErrOutletLimitReached
	}

	const maxCodeRetry = 3

	var (
		outlet *orgModels.Outlet
		err    error
	)

	for attempt := 0; attempt < maxCodeRetry; attempt++ {
		code, codeErr := uc.repo.GetNextCode(ctx)
		if codeErr != nil {
			return nil, fmt.Errorf("failed to generate outlet code: %w", codeErr)
		}

		outlet = uc.mapper.FromCreateRequest(req)
		outlet.Code = code

		// Use transaction so outlet + warehouse are atomic
		err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(outlet).Error; err != nil {
				return err
			}

			// Optionally create a linked warehouse
			if req.CreateWarehouse {
				whCode, err := uc.warehouseRepo.GetNextCode(ctx)
				if err != nil {
					return fmt.Errorf("failed to generate warehouse code: %w", err)
				}

				warehouse := &warehouseModels.Warehouse{
					Code:        whCode,
					Name:        fmt.Sprintf("%s Warehouse", outlet.Name),
					Address:     outlet.Address,
					ProvinceID:  outlet.ProvinceID,
					CityID:      outlet.CityID,
					DistrictID:  outlet.DistrictID,
					VillageID:   outlet.VillageID,
					Latitude:    outlet.Latitude,
					Longitude:   outlet.Longitude,
					IsPosOutlet: true,
					OutletID:    &outlet.ID,
					IsActive:    true,
				}

				if err := tx.Create(warehouse).Error; err != nil {
					return fmt.Errorf("failed to create linked warehouse: %w", err)
				}

				// Store warehouse reference on outlet
				outlet.WarehouseID = &warehouse.ID
				if err := tx.Model(outlet).Update("warehouse_id", warehouse.ID).Error; err != nil {
					return err
				}
			}

			return nil
		})

		if err == nil {
			break
		}

		if !isOutletCodeDuplicateError(err) || attempt == maxCodeRetry-1 {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	// Reload with relations
	outlet, err = uc.repo.GetByID(ctx, outlet.ID)
	if err != nil {
		return nil, err
	}

	return uc.mapper.ToResponse(outlet), nil
}

func isOutletCodeDuplicateError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "duplicate key value") &&
		strings.Contains(errMsg, "idx_outlets_code") &&
		strings.Contains(errMsg, "23505")
}

// GetByID retrieves an outlet by ID
func (uc *outletUsecase) GetByID(ctx context.Context, id string) (*dto.OutletResponse, error) {
	outlet, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOutletNotFound
		}
		return nil, err
	}
	return uc.mapper.ToResponse(outlet), nil
}

// List retrieves outlets with pagination
func (uc *outletUsecase) List(ctx context.Context, params repositories.OutletListParams) ([]*dto.OutletResponse, int64, error) {
	outlets, total, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return uc.mapper.ToResponseList(outlets), total, nil
}

// Update updates an existing outlet and synchronizes linked warehouse status
// when outlet is activated or deactivated.
func (uc *outletUsecase) Update(ctx context.Context, id string, req dto.UpdateOutletRequest) (*dto.OutletResponse, error) {
	outlet, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOutletNotFound
		}
		return nil, err
	}

	wasActive := outlet.IsActive
	uc.mapper.ApplyUpdateRequest(outlet, req)

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(outlet).Error; err != nil {
			return err
		}

		// Keep outlet-warehouse active flag in sync on status transitions.
		if wasActive != outlet.IsActive {
			if err := uc.syncLinkedWarehousesStatus(tx, outlet.ID, outlet.WarehouseID, outlet.IsActive); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reload with relations
	outlet, err = uc.repo.GetByID(ctx, outlet.ID)
	if err != nil {
		return nil, err
	}

	return uc.mapper.ToResponse(outlet), nil
}

// Delete soft-deletes an outlet
func (uc *outletUsecase) Delete(ctx context.Context, id string) error {
	if _, err := uc.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOutletNotFound
		}
		return err
	}
	return uc.repo.Delete(ctx, id)
}

// GetFormData returns form data (managers and companies) for outlet create/edit forms.
// Managers are filtered to include employees with management-level roles:
// - Roles with "manager" in the name (case-insensitive)
// - Roles with OUTLET, DIVISION, or AREA data scope (management scopes)
// This ensures outlet managers, team managers, and supervisors are all eligible.
func (uc *outletUsecase) GetFormData(ctx context.Context) (*dto.OutletFormDataResponse, error) {
	// Fetch employees with management-level roles for the manager dropdown
	employees, err := uc.employeeRepo.FindByRoleDataScope(ctx, roleModels.DataScopeOutlet)
	if err != nil {
		return nil, err
	}

	managers := make([]dto.ManagerResponse, 0, len(employees))
	for _, emp := range employees {
		managers = append(managers, dto.ManagerResponse{
			ID:   emp.ID,
			Name: emp.Name,
		})
	}

	// Fetch companies for company dropdown
	companies, err := uc.companyRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	companyOptions := make([]dto.CompanySimpleResponse, 0, len(companies))
	for _, c := range companies {
		companyOptions = append(companyOptions, dto.CompanySimpleResponse{
			ID:   c.ID,
			Name: c.Name,
		})
	}

	return &dto.OutletFormDataResponse{
		Managers:  managers,
		Companies: companyOptions,
	}, nil
}
