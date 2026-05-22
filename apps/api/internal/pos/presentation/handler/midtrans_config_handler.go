package handler

import (
	"errors"
	"strings"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/response"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	posModels "github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/usecase"
	"github.com/gin-gonic/gin"
)

// XenditConfigHandler handles Xendit gateway configuration per company
type XenditConfigHandler struct {
	uc usecase.XenditConfigUsecase
}

// NewXenditConfigHandler creates the handler
func NewXenditConfigHandler(uc usecase.XenditConfigUsecase) *XenditConfigHandler {
	return &XenditConfigHandler{uc: uc}
}

// Get returns the Xendit config for the authenticated user's company (requires pos.payment.manage)
func (h *XenditConfigHandler) Get(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		coreErrors.UnauthorizedResponse(c, "missing user context")
		return
	}
	userID, ok := uid.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		coreErrors.UnauthorizedResponse(c, "invalid user context")
		return
	}

	companyID := resolveCompanyIDFromRequestContext(c, userID)
	if strings.TrimSpace(companyID) == "" {
		response.SuccessResponse(c, &dto.XenditConfigResponse{
			ConnectionStatus: string(posModels.XenditStatusNotConnected),
			IsActive:         false,
		}, nil)
		return
	}

	cfg, err := h.uc.Get(c.Request.Context(), companyID)
	if err != nil {
		if errors.Is(err, usecase.ErrXenditConfigNotFound) {
			response.SuccessResponse(c, &dto.XenditConfigResponse{
				CompanyID:        companyID,
				ConnectionStatus: string(posModels.XenditStatusNotConnected),
				IsActive:         false,
			}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, cfg, nil)
}

// GetStatus is a lightweight endpoint for cashiers to check if digital payment is available.
// Requires only pos.order.create permission (cashier-level).
func (h *XenditConfigHandler) GetStatus(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		coreErrors.UnauthorizedResponse(c, "missing user context")
		return
	}
	userID, ok := uid.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		coreErrors.UnauthorizedResponse(c, "invalid user context")
		return
	}

	companyID := resolveCompanyIDFromRequestContext(c, userID)
	if strings.TrimSpace(companyID) == "" {
		response.SuccessResponse(c, &dto.XenditConnectionStatusResponse{
			IsConnected: false,
			Status:      string(posModels.XenditStatusNotConnected),
		}, nil)
		return
	}

	status, err := h.uc.GetConnectionStatus(c.Request.Context(), companyID)
	if err != nil {
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, status, nil)
}

func resolveCompanyIDFromRequestContext(c *gin.Context, userID string) string {
	companyID, _ := c.Get("user_company_id")
	companyIDStr, _ := companyID.(string)
	companyIDStr = strings.TrimSpace(companyIDStr)
	if companyIDStr != "" {
		return companyIDStr
	}

	var employee orgModels.Employee
	err := database.DB.
		Select("company_id").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		First(&employee).Error
	if err == nil && employee.CompanyID != nil {
		companyIDStr = strings.TrimSpace(*employee.CompanyID)
		if companyIDStr != "" {
			c.Set("user_company_id", companyIDStr)
			return companyIDStr
		}
	}

	tenantID, _ := c.Get("tenant_id")
	tenantIDStr, _ := tenantID.(string)
	tenantIDStr = strings.TrimSpace(tenantIDStr)
	if tenantIDStr == "" {
		return ""
	}

	var companyIDs []string
	err = database.DB.
		Model(&orgModels.Company{}).
		Select("id").
		Where("tenant_id = ?", tenantIDStr).
		Where("deleted_at IS NULL").
		Limit(2).
		Find(&companyIDs).Error
	if err != nil || len(companyIDs) != 1 {
		return ""
	}

	companyIDStr = strings.TrimSpace(companyIDs[0])
	if companyIDStr != "" {
		c.Set("user_company_id", companyIDStr)
	}
	return companyIDStr
}

// Connect saves Xendit credentials and sets the account as connected (requires pos.payment.manage)
func (h *XenditConfigHandler) Connect(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	var req dto.ConnectXenditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	cfg, err := h.uc.Connect(c.Request.Context(), uc.companyID, &req, uc.userID)
	if err != nil {
		if errors.Is(err, usecase.ErrXenditConnectionValidation) {
			coreErrors.ErrorResponse(c, "XENDIT_CONNECTION_TEST_FAILED", map[string]interface{}{"reason": "Connection validation failed"}, nil)
			return
		}
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, cfg, nil)
}

// Update patches non-credential Xendit settings (requires pos.payment.manage)
func (h *XenditConfigHandler) Update(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	var req dto.UpdateXenditConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	cfg, err := h.uc.Update(c.Request.Context(), uc.companyID, &req, uc.userID)
	if err != nil {
		if errors.Is(err, usecase.ErrXenditConfigNotFound) {
			coreErrors.NotFoundResponse(c, "xendit_config", "")
			return
		}
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, cfg, nil)
}

// Disconnect removes Xendit credentials and marks account as not connected (requires pos.payment.manage)
func (h *XenditConfigHandler) Disconnect(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	cfg, err := h.uc.Disconnect(c.Request.Context(), uc.companyID, uc.userID)
	if err != nil {
		if errors.Is(err, usecase.ErrXenditConfigNotFound) {
			coreErrors.NotFoundResponse(c, "xendit_config", "")
			return
		}
		coreErrors.InternalServerErrorResponse(c, "")
		return
	}
	response.SuccessResponse(c, cfg, nil)
}
