package handler

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"

	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/pos/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/usecase"
)

// FloorPlanHandler handles HTTP requests for floor plans
type FloorPlanHandler struct {
	uc usecase.FloorPlanUsecase
}

// NewFloorPlanHandler creates a new handler
func NewFloorPlanHandler(uc usecase.FloorPlanUsecase) *FloorPlanHandler {
	return &FloorPlanHandler{uc: uc}
}

type userContext struct {
	userID    string
	companyID string
	isOwner   bool
}

func extractUserContext(c *gin.Context) (*userContext, bool) {
	uid, exists := c.Get("user_id")
	if !exists {
		coreErrors.UnauthorizedResponse(c, "missing user context")
		return nil, false
	}
	userID, ok := uid.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		coreErrors.UnauthorizedResponse(c, "invalid user context")
		return nil, false
	}

	scope, _ := c.Get("permission_scope")
	scopeStr, _ := scope.(string)

	companyID, _ := c.Get("user_company_id")
	companyIDStr, _ := companyID.(string)

	// If company ID is not in context, try to resolve it from employee profile
	// using tenant-scoped query to match the database.GetDB pattern.
	if strings.TrimSpace(companyIDStr) == "" {
		tenantID, _ := c.Get("tenant_id")
		tenantIDStr, _ := tenantID.(string)

		var employee struct {
			CompanyID *string
		}
		query := database.DB.Table("employees").
			Select("company_id").
			Where("user_id = ?", userID).
			Where("deleted_at IS NULL")

		// Apply tenant scoping if available
		if tenantIDStr != "" {
			query = query.Where("tenant_id = ?", tenantIDStr)
		}

		err := query.First(&employee).Error
		if err == nil && employee.CompanyID != nil {
			companyIDStr = *employee.CompanyID
			c.Set("user_company_id", companyIDStr)
		}
	}

	// For POS, if company context is still missing but we have a tenant ID, 
	// we allow it to proceed as the infrastructure-level ApplyScopeFilter 
	// and tenant scoping will handle data isolation. 
	// We only fail if we're doing a write operation that EXPLICITLY requires company ID.

	return &userContext{
		userID:    userID,
		companyID: companyIDStr,
		isOwner:   scopeStr == "ALL",
	}, true
}

func handleFloorPlanError(c *gin.Context, err error) {
	switch err {
	case usecase.ErrFloorPlanNotFound:
		coreErrors.NotFoundResponse(c, "floor_plan", "")
	case usecase.ErrFloorPlanAlreadyExists:
		coreErrors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
			"resource": "floor_plan",
			"field":    "outlet_id",
		}, nil)
	case usecase.ErrFloorPlanAlreadyPublished:
		coreErrors.ErrorResponse(c, "RESOURCE_ALREADY_EXISTS", map[string]interface{}{
			"resource": "floor_plan",
			"field":    "status",
		}, nil)
	case usecase.ErrFloorPlanForbidden:
		coreErrors.ForbiddenResponse(c, "pos.layout", nil)
	case usecase.ErrVersionNotFound:
		coreErrors.NotFoundResponse(c, "layout_version", "")
	case usecase.ErrTableTokenStorageNotReady:
		coreErrors.ErrorResponse(c, "SERVICE_UNAVAILABLE", map[string]interface{}{
			"resource":    "pos_table_qr_tokens",
			"reason":      "database schema is not migrated",
			"action":      "run migrations",
			"retryable":   true,
			"tenant_scope": true,
		}, nil)
	default:
		coreErrors.InternalServerErrorResponse(c, "")
	}
}

// GetFormData returns dropdown options for floor plan forms
func (h *FloorPlanHandler) GetFormData(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	formData, err := h.uc.GetFormData(c.Request.Context(), uc.companyID, uc.isOwner)
	if err != nil {
		handleFloorPlanError(c, err)
		return
	}

	response.SuccessResponse(c, formData, nil)
}

// Create handles floor plan creation
func (h *FloorPlanHandler) Create(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	var req dto.CreateFloorPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	result, err := h.uc.Create(c.Request.Context(), &req, uc.userID, uc.companyID, uc.isOwner)
	if err != nil {
		handleFloorPlanError(c, err)
		return
	}

	response.SuccessResponseCreated(c, result, nil)
}

// List handles listing floor plans
func (h *FloorPlanHandler) List(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage > 100 {
		perPage = 100
	}
	if page < 1 {
		page = 1
	}

	params := repositories.FloorPlanListParams{
		OutletID:  c.Query("outlet_id"),
		CompanyID: c.Query("company_id"),
		Search:    c.Query("search"),
		Status:    c.Query("status"),
		SortBy:    c.Query("sort_by"),
		SortDir:   c.Query("sort_dir"),
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	}

	plans, total, err := h.uc.List(c.Request.Context(), params, uc.companyID, uc.isOwner)
	if err != nil {
		handleFloorPlanError(c, err)
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	hasNext := page < totalPages
	hasPrev := page > 1
	var nextPage, prevPage *int
	if hasNext {
		n := page + 1
		nextPage = &n
	}
	if hasPrev {
		p := page - 1
		prevPage = &p
	}

	meta := &response.Meta{
		Pagination: &response.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      int(total),
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			NextPage:   nextPage,
			PrevPage:   prevPage,
		},
	}

	response.SuccessResponse(c, plans, meta)
}

// GetByID handles getting a single floor plan
func (h *FloorPlanHandler) GetByID(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	id := c.Param("id")
	result, err := h.uc.GetByID(c.Request.Context(), id, uc.companyID, uc.isOwner)
	if err != nil {
		handleFloorPlanError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Update handles floor plan metadata update
func (h *FloorPlanHandler) Update(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	id := c.Param("id")
	var req dto.UpdateFloorPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	result, err := h.uc.Update(c.Request.Context(), id, &req, uc.companyID, uc.isOwner)
	if err != nil {
		handleFloorPlanError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// SaveLayoutData handles saving canvas layout data
func (h *FloorPlanHandler) SaveLayoutData(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	id := c.Param("id")
	var req dto.SaveLayoutDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	result, err := h.uc.SaveLayoutData(c.Request.Context(), id, &req, uc.companyID, uc.isOwner)
	if err != nil {
		handleFloorPlanError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// Delete handles floor plan deletion
func (h *FloorPlanHandler) Delete(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	id := c.Param("id")
	if err := h.uc.Delete(c.Request.Context(), id, uc.companyID, uc.isOwner); err != nil {
		handleFloorPlanError(c, err)
		return
	}

	response.SuccessResponseDeleted(c, "floor_plan", id, nil)
}

// Publish handles publishing a floor plan
func (h *FloorPlanHandler) Publish(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	id := c.Param("id")
	result, err := h.uc.Publish(c.Request.Context(), id, uc.userID, uc.companyID, uc.isOwner)
	if err != nil {
		handleFloorPlanError(c, err)
		return
	}

	response.SuccessResponse(c, result, nil)
}

// ListVersions handles listing layout versions
func (h *FloorPlanHandler) ListVersions(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	id := c.Param("id")
	versions, err := h.uc.ListVersions(c.Request.Context(), id, uc.companyID, uc.isOwner)
	if err != nil {
		handleFloorPlanError(c, err)
		return
	}

	response.SuccessResponse(c, versions, nil)
}

// ─── Table QR Token Handlers ─────────────────────────────────────────────────

// GenerateTableToken creates or rotates a QR token for a specific table object.
func (h *FloorPlanHandler) GenerateTableToken(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	floorPlanID := c.Param("id")
	tableObjectID := c.Param("objectId")

	var req dto.GenerateTableTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		coreErrors.HandleValidationError(c, err)
		return
	}

	token, err := h.uc.GenerateTableToken(c.Request.Context(), floorPlanID, tableObjectID, &req, uc.companyID, uc.isOwner)
	if err != nil {
		handleFloorPlanError(c, err)
		return
	}
	enrichTableQRToken(c, token)
	response.SuccessResponse(c, token, nil)
}

// ListTableTokens returns all active QR tokens for a floor plan.
func (h *FloorPlanHandler) ListTableTokens(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	floorPlanID := c.Param("id")
	tokens, err := h.uc.ListTableTokens(c.Request.Context(), floorPlanID, uc.companyID, uc.isOwner)
	if err != nil {
		handleFloorPlanError(c, err)
		return
	}
	for i := range tokens {
		enrichTableQRToken(c, &tokens[i])
	}
	response.SuccessResponse(c, tokens, nil)
}

// enrichTableQRToken populates SelfOrderURL and QRBase64 on the token response.
// The self-order URL is derived from the request origin so it works in both
// development and production without any static configuration.
func enrichTableQRToken(c *gin.Context, t *dto.TableQRTokenResponse) {
	if t == nil || t.Token == "" {
		return
	}
	baseURL := resolveFrontendBaseURL(c)
	selfOrderURL := fmt.Sprintf("%s/id/pos/%s", baseURL, t.Token)
	t.SelfOrderURL = selfOrderURL
	// Generate a 200×200 PNG QR code and embed as a data URI for inline display.
	if pngBytes, err := qrcode.Encode(selfOrderURL, qrcode.Medium, 200); err == nil {
		t.QRBase64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
	}
}

// RevokeTableToken deactivates a QR token so its URL stops working.
func (h *FloorPlanHandler) RevokeTableToken(c *gin.Context) {
	uc, ok := extractUserContext(c)
	if !ok {
		return
	}

	floorPlanID := c.Param("id")
	tableObjectID := c.Param("objectId")

	if err := h.uc.RevokeTableToken(c.Request.Context(), floorPlanID, tableObjectID, uc.companyID, uc.isOwner); err != nil {
		handleFloorPlanError(c, err)
		return
	}
	response.SuccessResponse(c, gin.H{"revoked": true}, nil)
}
