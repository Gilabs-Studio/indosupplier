package handler

import (
	"context"
	stderrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/response"
	permissionModels "github.com/gilabs/gims/api/internal/permission/data/models"
	permissionRepos "github.com/gilabs/gims/api/internal/permission/data/repositories"
	tenantModels "github.com/gilabs/gims/api/internal/tenant/data/models"
	"github.com/gilabs/gims/api/internal/tenant/domain/dto"
	"github.com/gilabs/gims/api/internal/tenant/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func resolveSystemAdminCookieOptions() (string, http.SameSite, bool) {
	if config.AppConfig == nil {
		return "", http.SameSiteLaxMode, true
	}

	env := strings.ToLower(strings.TrimSpace(config.AppConfig.Server.Env))
	if env == "production" || env == "prod" {
		return config.AppConfig.Server.RootDomain, http.SameSiteNoneMode, true
	}

	return "", http.SameSiteLaxMode, true
}

func setSystemAdminAccessTokenCookie(c *gin.Context, token string, maxAge int) {
	domain, sameSite, secure := resolveSystemAdminCookieOptions()

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "gims_sys_access_token",
		Value:    token,
		Path:     "/internal",
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
}

func expireSystemAdminAccessTokenCookie(c *gin.Context, path string, domain string) {
	_, sameSite, secure := resolveSystemAdminCookieOptions()

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "gims_sys_access_token",
		Value:    "",
		Path:     path,
		Domain:   domain,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
}

func clearSystemAdminAccessTokenCookies(c *gin.Context) {
	domain, _, _ := resolveSystemAdminCookieOptions()
	paths := []string{"/internal", "/"}
	domains := []string{"", domain}

	for _, p := range paths {
		for _, d := range domains {
			expireSystemAdminAccessTokenCookie(c, p, d)
		}
	}
}

// SystemAdminHandler handles system admin HTTP requests.
type SystemAdminHandler struct {
	sysAdminUC     usecase.SystemAdminUsecase
	tenantUC       usecase.TenantUsecase
	couponUC       usecase.CouponUsecase
	subUC          usecase.SubscriptionUsecase
	planUC         usecase.SubscriptionPlanUsecase
	permissionRepo permissionRepos.PermissionRepository
}

type dashboardBusinessHealth struct {
	MRRIDR                 int64   `json:"mrr_idr"`
	ARRIDR                 int64   `json:"arr_idr"`
	ChurnRatePercent       float64 `json:"churn_rate_percent"`
	NetRevenueRetentionPct float64 `json:"net_revenue_retention_percent"`
	TotalTenantAktif       int64   `json:"total_tenant_aktif"`
	TotalSeatTerjual       int64   `json:"total_seat_terjual"`
	TrialAktif             int64   `json:"trial_aktif"`
	OverdueInvoice         int64   `json:"overdue_invoice"`
	SuspendedTenant        int64   `json:"suspended_tenant"`
	DeltaTenantBulanIni    int64   `json:"delta_tenant_bulan_ini"`
}

type dashboardPlanMixRow struct {
	Plan         string  `json:"plan"`
	Label        string  `json:"label"`
	TenantCount  int64   `json:"tenant_count"`
	SeatCount    int64   `json:"seat_count"`
	SharePercent float64 `json:"share_percent"`
}

type dashboardBillingActivity struct {
	SuccessCount  int64 `json:"success_count"`
	FailedRetry   int64 `json:"failed_retry"`
	RefundProcess int64 `json:"refund_process"`
	UpgradePlan   int64 `json:"upgrade_plan"`
	AddSeat       int64 `json:"add_seat"`
}

type dashboardSupportSummary struct {
	OpenTickets     int64   `json:"open_tickets"`
	SLABreach       int64   `json:"sla_breach"`
	CSATAverage     float64 `json:"csat_average"`
	ActiveIncidents int64   `json:"active_incidents"`
	TodoNote        string  `json:"todo_note"`
	IsTodo          bool    `json:"is_todo"`
}

type dashboardTenantRow struct {
	TenantName string    `json:"tenant_name"`
	Plan       string    `json:"plan"`
	Seat       int       `json:"seat"`
	Status     string    `json:"status"`
	MRRIDR     int64     `json:"mrr_idr"`
	JoinedAt   time.Time `json:"joined_at"`
}

type dashboardInfrastructureSummary struct {
	Uptime30DPercent float64 `json:"uptime_30d_percent"`
	APILatencyAvgMs  float64 `json:"api_latency_avg_ms"`
	ErrorRatePercent float64 `json:"error_rate_percent"`
	DBSizeBytes      int64   `json:"db_size_bytes"`
	DBSizeLabel      string  `json:"db_size_label"`
	ObservationNote  string  `json:"observation_note"`
}

type dashboardPaymentRow struct {
	ID              string     `json:"id"`
	TenantName      string     `json:"tenant_name"`
	Plan            string     `json:"plan"`
	BillingPeriod   string     `json:"billing_period"`
	AmountPaidIDR   int64      `json:"amount_paid_idr"`
	CouponCode      string     `json:"coupon_code,omitempty"`
	XenditInvoiceID string     `json:"xendit_invoice_id,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	StartsAt        time.Time  `json:"starts_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// NewSystemAdminHandler creates a new SystemAdminHandler.
func NewSystemAdminHandler(
	sysAdminUC usecase.SystemAdminUsecase,
	tenantUC usecase.TenantUsecase,
	couponUC usecase.CouponUsecase,
	subUC usecase.SubscriptionUsecase,
	planUC usecase.SubscriptionPlanUsecase,
	permissionRepo permissionRepos.PermissionRepository,
) *SystemAdminHandler {
	return &SystemAdminHandler{
		sysAdminUC:     sysAdminUC,
		tenantUC:       tenantUC,
		couponUC:       couponUC,
		subUC:          subUC,
		planUC:         planUC,
		permissionRepo: permissionRepo,
	}
}

// ─── Permission Management ────────────────────────────────────────────────────

type sysAdminPermissionRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=255"`
	Code        string  `json:"code" binding:"required"`
	Action      string  `json:"action" binding:"required"`
	Resource    string  `json:"resource"`
	Description string  `json:"description"`
	MenuID      *string `json:"menu_id"`
}

type sysAdminPermissionMenuOption struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	URL      string  `json:"url"`
	ParentID *string `json:"parent_id,omitempty"`
	Depth    int     `json:"depth"`
	Label    string  `json:"label"`
}

// ListPermissions returns a paginated list of all permissions.
func (h *SystemAdminHandler) ListPermissions(c *gin.Context) {
	page := 1
	limit := 50
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	search := strings.TrimSpace(c.Query("search"))

	perms, total, err := h.permissionRepo.ListPaginated(c.Request.Context(), page, limit, search)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{
		Pagination: response.NewPaginationMeta(page, limit, int(total)),
	}

	response.SuccessResponse(c, perms, meta)
}

// ListPermissionMenus returns flattened menu options for permission assignment.
func (h *SystemAdminHandler) ListPermissionMenus(c *gin.Context) {
	menus, err := h.permissionRepo.GetRootMenusWithChildren(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	options := make([]sysAdminPermissionMenuOption, 0, len(menus))

	var flatten func(items []permissionModels.Menu, depth int, parents []string)
	flatten = func(items []permissionModels.Menu, depth int, parents []string) {
		for _, menu := range items {
			path := append(parents, menu.Name)
			options = append(options, sysAdminPermissionMenuOption{
				ID:       menu.ID,
				Name:     menu.Name,
				URL:      menu.URL,
				ParentID: menu.ParentID,
				Depth:    depth,
				Label:    strings.Join(path, " / "),
			})

			if len(menu.Children) > 0 {
				flatten(menu.Children, depth+1, path)
			}
		}
	}

	flatten(menus, 0, nil)
	response.SuccessResponse(c, options, nil)
}

// GetPermission returns a single permission by ID.
func (h *SystemAdminHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")
	p, err := h.permissionRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			errors.NotFoundResponse(c, "permission", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, p, nil)
}

// CreatePermission creates a new permission record.
func (h *SystemAdminHandler) CreatePermission(c *gin.Context) {
	var req sysAdminPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, ve)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	// Prevent duplicate codes.
	if existing, _ := h.permissionRepo.FindByCode(c.Request.Context(), req.Code); existing != nil {
		errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{"field": "code", "message": "permission code already exists"}, nil)
		return
	}

	p := &permissionModels.Permission{
		Name:        req.Name,
		Code:        req.Code,
		Action:      req.Action,
		Resource:    req.Resource,
		Description: req.Description,
		MenuID:      req.MenuID,
	}

	if err := h.permissionRepo.Create(c.Request.Context(), p); err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseCreated(c, p, nil)
}

// UpdatePermission modifies an existing permission.
func (h *SystemAdminHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")

	p, err := h.permissionRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			errors.NotFoundResponse(c, "permission", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	var req sysAdminPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, ve)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	// Guard code uniqueness when code is changed.
	if req.Code != p.Code {
		if existing, _ := h.permissionRepo.FindByCode(c.Request.Context(), req.Code); existing != nil {
			errors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{"field": "code", "message": "permission code already exists"}, nil)
			return
		}
	}

	p.Name = req.Name
	p.Code = req.Code
	p.Action = req.Action
	p.Resource = req.Resource
	p.Description = req.Description
	p.MenuID = req.MenuID

	if err := h.permissionRepo.Update(c.Request.Context(), p); err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, p, nil)
}

// DeletePermission soft-deletes a permission by ID.
func (h *SystemAdminHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")

	if _, err := h.permissionRepo.FindByID(c.Request.Context(), id); err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			errors.NotFoundResponse(c, "permission", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	if err := h.permissionRepo.Delete(c.Request.Context(), id); err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseDeleted(c, "permission", id, nil)
}

// Login handles system admin login via /internal/sys-login
func (h *SystemAdminHandler) Login(c *gin.Context) {
	var req dto.SystemAdminLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	loginResp, err := h.sysAdminUC.Login(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrSysAdminInvalidCredentials {
			errors.ErrorResponse(c, "INVALID_CREDENTIALS", nil, nil)
			return
		}
		if err == usecase.ErrSysAdminDisabled {
			errors.ErrorResponse(c, "ACCOUNT_DISABLED", map[string]interface{}{
				"reason": "System admin account is disabled",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	clearSystemAdminAccessTokenCookies(c)
	setSystemAdminAccessTokenCookie(c, loginResp.AccessToken, config.AppConfig.JWT.AccessTokenTTL*3600)
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")

	// Clear access token from response body
	loginResp.AccessToken = ""
	loginResp.RefreshToken = ""

	response.SuccessResponse(c, loginResp, nil)
}

// Logout handles system admin logout
func (h *SystemAdminHandler) Logout(c *gin.Context) {
	clearSystemAdminAccessTokenCookies(c)
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")

	response.SuccessResponse(c, gin.H{"message": "logged out"}, nil)
}

// Me returns the current system admin info from the JWT session
func (h *SystemAdminHandler) Me(c *gin.Context) {
	adminID, _ := c.Get("user_id")
	adminEmail, _ := c.Get("user_email")
	adminName, _ := c.Get("user_name")

	id, _ := adminID.(string)
	email, _ := adminEmail.(string)
	name, _ := adminName.(string)
	if strings.TrimSpace(name) == "" {
		name = email
	}

	response.SuccessResponse(c, gin.H{
		"id":       id,
		"email":    email,
		"name":     name,
		"username": name,
		"role":     "system_admin",
	}, nil)
}

// UpdateProfile updates system admin username/email.
func (h *SystemAdminHandler) UpdateProfile(c *gin.Context) {
	adminID, _ := c.Get("user_id")
	id, _ := adminID.(string)
	if strings.TrimSpace(id) == "" {
		errors.UnauthorizedResponse(c, "token missing")
		return
	}

	var req dto.SystemAdminUpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	updated, err := h.sysAdminUC.UpdateProfile(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case usecase.ErrSysAdminNotFound:
			errors.NotFoundResponse(c, "system admin", id)
		case usecase.ErrSysAdminEmailAlreadyTaken:
			errors.ErrorResponse(c, "EMAIL_ALREADY_TAKEN", nil, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	response.SuccessResponse(c, updated, nil)
}

// ChangePassword changes the current system admin password.
func (h *SystemAdminHandler) ChangePassword(c *gin.Context) {
	adminID, _ := c.Get("user_id")
	id, _ := adminID.(string)
	if strings.TrimSpace(id) == "" {
		errors.UnauthorizedResponse(c, "token missing")
		return
	}

	var req dto.SystemAdminChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	if strings.TrimSpace(req.CurrentPassword) == strings.TrimSpace(req.NewPassword) {
		errors.ErrorResponse(c, "VALIDATION_ERROR", nil, []response.FieldError{
			{
				Field:   "new_password",
				Code:    "SAME_AS_CURRENT",
				Message: "new password must be different from current password",
			},
		})
		return
	}

	err := h.sysAdminUC.ChangePassword(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case usecase.ErrSysAdminNotFound:
			errors.NotFoundResponse(c, "system admin", id)
		case usecase.ErrSysAdminInvalidCredentials:
			errors.ErrorResponse(c, "INVALID_CREDENTIALS", nil, []response.FieldError{
				{
					Field:   "current_password",
					Code:    "INVALID_CURRENT_PASSWORD",
					Message: "current password is invalid",
				},
			})
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	response.SuccessResponse(c, gin.H{"message": "password updated"}, nil)
}

// ListTenants returns all tenants with pagination and search (system admin only)
func (h *SystemAdminHandler) ListTenants(c *gin.Context) {
	var params dto.TenantListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		errors.InvalidRequestBodyResponse(c)
		return
	}

	tenants, pagination, err := h.tenantUC.ListTenantsPaginated(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: pagination}
	response.SuccessResponse(c, tenants, meta)
}

// GetTenantDetail returns detail for a single tenant (system admin only)
func (h *SystemAdminHandler) GetTenantDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"id": "required"}, nil)
		return
	}

	tenant, err := h.tenantUC.GetTenantDetail(c.Request.Context(), id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			errors.NotFoundResponse(c, "tenant", id)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, tenant, nil)
}

// RecoverTenantDeletion cancels a pending tenant deletion request (system admin only).
func (h *SystemAdminHandler) RecoverTenantDeletion(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"id": "required"}, nil)
		return
	}

	tenant, err := h.tenantUC.RecoverTenantDeletion(c.Request.Context(), id)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			errors.NotFoundResponse(c, "tenant", id)
		case usecase.ErrTenantDeletionNotScheduled:
			errors.ErrorResponse(c, "CONFLICT", map[string]interface{}{
				"reason": "tenant deletion is not scheduled",
			}, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	response.SuccessResponse(c, tenant, nil)
}

// Dashboard returns system admin dashboard summary
func (h *SystemAdminHandler) Dashboard(c *gin.Context) {
	tenants, err := h.tenantUC.ListTenants(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	businessHealth, planMix, billingActivity, supportSummary, tenantOverview, infrastructure := h.loadDashboardSummary(c.Request.Context(), tenants)

	response.SuccessResponse(c, gin.H{
		"total_tenants":              len(tenants),
		"tenants":                    tenants,
		"business_health":            businessHealth,
		"plan_mix":                   planMix,
		"billing_activity":           billingActivity,
		"support_incidents":          supportSummary,
		"tenant_overview":            tenantOverview,
		"infrastructure_performance": infrastructure,
	}, nil)
}

func (h *SystemAdminHandler) loadDashboardSummary(ctx context.Context, tenants []dto.TenantListResponse) (dashboardBusinessHealth, []dashboardPlanMixRow, dashboardBillingActivity, dashboardSupportSummary, []dashboardTenantRow, dashboardInfrastructureSummary) {
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	previousThirtyDaysAgo := now.AddDate(0, 0, -60)

	businessHealth := h.loadBusinessHealth(ctx, tenants, thirtyDaysAgo, previousThirtyDaysAgo, now)
	planMix := h.loadPlanMix(ctx)
	billingActivity := h.loadBillingActivity(ctx, thirtyDaysAgo)
	supportSummary := dashboardSupportSummary{
		OpenTickets:     24,
		SLABreach:       2,
		CSATAverage:     4.6,
		ActiveIncidents: 1,
		TodoNote:        "TODO: sambungkan data helpdesk dan incident tracker ke dashboard ini.",
		IsTodo:          true,
	}
	tenantOverview := h.loadTenantOverview(ctx, tenants)
	infrastructure := h.loadInfrastructureSummary(ctx)

	return businessHealth, planMix, billingActivity, supportSummary, tenantOverview, infrastructure
}

func (h *SystemAdminHandler) loadBusinessHealth(ctx context.Context, tenants []dto.TenantListResponse, thirtyDaysAgo, previousThirtyDaysAgo, now time.Time) dashboardBusinessHealth {
	type subscriptionAgg struct {
		TotalPaidCurrent  int64 `gorm:"column:total_paid_current"`
		TotalPaidPrevious int64 `gorm:"column:total_paid_previous"`
		ActiveCount       int64 `gorm:"column:active_count"`
		TrialCount        int64 `gorm:"column:trial_count"`
		OverdueCount      int64 `gorm:"column:overdue_count"`
		SeatCount         int64 `gorm:"column:seat_count"`
		ExpiredCount      int64 `gorm:"column:expired_count"`
	}

	var subscriptionSummary subscriptionAgg
	_ = database.DB.WithContext(ctx).
		Table("tenant_subscriptions").
		Select(`
			COALESCE(SUM(CASE WHEN created_at >= ? AND status = 'active' THEN amount_paid_idr ELSE 0 END), 0) AS total_paid_current,
			COALESCE(SUM(CASE WHEN created_at >= ? AND created_at < ? AND status = 'active' THEN amount_paid_idr ELSE 0 END), 0) AS total_paid_previous,
			COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0) AS active_count,
			COALESCE(SUM(CASE WHEN status = 'trial' THEN 1 ELSE 0 END), 0) AS trial_count,
			COALESCE(SUM(CASE WHEN status = 'active' AND expires_at IS NOT NULL AND expires_at < ? THEN 1 ELSE 0 END), 0) AS overdue_count,
			COALESCE(SUM(CASE WHEN status IN ('active', 'trial') THEN user_count ELSE 0 END), 0) AS seat_count,
			COALESCE(SUM(CASE WHEN status = 'expired' THEN 1 ELSE 0 END), 0) AS expired_count
		`, thirtyDaysAgo, previousThirtyDaysAgo, thirtyDaysAgo, now).
		Where("deleted_at IS NULL").
		Scan(&subscriptionSummary).Error

	activeTenants := int64(0)
	trialTenants := int64(0)
	suspendedTenants := int64(0)
	for _, tenant := range tenants {
		switch strings.ToLower(strings.TrimSpace(tenant.Status)) {
		case "active":
			activeTenants++
		case "trial":
			trialTenants++
		case "suspended":
			suspendedTenants++
		}
	}

	var payments30d struct {
		CurrentRevenue  int64 `gorm:"column:current_revenue"`
		PreviousRevenue int64 `gorm:"column:previous_revenue"`
	}
	_ = database.DB.WithContext(ctx).
		Table("payment_transactions").
		Select(`
			COALESCE(SUM(CASE WHEN status = 'paid' AND created_at >= ? THEN amount_idr ELSE 0 END), 0) AS current_revenue,
			COALESCE(SUM(CASE WHEN status = 'paid' AND created_at >= ? AND created_at < ? THEN amount_idr ELSE 0 END), 0) AS previous_revenue
		`, thirtyDaysAgo, previousThirtyDaysAgo, thirtyDaysAgo).
		Where("deleted_at IS NULL").
		Scan(&payments30d).Error

	mrr := payments30d.CurrentRevenue
	if mrr == 0 {
		mrr = subscriptionSummary.TotalPaidCurrent
	}
	arr := mrr * 12
	nrr := float64(100)
	if payments30d.PreviousRevenue > 0 {
		nrr = float64(payments30d.CurrentRevenue) * 100 / float64(payments30d.PreviousRevenue)
	}
	churn := float64(0)
	if activeTenants > 0 {
		churn = float64(subscriptionSummary.ExpiredCount) * 100 / float64(activeTenants)
	}

	planDelta := activeTenants - subscriptionSummary.ActiveCount
	if planDelta < 0 {
		planDelta = 0
	}

	return dashboardBusinessHealth{
		MRRIDR:                 mrr,
		ARRIDR:                 arr,
		ChurnRatePercent:       churn,
		NetRevenueRetentionPct: nrr,
		TotalTenantAktif:       activeTenants,
		TotalSeatTerjual:       subscriptionSummary.SeatCount,
		TrialAktif:             trialTenants,
		OverdueInvoice:         subscriptionSummary.OverdueCount,
		SuspendedTenant:        suspendedTenants,
		DeltaTenantBulanIni:    planDelta,
	}
}

func (h *SystemAdminHandler) loadPlanMix(ctx context.Context) []dashboardPlanMixRow {
	type planRow struct {
		Plan        string `gorm:"column:plan"`
		TenantCount int64  `gorm:"column:tenant_count"`
		SeatCount   int64  `gorm:"column:seat_count"`
	}

	var rows []planRow
	_ = database.DB.WithContext(ctx).
		Table("tenant_subscriptions").
		Select("plan, COUNT(DISTINCT tenant_id) AS tenant_count, COALESCE(SUM(user_count), 0) AS seat_count").
		Where("deleted_at IS NULL AND status IN ?", []tenantModels.SubscriptionStatus{tenantModels.SubscriptionActive, tenantModels.SubscriptionTrial}).
		Group("plan").
		Order("tenant_count DESC, seat_count DESC").
		Scan(&rows).Error

	total := int64(0)
	for _, row := range rows {
		total += row.TenantCount
	}

	result := make([]dashboardPlanMixRow, 0, len(rows))
	for _, row := range rows {
		share := 0.0
		if total > 0 {
			share = float64(row.TenantCount) * 100 / float64(total)
		}
		result = append(result, dashboardPlanMixRow{
			Plan:         row.Plan,
			Label:        humanizePlanLabel(row.Plan),
			TenantCount:  row.TenantCount,
			SeatCount:    row.SeatCount,
			SharePercent: share,
		})
	}

	return result
}

func (h *SystemAdminHandler) loadBillingActivity(ctx context.Context, since time.Time) dashboardBillingActivity {
	type billingRow struct {
		SuccessCount  int64 `gorm:"column:success_count"`
		FailedRetry   int64 `gorm:"column:failed_retry"`
		RefundProcess int64 `gorm:"column:refund_process"`
		UpgradePlan   int64 `gorm:"column:upgrade_plan"`
		AddSeat       int64 `gorm:"column:add_seat"`
	}

	var billing billingRow
	if err := database.DB.WithContext(ctx).
		Table("payment_transactions").
		Select(`
			COALESCE(SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END), 0) AS success_count,
			COALESCE(SUM(CASE WHEN status IN ('failed', 'expired') THEN 1 ELSE 0 END), 0) AS failed_retry,
			COALESCE(SUM(CASE WHEN status = 'canceled' THEN 1 ELSE 0 END), 0) AS refund_process,
			COALESCE(SUM(CASE WHEN LOWER(COALESCE(description, '')) LIKE '%upgrade%' OR LOWER(COALESCE(notes, '')) LIKE '%upgrade%' THEN 1 ELSE 0 END), 0) AS upgrade_plan,
			COALESCE(SUM(CASE WHEN LOWER(COALESCE(description, '')) LIKE '%seat%' OR LOWER(COALESCE(notes, '')) LIKE '%seat%' THEN 1 ELSE 0 END), 0) AS add_seat
		`).
		Where("deleted_at IS NULL AND created_at >= ?", since).
		Scan(&billing).Error; err != nil {
		// Non-fatal: return zeros on error
		return dashboardBillingActivity{}
	}

	return dashboardBillingActivity{
		SuccessCount:  billing.SuccessCount,
		FailedRetry:   billing.FailedRetry,
		RefundProcess: billing.RefundProcess,
		UpgradePlan:   billing.UpgradePlan,
		AddSeat:       billing.AddSeat,
	}
}

func (h *SystemAdminHandler) loadTenantOverview(ctx context.Context, tenants []dto.TenantListResponse) []dashboardTenantRow {
	rows := make([]dashboardTenantRow, 0, minInt(len(tenants), 4))
	for _, tenant := range tenants {
		joinedAt := time.Now()
		if tenant.CreatedAt != nil {
			if parsed, err := time.Parse(time.RFC3339, *tenant.CreatedAt); err == nil {
				joinedAt = parsed
			}
		}

		// Prefer active subscription plan over tenants.plan (which may be stale 'trial')
		plan := tenant.Plan
		seat := tenant.MaxUsers
		var sub struct {
			Plan      string `gorm:"column:plan"`
			SeatLimit int    `gorm:"column:seat_limit"`
		}
		if err := database.DB.WithContext(ctx).
			Table("tenant_subscriptions").
			Select("plan, seat_limit").
			Where("tenant_id = ? AND status IN ('active','trial') AND deleted_at IS NULL", tenant.ID).
			Order("created_at DESC").
			Limit(1).
			Scan(&sub).Error; err == nil && sub.Plan != "" {
			plan = sub.Plan
			if sub.SeatLimit > 0 {
				seat = sub.SeatLimit
			}
		}

		rows = append(rows, dashboardTenantRow{
			TenantName: tenant.Name,
			Plan:       plan,
			Seat:       seat,
			Status:     tenant.Status,
			JoinedAt:   joinedAt,
		})
		if len(rows) >= 4 {
			break
		}
	}

	return rows
}

func (h *SystemAdminHandler) loadInfrastructureSummary(ctx context.Context) dashboardInfrastructureSummary {
	metrics := middleware.MetricsSnapshot()
	var dbSizeBytes int64
	_ = database.DB.WithContext(ctx).Raw("SELECT COALESCE(pg_database_size(current_database()), 0)").Scan(&dbSizeBytes).Error

	return dashboardInfrastructureSummary{
		Uptime30DPercent: uptimePercentFromMetrics(metrics),
		APILatencyAvgMs:  metrics.AvgDurationMs,
		ErrorRatePercent: errorRatePercentFromMetrics(metrics),
		DBSizeBytes:      dbSizeBytes,
		DBSizeLabel:      humanizeBytes(dbSizeBytes),
		ObservationNote:  "Live metrics from request middleware and PostgreSQL storage size.",
	}
}

func humanizePlanLabel(plan string) string {
	if plan == "" {
		return "Unknown"
	}
	label := strings.ReplaceAll(plan, "_", " ")
	parts := strings.Fields(label)
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func humanizeBytes(size int64) string {
	if size <= 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	unitIndex := 0
	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}
	return fmt.Sprintf("%.2f %s", value, units[unitIndex])
}

func uptimePercentFromMetrics(metrics middleware.MetricsSnapshotData) float64 {
	if metrics.TotalRequests == 0 {
		return 100
	}
	uptime := 100 - errorRatePercentFromMetrics(metrics)
	if uptime < 0 {
		return 0
	}
	return uptime
}

func errorRatePercentFromMetrics(metrics middleware.MetricsSnapshotData) float64 {
	if metrics.TotalRequests == 0 {
		return 0
	}
	return float64(metrics.TotalErrors5xx) * 100 / float64(metrics.TotalRequests)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ─── Coupon endpoints ────────────────────────────────────────────────────────

// CreateCoupon generates a new coupon (system admin only).
func (h *SystemAdminHandler) CreateCoupon(c *gin.Context) {
	var req dto.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, ve)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	adminEmail, _ := c.Get("user_email")
	email, _ := adminEmail.(string)

	coupon, err := h.couponUC.Create(c.Request.Context(), &req, email)
	if err != nil {
		if err == usecase.ErrCouponDuplicate {
			errors.ErrorResponse(c, "COUPON_CODE_EXISTS", map[string]interface{}{
				"field": "code",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, coupon, nil)
}

// ListCoupons returns a paginated list of coupons (system admin only).
func (h *SystemAdminHandler) ListCoupons(c *gin.Context) {
	var params dto.CouponListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		errors.InvalidRequestBodyResponse(c)
		return
	}

	coupons, pagination, err := h.couponUC.List(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: pagination}
	response.SuccessResponse(c, coupons, meta)
}

// UpdateCoupon updates an existing coupon (system admin only).
func (h *SystemAdminHandler) UpdateCoupon(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"id": "required"}, nil)
		return
	}

	var req dto.UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, ve)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	coupon, err := h.couponUC.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == usecase.ErrCouponNotFound {
			errors.ErrorResponse(c, "COUPON_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, coupon, nil)
}

// SetCouponStatus activates or deactivates a coupon (system admin only).
func (h *SystemAdminHandler) SetCouponStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"id": "required"}, nil)
		return
	}

	var req dto.SetCouponStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.InvalidRequestBodyResponse(c)
		return
	}

	if err := h.couponUC.SetActive(c.Request.Context(), id, req.IsActive); err != nil {
		if err == usecase.ErrCouponNotFound {
			errors.ErrorResponse(c, "COUPON_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponseNoContent(c)
}

// ─── Subscription endpoints ──────────────────────────────────────────────────

// ListSubscriptions returns all tenant subscriptions (system admin only).
func (h *SystemAdminHandler) ListSubscriptions(c *gin.Context) {
	var params dto.SubscriptionListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		errors.InvalidRequestBodyResponse(c)
		return
	}

	subs, pagination, err := h.subUC.ListAll(c.Request.Context(), params)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	meta := &response.Meta{Pagination: pagination}
	response.SuccessResponse(c, subs, meta)
}

// GetTenantSubscription returns the active subscription for a specific tenant (system admin only).
func (h *SystemAdminHandler) GetTenantSubscription(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"tenant_id": "required"}, nil)
		return
	}

	sub, err := h.subUC.GetActiveByTenant(c.Request.Context(), tenantID)
	if err != nil {
		if err == usecase.ErrSubscriptionNotFound {
			errors.ErrorResponse(c, "SUBSCRIPTION_NOT_FOUND", nil, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, sub, nil)
}

// ─── Subscription Plan endpoints ─────────────────────────────────────────────

// ListPlans returns all subscription plan configs (system admin only).
func (h *SystemAdminHandler) ListPlans(c *gin.Context) {
	page := 1
	perPage := 100
	plans, _, err := h.planUC.ListAll(c.Request.Context(), page, perPage)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, plans, nil)
}

// UpsertPlan creates or updates a subscription plan config (system admin only).
func (h *SystemAdminHandler) UpsertPlan(c *gin.Context) {
	var req dto.UpsertPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	plan, err := h.planUC.Upsert(c.Request.Context(), &req)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, plan, nil)
}

// SetPlanActive toggles a plan's active state (system admin only).
func (h *SystemAdminHandler) SetPlanActive(c *gin.Context) {
	slug := c.Param("slug")
	type reqBody struct {
		IsActive bool `json:"is_active"`
	}
	var body reqBody
	if err := c.ShouldBindJSON(&body); err != nil {
		errors.InvalidRequestBodyResponse(c)
		return
	}

	if err := h.planUC.SetActive(c.Request.Context(), slug, body.IsActive); err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, gin.H{"slug": slug, "is_active": body.IsActive}, nil)
}

// SyncPlanEntitlements replaces the menu entitlements for a plan (system admin only).
func (h *SystemAdminHandler) SyncPlanEntitlements(c *gin.Context) {
	slug := c.Param("slug")
	type reqBody struct {
		MenuURLs []string `json:"menu_urls" binding:"required"`
	}
	var body reqBody
	if err := c.ShouldBindJSON(&body); err != nil {
		errors.InvalidRequestBodyResponse(c)
		return
	}

	if err := h.planUC.SyncMenuEntitlements(c.Request.Context(), slug, body.MenuURLs); err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, gin.H{"slug": slug, "menu_urls": body.MenuURLs}, nil)
}
