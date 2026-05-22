package mapper

import (
	"strings"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

type PaymentMapper struct {
	coaMapper *ChartOfAccountMapper
}

func NewPaymentMapper(coaMapper *ChartOfAccountMapper) *PaymentMapper {
	return &PaymentMapper{coaMapper: coaMapper}
}

func (m *PaymentMapper) ToResponse(item *financeModels.Payment) dto.PaymentResponse {
	if item == nil {
		return dto.PaymentResponse{}
	}

	resp := dto.PaymentResponse{
		ID:             item.ID,
		PaymentDate:    item.PaymentDate,
		Description:    item.Description,
		BankAccountID:  item.BankAccountID,
		TotalAmount:    item.TotalAmount,
		Status:         item.Status,
		JournalEntryID: item.JournalEntryID,
		ApprovedAt:     item.ApprovedAt,
		ApprovedBy:     item.ApprovedBy,
		PostedAt:       item.PostedAt,
		PostedBy:       item.PostedBy,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}

	if strings.TrimSpace(item.BankAccountNameSnapshot) != "" ||
		strings.TrimSpace(item.BankAccountNumberSnapshot) != "" ||
		strings.TrimSpace(item.BankAccountHolderSnapshot) != "" ||
		strings.TrimSpace(item.BankAccountCurrencySnapshot) != "" {
		resp.BankAccount = &dto.BankAccountMini{
			ID:            item.BankAccountID,
			Name:          strings.TrimSpace(item.BankAccountNameSnapshot),
			AccountNumber: strings.TrimSpace(item.BankAccountNumberSnapshot),
			AccountHolder: strings.TrimSpace(item.BankAccountHolderSnapshot),
			Currency:      strings.TrimSpace(item.BankAccountCurrencySnapshot),
		}
	}

	if len(item.Allocations) > 0 {
		resp.Allocations = make([]dto.PaymentAllocationResponse, 0, len(item.Allocations))
		for _, al := range item.Allocations {
			var coaResp *dto.ChartOfAccountResponse
			if strings.TrimSpace(al.ChartOfAccountCodeSnapshot) != "" || strings.TrimSpace(al.ChartOfAccountNameSnapshot) != "" || strings.TrimSpace(al.ChartOfAccountTypeSnapshot) != "" {
				coaResp = &dto.ChartOfAccountResponse{
					ID:        al.ChartOfAccountID,
					Code:      strings.TrimSpace(al.ChartOfAccountCodeSnapshot),
					Name:      strings.TrimSpace(al.ChartOfAccountNameSnapshot),
					Type:      financeModels.AccountType(strings.TrimSpace(al.ChartOfAccountTypeSnapshot)),
					ParentID:  nil,
					IsActive:  true,
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}
			} else if al.ChartOfAccount != nil {
				mapped := m.coaMapper.ToResponse(al.ChartOfAccount)
				coaResp = &mapped
			}
			resp.Allocations = append(resp.Allocations, dto.PaymentAllocationResponse{
				ID:              al.ID,
				ChartOfAccountID: al.ChartOfAccountID,
				ChartOfAccount:  coaResp,
				ReferenceType:   al.ReferenceType,
				ReferenceID:     al.ReferenceID,
				Amount:          al.Amount,
				Memo:            al.Memo,
				CreatedAt:       al.CreatedAt,
				UpdatedAt:       al.UpdatedAt,
			})
		}
	}

	return resp
}
