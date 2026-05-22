package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/gilabs/gims/api/internal/general/data/repositories"
	"github.com/gilabs/gims/api/internal/general/domain/dto"
	"gorm.io/gorm"
)

// DashboardUsecase defines the business logic interface for dashboard aggregation.
type DashboardUsecase interface {
	GetOverview(ctx context.Context, req dto.DashboardRequest) (*dto.DashboardOverviewResponse, error)
	GetOverviewByScope(ctx context.Context, req dto.DashboardRequest) (*dto.DashboardScopedOverviewResponse, error)
	GetLayout(ctx context.Context, userID, dashboardType string) (*dto.DashboardLayoutResponse, error)
	SaveLayout(ctx context.Context, userID string, req dto.DashboardLayoutSaveRequest) error
}

type dashboardUsecase struct {
	repo       repositories.DashboardRepository
	layoutRepo repositories.DashboardLayoutRepository
}

const dateLayout = "2006-01-02"

// NewDashboardUsecase creates a new DashboardUsecase.
func NewDashboardUsecase(repo repositories.DashboardRepository, layoutRepo repositories.DashboardLayoutRepository) DashboardUsecase {
	return &dashboardUsecase{repo: repo, layoutRepo: layoutRepo}
}

// formatIDR formats a float64 value as Indonesian Rupiah with compact notation.
func formatIDR(v float64) string {
	switch {
	case math.Abs(v) >= 1_000_000_000:
		return fmt.Sprintf("Rp %.1fM", v/1_000_000_000)
	case math.Abs(v) >= 1_000_000:
		return fmt.Sprintf("Rp %.1fjt", v/1_000_000)
	case math.Abs(v) >= 1_000:
		return fmt.Sprintf("Rp %.0frb", v/1_000)
	default:
		return fmt.Sprintf("Rp %.0f", v)
	}
}

// formatCount formats a numeric count as a compact string.
func formatCount(v float64) string {
	switch {
	case v >= 1_000_000:
		return fmt.Sprintf("%.1fM", v/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.1fK", v/1_000)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

// toKPICardFormatted adds a formatted string to a KPICard, using currency for
// monetary values and count notation for others.
func toKPICardFormatted(card dto.KPICard, isCurrency bool) dto.KPICard {
	if isCurrency {
		card.Formatted = formatIDR(card.Value)
	} else {
		card.Formatted = formatCount(card.Value)
	}
	return card
}

// toPeriodChartData converts a flat []PeriodChartPoint into the PeriodChartData
// format consumed by the frontend chart components.
func toPeriodChartData(points []dto.PeriodChartPoint, seriesLabel string, valueField string) dto.PeriodChartData {
	periods := make([]string, 0, len(points))
	data := make([]float64, 0, len(points))
	formatted := make([]string, 0, len(points))
	for _, p := range points {
		var v float64
		switch valueField {
		case "revenue":
			v = p.Revenue
		case "costs":
			v = p.Costs
		case "amount":
			v = p.Amount
		}
		periods = append(periods, p.Period)
		data = append(data, v)
		formatted = append(formatted, formatIDR(v))
	}
	return dto.PeriodChartData{
		Period: periods,
		Series: []dto.ChartSeriesData{{Label: seriesLabel, Data: data, Formatted: formatted}},
	}
}

// toRevenueVsCostsChartData merges two value fields into a dual-series PeriodChartData.
func toRevenueVsCostsChartData(points []dto.PeriodChartPoint, revLabel, costLabel string) dto.PeriodChartData {
	periods := make([]string, 0, len(points))
	revData := make([]float64, 0, len(points))
	costData := make([]float64, 0, len(points))
	revFmt := make([]string, 0, len(points))
	costFmt := make([]string, 0, len(points))
	for _, p := range points {
		periods = append(periods, p.Period)
		revData = append(revData, p.Revenue)
		costData = append(costData, p.Costs)
		revFmt = append(revFmt, formatIDR(p.Revenue))
		costFmt = append(costFmt, formatIDR(p.Costs))
	}
	return dto.PeriodChartData{
		Period: periods,
		Series: []dto.ChartSeriesData{
			{Label: revLabel, Data: revData, Formatted: revFmt},
			{Label: costLabel, Data: costData, Formatted: costFmt},
		},
	}
}

// GetOverview aggregates all dashboard data. Each section is fetched independently
// so that a failure in one section does not prevent other sections from loading.
func (u *dashboardUsecase) GetOverview(ctx context.Context, req dto.DashboardRequest) (*dto.DashboardOverviewResponse, error) {
	start, end := resolveDateRange(req)
	resp := &dto.DashboardOverviewResponse{}
	u.loadKPI(ctx, start, end, resp)
	u.loadCharts(ctx, start, end, resp)
	u.loadBalance(ctx, start, end, resp)
	u.loadCostsByCategory(ctx, start, end, resp)
	u.loadInvoices(ctx, start, end, resp)
	u.loadSalesPerformance(ctx, start, end, resp)
	u.loadTopProducts(ctx, start, end, resp)
	u.loadDeliveryStatus(ctx, start, end, resp)
	u.loadGeoOverview(ctx, start, end, resp)
	u.loadWarehouseOverview(ctx, resp)

	return resp, nil
}

// GetOverviewByScope aggregates only the requested dashboard section.
func (u *dashboardUsecase) GetOverviewByScope(ctx context.Context, req dto.DashboardRequest) (*dto.DashboardScopedOverviewResponse, error) {
	if !req.Scope.IsValid() {
		return nil, fmt.Errorf("invalid dashboard scope: %s", req.Scope)
	}

	start, end := resolveDateRange(req)
	overview := &dto.DashboardOverviewResponse{}

	switch req.Scope {
	case dto.DashboardOverviewScopeKPI:
		u.loadKPI(ctx, start, end, overview)
	case dto.DashboardOverviewScopeCharts:
		u.loadCharts(ctx, start, end, overview)
	case dto.DashboardOverviewScopeBalance:
		u.loadBalance(ctx, start, end, overview)
	case dto.DashboardOverviewScopeCosts:
		u.loadCostsByCategory(ctx, start, end, overview)
	case dto.DashboardOverviewScopeInvoices:
		u.loadInvoices(ctx, start, end, overview)
	case dto.DashboardOverviewScopeSalesPerformance:
		u.loadSalesPerformance(ctx, start, end, overview)
	case dto.DashboardOverviewScopeProducts:
		u.loadTopProducts(ctx, start, end, overview)
	case dto.DashboardOverviewScopeDelivery:
		u.loadDeliveryStatus(ctx, start, end, overview)
	case dto.DashboardOverviewScopeGeo:
		u.loadGeoOverview(ctx, start, end, overview)
	case dto.DashboardOverviewScopeWarehouse:
		u.loadWarehouseOverview(ctx, overview)
	case dto.DashboardOverviewScopePOS:
		return u.loadPOSScopedResponse(ctx, req)
	case dto.DashboardOverviewScopeHR:
		return u.loadHRScopedResponse(ctx)
	case dto.DashboardOverviewScopeCRM:
		return u.loadCRMScopedResponse(ctx)
	}

	return toScopedOverviewResponse(overview, req.Scope), nil
}

func toScopedOverviewResponse(
	overview *dto.DashboardOverviewResponse,
	scope dto.DashboardOverviewScope,
) *dto.DashboardScopedOverviewResponse {
	resp := &dto.DashboardScopedOverviewResponse{}

	switch scope {
	case dto.DashboardOverviewScopeKPI:
		resp.KPI = &overview.KPI
	case dto.DashboardOverviewScopeCharts:
		resp.RevenueChart = &overview.RevenueChart
		resp.CostsChart = &overview.CostsChart
		resp.RevenueVsCosts = &overview.RevenueVsCosts
	case dto.DashboardOverviewScopeBalance:
		resp.BalanceOverview = &overview.BalanceOverview
	case dto.DashboardOverviewScopeCosts:
		resp.CostsByCategory = overview.CostsByCategory
	case dto.DashboardOverviewScopeInvoices:
		resp.InvoicesSummary = &overview.InvoicesSummary
		resp.RecentInvoices = overview.RecentInvoices
	case dto.DashboardOverviewScopeSalesPerformance:
		resp.SalesPerformance = overview.SalesPerformance
	case dto.DashboardOverviewScopeProducts:
		resp.TopProducts = overview.TopProducts
	case dto.DashboardOverviewScopeDelivery:
		resp.DeliveryStatus = &overview.DeliveryStatus
	case dto.DashboardOverviewScopeGeo:
		resp.GeographicOverview = &overview.GeographicOverview
	case dto.DashboardOverviewScopeWarehouse:
		resp.WarehouseOverview = &overview.WarehouseOverview
	}

	return resp
}

func (u *dashboardUsecase) loadKPI(ctx context.Context, start, end time.Time, resp *dto.DashboardOverviewResponse) {
	if kpi, err := u.repo.GetRevenueKPI(ctx, start, end); err != nil {
		log.Printf("[dashboard] revenue KPI error: %v", err)
	} else {
		resp.KPI.TotalRevenue = toKPICardFormatted(kpi, true)
	}
	if kpi, err := u.repo.GetOrdersKPI(ctx, start, end); err != nil {
		log.Printf("[dashboard] orders KPI error: %v", err)
	} else {
		resp.KPI.TotalOrders = toKPICardFormatted(kpi, false)
	}
	if kpi, err := u.repo.GetCustomersKPI(ctx); err != nil {
		log.Printf("[dashboard] customers KPI error: %v", err)
	} else {
		resp.KPI.TotalCustomers = toKPICardFormatted(kpi, false)
	}
	if kpi, err := u.repo.GetProductsKPI(ctx); err != nil {
		log.Printf("[dashboard] products KPI error: %v", err)
	} else {
		resp.KPI.TotalProducts = toKPICardFormatted(kpi, false)
	}
	if kpi, err := u.repo.GetEmployeeCountKPI(ctx); err != nil {
		log.Printf("[dashboard] employee count KPI error: %v", err)
	} else {
		resp.KPI.EmployeeCount = toKPICardFormatted(kpi, false)
	}
}

func (u *dashboardUsecase) loadCharts(ctx context.Context, start, end time.Time, resp *dto.DashboardOverviewResponse) {
	if data, err := u.repo.GetRevenueChart(ctx, start, end); err != nil {
		log.Printf("[dashboard] revenue chart error: %v", err)
		resp.RevenueChart = dto.PeriodChartData{Series: []dto.ChartSeriesData{}, Period: []string{}}
	} else {
		resp.RevenueChart = toPeriodChartData(data, "Revenue", "revenue")
	}
	if data, err := u.repo.GetCostsChart(ctx, start, end); err != nil {
		log.Printf("[dashboard] costs chart error: %v", err)
		resp.CostsChart = dto.PeriodChartData{Series: []dto.ChartSeriesData{}, Period: []string{}}
	} else {
		resp.CostsChart = toPeriodChartData(data, "Costs", "costs")
	}
	if data, err := u.repo.GetRevenueVsCosts(ctx, start, end); err != nil {
		log.Printf("[dashboard] revenue vs costs error: %v", err)
		resp.RevenueVsCosts = dto.PeriodChartData{Series: []dto.ChartSeriesData{}, Period: []string{}}
	} else {
		resp.RevenueVsCosts = toRevenueVsCostsChartData(data, "Revenue", "Costs")
	}
}

func (u *dashboardUsecase) loadBalance(ctx context.Context, start, end time.Time, resp *dto.DashboardOverviewResponse) {
	if data, err := u.repo.GetBalance(ctx, start, end); err != nil {
		log.Printf("[dashboard] balance error: %v", err)
		resp.BalanceOverview = dto.BalanceOverviewData{ChartData: []dto.BalanceChartPoint{}}
	} else {
		chartData := make([]dto.BalanceChartPoint, 0, len(data.Trend))
		for _, p := range data.Trend {
			chartData = append(chartData, dto.BalanceChartPoint{
				Period:    p.Period,
				Value:     p.Amount,
				Formatted: formatIDR(p.Amount),
			})
		}
		resp.BalanceOverview = dto.BalanceOverviewData{
			Value:         data.Current,
			Formatted:     formatIDR(data.Current),
			ChangePercent: data.Change,
			ChartData:     chartData,
		}
	}
}

func (u *dashboardUsecase) loadCostsByCategory(ctx context.Context, start, end time.Time, resp *dto.DashboardOverviewResponse) {
	if data, err := u.repo.GetCostsByCategory(ctx, start, end); err != nil {
		log.Printf("[dashboard] costs by category error: %v", err)
		resp.CostsByCategory = []dto.CostCategoryItem{}
	} else {
		for i := range data {
			data[i].AmountFormatted = formatIDR(data[i].Amount)
		}
		resp.CostsByCategory = data
	}
}

func (u *dashboardUsecase) loadInvoices(ctx context.Context, start, end time.Time, resp *dto.DashboardOverviewResponse) {
	if data, err := u.repo.GetInvoiceSummary(ctx, start, end); err != nil {
		log.Printf("[dashboard] invoice summary error: %v", err)
	} else {
		resp.InvoicesSummary = data
	}
	if data, err := u.repo.GetRecentInvoices(ctx, 10); err != nil {
		log.Printf("[dashboard] recent invoices error: %v", err)
		resp.RecentInvoices = []dto.InvoiceRow{}
	} else {
		for i := range data {
			data[i].ValueFormatted = formatIDR(data[i].Value)
		}
		resp.RecentInvoices = data
	}
}

func (u *dashboardUsecase) loadSalesPerformance(ctx context.Context, start, end time.Time, resp *dto.DashboardOverviewResponse) {
	if data, err := u.repo.GetSalesPerformance(ctx, start, end, 5); err != nil {
		log.Printf("[dashboard] sales performance error: %v", err)
		resp.SalesPerformance = []dto.SalesPerformanceRow{}
	} else {
		for i := range data {
			data[i].RevenueFormatted = formatIDR(data[i].Revenue)
		}
		resp.SalesPerformance = data
	}
}

func (u *dashboardUsecase) loadTopProducts(ctx context.Context, start, end time.Time, resp *dto.DashboardOverviewResponse) {
	if data, err := u.repo.GetTopProducts(ctx, start, end, 6); err != nil {
		log.Printf("[dashboard] top products error: %v", err)
		resp.TopProducts = []dto.TopProductRow{}
	} else {
		for i := range data {
			data[i].RevenueFormatted = formatIDR(data[i].Revenue)
		}
		resp.TopProducts = data
	}
}

func (u *dashboardUsecase) loadDeliveryStatus(ctx context.Context, start, end time.Time, resp *dto.DashboardOverviewResponse) {
	if data, err := u.repo.GetDeliveryStatus(ctx, start, end); err != nil {
		log.Printf("[dashboard] delivery status error: %v", err)
	} else {
		resp.DeliveryStatus = data
	}
}

func (u *dashboardUsecase) loadGeoOverview(ctx context.Context, start, end time.Time, resp *dto.DashboardOverviewResponse) {
	if data, err := u.repo.GetGeoOverview(ctx, start, end); err != nil {
		log.Printf("[dashboard] geo overview error: %v", err)
		resp.GeographicOverview = dto.GeoOverviewData{Regions: []dto.GeoRegionData{}}
	} else {
		for i := range data.Regions {
			data.Regions[i].Formatted = formatIDR(data.Regions[i].Value)
		}
		data.TotalFormatted = formatIDR(data.TotalValue)
		resp.GeographicOverview = data
	}
}

func (u *dashboardUsecase) loadWarehouseOverview(ctx context.Context, resp *dto.DashboardOverviewResponse) {
	if data, err := u.repo.GetWarehouses(ctx); err != nil {
		log.Printf("[dashboard] warehouses error: %v", err)
		resp.WarehouseOverview = dto.WarehouseOverviewData{Warehouses: []dto.WarehouseItem{}}
	} else {
		var totalStock float64
		for i := range data {
			data[i].StockFormatted = formatIDR(data[i].StockValue)
			totalStock += data[i].StockValue
		}
		resp.WarehouseOverview = dto.WarehouseOverviewData{
			Warehouses:          data,
			TotalStockValue:     totalStock,
			TotalStockFormatted: formatIDR(totalStock),
		}
	}
}

// resolveDateRange determines start/end dates from the request.
// Defaults to last 12 months when no range or year is specified.
func resolveDateRange(req dto.DashboardRequest) (time.Time, time.Time) {
	now := time.Now()
	if req.StartDate != "" && req.EndDate != "" {
		start, err1 := time.Parse(dateLayout, req.StartDate)
		end, err2 := time.Parse(dateLayout, req.EndDate)
		if err1 == nil && err2 == nil {
			return start, end.Add(24*time.Hour - time.Second)
		}
	}
	if req.Year > 0 {
		return time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(req.Year, 12, 31, 23, 59, 59, 0, time.UTC)
	}
	start := time.Date(now.Year()-1, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	return start, end
}

// resolvePOSDateRange determines the POS window. It defaults to the current day
// so POS widgets behave like operational dashboards unless the caller requests a range.
func resolvePOSDateRange(req dto.DashboardRequest) (time.Time, time.Time) {
	now := time.Now()
	if req.StartDate != "" && req.EndDate != "" {
		start, err1 := time.Parse(dateLayout, req.StartDate)
		end, err2 := time.Parse(dateLayout, req.EndDate)
		if err1 == nil && err2 == nil {
			return start, end.Add(24*time.Hour - time.Second)
		}
	}
	if req.Year > 0 {
		return time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(req.Year, 12, 31, 23, 59, 59, 0, time.UTC)
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
}

// GetLayout fetches the saved dashboard layout for a user. Returns nil if no layout has been saved yet.
func (u *dashboardUsecase) GetLayout(ctx context.Context, userID, dashboardType string) (*dto.DashboardLayoutResponse, error) {
	layoutJSON, err := u.layoutRepo.GetLayout(ctx, userID, dashboardType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &dto.DashboardLayoutResponse{
		DashboardType: dashboardType,
		LayoutJSON:    layoutJSON,
	}, nil
}

// SaveLayout persists the user's custom dashboard layout to the database.
func (u *dashboardUsecase) SaveLayout(ctx context.Context, userID string, req dto.DashboardLayoutSaveRequest) error {
	return u.layoutRepo.SaveLayout(ctx, userID, req.DashboardType, req.LayoutJSON)
}

// loadPOSScopedResponse returns POS summary with RBAC outlet filtering.
func (u *dashboardUsecase) loadPOSScopedResponse(ctx context.Context, req dto.DashboardRequest) (*dto.DashboardScopedOverviewResponse, error) {
	start, end := resolvePOSDateRange(req)
	summary, err := u.repo.GetPOSSummary(ctx, req.UserID, req.OutletID, start, end)
	if err != nil {
		log.Printf("[dashboard] pos summary error: %v", err)
		return nil, err
	}
	return &dto.DashboardScopedOverviewResponse{POSSummary: summary}, nil
}

// loadHRScopedResponse returns HR attendance and leave summary.
func (u *dashboardUsecase) loadHRScopedResponse(ctx context.Context) (*dto.DashboardScopedOverviewResponse, error) {
	summary, err := u.repo.GetHRSummary(ctx)
	if err != nil {
		log.Printf("[dashboard] hr summary error: %v", err)
		summary = &dto.HRSummaryData{}
	}
	return &dto.DashboardScopedOverviewResponse{HRSummary: summary}, nil
}

// loadCRMScopedResponse returns CRM aggregated metrics.
func (u *dashboardUsecase) loadCRMScopedResponse(ctx context.Context) (*dto.DashboardScopedOverviewResponse, error) {
	summary, err := u.repo.GetCRMSummary(ctx)
	if err != nil {
		log.Printf("[dashboard] crm summary error: %v", err)
		summary = &dto.CRMSummaryData{}
	}
	return &dto.DashboardScopedOverviewResponse{CRMSummary: summary}, nil
}
