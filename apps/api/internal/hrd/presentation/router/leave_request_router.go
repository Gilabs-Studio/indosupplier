package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/hrd/presentation/handler"
	"github.com/gin-gonic/gin"
)

const (
	leaveRequestCreatePermission     = "leave_request.create"
	leaveRequestReadPermission       = "leave_request.read"
	leaveRequestUpdatePermission     = "leave_request.update"
	leaveRequestDeletePermission     = "leave_request.delete"
	leaveRequestApprovePermission    = "leave_request.approve"
	leaveRequestAuditTrailPermission = "leave_request.audit_trail"
)

// RegisterLeaveRequestRoutes registers all leave request routes with permission-based access control
func RegisterLeaveRequestRoutes(r *gin.RouterGroup, leaveRequestHandler *handler.LeaveRequestHandler) {
	leaveRequests := r.Group("/leave-requests")
	{
		// Self-service endpoints (employee own leave requests)
		leaveRequests.GET("/self", leaveRequestHandler.ListSelf)
		leaveRequests.POST("/self", leaveRequestHandler.CreateSelf)
		leaveRequests.GET("/self/:id", leaveRequestHandler.GetSelfByID)
		leaveRequests.PUT("/self/:id", leaveRequestHandler.UpdateSelf)
		leaveRequests.POST("/self/:id/cancel", leaveRequestHandler.CancelSelf)
		leaveRequests.GET("/my-balance", leaveRequestHandler.GetMyBalance)
		leaveRequests.GET("/my-form-data", leaveRequestHandler.GetMyFormData)

		// Form data endpoint (must come before /:id to avoid route conflicts)
		// Requires leave.read permission to access dropdown options
		leaveRequests.GET("/form-data", middleware.RequirePermission(leaveRequestReadPermission), leaveRequestHandler.GetFormData)

		// CRUD endpoints - Only HR/approvers with appropriate permissions can manage leave requests
		leaveRequests.POST("", middleware.RequirePermission(leaveRequestCreatePermission), leaveRequestHandler.Create) // Create new leave request
		leaveRequests.GET("", middleware.RequirePermission(leaveRequestReadPermission), leaveRequestHandler.List)      // List with filters
		leaveRequests.GET("/:id/audit-trail", middleware.RequirePermission(leaveRequestAuditTrailPermission), leaveRequestHandler.AuditTrail)
		leaveRequests.GET("/:id", middleware.RequirePermission(leaveRequestReadPermission), leaveRequestHandler.GetByID)     // Get by ID
		leaveRequests.PUT("/:id", middleware.RequirePermission(leaveRequestUpdatePermission), leaveRequestHandler.Update)    // Update existing
		leaveRequests.DELETE("/:id", middleware.RequirePermission(leaveRequestDeletePermission), leaveRequestHandler.Delete) // Soft delete

		// Balance endpoint - Requires read permission
		leaveRequests.GET("/balance/:employee_id", middleware.RequirePermission(leaveRequestReadPermission), leaveRequestHandler.GetBalance)
		leaveRequests.GET("/employee/:employee_id/balance", middleware.RequirePermission(leaveRequestReadPermission), leaveRequestHandler.GetBalance)

		// Approval workflow endpoints - Only approvers can approve/reject/cancel
		leaveRequests.POST("/:id/approve", middleware.RequirePermission(leaveRequestApprovePermission), leaveRequestHandler.Approve)     // Approve request
		leaveRequests.POST("/:id/reject", middleware.RequirePermission(leaveRequestApprovePermission), leaveRequestHandler.Reject)       // Reject request
		leaveRequests.POST("/:id/cancel", middleware.RequirePermission(leaveRequestApprovePermission), leaveRequestHandler.Cancel)       // Cancel request
		leaveRequests.POST("/:id/reapprove", middleware.RequirePermission(leaveRequestApprovePermission), leaveRequestHandler.Reapprove) // Re-approve cancelled/rejected request
	}
}
