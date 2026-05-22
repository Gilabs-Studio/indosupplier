package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	crmModels "github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/travel_planner/data/models"
	"github.com/gilabs/gims/api/internal/travel_planner/data/repositories"
	"github.com/gilabs/gims/api/internal/travel_planner/domain/dto"
	"github.com/gilabs/gims/api/internal/travel_planner/domain/mapper"
	"github.com/go-pdf/fpdf"
	"gorm.io/gorm"
)

//go:embed travel_plan_report.html
var travelPlanReportHTMLTemplate string

var (
	ErrTravelPlanNotFound    = errors.New("travel plan not found")
	ErrInvalidDateRange      = errors.New("invalid date range")
	ErrInvalidTravelMode     = errors.New("invalid travel mode")
	ErrInvalidStatus         = errors.New("invalid travel status")
	ErrInvalidDayDate        = errors.New("invalid day date")
	ErrInvalidStopCategory   = errors.New("invalid stop category")
	ErrInvalidStopSource     = errors.New("invalid stop source")
	ErrInvalidExpenseType    = errors.New("invalid expense type")
	ErrInvalidBudgetAmount   = errors.New("invalid budget amount")
	ErrTravelExpenseNotFound = errors.New("travel expense not found")
	ErrVisitNotFound         = errors.New("visit not found")
	ErrInvalidSearchQuery    = errors.New("search query must contain at least 2 characters")
	ErrInvalidCheckpoint     = errors.New("invalid checkpoint")
)

type TravelPlanUsecase interface {
	Create(ctx context.Context, req *dto.CreateTravelPlanRequest) (*dto.TravelPlanResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateTravelPlanRequest) (*dto.TravelPlanResponse, error)
	UpdateParticipants(ctx context.Context, id string, participantIDs []string) (*dto.TravelPlanResponse, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*dto.TravelPlanResponse, error)
	List(ctx context.Context, req *dto.ListTravelPlansRequest) ([]dto.TravelPlanResponse, int64, int, int, error)
	GetFormData(ctx context.Context) (*dto.TravelPlannerFormDataResponse, error)
	GetVisitPlannerFormData(ctx context.Context, req *dto.VisitPlannerFormDataRequest) (*dto.VisitPlannerFormDataResponse, error)
	CreateVisitPlannerPlan(ctx context.Context, req *dto.CreateVisitPlannerPlanRequest) (*dto.CreateVisitPlannerPlanResponse, error)
	ListParticipants(ctx context.Context, req *dto.ListTravelPlanParticipantsRequest) ([]dto.EmployeeFormOption, int64, int, int, error)
	SearchPlaces(ctx context.Context, query string, provider string) ([]dto.PlaceSearchResult, error)
	OptimizeRoute(ctx context.Context, planID string) (*dto.RouteOptimizationResponse, error)
	OptimizeRouteForVisit(ctx context.Context, req *dto.OptimizeNavigationRequest) (*dto.OptimizeNavigationResponse, error)
	GetGoogleMapsLinks(ctx context.Context, planID string) ([]dto.DayGoogleMapsLink, error)
	ExportPDF(ctx context.Context, planID string, dayIndex *int) ([]byte, string, error)
	ExportReportHTML(ctx context.Context, planID string) ([]byte, string, error)
	ListExpenses(ctx context.Context, planID string) (*dto.TravelExpenseListResponse, error)
	CreateExpense(ctx context.Context, planID string, req *dto.CreateTravelExpenseRequest) (*dto.TravelExpenseResponse, error)
	DeleteExpense(ctx context.Context, planID string, expenseID string) error
	ListVisitPlannerRoutes(ctx context.Context, req *dto.ListVisitPlannerRoutesRequest) ([]dto.ActiveVisitRouteResponse, int64, int, int, error)
	ListVisits(ctx context.Context, planID string) ([]dto.TravelPlanVisitResponse, error)
	ListAvailableVisits(ctx context.Context, search string) ([]dto.TravelPlanVisitResponse, error)
	LinkVisits(ctx context.Context, planID string, req *dto.LinkTravelPlanVisitsRequest) (int64, error)
	UnlinkVisit(ctx context.Context, planID string, visitID string) error
	CreateVisitFromTrip(ctx context.Context, planID string, req *dto.CreateTravelPlanVisitRequest) (*dto.TravelPlanVisitResponse, error)
	UpsertVisitLog(ctx context.Context, req *dto.UpsertVisitLogRequest) (*dto.VisitLogResponse, error)
	UpsertLocation(ctx context.Context, req *dto.LocationUpdateRequest) (*dto.LocationUpdateResponse, error)
	StartNavigation(ctx context.Context, req *dto.StartNavigationRequest) (*dto.NavigationStatusResponse, error)
	StopNavigation(ctx context.Context, req *dto.StopNavigationRequest) (*dto.NavigationStatusResponse, error)
	ResolveVisibleEmployeeIDs(ctx context.Context, requested []string) ([]string, error)
}

type travelPlanUsecase struct {
	db           *gorm.DB
	repo         repositories.TravelPlanRepository
	mapper       *mapper.TravelPlanMapper
	httpClient   *http.Client
	googleAPIKey string
}

func NewTravelPlanUsecase(db *gorm.DB, repo repositories.TravelPlanRepository, planMapper *mapper.TravelPlanMapper) TravelPlanUsecase {
	return &travelPlanUsecase{
		db:           db,
		repo:         repo,
		mapper:       planMapper,
		httpClient:   &http.Client{Timeout: 12 * time.Second},
		googleAPIKey: strings.TrimSpace(os.Getenv("GOOGLE_MAPS_API_KEY")),
	}
}

func (uc *travelPlanUsecase) Create(ctx context.Context, req *dto.CreateTravelPlanRequest) (*dto.TravelPlanResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	mode, err := parseTravelMode(req.Mode)
	if err != nil {
		return nil, err
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return nil, err
	}
	if endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}

	budgetAmount, err := parseBudgetAmount(req.BudgetAmount)
	if err != nil {
		return nil, err
	}

	days, err := uc.buildDays(req.Days)
	if err != nil {
		return nil, err
	}

	code, err := uc.repo.GenerateCode(ctx, apptime.Now())
	if err != nil {
		return nil, err
	}

	createdBy := strings.TrimSpace(getActorID(ctx))
	var createdByPtr *string
	if createdBy != "" {
		createdByPtr = &createdBy
	}

	plan := &models.TravelPlan{
		Code:         code,
		Title:        strings.TrimSpace(req.Title),
		PlanType:     models.TravelPlanTypeUpCountryCost,
		Mode:         mode,
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       models.TravelPlanStatusDraft,
		BudgetAmount: budgetAmount,
		Notes:        strings.TrimSpace(req.Notes),
		Days:         days,
		CreatedBy:    createdByPtr,
	}

	if err := uc.repo.Create(ctx, plan); err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, plan.ID, true)
	if err != nil {
		return nil, err
	}

	response := uc.mapper.ToResponse(full)
	return &response, nil
}

func (uc *travelPlanUsecase) Update(ctx context.Context, id string, req *dto.UpdateTravelPlanRequest) (*dto.TravelPlanResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	id = strings.TrimSpace(id)
	existing, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTravelPlanNotFound
		}
		return nil, err
	}

	mode, err := parseTravelMode(req.Mode)
	if err != nil {
		return nil, err
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return nil, err
	}
	if endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}

	budgetAmount, err := parseBudgetAmount(req.BudgetAmount)
	if err != nil {
		return nil, err
	}

	status := existing.Status
	if strings.TrimSpace(req.Status) != "" {
		status, err = parseTravelStatus(req.Status)
		if err != nil {
			return nil, err
		}
	}

	days, err := uc.buildDays(req.Days)
	if err != nil {
		return nil, err
	}

	existing.Title = strings.TrimSpace(req.Title)
	existing.Mode = mode
	existing.StartDate = startDate
	existing.EndDate = endDate
	existing.Status = status
	existing.BudgetAmount = budgetAmount
	existing.Notes = strings.TrimSpace(req.Notes)

	if err := uc.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	if err := uc.repo.ReplaceDays(ctx, existing.ID, days); err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, existing.ID, true)
	if err != nil {
		return nil, err
	}

	response := uc.mapper.ToResponse(full)
	return &response, nil
}

func (uc *travelPlanUsecase) UpdateParticipants(ctx context.Context, id string, participantIDs []string) (*dto.TravelPlanResponse, error) {
	id = strings.TrimSpace(id)
	existing, err := uc.repo.FindByID(ctx, id, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTravelPlanNotFound
		}
		return nil, err
	}

	existing.Notes = mergeParticipantMeta(strings.TrimSpace(existing.Notes), participantIDs)

	if err := uc.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	full, err := uc.repo.FindByID(ctx, existing.ID, true)
	if err != nil {
		return nil, err
	}

	response := uc.mapper.ToResponse(full)
	return &response, nil
}

func (uc *travelPlanUsecase) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrTravelPlanNotFound
	}

	if _, err := uc.repo.FindByID(ctx, id, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTravelPlanNotFound
		}
		return err
	}

	return uc.repo.Delete(ctx, id)
}

func (uc *travelPlanUsecase) GetByID(ctx context.Context, id string) (*dto.TravelPlanResponse, error) {
	id = strings.TrimSpace(id)
	plan, err := uc.repo.FindByID(ctx, id, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTravelPlanNotFound
		}
		return nil, err
	}

	response := uc.mapper.ToResponse(plan)
	return &response, nil
}

func (uc *travelPlanUsecase) List(ctx context.Context, req *dto.ListTravelPlansRequest) ([]dto.TravelPlanResponse, int64, int, int, error) {
	if req == nil {
		req = &dto.ListTravelPlansRequest{}
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

	var planType *models.TravelPlanType
	if req.PlanType != nil && strings.TrimSpace(*req.PlanType) != "" {
		parsedPlanType, err := parseTravelPlanType(*req.PlanType)
		if err != nil {
			return nil, 0, page, perPage, err
		}
		planType = &parsedPlanType
	}

	var mode *models.TravelMode
	if req.Mode != nil && strings.TrimSpace(*req.Mode) != "" {
		parsedMode, err := parseTravelMode(*req.Mode)
		if err != nil {
			return nil, 0, page, perPage, err
		}
		mode = &parsedMode
	}

	var status *models.TravelPlanStatus
	if req.Status != nil && strings.TrimSpace(*req.Status) != "" {
		parsedStatus, err := parseTravelStatus(*req.Status)
		if err != nil {
			return nil, 0, page, perPage, err
		}
		status = &parsedStatus
	}

	var startDate *time.Time
	if req.StartDate != nil && strings.TrimSpace(*req.StartDate) != "" {
		parsedStartDate, err := parseDate(*req.StartDate)
		if err != nil {
			return nil, 0, page, perPage, err
		}
		startDate = &parsedStartDate
	}

	var endDate *time.Time
	if req.EndDate != nil && strings.TrimSpace(*req.EndDate) != "" {
		parsedEndDate, err := parseDate(*req.EndDate)
		if err != nil {
			return nil, 0, page, perPage, err
		}
		endDate = &parsedEndDate
	}

	plans, total, err := uc.repo.List(ctx, repositories.TravelPlanListParams{
		Search:    strings.TrimSpace(req.Search),
		PlanType:  planType,
		Mode:      mode,
		Status:    status,
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	})
	if err == nil && planType != nil && *planType == models.TravelPlanTypeVisitReport && total == 0 {
		uc.backfillVisitReportPlans(ctx)
		plans, total, err = uc.repo.List(ctx, repositories.TravelPlanListParams{
			Search:    strings.TrimSpace(req.Search),
			PlanType:  planType,
			Mode:      mode,
			Status:    status,
			StartDate: startDate,
			EndDate:   endDate,
			Limit:     perPage,
			Offset:    (page - 1) * perPage,
		})
	}
	if err != nil {
		return nil, 0, page, perPage, err
	}

	responses := uc.mapper.ToResponseList(plans)
	return responses, total, page, perPage, nil
}

// backfillVisitReportPlans creates travel planner plans for submitted/approved/rejected visits that are not linked yet.
func (uc *travelPlanUsecase) backfillVisitReportPlans(ctx context.Context) {
	visits := make([]crmModels.VisitReport, 0)
	if err := uc.db.WithContext(ctx).
		Where("travel_plan_id IS NULL").
		Where("status IN ?", []string{"submitted", "approved", "rejected"}).
		Order("created_at DESC").
		Limit(500).
		Find(&visits).Error; err != nil {
		return
	}

	for _, visit := range visits {
		now := apptime.Now()
		prefix := fmt.Sprintf("TPL-%s", now.Format("200601"))
		var count int64
		if err := uc.db.WithContext(ctx).
			Model(&models.TravelPlan{}).
			Where("code LIKE ?", prefix+"-%").
			Count(&count).Error; err != nil {
			continue
		}

		code := fmt.Sprintf("%s-%04d", prefix, count+1)
		visitDate := time.Date(visit.VisitDate.Year(), visit.VisitDate.Month(), visit.VisitDate.Day(), 0, 0, 0, 0, apptime.Location())
		plan := models.TravelPlan{
			Code:         code,
			Title:        fmt.Sprintf("Visit %s", visit.Code),
			PlanType:     models.TravelPlanTypeVisitReport,
			Mode:         models.TravelModeMilestone,
			StartDate:    visitDate,
			EndDate:      visitDate,
			Status:       models.TravelPlanStatusActive,
			BudgetAmount: 0,
			Notes:        visit.Purpose,
			CreatedBy:    visit.CreatedBy,
		}

		if err := uc.db.WithContext(ctx).Create(&plan).Error; err != nil {
			continue
		}

		_ = uc.db.WithContext(ctx).
			Model(&crmModels.VisitReport{}).
			Where("id = ?", visit.ID).
			Update("travel_plan_id", plan.ID).Error
	}
}

func (uc *travelPlanUsecase) GetFormData(ctx context.Context) (*dto.TravelPlannerFormDataResponse, error) {
	type employeeFormRow struct {
		ID           string `gorm:"column:id"`
		EmployeeCode string `gorm:"column:employee_code"`
		Name         string `gorm:"column:name"`
		AvatarURL    string `gorm:"column:avatar_url"`
	}

	rows := make([]employeeFormRow, 0)
	err := uc.db.WithContext(ctx).
		Table("employees AS e").
		Select("e.id, e.employee_code, e.name, COALESCE(u.avatar_url, '') AS avatar_url").
		Joins("LEFT JOIN users AS u ON u.id = e.user_id").
		Where("e.deleted_at IS NULL").
		Where("e.is_active = ?", true).
		Order("e.name ASC").
		Limit(200).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	employees := make([]dto.EmployeeFormOption, 0, len(rows))
	for _, row := range rows {
		employees = append(employees, dto.EmployeeFormOption{
			ID:           row.ID,
			EmployeeCode: row.EmployeeCode,
			Name:         row.Name,
			AvatarURL:    row.AvatarURL,
		})
	}

	return &dto.TravelPlannerFormDataResponse{
		Modes: []dto.EnumOption{
			{Value: string(models.TravelModeLogistic), Label: "Logistic"},
			{Value: string(models.TravelModeCargo), Label: "Cargo"},
			{Value: string(models.TravelModeVessel), Label: "Vessel"},
			{Value: string(models.TravelModeMilestone), Label: "Milestone"},
		},
		Categories: []dto.EnumOption{
			{Value: string(models.TravelStopCategoryPickup), Label: "Pickup"},
			{Value: string(models.TravelStopCategoryDropoff), Label: "Dropoff"},
			{Value: string(models.TravelStopCategoryRefuel), Label: "Refuel"},
			{Value: string(models.TravelStopCategoryCheckpoint), Label: "Checkpoint"},
			{Value: string(models.TravelStopCategoryRest), Label: "Rest"},
			{Value: string(models.TravelStopCategoryCustom), Label: "Custom"},
		},
		Sources: []dto.EnumOption{
			{Value: string(models.TravelStopSourceManual), Label: "Manual"},
			{Value: string(models.TravelStopSourceGooglePlaces), Label: "Google Places"},
			{Value: string(models.TravelStopSourceOpenStreetMap), Label: "OpenStreetMap"},
		},
		Employees: employees,
		ExpenseTypes: []dto.EnumOption{
			{Value: string(models.TravelExpenseTypeTransport), Label: "Transport"},
			{Value: string(models.TravelExpenseTypeAccommodation), Label: "Accommodation"},
			{Value: string(models.TravelExpenseTypeMeal), Label: "Meal"},
			{Value: string(models.TravelExpenseTypeFuel), Label: "Fuel"},
			{Value: string(models.TravelExpenseTypeToll), Label: "Toll"},
			{Value: string(models.TravelExpenseTypeParking), Label: "Parking"},
			{Value: string(models.TravelExpenseTypeOther), Label: "Other"},
		},
	}, nil
}

func (uc *travelPlanUsecase) ListParticipants(
	ctx context.Context,
	req *dto.ListTravelPlanParticipantsRequest,
) ([]dto.EmployeeFormOption, int64, int, int, error) {
	if req == nil {
		req = &dto.ListTravelPlanParticipantsRequest{}
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

	search := strings.TrimSpace(req.Search)

	baseQuery := uc.db.WithContext(ctx).
		Table("employees AS e").
		Joins("LEFT JOIN users AS u ON u.id = e.user_id").
		Where("e.deleted_at IS NULL").
		Where("e.is_active = ?", true)

	if search != "" {
		like := search + "%"
		baseQuery = baseQuery.Where("e.name ILIKE ? OR e.employee_code ILIKE ?", like, like)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, page, perPage, err
	}

	type employeeFormRow struct {
		ID           string `gorm:"column:id"`
		EmployeeCode string `gorm:"column:employee_code"`
		Name         string `gorm:"column:name"`
		AvatarURL    string `gorm:"column:avatar_url"`
	}

	rows := make([]employeeFormRow, 0)
	if err := baseQuery.
		Select("e.id, e.employee_code, e.name, COALESCE(u.avatar_url, '') AS avatar_url").
		Order("e.name ASC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&rows).Error; err != nil {
		return nil, 0, page, perPage, err
	}

	items := make([]dto.EmployeeFormOption, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.EmployeeFormOption{
			ID:           row.ID,
			EmployeeCode: row.EmployeeCode,
			Name:         row.Name,
			AvatarURL:    row.AvatarURL,
		})
	}

	return items, total, page, perPage, nil
}

func (uc *travelPlanUsecase) SearchPlaces(ctx context.Context, query string, provider string) ([]dto.PlaceSearchResult, error) {
	searchQuery := strings.TrimSpace(query)
	if len(searchQuery) < 2 {
		return nil, ErrInvalidSearchQuery
	}

	provider = strings.ToLower(strings.TrimSpace(provider))

	if provider == "google" || provider == "google_places" {
		if uc.googleAPIKey != "" {
			googleResults, err := uc.searchGooglePlaces(ctx, searchQuery)
			if err == nil {
				return googleResults, nil
			}
		}
		return uc.searchOpenStreetMap(ctx, searchQuery)
	}

	if provider == "osm" || provider == "open_street_map" {
		return uc.searchOpenStreetMap(ctx, searchQuery)
	}

	if uc.googleAPIKey != "" {
		googleResults, err := uc.searchGooglePlaces(ctx, searchQuery)
		if err == nil {
			return googleResults, nil
		}
	}

	return uc.searchOpenStreetMap(ctx, searchQuery)
}

func (uc *travelPlanUsecase) OptimizeRoute(ctx context.Context, planID string) (*dto.RouteOptimizationResponse, error) {
	planID = strings.TrimSpace(planID)
	plan, err := uc.repo.FindByID(ctx, planID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTravelPlanNotFound
		}
		return nil, err
	}

	summaries := make([]dto.RouteOptimizationDaySummary, 0, len(plan.Days))
	optimizedAt := apptime.Now()

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for dayIndex := range plan.Days {
			day := &plan.Days[dayIndex]
			if len(day.Stops) == 0 {
				continue
			}

			optimizedStops, totalDistance := optimizeStops(day.Stops)
			for stopIdx := range optimizedStops {
				optimizedStops[stopIdx].OrderIndex = stopIdx + 1
				if err := tx.Model(&models.TravelPlanStop{}).
					Where("id = ?", optimizedStops[stopIdx].ID).
					Update("order_index", optimizedStops[stopIdx].OrderIndex).Error; err != nil {
					return err
				}
			}

			day.Stops = optimizedStops
			optimizedStopIDs := make([]string, 0, len(optimizedStops))
			for _, stop := range optimizedStops {
				optimizedStopIDs = append(optimizedStopIDs, stop.ID)
			}

			summaries = append(summaries, dto.RouteOptimizationDaySummary{
				DayID:            day.ID,
				DayIndex:         day.DayIndex,
				TotalDistanceKM:  roundToTwo(totalDistance),
				GoogleMapsURL:    buildGoogleMapsURL(optimizedStops),
				OptimizedStopIDs: optimizedStopIDs,
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(summaries, func(i, j int) bool {
		return summaries[i].DayIndex < summaries[j].DayIndex
	})

	return &dto.RouteOptimizationResponse{
		PlanID:      plan.ID,
		OptimizedAt: optimizedAt.Format(time.RFC3339),
		Days:        summaries,
	}, nil
}

func (uc *travelPlanUsecase) GetGoogleMapsLinks(ctx context.Context, planID string) ([]dto.DayGoogleMapsLink, error) {
	planID = strings.TrimSpace(planID)
	plan, err := uc.repo.FindByID(ctx, planID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTravelPlanNotFound
		}
		return nil, err
	}

	links := make([]dto.DayGoogleMapsLink, 0, len(plan.Days))
	for _, day := range plan.Days {
		sortedStops := append([]models.TravelPlanStop{}, day.Stops...)
		sort.SliceStable(sortedStops, func(i, j int) bool {
			return sortedStops[i].OrderIndex < sortedStops[j].OrderIndex
		})
		links = append(links, dto.DayGoogleMapsLink{
			DayID:    day.ID,
			DayIndex: day.DayIndex,
			URL:      buildGoogleMapsURL(sortedStops),
		})
	}

	sort.SliceStable(links, func(i, j int) bool {
		return links[i].DayIndex < links[j].DayIndex
	})

	return links, nil
}

func (uc *travelPlanUsecase) ExportPDF(ctx context.Context, planID string, dayIndex *int) ([]byte, string, error) {
	planID = strings.TrimSpace(planID)
	plan, err := uc.repo.FindByID(ctx, planID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrTravelPlanNotFound
		}
		return nil, "", err
	}

	days := append([]models.TravelPlanDay{}, plan.Days...)
	sort.SliceStable(days, func(i, j int) bool {
		return days[i].DayIndex < days[j].DayIndex
	})

	if dayIndex != nil {
		filtered := make([]models.TravelPlanDay, 0, 1)
		for _, day := range days {
			if day.DayIndex == *dayIndex {
				filtered = append(filtered, day)
				break
			}
		}
		if len(filtered) == 0 {
			return nil, "", errors.New("day not found for export")
		}
		days = filtered
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 12)

	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Travel Planner Brief")
	pdf.Ln(9)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, fmt.Sprintf("Plan Code: %s", plan.Code))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("Title: %s", plan.Title))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("Mode: %s", strings.Title(string(plan.Mode))))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("Period: %s to %s", plan.StartDate.Format("2006-01-02"), plan.EndDate.Format("2006-01-02")))
	pdf.Ln(8)
	if strings.TrimSpace(plan.Notes) != "" {
		pdf.MultiCell(0, 6, "Notes: "+strings.TrimSpace(plan.Notes), "", "L", false)
		pdf.Ln(2)
	}

	for _, day := range days {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 8, fmt.Sprintf("Day %d - %s", day.DayIndex, day.DayDate.Format("2006-01-02")))
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 11)
		if strings.TrimSpace(day.Summary) != "" {
			pdf.MultiCell(0, 6, "Summary: "+strings.TrimSpace(day.Summary), "", "L", false)
			pdf.Ln(2)
		}

		sortedStops := append([]models.TravelPlanStop{}, day.Stops...)
		sort.SliceStable(sortedStops, func(i, j int) bool {
			return sortedStops[i].OrderIndex < sortedStops[j].OrderIndex
		})

		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 7, "Itinerary Stops")
		pdf.Ln(7)
		pdf.SetFont("Arial", "", 10)
		for _, stop := range sortedStops {
			lockTag := ""
			if stop.IsLocked {
				lockTag = " [LOCKED]"
			}
			pdf.MultiCell(
				0,
				5,
				fmt.Sprintf("%d. %s%s (%s) [%.6f, %.6f]", stop.OrderIndex, stop.PlaceName, lockTag, strings.Title(string(stop.Category)), stop.Latitude, stop.Longitude),
				"",
				"L",
				false,
			)
		}

		sortedNotes := append([]models.TravelPlanDayNote{}, day.Notes...)
		sort.SliceStable(sortedNotes, func(i, j int) bool {
			return sortedNotes[i].OrderIndex < sortedNotes[j].OrderIndex
		})
		if len(sortedNotes) > 0 {
			pdf.Ln(2)
			pdf.SetFont("Arial", "B", 11)
			pdf.Cell(0, 7, "Day Notes")
			pdf.Ln(7)
			pdf.SetFont("Arial", "", 10)
			for _, note := range sortedNotes {
				noteTime := strings.TrimSpace(note.NoteTime)
				if noteTime == "" {
					noteTime = "--:--"
				}
				iconTag := strings.TrimSpace(note.IconTag)
				if iconTag == "" {
					iconTag = "note"
				}
				pdf.MultiCell(0, 5, fmt.Sprintf("- [%s] %s %s", iconTag, noteTime, note.NoteText), "", "L", false)
			}
		}
	}

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("travel-plan-%s.pdf", strings.ToLower(plan.Code))
	if dayIndex != nil {
		filename = fmt.Sprintf("travel-plan-%s-day-%d.pdf", strings.ToLower(plan.Code), *dayIndex)
	}

	return buffer.Bytes(), filename, nil
}

func (uc *travelPlanUsecase) ExportReportHTML(ctx context.Context, planID string) ([]byte, string, error) {
	planID = strings.TrimSpace(planID)
	plan, err := uc.repo.FindByID(ctx, planID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrTravelPlanNotFound
		}
		return nil, "", err
	}

	expenses, err := uc.repo.ListExpenses(ctx, planID)
	if err != nil {
		return nil, "", err
	}

	// ── Template data types ─────────────────────────────────────────────────

	type reportStop struct {
		PlaceName string
		Note      string
		Time      string
		Category  string
	}

	type reportDay struct {
		DayIndex int
		DayDate  string
		Summary  string
		Stops    []reportStop
	}

	type reportExpense struct {
		No          int
		Description string
		Type        string
		Date        string
		Amount      float64
	}

	type reportData struct {
		Code          string
		CompanyName   string
		CompanyAddress string
		CompanyPhone  string
		CompanyEmail  string
		PlanType      string
		Mode          string
		Status        string
		StartDate     string
		EndDate       string
		BudgetAmount  float64
		Notes         string
		PrintDate     string
		Days          []reportDay
		Expenses      []reportExpense
		TotalExpense  float64
		Remaining     float64
	}

	// ── Format helpers ──────────────────────────────────────────────────────

	formatDateStr := func(t time.Time) string {
		return t.Format("02 Jan 2006")
	}

	titleCase := func(s string) string {
		if s == "" {
			return s
		}
		parts := strings.Split(strings.ReplaceAll(s, "_", " "), " ")
		for i, p := range parts {
			if len(p) > 0 {
				parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
			}
		}
		return strings.Join(parts, " ")
	}

	// ── Build days with stops ───────────────────────────────────────────────

	sortedDays := append([]models.TravelPlanDay{}, plan.Days...)
	sort.SliceStable(sortedDays, func(i, j int) bool {
		return sortedDays[i].DayIndex < sortedDays[j].DayIndex
	})

	days := make([]reportDay, 0, len(sortedDays))
	for _, day := range sortedDays {
		sortedStops := append([]models.TravelPlanStop{}, day.Stops...)
		sort.SliceStable(sortedStops, func(i, j int) bool {
			return sortedStops[i].OrderIndex < sortedStops[j].OrderIndex
		})

		stops := make([]reportStop, 0, len(sortedStops))
		for _, stop := range sortedStops {
			// Extract ETA time from note if present (format: "ETA HH:MM ...")
			eta := ""
			noteText := strings.TrimSpace(stop.Note)
			if strings.HasPrefix(noteText, "ETA ") {
				parts := strings.SplitN(noteText[4:], " ", 2)
				eta = parts[0]
				if len(parts) > 1 {
					noteText = strings.TrimPrefix(strings.TrimSpace(parts[1]), "— ")
				} else {
					noteText = ""
				}
			}

			stops = append(stops, reportStop{
				PlaceName: stop.PlaceName,
				Note:      noteText,
				Time:      eta,
				Category:  titleCase(string(stop.Category)),
			})
		}

		dayDate := ""
		if !day.DayDate.IsZero() {
			dayDate = formatDateStr(day.DayDate)
		}

		days = append(days, reportDay{
			DayIndex: day.DayIndex,
			DayDate:  dayDate,
			Summary:  strings.TrimSpace(day.Summary),
			Stops:    stops,
		})
	}

	// ── Build expense rows ──────────────────────────────────────────────────

	expenseRows := make([]reportExpense, 0, len(expenses))
	totalExpense := 0.0
	for i, exp := range expenses {
		totalExpense += exp.Amount
		desc := strings.TrimSpace(exp.Description)
		if desc == "" {
			desc = titleCase(string(exp.ExpenseType))
		}
		expenseRows = append(expenseRows, reportExpense{
			No:          i + 1,
			Description: desc,
			Type:        titleCase(string(exp.ExpenseType)),
			Date:        exp.ExpenseDate.Format("02 Jan 2006"),
			Amount:      exp.Amount,
		})
	}

	if len(expenseRows) == 0 {
		expenseRows = append(expenseRows, reportExpense{
			No:          1,
			Description: "No expenses recorded",
			Type:        "-",
			Date:        "-",
			Amount:      0,
		})
	}

	remaining := plan.BudgetAmount - totalExpense
	if remaining < 0 {
		remaining = 0
	}

	// ── Template functions ──────────────────────────────────────────────────

	funcMap := template.FuncMap{
		"formatMoney": func(v float64) string {
			// Format as "Rp 1.234.567"
			abs := math.Abs(v)
			intPart := int64(abs)
			s := fmt.Sprintf("%d", intPart)
			n := len(s)
			var out []byte
			for i, c := range s {
				if i > 0 && (n-i)%3 == 0 {
					out = append(out, '.')
				}
				out = append(out, byte(c))
			}
			if v < 0 {
				return "Rp -" + string(out)
			}
			return "Rp " + string(out)
		},
		"inc": func(i int) int { return i + 1 },
		"gt": func(a, b float64) bool { return a > b },
		"len": func(v interface{}) int {
			switch val := v.(type) {
			case []reportDay:
				return len(val)
			case []reportExpense:
				return len(val)
			default:
				return 0
			}
		},
	}

	tpl, err := template.New("travel_report").Funcs(funcMap).Parse(travelPlanReportHTMLTemplate)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse travel report template: %w", err)
	}

	data := reportData{
		Code:          plan.Code,
		CompanyName:   "GIMS",
		CompanyAddress: "GiLabs Integrated Management Suite",
		CompanyPhone:  "-",
		CompanyEmail:  "-",
		PlanType:      titleCase(string(plan.PlanType)),
		Mode:          titleCase(string(plan.Mode)),
		Status:        titleCase(string(plan.Status)),
		StartDate:     formatDateStr(plan.StartDate),
		EndDate:       formatDateStr(plan.EndDate),
		BudgetAmount:  plan.BudgetAmount,
		Notes:         strings.TrimSpace(plan.Notes),
		PrintDate:     apptime.Now().Format("02 Jan 2006 15:04"),
		Days:          days,
		Expenses:      expenseRows,
		TotalExpense:  totalExpense,
		Remaining:     remaining,
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, "", fmt.Errorf("failed to render travel report: %w", err)
	}

	filename := fmt.Sprintf("travel-plan-%s-report.html", strings.ToLower(plan.Code))
	return buf.Bytes(), filename, nil
}

func (uc *travelPlanUsecase) ListExpenses(ctx context.Context, planID string) (*dto.TravelExpenseListResponse, error) {
	planID = strings.TrimSpace(planID)
	if _, err := uc.repo.FindByID(ctx, planID, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTravelPlanNotFound
		}
		return nil, err
	}

	expenses, err := uc.repo.ListExpenses(ctx, planID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.TravelExpenseResponse, 0, len(expenses))
	totalAmount := 0.0
	for i := range expenses {
		totalAmount += expenses[i].Amount
		items = append(items, mapTravelExpenseToResponse(&expenses[i]))
	}

	return &dto.TravelExpenseListResponse{
		Items:       items,
		TotalAmount: roundToTwo(totalAmount),
	}, nil
}

func (uc *travelPlanUsecase) CreateExpense(ctx context.Context, planID string, req *dto.CreateTravelExpenseRequest) (*dto.TravelExpenseResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	planID = strings.TrimSpace(planID)
	if _, err := uc.repo.FindByID(ctx, planID, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTravelPlanNotFound
		}
		return nil, err
	}

	expenseType, err := parseExpenseType(req.ExpenseType)
	if err != nil {
		return nil, err
	}

	expenseDate, err := parseDate(req.ExpenseDate)
	if err != nil {
		return nil, ErrInvalidDayDate
	}

	actorID := strings.TrimSpace(getActorID(ctx))
	var createdBy *string
	if actorID != "" {
		createdBy = &actorID
	}

	expense := &models.TravelPlanExpense{
		TravelPlanID: planID,
		ExpenseType:  expenseType,
		Description:  strings.TrimSpace(req.Description),
		Amount:       req.Amount,
		ExpenseDate:  expenseDate,
		ReceiptURL:   strings.TrimSpace(req.ReceiptURL),
		CreatedBy:    createdBy,
	}

	if err := uc.repo.CreateExpense(ctx, expense); err != nil {
		return nil, err
	}

	response := mapTravelExpenseToResponse(expense)
	return &response, nil
}

func (uc *travelPlanUsecase) DeleteExpense(ctx context.Context, planID string, expenseID string) error {
	planID = strings.TrimSpace(planID)
	expenseID = strings.TrimSpace(expenseID)

	if _, err := uc.repo.FindByID(ctx, planID, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTravelPlanNotFound
		}
		return err
	}

	expenses, err := uc.repo.ListExpenses(ctx, planID)
	if err != nil {
		return err
	}

	found := false
	for _, expense := range expenses {
		if expense.ID == expenseID {
			found = true
			break
		}
	}
	if !found {
		return ErrTravelExpenseNotFound
	}

	return uc.repo.DeleteExpense(ctx, planID, expenseID)
}

func (uc *travelPlanUsecase) ListVisits(ctx context.Context, planID string) ([]dto.TravelPlanVisitResponse, error) {
	planID = strings.TrimSpace(planID)
	if _, err := uc.repo.FindByID(ctx, planID, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTravelPlanNotFound
		}
		return nil, err
	}

	visits := make([]crmModels.VisitReport, 0)
	err := uc.db.WithContext(ctx).
		Preload("Customer").
		Preload("Employee").
		Preload("Employee.User").
		Where("travel_plan_id = ?", planID).
		Order("visit_date DESC").
		Order("created_at DESC").
		Find(&visits).Error
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TravelPlanVisitResponse, 0, len(visits))
	for i := range visits {
		responses = append(responses, mapTravelPlanVisitToResponse(&visits[i]))
	}

	return responses, nil
}

func (uc *travelPlanUsecase) ListAvailableVisits(ctx context.Context, search string) ([]dto.TravelPlanVisitResponse, error) {
	trimmedSearch := strings.TrimSpace(search)
	if trimmedSearch != "" && len(trimmedSearch) < 2 {
		return nil, ErrInvalidSearchQuery
	}

	query := uc.db.WithContext(ctx).
		Model(&crmModels.VisitReport{}).
		Preload("Customer").
		Preload("Employee").
		Preload("Employee.User").
		Where("travel_plan_id IS NULL").
		Order("visit_date DESC").
		Order("created_at DESC").
		Limit(50)

	if trimmedSearch != "" {
		like := trimmedSearch + "%"
		query = query.Where(
			"code ILIKE ? OR purpose ILIKE ? OR notes ILIKE ? OR contact_person ILIKE ?",
			like,
			like,
			like,
			like,
		)
	}

	visits := make([]crmModels.VisitReport, 0)
	if err := query.Find(&visits).Error; err != nil {
		return nil, err
	}

	responses := make([]dto.TravelPlanVisitResponse, 0, len(visits))
	for i := range visits {
		responses = append(responses, mapTravelPlanVisitToResponse(&visits[i]))
	}

	return responses, nil
}

func (uc *travelPlanUsecase) LinkVisits(ctx context.Context, planID string, req *dto.LinkTravelPlanVisitsRequest) (int64, error) {
	if req == nil || len(req.VisitIDs) == 0 {
		return 0, errors.New("visit_ids is required")
	}

	planID = strings.TrimSpace(planID)
	if _, err := uc.repo.FindByID(ctx, planID, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrTravelPlanNotFound
		}
		return 0, err
	}

	result := uc.db.WithContext(ctx).
		Model(&crmModels.VisitReport{}).
		Where("id IN ? AND (travel_plan_id IS NULL OR travel_plan_id = ?)", req.VisitIDs, planID).
		Updates(map[string]interface{}{
			"travel_plan_id": planID,
			"updated_at":     apptime.Now(),
		})
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, ErrVisitNotFound
	}

	return result.RowsAffected, nil
}

func (uc *travelPlanUsecase) UnlinkVisit(ctx context.Context, planID string, visitID string) error {
	planID = strings.TrimSpace(planID)
	visitID = strings.TrimSpace(visitID)

	if _, err := uc.repo.FindByID(ctx, planID, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTravelPlanNotFound
		}
		return err
	}

	result := uc.db.WithContext(ctx).
		Model(&crmModels.VisitReport{}).
		Where("id = ? AND travel_plan_id = ?", visitID, planID).
		Updates(map[string]interface{}{
			"travel_plan_id": nil,
			"updated_at":     apptime.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrVisitNotFound
	}

	return nil
}

func (uc *travelPlanUsecase) CreateVisitFromTrip(ctx context.Context, planID string, req *dto.CreateTravelPlanVisitRequest) (*dto.TravelPlanVisitResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	planID = strings.TrimSpace(planID)
	if _, err := uc.repo.FindByID(ctx, planID, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTravelPlanNotFound
		}
		return nil, err
	}

	visitDate, err := parseDate(req.VisitDate)
	if err != nil {
		return nil, ErrInvalidDayDate
	}

	code, err := uc.generateVisitCode(ctx, apptime.Now())
	if err != nil {
		return nil, err
	}

	actorID := strings.TrimSpace(getActorID(ctx))
	var createdBy *string
	if actorID != "" {
		createdBy = &actorID
	}

	visit := &crmModels.VisitReport{
		Code:          code,
		TravelPlanID:  &planID,
		VisitDate:     visitDate,
		EmployeeID:    strings.TrimSpace(req.EmployeeID),
		CustomerID:    req.CustomerID,
		ContactID:     req.ContactID,
		DealID:        req.DealID,
		LeadID:        req.LeadID,
		VillageID:     req.VillageID,
		ContactPerson: strings.TrimSpace(req.ContactPerson),
		ContactPhone:  strings.TrimSpace(req.ContactPhone),
		Address:       strings.TrimSpace(req.Address),
		Purpose:       strings.TrimSpace(req.Purpose),
		Notes:         strings.TrimSpace(req.Notes),
		Status:        crmModels.VisitReportStatusDraft,
		CreatedBy:     createdBy,
	}

	if err := uc.db.WithContext(ctx).Create(visit).Error; err != nil {
		return nil, err
	}

	created := crmModels.VisitReport{}
	if err := uc.db.WithContext(ctx).
		Preload("Customer").
		Preload("Employee").
		Preload("Employee.User").
		Where("id = ?", visit.ID).
		First(&created).Error; err != nil {
		return nil, err
	}

	response := mapTravelPlanVisitToResponse(&created)
	return &response, nil
}

func (uc *travelPlanUsecase) buildDays(dayRequests []dto.TravelPlanDayRequest) ([]models.TravelPlanDay, error) {
	days := make([]models.TravelPlanDay, 0, len(dayRequests))
	for _, dayRequest := range dayRequests {
		dayDate, err := parseDate(dayRequest.DayDate)
		if err != nil {
			return nil, ErrInvalidDayDate
		}

		stops := make([]models.TravelPlanStop, 0, len(dayRequest.Stops))
		for stopIndex, stopRequest := range dayRequest.Stops {
			category, err := parseStopCategory(stopRequest.Category)
			if err != nil {
				return nil, err
			}
			source, err := parseStopSource(stopRequest.Source)
			if err != nil {
				return nil, err
			}

			orderIndex := stopRequest.OrderIndex
			if orderIndex <= 0 {
				orderIndex = stopIndex + 1
			}

			stops = append(stops, models.TravelPlanStop{
				PlaceName:  strings.TrimSpace(stopRequest.PlaceName),
				Latitude:   stopRequest.Latitude,
				Longitude:  stopRequest.Longitude,
				Category:   category,
				OrderIndex: orderIndex,
				IsLocked:   stopRequest.IsLocked,
				Source:     source,
				PhotoURL:   strings.TrimSpace(stopRequest.PhotoURL),
				Note:       strings.TrimSpace(stopRequest.Note),
			})
		}
		sort.SliceStable(stops, func(i, j int) bool {
			return stops[i].OrderIndex < stops[j].OrderIndex
		})
		for stopIndex := range stops {
			stops[stopIndex].OrderIndex = stopIndex + 1
		}

		notes := make([]models.TravelPlanDayNote, 0, len(dayRequest.Notes))
		for noteIndex, noteRequest := range dayRequest.Notes {
			orderIndex := noteRequest.OrderIndex
			if orderIndex <= 0 {
				orderIndex = noteIndex + 1
			}
			notes = append(notes, models.TravelPlanDayNote{
				IconTag:    strings.TrimSpace(noteRequest.IconTag),
				NoteText:   strings.TrimSpace(noteRequest.NoteText),
				NoteTime:   normalizeNoteTime(noteRequest.NoteTime),
				OrderIndex: orderIndex,
			})
		}
		sort.SliceStable(notes, func(i, j int) bool {
			return notes[i].OrderIndex < notes[j].OrderIndex
		})
		for noteIndex := range notes {
			notes[noteIndex].OrderIndex = noteIndex + 1
		}

		days = append(days, models.TravelPlanDay{
			DayIndex: dayRequest.DayIndex,
			DayDate:  dayDate,
			Summary:  strings.TrimSpace(dayRequest.Summary),
			Stops:    stops,
			Notes:    notes,
		})
	}

	sort.SliceStable(days, func(i, j int) bool {
		return days[i].DayIndex < days[j].DayIndex
	})
	for dayIndex := range days {
		days[dayIndex].DayIndex = dayIndex + 1
	}

	return days, nil
}

func parseDate(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, ErrInvalidDayDate
	}
	return parsed, nil
}

func mergeParticipantMeta(notes string, participantIDs []string) string {
	cleanNotes := strings.TrimSpace(notes)
	metaRe := regexp.MustCompile(`\n?\[participants:[^\]]*\]$`)
	cleanNotes = strings.TrimSpace(metaRe.ReplaceAllString(cleanNotes, ""))

	ids := make([]string, 0, len(participantIDs))
	seen := make(map[string]struct{}, len(participantIDs))
	for _, participantID := range participantIDs {
		trimmedID := strings.TrimSpace(participantID)
		if trimmedID == "" {
			continue
		}
		if _, exists := seen[trimmedID]; exists {
			continue
		}
		seen[trimmedID] = struct{}{}
		ids = append(ids, trimmedID)
	}

	if len(ids) == 0 {
		return cleanNotes
	}

	participantMeta := fmt.Sprintf("[participants:%s]", strings.Join(ids, ","))
	if cleanNotes == "" {
		return participantMeta
	}

	return cleanNotes + "\n" + participantMeta
}

func parseTravelMode(mode string) (models.TravelMode, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	switch models.TravelMode(normalized) {
	case models.TravelModeLogistic, models.TravelModeCargo, models.TravelModeVessel, models.TravelModeMilestone:
		return models.TravelMode(normalized), nil
	default:
		return "", ErrInvalidTravelMode
	}
}

func parseTravelPlanType(planType string) (models.TravelPlanType, error) {
	normalized := strings.ToLower(strings.TrimSpace(planType))
	switch models.TravelPlanType(normalized) {
	case models.TravelPlanTypeUpCountryCost, models.TravelPlanTypeVisitReport:
		return models.TravelPlanType(normalized), nil
	default:
		return "", errors.New("invalid travel plan type")
	}
}

func parseTravelStatus(status string) (models.TravelPlanStatus, error) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch models.TravelPlanStatus(normalized) {
	case models.TravelPlanStatusDraft, models.TravelPlanStatusActive, models.TravelPlanStatusCompleted, models.TravelPlanStatusCancelled:
		return models.TravelPlanStatus(normalized), nil
	default:
		return "", ErrInvalidStatus
	}
}

func parseStopCategory(category string) (models.TravelStopCategory, error) {
	normalized := strings.ToLower(strings.TrimSpace(category))
	switch models.TravelStopCategory(normalized) {
	case models.TravelStopCategoryPickup,
		models.TravelStopCategoryDropoff,
		models.TravelStopCategoryRefuel,
		models.TravelStopCategoryCheckpoint,
		models.TravelStopCategoryRest,
		models.TravelStopCategoryCustom:
		return models.TravelStopCategory(normalized), nil
	default:
		return "", ErrInvalidStopCategory
	}
}

func parseStopSource(source string) (models.TravelStopSource, error) {
	normalized := strings.ToLower(strings.TrimSpace(source))
	if normalized == "" {
		return models.TravelStopSourceManual, nil
	}
	switch models.TravelStopSource(normalized) {
	case models.TravelStopSourceManual,
		models.TravelStopSourceGooglePlaces,
		models.TravelStopSourceOpenStreetMap:
		return models.TravelStopSource(normalized), nil
	default:
		return "", ErrInvalidStopSource
	}
}

func parseExpenseType(expenseType string) (models.TravelExpenseType, error) {
	normalized := strings.ToLower(strings.TrimSpace(expenseType))
	switch models.TravelExpenseType(normalized) {
	case models.TravelExpenseTypeTransport,
		models.TravelExpenseTypeAccommodation,
		models.TravelExpenseTypeMeal,
		models.TravelExpenseTypeFuel,
		models.TravelExpenseTypeToll,
		models.TravelExpenseTypeParking,
		models.TravelExpenseTypeOther:
		return models.TravelExpenseType(normalized), nil
	default:
		return "", ErrInvalidExpenseType
	}
}

func parseBudgetAmount(value float64) (float64, error) {
	if value < 0 {
		return 0, ErrInvalidBudgetAmount
	}
	return roundToTwo(value), nil
}

func normalizeNoteTime(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) >= 5 {
		return trimmed[:5]
	}
	return trimmed
}

func getActorID(ctx context.Context) string {
	actorID, _ := ctx.Value("user_id").(string)
	return strings.TrimSpace(actorID)
}

func optimizeStops(stops []models.TravelPlanStop) ([]models.TravelPlanStop, float64) {
	sortedStops := append([]models.TravelPlanStop{}, stops...)
	sort.SliceStable(sortedStops, func(i, j int) bool {
		return sortedStops[i].OrderIndex < sortedStops[j].OrderIndex
	})

	if len(sortedStops) <= 2 {
		for index := range sortedStops {
			sortedStops[index].OrderIndex = index + 1
		}
		return sortedStops, calculateTotalDistance(sortedStops)
	}

	lockedByPosition := make(map[int]models.TravelPlanStop)
	unlocked := make([]models.TravelPlanStop, 0, len(sortedStops))
	for index, stop := range sortedStops {
		if stop.IsLocked {
			lockedByPosition[index] = stop
			continue
		}
		unlocked = append(unlocked, stop)
	}

	if len(unlocked) <= 1 {
		for index := range sortedStops {
			sortedStops[index].OrderIndex = index + 1
		}
		return sortedStops, calculateTotalDistance(sortedStops)
	}

	optimizedUnlocked := nearestNeighborOrder(unlocked)
	optimized := make([]models.TravelPlanStop, len(sortedStops))
	unlockedIndex := 0
	for index := range optimized {
		if lockedStop, exists := lockedByPosition[index]; exists {
			optimized[index] = lockedStop
			continue
		}
		optimized[index] = optimizedUnlocked[unlockedIndex]
		unlockedIndex++
	}

	for index := range optimized {
		optimized[index].OrderIndex = index + 1
	}

	return optimized, calculateTotalDistance(optimized)
}

func nearestNeighborOrder(stops []models.TravelPlanStop) []models.TravelPlanStop {
	if len(stops) <= 1 {
		return append([]models.TravelPlanStop{}, stops...)
	}

	remaining := append([]models.TravelPlanStop{}, stops...)
	ordered := make([]models.TravelPlanStop, 0, len(stops))

	ordered = append(ordered, remaining[0])
	remaining = remaining[1:]

	for len(remaining) > 0 {
		last := ordered[len(ordered)-1]
		closestIndex := 0
		closestDistance := math.MaxFloat64
		for index, candidate := range remaining {
			distance := haversine(last.Latitude, last.Longitude, candidate.Latitude, candidate.Longitude)
			if distance < closestDistance {
				closestDistance = distance
				closestIndex = index
			}
		}

		ordered = append(ordered, remaining[closestIndex])
		remaining = append(remaining[:closestIndex], remaining[closestIndex+1:]...)
	}

	return ordered
}

func calculateTotalDistance(stops []models.TravelPlanStop) float64 {
	if len(stops) <= 1 {
		return 0
	}

	totalDistance := 0.0
	for index := 1; index < len(stops); index++ {
		previous := stops[index-1]
		current := stops[index]
		totalDistance += haversine(previous.Latitude, previous.Longitude, current.Latitude, current.Longitude)
	}

	return totalDistance
}

func roundToTwo(value float64) float64 {
	return math.Round(value*100) / 100
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	toRadians := func(value float64) float64 {
		return value * math.Pi / 180
	}

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	rLat1 := toRadians(lat1)
	rLat2 := toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func buildGoogleMapsURL(stops []models.TravelPlanStop) string {
	if len(stops) == 0 {
		return ""
	}

	sortedStops := append([]models.TravelPlanStop{}, stops...)
	sort.SliceStable(sortedStops, func(i, j int) bool {
		return sortedStops[i].OrderIndex < sortedStops[j].OrderIndex
	})

	origin := fmt.Sprintf("%.6f,%.6f", sortedStops[0].Latitude, sortedStops[0].Longitude)
	destination := origin
	if len(sortedStops) > 1 {
		lastStop := sortedStops[len(sortedStops)-1]
		destination = fmt.Sprintf("%.6f,%.6f", lastStop.Latitude, lastStop.Longitude)
	}

	waypoints := make([]string, 0)
	if len(sortedStops) > 2 {
		for _, stop := range sortedStops[1 : len(sortedStops)-1] {
			waypoints = append(waypoints, fmt.Sprintf("%.6f,%.6f", stop.Latitude, stop.Longitude))
			if len(waypoints) >= 10 {
				break
			}
		}
	}

	values := url.Values{}
	values.Set("api", "1")
	values.Set("origin", origin)
	values.Set("destination", destination)
	if len(waypoints) > 0 {
		values.Set("waypoints", strings.Join(waypoints, "|"))
	}

	return "https://www.google.com/maps/dir/?" + values.Encode()
}

func (uc *travelPlanUsecase) searchGooglePlaces(ctx context.Context, query string) ([]dto.PlaceSearchResult, error) {
	if uc.googleAPIKey == "" {
		return nil, errors.New("google places api key is not configured")
	}

	values := url.Values{}
	values.Set("query", query)
	values.Set("key", uc.googleAPIKey)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://maps.googleapis.com/maps/api/place/textsearch/json?"+values.Encode(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	response, err := uc.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("google places error status: %d", response.StatusCode)
	}

	var payload struct {
		Status  string `json:"status"`
		Results []struct {
			Name             string   `json:"name"`
			FormattedAddress string   `json:"formatted_address"`
			Types            []string `json:"types"`
			Rating           *float64 `json:"rating"`
			Geometry         struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
			Photos []struct {
				PhotoReference string `json:"photo_reference"`
			} `json:"photos"`
		} `json:"results"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if payload.Status != "OK" && payload.Status != "ZERO_RESULTS" {
		return nil, fmt.Errorf("google places status: %s", payload.Status)
	}

	results := make([]dto.PlaceSearchResult, 0, len(payload.Results))
	for _, place := range payload.Results {
		category := "other"
		if len(place.Types) > 0 {
			category = place.Types[0]
		}
		photoURL := ""
		if len(place.Photos) > 0 && strings.TrimSpace(place.Photos[0].PhotoReference) != "" {
			photoURL = "https://maps.googleapis.com/maps/api/place/photo?maxwidth=640&photoreference=" +
				url.QueryEscape(place.Photos[0].PhotoReference) +
				"&key=" + url.QueryEscape(uc.googleAPIKey)
		}

		results = append(results, dto.PlaceSearchResult{
			Provider:  "google_places",
			PlaceName: place.Name,
			Address:   place.FormattedAddress,
			Latitude:  place.Geometry.Location.Lat,
			Longitude: place.Geometry.Location.Lng,
			Category:  category,
			PhotoURL:  photoURL,
			Rating:    place.Rating,
		})
		if len(results) >= 20 {
			break
		}
	}

	return results, nil
}

func (uc *travelPlanUsecase) searchOpenStreetMap(ctx context.Context, query string) ([]dto.PlaceSearchResult, error) {
	values := url.Values{}
	values.Set("q", query)
	values.Set("format", "json")
	values.Set("addressdetails", "1")
	values.Set("limit", "20")

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://nominatim.openstreetmap.org/search?"+values.Encode(),
		nil,
	)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "GIMS-TravelPlanner/1.0")

	response, err := uc.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("openstreetmap error status: %d", response.StatusCode)
	}

	var payload []struct {
		DisplayName string `json:"display_name"`
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
		Type        string `json:"type"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	results := make([]dto.PlaceSearchResult, 0, len(payload))
	for _, place := range payload {
		latitude, err := strconv.ParseFloat(place.Lat, 64)
		if err != nil {
			continue
		}
		longitude, err := strconv.ParseFloat(place.Lon, 64)
		if err != nil {
			continue
		}

		placeName := place.DisplayName
		if firstPart := strings.Split(place.DisplayName, ","); len(firstPart) > 0 {
			placeName = strings.TrimSpace(firstPart[0])
		}

		results = append(results, dto.PlaceSearchResult{
			Provider:  "open_street_map",
			PlaceName: placeName,
			Address:   place.DisplayName,
			Latitude:  latitude,
			Longitude: longitude,
			Category:  strings.TrimSpace(place.Type),
		})
	}

	return results, nil
}

func mapTravelExpenseToResponse(expense *models.TravelPlanExpense) dto.TravelExpenseResponse {
	return dto.TravelExpenseResponse{
		ID:           expense.ID,
		TravelPlanID: expense.TravelPlanID,
		ExpenseType:  string(expense.ExpenseType),
		Description:  expense.Description,
		Amount:       expense.Amount,
		ExpenseDate:  expense.ExpenseDate.Format("2006-01-02"),
		ReceiptURL:   expense.ReceiptURL,
		CreatedBy:    expense.CreatedBy,
		CreatedAt:    expense.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    expense.UpdatedAt.Format(time.RFC3339),
	}
}

func mapTravelPlanVisitToResponse(visit *crmModels.VisitReport) dto.TravelPlanVisitResponse {
	customerName := ""
	if visit.Customer != nil {
		customerName = strings.TrimSpace(visit.Customer.Name)
	}

	employeeName := ""
	if visit.Employee != nil {
		employeeName = strings.TrimSpace(visit.Employee.Name)
	}

	resp := dto.TravelPlanVisitResponse{
		ID:           visit.ID,
		Code:         visit.Code,
		TravelPlanID: visit.TravelPlanID,
		VisitDate:    visit.VisitDate.Format("2006-01-02"),
		EmployeeID:   visit.EmployeeID,
		EmployeeName: employeeName,
		EmployeeAvatarURL: func() string {
			if visit.Employee != nil && visit.Employee.User != nil {
				return strings.TrimSpace(visit.Employee.User.AvatarURL)
			}
			return ""
		}(),
		CustomerID:           visit.CustomerID,
		CustomerName:         customerName,
		Status:               string(visit.Status),
		Purpose:              visit.Purpose,
		Outcome:              visit.Outcome,
		CreatedAt:            visit.CreatedAt.Format(time.RFC3339),
		Photos:               visit.Photos,
		Notes:                visit.Notes,
		Result:               visit.Result,
		ProductInterestCount: len(visit.Details),
	}

	// Map check-in/out timestamps and locations
	if visit.CheckInAt != nil {
		checkInStr := visit.CheckInAt.Format(time.RFC3339)
		resp.CheckInAt = &checkInStr
	}
	if visit.CheckOutAt != nil {
		checkOutStr := visit.CheckOutAt.Format(time.RFC3339)
		resp.CheckOutAt = &checkOutStr
	}
	resp.CheckInLocation = visit.CheckInLocation
	resp.CheckOutLocation = visit.CheckOutLocation

	return resp
}

func (uc *travelPlanUsecase) generateVisitCode(ctx context.Context, now time.Time) (string, error) {
	prefix := fmt.Sprintf("VISIT-%s-", now.Format("200601"))
	// Cryptographically secure random 6-digit suffix to avoid collisions
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%06d", prefix, n.Int64()), nil
}
