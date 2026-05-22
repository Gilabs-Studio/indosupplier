package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/feedback/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterFeedbackRoutes registers all protected feedback routes under /feedback.
func RegisterFeedbackRoutes(group *gin.RouterGroup, h *handler.FeedbackHandler) {
	fb := group.Group("/feedback")

	// Form management
	forms := fb.Group("/forms")
	forms.GET("", middleware.RequirePermission("feedback.read"), h.ListForms)
	forms.GET("/by-outlet", middleware.RequirePermission("feedback.read"), h.GetFormsByOutlet)
	forms.GET("/:id", middleware.RequirePermission("feedback.read"), h.GetForm)
	forms.POST("", middleware.RequirePermission("feedback.create"), h.CreateForm)
	forms.POST("/:id/copy", middleware.RequirePermission("feedback.create"), h.CopyForm)
	forms.PUT("/:id", middleware.RequirePermission("feedback.update"), h.UpdateForm)
	forms.DELETE("/:id", middleware.RequirePermission("feedback.delete"), h.DeleteForm)

	// Token generation (called by POS internally — staff permission)
	fb.POST("/tokens", middleware.RequirePermission("feedback.read"), h.GenerateToken)

	// Responses (admin view)
	responses := fb.Group("/responses")
	responses.GET("", middleware.RequirePermission("feedback.read"), h.ListResponses)
	responses.GET("/:id", middleware.RequirePermission("feedback.read"), h.GetResponse)
}

// RegisterFeedbackPublicRoutes registers unauthenticated routes for the public feedback page.
// These routes live directly on the root engine under /api/v1/public/feedback.
func RegisterFeedbackPublicRoutes(r *gin.RouterGroup, h *handler.FeedbackHandler) {
	pub := r.Group("/public/feedback")
	pub.GET("/:token", h.GetPublicForm)
	pub.POST("/:token/submit", h.SubmitFeedback)
}
