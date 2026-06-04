package dto

type FeatureStatusResponse struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	VisibleInUI bool   `json:"visible_in_ui"`
}

type MenuItemResponse struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Path        string `json:"path"`
	Persona     string `json:"persona"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type PlatformCatalogResponse struct {
	Personas      []string                `json:"personas"`
	Menus         []MenuItemResponse      `json:"menus"`
	FeatureStatus []FeatureStatusResponse `json:"feature_status"`
}

type DashboardMetricResponse struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Trend string `json:"trend"`
}

type DashboardResponse struct {
	Persona string                    `json:"persona"`
	Metrics []DashboardMetricResponse `json:"metrics"`
}
