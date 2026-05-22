package mapper

import (
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToLeadResponse converts a Lead model to LeadResponse DTO
func ToLeadResponse(lead *models.Lead) dto.LeadResponse {
	resp := dto.LeadResponse{
		ID:                   lead.ID,
		Code:                 lead.Code,
		FirstName:            lead.FirstName,
		LastName:             lead.LastName,
		CompanyName:          lead.CompanyName,
		Email:                lead.Email,
		Phone:                lead.Phone,
		ContactRoleID:        lead.ContactRoleID,
		JobTitle:             lead.JobTitle,
		Address:              lead.Address,
		City:                 lead.City,
		Province:             lead.Province,
		ProvinceID:           lead.ProvinceID,
		CityID:               lead.CityID,
		DistrictID:           lead.DistrictID,
		VillageName:          lead.VillageName,
		LeadSourceID:         lead.LeadSourceID,
		LeadStatusID:         lead.LeadStatusID,
		LeadScore:            lead.LeadScore,
		Probability:          lead.Probability,
		EstimatedValue:       lead.EstimatedValue,
		BudgetConfirmed:      lead.BudgetConfirmed,
		BudgetAmount:         lead.BudgetAmount,
		AuthConfirmed:        lead.AuthConfirmed,
		AuthPerson:           lead.AuthPerson,
		NeedConfirmed:        lead.NeedConfirmed,
		NeedDescription:      lead.NeedDescription,
		TimeConfirmed:        lead.TimeConfirmed,
		AssignedTo:           lead.AssignedTo,
		CustomerID:           lead.CustomerID,
		ContactID:            lead.ContactID,
		DealID:               lead.DealID,
		ConvertedBy:          lead.ConvertedBy,
		Notes:                lead.Notes,
		NPWP:                 lead.NPWP,
		CreatedBy:            lead.CreatedBy,
		CreatedAt:            lead.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
		UpdatedAt:            lead.UpdatedAt.Format("2006-01-02T15:04:05+07:00"),
		Latitude:             lead.Latitude,
		Longitude:            lead.Longitude,
		Rating:               lead.Rating,
		RatingCount:          lead.RatingCount,
		Types:                lead.Types,
		OpeningHours:         lead.OpeningHours,
		ThumbnailURL:         lead.ThumbnailURL,
		CID:                  lead.CID,
		PlaceID:              lead.PlaceID,
		Website:              lead.Website,
		BankAccountID:        lead.BankAccountID,
		BankAccountReference: lead.BankAccountReference,
		BusinessTypeID:       lead.BusinessTypeID,
		AreaID:               lead.AreaID,
		PaymentTermsID:       lead.PaymentTermsID,
	}

	if lead.TimeExpected != nil {
		t := lead.TimeExpected.Format("2006-01-02")
		resp.TimeExpected = &t
	}

	if lead.ConvertedAt != nil {
		t := lead.ConvertedAt.Format("2006-01-02T15:04:05+07:00")
		resp.ConvertedAt = &t
	}

	if lead.LeadSource != nil {
		resp.LeadSource = &dto.LeadSourceInfo{
			ID:   lead.LeadSource.ID,
			Name: lead.LeadSource.Name,
			Code: lead.LeadSource.Code,
		}
	}

	if lead.LeadStatus != nil {
		resp.LeadStatus = &dto.LeadStatusInfo{
			ID:          lead.LeadStatus.ID,
			Name:        lead.LeadStatus.Name,
			Code:        lead.LeadStatus.Code,
			Color:       lead.LeadStatus.Color,
			Score:       lead.LeadStatus.Score,
			IsConverted: lead.LeadStatus.IsConverted,
		}
	}

	if lead.AssignedEmployee != nil {
		resp.AssignedEmployee = &dto.LeadEmployeeInfo{
			ID:           lead.AssignedEmployee.ID,
			EmployeeCode: lead.AssignedEmployee.EmployeeCode,
			Name:         lead.AssignedEmployee.Name,
		}
	}

	if lead.ContactRole != nil {
		resp.ContactRole = &dto.LeadContactRoleInfo{
			ID:         lead.ContactRole.ID,
			Name:       lead.ContactRole.Name,
			Code:       lead.ContactRole.Code,
			BadgeColor: lead.ContactRole.BadgeColor,
		}
	}

	if lead.Customer != nil {
		resp.Customer = &dto.LeadCustomerInfo{
			ID:   lead.Customer.ID,
			Code: lead.Customer.Code,
			Name: lead.Customer.Name,
		}
	}

	if lead.BusinessType != nil {
		resp.BusinessType = &dto.LeadBusinessTypeInfo{
			ID:   lead.BusinessType.ID,
			Name: lead.BusinessType.Name,
		}
	}

	if lead.Area != nil {
		resp.Area = &dto.LeadAreaInfo{
			ID:   lead.Area.ID,
			Name: lead.Area.Name,
		}
	}

	if lead.Deal != nil {
		stageName := ""
		if lead.Deal.PipelineStage != nil {
			stageName = lead.Deal.PipelineStage.Name
		}
		resp.Deal = &dto.LeadDealInfo{
			ID:     lead.Deal.ID,
			Code:   lead.Deal.Code,
			Title:  lead.Deal.Title,
			Status: string(lead.Deal.Status),
			Stage:  stageName,
		}
	}

	if len(lead.Activities) > 0 {
		resp.Activities = ToActivityResponseList(lead.Activities)
	}

	if len(lead.Tasks) > 0 {
		resp.Tasks = ToTaskSummaryResponseList(lead.Tasks)
	}

	if len(lead.ProductItems) > 0 {
		items := make([]dto.LeadProductItemResponse, 0, len(lead.ProductItems))
		for _, item := range lead.ProductItems {
			items = append(items, dto.LeadProductItemResponse{
				ID:                  item.ID,
				LeadID:              item.LeadID,
				ProductID:           item.ProductID,
				ProductName:         item.ProductName,
				ProductSKU:          item.ProductSKU,
				InterestLevel:       item.InterestLevel,
				Quantity:            item.Quantity,
				UnitPrice:           item.UnitPrice,
				Notes:               item.Notes,
				SourceVisitReportID: item.SourceVisitReportID,
				LastSurveyAnswers:   item.LastSurveyAnswers,
				CreatedAt:           item.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
			})
		}
		resp.ProductItems = items
	}

	return resp
}

// ToLeadResponseList converts a slice of Lead models to LeadResponse DTOs
func ToLeadResponseList(leads []models.Lead) []dto.LeadResponse {
	result := make([]dto.LeadResponse, 0, len(leads))
	for i := range leads {
		result = append(result, ToLeadResponse(&leads[i]))
	}
	return result
}
