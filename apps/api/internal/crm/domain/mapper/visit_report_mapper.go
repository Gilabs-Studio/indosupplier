package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
)

// MapVisitReportToResponse converts a VisitReport model to response DTO
func MapVisitReportToResponse(report *models.VisitReport) *dto.VisitReportResponse {
	if report == nil {
		return nil
	}

	resp := &dto.VisitReportResponse{
		ID:               report.ID,
		Code:             report.Code,
		CustomerID:       report.CustomerID,
		ContactID:        report.ContactID,
		DealID:           report.DealID,
		LeadID:           report.LeadID,
		TravelPlanID:     report.TravelPlanID,
		EmployeeID:       report.EmployeeID,
		VisitDate:        report.VisitDate.Format("2006-01-02"),
		Address:          report.Address,
		VillageID:        report.VillageID,
		Latitude:         report.Latitude,
		Longitude:        report.Longitude,
		Purpose:          report.Purpose,
		Notes:            report.Notes,
		Result:           report.Result,
		Outcome:          report.Outcome,
		NextSteps:        report.NextSteps,
		ContactPerson:    report.ContactPerson,
		ContactPhone:     report.ContactPhone,
		Photos:           report.Photos,
		CreatedBy:        report.CreatedBy,
		CheckInLocation:  report.CheckInLocation,
		CheckOutLocation: report.CheckOutLocation,
		CreatedAt:        report.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        report.UpdatedAt.Format(time.RFC3339),
	}

	if report.ScheduledTime != nil {
		st := report.ScheduledTime.Format("15:04")
		resp.ScheduledTime = &st
	}
	if report.ActualTime != nil {
		at := report.ActualTime.Format("15:04")
		resp.ActualTime = &at
	}
	if report.CheckInAt != nil {
		ci := report.CheckInAt.Format(time.RFC3339)
		resp.CheckInAt = &ci
	}
	if report.CheckOutAt != nil {
		co := report.CheckOutAt.Format(time.RFC3339)
		resp.CheckOutAt = &co
	}

	// Map Employee
	if report.Employee != nil {
		resp.Employee = &dto.VisitEmployeeBrief{
			ID:           report.Employee.ID,
			EmployeeCode: report.Employee.EmployeeCode,
			Name:         report.Employee.Name,
			Email:        report.Employee.Email,
			Phone:        report.Employee.Phone,
		}
	}

	// Map Customer
	if report.Customer != nil {
		customerPhone := report.ContactPhone
		resp.Customer = &dto.VisitCustomerBrief{
			ID:      report.Customer.ID,
			Name:    report.Customer.Name,
			Address: report.Customer.Address,
			Phone:   customerPhone,
		}
	}

	// Map Contact
	if report.Contact != nil {
		resp.Contact = &dto.VisitContactBrief{
			ID:    report.Contact.ID,
			Name:  report.Contact.Name,
			Phone: report.Contact.Phone,
			Email: report.Contact.Email,
		}
	}

	// Map Deal
	if report.Deal != nil {
		resp.Deal = &dto.VisitDealBrief{
			ID:    report.Deal.ID,
			Code:  report.Deal.Code,
			Title: report.Deal.Title,
		}
	}

	// Map Lead
	if report.Lead != nil {
		resp.Lead = &dto.VisitLeadBrief{
			ID:        report.Lead.ID,
			Code:      report.Lead.Code,
			FirstName: report.Lead.FirstName,
			LastName:  report.Lead.LastName,
		}
	}

	// Map Village with hierarchy
	if report.Village != nil {
		resp.Village = &dto.VisitVillageResponse{
			ID:   report.Village.ID,
			Name: report.Village.Name,
		}
		if report.Village.District != nil {
			resp.Village.District = &dto.VisitDistrictResponse{
				ID:   report.Village.District.ID,
				Name: report.Village.District.Name,
			}
			if report.Village.District.City != nil {
				resp.Village.District.Regency = &dto.VisitRegencyResponse{
					ID:   report.Village.District.City.ID,
					Name: report.Village.District.City.Name,
				}
				if report.Village.District.City.Province != nil {
					resp.Village.District.Regency.Province = &dto.VisitProvinceResponse{
						ID:   report.Village.District.City.Province.ID,
						Name: report.Village.District.City.Province.Name,
					}
				}
			}
		}
	}

	// Map Details
	if len(report.Details) > 0 {
		resp.Details = make([]dto.VisitReportDetailResponse, len(report.Details))
		for i, detail := range report.Details {
			resp.Details[i] = *MapVisitReportDetailToResponse(&detail)
		}
	}

	return resp
}

// MapVisitReportDetailToResponse converts a detail model to response
func MapVisitReportDetailToResponse(detail *models.VisitReportDetail) *dto.VisitReportDetailResponse {
	if detail == nil {
		return nil
	}

	resp := &dto.VisitReportDetailResponse{
		ID:            detail.ID,
		VisitReportID: detail.VisitReportID,
		ProductID:     detail.ProductID,
		InterestLevel: detail.InterestLevel,
		Notes:         detail.Notes,
		Quantity:      detail.Quantity,
		Price:         detail.Price,
		CreatedAt:     detail.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     detail.UpdatedAt.Format(time.RFC3339),
	}

	if detail.Product != nil {
		imageURL := ""
		if detail.Product.ImageURL != nil {
			imageURL = *detail.Product.ImageURL
		}
		resp.Product = &dto.VisitProductBrief{
			ID:           detail.Product.ID,
			Code:         detail.Product.Code,
			Name:         detail.Product.Name,
			SellingPrice: detail.Product.SellingPrice,
			ImageURL:     imageURL,
		}
	}

	if len(detail.Answers) > 0 {
		resp.Answers = make([]dto.VisitReportInterestAnswerResponse, len(detail.Answers))
		for i, answer := range detail.Answers {
			resp.Answers[i] = dto.VisitReportInterestAnswerResponse{
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

// MapVisitReportsToResponse converts a slice of VisitReport models to responses
func MapVisitReportsToResponse(reports []models.VisitReport) []dto.VisitReportResponse {
	result := make([]dto.VisitReportResponse, len(reports))
	for i, report := range reports {
		result[i] = *MapVisitReportToResponse(&report)
	}
	return result
}

// MapVisitReportProgressHistoryToResponse converts a progress history model to response.
func MapVisitReportProgressHistoryToResponse(h *models.VisitReportProgressHistory) *dto.VisitReportProgressHistoryResponse {
	if h == nil {
		return nil
	}

	return &dto.VisitReportProgressHistoryResponse{
		ID:            h.ID,
		VisitReportID: h.VisitReportID,
		FromStatus:    string(h.FromStatus),
		ToStatus:      string(h.ToStatus),
		Notes:         h.Notes,
		ChangedBy:     h.ChangedBy,
		CreatedAt:     h.CreatedAt.Format(time.RFC3339),
	}
}

// MapVisitReportProgressHistoryListToResponse converts a slice of progress history models.
func MapVisitReportProgressHistoryListToResponse(historyList []models.VisitReportProgressHistory) []dto.VisitReportProgressHistoryResponse {
	result := make([]dto.VisitReportProgressHistoryResponse, len(historyList))
	for i, h := range historyList {
		result[i] = *MapVisitReportProgressHistoryToResponse(&h)
	}

	return result
}

// MapInterestQuestionToResponse converts an interest question model to response
func MapInterestQuestionToResponse(question *salesModels.SalesVisitInterestQuestion) *dto.VisitInterestQuestionResponse {
	if question == nil {
		return nil
	}
	resp := &dto.VisitInterestQuestionResponse{
		ID:           question.ID,
		QuestionText: question.QuestionText,
		Sequence:     question.Sequence,
	}
	if len(question.Options) > 0 {
		resp.Options = make([]dto.VisitInterestOptionResponse, len(question.Options))
		for i, option := range question.Options {
			resp.Options[i] = dto.VisitInterestOptionResponse{
				ID:         option.ID,
				OptionText: option.OptionText,
				Score:      option.Score,
			}
		}
	}
	return resp
}

// MapInterestQuestionsToResponse converts a slice of interest questions
func MapInterestQuestionsToResponse(questions []salesModels.SalesVisitInterestQuestion) []dto.VisitInterestQuestionResponse {
	result := make([]dto.VisitInterestQuestionResponse, len(questions))
	for i, q := range questions {
		result[i] = *MapInterestQuestionToResponse(&q)
	}
	return result
}
