package dto

// PermissionResponse represents permission response DTO (simplified)
type PermissionResponse struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Code   string        `json:"code"`
	Action string        `json:"action,omitempty"`
	Scope  string        `json:"scope,omitempty"` // Populated when part of a role assignment
	MenuID *string       `json:"menu_id,omitempty"`
	Menu   *MenuResponse `json:"menu,omitempty"`
}

// MenuResponse represents menu response DTO (with nested structure)
type MenuResponse struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Icon     string         `json:"icon"`
	URL      string         `json:"url"`
	Access   bool           `json:"access"`
	ParentID *string        `json:"parent_id,omitempty"`
	Parent   *MenuResponse  `json:"parent,omitempty"`
	Children []MenuResponse `json:"children,omitempty"`
	Order    int            `json:"order"`
	Status   string         `json:"status"`
}

// ActionResponse represents action response DTO
type ActionResponse struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Action string `json:"action"` // Generic action type (VIEW, CREATE, etc.)
	Access bool   `json:"access"`
}

// GetUserPermissionsResponse represents user permissions response (hierarchical menu structure)
type GetUserPermissionsResponse struct {
	Menus []MenuWithActionsResponse `json:"menus"`
}

// MenuWithActionsResponse represents menu with actions response
type MenuWithActionsResponse struct {
	ID       string                    `json:"id"`
	Name     string                    `json:"name"`
	Icon     string                    `json:"icon"`
	URL      string                    `json:"url"`
	Children []MenuWithActionsResponse `json:"children,omitempty"`
	Actions  []ActionResponse          `json:"actions,omitempty"`
}

// MenuCategoryResponse represents menu category for dynamic grouping
type MenuCategoryResponse struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Icon     string                 `json:"icon"`
	Order    int                    `json:"order"`
	Children []MenuCategoryResponse `json:"children,omitempty"`
}
