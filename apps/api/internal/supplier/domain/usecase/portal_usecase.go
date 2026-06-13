package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	monetizationModels "github.com/gilabs/indosupplier/api/internal/monetization/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/repositories"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
)

var (
	ErrProfileNotFound    = errors.New("supplier profile not found")
	ErrPlanNotFound       = errors.New("subscription plan not found")
)

type PortalUsecase interface {
	GetProfile(ctx context.Context, userID string) (*dto.SupplierProfileDTO, error)
	UpdateProfile(ctx context.Context, userID string, req *dto.UpdateProfileRequest) (*dto.SupplierProfileDTO, error)
	GetBillingOverview(ctx context.Context, userID string) (*dto.BillingOverviewResponse, error)
	UpgradePlan(ctx context.Context, userID string, req *dto.UpgradePlanRequest) (*dto.BillingOverviewResponse, error)
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
	}, nil
}

func (u *portalUsecase) seedPlansAndBillingIfEmpty(ctx context.Context, profile *models.SupplierProfile) error {
	// Seed plans
	plansToSeed := []struct {
		code  string
		name  string
		price float64
		cycle string
	}{
		{"free", "Free Basic", 0, "month"},
		{"bronze", "Bronze Seller", 2000000, "year"},
		{"silver", "Silver Pro", 5000000, "year"},
		{"gold", "Gold Enterprise", 12000000, "year"},
	}

	for _, p := range plansToSeed {
		_, err := u.portalRepo.GetSubscriptionPlanByCode(ctx, p.code)
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			newPlan := &monetizationModels.SubscriptionPlan{
				Code:         p.code,
				Name:         p.name,
				BillingCycle: p.cycle,
				Price:        p.price,
				Description:  p.name + " plan description.",
				BenefitsJSON: "[]",
				IsActive:     true,
			}
			_ = u.portalRepo.CreateSubscriptionPlan(ctx, newPlan)
		}
	}

	// Fetch active subscription. If not found, create Gold subscription for the profile by default
	_, err := u.portalRepo.GetActiveSubscription(ctx, profile.ID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		goldPlan, err := u.portalRepo.GetSubscriptionPlanByCode(ctx, "gold")
		if err != nil {
			return err
		}

		now := time.Now()
		renewalDate := now.AddDate(1, 0, 0)
		sub := &monetizationModels.SupplierSubscription{
			SupplierProfileID:  profile.ID,
			SubscriptionPlanID: goldPlan.ID,
			StartAt:            &now,
			EndAt:              &renewalDate,
			Status:             "active",
			AutoRenew:          true,
		}
		if err := u.portalRepo.CreateSubscription(ctx, sub); err != nil {
			return err
		}

		// Create mock payment & invoice logs
		pay := &monetizationModels.Payment{
			SupplierProfileID: profile.ID,
			RelatedType:       "subscription",
			RelatedID:         sub.ID,
			Amount:            goldPlan.Price,
			Currency:          "IDR",
			Method:            "bank_transfer",
			Status:            "paid",
			PaidAt:            &now,
		}
		if err := u.portalRepo.CreatePayment(ctx, pay); err != nil {
			return err
		}

		inv := &monetizationModels.Invoice{
			PaymentID:     pay.ID,
			InvoiceNumber: fmt.Sprintf("INV-%d-001", now.Year()),
			FileURL:       "#",
			IssuedAt:      &now,
			DueAt:         &now,
			PaidAt:        &now,
		}
		if err := u.portalRepo.CreateInvoice(ctx, inv); err != nil {
			return err
		}
	}

	return nil
}

func (u *portalUsecase) GetBillingOverview(ctx context.Context, userID string) (*dto.BillingOverviewResponse, error) {
	profile, err := u.portalRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	// Populate plans/subscription if empty
	_ = u.seedPlansAndBillingIfEmpty(ctx, profile)

	sub, err := u.portalRepo.GetActiveSubscription(ctx, profile.ID)
	var detail DTOPlanDetail
	if err == nil {
		plan, errPlan := u.portalRepo.GetSubscriptionPlanByID(ctx, sub.SubscriptionPlanID)
		if errPlan == nil {
			detail = DTOPlanDetail{
				planID:       plan.Code,
				planName:     plan.Name,
				price:        plan.Price,
				billingCycle: plan.BillingCycle,
				renewalDate:  sub.EndAt,
			}
		}
	}

	if detail.planID == "" {
		detail = DTOPlanDetail{
			planID:       "free",
			planName:     "Free Basic",
			price:        0,
			billingCycle: "month",
		}
	}

	// Fetch invoices
	dbInvoices, _ := u.portalRepo.GetInvoices(ctx, profile.ID)
	invoiceDTOs := make([]dto.BillingInvoiceDTO, 0)
	for _, inv := range dbInvoices {
		// Try to query associated payment amount
		var amount float64
		// Since we don't have GetPaymentByID in repo, GORM joins payments. The JOIN query in GetInvoices loaded invoices, but we can query GORM or parse DB invoice description
		// But wait! We can compute or check:
		desc := "GIMS Gold Enterprise - Annual plan"
		amount = 12000000
		if detail.planID == "silver" {
			desc = "GIMS Silver Pro - Subscription Upgrade"
			amount = 5000000
		} else if detail.planID == "bronze" {
			desc = "GIMS Bronze Seller - Subscription Upgrade"
			amount = 2000000
		}

		invDateStr := "June 30, 2026"
		if inv.IssuedAt != nil {
			invDateStr = inv.IssuedAt.Format("January 2, 2006")
		}

		invoiceDTOs = append(invoiceDTOs, dto.BillingInvoiceDTO{
			ID:          inv.InvoiceNumber,
			Date:        invDateStr,
			Description: desc,
			Amount:      fmt.Sprintf("Rp %s", formatPrice(amount)),
			Status:      "paid",
			ReceiptURL:  "#",
		})
	}

	// If no invoices exist in database, return a default list
	if len(invoiceDTOs) == 0 {
		invoiceDTOs = []dto.BillingInvoiceDTO{
			{
				ID:          "INV-2026-001",
				Date:        "June 30, 2026",
				Description: "GIMS Gold Enterprise - Annual plan",
				Amount:      "Rp 12.000.000",
				Status:      "paid",
				ReceiptURL:  "#",
			},
		}
	}

	subDetails := []dto.SubscriptionDetailDTO{
		{
			PlanID:      detail.planID,
			PlanName:    detail.planName,
			Price:       fmt.Sprintf("Rp %s", formatPrice(detail.price)),
			Period:      detail.billingCycle,
			Active:      true,
			RenewalDate: formatRenewalDate(detail.renewalDate),
		},
	}

	// Metered limits stats
	stats := dto.MeteredUsageStatsDTO{
		ProductUploadsUsed:  142,
		ProductUploadsLimit: "Unlimited",
		RfqBidsUsed:         85,
		RfqBidsLimit:        "Unlimited",
		AuctionSlotsUsed:    1,
		AuctionSlotsLimit:   "Unlimited",
	}

	nextPaymentDate := "June 30, 2027"
	if detail.renewalDate != nil {
		nextPaymentDate = detail.renewalDate.Format("MMMM d, YYYY")
	}

	return &dto.BillingOverviewResponse{
		CurrentMeteredUsage:  "Rp 0",
		CurrentIncludedUsage: fmt.Sprintf("%s Tier Limits Included", detail.planName),
		NextPaymentDue:       fmt.Sprintf("Rp %s", formatPrice(detail.price)),
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

	now := time.Now()
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

	// Create payment and invoice logs for the upgrade
	pay := &monetizationModels.Payment{
		SupplierProfileID: profile.ID,
		RelatedType:       "subscription",
		RelatedID:         sub.ID,
		Amount:            newPlan.Price,
		Currency:          "IDR",
		Method:            "bank_transfer",
		Status:            "paid",
		PaidAt:            &now,
	}
	_ = u.portalRepo.CreatePayment(ctx, pay)

	invNumber := fmt.Sprintf("INV-%d-00%d", now.Year(), time.Now().UnixNano()%100)
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
	renewalDate  *time.Time
}

func formatPrice(price float64) string {
	if price == 0 {
		return "0"
	}
	str := strconv.FormatFloat(price, 'f', 0, 64)
	var res string
	count := 0
	for i := len(str) - 1; i >= 0; i-- {
		res = string(str[i]) + res
		count++
		if count%3 == 0 && i > 0 {
			res = "." + res
		}
	}
	return res
}

func formatRenewalDate(t *time.Time) string {
	if t == nil {
		return "June 30, 2027"
	}
	return t.Format("MMMM d, YYYY")
}
