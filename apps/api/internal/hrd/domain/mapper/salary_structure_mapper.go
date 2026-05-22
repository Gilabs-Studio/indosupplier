package mapper

import (
	hrdModels "github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
)

type SalaryStructureMapper struct{}

func NewSalaryStructureMapper() *SalaryStructureMapper {
	return &SalaryStructureMapper{}
}

func (m *SalaryStructureMapper) ToResponse(item *hrdModels.SalaryStructure) dto.SalaryStructureResponse {
	if item == nil {
		return dto.SalaryStructureResponse{}
	}

	empInfo := dto.EmployeeInfo{
		ID: item.EmployeeID,
	}
	if item.Employee != nil {
		empInfo.ID = item.Employee.ID
		empInfo.Name = item.Employee.Name
		empInfo.EmployeeCode = item.Employee.EmployeeCode
		empInfo.Email = item.Employee.Email
		// Avatar from User if linked
		if item.Employee.User != nil {
			empInfo.AvatarURL = item.Employee.User.AvatarURL
		}
	}

	return dto.SalaryStructureResponse{
		ID:            item.ID,
		EmployeeID:    item.EmployeeID,
		Employee:      empInfo,
		EffectiveDate: item.EffectiveDate,
		BasicSalary:   item.BasicSalary,
		Notes:         item.Notes,
		Status:        string(item.Status),
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}
