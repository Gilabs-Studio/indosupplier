package mapper

import (
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

type UpCountryCostMapper struct{}

func NewUpCountryCostMapper() *UpCountryCostMapper {
	return &UpCountryCostMapper{}
}

func (m *UpCountryCostMapper) ToResponse(item *financeModels.UpCountryCost) dto.UpCountryCostResponse {
	if item == nil {
		return dto.UpCountryCostResponse{}
	}

	employees := make([]dto.UpCountryCostEmployeeResponse, 0, len(item.Employees))
	for _, e := range item.Employees {
		employees = append(employees, dto.UpCountryCostEmployeeResponse{
			ID:         e.ID,
			EmployeeID: e.EmployeeID,
		})
	}

	items := make([]dto.UpCountryCostItemResponse, 0, len(item.Items))
	var totalAmount float64
	for _, it := range item.Items {
		items = append(items, dto.UpCountryCostItemResponse{
			ID:          it.ID,
			CostType:    string(it.CostType),
			Description: it.Description,
			Amount:      it.Amount,
			ExpenseDate: it.ExpenseDate,
		})
		totalAmount += it.Amount
	}

	return dto.UpCountryCostResponse{
		ID:                item.ID,
		Code:              item.Code,
		Purpose:           item.Purpose,
		Location:          item.Location,
		StartDate:         item.StartDate,
		EndDate:           item.EndDate,
		Status:            string(item.Status),
		Notes:             item.Notes,
		Employees:         employees,
		Items:             items,
		TotalAmount:       totalAmount,
		SubmittedAt:       item.SubmittedAt,
		SubmittedBy:       item.SubmittedBy,
		ManagerApprovedAt: item.ManagerApprovedAt,
		ManagerApprovedBy: item.ManagerApprovedBy,
		ManagerComment:    item.ManagerComment,
		FinanceApprovedAt: item.FinanceApprovedAt,
		FinanceApprovedBy: item.FinanceApprovedBy,
		PaidAt:            item.PaidAt,
		PaidBy:            item.PaidBy,
		CreatedBy:         item.CreatedBy,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}
