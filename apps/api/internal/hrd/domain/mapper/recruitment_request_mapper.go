package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/google/uuid"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
)

// ToRecruitmentRequestResponse converts RecruitmentRequest model to response DTO
func ToRecruitmentRequestResponse(req *models.RecruitmentRequest) *dto.RecruitmentRequestResponse {
	if req == nil {
		return nil
	}

	return &dto.RecruitmentRequestResponse{
		ID:                req.ID,
		RequestCode:       req.RequestCode,
		RequestedByID:     req.RequestedByID,
		RequestDate:       req.RequestDate.Format("2006-01-02"),
		DivisionID:        req.DivisionID,
		PositionID:        req.PositionID,
		RequiredCount:     req.RequiredCount,
		FilledCount:       req.FilledCount,
		OpenPositions:     req.OpenPositions(),
		EmploymentType:    string(req.EmploymentType),
		ExpectedStartDate: req.ExpectedStartDate.Format("2006-01-02"),
		SalaryRangeMin:    req.SalaryRangeMin,
		SalaryRangeMax:    req.SalaryRangeMax,
		JobDescription:    req.JobDescription,
		Qualifications:    req.Qualifications,
		Priority:          string(req.Priority),
		Status:            string(req.Status),
		Notes:             req.Notes,
		ApprovedByID:      req.ApprovedByID,
		ApprovedAt:        req.ApprovedAt,
		RejectedByID:      req.RejectedByID,
		RejectedAt:        req.RejectedAt,
		RejectionNotes:    req.RejectionNotes,
		ClosedAt:          req.ClosedAt,
		CreatedAt:         req.CreatedAt,
		UpdatedAt:         req.UpdatedAt,
	}
}

// ToRecruitmentRequestResponseList converts a slice of RecruitmentRequest models to response DTOs
func ToRecruitmentRequestResponseList(requests []models.RecruitmentRequest) []*dto.RecruitmentRequestResponse {
	responses := make([]*dto.RecruitmentRequestResponse, 0, len(requests))
	for i := range requests {
		responses = append(responses, ToRecruitmentRequestResponse(&requests[i]))
	}
	return responses
}

// EnrichRecruitmentResponse populates employee and organization names into the response
func EnrichRecruitmentResponse(
	resp *dto.RecruitmentRequestResponse,
	employeeMap map[string]*orgModels.Employee,
	divisionMap map[string]*orgModels.Division,
	positionMap map[string]*orgModels.JobPosition,
) {
	if resp == nil {
		return
	}

	// Requester
	if emp, ok := employeeMap[resp.RequestedByID]; ok {
		empID, _ := uuid.Parse(emp.ID)
		resp.RequestedBy = &dto.EmployeeSimpleResponse{
			ID:           empID,
			EmployeeCode: emp.EmployeeCode,
			Name:         emp.Name,
		}
	}

	// Division
	if div, ok := divisionMap[resp.DivisionID]; ok {
		resp.DivisionName = div.Name
	}

	// Position
	if pos, ok := positionMap[resp.PositionID]; ok {
		resp.PositionName = pos.Name
	}

	// Approved by
	if resp.ApprovedByID != nil {
		if emp, ok := employeeMap[*resp.ApprovedByID]; ok {
			approverID, _ := uuid.Parse(emp.ID)
			resp.ApprovedBy = &dto.EmployeeSimpleResponse{
				ID:           approverID,
				EmployeeCode: emp.EmployeeCode,
				Name:         emp.Name,
			}
		}
	}
}

// ToRecruitmentRequestModel converts CreateRecruitmentRequestDTO to RecruitmentRequest model
func ToRecruitmentRequestModel(req *dto.CreateRecruitmentRequestDTO, id, requestCode, requestedByID string) (*models.RecruitmentRequest, error) {
	expectedStartDate, err := time.Parse("2006-01-02", req.ExpectedStartDate)
	if err != nil {
		return nil, err
	}

	priority := models.RecruitmentPriorityMedium
	if req.Priority != "" {
		priority = models.RecruitmentPriority(req.Priority)
	}

	return &models.RecruitmentRequest{
		ID:                id,
		RequestCode:       requestCode,
		RequestedByID:     requestedByID,
		RequestDate:       apptime.Now(),
		DivisionID:        req.DivisionID,
		PositionID:        req.PositionID,
		RequiredCount:     req.RequiredCount,
		FilledCount:       0,
		EmploymentType:    models.RecruitmentEmploymentType(req.EmploymentType),
		ExpectedStartDate: expectedStartDate,
		SalaryRangeMin:    req.SalaryRangeMin,
		SalaryRangeMax:    req.SalaryRangeMax,
		JobDescription:    req.JobDescription,
		Qualifications:    req.Qualifications,
		Priority:          priority,
		Status:            models.RecruitmentStatusDraft,
		Notes:             req.Notes,
		CreatedBy:         &requestedByID,
	}, nil
}

// ApplyRecruitmentUpdateDTO applies partial update fields to the model
func ApplyRecruitmentUpdateDTO(model *models.RecruitmentRequest, req *dto.UpdateRecruitmentRequestDTO) error {
	if req.DivisionID != nil {
		model.DivisionID = *req.DivisionID
	}
	if req.PositionID != nil {
		model.PositionID = *req.PositionID
	}
	if req.RequiredCount != nil {
		model.RequiredCount = *req.RequiredCount
	}
	if req.EmploymentType != nil {
		model.EmploymentType = models.RecruitmentEmploymentType(*req.EmploymentType)
	}
	if req.ExpectedStartDate != nil {
		expectedDate, err := time.Parse("2006-01-02", *req.ExpectedStartDate)
		if err != nil {
			return err
		}
		model.ExpectedStartDate = expectedDate
	}
	if req.SalaryRangeMin != nil {
		model.SalaryRangeMin = req.SalaryRangeMin
	}
	if req.SalaryRangeMax != nil {
		model.SalaryRangeMax = req.SalaryRangeMax
	}
	if req.JobDescription != nil {
		model.JobDescription = *req.JobDescription
	}
	if req.Qualifications != nil {
		model.Qualifications = *req.Qualifications
	}
	if req.Priority != nil {
		model.Priority = models.RecruitmentPriority(*req.Priority)
	}
	if req.Notes != nil {
		model.Notes = req.Notes
	}
	return nil
}
