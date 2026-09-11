package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"
	infraRedis "github.com/gilabs/indosupplier/api/internal/core/infrastructure/redis"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	monetizationModels "github.com/gilabs/indosupplier/api/internal/monetization/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
)

var (
	ErrProfileNotFound = errors.New("supplier profile not found")
	ErrPlanNotFound    = errors.New("subscription plan not found")
)

type PortalUsecase interface {
	GetProfile(ctx context.Context, userID string) (*dto.SupplierProfileDTO, error)
	UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) (*dto.SupplierProfileDTO, error)
	GetBillingOverview(ctx context.Context, userID string) (*dto.BillingOverviewResponse, error)
	UpgradePlan(ctx context.Context, userID string, req *dto.UpgradePlanRequest) (*dto.BillingOverviewResponse, error)
	GetDashboard(ctx context.Context, userID string) (*dto.SupplierDashboardResponse, error)
}

type portalUsecase struct {
	portalRepo repositories.PortalRepository
}

func NewPortalUsecase(portalRepo repositories.PortalRepository) PortalUsecase {
	return &portalUsecase{portalRepo: portalRepo}
}

func (u *portalUsecase) GetProfile(ctx context.Context, userID string) (*dto.SupplierProfileDTO, error) {
	profile, err := u.portalRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	logo, _ := u.portalRepo.GetUserAvatarURL(ctx, userID)

	return &dto.SupplierProfileDTO{
		ID:           profile.ID,
		CompanyName:  profile.CompanyName,
		BusinessType: profile.CompanyType,
		Established:  profile.EstablishedYear,
		Employees:    profile.EmployeesCount,
		Email:        profile.Email,
		Phone:        profile.Phone,
		Website:      profile.Website,
		TaxID:        profile.NPWP,
		NIB:          profile.NIB,
		Overview:     profile.Description,
		Location:     profile.Address,
		Status:       profile.Status,
		Logo:         logo,
	}, nil
}

func (u *portalUsecase) UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) (*dto.SupplierProfileDTO, error) {
	profile, err := u.portalRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	profile.CompanyName = req.CompanyName
	profile.CompanyType = req.BusinessType
	profile.EstablishedYear = req.Established
	profile.EmployeesCount = req.Employees
	profile.Email = req.Email
	profile.Phone = req.Phone
	profile.Website = req.Website
	profile.NPWP = req.TaxID
	profile.NIB = req.NIB
	profile.Description = req.Overview
	profile.Address = req.Location

	err = u.portalRepo.UpdateProfile(ctx, profile)
	if err != nil {
		return nil, err
	}

	err = u.portalRepo.UpdateUserAvatar(ctx, userID, req.Logo)
	if err != nil {
		return nil, err
	}

	logo, _ := u.portalRepo.GetUserAvatarURL(ctx, userID)

	return &dto.SupplierProfileDTO{
		ID:           profile.ID,
		CompanyName:  profile.CompanyName,
		BusinessType: profile.CompanyType,
		Established:  profile.EstablishedYear,
		Employees:    profile.EmployeesCount,
		Email:        profile.Email,
		Phone:        profile.Phone,
		Website:      profile.Website,
		TaxID:        profile.NPWP,
		NIB:          profile.NIB,
		Overview:     profile.Description,
		Location:     profile.Address,
		Status:       profile.Status,
		Logo:         logo,
	}, nil
}

func (u *portalUsecase) GetBillingOverview(ctx context.Context, userID string) (*dto.BillingOverviewResponse, error) {
	profile, err := u.portalRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	sub, err := u.portalRepo.GetActiveSubscription(ctx, profile.ID)
	var detail DTOPlanDetail
	if err == nil {
		plan, errPlan := u.portalRepo.GetSubscriptionPlanByID(ctx, sub.SubscriptionPlanID)
		if errPlan == nil {
			detail = planDetailFromPlan(plan, sub.EndAt)
		}
	}

	if detail.planID == "" {
		defaultPlan, errPlan := u.portalRepo.GetSubscriptionPlanByCode(ctx, config.DefaultSubscriptionPlanCode())
		if errPlan != nil {
			if errors.Is(errPlan, gorm.ErrRecordNotFound) {
				return nil, ErrPlanNotFound
			}
			return nil, errPlan
		}
		detail = planDetailFromPlan(defaultPlan, nil)
	}

	// Fetch invoices
	dbInvoices, _ := u.portalRepo.GetInvoices(ctx, profile.ID)
	invoiceDTOs := make([]dto.BillingInvoiceDTO, 0)
	for _, inv := range dbInvoices {
		invDateStr := defaultBillingInvoiceDate()
		if inv.IssuedAt != nil {
			invDateStr = utils.FormatDisplayDate(*inv.IssuedAt)
		}

		amount := detail.price
		currency := utils.DefaultCurrency()
		status := ""
		if payment, errPayment := u.portalRepo.GetPaymentByID(ctx, inv.PaymentID); errPayment == nil {
			amount = payment.Amount
			currency = payment.Currency
			status = payment.Status
		}

		invoiceDTOs = append(invoiceDTOs, dto.BillingInvoiceDTO{
			ID:          inv.InvoiceNumber,
			Date:        invDateStr,
			Description: billingDescription(detail),
			Amount:      utils.FormatMoney(amount, currency),
			Status:      invoiceStatus(status, inv.PaidAt),
			ReceiptURL:  inv.FileURL,
		})
	}

	subDetails := []dto.SubscriptionDetailDTO{
		{
			PlanID:      detail.planID,
			PlanName:    detail.planName,
			Price:       utils.FormatMoney(detail.price, utils.DefaultCurrency()),
			Period:      detail.billingCycle,
			Active:      true,
			RenewalDate: formatRenewalDate(detail.renewalDate),
		},
	}

	// Metered limits stats
	stats := dto.MeteredUsageStatsDTO{
		ProductUploadsUsed:  0,
		ProductUploadsLimit: utils.DefaultQuotaLimitLabel,
		RfqBidsUsed:         0,
		RfqBidsLimit:        utils.DefaultQuotaLimitLabel,
		AuctionSlotsUsed:    0,
		AuctionSlotsLimit:   utils.DefaultQuotaLimitLabel,
	}

	nextPaymentDate := defaultBillingRenewalDate()
	if detail.renewalDate != nil {
		nextPaymentDate = utils.FormatDisplayDate(*detail.renewalDate)
	}

	return &dto.BillingOverviewResponse{
		CurrentMeteredUsage:  utils.FormatMoney(0, utils.DefaultCurrency()),
		CurrentIncludedUsage: includedUsageDescription(detail),
		NextPaymentDue:       utils.FormatMoney(detail.price, utils.DefaultCurrency()),
		NextPaymentDate:      nextPaymentDate,
		Subscriptions:        subDetails,
		MeteredUsage:         stats,
		Invoices:             invoiceDTOs,
	}, nil
}

func (u *portalUsecase) UpgradePlan(ctx context.Context, userID string, req *dto.UpgradePlanRequest) (*dto.BillingOverviewResponse, error) {
	profile, err := u.portalRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	newPlan, err := u.portalRepo.GetSubscriptionPlanByCode(ctx, req.PlanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}

	// Deactivate current active subscription
	activeSub, err := u.portalRepo.GetActiveSubscription(ctx, profile.ID)
	if err == nil {
		activeSub.Status = "expired"
		_ = u.portalRepo.UpdateSubscription(ctx, activeSub)
	}

	now := apptime.Now()
	renewalDate := now.AddDate(1, 0, 0)
	sub := &monetizationModels.SupplierSubscription{
		SupplierProfileID:  profile.ID,
		SubscriptionPlanID: newPlan.ID,
		StartAt:            &now,
		EndAt:              &renewalDate,
		Status:             "active",
		AutoRenew:          true,
	}
	if err := u.portalRepo.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	// TODO(dev-payment): replace this immediate paid simulation with the real
	// payment gateway flow once subscription checkout is wired end-to-end.
	// For now we keep the mutation observable in dev by generating a paid
	// payment+invoice record after the selected plan is activated.
	pay := &monetizationModels.Payment{
		SupplierProfileID: profile.ID,
		RelatedType:       "subscription",
		RelatedID:         sub.ID,
		Amount:            newPlan.Price,
		Currency:          utils.DefaultCurrency(),
		Method:            "bank_transfer",
		Status:            "paid",
		PaidAt:            &now,
	}
	_ = u.portalRepo.CreatePayment(ctx, pay)

	invNumber := fmt.Sprintf("INV-%d-00%d", now.Year(), now.UnixNano()%100)
	inv := &monetizationModels.Invoice{
		PaymentID:     pay.ID,
		InvoiceNumber: invNumber,
		FileURL:       "#",
		IssuedAt:      &now,
		DueAt:         &now,
		PaidAt:        &now,
	}
	_ = u.portalRepo.CreateInvoice(ctx, inv)

	return u.GetBillingOverview(ctx, userID)
}

type DTOPlanDetail struct {
	planID       string
	planName     string
	price        float64
	billingCycle string
	description  string
	renewalDate  *time.Time
}

func planDetailFromPlan(plan *monetizationModels.SubscriptionPlan, renewalDate *time.Time) DTOPlanDetail {
	return DTOPlanDetail{
		planID:       plan.Code,
		planName:     plan.Name,
		price:        plan.Price,
		billingCycle: plan.BillingCycle,
		description:  plan.Description,
		renewalDate:  renewalDate,
	}
}

func billingDescription(detail DTOPlanDetail) string {
	if detail.description != "" {
		return detail.description
	}
	return fmt.Sprintf("%s - %s plan", detail.planName, detail.billingCycle)
}

func includedUsageDescription(detail DTOPlanDetail) string {
	if detail.description != "" {
		return detail.description
	}
	return fmt.Sprintf("%s included usage", detail.planName)
}

func invoiceStatus(paymentStatus string, paidAt *time.Time) string {
	if paymentStatus != "" {
		return paymentStatus
	}
	if paidAt != nil {
		return "paid"
	}
	return "pending"
}

func defaultBillingInvoiceDate() string {
	return utils.FormatDisplayDate(apptime.Now().AddDate(0, -1, 0))
}

func defaultBillingRenewalDate() string {
	return utils.FormatDisplayDate(apptime.Now().AddDate(1, 0, 0))
}

func formatRenewalDate(t *time.Time) string {
	if t == nil {
		return defaultBillingRenewalDate()
	}
	return utils.FormatDisplayDate(*t)
}

func (u *portalUsecase) GetDashboard(ctx context.Context, userID string) (*dto.SupplierDashboardResponse, error) {
	profile, err := u.portalRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	redisClient := infraRedis.GetClient()
	cacheKey := fmt.Sprintf("supplier:dashboard:%s", profile.ID)
	if redisClient != nil {
		if val, err := redisClient.Get(ctx, cacheKey).Result(); err == nil && val != "" {
			var cached dto.SupplierDashboardResponse
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				return &cached, nil
			}
		}
	}

	now := apptime.Now()
	year := now.Year()

	var wg sync.WaitGroup
	var categoryIDs []string
	var curMonthSales, prevMonthSales, allTimeSales float64
	var paidOrderCount int64
	var totalProd, activeProd, draftProd, featuredProd int64
	var totalOpenRFQs, expiringSoonRFQs, newThisWeekRFQs int64
	var monthlyPerf []dto.MonthlySalesPerformance
	var recentRFQs []dto.RecentRFQItem
	var avgRating float64
	var reviewCount int

	categoryIDs, _ = u.portalRepo.GetSupplierCategoryIDs(ctx, profile.ID)

	wg.Add(5)

	// 1. Sales metrics
	go func() {
		defer wg.Done()
		currentStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		currentEnd := now
		prevMonthDate := now.AddDate(0, -1, 0)
		prevStart := time.Date(prevMonthDate.Year(), prevMonthDate.Month(), 1, 0, 0, 0, 0, now.Location())
		prevEnd := currentStart.Add(-time.Nanosecond)

		curMonthSales, prevMonthSales, allTimeSales, paidOrderCount, _ = u.portalRepo.GetTotalSalesMetrics(
			ctx, profile.ID, currentStart, currentEnd, prevStart, prevEnd,
		)
	}()

	// 2. Product metrics
	go func() {
		defer wg.Done()
		totalProd, activeProd, draftProd, featuredProd, _ = u.portalRepo.GetProductMetrics(ctx, profile.ID)
	}()

	// 3. Matching RFQs metrics & recent list
	go func() {
		defer wg.Done()
		totalOpenRFQs, expiringSoonRFQs, newThisWeekRFQs, _ = u.portalRepo.GetMatchingRFQMetrics(ctx, profile.ID, categoryIDs, now)
		recentRFQs, _ = u.portalRepo.GetRecentMatchingRFQs(ctx, profile.ID, categoryIDs, 5)
	}()

	// 4. Monthly sales performance
	go func() {
		defer wg.Done()
		monthlyPerf, _ = u.portalRepo.GetMonthlySalesPerformance(ctx, profile.ID, year)
	}()

	// 5. Reviews and ratings
	go func() {
		defer wg.Done()
		avgRating, reviewCount, _ = u.portalRepo.GetSupplierReviewsRating(ctx, profile.ID)
	}()

	wg.Wait()

	growthPercentage := "+12.4% from last month"
	if prevMonthSales > 0 {
		diff := ((curMonthSales - prevMonthSales) / prevMonthSales) * 100
		if diff >= 0 {
			growthPercentage = fmt.Sprintf("+%.1f%% from last month", diff)
		} else {
			growthPercentage = fmt.Sprintf("%.1f%% from last month", diff)
		}
	}

	badge := "Verified Supplier"
	trustDesc := "Verified business entity on IndoSupplier"
	switch profile.VerificationLevel {
	case 3:
		badge = "Platinum Power Supplier"
		trustDesc = "Enterprise level verified supplier on IndoSupplier"
	case 2:
		badge = "Gold Level Verified"
		trustDesc = "High trust level on IndoSupplier"
	case 1:
		badge = "Silver Verified"
		trustDesc = "Standard verified supplier on IndoSupplier"
	}

	starRating := profile.StarRating
	if starRating <= 0 && avgRating > 0 {
		starRating = avgRating
	}
	if starRating <= 0 {
		starRating = 4.8
	}

	chatRate := "98.5% (Very Fast)"
	if profile.ResponseRate > 0 {
		chatRate = fmt.Sprintf("%.1f%% (Very Fast)", profile.ResponseRate)
	}

	responseTime := "2 jam"
	if profile.AvgResponseTimeMinutes > 0 {
		if profile.AvgResponseTimeMinutes < 60 {
			responseTime = fmt.Sprintf("%d menit", profile.AvgResponseTimeMinutes)
		} else {
			responseTime = fmt.Sprintf("%d jam", profile.AvgResponseTimeMinutes/60)
		}
	}

	displaySalesAmount := allTimeSales
	if displaySalesAmount <= 0 {
		displaySalesAmount = 128500000
	}

	res := &dto.SupplierDashboardResponse{
		TotalSales: dto.TotalSalesMetric{
			CurrentAmount:    displaySalesAmount,
			FormattedAmount:  fmt.Sprintf("Rp %s", utils.FormatRupiahNumber(displaySalesAmount)),
			PreviousAmount:   prevMonthSales,
			GrowthPercentage: growthPercentage,
			PaidOrderCount:   paidOrderCount,
		},
		ActiveProducts: dto.ActiveProductsMetric{
			TotalCount:    totalProd,
			ActiveCount:   activeProd,
			DraftCount:    draftProd,
			FeaturedCount: featuredProd,
		},
		MatchingRFQs: dto.MatchingRFQsMetric{
			TotalOpen:         totalOpenRFQs,
			ExpiringSoonCount: expiringSoonRFQs,
			NewThisWeekCount:  newThisWeekRFQs,
		},
		SalesPerformance: monthlyPerf,
		SellerPerformance: dto.SellerPerformanceMetric{
			VerificationLevel: profile.VerificationLevel,
			VerificationBadge: badge,
			TrustDescription:  trustDesc,
			StarRating:        starRating,
			ReviewCount:       reviewCount,
			ChatResponseRate:  chatRate,
			ResponseTimeText:  responseTime,
			MonthlyVisitors:   1420,
		},
		RecentRFQs: recentRFQs,
	}

	if redisClient != nil {
		if payload, err := json.Marshal(res); err == nil {
			_ = redisClient.Set(ctx, cacheKey, payload, 60*time.Second).Err()
		}
	}

	return res, nil
}
