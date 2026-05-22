package dto

import "encoding/json"

// CreateAreaRequest represents create area request — enhanced in Sprint 24 with territory fields
type CreateAreaRequest struct {
	Name        string           `json:"name" binding:"required,min=2,max=100"`
	Description string           `json:"description" binding:"omitempty,max=500"`
	IsActive    *bool            `json:"is_active"`
	Code        string           `json:"code" binding:"omitempty,max=50"`
	Polygon     *json.RawMessage `json:"polygon"`          // GeoJSON polygon coordinates
	Color       string           `json:"color" binding:"omitempty,max=20"`
	ManagerID   *string          `json:"manager_id" binding:"omitempty"`
	Province    string           `json:"province" binding:"omitempty"`
	Regency     string           `json:"regency" binding:"omitempty"`
	District    string           `json:"district" binding:"omitempty"`
}

// UpdateAreaRequest represents update area request — enhanced in Sprint 24 with territory fields
type UpdateAreaRequest struct {
	Name        string           `json:"name" binding:"omitempty,min=2,max=100"`
	Description string           `json:"description" binding:"omitempty,max=500"`
	IsActive    *bool            `json:"is_active"`
	Code        string           `json:"code" binding:"omitempty,max=50"`
	Polygon     *json.RawMessage `json:"polygon"`
	Color       string           `json:"color" binding:"omitempty,max=20"`
	ManagerID   *string          `json:"manager_id"`
	Province    string           `json:"province" binding:"omitempty"`
	Regency     string           `json:"regency" binding:"omitempty"`
	District    string           `json:"district" binding:"omitempty"`
}

// ListAreasRequest represents list areas request with optional filters
type ListAreasRequest struct {
	Page          int    `form:"page" binding:"omitempty,min=1"`
	PerPage       int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search        string `form:"search" binding:"omitempty,max=100"`
	HasSupervisor *bool  `form:"has_supervisor"`
	HasMembers    *bool  `form:"has_members"`
	Province      string `form:"province" binding:"omitempty,max=100"`
	SortBy        string `form:"sort_by" binding:"omitempty,oneof=name code province created_at updated_at"`
	SortDir       string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// AssignAreaMembersRequest represents the request to assign employees as members of an area
type AssignAreaMembersRequest struct {
	EmployeeIDs []string `json:"employee_ids" binding:"required,min=1"`
}

// AssignAreaSupervisorsRequest represents the request to assign employees as supervisors of an area
type AssignAreaSupervisorsRequest struct {
	EmployeeIDs []string `json:"employee_ids" binding:"required,min=1"`
}

// RemoveAreaEmployeeRequest represents the request to remove an employee from an area
type RemoveAreaEmployeeRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
}

// EmployeeInAreaResponse represents a brief employee record in the context of an area assignment
type EmployeeInAreaResponse struct {
	ID           string  `json:"id"`
	EmployeeCode string  `json:"employee_code"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	DivisionID   *string `json:"division_id,omitempty"`
	DivisionName string  `json:"division_name,omitempty"`
	JobPosition  string  `json:"job_position,omitempty"`
	IsSupervisor bool    `json:"is_supervisor"`
}

// ManagerResponse represents manager info for the area
type ManagerResponse struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

// AreaResponse represents area response for list endpoints — enhanced with territory fields
type AreaResponse struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	IsActive        bool             `json:"is_active"`
	Code            string           `json:"code"`
	Polygon         *json.RawMessage `json:"polygon,omitempty"`
	Color           string           `json:"color"`
	ManagerID       *string          `json:"manager_id,omitempty"`
	Manager         *ManagerResponse `json:"manager,omitempty"`
	Province        string           `json:"province"`
	Regency         string           `json:"regency"`
	District        string           `json:"district"`
	SupervisorCount int              `json:"supervisor_count"`
	SupervisorNames []string         `json:"supervisor_names"`
	MemberCount     int              `json:"member_count"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
}

// AreaDetailResponse represents a detailed area response, including full member/supervisor lists
type AreaDetailResponse struct {
	ID              string                   `json:"id"`
	Name            string                   `json:"name"`
	Description     string                   `json:"description"`
	IsActive        bool                     `json:"is_active"`
	Code            string                   `json:"code"`
	Polygon         *json.RawMessage         `json:"polygon,omitempty"`
	Color           string                   `json:"color"`
	ManagerID       *string                  `json:"manager_id,omitempty"`
	Manager         *ManagerResponse         `json:"manager,omitempty"`
	Province        string                   `json:"province"`
	Regency         string                   `json:"regency"`
	District        string                   `json:"district"`
	Supervisors     []EmployeeInAreaResponse `json:"supervisors"`
	Members         []EmployeeInAreaResponse `json:"members"`
	SupervisorCount int                      `json:"supervisor_count"`
	MemberCount     int                      `json:"member_count"`
	CreatedAt       string                   `json:"created_at"`
	UpdatedAt       string                   `json:"updated_at"`
}

// AreaFormDataResponse contains dropdown options for area forms
type AreaFormDataResponse struct {
	Employees []EmployeeFormOption `json:"employees"`
}

// EmployeeFormOption represents an employee option for dropdowns
type EmployeeFormOption struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}
