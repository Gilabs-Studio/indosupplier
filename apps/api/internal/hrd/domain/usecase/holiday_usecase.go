package usecase

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/mapper"
	"gorm.io/gorm"
)

var (
	ErrHolidayNotFound      = errors.New("holiday not found")
	ErrHolidayAlreadyExists = errors.New("holiday on this date already exists")
	ErrInvalidCSVFormat     = errors.New("invalid CSV format")
)

// HolidayUsecase defines the interface for holiday business logic
type HolidayUsecase interface {
	List(ctx context.Context, req *dto.ListHolidaysRequest) ([]dto.HolidayResponse, *utils.PaginationResult, error)
	GetByID(ctx context.Context, id string) (*dto.HolidayResponse, error)
	GetByYear(ctx context.Context, year int) ([]dto.HolidayResponse, error)
	GetCalendar(ctx context.Context, year int) (*dto.HolidayCalendarResponse, error)
	Create(ctx context.Context, req *dto.CreateHolidayRequest) (*dto.HolidayResponse, error)
	CreateBatch(ctx context.Context, holidays []dto.CreateHolidayRequest) ([]dto.HolidayResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateHolidayRequest) (*dto.HolidayResponse, error)
	Delete(ctx context.Context, id string) error
	ImportFromCSV(ctx context.Context, rows []dto.HolidayCSVRow, year int, overwrite bool) (int, error)
	IsHoliday(ctx context.Context, dateStr string) (bool, *dto.HolidayResponse, error)
}

type holidayUsecase struct {
	repo   repositories.HolidayRepository
	mapper *mapper.HolidayMapper
}

// NewHolidayUsecase creates a new HolidayUsecase
func NewHolidayUsecase(repo repositories.HolidayRepository) HolidayUsecase {
	return &holidayUsecase{
		repo:   repo,
		mapper: mapper.NewHolidayMapper(),
	}
}

func (u *holidayUsecase) List(ctx context.Context, req *dto.ListHolidaysRequest) ([]dto.HolidayResponse, *utils.PaginationResult, error) {
	holidays, total, err := u.repo.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	responses := u.mapper.ToResponseList(holidays)

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

func (u *holidayUsecase) GetByID(ctx context.Context, id string) (*dto.HolidayResponse, error) {
	h, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHolidayNotFound
		}
		return nil, err
	}
	return u.mapper.ToResponse(h), nil
}

func (u *holidayUsecase) GetByYear(ctx context.Context, year int) ([]dto.HolidayResponse, error) {
	holidays, err := u.repo.FindByYear(ctx, year)
	if err != nil {
		return nil, err
	}
	return u.mapper.ToResponseList(holidays), nil
}

func (u *holidayUsecase) GetCalendar(ctx context.Context, year int) (*dto.HolidayCalendarResponse, error) {
	holidays, err := u.repo.FindByYear(ctx, year)
	if err != nil {
		return nil, err
	}
	return u.mapper.ToCalendarResponse(holidays, year), nil
}

func (u *holidayUsecase) Create(ctx context.Context, req *dto.CreateHolidayRequest) (*dto.HolidayResponse, error) {
	h, err := u.mapper.ToModel(req)
	if err != nil {
		return nil, err
	}

	if err := u.repo.Create(ctx, h); err != nil {
		return nil, err
	}

	return u.mapper.ToResponse(h), nil
}

func (u *holidayUsecase) CreateBatch(ctx context.Context, holidays []dto.CreateHolidayRequest) ([]dto.HolidayResponse, error) {
	var holidayModels []models.Holiday
	for _, req := range holidays {
		h, err := u.mapper.ToModel(&req)
		if err != nil {
			return nil, err
		}
		holidayModels = append(holidayModels, *h)
	}

	if err := u.repo.CreateBatch(ctx, holidayModels); err != nil {
		return nil, err
	}

	responses := make([]dto.HolidayResponse, len(holidayModels))
	for i := range holidayModels {
		responses[i] = *u.mapper.ToResponse(&holidayModels[i])
	}

	return responses, nil
}

func (u *holidayUsecase) Update(ctx context.Context, id string, req *dto.UpdateHolidayRequest) (*dto.HolidayResponse, error) {
	h, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHolidayNotFound
		}
		return nil, err
	}

	if err := u.mapper.ApplyUpdate(h, req); err != nil {
		return nil, err
	}

	if err := u.repo.Update(ctx, h); err != nil {
		return nil, err
	}

	return u.mapper.ToResponse(h), nil
}

func (u *holidayUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHolidayNotFound
		}
		return err
	}

	return u.repo.Delete(ctx, id)
}

func (u *holidayUsecase) ImportFromCSV(ctx context.Context, rows []dto.HolidayCSVRow, year int, overwrite bool) (int, error) {
	var holidayModels []models.Holiday

	for _, row := range rows {
		h, err := u.mapper.CSVRowToModel(&row, year)
		if err != nil {
			// Skip invalid rows, log error
			continue
		}
		holidayModels = append(holidayModels, *h)
	}

	if len(holidayModels) == 0 {
		return 0, ErrInvalidCSVFormat
	}

	// If overwrite, we would delete existing holidays for the year first
	// This would be done in a transaction in production

	if err := u.repo.CreateBatch(ctx, holidayModels); err != nil {
		return 0, err
	}

	return len(holidayModels), nil
}

func (u *holidayUsecase) IsHoliday(ctx context.Context, dateStr string) (bool, *dto.HolidayResponse, error) {
	// Parse the date string
	// This is simplified - in production, use proper date parsing
	req := &dto.ListHolidaysRequest{
		DateFrom: dateStr,
		DateTo:   dateStr,
		Page:     1,
		PerPage:  1,
	}

	holidays, _, err := u.repo.List(ctx, req)
	if err != nil {
		return false, nil, err
	}

	if len(holidays) > 0 {
		return true, u.mapper.ToResponse(&holidays[0]), nil
	}

	return false, nil, nil
}
