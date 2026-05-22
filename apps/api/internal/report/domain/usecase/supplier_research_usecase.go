package usecase

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/report/data/repositories"
	"github.com/gilabs/gims/api/internal/report/domain/dto"
)

// SupplierResearchUsecase defines business logic for supplier research reports.
type SupplierResearchUsecase interface {
	GetKpis(ctx context.Context, req dto.SupplierResearchKpisRequest) (*dto.SupplierResearchKpisResponse, error)
	ListPurchaseVolume(ctx context.Context, req dto.ListSupplierPurchaseVolumeRequest) ([]dto.SupplierPurchaseVolumeResponse, utils.PaginationResult, error)
	ListDeliveryTime(ctx context.Context, req dto.ListSupplierDeliveryTimeRequest) ([]dto.SupplierDeliveryTimeResponse, utils.PaginationResult, error)
	GetSpendTrend(ctx context.Context, req dto.SupplierSpendTrendRequest) (*dto.SupplierSpendTrendResponse, error)
	ListSuppliers(ctx context.Context, req dto.ListSuppliersRequest) ([]dto.SupplierTableRowResponse, utils.PaginationResult, error)
	GetSupplierDetail(ctx context.Context, supplierID string, req dto.SupplierResearchKpisRequest) (*dto.SupplierDetailResponse, error)
}

type supplierResearchUsecase struct {
	repo repositories.SupplierResearchRepository
}

func NewSupplierResearchUsecase(repo repositories.SupplierResearchRepository) SupplierResearchUsecase {
	return &supplierResearchUsecase{repo: repo}
}

func (uc *supplierResearchUsecase) GetKpis(ctx context.Context, req dto.SupplierResearchKpisRequest) (*dto.SupplierResearchKpisResponse, error) {
	params := parseSupplierFilterParams(req.StartDate, req.EndDate, req.DateMode, req.Year, req.CategoryIDs, "", req.MinPurchaseValue, req.MaxPurchaseValue)

	row, err := uc.repo.GetKpis(ctx, params)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return &dto.SupplierResearchKpisResponse{}, nil
	}

	return &dto.SupplierResearchKpisResponse{
		TotalSuppliers:      row.TotalSuppliers,
		ActiveSuppliers:     row.ActiveSuppliers,
		TotalPurchaseValue:  row.TotalPurchaseValue,
		AverageLeadTimeDays: round2(row.AverageLeadTimeDays),
	}, nil
}

func (uc *supplierResearchUsecase) ListPurchaseVolume(ctx context.Context, req dto.ListSupplierPurchaseVolumeRequest) ([]dto.SupplierPurchaseVolumeResponse, utils.PaginationResult, error) {
	listParams := repositories.SupplierResearchListParams{
		SupplierResearchFilterParams: parseSupplierFilterParams(req.StartDate, req.EndDate, req.DateMode, req.Year, req.CategoryIDs, req.Search, req.MinPurchaseValue, req.MaxPurchaseValue),
		Page:                         req.Page,
		PerPage:                      req.PerPage,
		SortBy:                       req.SortBy,
		Order:                        req.Order,
	}

	rows, pagination, err := uc.repo.ListPurchaseVolume(ctx, listParams)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.SupplierPurchaseVolumeResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.SupplierPurchaseVolumeResponse{
			SupplierID:          row.SupplierID,
			SupplierCode:        row.SupplierCode,
			SupplierName:        row.SupplierName,
			CategoryName:        row.CategoryName,
			TotalPurchaseValue:  row.TotalPurchaseValue,
			TotalPurchaseOrders: row.TotalPurchaseOrders,
			DependencyScore:     round2(row.DependencyScore),
		})
	}

	return results, pagination, nil
}

func (uc *supplierResearchUsecase) ListDeliveryTime(ctx context.Context, req dto.ListSupplierDeliveryTimeRequest) ([]dto.SupplierDeliveryTimeResponse, utils.PaginationResult, error) {
	listParams := repositories.SupplierResearchListParams{
		SupplierResearchFilterParams: parseSupplierFilterParams(req.StartDate, req.EndDate, req.DateMode, req.Year, req.CategoryIDs, req.Search, req.MinPurchaseValue, req.MaxPurchaseValue),
		Page:                         req.Page,
		PerPage:                      req.PerPage,
		SortBy:                       req.SortBy,
		Order:                        req.Order,
	}

	rows, pagination, err := uc.repo.ListDeliveryTime(ctx, listParams)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.SupplierDeliveryTimeResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.SupplierDeliveryTimeResponse{
			SupplierID:          row.SupplierID,
			SupplierName:        row.SupplierName,
			AverageLeadTimeDays: round2(row.AverageLeadTimeDays),
			SupplierOnTimeRate:  round2(row.SupplierOnTimeRate),
			LateDeliveryCount:   row.LateDeliveryCount,
		})
	}

	return results, pagination, nil
}

func (uc *supplierResearchUsecase) GetSpendTrend(ctx context.Context, req dto.SupplierSpendTrendRequest) (*dto.SupplierSpendTrendResponse, error) {
	interval := strings.ToLower(strings.TrimSpace(req.Interval))
	if interval == "" {
		interval = "monthly"
	}
	if interval != "daily" && interval != "weekly" && interval != "monthly" {
		interval = "monthly"
	}

	params := parseSupplierFilterParams(req.StartDate, req.EndDate, req.DateMode, req.Year, req.CategoryIDs, "", req.MinPurchaseValue, req.MaxPurchaseValue)
	rows, err := uc.repo.GetSpendTrend(ctx, params, interval)
	if err != nil {
		return nil, err
	}

	timeline := make([]dto.SupplierSpendTrendPointResponse, 0, len(rows))
	for _, row := range rows {
		timeline = append(timeline, dto.SupplierSpendTrendPointResponse{
			Period:             row.Period,
			TotalPurchaseValue: row.TotalPurchaseValue,
		})
	}

	return &dto.SupplierSpendTrendResponse{
		Interval: interval,
		Timeline: timeline,
	}, nil
}

func (uc *supplierResearchUsecase) ListSuppliers(ctx context.Context, req dto.ListSuppliersRequest) ([]dto.SupplierTableRowResponse, utils.PaginationResult, error) {
	tab := strings.ToLower(strings.TrimSpace(req.Tab))
	if tab == "" {
		tab = "top_spenders"
	}

	listParams := repositories.SupplierResearchListParams{
		SupplierResearchFilterParams: parseSupplierFilterParams(req.StartDate, req.EndDate, req.DateMode, req.Year, req.CategoryIDs, req.Search, req.MinPurchaseValue, req.MaxPurchaseValue),
		Page:                         req.Page,
		PerPage:                      req.PerPage,
		SortBy:                       req.SortBy,
		Order:                        req.Order,
		Tab:                          tab,
	}

	rows, pagination, err := uc.repo.ListSuppliers(ctx, listParams)
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	results := make([]dto.SupplierTableRowResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, dto.SupplierTableRowResponse{
			SupplierID:               row.SupplierID,
			SupplierCode:             row.SupplierCode,
			SupplierName:             row.SupplierName,
			CategoryName:             row.CategoryName,
			TotalPurchaseValue:       row.TotalPurchaseValue,
			TotalPurchaseOrders:      row.TotalPurchaseOrders,
			AverageLeadTimeDays:      round2(row.AverageLeadTimeDays),
			LateDeliveryCount:        row.LateDeliveryCount,
			SupplierOnTimeRate:       round2(row.SupplierOnTimeRate),
			DependencyScore:          round2(row.DependencyScore),
			ActivePurchaseOrderCount: row.ActivePurchaseOrderCount,
		})
	}

	return results, pagination, nil
}

func (uc *supplierResearchUsecase) GetSupplierDetail(ctx context.Context, supplierID string, req dto.SupplierResearchKpisRequest) (*dto.SupplierDetailResponse, error) {
	params := parseSupplierFilterParams(req.StartDate, req.EndDate, req.DateMode, req.Year, req.CategoryIDs, "", req.MinPurchaseValue, req.MaxPurchaseValue)
	row, err := uc.repo.GetSupplierDetail(ctx, supplierID, params)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("supplier not found")
	}

	purchaseOrders := make([]dto.SupplierDetailPurchaseOrderResponse, 0, len(row.PurchaseOrders))
	for _, po := range row.PurchaseOrders {
		purchaseOrders = append(purchaseOrders, dto.SupplierDetailPurchaseOrderResponse{
			PurchaseOrderID: po.PurchaseOrderID,
			Code:            po.Code,
			Status:          po.Status,
			OrderDate:       po.OrderDate,
			TotalAmount:     po.TotalAmount,
		})
	}

	products := make([]dto.SupplierDetailPurchasedProductResponse, 0, len(row.Products))
	for _, item := range row.Products {
		products = append(products, dto.SupplierDetailPurchasedProductResponse{
			ProductID:   item.ProductID,
			ProductCode: item.ProductCode,
			ProductName: item.ProductName,
			TotalQty:    round2(item.TotalQuantity),
			TotalOrders: item.TotalOrders,
			TotalAmount: item.TotalAmount,
		})
	}

	return &dto.SupplierDetailResponse{
		SupplierID:          row.SupplierID,
		SupplierCode:        row.SupplierCode,
		SupplierName:        row.SupplierName,
		CategoryName:        row.CategoryName,
		TotalPurchaseValue:  row.TotalPurchaseValue,
		TotalPurchaseOrders: row.TotalPurchaseOrders,
		AverageLeadTimeDays: round2(row.AverageLeadTimeDays),
		SupplierOnTimeRate:  round2(row.SupplierOnTimeRate),
		LateDeliveryCount:   row.LateDeliveryCount,
		DependencyScore:     round2(row.DependencyScore),
		Products:            products,
		PurchaseOrders:      purchaseOrders,
	}, nil
}

func parseSupplierFilterParams(startDate, endDate, dateMode string, year int, categoryIDsCSV, search string, minPurchaseValue, maxPurchaseValue float64) repositories.SupplierResearchFilterParams {
	parsedStartDate, parsedEndDate := parseSupplierDateRange(startDate, endDate, dateMode, year)

	categoryIDs := []string{}
	for _, raw := range strings.Split(categoryIDsCSV, ",") {
		item := strings.TrimSpace(raw)
		if item != "" {
			categoryIDs = append(categoryIDs, item)
		}
	}

	return repositories.SupplierResearchFilterParams{
		Search:           search,
		StartDate:        parsedStartDate,
		EndDate:          parsedEndDate,
		CategoryIDs:      categoryIDs,
		MinPurchaseValue: minPurchaseValue,
		MaxPurchaseValue: maxPurchaseValue,
	}
}

func parseSupplierDateRange(startDate, endDate, dateMode string, year int) (time.Time, time.Time) {
	now := apptime.Now()
	defaultStart := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location())
	defaultEnd := time.Date(now.Year()-1, 12, 31, 23, 59, 59, 0, now.Location())

	if strings.EqualFold(strings.TrimSpace(dateMode), "year") && year >= 2000 && year <= now.Year()+1 {
		return time.Date(year, 1, 1, 0, 0, 0, 0, now.Location()), time.Date(year, 12, 31, 23, 59, 59, 0, now.Location())
	}

	start := defaultStart
	end := defaultEnd

	if startDate != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", startDate, now.Location()); err == nil {
			start = parsed
		}
	}
	if endDate != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", endDate, now.Location()); err == nil {
			end = parsed
		}
	}

	if end.Before(start) {
		start, end = end, start
	}

	return start, end
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
