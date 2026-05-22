package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/loyalty/data/repositories"
	"github.com/gilabs/gims/api/internal/loyalty/domain/dto"
	"github.com/gilabs/gims/api/internal/loyalty/domain/usecase"
	"github.com/gin-gonic/gin"
)

type LoyaltyHandler struct {
uc usecase.LoyaltyUsecase
}

func NewLoyaltyHandler(uc usecase.LoyaltyUsecase) *LoyaltyHandler {
return &LoyaltyHandler{uc: uc}
}

func extractUserID(c *gin.Context) string {
if id, exists := c.Get("user_id"); exists {
if s, ok := id.(string); ok {
return s
}
}
return ""
}

func clampPerPage(v int) int {
if v < 1 {
return 20
}
if v > 100 {
return 100
}
return v
}

// ─── Program handlers ─────────────────────────────────────────────────────────

func (h *LoyaltyHandler) CreateProgram(c *gin.Context) {
var req dto.CreateLoyaltyProgramRequest
if err := c.ShouldBindJSON(&req); err != nil {
response.ValidationErrorResponse(c, nil)
return
}

userID := extractUserID(c)
resp, err := h.uc.CreateProgram(c.Request.Context(), userID, &req)
if err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponseCreated(c, resp, nil)
}

func (h *LoyaltyHandler) UpdateProgram(c *gin.Context) {
id := c.Param("id")
var req dto.UpdateLoyaltyProgramRequest
if err := c.ShouldBindJSON(&req); err != nil {
response.ValidationErrorResponse(c, nil)
return
}
userID := extractUserID(c)
resp, err := h.uc.UpdateProgram(c.Request.Context(), id, userID, &req)
if err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponse(c, resp, nil)
}

func (h *LoyaltyHandler) DeleteProgram(c *gin.Context) {
id := c.Param("id")
if err := h.uc.DeleteProgram(c.Request.Context(), id); err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponseNoContent(c)
}

func (h *LoyaltyHandler) GetProgram(c *gin.Context) {
id := c.Param("id")
resp, err := h.uc.GetProgram(c.Request.Context(), id)
if err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponse(c, resp, nil)
}

func (h *LoyaltyHandler) ListPrograms(c *gin.Context) {
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
perPage = clampPerPage(perPage)
search := c.Query("search")

items, pagination, err := h.uc.ListPrograms(c.Request.Context(), page, perPage, search)
if err != nil {
handleLoyaltyError(c, err)
return
}
meta := &response.Meta{
Pagination: response.NewPaginationMeta(pagination.Page, pagination.PerPage, pagination.Total),
}
response.SuccessResponse(c, items, meta)
}

func (h *LoyaltyHandler) ToggleProgramActive(c *gin.Context) {
id := c.Param("id")
userID := extractUserID(c)
resp, err := h.uc.ToggleProgramActive(c.Request.Context(), id, userID)
if err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponse(c, resp, nil)
}

// ─── Member handlers ──────────────────────────────────────────────────────────

func (h *LoyaltyHandler) EnrollMember(c *gin.Context) {
var req dto.EnrollMemberRequest
if err := c.ShouldBindJSON(&req); err != nil {
response.ValidationErrorResponse(c, nil)
return
}
userID := extractUserID(c)
resp, err := h.uc.EnrollMember(c.Request.Context(), userID, &req)
if err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponseCreated(c, resp, nil)
}
func (h *LoyaltyHandler) ChangeProgram(c *gin.Context) {
	id := c.Param("id")
	var req dto.ChangeProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationErrorResponse(c, nil)
		return
	}
	userID := extractUserID(c)
	resp, err := h.uc.ChangeProgram(c.Request.Context(), id, userID, &req)
	if err != nil {
		handleLoyaltyError(c, err)
		return
	}
	response.SuccessResponse(c, resp, nil)
}
func (h *LoyaltyHandler) GetMember(c *gin.Context) {
id := c.Param("id")
resp, err := h.uc.GetMember(c.Request.Context(), id)
if err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponse(c, resp, nil)
}

func (h *LoyaltyHandler) GetMemberByCustomer(c *gin.Context) {
customerID := c.Param("customer_id")
resp, err := h.uc.GetMemberByCustomerID(c.Request.Context(), customerID)
if err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponse(c, resp, nil)
}

// LookupMember is called by the POS terminal to detect whether a customer is a member.
func (h *LoyaltyHandler) LookupMember(c *gin.Context) {
name := c.Query("name")
outletID := c.Query("outlet_id")
if name == "" || outletID == "" {
response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "name and outlet_id are required", nil, nil)
return
}

resp, err := h.uc.LookupMember(c.Request.Context(), name, outletID)
if err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponse(c, resp, nil)
}

func (h *LoyaltyHandler) ListMembers(c *gin.Context) {
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
perPage = clampPerPage(perPage)

params := repositories.MemberListParams{
Page:     page,
PerPage:  perPage,
Tier:     c.Query("tier"),
OutletID: c.Query("outlet_id"),
Search:   c.Query("search"),
}

items, pagination, err := h.uc.ListMembers(c.Request.Context(), params)
if err != nil {
handleLoyaltyError(c, err)
return
}
meta := &response.Meta{
Pagination: response.NewPaginationMeta(pagination.Page, pagination.PerPage, pagination.Total),
}
response.SuccessResponse(c, items, meta)
}

// ─── Points handlers ──────────────────────────────────────────────────────────

func (h *LoyaltyHandler) RedeemPoints(c *gin.Context) {
var req dto.RedeemPointsRequest
if err := c.ShouldBindJSON(&req); err != nil {
response.ValidationErrorResponse(c, nil)
return
}
userID := extractUserID(c)
req.ProcessedBy = &userID

resp, err := h.uc.RedeemPoints(c.Request.Context(), &req)
if err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponse(c, resp, nil)
}

func (h *LoyaltyHandler) AdjustPoints(c *gin.Context) {
var req dto.AdjustPointsRequest
if err := c.ShouldBindJSON(&req); err != nil {
response.ValidationErrorResponse(c, nil)
return
}
userID := extractUserID(c)
req.ProcessedBy = &userID

if err := h.uc.AdjustPoints(c.Request.Context(), &req); err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponseNoContent(c)
}

func (h *LoyaltyHandler) ListLedger(c *gin.Context) {
memberID := c.Param("member_id")
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
perPage = clampPerPage(perPage)

params := repositories.LedgerListParams{
MemberID: memberID,
Page:     page,
PerPage:  perPage,
}

items, pagination, err := h.uc.ListLedger(c.Request.Context(), params)
if err != nil {
handleLoyaltyError(c, err)
return
}
meta := &response.Meta{
Pagination: response.NewPaginationMeta(pagination.Page, pagination.PerPage, pagination.Total),
}
response.SuccessResponse(c, items, meta)
}

// ─── Public handler ───────────────────────────────────────────────────────────

func (h *LoyaltyHandler) PublicSelfRegister(c *gin.Context) {
var req dto.PublicSelfRegisterRequest
if err := c.ShouldBindJSON(&req); err != nil {
response.ValidationErrorResponse(c, nil)
return
}
resp, err := h.uc.PublicSelfRegister(c.Request.Context(), &req)
if err != nil {
handleLoyaltyError(c, err)
return
}
response.SuccessResponseCreated(c, resp, nil)
}

// ─── Error mapping ────────────────────────────────────────────────────────────

func handleLoyaltyError(c *gin.Context, err error) {
switch {
case errors.Is(err, usecase.ErrLoyaltyProgramNotFound):
response.ErrorResponse(c, http.StatusNotFound, "LOYALTY_PROGRAM_NOT_FOUND", err.Error(), nil, nil)
case errors.Is(err, usecase.ErrLoyaltyMemberNotFound):
response.ErrorResponse(c, http.StatusNotFound, "LOYALTY_MEMBER_NOT_FOUND", err.Error(), nil, nil)
case errors.Is(err, usecase.ErrAlreadyEnrolled):
response.ErrorResponse(c, http.StatusConflict, "LOYALTY_ALREADY_ENROLLED", err.Error(), nil, nil)
case errors.Is(err, usecase.ErrInvalidLoyaltyConfig):
response.ErrorResponse(c, http.StatusBadRequest, "LOYALTY_INVALID_CONFIG", err.Error(), nil, nil)
case errors.Is(err, usecase.ErrRewardNotFound):
response.ErrorResponse(c, http.StatusNotFound, "LOYALTY_REWARD_NOT_FOUND", err.Error(), nil, nil)
case errors.Is(err, usecase.ErrInsufficientPoints):
response.ErrorResponse(c, http.StatusConflict, "LOYALTY_INSUFFICIENT_POINTS", err.Error(), nil, nil)
case errors.Is(err, usecase.ErrPointsAlreadyAwarded):
response.ErrorResponse(c, http.StatusConflict, "LOYALTY_POINTS_ALREADY_AWARDED", err.Error(), nil, nil)
case errors.Is(err, usecase.ErrNoActiveProgram):
response.ErrorResponse(c, http.StatusNotFound, "LOYALTY_NO_ACTIVE_PROGRAM", err.Error(), nil, nil)
case errors.Is(err, usecase.ErrLoyaltyForbidden):
response.ErrorResponse(c, http.StatusForbidden, "LOYALTY_FORBIDDEN", err.Error(), nil, nil)
default:
response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", nil, nil)
}
}
