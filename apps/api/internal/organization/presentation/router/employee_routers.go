package router

import (
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/organization/presentation/handler"
	"github.com/gin-gonic/gin"
)

// RegisterEmployeeRoutes registers employee routes
func RegisterEmployeeRoutes(rg *gin.RouterGroup, h *handler.EmployeeHandler) {
	g := rg.Group("/employees")

	// Static routes BEFORE parameterized /:id to prevent path conflicts
	g.GET("/form-data", middleware.RequirePermission("employee.read"), h.GetFormData)

	g.GET("", middleware.RequirePermission("employee.read"), h.List)
	g.GET("/:id", middleware.RequirePermission("employee.read"), h.GetByID)
	g.POST("", middleware.RequirePermission("employee.create"), h.Create)
	g.PUT("/:id", middleware.RequirePermission("employee.update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission("employee.delete"), h.Delete)
	g.POST("/:id/areas", middleware.RequirePermission("employee.update"), h.AssignAreas)
	g.PUT("/:id/areas", middleware.RequirePermission("employee.update"), h.BulkUpdateAreas)
	g.DELETE("/:id/areas/:area_id", middleware.RequirePermission("employee.update"), h.RemoveAreaAssignment)
	g.POST("/:id/supervisor-areas", middleware.RequirePermission("employee.assign_area"), h.AssignSupervisorAreas)

	// Employee contract management routes
	g.GET("/:id/contracts", middleware.RequirePermission("employee.read"), h.GetEmployeeContracts)
	g.POST("/:id/contracts", middleware.RequirePermission("employee.update"), h.CreateEmployeeContract)
	g.GET("/:id/contracts/active", middleware.RequirePermission("employee.read"), h.GetActiveContract)
	g.PUT("/:id/contracts/:contract_id", middleware.RequirePermission("employee.update"), h.UpdateEmployeeContract)
	g.DELETE("/:id/contracts/:contract_id", middleware.RequirePermission("employee.delete"), h.DeleteEmployeeContract)
	g.POST("/:id/contracts/:contract_id/terminate", middleware.RequirePermission("employee.update"), h.TerminateEmployeeContract)
	g.POST("/:id/contracts/:contract_id/renew", middleware.RequirePermission("employee.update"), h.RenewEmployeeContract)
	g.PATCH("/:id/contracts/active", middleware.RequirePermission("employee.update"), h.CorrectActiveEmployeeContract)

	// Employee education history management routes
	g.GET("/:id/education-histories", middleware.RequirePermission("employee.read"), h.GetEmployeeEducationHistories)
	g.POST("/:id/education-histories", middleware.RequirePermission("employee.update"), h.CreateEmployeeEducationHistory)
	g.PUT("/:id/education-histories/:education_id", middleware.RequirePermission("employee.update"), h.UpdateEmployeeEducationHistory)
	g.DELETE("/:id/education-histories/:education_id", middleware.RequirePermission("employee.delete"), h.DeleteEmployeeEducationHistory)

	// Employee certification management routes
	g.GET("/:id/certifications", middleware.RequirePermission("employee.read"), h.GetEmployeeCertifications)
	g.POST("/:id/certifications", middleware.RequirePermission("employee.update"), h.CreateEmployeeCertification)
	g.PUT("/:id/certifications/:certification_id", middleware.RequirePermission("employee.update"), h.UpdateEmployeeCertification)
	g.DELETE("/:id/certifications/:certification_id", middleware.RequirePermission("employee.delete"), h.DeleteEmployeeCertification)

	// Employee asset management routes
	g.GET("/:id/assets", middleware.RequirePermission("employee.read"), h.GetEmployeeAssets)
	g.POST("/:id/assets", middleware.RequirePermission("employee.update"), h.CreateEmployeeAsset)
	g.PUT("/:id/assets/:asset_id", middleware.RequirePermission("employee.update"), h.UpdateEmployeeAsset)
	g.POST("/:id/assets/:asset_id/return", middleware.RequirePermission("employee.update"), h.ReturnEmployeeAsset)
	g.DELETE("/:id/assets/:asset_id", middleware.RequirePermission("employee.delete"), h.DeleteEmployeeAsset)

	// Employee signature management routes
	g.GET("/:id/signature", middleware.RequirePermission("employee.read"), h.GetEmployeeSignature)
	g.POST("/:id/signature", middleware.RequirePermission("employee.update"), h.UploadEmployeeSignature)
	g.DELETE("/:id/signature", middleware.RequirePermission("employee.update"), h.DeleteEmployeeSignature)

	// Employee outlet assignment routes
	g.POST("/:id/outlets", middleware.RequirePermission("employee.update"), h.AssignOutlets)
	g.PUT("/:id/outlets", middleware.RequirePermission("employee.update"), h.BulkUpdateOutlets)

	// Employee warehouse assignment routes
	g.POST("/:id/warehouses", middleware.RequirePermission("employee.update"), h.AssignWarehouses)
	g.PUT("/:id/warehouses", middleware.RequirePermission("employee.update"), h.BulkUpdateWarehouses)
}
