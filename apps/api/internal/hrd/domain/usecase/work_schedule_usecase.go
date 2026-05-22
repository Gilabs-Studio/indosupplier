package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/mapper"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	orgDTO "github.com/gilabs/gims/api/internal/organization/domain/dto"
	"gorm.io/gorm"
)

var (
	ErrWorkScheduleNotFound               = errors.New("work schedule not found")
	ErrWorkScheduleAlreadyExists          = errors.New("work schedule with this name already exists")
	ErrCannotDeleteDefaultSchedule        = errors.New("cannot delete default work schedule")
	ErrCannotSetDivisionScheduleAsDefault = errors.New("cannot set a division-specific schedule as default")
)

// WorkScheduleUsecase defines the interface for work schedule business logic
type WorkScheduleUsecase interface {
	List(ctx context.Context, req *dto.ListWorkSchedulesRequest) ([]dto.WorkScheduleResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.WorkScheduleResponse, error)
	GetByDivisionID(ctx context.Context, divisionID string) (*dto.WorkScheduleResponse, error)
	GetDefault(ctx context.Context) (*dto.WorkScheduleResponse, error)
	Create(ctx context.Context, req *dto.CreateWorkScheduleRequest) (*dto.WorkScheduleResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateWorkScheduleRequest) (*dto.WorkScheduleResponse, error)
	Delete(ctx context.Context, id string) error
	SetDefault(ctx context.Context, id string) error
	GetFormData(ctx context.Context) (*dto.WorkScheduleFormDataResponse, error)
}

type workScheduleUsecase struct {
	repo         repositories.WorkScheduleRepository
	divisionRepo orgRepos.DivisionRepository
	companyRepo  orgRepos.CompanyRepository
	mapper       *mapper.WorkScheduleMapper
}

// NewWorkScheduleUsecase creates a new WorkScheduleUsecase
func NewWorkScheduleUsecase(
	repo repositories.WorkScheduleRepository,
	divisionRepo orgRepos.DivisionRepository,
	companyRepo orgRepos.CompanyRepository,
) WorkScheduleUsecase {
	return &workScheduleUsecase{
		repo:         repo,
		divisionRepo: divisionRepo,
		companyRepo:  companyRepo,
		mapper:       mapper.NewWorkScheduleMapper(),
	}
}

func (u *workScheduleUsecase) List(ctx context.Context, req *dto.ListWorkSchedulesRequest) ([]dto.WorkScheduleResponse, *utils.PaginationResult, error) {
	schedules, total, err := u.repo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := u.mapper.ToResponseList(schedules)

	// Enrich division names
	u.enrichDivisionNames(ctx, responses)

	// Calculate pagination
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

func (u *workScheduleUsecase) GetByID(ctx context.Context, id string) (*dto.WorkScheduleResponse, error) {
	ws, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkScheduleNotFound
		}
		return nil, err
	}
	resp := u.mapper.ToResponse(ws)
	u.enrichDivisionName(ctx, resp)
	return resp, nil
}

func (u *workScheduleUsecase) GetByDivisionID(ctx context.Context, divisionID string) (*dto.WorkScheduleResponse, error) {
	ws, err := u.repo.FindByDivisionID(ctx, divisionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fall back to default schedule
			return u.GetDefault(ctx)
		}
		return nil, err
	}
	resp := u.mapper.ToResponse(ws)
	u.enrichDivisionName(ctx, resp)
	return resp, nil
}

func (u *workScheduleUsecase) GetDefault(ctx context.Context) (*dto.WorkScheduleResponse, error) {
	ws, err := u.repo.FindDefault(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkScheduleNotFound
		}
		return nil, err
	}
	resp := u.mapper.ToResponse(ws)
	u.enrichDivisionName(ctx, resp)
	return resp, nil
}

func (u *workScheduleUsecase) Create(ctx context.Context, req *dto.CreateWorkScheduleRequest) (*dto.WorkScheduleResponse, error) {
	ws := u.mapper.ToModel(req)

	// If this is marked as default, unset other defaults
	if ws.IsDefault {
		// This will be handled in a transaction if needed
	}

	if err := u.repo.Create(ctx, ws); err != nil {
		return nil, err
	}

	// If marked as default, set it as default (handles unsetting others)
	if ws.IsDefault {
		if err := u.repo.SetDefault(ctx, ws.ID); err != nil {
			return nil, err
		}
	}

	return u.mapper.ToResponse(ws), nil
}

func (u *workScheduleUsecase) Update(ctx context.Context, id string, req *dto.UpdateWorkScheduleRequest) (*dto.WorkScheduleResponse, error) {
	ws, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkScheduleNotFound
		}
		return nil, err
	}

	u.mapper.ApplyUpdate(ws, req)

	if err := u.repo.Update(ctx, ws); err != nil {
		return nil, err
	}

	// If marked as default, set it as default (handles unsetting others)
	if req.IsDefault != nil && *req.IsDefault {
		if err := u.repo.SetDefault(ctx, ws.ID); err != nil {
			return nil, err
		}
	}

	return u.mapper.ToResponse(ws), nil
}

func (u *workScheduleUsecase) Delete(ctx context.Context, id string) error {
	ws, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkScheduleNotFound
		}
		return err
	}

	// Prevent deleting default schedule
	if ws.IsDefault {
		return ErrCannotDeleteDefaultSchedule
	}

	return u.repo.Delete(ctx, id)
}

func (u *workScheduleUsecase) SetDefault(ctx context.Context, id string) error {
	// Verify schedule exists
	schedule, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkScheduleNotFound
		}
		return err
	}

	// Only general (non-division) schedules can be set as default
	if schedule.DivisionID != nil && *schedule.DivisionID != "" {
		return ErrCannotSetDivisionScheduleAsDefault
	}

	return u.repo.SetDefault(ctx, id)
}

// GetWorkScheduleForEmployee gets the appropriate work schedule for an employee
// This can be extended to check employee's division or other factors
func (u *workScheduleUsecase) GetWorkScheduleForEmployee(ctx context.Context, divisionID *string) (*models.WorkSchedule, error) {
	var ws *models.WorkSchedule
	var err error

	// Try division-specific schedule first
	if divisionID != nil && *divisionID != "" {
		ws, err = u.repo.FindByDivisionID(ctx, *divisionID)
		if err == nil {
			return ws, nil
		}
		// If not found, fall back to default
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// Get default schedule
	ws, err = u.repo.FindDefault(ctx)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

// GetFormData returns dropdown data for the work schedule form (divisions and companies)
func (u *workScheduleUsecase) GetFormData(ctx context.Context) (*dto.WorkScheduleFormDataResponse, error) {
	// Get active divisions
	divListReq := &orgDTO.ListDivisionsRequest{Page: 1, PerPage: 100}
	divisions, _, err := u.divisionRepo.List(ctx, divListReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch divisions: %w", err)
	}
	divisionOptions := make([]dto.DivisionFormOption, 0, len(divisions))
	for _, div := range divisions {
		if div.IsActive {
			divisionOptions = append(divisionOptions, dto.DivisionFormOption{
				ID:   div.ID,
				Name: div.Name,
			})
		}
	}

	// Get active companies with coordinates
	isActive := true
	compListReq := &orgDTO.ListCompaniesRequest{Page: 1, PerPage: 100, IsActive: &isActive}
	companies, _, err := u.companyRepo.List(ctx, compListReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch companies: %w", err)
	}
	companyOptions := make([]dto.CompanyFormOption, 0, len(companies))
	for _, c := range companies {
		companyOptions = append(companyOptions, dto.CompanyFormOption{
			ID:        c.ID,
			Name:      c.Name,
			Latitude:  c.Latitude,
			Longitude: c.Longitude,
		})
	}

	return &dto.WorkScheduleFormDataResponse{
		Divisions: divisionOptions,
		Companies: companyOptions,
	}, nil
}

// enrichDivisionName populates the DivisionName field for a single response
func (u *workScheduleUsecase) enrichDivisionName(ctx context.Context, resp *dto.WorkScheduleResponse) {
	if resp.DivisionID != nil && *resp.DivisionID != "" {
		division, err := u.divisionRepo.FindByID(ctx, *resp.DivisionID)
		if err == nil {
			resp.DivisionName = division.Name
		}
	}
}

// enrichDivisionNames populates DivisionName for a list of responses
func (u *workScheduleUsecase) enrichDivisionNames(ctx context.Context, responses []dto.WorkScheduleResponse) {
	// Collect unique division IDs
	divisionIDs := make(map[string]bool)
	for _, resp := range responses {
		if resp.DivisionID != nil && *resp.DivisionID != "" {
			divisionIDs[*resp.DivisionID] = true
		}
	}

	// Fetch division names
	divisionNames := make(map[string]string)
	for id := range divisionIDs {
		division, err := u.divisionRepo.FindByID(ctx, id)
		if err == nil {
			divisionNames[id] = division.Name
		}
	}

	// Enrich responses
	for i := range responses {
		if responses[i].DivisionID != nil {
			if name, ok := divisionNames[*responses[i].DivisionID]; ok {
				responses[i].DivisionName = name
			}
		}
	}
}
