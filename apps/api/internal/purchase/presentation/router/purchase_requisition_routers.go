package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/purchase/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	purchaseRequisitionCreate  = "purchase_requisition.create"
	purchaseRequisitionRead    = "purchase_requisition.read"
	purchaseRequisitionUpdate  = "purchase_requisition.update"
	purchaseRequisitionDelete  = "purchase_requisition.delete"
	purchaseRequisitionSubmit  = "purchase_requisition.submit"
	purchaseRequisitionApprove = "purchase_requisition.approve"
	purchaseRequisitionReject  = "purchase_requisition.reject"
	purchaseRequisitionConvert = "purchase_requisition.convert"
	purchaseRequisitionExport  = "purchase_requisition.export"
	purchaseRequisitionPrint   = "purchase_requisition.print"
)

func RegisterPurchaseRequisitionRoutes(rg *gin.RouterGroup, h *handler.PurchaseRequisitionHandler, printH *handler.PurchaseRequisitionPrintHandler) {
	g := rg.Group("/purchase-requisitions")
	{
		g.GET("", middleware.RequirePermission(purchaseRequisitionRead), h.List)
		g.GET("/add", middleware.RequirePermission(purchaseRequisitionCreate), h.AddData)
		g.GET("/export", middleware.RequirePermission(purchaseRequisitionExport), h.Export)
		g.POST("", middleware.RequirePermission(purchaseRequisitionCreate), h.Create)
		g.GET("/:id", middleware.RequirePermission(purchaseRequisitionRead), h.GetByID)
		g.PUT("/:id", middleware.RequirePermission(purchaseRequisitionUpdate), h.Update)
		g.POST("/:id/submit", middleware.RequirePermission(purchaseRequisitionSubmit), h.Submit)
		g.POST("/:id/approve", middleware.RequirePermission(purchaseRequisitionApprove), h.Approve)
		g.POST("/:id/reject", middleware.RequirePermission(purchaseRequisitionReject), h.Reject)
		g.PATCH("/:id/submit", middleware.RequirePermission(purchaseRequisitionSubmit), h.Submit)
		g.PATCH("/:id/approve", middleware.RequirePermission(purchaseRequisitionApprove), h.Approve)
		g.PATCH("/:id/reject", middleware.RequirePermission(purchaseRequisitionReject), h.Reject)
		g.POST("/:id/convert", middleware.RequirePermission(purchaseRequisitionConvert), h.Convert)
		g.GET("/:id/audit-trail", middleware.RequirePermission(purchaseRequisitionRead), h.AuditTrail)
		g.GET("/:id/print", middleware.RequirePermission(purchaseRequisitionPrint), printH.PrintPurchaseRequisition)
		g.DELETE("/:id", middleware.RequirePermission(purchaseRequisitionDelete), h.Delete)
	}
}
