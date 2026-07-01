package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"
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
