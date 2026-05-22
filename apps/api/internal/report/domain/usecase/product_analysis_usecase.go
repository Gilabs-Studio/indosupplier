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

// ProductAnalysisUsecase defines business logic for product analysis reports
type ProductAnalysisUsecase interface {
	ListProductPerformance(ctx context.Context, req dto.ListProductPerformanceRequest) ([]dto.ProductPerformanceResponse, utils.PaginationResult, error)
	GetMonthlyProductSales(ctx context.Context, req dto.MonthlyProductSalesRequest) (*dto.MonthlyProductSalesResponse, error)
	GetProductDetail(ctx context.Context, productID string, req dto.GetProductDetailRequest) (*dto.ProductDetailResponse, error)
	GetProductCustomers(ctx context.Context, productID string, req dto.ListProductCustomersRequest) ([]dto.ProductCustomerResponse, utils.PaginationResult, error)
	GetProductSalesReps(ctx context.Context, productID string, req dto.ListProductSalesRepsRequest) ([]dto.ProductSalesRepResponse, utils.PaginationResult, error)
	GetProductMonthlyTrend(ctx context.Context, productID string, req dto.GetProductMonthlyTrendRequest) (*dto.ProductMonthlyTrendResponse, error)
	ListCategoryPerformance(ctx context.Context, req dto.ListCategoryPerformanceRequest) ([]dto.CategoryPerformanceResponse, utils.PaginationResult, error)
	ListSegmentPerformance(ctx context.Context, req dto.ListSegmentPerformanceRequest) ([]dto.SegmentPerformanceResponse, utils.PaginationResult, error)
	ListTypePerformance(ctx context.Context, req dto.ListTypePerformanceRequest) ([]dto.TypePerformanceResponse, utils.PaginationResult, error)
	ListPackagingPerformance(ctx context.Context, req dto.ListPackagingPerformanceRequest) ([]dto.PackagingPerformanceResponse, utils.PaginationResult, error)
	ListProcurementTypePerformance(ctx context.Context, req dto.ListProcurementTypePerformanceRequest) ([]dto.ProcurementTypePerformanceResponse, utils.PaginationResult, error)
}

type productAnalysisUsecase struct {
	repo repositories.ProductAnalysisRepository
}

// NewProductAnalysisUsecase creates a new ProductAnalysisUsecase instance
func NewProductAnalysisUsecase(repo repositories.ProductAnalysisRepository) ProductAnalysisUsecase {
	return &productAnalysisUsecase{repo: repo}
}

func (uc *productAnalysisUsecase) ListProductPerformance(ctx context.Context, req dto.ListProductPerformanceRequest) ([]dto.ProductPerformanceResponse, utils.PaginationResult, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	params := repositories.ListProductPerformanceParams{
		Search:     req.Search,
		CategoryID: req.CategoryID,
		StartDate:  startDate,
		EndDate:    endDate,
		Page:       req.Page,
		PerPage:    req.PerPage,
		SortBy:     req.SortBy,
		Order:      req.Order,
	}

	rows, pagination, err := uc.repo.ListProductPerformance(ctx, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.ProductPerformanceResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.ProductPerformanceResponse{
			ProductID:             row.ProductID,
			ProductCode:           row.ProductCode,
			ProductName:           row.ProductName,
			ProductSKU:            row.ProductSKU,
			ProductImage:          row.ProductImage,
			CategoryName:          row.CategoryName,
			TotalQty:              row.TotalQty,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatProductCurrencyIDR(row.TotalRevenue),
			TotalOrders:           row.TotalOrders,
			AvgPrice:              row.AvgPrice,
			AvgPriceFormatted:     formatProductCurrencyIDR(row.AvgPrice),
		})
	}

	return results, pagination, nil
}

func (uc *productAnalysisUsecase) GetMonthlyProductSales(ctx context.Context, req dto.MonthlyProductSalesRequest) (*dto.MonthlyProductSalesResponse, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	rows, err := uc.repo.GetMonthlyProductSales(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var totalRevenue float64
	var totalQty float64
	var totalOrders int

	monthlyData := make([]dto.MonthlyProductSalesDataResponse, 0, len(rows))
	for _, row := range rows {
		totalRevenue += row.TotalRevenue
		totalQty += row.TotalQty
		totalOrders += row.TotalOrders

		monthlyData = append(monthlyData, dto.MonthlyProductSalesDataResponse{
			Month:        row.Month,
			MonthName:    productMonthName(row.Month),
			Year:         row.Year,
			TotalRevenue: row.TotalRevenue,
			TotalQty:     row.TotalQty,
			TotalOrders:  row.TotalOrders,
		})
	}

	return &dto.MonthlyProductSalesResponse{
		MonthlyData:  monthlyData,
		TotalRevenue: totalRevenue,
		TotalQty:     totalQty,
		TotalOrders:  totalOrders,
	}, nil
}

func (uc *productAnalysisUsecase) GetProductDetail(ctx context.Context, productID string, req dto.GetProductDetailRequest) (*dto.ProductDetailResponse, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	row, err := uc.repo.GetProductDetail(ctx, productID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("product not found")
	}

	avgPrice := float64(0)
	if row.TotalQty > 0 {
		avgPrice = row.TotalRevenue / row.TotalQty
	}

	// Calculate period comparison
	var comparison *dto.ProductPeriodChange
	if row.PrevRevenue > 0 || row.PrevQty > 0 || row.PrevOrders > 0 {
		revenueChange := float64(0)
		if row.PrevRevenue > 0 {
			revenueChange = math.Round(((row.TotalRevenue-row.PrevRevenue)/row.PrevRevenue)*10000) / 100
		}
		qtyChange := float64(0)
		if row.PrevQty > 0 {
			qtyChange = math.Round(((row.TotalQty-row.PrevQty)/row.PrevQty)*10000) / 100
		}
		ordersChange := float64(0)
		if row.PrevOrders > 0 {
			ordersChange = math.Round(((float64(row.TotalOrders)-float64(row.PrevOrders))/float64(row.PrevOrders))*10000) / 100
		}
		comparison = &dto.ProductPeriodChange{
			RevenueChange: revenueChange,
			QtyChange:     qtyChange,
			OrdersChange:  ordersChange,
		}
	}

	return &dto.ProductDetailResponse{
		ProductID:    row.ProductID,
		ProductCode:  row.ProductCode,
		ProductName:  row.ProductName,
		ProductSKU:   row.ProductSKU,
		ProductImage: row.ProductImage,
		CategoryName: row.CategoryName,
		BrandName:    row.BrandName,
		SellingPrice: row.SellingPrice,
		CostPrice:    row.CostPrice,
		CurrentStock: row.CurrentStock,
		Statistics: &dto.ProductDetailStatistics{
			TotalQty:              row.TotalQty,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatProductCurrencyIDR(row.TotalRevenue),
			TotalOrders:           row.TotalOrders,
			AvgPrice:              avgPrice,
			AvgPriceFormatted:     formatProductCurrencyIDR(avgPrice),
			PeriodComparison:      comparison,
		},
	}, nil
}

func (uc *productAnalysisUsecase) GetProductCustomers(ctx context.Context, productID string, req dto.ListProductCustomersRequest) ([]dto.ProductCustomerResponse, utils.PaginationResult, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	params := repositories.ProductCustomerParams{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	}

	rows, pagination, err := uc.repo.GetProductTopCustomers(ctx, productID, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.ProductCustomerResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.ProductCustomerResponse{
			CustomerID:            row.CustomerID,
			CustomerName:          row.CustomerName,
			CustomerCode:          row.CustomerCode,
			CustomerType:          row.CustomerType,
			CityName:              row.CityName,
			TotalQty:              row.TotalQty,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatProductCurrencyIDR(row.TotalRevenue),
			TotalOrders:           row.TotalOrders,
		})
	}

	return results, pagination, nil
}

func (uc *productAnalysisUsecase) GetProductSalesReps(ctx context.Context, productID string, req dto.ListProductSalesRepsRequest) ([]dto.ProductSalesRepResponse, utils.PaginationResult, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	params := repositories.ProductSalesRepParams{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	}

	rows, pagination, err := uc.repo.GetProductTopSalesReps(ctx, productID, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.ProductSalesRepResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.ProductSalesRepResponse{
			EmployeeID:            row.EmployeeID,
			EmployeeCode:          row.EmployeeCode,
			Name:                  row.Name,
			AvatarURL:             row.AvatarURL,
			PositionName:          row.PositionName,
			TotalQty:              row.TotalQty,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatProductCurrencyIDR(row.TotalRevenue),
			TotalOrders:           row.TotalOrders,
		})
	}

	return results, pagination, nil
}

func (uc *productAnalysisUsecase) GetProductMonthlyTrend(ctx context.Context, productID string, req dto.GetProductMonthlyTrendRequest) (*dto.ProductMonthlyTrendResponse, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	rows, err := uc.repo.GetProductMonthlyTrend(ctx, productID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	monthlyData := make([]dto.ProductMonthlyTrendDataResponse, 0, len(rows))
	for _, row := range rows {
		monthlyData = append(monthlyData, dto.ProductMonthlyTrendDataResponse{
			Month:        row.Month,
			MonthName:    productMonthName(row.Month),
			Year:         row.Year,
			TotalRevenue: row.TotalRevenue,
			TotalQty:     row.TotalQty,
			TotalOrders:  row.TotalOrders,
		})
	}

	// Retrieve product name from detail query (basic info only)
	detail, err := uc.repo.GetProductDetail(ctx, productID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	productName := ""
	if detail != nil {
		productName = detail.ProductName
	}

	return &dto.ProductMonthlyTrendResponse{
		ProductID:   productID,
		ProductName: productName,
		MonthlyData: monthlyData,
	}, nil
}

func (uc *productAnalysisUsecase) ListCategoryPerformance(ctx context.Context, req dto.ListCategoryPerformanceRequest) ([]dto.CategoryPerformanceResponse, utils.PaginationResult, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	params := repositories.ListCategoryPerformanceParams{
		Search:    req.Search,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	}

	rows, pagination, err := uc.repo.ListCategoryPerformance(ctx, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.CategoryPerformanceResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.CategoryPerformanceResponse{
			CategoryID:            row.CategoryID,
			CategoryName:          row.CategoryName,
			ProductCount:          row.ProductCount,
			TotalQty:              row.TotalQty,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatProductCurrencyIDR(row.TotalRevenue),
			TotalOrders:           row.TotalOrders,
			AvgPrice:              row.AvgPrice,
			AvgPriceFormatted:     formatProductCurrencyIDR(row.AvgPrice),
		})
	}

	return results, pagination, nil
}

func (uc *productAnalysisUsecase) ListSegmentPerformance(ctx context.Context, req dto.ListSegmentPerformanceRequest) ([]dto.SegmentPerformanceResponse, utils.PaginationResult, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	params := repositories.ListDimensionPerformanceParams{
		Search:    req.Search,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	}

	rows, pagination, err := uc.repo.ListSegmentPerformance(ctx, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.SegmentPerformanceResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.SegmentPerformanceResponse{
			SegmentID:             row.DimensionID,
			SegmentName:           row.DimensionName,
			ProductCount:          row.ProductCount,
			TotalQty:              row.TotalQty,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatProductCurrencyIDR(row.TotalRevenue),
			TotalOrders:           row.TotalOrders,
			AvgPrice:              row.AvgPrice,
			AvgPriceFormatted:     formatProductCurrencyIDR(row.AvgPrice),
		})
	}

	return results, pagination, nil
}

func (uc *productAnalysisUsecase) ListTypePerformance(ctx context.Context, req dto.ListTypePerformanceRequest) ([]dto.TypePerformanceResponse, utils.PaginationResult, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	params := repositories.ListDimensionPerformanceParams{
		Search:    req.Search,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	}

	rows, pagination, err := uc.repo.ListTypePerformance(ctx, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.TypePerformanceResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.TypePerformanceResponse{
			TypeID:                row.DimensionID,
			TypeName:              row.DimensionName,
			ProductCount:          row.ProductCount,
			TotalQty:              row.TotalQty,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatProductCurrencyIDR(row.TotalRevenue),
			TotalOrders:           row.TotalOrders,
			AvgPrice:              row.AvgPrice,
			AvgPriceFormatted:     formatProductCurrencyIDR(row.AvgPrice),
		})
	}

	return results, pagination, nil
}

func (uc *productAnalysisUsecase) ListPackagingPerformance(ctx context.Context, req dto.ListPackagingPerformanceRequest) ([]dto.PackagingPerformanceResponse, utils.PaginationResult, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	params := repositories.ListDimensionPerformanceParams{
		Search:    req.Search,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	}

	rows, pagination, err := uc.repo.ListPackagingPerformance(ctx, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.PackagingPerformanceResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.PackagingPerformanceResponse{
			PackagingID:           row.DimensionID,
			PackagingName:         row.DimensionName,
			ProductCount:          row.ProductCount,
			TotalQty:              row.TotalQty,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatProductCurrencyIDR(row.TotalRevenue),
			TotalOrders:           row.TotalOrders,
			AvgPrice:              row.AvgPrice,
			AvgPriceFormatted:     formatProductCurrencyIDR(row.AvgPrice),
		})
	}

	return results, pagination, nil
}

func (uc *productAnalysisUsecase) ListProcurementTypePerformance(ctx context.Context, req dto.ListProcurementTypePerformanceRequest) ([]dto.ProcurementTypePerformanceResponse, utils.PaginationResult, error) {
	startDate, endDate := parseProductDateRange(req.StartDate, req.EndDate)

	params := repositories.ListDimensionPerformanceParams{
		Search:    req.Search,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	}

	rows, pagination, err := uc.repo.ListProcurementTypePerformance(ctx, params)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.ProcurementTypePerformanceResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.ProcurementTypePerformanceResponse{
			ProcurementTypeID:     row.DimensionID,
			ProcurementTypeName:   row.DimensionName,
			ProductCount:          row.ProductCount,
			TotalQty:              row.TotalQty,
			TotalRevenue:          row.TotalRevenue,
			TotalRevenueFormatted: formatProductCurrencyIDR(row.TotalRevenue),
			TotalOrders:           row.TotalOrders,
			AvgPrice:              row.AvgPrice,
			AvgPriceFormatted:     formatProductCurrencyIDR(row.AvgPrice),
		})
	}

	return results, pagination, nil
}

// --- Helpers (scoped to product analysis) ---

func parseProductDateRange(startStr, endStr string) (time.Time, time.Time) {
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

func formatProductCurrencyIDR(amount float64) string {
	rounded := math.Round(amount)
	if rounded >= 1_000_000_000 {
		return fmt.Sprintf("Rp %.1fM", rounded/1_000_000_000)
	}
	if rounded >= 1_000_000 {
		return fmt.Sprintf("Rp %.1fjt", rounded/1_000_000)
	}
	return fmt.Sprintf("Rp %s", formatProductNumber(rounded))
}

func formatProductNumber(n float64) string {
	s := fmt.Sprintf("%.0f", n)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return result
}

func productMonthName(month int) string {
	names := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	if month >= 1 && month <= 12 {
		return names[month]
	}
	return ""
}
