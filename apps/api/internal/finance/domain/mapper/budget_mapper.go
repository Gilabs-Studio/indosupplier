package mapper

import (
	"strings"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

type BudgetMapper struct {
	coaMapper *ChartOfAccountMapper
}

func NewBudgetMapper(coaMapper *ChartOfAccountMapper) *BudgetMapper {
	return &BudgetMapper{coaMapper: coaMapper}
}

func (m *BudgetMapper) ToResponse(item *financeModels.Budget) dto.BudgetResponse {
	if item == nil {
		return dto.BudgetResponse{}
	}

	resp := dto.BudgetResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		StartDate:   item.StartDate,
		EndDate:     item.EndDate,
		TotalAmount: item.TotalAmount,
		Status:      item.Status,
		ApprovedAt:  item.ApprovedAt,
		ApprovedBy:  item.ApprovedBy,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}

	if len(item.Items) > 0 {
		resp.Items = make([]dto.BudgetItemResponse, 0, len(item.Items))
		for _, it := range item.Items {
			var coaResp *dto.ChartOfAccountResponse
			if strings.TrimSpace(it.ChartOfAccountCodeSnapshot) != "" || strings.TrimSpace(it.ChartOfAccountNameSnapshot) != "" || strings.TrimSpace(it.ChartOfAccountTypeSnapshot) != "" {
				coaResp = &dto.ChartOfAccountResponse{
					ID:        it.ChartOfAccountID,
					Code:      strings.TrimSpace(it.ChartOfAccountCodeSnapshot),
					Name:      strings.TrimSpace(it.ChartOfAccountNameSnapshot),
					Type:      financeModels.AccountType(strings.TrimSpace(it.ChartOfAccountTypeSnapshot)),
					ParentID:  nil,
					IsActive:  true,
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}
			} else if it.ChartOfAccount != nil {
				mapped := m.coaMapper.ToResponse(it.ChartOfAccount)
				coaResp = &mapped
			}
			resp.Items = append(resp.Items, dto.BudgetItemResponse{
				ID:               it.ID,
				ChartOfAccountID: it.ChartOfAccountID,
				ChartOfAccount:   coaResp,
				Amount:           it.Amount,
				ActualAmount:     it.ActualAmount,
				Memo:             it.Memo,
				CreatedAt:        it.CreatedAt,
				UpdatedAt:        it.UpdatedAt,
			})
		}
	}

	return resp
}
