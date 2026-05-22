package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	xenditClient "github.com/gilabs/gims/api/internal/core/infrastructure/xendit"
	"github.com/gilabs/gims/api/internal/tenant/data/models"
	tenantDTO "github.com/gilabs/gims/api/internal/tenant/domain/dto"
	couponUC "github.com/gilabs/gims/api/internal/tenant/domain/usecase"
	"github.com/google/uuid"
	"gorm.io/gorm"
	gormclause "gorm.io/gorm/clause"
)

// PendingBillingChange is cached until Xendit confirms the invoice payment.
type PendingBillingChange struct {
	Token                string `json:"token"`
	TenantID             string `json:"tenant_id"`
	TenantName           string `json:"tenant_name,omitempty"`
	TenantOwnerEmail     string `json:"tenant_owner_email,omitempty"`
	Action               string `json:"action"`
	Target               string `json:"target"`
	ActionDate           string `json:"action_date"`
	CurrentPlan          string `json:"current_plan"`
	CurrentBillingPeriod string `json:"current_billing_period"`
	CurrentSeatLimit     int    `json:"current_seat_limit"`
	CurrentOutletLimit   int    `json:"current_outlet_limit"`
	ActiveUsers          int    `json:"active_users"`
	CurrentAmount        int64  `json:"current_amount"`
	NewSeatLimit         int    `json:"new_seat_limit"`
	NewOutletLimit       int    `json:"new_outlet_limit"`
	NewPlan              string `json:"new_plan"`
	NewAmount            int64  `json:"new_amount"`
	ProrationAmount      int64  `json:"proration_amount"`
	CouponCode           string `json:"coupon_code,omitempty"`
	CouponID             string `json:"coupon_id,omitempty"`
	CouponMarkAsUsed     bool   `json:"coupon_mark_as_used"`
	InvoiceID            string `json:"invoice_id,omitempty"`
	XenditCustomerID     string `json:"xendit_customer_id,omitempty"`
	XenditPlanID         string `json:"xendit_plan_id,omitempty"`
	IdempotencyKey       string `json:"idempotency_key"`
	CreatedAt            string `json:"created_at"`
}

type billingSubscriptionSnapshot struct {
	SubscriptionID        string
	TenantName            string
	OwnerEmail            string
	CurrentSeatLimit      int
	CurrentOutletLimit    int
	ActiveUsers           int
	CurrentAmount         int64
	PlanSlug              string
	BillingPeriod         string
	NextBillingAt         *time.Time
	ExpiresAt             *time.Time
	Subscription          *models.TenantSubscription
	CurrentPlan           *models.SubscriptionPlanConfig
	TargetPlan            *models.SubscriptionPlanConfig
	ActionDate            time.Time
	CurrentBillingPeriod  models.SubscriptionBillingPeriod
	CurrentPlanPrice      int64
	TargetPlanPrice       int64
	CurrentAmountPerCycle int64
	NewAmountPerCycle     int64
	Coupon                *models.Coupon
}

func (u *authUsecase) CreateBillingChangeInvoice(ctx context.Context, req *tenantDTO.BillingChangeRequest, tenantID string) (*tenantDTO.BillingChangeResponse, error) {
	if u.xendit == nil || !u.xendit.IsConfigured() {
		return nil, ErrPaymentGatewayUnavailable
	}

	snapshot, err := u.loadBillingChangeSnapshot(ctx, req, tenantID)
	if err != nil {
		return nil, err
	}

	billingChange, couponApplied, invoicePayload, pending, err := u.buildBillingChange(ctx, req, snapshot)
	if err != nil {
		return nil, err
	}

	response := &tenantDTO.BillingChangeResponse{Status: "ok", SyncRequired: snapshot.CurrentAmount != billingChange.OldAmountPerCycle}
	response.BillingChange.OldSeatLimit = billingChange.OldSeatLimit
	response.BillingChange.NewSeatLimit = billingChange.NewSeatLimit
	response.BillingChange.OldOutletLimit = billingChange.OldOutletLimit
	response.BillingChange.NewOutletLimit = billingChange.NewOutletLimit
	response.BillingChange.OldAmountPerCycle = billingChange.OldAmountPerCycle
	response.BillingChange.NewAmountPerCycle = billingChange.NewAmountPerCycle
	response.BillingChange.ProrationAmount = billingChange.ProrationAmount
	response.BillingChange.ProrationWaivedReason = billingChange.ProrationWaivedReason
	response.CouponApplied.Code = couponApplied.Code
	response.CouponApplied.DiscountAmount = couponApplied.DiscountAmount
	response.CouponApplied.MarkAsUsed = couponApplied.MarkAsUsed
	response.UserNotification.Title = billingChange.NotificationTitle
	response.UserNotification.Message = billingChange.NotificationMessage
	response.UserNotification.AmountDueNow = billingChange.AmountDueNow

	if billingChange.AmountDueNow <= 0 {
		if err := u.applyBillingChangeImmediately(ctx, snapshot, billingChange, couponApplied); err != nil {
			return nil, err
		}
		response.XenditAction = "none"
		response.XenditPayload = map[string]any{}
		return response, nil
	}

	invoice, err := u.xendit.CreateInvoice(ctx, invoicePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to create billing invoice: %w", err)
	}
	pending.InvoiceID = invoice.ID
	pending.CreatedAt = apptime.Now().Format(time.RFC3339)
	if err := u.storePendingBillingChange(ctx, pending); err != nil {
		return nil, err
	}

	response.XenditAction = "create_invoice"
	response.XenditPayload = map[string]any{
		"external_id":          invoice.ExternalID,
		"invoice_id":           invoice.ID,
		"invoice_url":          invoice.InvoiceURL,
		"amount":               invoice.Amount,
		"currency":             invoice.Currency,
		"description":          invoice.Description,
		"expiry_date":          invoice.ExpiryDate.Format(time.RFC3339),
		"idempotency_key":      pending.IdempotencyKey,
		"xendit_customer_id":   pending.XenditCustomerID,
		"xendit_plan_id":       pending.XenditPlanID,
		"action":               pending.Action,
		"target":               pending.Target,
		"proration_amount":     pending.ProrationAmount,
		"new_amount_per_cycle": pending.NewAmount,
	}
	return response, nil
}

func (u *authUsecase) CompletePendingBillingChange(ctx context.Context, token string) error {
	if u.redis == nil {
		return ErrPendingBillingChangeNotFound
	}

	key := pendingBillingChangeKeyPrefix + token
	val, err := u.redis.Get(ctx, key).Result()
	if err != nil {
		return ErrPendingBillingChangeNotFound
	}

	var pending PendingBillingChange
	if err := json.Unmarshal([]byte(val), &pending); err != nil {
		return ErrPendingBillingChangeInvalid
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sub models.TenantSubscription
		if err := tx.Where("tenant_id = ? AND status IN ? AND deleted_at IS NULL", pending.TenantID, []string{string(models.SubscriptionActive), string(models.SubscriptionTrial)}).
			Order("starts_at DESC").First(&sub).Error; err != nil {
			return err
		}

		updates := map[string]any{
			"seat_limit":        pending.NewSeatLimit,
			"outlet_limit":      pending.NewOutletLimit,
			"user_count":        pending.NewSeatLimit,
			"xendit_invoice_id": pending.InvoiceID,
		}
		if amountColumn, colErr := u.resolveTenantSubscriptionAmountColumn(ctx); colErr != nil {
			return colErr
		} else if amountColumn != "" {
			updates[amountColumn] = pending.NewAmount
		}
		if pending.NewPlan != "" {
			updates["plan"] = models.SubscriptionPlan(pending.NewPlan)
		}
		if pending.CouponID != "" {
			updates["coupon_id"] = pending.CouponID
		}

		if err := tx.Model(&models.TenantSubscription{}).Where("id = ?", sub.ID).Updates(updates).Error; err != nil {
			return err
		}

		hasTenantColumn, schemaErr := u.paymentTransactionsHasTenantColumn(ctx)
		if schemaErr != nil {
			return schemaErr
		}
		supportsOnConflict, conflictErr := u.paymentTransactionsSupportsProviderInvoiceOnConflict(ctx)
		if conflictErr != nil {
			return conflictErr
		}
		billingChangeMeta := fmt.Sprintf(`{"action":"%s","idempotency_key":"%s"}`,
			strings.TrimSpace(pending.Action), strings.TrimSpace(pending.IdempotencyKey))
		if hasTenantColumn {
			paymentTxn := &models.PaymentTransaction{
				TenantID:          pending.TenantID,
				Provider:          models.PaymentProviderXendit,
				Status:            models.PaymentStatusPaid,
				AmountIDR:         pending.ProrationAmount,
				ProviderInvoiceID: pending.InvoiceID,
				Description:       fmt.Sprintf("Billing change for %s", pending.Target),
				PaidAt:            ptrTime(apptime.Now()),
				SubscriptionID:    sub.ID,
				Metadata:          billingChangeMeta,
				Notes:             billingChangeMeta,
			}
			if pending.ProrationAmount == 0 {
				paymentTxn.AmountIDR = 0
			}
			if supportsOnConflict {
				if err := tx.Clauses(gormclause.OnConflict{Columns: []gormclause.Column{{Name: "provider"}, {Name: "provider_invoice_id"}}, DoNothing: true}).Create(paymentTxn).Error; err != nil {
					return err
				}
			} else {
				exists, existsErr := paymentTransactionExists(tx, models.PaymentProviderXendit, pending.InvoiceID)
				if existsErr != nil {
					return existsErr
				}
				if !exists {
					if err := tx.Create(paymentTxn).Error; err != nil {
						return err
					}
				}
			}
		} else {
			now := apptime.Now()
			legacyTxn := map[string]any{
				"id":                  uuid.NewString(),
				"subscription_id":     sub.ID,
				"provider":            models.PaymentProviderXendit,
				"status":              models.PaymentStatusPaid,
				"amount_idr":          pending.ProrationAmount,
				"provider_invoice_id": pending.InvoiceID,
				"description":         fmt.Sprintf("Billing change for %s", pending.Target),
				"paid_at":             ptrTime(now),
				"metadata":            billingChangeMeta,
				"notes":               billingChangeMeta,
				"created_at":          now,
				"updated_at":          now,
			}
			if supportsOnConflict {
				if err := tx.Table("payment_transactions").
					Clauses(gormclause.OnConflict{Columns: []gormclause.Column{{Name: "provider"}, {Name: "provider_invoice_id"}}, DoNothing: true}).
					Create(legacyTxn).Error; err != nil {
					return err
				}
			} else {
				exists, existsErr := paymentTransactionExists(tx, models.PaymentProviderXendit, pending.InvoiceID)
				if existsErr != nil {
					return existsErr
				}
				if !exists {
					if err := tx.Table("payment_transactions").Create(legacyTxn).Error; err != nil {
						return err
					}
				}
			}
		}

		if pending.CouponMarkAsUsed && pending.CouponID != "" {
			if err := tx.Model(&models.Coupon{}).Where("id = ?", pending.CouponID).
				Updates(map[string]any{"used_count": gorm.Expr("used_count + 1")}).Error; err != nil {
				return err
			}
		}

		return u.redis.Del(ctx, key).Err()
	})
}

type billingChangeDecision struct {
	OldSeatLimit          int
	NewSeatLimit          int
	OldOutletLimit        int
	NewOutletLimit        int
	OldAmountPerCycle     int64
	NewAmountPerCycle     int64
	ProrationAmount       int64
	ProrationWaivedReason string
	AmountDueNow          int64
	NotificationTitle     string
	NotificationMessage   string
	NewPlanSlug           string
	IdempotencyKey        string
	XenditCustomerID      string
	XenditPlanID          string
	ExternalID            string
	Action                string
	Target                string
}

type couponApplicationDecision struct {
	Code           string
	DiscountAmount int64
	MarkAsUsed     bool
}

func (u *authUsecase) loadBillingChangeSnapshot(ctx context.Context, req *tenantDTO.BillingChangeRequest, tenantID string) (*billingSubscriptionSnapshot, error) {
	var sub models.TenantSubscription
	if err := u.db.WithContext(ctx).
		Preload("Coupon").
		Where("tenant_id = ? AND status IN ? AND deleted_at IS NULL", tenantID, []string{string(models.SubscriptionActive), string(models.SubscriptionTrial)}).
		Order("starts_at DESC").
		First(&sub).Error; err != nil {
		return nil, ErrPendingBillingChangeNotFound
	}

	activeUsers, err := u.countActiveTenantUsers(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	ownerEmail, tenantName, err := u.resolveTenantOwnerEmail(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	currentSeatLimit := sub.SeatLimit
	if sub.UserCount > currentSeatLimit {
		currentSeatLimit = sub.UserCount
	}
	if currentSeatLimit <= 0 {
		currentSeatLimit = sub.UserCount
	}
	if currentSeatLimit <= 0 {
		currentSeatLimit = 1
	}

	currentOutletLimit := sub.OutletLimit
	if currentOutletLimit <= 0 {
		currentOutletLimit = 1
	}

	currentPlan, err := u.planRepo.FindBySlug(ctx, string(sub.Plan))
	if err != nil {
		return nil, err
	}

	actionDate, err := time.Parse("2006-01-02", req.ActionDate)
	if err != nil {
		return nil, err
	}

	return &billingSubscriptionSnapshot{
		SubscriptionID:        sub.ID,
		TenantName:            tenantName,
		OwnerEmail:            ownerEmail,
		CurrentSeatLimit:      currentSeatLimit,
		CurrentOutletLimit:    currentOutletLimit,
		ActiveUsers:           int(activeUsers),
		CurrentAmount:         sub.AmountPaidIDR,
		PlanSlug:              string(sub.Plan),
		BillingPeriod:         string(sub.BillingPeriod),
		NextBillingAt:         sub.NextBillingAt,
		ExpiresAt:             sub.ExpiresAt,
		Subscription:          &sub,
		CurrentPlan:           currentPlan,
		CurrentBillingPeriod:  models.SubscriptionBillingPeriod(sub.BillingPeriod),
		ActionDate:            actionDate,
		CurrentAmountPerCycle: sub.AmountPaidIDR,
	}, nil
}

func (u *authUsecase) buildBillingChange(ctx context.Context, req *tenantDTO.BillingChangeRequest, snapshot *billingSubscriptionSnapshot) (*billingChangeDecision, couponApplicationDecision, xenditClient.CreateInvoiceRequest, PendingBillingChange, error) {
	decision := &billingChangeDecision{
		OldSeatLimit:      snapshot.CurrentSeatLimit,
		OldOutletLimit:    snapshot.CurrentOutletLimit,
		OldAmountPerCycle: snapshot.CurrentAmountPerCycle,
		Action:            string(req.Action),
		Target:            req.Target,
		IdempotencyKey:    deriveBillingIdempotencyKey(req, snapshot),
		XenditCustomerID:  req.XenditCustomerID,
		XenditPlanID:      req.XenditPlanID,
	}

	newSeatLimit := snapshot.CurrentSeatLimit
	newOutletLimit := snapshot.CurrentOutletLimit
	newPlanSlug := snapshot.PlanSlug
	var targetPlan *models.SubscriptionPlanConfig
	var oneTimeOutletAmount int64

	switch req.Action {
	case tenantDTO.BillingChangeActionAddSeat:
		parsedSeatLimit, parseErr := strconv.Atoi(strings.TrimSpace(req.Target))
		if parseErr != nil || parsedSeatLimit < snapshot.CurrentSeatLimit {
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, ErrSeatLimitExceeded
		}
		newSeatLimit = parsedSeatLimit
	case tenantDTO.BillingChangeActionAddOutlet:
		parsedOutletLimit, parseErr := strconv.Atoi(strings.TrimSpace(req.Target))
		if parseErr != nil || parsedOutletLimit < snapshot.CurrentOutletLimit {
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, ErrBillingChangeUnsupported
		}
		newOutletLimit = parsedOutletLimit
		outletAddonCount := newOutletLimit - snapshot.CurrentOutletLimit
		if outletAddonCount < 0 {
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, ErrBillingChangeUnsupported
		}
		oneTimeOutletAmount = int64(outletAddonCount) * snapshot.CurrentPlan.OutletAddonPriceIDR(snapshot.BillingPeriod)
	case tenantDTO.BillingChangeActionDowngradeOutlet:
		parsedOutletLimit, parseErr := strconv.Atoi(strings.TrimSpace(req.Target))
		if parseErr != nil || parsedOutletLimit >= snapshot.CurrentOutletLimit {
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, ErrBillingChangeUnsupported
		}
		if parsedOutletLimit < 1 {
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, ErrBillingChangeUnsupported
		}
		newOutletLimit = parsedOutletLimit
	case tenantDTO.BillingChangeActionDowngrade:
		parsedSeatLimit, parseErr := strconv.Atoi(strings.TrimSpace(req.Target))
		if parseErr != nil || parsedSeatLimit >= snapshot.CurrentSeatLimit {
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, ErrBillingChangeUnsupported
		}
		if parsedSeatLimit < snapshot.ActiveUsers {
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, ErrSeatLimitExceeded
		}
		newSeatLimit = parsedSeatLimit
	case tenantDTO.BillingChangeActionUpgradePlan:
		targetPlanModel, err := u.planRepo.FindBySlug(ctx, strings.TrimSpace(req.Target))
		if err != nil {
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, err
		}
		targetPlan = targetPlanModel
		newPlanSlug = targetPlanModel.Slug
	default:
		return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, ErrBillingChangeUnsupported
	}

	currentAmount := snapshot.CurrentPlan.TotalPriceWithOutletAddonsIDR(snapshot.BillingPeriod, snapshot.CurrentSeatLimit, snapshot.CurrentOutletLimit)
	var newAmount int64
	if targetPlan != nil {
		newAmount = targetPlan.TotalPriceWithOutletAddonsIDR(snapshot.BillingPeriod, newSeatLimit, newOutletLimit)
	} else {
		newAmount = snapshot.CurrentPlan.TotalPriceWithOutletAddonsIDR(snapshot.BillingPeriod, newSeatLimit, newOutletLimit)
	}

	if newAmount < currentAmount && req.Action == tenantDTO.BillingChangeActionUpgradePlan {
		return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, ErrBillingChangeUnsupported
	}

	decision.NewSeatLimit = newSeatLimit
	decision.NewOutletLimit = newOutletLimit
	decision.NewAmountPerCycle = newAmount
	decision.NewPlanSlug = newPlanSlug
	if req.Action == tenantDTO.BillingChangeActionAddOutlet {
		decision.NewAmountPerCycle = currentAmount
		decision.ProrationAmount = oneTimeOutletAmount
		decision.ProrationWaivedReason = "one_time_outlet_addon"
		decision.AmountDueNow = oneTimeOutletAmount
	} else {
		prorationAmount, waivedReason := calculateProrationAmount(snapshot.ActionDate, snapshot.NextBillingAt, snapshot.ExpiresAt, currentAmount, newAmount)
		decision.ProrationAmount = prorationAmount
		decision.ProrationWaivedReason = waivedReason
		decision.AmountDueNow = prorationAmount
	}

	couponDecision := couponApplicationDecision{Code: strings.ToUpper(strings.TrimSpace(req.CouponCode))}
	if couponDecision.Code != "" {
		couponResp, err := u.couponUC.Validate(ctx, couponDecision.Code)
		if err != nil || couponResp == nil || !couponResp.Valid {
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, errors.New("COUPON_INVALID")
		}
		var coupon models.Coupon
		if err := u.db.WithContext(ctx).Where("code = ? AND deleted_at IS NULL", couponDecision.Code).First(&coupon).Error; err != nil {
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, errors.New("COUPON_INVALID")
		}
		couponDecision.MarkAsUsed = coupon.MaxUses <= 1
		baseAmount := newAmount
		discountUserCount := newSeatLimit
		if req.Action == tenantDTO.BillingChangeActionAddOutlet {
			baseAmount = decision.AmountDueNow
			discountUserCount = 1
		}
		discountedAmount, discountErr := u.couponUC.ApplyDiscount(
			ctx,
			couponDecision.Code,
			newPlanSlug,
			baseAmount,
			discountUserCount,
			snapshot.BillingPeriod,
		)
		if discountErr != nil {
			if errors.Is(discountErr, couponUC.ErrCouponUserLimit) {
				return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, ErrCouponUserLimitExceeded
			}
			return nil, couponApplicationDecision{}, xenditClient.CreateInvoiceRequest{}, PendingBillingChange{}, discountErr
		}
		couponDecision.DiscountAmount = baseAmount - discountedAmount
		if req.Action == tenantDTO.BillingChangeActionAddOutlet {
			decision.AmountDueNow = discountedAmount
			decision.ProrationAmount = discountedAmount
		} else {
			decision.NewAmountPerCycle = discountedAmount
		}
	}

	pending := PendingBillingChange{
		Token:                decision.IdempotencyKey,
		TenantID:             snapshot.Subscription.TenantID,
		TenantName:           snapshot.TenantName,
		TenantOwnerEmail:     snapshot.OwnerEmail,
		Action:               decision.Action,
		Target:               decision.Target,
		ActionDate:           snapshot.ActionDate.Format(time.RFC3339),
		CurrentPlan:          snapshot.PlanSlug,
		CurrentBillingPeriod: snapshot.BillingPeriod,
		CurrentSeatLimit:     snapshot.CurrentSeatLimit,
		CurrentOutletLimit:   snapshot.CurrentOutletLimit,
		ActiveUsers:          snapshot.ActiveUsers,
		CurrentAmount:        currentAmount,
		NewSeatLimit:         newSeatLimit,
		NewOutletLimit:       newOutletLimit,
		NewPlan:              newPlanSlug,
		NewAmount:            decision.NewAmountPerCycle,
		ProrationAmount:      decision.ProrationAmount,
		CouponCode:           couponDecision.Code,
		CouponID:             couponIDIfAny(couponDecision.Code, u.db, ctx),
		CouponMarkAsUsed:     couponDecision.MarkAsUsed,
		XenditCustomerID:     req.XenditCustomerID,
		XenditPlanID:         req.XenditPlanID,
		IdempotencyKey:       decision.IdempotencyKey,
		CreatedAt:            apptime.Now().Format(time.RFC3339),
	}

	if decision.AmountDueNow <= 0 {
		if decision.NewAmountPerCycle < decision.OldAmountPerCycle {
			decision.NotificationTitle = "Downgrade langganan disiapkan"
			decision.NotificationMessage = "Seat limit Anda akan diturunkan sekarang. Tidak ada refund untuk siklus berjalan; biaya recurring yang lebih rendah akan berlaku pada siklus billing berikutnya sesuai alur subscription Xendit."
		} else {
			decision.NotificationTitle = "Perubahan langganan disiapkan"
			decision.NotificationMessage = "Perubahan langganan akan diterapkan tanpa tagihan tambahan sekarang karena sisa periode aktif sudah terlalu singkat."
		}
	} else {
		if req.Action == tenantDTO.BillingChangeActionAddOutlet {
			decision.NotificationTitle = "Pembayaran add-on outlet siap"
			decision.NotificationMessage = "Kami menyiapkan invoice sekali bayar untuk penambahan outlet. Setelah pembayaran berhasil, limit outlet Anda akan diperbarui otomatis tanpa mengubah biaya recurring."
		} else {
			decision.NotificationTitle = "Tagihan perubahan langganan siap"
			decision.NotificationMessage = "Kami menyiapkan invoice perubahan langganan. Setelah pembayaran berhasil, limit dan paket Anda akan diperbarui otomatis."
		}
	}

	frontendBaseURL := ""
	if config.AppConfig != nil {
		frontendBaseURL = config.AppConfig.Server.FrontendBaseURL
	}

	invoiceReq := xenditClient.CreateInvoiceRequest{
		ExternalID:         decision.IdempotencyKey,
		Amount:             decision.AmountDueNow,
		PayerEmail:         snapshot.OwnerEmail,
		Description:        fmt.Sprintf("Billing change for %s (%s)", snapshot.TenantName, decision.Target),
		SuccessRedirectURL: frontendBaseURL + "/profile?tab=billing&billing_status=paid&billing_token=" + decision.IdempotencyKey,
		FailureRedirectURL: frontendBaseURL + "/profile?tab=billing&billing_status=failed&billing_token=" + decision.IdempotencyKey,
		Currency:           "IDR",
		InvoiceDuration:    86400,
	}
	if req.Action == tenantDTO.BillingChangeActionAddOutlet {
		invoiceReq.Description = fmt.Sprintf("One-time outlet addon for %s (%s)", snapshot.TenantName, decision.Target)
	}

	return decision, couponDecision, invoiceReq, pending, nil
}

func (u *authUsecase) applyBillingChangeImmediately(ctx context.Context, snapshot *billingSubscriptionSnapshot, decision *billingChangeDecision, couponDecision couponApplicationDecision) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"seat_limit":   decision.NewSeatLimit,
			"outlet_limit": decision.NewOutletLimit,
			"user_count":   decision.NewSeatLimit,
		}
		if amountColumn, colErr := u.resolveTenantSubscriptionAmountColumn(ctx); colErr != nil {
			return colErr
		} else if amountColumn != "" {
			updates[amountColumn] = decision.NewAmountPerCycle
		}
		if decision.NewPlanSlug != "" {
			updates["plan"] = models.SubscriptionPlan(decision.NewPlanSlug)
		}
		if couponDecision.Code != "" {
			coupon, err := loadCouponByCode(tx, couponDecision.Code)
			if err != nil {
				return err
			}
			updates["coupon_id"] = coupon.ID
			if couponDecision.MarkAsUsed {
				if err := markCouponUsed(tx, coupon.ID); err != nil {
					return err
				}
			}
		}
		if err := tx.Model(&models.TenantSubscription{}).Where("id = ?", snapshot.SubscriptionID).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
}

func (u *authUsecase) paymentTransactionsHasTenantColumn(ctx context.Context) (bool, error) {
	type result struct {
		Exists bool `gorm:"column:exists"`
	}
	var row result
	err := u.db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = 'payment_transactions'
			  AND column_name = 'tenant_id'
		) AS exists
	`).Scan(&row).Error
	if err != nil {
		return false, err
	}
	return row.Exists, nil
}

func (u *authUsecase) resolveTenantSubscriptionAmountColumn(ctx context.Context) (string, error) {
	type result struct {
		Exists bool `gorm:"column:exists"`
	}
	checkColumn := func(columnName string) (bool, error) {
		var row result
		err := u.db.WithContext(ctx).Raw(`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = current_schema()
				  AND table_name = 'tenant_subscriptions'
				  AND column_name = ?
			) AS exists
		`, columnName).Scan(&row).Error
		if err != nil {
			return false, err
		}
		return row.Exists, nil
	}

	if ok, err := checkColumn("amount_paid_idr"); err != nil {
		return "", err
	} else if ok {
		return "amount_paid_idr", nil
	}

	if ok, err := checkColumn("amount_idr"); err != nil {
		return "", err
	} else if ok {
		return "amount_idr", nil
	}

	return "", nil
}

func (u *authUsecase) paymentTransactionsSupportsProviderInvoiceOnConflict(ctx context.Context) (bool, error) {
	type result struct {
		Exists bool `gorm:"column:exists"`
	}
	var row result
	err := u.db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
			  ON tc.constraint_name = kcu.constraint_name
			 AND tc.table_schema = kcu.table_schema
			WHERE tc.table_schema = current_schema()
			  AND tc.table_name = 'payment_transactions'
			  AND tc.constraint_type = 'UNIQUE'
			GROUP BY tc.constraint_name
			HAVING BOOL_OR(kcu.column_name = 'provider')
			   AND BOOL_OR(kcu.column_name = 'provider_invoice_id')
		)
	`).Scan(&row).Error
	if err != nil {
		return false, err
	}
	return row.Exists, nil
}

func paymentTransactionExists(tx *gorm.DB, provider models.PaymentProvider, providerInvoiceID string) (bool, error) {
	var row struct {
		ID string `gorm:"column:id"`
	}
	err := tx.Table("payment_transactions").
		Select("id").
		Where("provider = ? AND provider_invoice_id = ?", provider, providerInvoiceID).
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(row.ID) != "", nil
}

func (u *authUsecase) storePendingBillingChange(ctx context.Context, pending PendingBillingChange) error {
	if u.redis == nil {
		return ErrPaymentGatewayUnavailable
	}
	encoded, err := json.Marshal(pending)
	if err != nil {
		return err
	}
	return u.redis.Set(ctx, pendingBillingChangeKeyPrefix+pending.Token, string(encoded), pendingRegTTL).Err()
}

func (u *authUsecase) ConfirmPendingBillingChange(ctx context.Context, token, tenantID string) error {
	if u.redis == nil {
		return ErrPendingBillingChangeNotFound
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrPendingBillingChangeInvalid
	}

	key := pendingBillingChangeKeyPrefix + token
	val, err := u.redis.Get(ctx, key).Result()
	if err != nil {
		return ErrPendingBillingChangeNotFound
	}

	var pending PendingBillingChange
	if err := json.Unmarshal([]byte(val), &pending); err != nil {
		return ErrPendingBillingChangeInvalid
	}
	if strings.TrimSpace(pending.TenantID) == "" || pending.TenantID != tenantID {
		return ErrPendingBillingChangeNotFound
	}

	return u.CompletePendingBillingChange(ctx, token)
}

func (u *authUsecase) countActiveTenantUsers(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := u.db.WithContext(ctx).
		Table("users").
		Where("tenant_id = ? AND status = ? AND deleted_at IS NULL", tenantID, "active").
		Count(&count).Error
	return count, err
}

func (u *authUsecase) resolveTenantOwnerEmail(ctx context.Context, tenantID string) (string, string, error) {
	type ownerRow struct {
		Email string `gorm:"column:email"`
		Name  string `gorm:"column:name"`
	}
	var row ownerRow
	err := u.db.WithContext(ctx).
		Table("tenants").
		Select("COALESCE(u.email, '') AS email, COALESCE(tenants.name, '') AS name").
		Joins("LEFT JOIN users u ON u.id = tenants.owner_user_id AND u.deleted_at IS NULL").
		Where("tenants.id = ? AND tenants.deleted_at IS NULL", tenantID).
		Scan(&row).Error
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(row.Email) == "" {
		var fallback struct {
			Email string `gorm:"column:email"`
		}
		if err := u.db.WithContext(ctx).
			Table("users").
			Select("email").
			Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
			Order("created_at ASC").
			Limit(1).
			Scan(&fallback).Error; err == nil {
			row.Email = fallback.Email
		}
	}
	return strings.TrimSpace(row.Email), strings.TrimSpace(row.Name), nil
}

func deriveBillingIdempotencyKey(req *tenantDTO.BillingChangeRequest, snapshot *billingSubscriptionSnapshot) string {
	if key := strings.TrimSpace(req.IdempotencyKey); key != "" {
		return key
	}
	planToken := strings.TrimSpace(req.XenditPlanID)
	if planToken == "" {
		planToken = snapshot.PlanSlug
	}
	return fmt.Sprintf("%s-%s-%s", planToken, req.Action, req.ActionDate)
}

func calculateProrationAmount(actionDate time.Time, nextBillingAt, expiresAt *time.Time, oldAmount, newAmount int64) (int64, string) {
	if newAmount <= oldAmount {
		return 0, ""
	}

	billingEnd := nextBillingAt
	if billingEnd == nil {
		billingEnd = expiresAt
	}
	if billingEnd == nil || billingEnd.Before(actionDate) {
		return 0, "LESS_THAN_3_DAYS"
	}

	remainingDays := int(math.Ceil(billingEnd.Sub(actionDate).Hours() / 24))
	if remainingDays < 3 {
		return 0, "LESS_THAN_3_DAYS"
	}

	daysInMonth := daysInMonth(actionDate)
	if daysInMonth <= 0 {
		daysInMonth = 30
	}
	delta := float64(newAmount - oldAmount)
	proration := int64(math.Round((delta / float64(daysInMonth)) * float64(remainingDays)))
	if proration < 0 {
		proration = 0
	}
	return proration, ""
}

func daysInMonth(t time.Time) int {
	firstOfNextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	lastOfCurrentMonth := firstOfNextMonth.Add(-24 * time.Hour)
	return lastOfCurrentMonth.Day()
}

func couponIDIfAny(code string, db *gorm.DB, ctx context.Context) string {
	if strings.TrimSpace(code) == "" {
		return ""
	}
	var coupon models.Coupon
	if err := loadCouponByCodeDB(ctx, db, code, &coupon); err != nil {
		return ""
	}
	return coupon.ID
}

func loadCouponByCode(tx *gorm.DB, code string) (*models.Coupon, error) {
	var coupon models.Coupon
	if err := loadCouponByCodeDB(context.Background(), tx, code, &coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

func loadCouponByCodeDB(ctx context.Context, db *gorm.DB, code string, coupon *models.Coupon) error {
	return db.WithContext(ctx).Where("code = ? AND deleted_at IS NULL", strings.ToUpper(strings.TrimSpace(code))).First(coupon).Error
}

func markCouponUsed(tx *gorm.DB, couponID string) error {
	return tx.Model(&models.Coupon{}).Where("id = ?", couponID).
		Updates(map[string]any{"used_count": gorm.Expr("used_count + 1")}).Error
}

func applyCouponToRecurringAmount(coupon *models.Coupon, baseAmount int64, billingPeriod string) int64 {
	if coupon == nil {
		return baseAmount
	}

	switch coupon.DiscountType {
	case models.CouponDiscountTrial:
		return 0
	case models.CouponDiscountPercent:
		if coupon.DiscountValue <= 0 {
			return baseAmount
		}
		discount := int64(math.Round(float64(baseAmount) * coupon.DiscountValue / 100))
		if final := baseAmount - discount; final >= 0 {
			return final
		}
		return 0
	case models.CouponDiscountAmount:
		discount := int64(coupon.DiscountValue)
		if billingPeriod == string(models.BillingYearly) {
			discount *= 12
		}
		if final := baseAmount - discount; final >= 0 {
			return final
		}
		return 0
	default:
		return baseAmount
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

// CancelSubscription cancels the active subscription for a tenant.
// If a Xendit recurring plan is linked, it is deactivated to stop future charges.
// The subscription status is set to "cancelled" but ExpiresAt is preserved so
// access continues until the end of the already-paid billing period.
func (u *authUsecase) CancelSubscription(ctx context.Context, tenantID string) error {
	var sub models.TenantSubscription
	err := u.db.WithContext(ctx).
		Where("tenant_id = ? AND status IN ? AND deleted_at IS NULL", tenantID,
			[]string{string(models.SubscriptionActive), string(models.SubscriptionTrial)}).
		Order("starts_at DESC").
		First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSubscriptionNotFound
		}
		return fmt.Errorf("cancel subscription: lookup failed: %w", err)
	}

	// Deactivate the Xendit recurring plan when present to stop future auto-charges.
	if sub.XenditSubscriptionID != nil && *sub.XenditSubscriptionID != "" {
		if u.xendit != nil && u.xendit.IsConfigured() {
			if _, xenditErr := u.xendit.CancelRecurringPlan(ctx, *sub.XenditSubscriptionID); xenditErr != nil {
				// Log and continue — the DB record must still be updated even if
				// the Xendit call fails (e.g. plan already inactive on their side).
				fmt.Printf("[cancel_subscription] xendit deactivate plan=%s tenant=%s err=%v\n",
					*sub.XenditSubscriptionID, tenantID, xenditErr)
			}
		}
	}

	updates := map[string]any{
		"status":                 models.SubscriptionCancelled,
		"xendit_subscription_id": nil,
		"next_billing_at":        nil,
	}
	if dbErr := u.db.WithContext(ctx).
		Model(&models.TenantSubscription{}).
		Where("id = ?", sub.ID).
		Updates(updates).Error; dbErr != nil {
		return fmt.Errorf("cancel subscription: update failed: %w", dbErr)
	}

	return nil
}
