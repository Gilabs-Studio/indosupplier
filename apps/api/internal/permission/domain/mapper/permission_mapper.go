package mapper

import (
	"github.com/gilabs/gims/api/internal/permission/data/models"
	"github.com/gilabs/gims/api/internal/permission/domain/dto"
)

// ToPermissionResponse converts Permission to PermissionResponse (simplified)
func ToPermissionResponse(p *models.Permission) *dto.PermissionResponse {
	resp := &dto.PermissionResponse{
		ID:     p.ID,
		Name:   p.Name,
		Code:   p.Code,
		Action: p.Action,
		MenuID: p.MenuID,
	}
	if p.Menu != nil {
		resp.Menu = ToMenuResponse(p.Menu)
	}
	return resp
}

// ToPermissionWithScopeResponse converts Permission to PermissionResponse with scope
func ToPermissionWithScopeResponse(p *models.Permission, scope string) *dto.PermissionResponse {
	resp := ToPermissionResponse(p)
	resp.Scope = scope
	return resp
}

// ToMenuResponse converts Menu to MenuResponse (recursive for children and parent)
func ToMenuResponse(m *models.Menu) *dto.MenuResponse {
	resp := &dto.MenuResponse{
		ID:       m.ID,
		Name:     m.Name,
		Icon:     m.Icon,
		URL:      m.URL,
		Access:   m.Access,
		ParentID: m.ParentID,
		Order:    m.Order,
		Status:   m.Status,
	}

	// Include parent info
	if m.Parent != nil {
		resp.Parent = &dto.MenuResponse{
			ID:       m.Parent.ID,
			Name:     m.Parent.Name,
			Icon:     m.Parent.Icon,
			URL:      m.Parent.URL,
			Access:   m.Parent.Access,
			ParentID: m.Parent.ParentID,
			Order:    m.Parent.Order,
			Status:   m.Parent.Status,
		}
	}

	// Include children
	if len(m.Children) > 0 {
		resp.Children = make([]dto.MenuResponse, len(m.Children))
		for i, child := range m.Children {
			resp.Children[i] = *ToMenuResponse(&child)
		}
	}
	return resp
}

// ToMenuCategoryResponse converts Menu to MenuCategoryResponse for category grouping
func ToMenuCategoryResponse(m *models.Menu) *dto.MenuCategoryResponse {
	resp := &dto.MenuCategoryResponse{
		ID:    m.ID,
		Name:  m.Name,
		Icon:  m.Icon,
		Order: m.Order,
	}

	if len(m.Children) > 0 {
		resp.Children = make([]dto.MenuCategoryResponse, len(m.Children))
		for i, child := range m.Children {
			resp.Children[i] = *ToMenuCategoryResponse(&child)
		}
	}
	return resp
}
