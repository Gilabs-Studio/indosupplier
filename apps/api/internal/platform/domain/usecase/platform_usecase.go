package usecase

import "github.com/gilabs/indosupplier/api/internal/platform/domain/dto"

type PlatformUsecase interface {
	GetFeatureStatus() []dto.FeatureStatusResponse
	GetCatalog() dto.PlatformCatalogResponse
	GetDashboard(persona string) dto.DashboardResponse
}

type platformUsecase struct{}

func NewPlatformUsecase() PlatformUsecase {
	return &platformUsecase{}
}

func (u *platformUsecase) GetFeatureStatus() []dto.FeatureStatusResponse {
	return []dto.FeatureStatusResponse{
		{
			Code:        "ai_search_agent",
			Name:        "AI Search Agent",
			Status:      "maintenance",
			Message:     "AI sourcing agent is visible in the interface and currently under maintenance.",
			VisibleInUI: true,
		},
		{
			Code:        "live_chat",
			Name:        "Live Chat",
			Status:      "maintenance",
			Message:     "Live chat is visible in the interface and currently under maintenance.",
			VisibleInUI: true,
		},
	}
}

func (u *platformUsecase) GetCatalog() dto.PlatformCatalogResponse {
	menus := []dto.MenuItemResponse{
		{Key: "public.search", Label: "Search", Path: "/search", Persona: "public", Status: "maintenance", Description: "Search and discovery with AI agent placeholder."},
		{Key: "public.categories", Label: "Categories", Path: "/category/:slug", Persona: "public", Status: "active", Description: "Browse suppliers by category."},
		{Key: "public.supplier", Label: "Supplier profile", Path: "/supplier/:id", Persona: "public", Status: "active", Description: "Public supplier profile."},
		{Key: "buyer.dashboard", Label: "Buyer dashboard", Path: "/buyer/dashboard", Persona: "buyer", Status: "active", Description: "Buyer RFQ and sourcing overview."},
		{Key: "buyer.rfq", Label: "Buyer RFQ", Path: "/buyer/rfq", Persona: "buyer", Status: "active", Description: "Buyer RFQ workflow."},
		{Key: "buyer.compare", Label: "Supplier comparison", Path: "/buyer/compare", Persona: "buyer", Status: "active", Description: "Side-by-side supplier comparison."},
		{Key: "supplier.dashboard", Label: "Supplier dashboard", Path: "/supplier/dashboard", Persona: "supplier", Status: "active", Description: "Supplier profile and RFQ overview."},
		{Key: "supplier.profile", Label: "Supplier profile", Path: "/supplier/profile", Persona: "supplier", Status: "active", Description: "Supplier public profile management."},
		{Key: "supplier.ads", Label: "Advertising", Path: "/supplier/ads", Persona: "supplier", Status: "active", Description: "Ad campaign and boost management."},
		{Key: "supplier.support", Label: "Support", Path: "/supplier/support", Persona: "supplier", Status: "maintenance", Description: "Ticket support with live chat placeholder."},
		{Key: "admin.dashboard", Label: "Admin dashboard", Path: "/admin/dashboard", Persona: "system_admin", Status: "active", Description: "Platform operation overview."},
		{Key: "admin.verifications", Label: "Verification queue", Path: "/admin/verifications", Persona: "system_admin", Status: "active", Description: "Supplier verification review queue."},
		{Key: "admin.support", Label: "Support queue", Path: "/admin/support", Persona: "system_admin", Status: "maintenance", Description: "Ticket queue with live chat placeholder."},
	}

	return dto.PlatformCatalogResponse{
		Personas:      []string{"buyer", "supplier", "system_admin"},
		Menus:         menus,
		FeatureStatus: u.GetFeatureStatus(),
	}
}

func (u *platformUsecase) GetDashboard(persona string) dto.DashboardResponse {
	switch persona {
	case "supplier":
		return dto.DashboardResponse{
			Persona: "supplier",
			Metrics: []dto.DashboardMetricResponse{
				{Label: "Profile completeness", Value: "72%", Trend: "+8%"},
				{Label: "Incoming RFQ", Value: "18", Trend: "+4"},
				{Label: "Profile views", Value: "1,240", Trend: "+12%"},
			},
		}
	case "system_admin":
		return dto.DashboardResponse{
			Persona: "system_admin",
			Metrics: []dto.DashboardMetricResponse{
				{Label: "Pending verification", Value: "42", Trend: "+9"},
				{Label: "Open support tickets", Value: "16", Trend: "-3"},
				{Label: "Campaigns in review", Value: "7", Trend: "+2"},
			},
		}
	default:
		return dto.DashboardResponse{
			Persona: "buyer",
			Metrics: []dto.DashboardMetricResponse{
				{Label: "Active RFQ", Value: "6", Trend: "+2"},
				{Label: "Saved suppliers", Value: "24", Trend: "+5"},
				{Label: "Unread notifications", Value: "9", Trend: "+1"},
			},
		}
	}
}
