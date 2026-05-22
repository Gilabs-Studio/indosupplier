package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/finance/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	assetTransferRead    = "asset.transfer.approve"
	assetTransferApprove = "asset.transfer.approve"
	assetTransferReject  = "asset.transfer.approve"
)

func RegisterAssetTransferRoutes(r *gin.RouterGroup, h *handler.AssetHandler) {
	registerAssetTransferRoutesInGroup(r.Group(""), h)
}

func registerAssetTransferRoutesInGroup(g *gin.RouterGroup, h *handler.AssetHandler) {
	g.GET("/transfers", middleware.RequirePermission(assetTransferRead), h.ListTransfers)
	g.POST("/approvals/:transfer_id/approve", middleware.RequirePermission(assetTransferApprove), h.ApproveTransfer)
	g.POST("/approvals/:transfer_id/reject", middleware.RequirePermission(assetTransferReject), h.RejectTransfer)
}