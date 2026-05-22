package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/purchase/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	goodsReceiptRead    = "goods_receipt.read"
	goodsReceiptCreate  = "goods_receipt.create"
	goodsReceiptUpdate  = "goods_receipt.update"
	goodsReceiptDelete  = "goods_receipt.delete"
	goodsReceiptConfirm = "goods_receipt.confirm"
	goodsReceiptSubmit  = "goods_receipt.submit"
	goodsReceiptApprove = "goods_receipt.approve"
	goodsReceiptReject  = "goods_receipt.reject"
	goodsReceiptClose   = "goods_receipt.close"
	goodsReceiptConvert = "goods_receipt.convert"
	goodsReceiptExport  = "goods_receipt.export"
	goodsReceiptPrint   = "goods_receipt.print"
)

func RegisterGoodsReceiptRoutes(r *gin.RouterGroup, h *handler.GoodsReceiptHandler, printH *handler.GoodsReceiptPrintHandler) {
	g := r.Group("/goods-receipt")
	g.GET("/add", middleware.RequirePermission(goodsReceiptCreate), h.Add)
	g.GET("", middleware.RequirePermission(goodsReceiptRead), h.List)
	g.GET("/export", middleware.RequirePermission(goodsReceiptExport), h.Export)
	g.POST("", middleware.RequirePermission(goodsReceiptCreate), h.Create)
	g.GET("/:id", middleware.RequirePermission(goodsReceiptRead), h.GetByID)
	g.GET("/:id/audit-trail", middleware.RequirePermission(goodsReceiptRead), h.AuditTrail)
	g.GET("/:id/print", middleware.RequirePermission(goodsReceiptPrint), printH.PrintGoodsReceipt)
	g.PUT("/:id", middleware.RequirePermission(goodsReceiptUpdate), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(goodsReceiptDelete), h.Delete)
	g.POST("/:id/confirm", middleware.RequirePermission(goodsReceiptConfirm), h.Confirm)
	g.POST("/:id/submit", middleware.RequirePermission(goodsReceiptSubmit), h.Submit)
	g.POST("/:id/approve", middleware.RequirePermission(goodsReceiptApprove), h.Approve)
	g.POST("/:id/reject", middleware.RequirePermission(goodsReceiptReject), h.Reject)
	g.POST("/:id/close", middleware.RequirePermission(goodsReceiptClose), h.CloseGR)
	g.PATCH("/:id/confirm", middleware.RequirePermission(goodsReceiptConfirm), h.Confirm)
	g.PATCH("/:id/submit", middleware.RequirePermission(goodsReceiptSubmit), h.Submit)
	g.PATCH("/:id/approve", middleware.RequirePermission(goodsReceiptApprove), h.Approve)
	g.PATCH("/:id/reject", middleware.RequirePermission(goodsReceiptReject), h.Reject)
	g.PATCH("/:id/close", middleware.RequirePermission(goodsReceiptClose), h.CloseGR)
	g.POST("/:id/convert", middleware.RequirePermission(goodsReceiptConvert), h.ConvertToSupplierInvoice)
}
