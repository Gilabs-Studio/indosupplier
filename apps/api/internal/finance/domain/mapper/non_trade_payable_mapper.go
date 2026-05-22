package mapper

import (
	"strings"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

type NonTradePayableMapper struct{}

func NewNonTradePayableMapper() *NonTradePayableMapper {
	return &NonTradePayableMapper{}
}

func (m *NonTradePayableMapper) ToResponse(item *financeModels.NonTradePayable) dto.NonTradePayableResponse {
	if item == nil {
		return dto.NonTradePayableResponse{}
	}

	var coaResp *dto.ChartOfAccountResponse
	if strings.TrimSpace(item.ChartOfAccountCodeSnapshot) != "" ||
		strings.TrimSpace(item.ChartOfAccountNameSnapshot) != "" ||
		strings.TrimSpace(item.ChartOfAccountTypeSnapshot) != "" {
		coaResp = &dto.ChartOfAccountResponse{
			ID:        item.ChartOfAccountID,
			Code:      strings.TrimSpace(item.ChartOfAccountCodeSnapshot),
			Name:      strings.TrimSpace(item.ChartOfAccountNameSnapshot),
			Type:      financeModels.AccountType(strings.TrimSpace(item.ChartOfAccountTypeSnapshot)),
			ParentID:  nil,
			IsActive:  true,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		}
	} else if item.ChartOfAccount != nil {
		mapped := (&ChartOfAccountMapper{}).ToResponse(item.ChartOfAccount)
		coaResp = &mapped
	}

	return dto.NonTradePayableResponse{
		ID:               item.ID,
		TransactionDate:  item.TransactionDate,
		Code:             item.Code,
		Description:      item.Description,
		ChartOfAccountID: item.ChartOfAccountID,
		ChartOfAccount:   coaResp,
		Amount:           item.Amount,
		VendorName:       item.VendorName,
		DueDate:          item.DueDate,
		ReferenceNo:      item.ReferenceNo,
		Status:           string(item.Status),
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}
