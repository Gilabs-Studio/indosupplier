package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
)

// MapSalesVisitToResponse maps SalesVisit model to response DTO
func MapSalesVisitToResponse(visit *models.SalesVisit) *dto.SalesVisitResponse {
	if visit == nil {
		return nil
	}

	resp := &dto.SalesVisitResponse{
		ID:            visit.ID,
		Code:          visit.Code,
		VisitDate:     visit.VisitDate.Format("2006-01-02"),
		EmployeeID:    visit.EmployeeID,
		CompanyID:     visit.CompanyID,
		ContactPerson: visit.ContactPerson,
		ContactPhone:  visit.ContactPhone,
		Address:       visit.Address,
		VillageID:     visit.VillageID,
		Latitude:      visit.Latitude,
		Longitude:     visit.Longitude,
		Purpose:       visit.Purpose,
		Notes:         visit.Notes,
		Result:        visit.Result,
		Status:        string(visit.Status),
		CreatedBy:     visit.CreatedBy,
		CancelledBy:   visit.CancelledBy,
		CreatedAt:     visit.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     visit.UpdatedAt.Format(time.RFC3339),
	}

	// Map optional time fields
	if visit.ScheduledTime != nil {
		scheduledTime := visit.ScheduledTime.Format("15:04")
		resp.ScheduledTime = &scheduledTime
	}

	if visit.ActualTime != nil {
		actualTime := visit.ActualTime.Format("15:04")
		resp.ActualTime = &actualTime
	}

	if visit.CheckInAt != nil {
		checkInAt := visit.CheckInAt.Format(time.RFC3339)
		resp.CheckInAt = &checkInAt
	}

	if visit.CheckOutAt != nil {
		checkOutAt := visit.CheckOutAt.Format(time.RFC3339)
		resp.CheckOutAt = &checkOutAt
	}

	if visit.CancelledAt != nil {
		cancelledAt := visit.CancelledAt.Format(time.RFC3339)
		resp.CancelledAt = &cancelledAt
	}

	// Map Employee
	if visit.Employee != nil {
		resp.Employee = &dto.EmployeeBriefResponse{
			ID:           visit.Employee.ID,
			EmployeeCode: visit.Employee.EmployeeCode,
			Name:         visit.Employee.Name,
			Email:        visit.Employee.Email,
			Phone:        visit.Employee.Phone,
		}
	}

	// Map Company
	if visit.Company != nil {
		resp.Company = &dto.CompanyBriefResponse{
			ID:      visit.Company.ID,
			Name:    visit.Company.Name,
			Address: visit.Company.Address,
			Phone:   visit.Company.Phone,
		}
	}

	// Map Village with hierarchy
	if visit.Village != nil {
		resp.Village = &dto.VillageResponse{
			ID:   visit.Village.ID,
			Name: visit.Village.Name,
		}
		if visit.Village.District != nil {
			resp.Village.District = &dto.DistrictResponse{
				ID:   visit.Village.District.ID,
				Name: visit.Village.District.Name,
			}
			if visit.Village.District.City != nil {
				resp.Village.District.Regency = &dto.RegencyResponse{
					ID:   visit.Village.District.City.ID,
					Name: visit.Village.District.City.Name,
				}
				if visit.Village.District.City.Province != nil {
					resp.Village.District.Regency.Province = &dto.ProvinceResponse{
						ID:   visit.Village.District.City.Province.ID,
						Name: visit.Village.District.City.Province.Name,
					}
				}
			}
		}
	}

	// Map Details
	if len(visit.Details) > 0 {
		resp.Details = make([]dto.SalesVisitDetailResponse, len(visit.Details))
		for i, detail := range visit.Details {
			resp.Details[i] = *MapSalesVisitDetailToResponse(&detail)
		}
	}

	return resp
}

// MapSalesVisitDetailToResponse maps SalesVisitDetail model to response DTO
func MapSalesVisitDetailToResponse(detail *models.SalesVisitDetail) *dto.SalesVisitDetailResponse {
	if detail == nil {
		return nil
	}

	resp := &dto.SalesVisitDetailResponse{
		ID:            detail.ID,
		SalesVisitID:  detail.SalesVisitID,
		ProductID:     detail.ProductID,
		InterestLevel: detail.InterestLevel,
		Notes:         detail.Notes,
		Quantity:      detail.Quantity,
		Price:         detail.Price,
		CreatedAt:     detail.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     detail.UpdatedAt.Format(time.RFC3339),
	}

	// Map Product
	if detail.Product != nil {
		resp.Product = &dto.ProductResponse{
			ID:           detail.Product.ID,
			Code:         detail.Product.Code,
			Name:         detail.Product.Name,
			SellingPrice: detail.Product.SellingPrice,
			ImageURL:     detail.Product.ImageURL,
		}
	}
	
	// Map Answers
	if len(detail.Answers) > 0 {
		resp.Answers = make([]dto.SalesVisitInterestAnswerResponse, len(detail.Answers))
		for i, answer := range detail.Answers {
			resp.Answers[i] = dto.SalesVisitInterestAnswerResponse{
				ID:         answer.ID,
				QuestionID: answer.QuestionID,
				OptionID:   answer.OptionID,
				Score:      answer.Score,
			}
			
			if answer.Question != nil {
				resp.Answers[i].QuestionText = answer.Question.QuestionText
			}
			if answer.Option != nil {
				resp.Answers[i].OptionText = answer.Option.OptionText
			}
		}
	}

	return resp
}

// MapSalesVisitProgressHistoryToResponse maps SalesVisitProgressHistory model to response DTO
func MapSalesVisitProgressHistoryToResponse(history *models.SalesVisitProgressHistory) *dto.SalesVisitProgressHistoryResponse {
	if history == nil {
		return nil
	}

	return &dto.SalesVisitProgressHistoryResponse{
		ID:           history.ID,
		SalesVisitID: history.SalesVisitID,
		FromStatus:   string(history.FromStatus),
		ToStatus:     string(history.ToStatus),
		Notes:        history.Notes,
		ChangedBy:    history.ChangedBy,
		CreatedAt:    history.CreatedAt.Format(time.RFC3339),
	}
}

// MapSalesVisitsToResponse maps a slice of SalesVisit models to response DTOs
func MapSalesVisitsToResponse(visits []models.SalesVisit) []dto.SalesVisitResponse {
	result := make([]dto.SalesVisitResponse, len(visits))
	for i, visit := range visits {
		result[i] = *MapSalesVisitToResponse(&visit)
	}
	return result
}

// MapSalesVisitDetailsToResponse maps a slice of SalesVisitDetail models to response DTOs
func MapSalesVisitDetailsToResponse(details []models.SalesVisitDetail) []dto.SalesVisitDetailResponse {
	result := make([]dto.SalesVisitDetailResponse, len(details))
	for i, detail := range details {
		result[i] = *MapSalesVisitDetailToResponse(&detail)
	}
	return result
}

// MapSalesVisitProgressHistoryListToResponse maps a slice of SalesVisitProgressHistory models to response DTOs
func MapSalesVisitProgressHistoryListToResponse(historyList []models.SalesVisitProgressHistory) []dto.SalesVisitProgressHistoryResponse {
	result := make([]dto.SalesVisitProgressHistoryResponse, len(historyList))
	for i, history := range historyList {
		result[i] = *MapSalesVisitProgressHistoryToResponse(&history)
	}
	return result
}

// MapSalesVisitInterestQuestionToResponse maps SalesVisitInterestQuestion model to response DTO
func MapSalesVisitInterestQuestionToResponse(question *models.SalesVisitInterestQuestion) *dto.SalesVisitInterestQuestionResponse {
	if question == nil {
		return nil
	}

	resp := &dto.SalesVisitInterestQuestionResponse{
		ID:           question.ID,
		QuestionText: question.QuestionText,
		Sequence:     question.Sequence,
	}

	if len(question.Options) > 0 {
		resp.Options = make([]dto.SalesVisitInterestOptionResponse, len(question.Options))
		for i, option := range question.Options {
			resp.Options[i] = dto.SalesVisitInterestOptionResponse{
				ID:         option.ID,
				OptionText: option.OptionText,
				Score:      option.Score,
			}
		}
	}

	return resp
}

// MapSalesVisitInterestQuestionsToResponse maps a slice of SalesVisitInterestQuestion models to response DTOs
func MapSalesVisitInterestQuestionsToResponse(questions []models.SalesVisitInterestQuestion) []dto.SalesVisitInterestQuestionResponse {
	result := make([]dto.SalesVisitInterestQuestionResponse, len(questions))
	for i, question := range questions {
		result[i] = *MapSalesVisitInterestQuestionToResponse(&question)
	}
	return result
}
