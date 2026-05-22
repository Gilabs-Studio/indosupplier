package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/report/data/repositories"
	"github.com/gilabs/gims/api/internal/report/domain/dto"
)

// SalesOverviewUsecase defines business logic for sales overview reports
type SalesOverviewUsecase interface {
	ListSalesRepPerformance(ctx context.Context, req dto.ListSalesRepPerformanceRequest) ([]dto.SalesRepPerformanceResponse, utils.PaginationResult, error)
	GetMonthlySalesOverview(ctx context.Context, req dto.MonthlySalesOverviewRequest) (*dto.MonthlySalesOverviewResponse, error)
	GetSalesRepDetail(ctx context.Context, employeeID string, req dto.GetSalesRepDetailRequest) (*dto.SalesRepDetailResponse, error)
	GetSalesRepCheckInLocations(ctx context.Context, employeeID string, req dto.GetSalesRepCheckInLocationsRequest) (*dto.SalesRepCheckInLocationsResponse, error)
	GetSalesRepProducts(ctx context.Context, employeeID string, req dto.ListSalesRepProductsRequest) ([]dto.SalesRepProductResponse, utils.PaginationResult, error)
	GetSalesRepCustomers(ctx context.Context, employeeID string, req dto.ListSalesRepCustomersRequest) ([]dto.SalesRepCustomerResponse, utils.PaginationResult, error)
	GetEmployeeDashboardMetrics(ctx context.Context, employeeID string, req dto.EmployeeDashboardMetricsRequest) (*dto.EmployeeDashboardMetricsResponse, error)
}

type salesOverviewUsecase struct {
	repo repositories.SalesOverviewRepository
}

func NewSalesOverviewUsecase(repo repositories.SalesOverviewRepository) SalesOverviewUsecase {
	return &salesOverviewUsecase{repo: repo}
}

// defaultDateRange returns default start/end dates (last 1 year)
func defaultDateRange() (time.Time, time.Time) {
	now := apptime.Now()
	return now.AddDate(-1, 0, 0), now
}

// parseDateRange parses start_date and end_date strings, falling back to defaults
func parseDateRange(startStr, endStr string) (time.Time, time.Time) {
	start, end := defaultDateRange()
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

func formatCurrencyIDR(amount float64) string {
	// Format as Indonesian Rupiah without decimal places, no abbreviations
	rounded := math.Round(amount)
	return fmt.Sprintf("Rp %s", formatNumber(rounded))
}

func formatNumber(n float64) string {
	s := fmt.Sprintf("%.0f", n)
	// Insert thousand separators
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return result
}

// monthName returns Indonesian month name
func monthName(month int) string {
	names := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	if month >= 1 && month <= 12 {
		return names[month]
	}
	return ""
}

func (uc *salesOverviewUsecase) ListSalesRepPerformance(ctx context.Context, req dto.ListSalesRepPerformanceRequest) ([]dto.SalesRepPerformanceResponse, utils.PaginationResult, error) {
	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)

	params := repositories.ListPerformanceParams{
		Search:    req.Search,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	}

	rows, pagination, err := uc.repo.ListSalesRepPerformance(ctx, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	// Batch-fetch targets for all visible employees in a single query to avoid N+1.
	employeeIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		employeeIDs = append(employeeIDs, row.EmployeeID)
	}
	targetAmounts, _ := uc.repo.GetEmployeeTargetAmounts(ctx, employeeIDs, startDate, endDate)

	results := make([]dto.SalesRepPerformanceResponse, 0, len(rows))
	for _, row := range rows {
		conversionRate := float64(0)
		if row.VisitsCompleted > 0 {
			conversionRate = math.Round((float64(row.TotalOrders)/float64(row.VisitsCompleted))*10000) / 100
		}

		avgOrderValue := float64(0)
		if row.TotalOrders > 0 {
			avgOrderValue = row.TotalRevenue / float64(row.TotalOrders)
		}

		resp := dto.SalesRepPerformanceResponse{
			EmployeeID:                 row.EmployeeID,
			EmployeeCode:               row.EmployeeCode,
			Name:                       row.Name,
			Email:                      row.Email,
			AvatarURL:                  row.AvatarURL,
			PositionName:               row.PositionName,
			DivisionName:               row.DivisionName,
			TotalRevenue:               row.TotalRevenue,
			TotalRevenueFormatted:      formatCurrencyIDR(row.TotalRevenue),
			TotalOrders:                row.TotalOrders,
			TotalDeliveries:            row.TotalDeliveries,
			TotalInvoices:              row.TotalInvoices,
			VisitsCompleted:            row.VisitsCompleted,
			TasksCompleted:             row.TasksCompleted,
			ConversionRate:             conversionRate,
			AverageOrderValue:          avgOrderValue,
			AverageOrderValueFormatted: formatCurrencyIDR(avgOrderValue),
		}

		// Always provide target fields (0 if none found) to avoid nulls in UI.
		target := targetAmounts[row.EmployeeID]
		formatted := formatCurrencyIDR(target)
		achievement := float64(0)
		if target > 0 {
			achievement = math.Round((row.TotalRevenue/target)*10000) / 100
		}
		resp.TargetAmount = &target
		resp.TargetAmountFormatted = &formatted
		resp.TargetAchievementPercentage = &achievement

		results = append(results, resp)
	}

	return results, pagination, nil
}

func (uc *salesOverviewUsecase) GetMonthlySalesOverview(ctx context.Context, req dto.MonthlySalesOverviewRequest) (*dto.MonthlySalesOverviewResponse, error) {
	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)

	rows, err := uc.repo.GetMonthlySalesOverview(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get monthly targets for the same period
	targets, _ := uc.repo.GetMonthlyTargets(ctx, startDate, endDate)
	targetMap := make(map[string]float64)
	for _, t := range targets {
		key := fmt.Sprintf("%d-%d", t.Year, t.Month)
		targetMap[key] = t.TargetAmount
	}

	var totalRevenue float64
	var totalCashIn float64
	var totalOrders, totalVisits, totalDeliveries int

	monthlyData := make([]dto.MonthlySalesDataResponse, 0, len(rows))
	for _, row := range rows {
		key := fmt.Sprintf("%d-%d", row.Year, row.Month)
		targetAmount := targetMap[key]

		monthlyData = append(monthlyData, dto.MonthlySalesDataResponse{
			Month:           row.Month,
			MonthName:       monthName(row.Month),
			Year:            row.Year,
			TotalRevenue:    row.TotalRevenue,
			TotalCashIn:     row.TotalCashIn,
			TotalOrders:     row.TotalOrders,
			TotalVisits:     row.TotalVisits,
			TotalDeliveries: row.TotalDeliveries,
			TargetAmount:    targetAmount,
		})

		totalRevenue += row.TotalRevenue
		totalCashIn += row.TotalCashIn
		totalOrders += row.TotalOrders
		totalVisits += row.TotalVisits
		totalDeliveries += row.TotalDeliveries
	}

	return &dto.MonthlySalesOverviewResponse{
		MonthlyData:     monthlyData,
		TotalRevenue:    totalRevenue,
		TotalCashIn:     totalCashIn,
		TotalOrders:     totalOrders,
		TotalVisits:     totalVisits,
		TotalDeliveries: totalDeliveries,
	}, nil
}

func (uc *salesOverviewUsecase) GetSalesRepDetail(ctx context.Context, employeeID string, req dto.GetSalesRepDetailRequest) (*dto.SalesRepDetailResponse, error) {
	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)

	row, err := uc.repo.GetSalesRepDetail(ctx, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}

	// Use the same target source as list endpoint to keep target values consistent.
	targetAmounts, err := uc.repo.GetEmployeeTargetAmounts(ctx, []string{employeeID}, startDate, endDate)
	if err != nil {
		return nil, err
	}
	targetFromRange := targetAmounts[employeeID]

	conversionRate := float64(0)
	if row.VisitsCompleted > 0 {
		conversionRate = math.Round((float64(row.TotalOrders)/float64(row.VisitsCompleted))*10000) / 100
	}

	avgOrderValue := float64(0)
	if row.TotalOrders > 0 {
		avgOrderValue = row.TotalRevenue / float64(row.TotalOrders)
	}

	// Calculate target achievement percentage
	targetAmount := targetFromRange
	formatted := formatCurrencyIDR(targetAmount)
	targetAmountFormatted := &formatted
	targetAchievementPercent := float64(0)
	if targetAmount > 0 {
		achievementPercent := math.Round((row.TotalRevenue/targetAmount)*10000) / 100
		targetAchievementPercent = achievementPercent
	}

	stats := &dto.SalesRepStatisticsResponse{
		TotalRevenue:               row.TotalRevenue,
		TotalRevenueFormatted:      formatCurrencyIDR(row.TotalRevenue),
		TotalOrders:                row.TotalOrders,
		VisitsCompleted:            row.VisitsCompleted,
		TasksCompleted:             row.TasksCompleted,
		ConversionRate:             conversionRate,
		AverageOrderValue:          avgOrderValue,
		AverageOrderValueFormatted: formatCurrencyIDR(avgOrderValue),
		TargetAmount:               &targetAmount,
		TargetAmountFormatted:      targetAmountFormatted,
		TargetAchievementPercent:   &targetAchievementPercent,
	}

	// Period comparison: calculate percentage change vs previous period
	if row.PrevRevenue > 0 || row.PrevOrders > 0 || row.PrevVisits > 0 {
		revenueChange := float64(0)
		if row.PrevRevenue > 0 {
			revenueChange = math.Round(((row.TotalRevenue-row.PrevRevenue)/row.PrevRevenue)*10000) / 100
		}
		ordersChange := float64(0)
		if row.PrevOrders > 0 {
			ordersChange = math.Round(((float64(row.TotalOrders)-float64(row.PrevOrders))/float64(row.PrevOrders))*10000) / 100
		}
		visitsChange := float64(0)
		if row.PrevVisits > 0 {
			visitsChange = math.Round(((float64(row.VisitsCompleted)-float64(row.PrevVisits))/float64(row.PrevVisits))*10000) / 100
		}

		stats.PeriodComparison = &dto.PeriodComparisonData{
			RevenueChange: revenueChange,
			OrdersChange:  ordersChange,
			VisitsChange:  visitsChange,
		}
	}

	return &dto.SalesRepDetailResponse{
		EmployeeID:   row.EmployeeID,
		EmployeeCode: row.EmployeeCode,
		Name:         row.Name,
		Email:        row.Email,
		AvatarURL:    row.AvatarURL,
		PositionName: row.PositionName,
		DivisionName: row.DivisionName,
		Statistics:   stats,
	}, nil
}

func (uc *salesOverviewUsecase) GetSalesRepCheckInLocations(ctx context.Context, employeeID string, req dto.GetSalesRepCheckInLocationsRequest) (*dto.SalesRepCheckInLocationsResponse, error) {
	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)

	params := repositories.CheckInParams{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
	}

	rows, totalVisits, err := uc.repo.GetSalesRepCheckInLocations(ctx, employeeID, params)
	if err != nil {
		return nil, err
	}

	locations := make([]dto.CheckInLocationResponse, 0, len(rows))
	for i, row := range rows {
		loc := dto.CheckInLocationResponse{
			VisitNumber:   i + 1 + ((params.Page - 1) * params.PerPage),
			VisitReportID: row.VisitID,
			VisitDate:     row.VisitDate.Format("2006-01-02"),
			Purpose:       row.Purpose,
		}

		if row.CheckInAt != nil {
			loc.CheckInTime = row.CheckInAt.Format(time.RFC3339)
		}

		if row.Latitude != nil && row.Longitude != nil {
			loc.Location = &dto.LocationData{
				Latitude:  *row.Latitude,
				Longitude: *row.Longitude,
				Address:   row.Address,
			}
		}

		if row.CompanyID != nil && *row.CompanyID != "" {
			loc.Customer = &dto.CustomerRefData{
				ID:   *row.CompanyID,
				Name: row.CompanyName,
			}
		}

		locations = append(locations, loc)
	}

	return &dto.SalesRepCheckInLocationsResponse{
		CheckInLocations: locations,
		TotalVisits:      totalVisits,
		Period: &dto.PeriodData{
			Start: startDate.Format("2006-01-02"),
			End:   endDate.Format("2006-01-02"),
		},
	}, nil
}

func (uc *salesOverviewUsecase) GetSalesRepProducts(ctx context.Context, employeeID string, req dto.ListSalesRepProductsRequest) ([]dto.SalesRepProductResponse, utils.PaginationResult, error) {
	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)

	params := repositories.ProductParams{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	}

	rows, pagination, err := uc.repo.GetSalesRepProducts(ctx, employeeID, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.SalesRepProductResponse, 0, len(rows))
	for _, row := range rows {
		avgPrice := float64(0)
		if row.TotalQty > 0 {
			avgPrice = row.TotalRevenue / row.TotalQty
		}

		resp := dto.SalesRepProductResponse{
			ProductID:             row.ProductID,
			ProductName:           row.ProductName,
			ProductSKU:            row.ProductSKU,
			ProductImage:          row.ProductImage,
			CategoryName:          row.CategoryName,
			TotalQuantity:         row.TotalQty,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatCurrencyIDR(row.TotalRevenue),
			AveragePrice:          avgPrice,
			AveragePriceFormatted: formatCurrencyIDR(avgPrice),
		}

		if row.LastSoldDate != nil {
			resp.LastSoldDate = row.LastSoldDate.Format("2006-01-02")
		}

		results = append(results, resp)
	}

	return results, pagination, nil
}

func (uc *salesOverviewUsecase) GetSalesRepCustomers(ctx context.Context, employeeID string, req dto.ListSalesRepCustomersRequest) ([]dto.SalesRepCustomerResponse, utils.PaginationResult, error) {
	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)

	params := repositories.CustomerParams{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	}

	rows, pagination, err := uc.repo.GetSalesRepCustomers(ctx, employeeID, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.SalesRepCustomerResponse, 0, len(rows))
	for _, row := range rows {
		status := "inactive"
		if row.IsActive {
			status = "active"
		}

		results = append(results, dto.SalesRepCustomerResponse{
			CustomerID:            row.CustomerID,
			CustomerName:          row.CustomerName,
			CustomerCode:          row.CustomerCode,
			CustomerType:          row.CustomerType,
			City:                  row.CityName,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatCurrencyIDR(row.TotalRevenue),
			TotalOrders:           row.TotalOrders,
			Status:                status,
		})
	}

	return results, pagination, nil
}

// GetEmployeeDashboardMetrics returns aggregated metrics for employee dashboard/profile
func (uc *salesOverviewUsecase) GetEmployeeDashboardMetrics(ctx context.Context, employeeID string, req dto.EmployeeDashboardMetricsRequest) (*dto.EmployeeDashboardMetricsResponse, error) {
	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)

	result := &dto.EmployeeDashboardMetricsResponse{}

	// Get check-in locations summary
	checkInParams := repositories.CheckInParams{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      1,
		PerPage:   100, // Get all records for summary
	}
	checkInRows, totalVisits, err := uc.repo.GetSalesRepCheckInLocations(ctx, employeeID, checkInParams)
	if err == nil {
		result.CheckInLocations = &dto.CheckInLocationsSummary{
			TotalLocations: len(checkInRows),
			TotalVisits:    totalVisits,
			Period: &dto.PeriodData{
				Start: startDate.Format("2006-01-02"),
				End:   endDate.Format("2006-01-02"),
			},
		}
	}

	// Get products sold summary
	productParams := repositories.ProductParams{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      1,
		PerPage:   100, // Get all records for summary
		SortBy:    "revenue",
		Order:     "desc",
	}
	productRows, _, err := uc.repo.GetSalesRepProducts(ctx, employeeID, productParams)
	if err == nil {
		totalProducts := len(productRows)
		totalQty := 0.0
		totalRev := 0.0
		for _, row := range productRows {
			totalQty += row.TotalQty
			totalRev += row.TotalRevenue
		}
		avgRev := 0.0
		if totalProducts > 0 {
			avgRev = totalRev / float64(totalProducts)
		}
		result.ProductsSold = &dto.ProductsSoldSummary{
			TotalProducts:           totalProducts,
			TotalQuantity:           totalQty,
			TotalRevenue:            totalRev,
			TotalRevenueFormatted:   formatCurrencyIDR(totalRev),
			AverageRevenue:          avgRev,
			AverageRevenueFormatted: formatCurrencyIDR(avgRev),
		}
	}

	// Get customers summary
	customerParams := repositories.CustomerParams{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      1,
		PerPage:   100, // Get all records for summary
		SortBy:    "revenue",
		Order:     "desc",
	}
	customerRows, _, err := uc.repo.GetSalesRepCustomers(ctx, employeeID, customerParams)
	if err == nil {
		totalCustomers := len(customerRows)
		totalRev := 0.0
		totalOrders := 0
		for _, row := range customerRows {
			totalRev += row.TotalRevenue
			totalOrders += row.TotalOrders
		}
		avgOrderValue := 0.0
		if totalOrders > 0 {
			avgOrderValue = totalRev / float64(totalOrders)
		}
		result.Customers = &dto.CustomersSummary{
			TotalCustomers:              totalCustomers,
			TotalRevenue:                totalRev,
			TotalRevenueFormatted:       formatCurrencyIDR(totalRev),
			AverageOrderValue:           avgOrderValue,
			AverageOrderValueFormatted:  formatCurrencyIDR(avgOrderValue),
			TotalOrders:                 totalOrders,
		}
	}

	return result, nil
}
