package usecase

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/utils"
	"github.com/gilabs/gims/api/internal/report/data/repositories"
	"github.com/gilabs/gims/api/internal/report/domain/dto"
)

// CustomerResearchUsecase defines business logic for customer research reports.
type CustomerResearchUsecase interface {
	GetKPIs(ctx context.Context, req dto.GetCustomerResearchKpisRequest) (*dto.CustomerResearchKpisResponse, error)
	GetRevenueTrend(ctx context.Context, req dto.GetRevenueTrendRequest) (*dto.RevenueTrendResponse, error)
	ListCustomers(ctx context.Context, req dto.ListCustomersRequest) (*dto.ListCustomersResponse, utils.PaginationResult, error)
	ListRevenueByCustomer(ctx context.Context, req dto.ListRevenueByCustomerRequest) (*dto.ListCustomersResponse, utils.PaginationResult, error)
	ListPurchaseFrequency(ctx context.Context, req dto.ListPurchaseFrequencyRequest) (*dto.ListCustomersResponse, utils.PaginationResult, error)
	GetCustomerDetail(ctx context.Context, customerID string, req dto.GetCustomerResearchKpisRequest) (*dto.CustomerDetailResponse, error)
	GetCustomerTopProducts(ctx context.Context, customerID string, req dto.GetCustomerTopProductsRequest) (*dto.CustomerTopProductsResponse, error)
}

type customerResearchUsecase struct {
	repo repositories.CustomerResearchRepository
}

// NewCustomerResearchUsecase creates a new CustomerResearchUsecase instance.
func NewCustomerResearchUsecase(repo repositories.CustomerResearchRepository) CustomerResearchUsecase {
	return &customerResearchUsecase{repo: repo}
}

func defaultCustomerResearchDateRange() (time.Time, time.Time) {
	now := apptime.Now()
	return now.AddDate(0, 0, -30), now
}

func parseCustomerResearchDateRange(startStr, endStr, dateMode string, year int) (time.Time, time.Time) {
	start, end := defaultCustomerResearchDateRange()
	now := apptime.Now()

	if strings.EqualFold(strings.TrimSpace(dateMode), "year") && year >= 2000 && year <= now.Year()+1 {
		return time.Date(year, 1, 1, 0, 0, 0, 0, now.Location()), time.Date(year, 12, 31, 23, 59, 59, 0, now.Location())
	}

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

func (uc *customerResearchUsecase) GetKPIs(ctx context.Context, req dto.GetCustomerResearchKpisRequest) (*dto.CustomerResearchKpisResponse, error) {
	start, end := parseCustomerResearchDateRange(req.StartDate, req.EndDate, req.DateMode, req.Year)
	row, err := uc.repo.GetKPIs(ctx, start, end)
	if err != nil {
		return nil, err
	}

	inactive := row.TotalCustomers - row.ActiveCustomers
	if inactive < 0 {
		inactive = 0
	}

	avgOrderValue := 0.0
	if row.TotalOrders > 0 {
		avgOrderValue = row.TotalRevenue / float64(row.TotalOrders)
	}

	return &dto.CustomerResearchKpisResponse{
		TotalCustomers:    row.TotalCustomers,
		ActiveCustomers:   row.ActiveCustomers,
		InactiveCustomers: inactive,
		TotalRevenue:      row.TotalRevenue,
		AverageOrderValue: math.Round(avgOrderValue*100) / 100,
	}, nil
}

func (uc *customerResearchUsecase) GetRevenueTrend(ctx context.Context, req dto.GetRevenueTrendRequest) (*dto.RevenueTrendResponse, error) {
	start, end := parseCustomerResearchDateRange(req.StartDate, req.EndDate, req.DateMode, req.Year)
	interval := strings.ToLower(strings.TrimSpace(req.Interval))
	if interval == "" {
		interval = "daily"
	}

	rows, err := uc.repo.GetRevenueTrend(ctx, start, end, interval)
	if err != nil {
		return nil, err
	}

	data := make([]dto.RevenueTrendData, 0, len(rows))
	for _, row := range rows {
		data = append(data, dto.RevenueTrendData{
			Period:       row.Period,
			TotalRevenue: row.TotalRevenue,
			TotalOrders:  row.TotalOrders,
		})
	}

	return &dto.RevenueTrendResponse{Data: data}, nil
}

func (uc *customerResearchUsecase) ListCustomers(ctx context.Context, req dto.ListCustomersRequest) (*dto.ListCustomersResponse, utils.PaginationResult, error) {
	start, end := parseCustomerResearchDateRange(req.StartDate, req.EndDate, req.DateMode, req.Year)

	rows, pagination, err := uc.repo.ListCustomers(ctx, repositories.ListCustomersParams{
		StartDate: start,
		EndDate:   end,
		Tab:       req.Tab,
		Search:    req.Search,
		Page:      req.Page,
		PerPage:   req.PerPage,
		SortBy:    req.SortBy,
		Order:     req.Order,
	})
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	items := mapCustomerRows(rows)
	return &dto.ListCustomersResponse{Data: items}, pagination, nil
}

func (uc *customerResearchUsecase) ListRevenueByCustomer(ctx context.Context, req dto.ListRevenueByCustomerRequest) (*dto.ListCustomersResponse, utils.PaginationResult, error) {
	start, end := parseCustomerResearchDateRange(req.StartDate, req.EndDate, req.DateMode, req.Year)
	rows, pagination, err := uc.repo.GetRevenueByCustomer(ctx, repositories.ListCustomersParams{
		StartDate: start,
		EndDate:   end,
		Search:    req.Search,
		Page:      req.Page,
		PerPage:   req.PerPage,
		Order:     req.Order,
	})
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	items := mapCustomerRows(rows)
	return &dto.ListCustomersResponse{Data: items}, pagination, nil
}

func (uc *customerResearchUsecase) ListPurchaseFrequency(ctx context.Context, req dto.ListPurchaseFrequencyRequest) (*dto.ListCustomersResponse, utils.PaginationResult, error) {
	start, end := parseCustomerResearchDateRange(req.StartDate, req.EndDate, req.DateMode, req.Year)
	rows, pagination, err := uc.repo.GetPurchaseFrequency(ctx, repositories.ListCustomersParams{
		StartDate: start,
		EndDate:   end,
		Search:    req.Search,
		Page:      req.Page,
		PerPage:   req.PerPage,
		Order:     req.Order,
	})
	if err != nil {
		return nil, utils.PaginationResult{}, err
	}

	items := mapCustomerRows(rows)
	return &dto.ListCustomersResponse{Data: items}, pagination, nil
}

func (uc *customerResearchUsecase) GetCustomerDetail(ctx context.Context, customerID string, req dto.GetCustomerResearchKpisRequest) (*dto.CustomerDetailResponse, error) {
	start, end := parseCustomerResearchDateRange(req.StartDate, req.EndDate, req.DateMode, req.Year)
	row, err := uc.repo.GetCustomerDetail(ctx, customerID, start, end)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}

	return &dto.CustomerDetailResponse{
		CustomerID:        row.CustomerID,
		CustomerName:      row.CustomerName,
		TotalRevenue:      row.TotalRevenue,
		TotalOrders:       row.TotalOrders,
		AverageOrderValue: row.AverageOrderValue,
		LastOrderDate:     row.LastOrderDate,
	}, nil
}

func (uc *customerResearchUsecase) GetCustomerTopProducts(ctx context.Context, customerID string, req dto.GetCustomerTopProductsRequest) (*dto.CustomerTopProductsResponse, error) {
	start, end := parseCustomerResearchDateRange(req.StartDate, req.EndDate, req.DateMode, req.Year)
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	rows, err := uc.repo.GetCustomerTopProducts(ctx, customerID, start, end, limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.CustomerProductItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.CustomerProductItem{
			ProductID:    row.ProductID,
			ProductCode:  row.ProductCode,
			ProductName:  row.ProductName,
			TotalQty:     row.TotalQty,
			TotalRevenue: row.TotalRevenue,
			TotalOrders:  row.TotalOrders,
		})
	}

	return &dto.CustomerTopProductsResponse{Data: items}, nil
}

func mapCustomerRows(rows []repositories.CustomerResearchRow) []dto.CustomerRow {
	items := make([]dto.CustomerRow, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.CustomerRow{
			CustomerID:            row.CustomerID,
			CustomerName:          row.CustomerName,
			TotalRevenue:          row.TotalRevenue,
			TotalOrders:           row.TotalOrders,
			AverageOrderValue:     row.AverageOrderValue,
			LastOrderDate:         row.LastOrderDate,
			ActiveSalesOrderCount: row.ActiveSalesOrderCount,
		})
	}
	return items
}
