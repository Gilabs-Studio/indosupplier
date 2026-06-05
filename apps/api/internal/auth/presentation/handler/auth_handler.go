package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/gilabs/indosupplier/api/internal/auth/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/auth/domain/usecase"
	authDTO "github.com/gilabs/indosupplier/api/internal/auth/presentation/dto"
	"github.com/gilabs/indosupplier/api/internal/core/errors"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/config"
	"github.com/gilabs/indosupplier/api/internal/core/response"
)

type AuthHandler struct {
	authUC usecase.AuthUsecase
}

func NewAuthHandler(authUC usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

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
		Name:     "indosupplier_access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   config.AppConfig.JWT.AccessTokenTTL * 3600,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "indosupplier_refresh_token",
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
		Name:     "indosupplier_access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "indosupplier_refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
}

func toAuthUserDTO(user *dto.UserResponse) authDTO.UserDTO {
	resp := authDTO.UserDTO{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Capabilities: authDTO.AccountCapabilitiesDTO{
			Buyer:    user.Capabilities.Buyer,
			Supplier: user.Capabilities.Supplier,
		},
	}
	if user.BuyerProfile != nil {
		resp.BuyerProfile = &authDTO.AccountProfileRefDTO{
			ID:     user.BuyerProfile.ID,
			Status: user.BuyerProfile.Status,
		}
	}
	if user.SupplierProfile != nil {
		resp.SupplierProfile = &authDTO.AccountProfileRefDTO{
			ID:     user.SupplierProfile.ID,
			Status: user.SupplierProfile.Status,
		}
	}
	return resp
}

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
			errors.ErrorResponse(c, "ACCOUNT_DISABLED", map[string]interface{}{"reason": "User account is inactive"}, nil)
			return
		}
		errors.InternalServerErrorResponse(c, err.Error())
		return
	}

	setAuthCookies(c, loginResponse.Token, loginResponse.RefreshToken)

	resp := authDTO.LoginResponseDTO{
		User:         toAuthUserDTO(loginResponse.User),
		AccessToken:  "",
		RefreshToken: "",
	}

	response.SuccessResponse(c, resp, nil)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("indosupplier_refresh_token")
	if err != nil || refreshToken == "" {
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
		clearAuthCookies(c)
		errors.ErrorResponse(c, "REFRESH_TOKEN_INVALID", nil, nil)
		return
	}

	setAuthCookies(c, loginResponse.Token, loginResponse.RefreshToken)

	resp := authDTO.LoginResponseDTO{
		User:         toAuthUserDTO(loginResponse.User),
		AccessToken:  "",
		RefreshToken: "",
	}

	response.SuccessResponse(c, resp, nil)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("indosupplier_refresh_token")
	if err != nil || refreshToken == "" {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.ShouldBindJSON(&req)
		refreshToken = req.RefreshToken
	}

	if refreshToken != "" {
		_ = h.authUC.Logout(c.Request.Context(), refreshToken)
	}

	clearAuthCookies(c)

	secure, sameSite := getCookieSecureAndSameSite(c)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "indosupplier_csrf_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Domain:   authCookieDomain(),
		Secure:   secure,
		HttpOnly: false,
		SameSite: sameSite,
	})

	response.SuccessResponse(c, map[string]string{"message": "Logout successful"}, nil)
}

func (h *AuthHandler) GetCSRFToken(c *gin.Context) {
	token := c.GetString("csrf_token")
	if token == "" {
		token = strings.TrimSpace(c.GetHeader("X-CSRF-Token"))
	}

	secure, sameSite := getCookieSecureAndSameSite(c)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "indosupplier_csrf_token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 3600,
		Domain:   authCookieDomain(),
		Secure:   secure,
		HttpOnly: false,
		SameSite: sameSite,
	})

	response.SuccessResponse(c, gin.H{"csrf_token": token}, nil)
}
