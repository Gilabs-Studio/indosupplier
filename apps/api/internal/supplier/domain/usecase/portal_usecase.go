package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	monetizationModels "github.com/gilabs/indosupplier/api/internal/monetization/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/models"
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
	for _, p := range billingPlanCatalog {
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

		now := apptime.Now()
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
			Currency:          utils.DefaultCurrency(),
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
		planInfo := billingPlanByCode(detail.planID)
		invDateStr := defaultBillingInvoiceDate()
		if inv.IssuedAt != nil {
			invDateStr = inv.IssuedAt.Format("January 2, 2006")
		}

		invoiceDTOs = append(invoiceDTOs, dto.BillingInvoiceDTO{
			ID:          inv.InvoiceNumber,
			Date:        invDateStr,
			Description: planInfo.invoiceDescription,
			Amount:      utils.FormatMoney(planInfo.price, utils.DefaultCurrency()),
			Status:      "paid",
			ReceiptURL:  "#",
		})
	}

	// If no invoices exist in database, return a default list
	if len(invoiceDTOs) == 0 {
		invoiceDTOs = []dto.BillingInvoiceDTO{
			{
				ID:          "INV-2026-001",
				Date:        defaultBillingInvoiceDate(),
				Description: billingPlanByCode(detail.planID).invoiceDescription,
				Amount:      utils.FormatMoney(billingPlanByCode(detail.planID).price, utils.DefaultCurrency()),
				Status:      "paid",
				ReceiptURL:  "#",
			},
		}
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
		ProductUploadsUsed:  142,
		ProductUploadsLimit: utils.DefaultQuotaLimitLabel,
		RfqBidsUsed:         85,
		RfqBidsLimit:        utils.DefaultQuotaLimitLabel,
		AuctionSlotsUsed:    1,
		AuctionSlotsLimit:   utils.DefaultQuotaLimitLabel,
	}

	nextPaymentDate := defaultBillingRenewalDate()
	if detail.renewalDate != nil {
		nextPaymentDate = detail.renewalDate.Format("January 2, 2006")
	}

	return &dto.BillingOverviewResponse{
		CurrentMeteredUsage:  utils.FormatMoney(0, utils.DefaultCurrency()),
		CurrentIncludedUsage: fmt.Sprintf("%s Tier Limits Included", detail.planName),
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

	// Create payment and invoice logs for the upgrade
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
	renewalDate  *time.Time
}

type billingPlanSeed struct {
	code               string
	name               string
	price              float64
	cycle              string
	invoiceDescription string
}

var billingPlanCatalog = []billingPlanSeed{
	{code: "free", name: "Free Basic", price: 0, cycle: "month", invoiceDescription: "GIMS Free Basic - Monthly plan"},
	{code: "bronze", name: "Bronze Seller", price: 2000000, cycle: "year", invoiceDescription: "GIMS Bronze Seller - Subscription Upgrade"},
	{code: "silver", name: "Silver Pro", price: 5000000, cycle: "year", invoiceDescription: "GIMS Silver Pro - Subscription Upgrade"},
	{code: "gold", name: "Gold Enterprise", price: 12000000, cycle: "year", invoiceDescription: "GIMS Gold Enterprise - Annual plan"},
}

func billingPlanByCode(code string) billingPlanSeed {
	for _, plan := range billingPlanCatalog {
		if plan.code == code {
			return plan
		}
	}
	return billingPlanCatalog[0]
}

func defaultBillingInvoiceDate() string {
	return apptime.Now().AddDate(0, -1, 0).Format("January 2, 2006")
}

func defaultBillingRenewalDate() string {
	return apptime.Now().AddDate(1, 0, 0).Format("January 2, 2006")
}

func formatRenewalDate(t *time.Time) string {
	if t == nil {
		return defaultBillingRenewalDate()
	}
	return t.Format("January 2, 2006")
}
