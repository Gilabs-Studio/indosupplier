package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AssetHandler struct {
	uc usecase.AssetUsecase
}

func NewAssetHandler(uc usecase.AssetUsecase) *AssetHandler {
	return &AssetHandler{uc: uc}
}

func (h *AssetHandler) Create(c *gin.Context) {
	var req dto.CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	ctx := context.WithValue(c.Request.Context(), "request_ip", c.ClientIP())
	ctx = context.WithValue(ctx, "user_agent", c.GetHeader("User-Agent"))
	res, err := h.uc.Create(ctx, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_CREATE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponseCreated(c, res, nil)
}

func (h *AssetHandler) Update(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_UPDATE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) EditAsset(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.EditAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.EditAsset(c.Request.Context(), id, &req)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "immutable") {
			response.ErrorResponse(c, http.StatusForbidden, "ASSET_FIELD_IMMUTABLE", err.Error(), nil, nil)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			response.ErrorResponse(c, http.StatusNotFound, "ASSET_NOT_FOUND", err.Error(), nil, nil)
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_EDIT_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}


func (h *AssetHandler) Delete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_DELETE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponseDeleted(c, "asset", id, nil)
}

func (h *AssetHandler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	res, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "ASSET_NOT_FOUND", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) List(c *gin.Context) {
	var req dto.ListAssetsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	req.Page = page
	req.PerPage = perPage

	items, total, err := h.uc.List(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ASSET_LIST_FAILED", err.Error(), nil, nil)
		return
	}
	meta := &response.Meta{Pagination: response.NewPaginationMeta(page, perPage, int(total))}
	response.SuccessResponse(c, items, meta)
}

func (h *AssetHandler) Depreciate(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.DepreciateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Depreciate(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_DEPRECIATE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) PreviewBatchDepreciation(c *gin.Context) {
	var req dto.BatchDepreciationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.PreviewBatchDepreciation(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_BATCH_DEPRECIATION_PREVIEW_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) RunBatchDepreciation(c *gin.Context) {
	var req dto.BatchDepreciationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.RunBatchDepreciation(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_BATCH_DEPRECIATION_RUN_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) GetDepreciationSchedule(c *gin.Context) {
	var req dto.RunDepreciationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	if strings.TrimSpace(req.Period) == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "period is required", nil, nil)
		return
	}
	res, err := h.uc.GetDepreciationSchedule(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "DEPRECIATION_SCHEDULE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) RunDepreciation(c *gin.Context) {
	var req dto.RunDepreciationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.RunDepreciation(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "DEPRECIATION_RUN_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) ApproveDepreciationRun(c *gin.Context) {
	var req dto.RunDepreciationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.ApproveDepreciationRun(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "DEPRECIATION_APPROVE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) GetDepreciationHistory(c *gin.Context) {
	period := strings.TrimSpace(c.Query("period"))
	if period == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "period is required", nil, nil)
		return
	}
	res, err := h.uc.GetDepreciationHistory(c.Request.Context(), period)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "DEPRECIATION_HISTORY_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) Transfer(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.TransferAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Transfer(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_TRANSFER_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) ListTransfers(c *gin.Context) {
	var req dto.ListTransfersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.ListTransfers(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "TRANSFER_LIST_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) ApproveTransfer(c *gin.Context) {
	transferID := strings.TrimSpace(c.Param("transfer_id"))
	res, err := h.uc.ApproveTransfer(c.Request.Context(), transferID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "TRANSFER_APPROVE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) RejectTransfer(c *gin.Context) {
	transferID := strings.TrimSpace(c.Param("transfer_id"))
	var req dto.RejectTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.RejectTransfer(c.Request.Context(), transferID, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "TRANSFER_REJECT_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) PreviewDisposal(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.PreviewDisposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}

	res, err := h.uc.PreviewDisposal(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_DISPOSAL_PREVIEW_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) Dispose(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.DisposeAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Dispose(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_DISPOSE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) ApproveDepreciation(c *gin.Context) {
	id := strings.TrimSpace(c.Param("dep_id"))
	res, err := h.uc.ApproveDepreciation(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ASSET_DEPRECIATE_APPROVE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) Revalue(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.RevalueAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Revalue(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ASSET_REVALUE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) Adjust(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.AdjustAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Adjust(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ASSET_ADJUST_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) ApproveTransaction(c *gin.Context) {
	id := strings.TrimSpace(c.Param("tx_id"))
	res, err := h.uc.ApproveTransaction(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ASSET_TRANSACTION_APPROVE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) Sell(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.SellAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Sell(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ASSET_SELL_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

// ========== Phase 2: Attachments ==========

func (h *AssetHandler) ListAttachments(c *gin.Context) {
	assetID := strings.TrimSpace(c.Param("id"))
	items, err := h.uc.ListAttachments(c.Request.Context(), assetID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ATTACHMENT_LIST_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, items, nil)
}

func (h *AssetHandler) CreateAttachment(c *gin.Context) {
	assetID := strings.TrimSpace(c.Param("id"))
	assetUUID, err := uuid.Parse(assetID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ASSET_ID", "Invalid asset ID", nil, nil)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "FILE_REQUIRED", "File is required", nil, nil)
		return
	}
	defer file.Close()

	fileType := c.PostForm("file_type")
	if fileType == "" {
		fileType = "document"
	}
	description := c.PostForm("description")

	// Get content type
	contentType := header.Header.Get("Content-Type")
	fileSize := int(header.Size)

	// For now, store file info (actual file storage would use cloud storage)
	tenantID, _ := c.Request.Context().Value("tenant_id").(string)
	att := &financeModels.AssetAttachment{
		AssetID:     assetUUID,
		FileName:    header.Filename,
		FilePath:    "/tenants/" + tenantID + "/assets/" + assetID + "/" + header.Filename,
		FileURL:     "/api/v1/fixed-assets/" + assetID + "/attachments/" + header.Filename,
		FileType:    fileType,
		FileSize:    &fileSize,
		MimeType:    &contentType,
		Description: &description,
	}

	// Set uploaded by
	actorID, _ := c.Request.Context().Value("user_id").(string)
	if actorID != "" {
		actorUUID, err := uuid.Parse(actorID)
		if err == nil {
			att.UploadedBy = &actorUUID
		}
	}
	if tenantID != "" {
		tenantUUID, err := uuid.Parse(tenantID)
		if err == nil {
			att.TenantID = &tenantUUID
		}
	}

	res, err := h.uc.CreateAttachment(c.Request.Context(), assetID, att)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ATTACHMENT_CREATE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponseCreated(c, res, nil)
}

func (h *AssetHandler) DeleteAttachment(c *gin.Context) {
	assetID := strings.TrimSpace(c.Param("id"))
	attachmentID := strings.TrimSpace(c.Param("attachment_id"))
	if err := h.uc.DeleteAttachment(c.Request.Context(), assetID, attachmentID); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ATTACHMENT_DELETE_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponseDeleted(c, "attachment", attachmentID, nil)
}

// ========== Phase 2: Assignments ==========

func (h *AssetHandler) AssignAsset(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.AssignAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Assign(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_ASSIGN_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

func (h *AssetHandler) ReturnAsset(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.ReturnAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, nil)
		return
	}
	res, err := h.uc.Return(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ASSET_RETURN_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, res, nil)
}

// ========== Phase 2: Audit Logs & Assignment History ==========

func (h *AssetHandler) ListAuditLogs(c *gin.Context) {
	assetID := strings.TrimSpace(c.Param("id"))
	items, err := h.uc.ListAuditLogs(c.Request.Context(), assetID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "AUDIT_LOG_LIST_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, items, nil)
}

func (h *AssetHandler) ListAssignmentHistory(c *gin.Context) {
	assetID := strings.TrimSpace(c.Param("id"))
	items, err := h.uc.ListAssignmentHistory(c.Request.Context(), assetID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "ASSIGNMENT_HISTORY_LIST_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, items, nil)
}

// GetAvailableAssets returns list of assets available for employee borrowing
func (h *AssetHandler) GetAvailableAssets(c *gin.Context) {
	items, err := h.uc.GetAvailableAssets(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "AVAILABLE_ASSETS_LIST_FAILED", err.Error(), nil, nil)
		return
	}
	response.SuccessResponse(c, items, nil)
}
