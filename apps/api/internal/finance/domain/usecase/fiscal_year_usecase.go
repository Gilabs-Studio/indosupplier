package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/middleware"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"gorm.io/gorm"
)

var (
	ErrFiscalYearNotFound                = errors.New("fiscal year not found")
	ErrFiscalYearNotDraft                = errors.New("fiscal year can only be edited in draft status")
	ErrFiscalYearDeleteNotDraft          = errors.New("fiscal year can only be deleted in draft status")
	ErrFiscalYearInvalidRange            = errors.New("fiscal year end date must be on or after start date")
	ErrFiscalYearActiveAlreadyUsed       = errors.New("another active fiscal year already exists")
	ErrFiscalYearLockInvalidStatus       = errors.New("fiscal year can only be locked from active status")
	ErrFiscalYearActivateInvalid         = errors.New("fiscal year can only be activated from draft status")
	ErrFiscalYearOpeningBalanceNotPosted = errors.New("opening balance must be posted before fiscal year can be locked")
	ErrFiscalYearHasTransactions         = errors.New("fiscal year cannot be locked because posted journal exists in period")
	ErrFiscalYearCompanyMismatch         = errors.New("fiscal year does not belong to the provided company")
)

type FiscalYearUsecase interface {
	Create(ctx context.Context, req *dto.CreateFiscalYearRequest, createdBy *string) (*dto.FiscalYearResponse, error)
	List(ctx context.Context, req *dto.ListFiscalYearsRequest) ([]dto.FiscalYearResponse, int64, error)
	GetByID(ctx context.Context, id, companyID string) (*dto.FiscalYearResponse, error)
	Update(ctx context.Context, id, companyID string, req *dto.UpdateFiscalYearRequest) (*dto.FiscalYearResponse, error)
	Delete(ctx context.Context, id, companyID string) error
	Activate(ctx context.Context, id, companyID string) (*dto.FiscalYearResponse, error)
	Lock(ctx context.Context, id, companyID string) (*dto.FiscalYearResponse, error)
}

type fiscalYearUsecase struct {
	repo               repositories.FiscalYearRepository
	openingBalanceRepo repositories.OpeningBalanceRepository
	db                 *gorm.DB
}

func NewFiscalYearUsecase(db *gorm.DB, repo repositories.FiscalYearRepository, openingBalanceRepo repositories.OpeningBalanceRepository) FiscalYearUsecase {
	return &fiscalYearUsecase{db: db, repo: repo, openingBalanceRepo: openingBalanceRepo}
}

func validateFiscalYearRange(startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return ErrFiscalYearInvalidRange
	}
	return nil
}

func (uc *fiscalYearUsecase) parseDates(startDateRaw, endDateRaw string) (time.Time, time.Time, error) {
	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(startDateRaw))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endDate, err := time.Parse("2006-01-02", strings.TrimSpace(endDateRaw))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if err := validateFiscalYearRange(startDate, endDate); err != nil {
		return time.Time{}, time.Time{}, err
	}
	return startDate, endDate, nil
}

func toFiscalYearResponse(item *financeModels.FiscalYear) *dto.FiscalYearResponse {
	return &dto.FiscalYearResponse{
		ID:        item.ID,
		CompanyID: item.CompanyID,
		Name:      item.Name,
		StartDate: item.StartDate,
		EndDate:   item.EndDate,
		Status:    item.Status,
		CreatedBy: item.CreatedBy,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func (uc *fiscalYearUsecase) Create(ctx context.Context, req *dto.CreateFiscalYearRequest, createdBy *string) (*dto.FiscalYearResponse, error) {
	startDate, endDate, err := uc.parseDates(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))

	item := &financeModels.FiscalYear{
		TenantID:  tenantID,
		CompanyID: req.CompanyID,
		Name:      strings.TrimSpace(req.Name),
		StartDate: startDate,
		EndDate:   endDate,
		Status:    financeModels.FiscalYearStatusDraft,
		CreatedBy: createdBy,
	}
	if err := uc.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return toFiscalYearResponse(item), nil
}

func (uc *fiscalYearUsecase) List(ctx context.Context, req *dto.ListFiscalYearsRequest) ([]dto.FiscalYearResponse, int64, error) {
	items, total, err := uc.repo.List(ctx, repositories.FiscalYearListParams{
		CompanyID: req.CompanyID,
		Status:    req.Status,
		Page:      req.Page,
		PerPage:   req.PerPage,
	})
	if err != nil {
		return nil, 0, err
	}
	res := make([]dto.FiscalYearResponse, 0, len(items))
	for i := range items {
		mapped := toFiscalYearResponse(&items[i])
		res = append(res, *mapped)
	}
	return res, total, nil
}

func (uc *fiscalYearUsecase) GetByID(ctx context.Context, id, companyID string) (*dto.FiscalYearResponse, error) {
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFiscalYearNotFound
		}
		return nil, err
	}
	if item.CompanyID != companyID {
		return nil, ErrFiscalYearCompanyMismatch
	}
	return toFiscalYearResponse(item), nil
}

func (uc *fiscalYearUsecase) Update(ctx context.Context, id, companyID string, req *dto.UpdateFiscalYearRequest) (*dto.FiscalYearResponse, error) {
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFiscalYearNotFound
		}
		return nil, err
	}
	if item.CompanyID != companyID {
		return nil, ErrFiscalYearCompanyMismatch
	}
	if item.Status != financeModels.FiscalYearStatusDraft {
		return nil, ErrFiscalYearNotDraft
	}
	startDate, endDate, err := uc.parseDates(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	item.Name = strings.TrimSpace(req.Name)
	item.StartDate = startDate
	item.EndDate = endDate
	if err := uc.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return toFiscalYearResponse(item), nil
}

func (uc *fiscalYearUsecase) Delete(ctx context.Context, id, companyID string) error {
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFiscalYearNotFound
		}
		return err
	}

	if item.CompanyID != companyID {
		return ErrFiscalYearCompanyMismatch
	}

	if item.Status != financeModels.FiscalYearStatusDraft {
		return ErrFiscalYearDeleteNotDraft
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFiscalYearNotFound
		}
		return err
	}

	return nil
}

func (uc *fiscalYearUsecase) Activate(ctx context.Context, id, companyID string) (*dto.FiscalYearResponse, error) {
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFiscalYearNotFound
		}
		return nil, err
	}
	if item.CompanyID != companyID {
		return nil, ErrFiscalYearCompanyMismatch
	}
	if item.Status != financeModels.FiscalYearStatusDraft {
		return nil, ErrFiscalYearActivateInvalid
	}

	active, err := uc.repo.FindActiveByCompany(ctx, companyID)
	if err == nil && active != nil && active.ID != item.ID {
		return nil, ErrFiscalYearActiveAlreadyUsed
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	item.Status = financeModels.FiscalYearStatusActive
	if err := uc.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return toFiscalYearResponse(item), nil
}

func (uc *fiscalYearUsecase) Lock(ctx context.Context, id, companyID string) (*dto.FiscalYearResponse, error) {
	item, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFiscalYearNotFound
		}
		return nil, err
	}
	if item.CompanyID != companyID {
		return nil, ErrFiscalYearCompanyMismatch
	}
	if item.Status != financeModels.FiscalYearStatusActive {
		return nil, ErrFiscalYearLockInvalidStatus
	}

	if uc.openingBalanceRepo != nil {
		hasOpeningBalance, err := uc.openingBalanceRepo.HasPostedOpeningJournal(ctx, companyID, item.ID)
		if err != nil {
			return nil, err
		}
		if !hasOpeningBalance {
			return nil, ErrFiscalYearOpeningBalanceNotPosted
		}
	}

	hasTx, err := uc.repo.HasPostedJournalInRange(ctx, companyID, item.StartDate, item.EndDate)
	if err != nil {
		return nil, err
	}
	if hasTx {
		return nil, ErrFiscalYearHasTransactions
	}

	item.Status = financeModels.FiscalYearStatusLocked
	if err := uc.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return toFiscalYearResponse(item), nil
}
