package mapper

import (
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/organization/domain/dto"
)

func ToEducationHistoryResponse(education *models.EmployeeEducationHistory) dto.EmployeeEducationHistoryResponse {
	resp := dto.EmployeeEducationHistoryResponse{
		ID:            education.ID,
		EmployeeID:    education.EmployeeID,
		Institution:   education.Institution,
		Degree:        string(education.Degree),
		FieldOfStudy:  education.FieldOfStudy,
		StartDate:     education.StartDate.Format("2006-01-02"),
		GPA:           education.GPA,
		Description:   education.Description,
		DocumentPath:  education.DocumentPath,
		IsCompleted:   education.IsCompleted(),
		DurationYears: education.GetDurationYears(),
		CreatedAt:     education.CreatedAt,
		UpdatedAt:     education.UpdatedAt,
	}

	if education.EndDate != nil {
		endDateStr := education.EndDate.Format("2006-01-02")
		resp.EndDate = &endDateStr
	}

	return resp
}

func ToEducationHistoryResponseList(educations []*models.EmployeeEducationHistory) []dto.EmployeeEducationHistoryResponse {
	responses := make([]dto.EmployeeEducationHistoryResponse, len(educations))
	for i, education := range educations {
		responses[i] = ToEducationHistoryResponse(education)
	}
	return responses
}

func ToEducationBriefResponse(education *models.EmployeeEducationHistory) *dto.EmployeeEducationBriefResponse {
	if education == nil {
		return nil
	}

	resp := &dto.EmployeeEducationBriefResponse{
		ID:           education.ID.String(),
		Institution:  education.Institution,
		Degree:       string(education.Degree),
		FieldOfStudy: education.FieldOfStudy,
		StartDate:    education.StartDate.Format("2006-01-02"),
		GPA:          education.GPA,
		DocumentPath: education.DocumentPath,
		IsCompleted:  education.IsCompleted(),
	}

	if education.EndDate != nil {
		endDateStr := education.EndDate.Format("2006-01-02")
		resp.EndDate = &endDateStr
	}

	return resp
}
