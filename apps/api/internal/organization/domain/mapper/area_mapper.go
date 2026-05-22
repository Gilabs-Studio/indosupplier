package mapper

import (
	"encoding/json"
	"time"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
)

// ToAreaResponse converts Area model to AreaResponse DTO.
// Supervisor/member counts are computed from preloaded EmployeeAreas.
func ToAreaResponse(m *models.Area) *dto.AreaResponse {
	if m == nil {
		return nil
	}

	supervisorCount := 0
	memberCount := 0
	supervisorNames := make([]string, 0)

	for _, ea := range m.EmployeeAreas {
		if ea.IsSupervisor {
			supervisorCount++
			if ea.Employee != nil {
				supervisorNames = append(supervisorNames, ea.Employee.Name)
			}
		} else {
			memberCount++
		}
	}

	resp := &dto.AreaResponse{
		ID:              m.ID,
		Name:            m.Name,
		Description:     m.Description,
		IsActive:        m.IsActive,
		Code:            m.Code,
		Color:           m.Color,
		ManagerID:       m.ManagerID,
		Province:        m.Province,
		Regency:         m.Regency,
		District:        m.District,
		SupervisorCount: supervisorCount,
		SupervisorNames: supervisorNames,
		MemberCount:     memberCount,
		CreatedAt:       m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       m.UpdatedAt.Format(time.RFC3339),
	}

	// Parse polygon JSON string to RawMessage for response
	if m.Polygon != nil && *m.Polygon != "" {
		raw := json.RawMessage(*m.Polygon)
		resp.Polygon = &raw
	}

	// Map manager relation if preloaded
	if m.Manager != nil {
		resp.Manager = &dto.ManagerResponse{
			ID:           m.Manager.ID,
			EmployeeCode: m.Manager.EmployeeCode,
			Name:         m.Manager.Name,
		}
	}

	return resp
}

// ToAreaDetailResponse converts an Area model (with preloaded EmployeeAreas) to a detailed response DTO.
func ToAreaDetailResponse(m *models.Area) *dto.AreaDetailResponse {
	if m == nil {
		return nil
	}

	supervisors := make([]dto.EmployeeInAreaResponse, 0)
	members := make([]dto.EmployeeInAreaResponse, 0)

	for _, ea := range m.EmployeeAreas {
		if ea.Employee == nil {
			continue
		}

		emp := ea.Employee
		empResp := dto.EmployeeInAreaResponse{
			ID:           emp.ID,
			EmployeeCode: emp.EmployeeCode,
			Name:         emp.Name,
			Email:        emp.Email,
			IsSupervisor: ea.IsSupervisor,
		}

		if emp.DivisionID != nil {
			empResp.DivisionID = emp.DivisionID
		}
		if emp.Division != nil {
			empResp.DivisionName = emp.Division.Name
		}
		if emp.JobPosition != nil {
			empResp.JobPosition = emp.JobPosition.Name
		}

		if ea.IsSupervisor {
			supervisors = append(supervisors, empResp)
		} else {
			members = append(members, empResp)
		}
	}

	detail := &dto.AreaDetailResponse{
		ID:              m.ID,
		Name:            m.Name,
		Description:     m.Description,
		IsActive:        m.IsActive,
		Code:            m.Code,
		Color:           m.Color,
		ManagerID:       m.ManagerID,
		Province:        m.Province,
		Regency:         m.Regency,
		District:        m.District,
		Supervisors:     supervisors,
		Members:         members,
		SupervisorCount: len(supervisors),
		MemberCount:     len(members),
		CreatedAt:       m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       m.UpdatedAt.Format(time.RFC3339),
	}

	// Parse polygon JSON string to RawMessage
	if m.Polygon != nil && *m.Polygon != "" {
		raw := json.RawMessage(*m.Polygon)
		detail.Polygon = &raw
	}

	// Map manager relation if preloaded
	if m.Manager != nil {
		detail.Manager = &dto.ManagerResponse{
			ID:           m.Manager.ID,
			EmployeeCode: m.Manager.EmployeeCode,
			Name:         m.Manager.Name,
		}
	}

	return detail
}

// ToAreaResponses converts slice of Area models to slice of AreaResponse DTOs
func ToAreaResponses(models []models.Area) []dto.AreaResponse {
	responses := make([]dto.AreaResponse, len(models))
	for i, m := range models {
		responses[i] = *ToAreaResponse(&m)
	}
	return responses
}

// AreaFromCreateRequest creates Area model from CreateAreaRequest
func AreaFromCreateRequest(req *dto.CreateAreaRequest) *models.Area {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	area := &models.Area{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
		Code:        req.Code,
		Color:       req.Color,
		ManagerID:   req.ManagerID,
		Province:    req.Province,
		Regency:     req.Regency,
		District:    req.District,
	}

	// Store polygon as JSON string
	if req.Polygon != nil {
		polygonStr := string(*req.Polygon)
		area.Polygon = &polygonStr
	}

	return area
}

// ApplyUpdateToArea applies UpdateAreaRequest fields to existing Area model
func ApplyUpdateToArea(area *models.Area, req *dto.UpdateAreaRequest) {
	if req.Name != "" {
		area.Name = req.Name
	}
	if req.Description != "" {
		area.Description = req.Description
	}
	if req.IsActive != nil {
		area.IsActive = *req.IsActive
	}
	if req.Code != "" {
		area.Code = req.Code
	}
	if req.Polygon != nil {
		polygonStr := string(*req.Polygon)
		area.Polygon = &polygonStr
	}
	if req.Color != "" {
		area.Color = req.Color
	}
	if req.ManagerID != nil {
		area.ManagerID = req.ManagerID
	}
	if req.Province != "" {
		area.Province = req.Province
	}
	if req.Regency != "" {
		area.Regency = req.Regency
	}
	if req.District != "" {
		area.District = req.District
	}
}
