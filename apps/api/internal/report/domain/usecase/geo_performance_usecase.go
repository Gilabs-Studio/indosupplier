package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/report/data/repositories"
	"github.com/gilabs/gims/api/internal/report/domain/dto"
)

// GeoPerformanceUsecase defines business logic for geo performance reports
type GeoPerformanceUsecase interface {
	GetGeoPerformance(ctx context.Context, req dto.GeoPerformanceRequest) (*dto.GeoPerformanceSummaryResponse, error)
	GetFormData(ctx context.Context) (*dto.GeoPerformanceFormDataResponse, error)
}

type geoPerformanceUsecase struct {
	repo repositories.GeoPerformanceRepository
}

// NewGeoPerformanceUsecase creates a new instance
func NewGeoPerformanceUsecase(repo repositories.GeoPerformanceRepository) GeoPerformanceUsecase {
	return &geoPerformanceUsecase{repo: repo}
}

func (uc *geoPerformanceUsecase) GetGeoPerformance(ctx context.Context, req dto.GeoPerformanceRequest) (*dto.GeoPerformanceSummaryResponse, error) {
	startDate, endDate := parseGeoDateRange(req.StartDate, req.EndDate)

	level := req.Level
	if level != "province" && level != "city" {
		level = "province"
	}

	params := repositories.GeoPerformanceParams{
		StartDate:  startDate,
		EndDate:    endDate,
		SalesRepID: req.SalesRepID,
		Level:      level,
	}

	var rows []repositories.GeoPerformanceRow
	var err error

	mode := req.Mode
	if mode != "sales_order" && mode != "paid_invoice" {
		mode = "sales_order"
	}

	switch mode {
	case "paid_invoice":
		rows, err = uc.repo.GetGeoPerformanceByPaidInvoice(ctx, params)
	default:
		rows, err = uc.repo.GetGeoPerformanceBySalesOrder(ctx, params)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get geo performance data: %w", err)
	}

	// Map rows to response DTOs and compute summary totals
	areas := make([]dto.GeoPerformanceAreaResponse, 0, len(rows))
	var totalRevenue float64
	var totalOrders int

	for _, row := range rows {
		avgOrderValue := 0.0
		if row.TotalOrders > 0 {
			avgOrderValue = math.Round(row.TotalRevenue / float64(row.TotalOrders))
		}

		areas = append(areas, dto.GeoPerformanceAreaResponse{
			AreaID:        row.AreaID,
			AreaName:      row.AreaName,
			ParentName:    row.ParentName,
			TotalRevenue:  row.TotalRevenue,
			TotalOrders:   row.TotalOrders,
			AvgOrderValue: avgOrderValue,
		})

		totalRevenue += row.TotalRevenue
		totalOrders += row.TotalOrders
	}

	overallAvg := 0.0
	if totalOrders > 0 {
		overallAvg = math.Round(totalRevenue / float64(totalOrders))
	}

	return &dto.GeoPerformanceSummaryResponse{
		Areas:         areas,
		TotalRevenue:  totalRevenue,
		TotalOrders:   totalOrders,
		AvgOrderValue: overallAvg,
		AreasWithData: len(areas),
		Mode:          mode,
		Level:         level,
	}, nil
}

func (uc *geoPerformanceUsecase) GetFormData(ctx context.Context) (*dto.GeoPerformanceFormDataResponse, error) {
	reps, err := uc.repo.GetSalesRepsForFilter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get form data: %w", err)
	}

	options := make([]dto.GeoSalesRepOption, 0, len(reps))
	for _, r := range reps {
		options = append(options, dto.GeoSalesRepOption{
			ID:   r.ID,
			Name: r.Name,
			Code: r.Code,
		})
	}

	return &dto.GeoPerformanceFormDataResponse{
		SalesReps: options,
	}, nil
}

// parseGeoDateRange parses start/end strings with a default of last 12 months
func parseGeoDateRange(startStr, endStr string) (time.Time, time.Time) {
	now := apptime.Now()
	start := now.AddDate(-1, 0, 0)
	end := now

	if startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			start = parsed
		}
	}
	if endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			end = parsed
		}
	}

	return start, end
}
