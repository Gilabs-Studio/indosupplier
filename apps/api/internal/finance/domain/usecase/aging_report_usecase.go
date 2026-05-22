package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/financesettings"
)

type AgingReportUsecase interface {
	ARAging(ctx context.Context, asOf time.Time, search string, page, perPage int) (*dto.ARAgingReportResponse, int64, error)
	APAging(ctx context.Context, asOf time.Time, search string, page, perPage int) (*dto.APAgingReportResponse, int64, error)
	ARAgingFinance(ctx context.Context, query dto.AgingFinanceQuery) (*dto.ARAgingReportResponse, error)
	APAgingFinance(ctx context.Context, query dto.AgingFinanceQuery) (*dto.APAgingReportResponse, error)
}

type agingReportUsecase struct {
	repo            repositories.AgingReportRepository
	settingsService financesettings.SettingsService
}

func NewAgingReportUsecase(repo repositories.AgingReportRepository, settingsService ...financesettings.SettingsService) AgingReportUsecase {
	var svc financesettings.SettingsService
	if len(settingsService) > 0 {
		svc = settingsService[0]
	}

	return &agingReportUsecase{
		repo:            repo,
		settingsService: svc,
	}
}

type agingBucketRule struct {
	Key     string
	Label   string
	MinDays *int
	MaxDays *int
}

func intPtr(value int) *int {
	v := value
	return &v
}

func defaultAgingBucketRules() []agingBucketRule {
	return []agingBucketRule{
		{Key: "current", Label: "Current", MinDays: nil, MaxDays: intPtr(0)},
		{Key: "days_1_30", Label: "1-30 Days", MinDays: intPtr(1), MaxDays: intPtr(30)},
		{Key: "days_31_60", Label: "31-60 Days", MinDays: intPtr(31), MaxDays: intPtr(60)},
		{Key: "days_61_90", Label: "61-90 Days", MinDays: intPtr(61), MaxDays: intPtr(90)},
		{Key: "over_90", Label: ">90 Days", MinDays: intPtr(91), MaxDays: nil},
	}
}

func normalizeBucketKey(raw string) string {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return normalized
}

func humanizeBucketKey(key string) string {
	parts := strings.Split(strings.ReplaceAll(key, "-", "_"), "_")
	humanized := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		humanized = append(humanized, strings.ToUpper(part[:1])+part[1:])
	}
	if len(humanized) == 0 {
		return "Bucket"
	}
	return strings.Join(humanized, " ")
}

func parseAgingBucketRules(raw string) ([]agingBucketRule, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("aging bucket setting is empty")
	}

	var definitions []dto.AgingBucketDefinition
	if err := json.Unmarshal([]byte(raw), &definitions); err != nil {
		return nil, err
	}

	rules := make([]agingBucketRule, 0, len(definitions))
	for _, definition := range definitions {
		key := normalizeBucketKey(definition.Key)
		if key == "" {
			continue
		}
		if definition.MinDays != nil && definition.MaxDays != nil && *definition.MinDays > *definition.MaxDays {
			continue
		}

		label := strings.TrimSpace(definition.Label)
		if label == "" {
			label = humanizeBucketKey(key)
		}

		rules = append(rules, agingBucketRule{
			Key:     key,
			Label:   label,
			MinDays: definition.MinDays,
			MaxDays: definition.MaxDays,
		})
	}

	if len(rules) == 0 {
		return nil, errors.New("aging bucket setting has no valid bucket")
	}

	return rules, nil
}

func (uc *agingReportUsecase) loadBucketRules(ctx context.Context, reportType string) []agingBucketRule {
	defaults := defaultAgingBucketRules()
	if uc.settingsService == nil {
		return defaults
	}

	keys := []string{financeModels.AgingBucketConfigSettingKey(reportType)}
	if keys[0] != financeModels.SettingAgingBucketConfig {
		keys = append(keys, financeModels.SettingAgingBucketConfig)
	}

	for _, key := range keys {
		rawValue, err := uc.settingsService.GetValue(ctx, key)
		if err != nil {
			continue
		}

		parsed, err := parseAgingBucketRules(rawValue)
		if err != nil {
			continue
		}
		return parsed
	}

	return defaults
}

func toBucketDefinitions(rules []agingBucketRule) []dto.AgingBucketDefinition {
	defs := make([]dto.AgingBucketDefinition, 0, len(rules))
	for _, rule := range rules {
		defs = append(defs, dto.AgingBucketDefinition{
			Key:     rule.Key,
			Label:   rule.Label,
			MinDays: rule.MinDays,
			MaxDays: rule.MaxDays,
		})
	}
	return defs
}

func (r agingBucketRule) matches(daysPastDue int) bool {
	if r.MinDays != nil && daysPastDue < *r.MinDays {
		return false
	}
	if r.MaxDays != nil && daysPastDue > *r.MaxDays {
		return false
	}
	return true
}

func daysPastDue(asOf time.Time, due time.Time) int {
	asOfDate := time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, time.UTC)
	d := int(asOfDate.Sub(dueDate).Hours() / 24)
	return d
}

func bucketize(amount float64, dpd int, rules []agingBucketRule) dto.AgingBuckets {
	buckets := dto.AgingBuckets{Dynamic: map[string]float64{}}
	if amount <= 0 {
		return buckets
	}

	matched := false
	for _, rule := range rules {
		if !rule.matches(dpd) {
			continue
		}
		buckets.Dynamic[rule.Key] += amount
		matched = true
		break
	}
	if !matched && len(rules) > 0 {
		buckets.Dynamic[rules[len(rules)-1].Key] += amount
	}

	if dpd <= 0 {
		buckets.Current = amount
		return buckets
	}
	if dpd <= 30 {
		buckets.Days1To30 = amount
		return buckets
	}
	if dpd <= 60 {
		buckets.Days31To60 = amount
		return buckets
	}
	if dpd <= 90 {
		buckets.Days61To90 = amount
		return buckets
	}
	buckets.Over90 = amount
	return buckets
}

func addBuckets(a dto.AgingBuckets, b dto.AgingBuckets) dto.AgingBuckets {
	result := dto.AgingBuckets{
		Current:    a.Current + b.Current,
		Days1To30:  a.Days1To30 + b.Days1To30,
		Days31To60: a.Days31To60 + b.Days31To60,
		Days61To90: a.Days61To90 + b.Days61To90,
		Over90:     a.Over90 + b.Over90,
	}
	if len(a.Dynamic) > 0 || len(b.Dynamic) > 0 {
		result.Dynamic = make(map[string]float64, len(a.Dynamic)+len(b.Dynamic))
		for key, value := range a.Dynamic {
			result.Dynamic[key] += value
		}
		for key, value := range b.Dynamic {
			result.Dynamic[key] += value
		}
	}
	return result
}

func bucketsOverdueTotal(b dto.AgingBuckets) float64 {
	return b.Days1To30 + b.Days31To60 + b.Days61To90 + b.Over90
}

type arGroupAccumulator struct {
	customerID       string
	customerName     string
	invoices         []dto.ARAgingInvoiceRow
	buckets          dto.AgingBuckets
	totalOutstanding float64
}

type apGroupAccumulator struct {
	supplierID       string
	supplierName     string
	invoices         []dto.APAgingInvoiceRow
	buckets          dto.AgingBuckets
	totalOutstanding float64
}

func (uc *agingReportUsecase) ARAging(ctx context.Context, asOf time.Time, search string, page, perPage int) (*dto.ARAgingReportResponse, int64, error) {
	// Normalize asOf to end of day
	asOf = time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 23, 59, 59, 999999999, asOf.Location())

	if uc.repo == nil {
		return nil, 0, errors.New("repository is not configured")
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	rows, total, err := uc.repo.ListARAging(ctx, repositories.AgingListParams{
		Search:   strings.TrimSpace(search),
		AsOfDate: asOf,
		Limit:    perPage,
		Offset:   (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}
	bucketRules := uc.loadBucketRules(ctx, financeModels.AgingReportTypeAR)
	bucketConfig := toBucketDefinitions(bucketRules)

	respRows := make([]dto.ARAgingInvoiceRow, 0, len(rows))
	totals := dto.AgingTotals{}
	for _, r := range rows {
		if r.RemainingAmount <= 0 || math.IsNaN(r.RemainingAmount) {
			continue
		}
		dpd := daysPastDue(asOf, r.DueDate)
		b := bucketize(r.RemainingAmount, dpd, bucketRules)
		respRows = append(respRows, dto.ARAgingInvoiceRow{
			InvoiceID:       r.InvoiceID,
			SourceType:      r.SourceType,
			Code:            r.Code,
			InvoiceNumber:   r.InvoiceNumber,
			CustomerID:      r.CustomerID,
			CustomerName:    r.CustomerName,
			InvoiceDate:     r.InvoiceDate,
			DueDate:         r.DueDate,
			DaysPastDue:     dpd,
			Amount:          r.Amount,
			RemainingAmount: r.RemainingAmount,
			Buckets:         b,
		})
		totals.Count++
		totals.Remaining += r.RemainingAmount
		totals.Buckets = addBuckets(totals.Buckets, b)
	}

	return &dto.ARAgingReportResponse{AsOfDate: asOf, BucketConfig: bucketConfig, Rows: respRows, Totals: totals}, total, nil
}

func (uc *agingReportUsecase) APAging(ctx context.Context, asOf time.Time, search string, page, perPage int) (*dto.APAgingReportResponse, int64, error) {
	// Normalize asOf to end of day
	asOf = time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 23, 59, 59, 999999999, asOf.Location())

	if uc.repo == nil {
		return nil, 0, errors.New("repository is not configured")
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	rows, total, err := uc.repo.ListAPAging(ctx, repositories.AgingListParams{
		Search:   strings.TrimSpace(search),
		AsOfDate: asOf,
		Limit:    perPage,
		Offset:   (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}
	bucketRules := uc.loadBucketRules(ctx, financeModels.AgingReportTypeAP)
	bucketConfig := toBucketDefinitions(bucketRules)

	respRows := make([]dto.APAgingInvoiceRow, 0, len(rows))
	totals := dto.AgingTotals{}
	for _, r := range rows {
		if r.RemainingAmount <= 0 || math.IsNaN(r.RemainingAmount) {
			continue
		}
		dpd := daysPastDue(asOf, r.DueDate)
		b := bucketize(r.RemainingAmount, dpd, bucketRules)
		respRows = append(respRows, dto.APAgingInvoiceRow{
			InvoiceID:       r.InvoiceID,
			SourceType:      r.SourceType,
			Code:            r.Code,
			InvoiceNumber:   r.InvoiceNumber,
			InvoiceDate:     r.InvoiceDate,
			DueDate:         r.DueDate,
			DaysPastDue:     dpd,
			SupplierID:      r.SupplierID,
			SupplierName:    r.SupplierName,
			Amount:          r.Amount,
			PaidAmount:      r.PaidAmount,
			RemainingAmount: r.RemainingAmount,
			Buckets:         b,
		})
		totals.Count++
		totals.Remaining += r.RemainingAmount
		totals.Buckets = addBuckets(totals.Buckets, b)
	}

	return &dto.APAgingReportResponse{AsOfDate: asOf, BucketConfig: bucketConfig, Rows: respRows, Totals: totals}, total, nil
}

func (uc *agingReportUsecase) ARAgingFinance(ctx context.Context, query dto.AgingFinanceQuery) (*dto.ARAgingReportResponse, error) {
	asOf := time.Date(query.AsOfDate.Year(), query.AsOfDate.Month(), query.AsOfDate.Day(), 23, 59, 59, 999999999, query.AsOfDate.Location())
	rows, err := uc.repo.ListARAgingFinance(ctx, repositories.AgingFinanceListParams{
		Search:    strings.TrimSpace(query.Search),
		AsOfDate:  asOf,
		PartnerID: strings.TrimSpace(query.PartnerID),
		MinAmount: query.MinAmount,
	})
	if err != nil {
		return nil, err
	}
	bucketRules := uc.loadBucketRules(ctx, financeModels.AgingReportTypeAR)
	bucketConfig := toBucketDefinitions(bucketRules)

	respRows := make([]dto.ARAgingInvoiceRow, 0, len(rows))
	totals := dto.AgingTotals{}
	summary := &dto.AgingSummary{}
	groups := make(map[string]*arGroupAccumulator)

	for _, r := range rows {
		if r.RemainingAmount <= 0 || math.IsNaN(r.RemainingAmount) {
			continue
		}
		dpd := daysPastDue(asOf, r.DueDate)
		if !query.IncludeCurrent && dpd <= 0 {
			continue
		}
		buckets := bucketize(r.RemainingAmount, dpd, bucketRules)
		sourceType := strings.TrimSpace(r.SourceType)
		if sourceType == "" {
			sourceType = "CUSTOMER_INVOICE"
		}

		row := dto.ARAgingInvoiceRow{
			InvoiceID:       r.InvoiceID,
			SourceType:      sourceType,
			Code:            r.Code,
			InvoiceNumber:   r.InvoiceNumber,
			CustomerID:      r.CustomerID,
			CustomerName:    r.CustomerName,
			InvoiceDate:     r.InvoiceDate,
			DueDate:         r.DueDate,
			DaysPastDue:     dpd,
			Amount:          r.Amount,
			RemainingAmount: r.RemainingAmount,
			Buckets:         buckets,
		}

		respRows = append(respRows, row)
		totals.Count++
		totals.Remaining += r.RemainingAmount
		totals.Buckets = addBuckets(totals.Buckets, buckets)

		groupKey := strings.TrimSpace(r.CustomerID)
		if groupKey == "" {
			groupKey = "__unknown__:" + strings.TrimSpace(r.CustomerName)
		}
		group := groups[groupKey]
		if group == nil {
			group = &arGroupAccumulator{
				customerID:   strings.TrimSpace(r.CustomerID),
				customerName: strings.TrimSpace(r.CustomerName),
			}
			if group.customerName == "" {
				group.customerName = "Customer"
			}
			groups[groupKey] = group
		}
		group.invoices = append(group.invoices, row)
		group.totalOutstanding += r.RemainingAmount
		group.buckets = addBuckets(group.buckets, buckets)
	}

	customers := make([]dto.ARAgingPartnerGroup, 0, len(groups))
	for _, group := range groups {
		customers = append(customers, dto.ARAgingPartnerGroup{
			CustomerID:       group.customerID,
			CustomerName:     group.customerName,
			InvoiceCount:     len(group.invoices),
			TotalOutstanding: group.totalOutstanding,
			Buckets:          group.buckets,
			Invoices:         group.invoices,
		})
	}
	sort.SliceStable(customers, func(i, j int) bool {
		if customers[i].TotalOutstanding == customers[j].TotalOutstanding {
			return customers[i].CustomerName < customers[j].CustomerName
		}
		return customers[i].TotalOutstanding > customers[j].TotalOutstanding
	})

	summary.InvoiceCount = totals.Count
	summary.PartnerCount = len(customers)
	summary.TotalOutstanding = totals.Remaining
	summary.Buckets = totals.Buckets
	summary.TotalCurrent = totals.Buckets.Current
	summary.TotalOverdue = bucketsOverdueTotal(totals.Buckets)

	return &dto.ARAgingReportResponse{
		AsOfDate:     asOf,
		BucketConfig: bucketConfig,
		Rows:         respRows,
		Totals:       totals,
		Summary:      summary,
		Customers:    customers,
	}, nil
}

func (uc *agingReportUsecase) APAgingFinance(ctx context.Context, query dto.AgingFinanceQuery) (*dto.APAgingReportResponse, error) {
	asOf := time.Date(query.AsOfDate.Year(), query.AsOfDate.Month(), query.AsOfDate.Day(), 23, 59, 59, 999999999, query.AsOfDate.Location())
	rows, err := uc.repo.ListAPAgingFinance(ctx, repositories.AgingFinanceListParams{
		Search:    strings.TrimSpace(query.Search),
		AsOfDate:  asOf,
		PartnerID: strings.TrimSpace(query.PartnerID),
		MinAmount: query.MinAmount,
	})
	if err != nil {
		return nil, err
	}
	bucketRules := uc.loadBucketRules(ctx, financeModels.AgingReportTypeAP)
	bucketConfig := toBucketDefinitions(bucketRules)

	respRows := make([]dto.APAgingInvoiceRow, 0, len(rows))
	totals := dto.AgingTotals{}
	summary := &dto.AgingSummary{}
	groups := make(map[string]*apGroupAccumulator)

	for _, r := range rows {
		if r.RemainingAmount <= 0 || math.IsNaN(r.RemainingAmount) {
			continue
		}
		dpd := daysPastDue(asOf, r.DueDate)
		if !query.IncludeCurrent && dpd <= 0 {
			continue
		}
		buckets := bucketize(r.RemainingAmount, dpd, bucketRules)
		sourceType := strings.TrimSpace(r.SourceType)
		if sourceType == "" {
			sourceType = "SUPPLIER_INVOICE"
		}

		row := dto.APAgingInvoiceRow{
			InvoiceID:       r.InvoiceID,
			SourceType:      sourceType,
			Code:            r.Code,
			InvoiceNumber:   r.InvoiceNumber,
			InvoiceDate:     r.InvoiceDate,
			DueDate:         r.DueDate,
			DaysPastDue:     dpd,
			SupplierID:      r.SupplierID,
			SupplierName:    r.SupplierName,
			Amount:          r.Amount,
			PaidAmount:      r.PaidAmount,
			RemainingAmount: r.RemainingAmount,
			Buckets:         buckets,
		}

		respRows = append(respRows, row)
		totals.Count++
		totals.Remaining += r.RemainingAmount
		totals.Buckets = addBuckets(totals.Buckets, buckets)

		groupKey := strings.TrimSpace(r.SupplierID)
		if groupKey == "" {
			groupKey = "__unknown__:" + strings.TrimSpace(r.SupplierName)
		}
		group := groups[groupKey]
		if group == nil {
			group = &apGroupAccumulator{
				supplierID:   strings.TrimSpace(r.SupplierID),
				supplierName: strings.TrimSpace(r.SupplierName),
			}
			if group.supplierName == "" {
				group.supplierName = "Supplier"
			}
			groups[groupKey] = group
		}
		group.invoices = append(group.invoices, row)
		group.totalOutstanding += r.RemainingAmount
		group.buckets = addBuckets(group.buckets, buckets)
	}

	suppliers := make([]dto.APAgingPartnerGroup, 0, len(groups))
	for _, group := range groups {
		suppliers = append(suppliers, dto.APAgingPartnerGroup{
			SupplierID:       group.supplierID,
			SupplierName:     group.supplierName,
			InvoiceCount:     len(group.invoices),
			TotalOutstanding: group.totalOutstanding,
			Buckets:          group.buckets,
			Invoices:         group.invoices,
		})
	}
	sort.SliceStable(suppliers, func(i, j int) bool {
		if suppliers[i].TotalOutstanding == suppliers[j].TotalOutstanding {
			return suppliers[i].SupplierName < suppliers[j].SupplierName
		}
		return suppliers[i].TotalOutstanding > suppliers[j].TotalOutstanding
	})

	summary.InvoiceCount = totals.Count
	summary.PartnerCount = len(suppliers)
	summary.TotalOutstanding = totals.Remaining
	summary.Buckets = totals.Buckets
	summary.TotalCurrent = totals.Buckets.Current
	summary.TotalOverdue = bucketsOverdueTotal(totals.Buckets)

	return &dto.APAgingReportResponse{
		AsOfDate:     asOf,
		BucketConfig: bucketConfig,
		Rows:         respRows,
		Totals:       totals,
		Summary:      summary,
		Suppliers:    suppliers,
	}, nil
}
