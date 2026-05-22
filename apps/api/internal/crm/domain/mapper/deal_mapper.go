package mapper

import (
	"strings"

	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

// ToDealResponse converts a Deal model to DealResponse DTO
func ToDealResponse(deal *models.Deal) dto.DealResponse {
	resp := dto.DealResponse{
		ID:                   deal.ID,
		Code:                 deal.Code,
		Title:                deal.Title,
		Description:          deal.Description,
		Status:               string(deal.Status),
		PipelineStageID:      deal.PipelineStageID,
		Value:                deal.Value,
		Probability:          deal.Probability,
		CloseReason:          deal.CloseReason,
		CustomerID:           deal.CustomerID,
		ContactID:            deal.ContactID,
		AssignedTo:           deal.AssignedTo,
		LeadID:               deal.LeadID,
		BudgetConfirmed:      deal.BudgetConfirmed,
		BudgetAmount:         deal.BudgetAmount,
		AuthConfirmed:        deal.AuthConfirmed,
		AuthPerson:           deal.AuthPerson,
		NeedConfirmed:        deal.NeedConfirmed,
		NeedDescription:      deal.NeedDescription,
		TimeConfirmed:        deal.TimeConfirmed,
		Notes:                deal.Notes,
		CreatedBy:            deal.CreatedBy,
		CreatedAt:            deal.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
		UpdatedAt:            deal.UpdatedAt.Format("2006-01-02T15:04:05+07:00"),
	}

	if deal.ExpectedCloseDate != nil {
		t := deal.ExpectedCloseDate.Format("2006-01-02")
		resp.ExpectedCloseDate = &t
	}
	if deal.ActualCloseDate != nil {
		t := deal.ActualCloseDate.Format("2006-01-02")
		resp.ActualCloseDate = &t
	}

	// Conversion tracking
	resp.ConvertedToQuotationID = deal.ConvertedToQuotationID
	if deal.ConvertedAt != nil {
		t := deal.ConvertedAt.Format("2006-01-02T15:04:05+07:00")
		resp.ConvertedAt = &t
	}

	if deal.PipelineStage != nil {
		resp.PipelineStage = &dto.DealPipelineStageInfo{
			ID:          deal.PipelineStage.ID,
			Name:        deal.PipelineStage.Name,
			Code:        deal.PipelineStage.Code,
			Color:       deal.PipelineStage.Color,
			Order:       deal.PipelineStage.Order,
			Probability: deal.PipelineStage.Probability,
			IsWon:       deal.PipelineStage.IsWon,
			IsLost:      deal.PipelineStage.IsLost,
		}
	}

	if deal.Customer != nil {
		resp.Customer = &dto.DealCustomerInfo{
			ID:   deal.Customer.ID,
			Code: deal.Customer.Code,
			Name: deal.Customer.Name,
		}
	}

	// If no customer linked but the deal has lead info, expose a snapshot
	// of the potential customer (no customer ID) so UI can display prospect name.
	if resp.Customer == nil && deal.Lead != nil {
		name := deal.Lead.CompanyName
		if name == "" {
			name = strings.TrimSpace(deal.Lead.FirstName + " " + deal.Lead.LastName)
		}
		if name != "" {
			resp.Customer = &dto.DealCustomerInfo{
				ID:   "",
				Code: deal.Lead.Code,
				Name: name,
			}
		}
	}

	if deal.Contact != nil {
		resp.Contact = &dto.DealContactInfo{
			ID:    deal.Contact.ID,
			Name:  deal.Contact.Name,
			Phone: deal.Contact.Phone,
			Email: deal.Contact.Email,
		}
	}

	// If no contact linked but lead has person info, expose a snapshot contact
	if resp.Contact == nil && deal.Lead != nil {
		contactName := strings.TrimSpace(deal.Lead.FirstName + " " + deal.Lead.LastName)
		if contactName == "" && deal.Lead.CompanyName != "" {
			// fallback to company as contact name when person not provided
			contactName = deal.Lead.CompanyName
		}
		if contactName != "" || deal.Lead.Phone != "" || deal.Lead.Email != "" {
			resp.Contact = &dto.DealContactInfo{
				ID:    "",
				Name:  contactName,
				Phone: deal.Lead.Phone,
				Email: deal.Lead.Email,
			}
		}
	}

	if deal.AssignedEmployee != nil {
		resp.AssignedEmployee = &dto.DealEmployeeInfo{
			ID:           deal.AssignedEmployee.ID,
			EmployeeCode: deal.AssignedEmployee.EmployeeCode,
			Name:         deal.AssignedEmployee.Name,
		}
	}

	if deal.Lead != nil {
		resp.Lead = &dto.DealLeadInfo{
			ID:          deal.Lead.ID,
			Code:        deal.Lead.Code,
			FirstName:   deal.Lead.FirstName,
			LastName:    deal.Lead.LastName,
			CompanyName: deal.Lead.CompanyName,
			Phone:       deal.Lead.Phone,
			Email:       deal.Lead.Email,
			Address:     deal.Lead.Address,
			City:        deal.Lead.City,
			Province:    deal.Lead.Province,
			Latitude:    deal.Lead.Latitude,
			Longitude:   deal.Lead.Longitude,
		}
	}

	// Map product items
	resp.Items = make([]dto.DealProductItemResponse, 0, len(deal.Items))
	for _, item := range deal.Items {
		resp.Items = append(resp.Items, dto.DealProductItemResponse{
			ID:              item.ID,
			DealID:          item.DealID,
			ProductID:       item.ProductID,
			ProductName:     item.ProductName,
			ProductSKU:      item.ProductSKU,
			UnitPrice:       item.UnitPrice,
			Quantity:        item.Quantity,
			DiscountPercent: item.DiscountPercent,
			DiscountAmount:  item.DiscountAmount,
			Subtotal:        item.Subtotal,
			Notes:           item.Notes,
			InterestLevel:   item.InterestLevel,
			IsDeleted:       item.DeletedAt.Valid,
		})
	}

	if len(deal.Tasks) > 0 {
		resp.Tasks = ToTaskSummaryResponseList(deal.Tasks)
	}

	return resp
}

// ToDealResponseList converts a slice of Deal models to DealResponse DTOs
func ToDealResponseList(deals []models.Deal) []dto.DealResponse {
	result := make([]dto.DealResponse, 0, len(deals))
	for i := range deals {
		result = append(result, ToDealResponse(&deals[i]))
	}
	return result
}

// ToDealHistoryResponse converts a DealHistory model to DealHistoryResponse DTO
func ToDealHistoryResponse(h *models.DealHistory) dto.DealHistoryResponse {
	resp := dto.DealHistoryResponse{
		ID:              h.ID,
		DealID:          h.DealID,
		FromStageID:     h.FromStageID,
		FromStageName:   h.FromStageName,
		ToStageID:       h.ToStageID,
		ToStageName:     h.ToStageName,
		FromProbability: h.FromProbability,
		ToProbability:   h.ToProbability,
		DaysInPrevStage: h.DaysInPrevStage,
		ChangedBy:       h.ChangedBy,
		ChangedAt:       h.ChangedAt.Format("2006-01-02T15:04:05+07:00"),
		Reason:          h.Reason,
		Notes:           h.Notes,
	}

	if h.FromStage != nil {
		resp.FromStage = &dto.DealPipelineStageInfo{
			ID:          h.FromStage.ID,
			Name:        h.FromStage.Name,
			Code:        h.FromStage.Code,
			Color:       h.FromStage.Color,
			Order:       h.FromStage.Order,
			Probability: h.FromStage.Probability,
			IsWon:       h.FromStage.IsWon,
			IsLost:      h.FromStage.IsLost,
		}
	}

	if h.ToStage != nil {
		resp.ToStage = &dto.DealPipelineStageInfo{
			ID:          h.ToStage.ID,
			Name:        h.ToStage.Name,
			Code:        h.ToStage.Code,
			Color:       h.ToStage.Color,
			Order:       h.ToStage.Order,
			Probability: h.ToStage.Probability,
			IsWon:       h.ToStage.IsWon,
			IsLost:      h.ToStage.IsLost,
		}
	}

	if h.ChangedByEmployee != nil {
		resp.ChangedByUser = &dto.DealEmployeeInfo{
			ID:           h.ChangedByEmployee.ID,
			EmployeeCode: h.ChangedByEmployee.EmployeeCode,
			Name:         h.ChangedByEmployee.Name,
		}
	}

	return resp
}

// ToDealHistoryResponseList converts a slice of DealHistory models to DTOs
func ToDealHistoryResponseList(history []models.DealHistory) []dto.DealHistoryResponse {
	result := make([]dto.DealHistoryResponse, 0, len(history))
	for i := range history {
		result = append(result, ToDealHistoryResponse(&history[i]))
	}
	return result
}

// ToPipelineSummaryResponse converts PipelineSummaryData to DTO
func ToPipelineSummaryResponse(data *repositories.PipelineSummaryData) dto.DealPipelineSummaryResponse {
	resp := dto.DealPipelineSummaryResponse{
		TotalDeals: data.TotalDeals,
		TotalValue: data.TotalValue,
		OpenDeals:  data.OpenDeals,
		OpenValue:  data.OpenValue,
		WonDeals:   data.WonDeals,
		WonValue:   data.WonValue,
		LostDeals:  data.LostDeals,
		LostValue:  data.LostValue,
	}

	resp.ByStage = make([]dto.DealStageSummaryResponse, 0, len(data.ByStage))
	for _, s := range data.ByStage {
		resp.ByStage = append(resp.ByStage, dto.DealStageSummaryResponse{
			StageID:    s.StageID,
			StageName:  s.StageName,
			StageColor: s.StageColor,
			DealCount:  s.DealCount,
			TotalValue: s.TotalValue,
		})
	}

	return resp
}

// ToForecastResponse converts ForecastData to DTO
func ToForecastResponse(data *repositories.ForecastData) dto.DealForecastResponse {
	resp := dto.DealForecastResponse{
		TotalWeightedValue: data.TotalWeightedValue,
		TotalDeals:         data.TotalDeals,
	}

	resp.ByStage = make([]dto.DealStageForecastResponse, 0, len(data.ByStage))
	for _, s := range data.ByStage {
		resp.ByStage = append(resp.ByStage, dto.DealStageForecastResponse{
			StageID:       s.StageID,
			StageName:     s.StageName,
			DealCount:     s.DealCount,
			TotalValue:    s.TotalValue,
			Probability:   s.Probability,
			WeightedValue: s.WeightedValue,
		})
	}

	return resp
}
