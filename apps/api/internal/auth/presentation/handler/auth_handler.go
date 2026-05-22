package handler

import (
	goerrors "errors"
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/auth/domain/dto"
	"github.com/gilabs/gims/api/internal/auth/domain/usecase"
	authDTO "github.com/gilabs/gims/api/internal/auth/presentation/dto"
	"github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/response"
	tenantDTO "github.com/gilabs/gims/api/internal/tenant/domain/dto"
	couponUC "github.com/gilabs/gims/api/internal/tenant/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AuthHandler handles all auth-related HTTP requests.
type AuthHandler struct {
	authUC   usecase.AuthUsecase
	couponUC couponUC.CouponUsecase
	planUC   couponUC.SubscriptionPlanUsecase
}

func NewAuthHandler(authUC usecase.AuthUsecase, couponUC couponUC.CouponUsecase, planUC couponUC.SubscriptionPlanUsecase) *AuthHandler {
	return &AuthHandler{
		authUC:   authUC,
		couponUC: couponUC,
		planUC:   planUC,
	}
}

// ValidateCoupon is a public endpoint that lets the registration form check a coupon
// before submission. When email is provided it also checks the one-time-per-email rule.
func (h *AuthHandler) ValidateCoupon(c *gin.Context) {
	code := strings.TrimSpace(c.Query("code"))
	if code == "" || len(code) > 64 {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"code": "required, max 64 chars"}, nil)
		return
	}

	email := strings.TrimSpace(c.Query("email"))

	var result interface{}
	var err error

	if email != "" {
		result, err = h.couponUC.ValidateForEmail(c.Request.Context(), code, email)
	} else {
		result, err = h.couponUC.Validate(c.Request.Context(), code)
	}

	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

// CheckAvailability checks if email and company name are available for registration.
// GET /api/v1/auth/check-availability?email=...&company_name=...
func (h *AuthHandler) CheckAvailability(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	companyName := strings.TrimSpace(c.Query("company_name"))

	if email == "" && companyName == "" {
		errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{"query": "either email or company_name is required"}, nil)
		return
	}

	result, err := h.authUC.CheckAvailability(c.Request.Context(), email, companyName)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	response.SuccessResponse(c, result, nil)
}

// isHTTPS returns true when the current request arrived over TLS, either
// directly or via a trusted proxy that sets X-Forwarded-Proto.
func isHTTPS(c *gin.Context) bool {
	if config.AppConfig == nil || config.AppConfig.Server.Env != "production" {
		return false
	}
	if c.Request.TLS != nil {
		return true
	}
	if config.AppConfig.Security.ProxyHeadersEnabled {
		xfp := strings.ToLower(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")))
		return xfp == "https"
	}
	return false
}

func authCookieDomain() string {
	if config.AppConfig != nil && config.AppConfig.Server.Env == "production" {
		return config.AppConfig.Server.RootDomain
	}
	return ""
}

func getCookieSecureAndSameSite(c *gin.Context) (bool, http.SameSite) {
	isSec := isHTTPS(c)
	if config.AppConfig != nil && config.AppConfig.Server.Env == "production" {
		isSec = true
	}
	if isSec {
		return true, http.SameSiteNoneMode
	}
	return false, http.SameSiteLaxMode
}

func setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	domain := authCookieDomain()
	secure, sameSite := getCookieSecureAndSameSite(c)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "gims_access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   config.AppConfig.JWT.AccessTokenTTL * 3600,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "gims_refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   config.AppConfig.JWT.RefreshTokenTTL * 24 * 3600,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
}

func clearAuthCookies(c *gin.Context) {
	domain := authCookieDomain()
	secure, sameSite := getCookieSecureAndSameSite(c)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "gims_access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "gims_refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
}

// Login handles login request
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	loginResponse, err := h.authUC.Login(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			errors.ErrorResponse(c, "INVALID_CREDENTIALS", nil, nil)
			return
		}
		if err == usecase.ErrUserInactive {
			errors.ErrorResponse(c, "ACCOUNT_DISABLED", map[string]interface{}{
				"reason": "User account is inactive",
			}, nil)
			return
		}
		if err == usecase.ErrSubscriptionSuspended {
			errors.ErrorResponse(c, "ACCOUNT_SUSPENDED", map[string]interface{}{
				"reason": "Subscription is suspended. Contact owner/admin to restore access.",
			}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	// Security: Set HttpOnly Cookies
	setAuthCookies(c, loginResponse.Token, loginResponse.RefreshToken)

	// Map to Presentation DTO (Strict Mode: No tokens in JSON)
	resp := authDTO.LoginResponseDTO{
		User: authDTO.UserDTO{
			ID:         loginResponse.User.ID,
			Name:       loginResponse.User.Name,
			Email:      loginResponse.User.Email,
			AvatarURL:  loginResponse.User.AvatarURL,
			EmployeeID: loginResponse.User.EmployeeID,
			Role: authDTO.RoleDTO{
				Code:      loginResponse.User.Role,
				Name:      loginResponse.User.RoleName,
				DataScope: loginResponse.User.RoleDataScope,
				IsOwner:   loginResponse.User.IsOwner,
			},
			Permissions:        loginResponse.User.Permissions,
			TenantID:           loginResponse.User.TenantID,
			TenantName:         loginResponse.User.TenantName,
			SubscriptionPlan:   loginResponse.User.SubscriptionPlan,
			SubscriptionAccess: mapSubscriptionAccess(loginResponse.User.SubscriptionAccess),
		},
		AccessToken:  "", // Removed for security
		RefreshToken: "", // Removed for security
	}

	response.SuccessResponse(c, resp, nil)
}

// RefreshToken handles refresh token request
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 1. Try to get refresh token from Cookie first (Strict/Browser)
	refreshToken, err := c.Cookie("gims_refresh_token")
	if err != nil || refreshToken == "" {
		// 2. Fallback to JSON body (Mobile/CLI)
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if errBind := c.ShouldBindJSON(&req); errBind == nil && req.RefreshToken != "" {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		errors.ErrorResponse(c, "REFRESH_TOKEN_REQUIRED", nil, nil)
		return
	}

	loginResponse, err := h.authUC.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		if err == usecase.ErrSubscriptionSuspended {
			errors.ErrorResponse(c, "ACCOUNT_SUSPENDED", map[string]interface{}{
				"reason": "Subscription is suspended. Contact owner/admin to restore access.",
			}, nil)
			return
		}
		if err == usecase.ErrUserNotFound {
			errors.ErrorResponse(c, "USER_NOT_FOUND", nil, nil)
			return
		}
		if err == usecase.ErrRefreshTokenInvalid || err == usecase.ErrRefreshTokenRevoked || err == usecase.ErrRefreshTokenExpired {
			clearAuthCookies(c)
			errors.ErrorResponse(c, "REFRESH_TOKEN_INVALID", nil, nil)
			return
		}
		// Clear cookies if refresh fails
		clearAuthCookies(c)

		errors.ErrorResponse(c, "REFRESH_TOKEN_INVALID", nil, nil)
		return
	}

	// Security: Set New HttpOnly Cookies
	setAuthCookies(c, loginResponse.Token, loginResponse.RefreshToken)

	// Map to Presentation DTO
	resp := authDTO.LoginResponseDTO{
		User: authDTO.UserDTO{
			ID:         loginResponse.User.ID,
			Name:       loginResponse.User.Name,
			Email:      loginResponse.User.Email,
			AvatarURL:  loginResponse.User.AvatarURL,
			EmployeeID: loginResponse.User.EmployeeID,
			Role: authDTO.RoleDTO{
				Code:      loginResponse.User.Role,
				Name:      loginResponse.User.RoleName,
				DataScope: loginResponse.User.RoleDataScope,
				IsOwner:   loginResponse.User.IsOwner,
			},
			Permissions:        loginResponse.User.Permissions,
			TenantID:           loginResponse.User.TenantID,
			TenantName:         loginResponse.User.TenantName,
			SubscriptionPlan:   loginResponse.User.SubscriptionPlan,
			SubscriptionAccess: mapSubscriptionAccess(loginResponse.User.SubscriptionAccess),
		},
		AccessToken:  "",
		RefreshToken: "",
	}

	response.SuccessResponse(c, resp, nil)
}

// Logout handles logout request
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is handled client-side
	// Server can maintain a blacklist if needed
	// Here we revoke the refresh token provided in the body or simply return success
	// The original implementation accepted no body and just returned success,
	// but the service had a Logout method taking a token.
	// Let's see if we can get the token from header or body.

	// 1. Try Cookie
	refreshToken, err := c.Cookie("gims_refresh_token")
	if err != nil || refreshToken == "" {
		// 2. Try JSON
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.ShouldBindJSON(&req)
		refreshToken = req.RefreshToken
	}

	if refreshToken != "" {
		_ = h.authUC.Logout(c.Request.Context(), refreshToken)
	}

	// ALWAYS Clear Cookies on Logout
	clearAuthCookies(c)
	// Also clear CSRF token cookie
	secure, sameSite := getCookieSecureAndSameSite(c)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "gims_csrf_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: false, // CSRF token is not HttpOnly so JS can read it
		SameSite: sameSite,
	})

	response.SuccessResponseNoContent(c)
}

// GetCSRFToken ensures the client has a CSRF cookie (middleware handles setting it)
// and returns the token value in the response body so cross-origin clients can
// read it without relying on CORS header exposure (which can silently fail in
// certain browser/CDN configurations).
func (h *AuthHandler) GetCSRFToken(c *gin.Context) {
	token, err := c.Cookie("gims_csrf_token")
	if err != nil || token == "" {
		// Middleware should have set it; if for some reason it didn't, return a
		// generic message — the header still carries the token for same-origin.
		response.SuccessResponse(c, gin.H{"message": "CSRF token set"}, nil)
		return
	}
	response.SuccessResponse(c, gin.H{"csrf_token": token}, nil)
}

// RegisterTenant handles self-service tenant registration.
// Creates a new tenant + admin account in one transaction and returns a JWT
// so the caller can immediately access the platform (no separate login required).
// When a paid plan is selected, returns an invoice URL for Xendit redirect instead.
func (h *AuthHandler) RegisterTenant(c *gin.Context) {
	var req dto.RegisterTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Provide enhanced error messages for user_count field validation
			var fieldErrors []response.FieldError
			hasUserCountError := false

			for _, fieldError := range validationErrors {
				if fieldError.StructField() == "UserCount" {
					hasUserCountError = true
					var message string
					switch fieldError.Tag() {
					case "max":
						message = "User count must not exceed 999 per registration (system limit)"
					case "min":
						message = "User count must be at least 1"
					default:
						message = "Invalid user count value"
					}
					fieldErrors = append(fieldErrors, response.FieldError{
						Field:   "user_count",
						Code:    fieldError.Tag(),
						Message: message,
					})
				}
			}

			// If we found user_count errors, return them with better context
			if hasUserCountError && len(fieldErrors) > 0 {
				errors.ValidationErrorResponse(c, fieldErrors)
				return
			}

			// For other validation errors, use generic handler
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	result, err := h.authUC.RegisterTenant(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrEmailAlreadyTaken:
			errors.ErrorResponse(c, "EMAIL_ALREADY_TAKEN", map[string]interface{}{
				"field": "email",
			}, nil)
		case usecase.ErrSlugAlreadyTaken:
			errors.ErrorResponse(c, "COMPANY_NAME_TAKEN", map[string]interface{}{
				"field": "company_name",
			}, nil)
		case usecase.ErrCompanyNameRequired:
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
				"field":   "company_name",
				"message": "Company name is required",
			}, nil)
		case usecase.ErrPaymentRequired:
			errors.ErrorResponse(c, "PAYMENT_REQUIRED", map[string]interface{}{
				"message": "A valid coupon or subscription plan is required to register",
			}, nil)
		case usecase.ErrPaymentGatewayUnavailable:
			errors.ErrorResponse(c, "SERVICE_UNAVAILABLE", map[string]interface{}{
				"reason": "Paid registration is temporarily unavailable because Xendit is not configured",
			}, nil)
		case usecase.ErrCouponInvalid:
			errors.ErrorResponse(c, "COUPON_INVALID", map[string]interface{}{
				"field":   "coupon",
				"message": "The coupon is invalid, inactive, or has reached its usage limit",
			}, nil)
		case usecase.ErrCouponAlreadyUsed:
			errors.ErrorResponse(c, "COUPON_ALREADY_USED", map[string]interface{}{
				"field":   "coupon",
				"message": "This coupon has already been used with your email address",
			}, nil)
		case usecase.ErrCouponUserLimitExceeded:
			errors.ErrorResponse(c, "COUPON_USER_LIMIT_EXCEEDED", map[string]interface{}{
				"field":   "user_count",
				"message": "Selected user count exceeds coupon limit",
				"hint":    "Please reduce the number of users to match your coupon's maximum allowed user count, or remove the coupon to use plan pricing instead",
			}, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	// Paid plan path: return invoice URL for frontend Xendit redirect.
	if result.RequiresPayment && result.InvoiceResponse != nil {
		response.SuccessResponse(c, result.InvoiceResponse, nil)
		return
	}

	loginResponse := result.LoginResponse

	// Set HttpOnly auth cookies so the browser-based frontend is immediately logged in.
	setAuthCookies(c, loginResponse.Token, loginResponse.RefreshToken)

	resp := authDTO.LoginResponseDTO{
		User: authDTO.UserDTO{
			ID:        loginResponse.User.ID,
			Name:      loginResponse.User.Name,
			Email:     loginResponse.User.Email,
			AvatarURL: loginResponse.User.AvatarURL,
			Role: authDTO.RoleDTO{
				Code:      loginResponse.User.Role,
				Name:      loginResponse.User.RoleName,
				DataScope: loginResponse.User.RoleDataScope,
				IsOwner:   loginResponse.User.IsOwner,
			},
			Permissions:        loginResponse.User.Permissions,
			TenantID:           loginResponse.User.TenantID,
			TenantName:         loginResponse.User.TenantName,
			SubscriptionPlan:   loginResponse.User.SubscriptionPlan,
			SubscriptionAccess: mapSubscriptionAccess(loginResponse.User.SubscriptionAccess),
		},
		AccessToken:  "",
		RefreshToken: "",
	}

	response.SuccessResponse(c, resp, nil)
}

// ConfirmPendingRegistration finalizes paid registration after payment callback and
// immediately logs the user in by setting HttpOnly auth cookies.
// POST /api/v1/auth/register/confirm
func (h *AuthHandler) ConfirmPendingRegistration(c *gin.Context) {
	var req dto.ConfirmPendingRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, validationErrors)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}

	loginResponse, err := h.authUC.ConfirmPendingRegistration(c.Request.Context(), req.Token)
	if err != nil {
		switch {
		case goerrors.Is(err, usecase.ErrPendingRegistrationNotFound):
			errors.ErrorResponse(c, "PENDING_REGISTRATION_NOT_FOUND", nil, nil)
		case goerrors.Is(err, usecase.ErrPendingRegistrationDataInvalid):
			errors.ErrorResponse(c, "INVALID_PAYLOAD", nil, nil)
		case goerrors.Is(err, usecase.ErrEmailAlreadyTaken):
			errors.ErrorResponse(c, "EMAIL_ALREADY_TAKEN", nil, nil)
		case goerrors.Is(err, usecase.ErrSlugAlreadyTaken):
			errors.ErrorResponse(c, "SLUG_ALREADY_TAKEN", nil, nil)
		default:
			errors.InternalServerErrorResponse(c, err.Error())
		}
		return
	}

	setAuthCookies(c, loginResponse.Token, loginResponse.RefreshToken)

	resp := authDTO.LoginResponseDTO{
		User: authDTO.UserDTO{
			ID:        loginResponse.User.ID,
			Name:      loginResponse.User.Name,
			Email:     loginResponse.User.Email,
			AvatarURL: loginResponse.User.AvatarURL,
			Role: authDTO.RoleDTO{
				Code:      loginResponse.User.Role,
				Name:      loginResponse.User.RoleName,
				DataScope: loginResponse.User.RoleDataScope,
				IsOwner:   loginResponse.User.IsOwner,
			},
			Permissions:        loginResponse.User.Permissions,
			TenantID:           loginResponse.User.TenantID,
			TenantName:         loginResponse.User.TenantName,
			SubscriptionPlan:   loginResponse.User.SubscriptionPlan,
			SubscriptionAccess: mapSubscriptionAccess(loginResponse.User.SubscriptionAccess),
		},
		AccessToken:  "",
		RefreshToken: "",
	}

	response.SuccessResponse(c, resp, nil)
}

// ListPublicPlans returns all active subscription plans for the public registration form.
// GET /api/v1/auth/plans
func (h *AuthHandler) ListPublicPlans(c *gin.Context) {
	plans, err := h.planUC.ListPublic(c.Request.Context())
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}
	response.SuccessResponse(c, plans, nil)
}

// ComputePrice computes the invoice amount for a plan/billing/usercount combination,
// optionally applying a coupon discount.
// POST /api/v1/auth/plans/compute-price
func (h *AuthHandler) ComputePrice(c *gin.Context) {
	var req struct {
		PlanSlug      string `json:"plan_slug"      binding:"required"`
		BillingPeriod string `json:"billing_period" binding:"required,oneof=monthly yearly"`
		UserCount     int    `json:"user_count"     binding:"omitempty,min=1,max=500"`
		CouponCode    string `json:"coupon_code"    binding:"omitempty,max=64"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			errors.HandleValidationError(c, ve)
			return
		}
		errors.InvalidRequestBodyResponse(c)
		return
	}
	if req.UserCount < 1 {
		req.UserCount = 1
	}

	ctx := c.Request.Context()
	baseAmount, err := h.planUC.ComputePrice(ctx, req.PlanSlug, req.BillingPeriod, req.UserCount)
	if err != nil {
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	resp := &tenantDTO.ComputePriceResponse{
		Slug:           req.PlanSlug,
		BillingPeriod:  req.BillingPeriod,
		UserCount:      req.UserCount,
		BaseAmountIDR:  baseAmount,
		FinalAmountIDR: baseAmount,
	}

	if couponCode := strings.TrimSpace(req.CouponCode); couponCode != "" {
		discounted, applyErr := h.couponUC.ApplyDiscount(ctx, couponCode, req.PlanSlug, baseAmount, req.UserCount, req.BillingPeriod)
		if goerrors.Is(applyErr, couponUC.ErrCouponUserLimit) {
			errors.ErrorResponse(c, "VALIDATION_ERROR", map[string]interface{}{
				"user_count": "selected user count exceeds coupon limit",
			}, nil)
			return
		}
		if applyErr == nil {
			resp.FinalAmountIDR = discounted
			resp.DiscountIDR = baseAmount - discounted
			resp.CouponApplied = true
			resp.CouponCode = couponCode
		}
	}

	response.SuccessResponse(c, resp, nil)
}

func mapSubscriptionAccess(access *dto.SubscriptionAccessResponse) *authDTO.SubscriptionAccessDTO {
	if access == nil {
		return nil
	}
	return &authDTO.SubscriptionAccessDTO{
		State:                access.State,
		Enforcement:          access.Enforcement,
		DaysOverdue:          access.DaysOverdue,
		GracePeriodDays:      access.GracePeriodDays,
		ForceBillingRedirect: access.ForceBillingRedirect,
		AllowRead:            access.AllowRead,
		AllowWrite:           access.AllowWrite,
		Message:              access.Message,
		BillingPath:          access.BillingPath,
	}
}
