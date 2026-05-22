package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
)

var monthNames = map[int]string{
	1: "Januari", 2: "Februari", 3: "Maret", 4: "April",
	5: "Mei", 6: "Juni", 7: "Juli", 8: "Agustus",
	9: "September", 10: "Oktober", 11: "November", 12: "Desember",
}

// ToSalesTargetResponse converts a SalesTarget model to response DTO
func ToSalesTargetResponse(m *salesModels.SalesTarget) dto.SalesTargetResponse {
	totalActual, achievementPercent := m.CalculateAchievements()

	response := dto.SalesTargetResponse{
		ID:                 m.ID,
		EmployeeID:         m.EmployeeID,
		Year:               m.Year,
		TotalTarget:        m.TotalTarget,
		TotalActual:        totalActual,
		AchievementPercent: achievementPercent,
		Notes:              m.Notes,
		CreatedAt:          m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          m.UpdatedAt.Format(time.RFC3339),
	}

	if m.Employee != nil {
		response.Employee = &dto.EmployeeResponse{
			ID:           m.Employee.ID,
			Name:         m.Employee.Name,
			EmployeeCode: m.Employee.EmployeeCode,
			Email:        m.Employee.Email,
			Phone:        m.Employee.Phone,
		}
	}

	// Map monthly targets
	if len(m.MonthlyTargets) > 0 {
		response.MonthlyTargets = make([]dto.MonthlySalesTargetResponse, len(m.MonthlyTargets))
		for i, mt := range m.MonthlyTargets {
			response.MonthlyTargets[i] = ToMonthlySalesTargetResponse(&mt)
		}
	}

	return response
}

// ToMonthlySalesTargetResponse converts a MonthlySalesTarget model to response DTO
func ToMonthlySalesTargetResponse(m *salesModels.MonthlySalesTarget) dto.MonthlySalesTargetResponse {
	return dto.MonthlySalesTargetResponse{
		ID:                 m.ID,
		Month:              m.Month,
		MonthName:          monthNames[m.Month],
		TargetAmount:       m.TargetAmount,
		ActualAmount:       m.ActualAmount,
		AchievementPercent: m.AchievementPercent,
		Notes:              m.Notes,
	}
}

// ToSalesTargetModel converts a CreateSalesTargetRequest to SalesTarget model
func ToSalesTargetModel(req *dto.CreateSalesTargetRequest) *salesModels.SalesTarget {
	target := &salesModels.SalesTarget{
		EmployeeID:  req.EmployeeID,
		Year:        req.Year,
		TotalTarget: req.TotalTarget,
		Notes:       req.Notes,
		CreatedAt:   apptime.Now(),
		UpdatedAt:   apptime.Now(),
	}

	// Map monthly targets
	if len(req.Months) > 0 {
		target.MonthlyTargets = make([]salesModels.MonthlySalesTarget, len(req.Months))
		for i, monthReq := range req.Months {
			target.MonthlyTargets[i] = salesModels.MonthlySalesTarget{
				Month:        monthReq.Month,
				TargetAmount: monthReq.TargetAmount,
				Notes:        monthReq.Notes,
				ActualAmount: 0,
				CreatedAt:    apptime.Now(),
				UpdatedAt:    apptime.Now(),
			}
		}
	}

	return target
}

// UpdateSalesTargetModel updates a SalesTarget model from UpdateSalesTargetRequest
func UpdateSalesTargetModel(m *salesModels.SalesTarget, req *dto.UpdateSalesTargetRequest) {
	if req.TotalTarget != nil {
		m.TotalTarget = *req.TotalTarget
	}

	if req.Notes != nil {
		m.Notes = *req.Notes
	}

	// Update monthly targets if provided
	if req.Months != nil && len(*req.Months) > 0 {
		m.MonthlyTargets = make([]salesModels.MonthlySalesTarget, len(*req.Months))
		for i, monthReq := range *req.Months {
			m.MonthlyTargets[i] = salesModels.MonthlySalesTarget{
				Month:        monthReq.Month,
				TargetAmount: monthReq.TargetAmount,
				Notes:        monthReq.Notes,
				UpdatedAt:    apptime.Now(),
			}
		}
	}

	m.UpdatedAt = apptime.Now()
}
