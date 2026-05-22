package mapper

import (
	"strings"

	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
)

type SalesPaymentMapper struct{}

func NewSalesPaymentMapper() *SalesPaymentMapper {
	return &SalesPaymentMapper{}
}

func (m *SalesPaymentMapper) toInvoiceSummary(si *models.CustomerInvoice) *dto.SalesPaymentInvoiceSummary {
	if si == nil {
		return nil
	}

	var dueDate *string
	if si.DueDate != nil {
		dd := si.DueDate.Format("2006-01-02")
		dueDate = &dd
	}

	return &dto.SalesPaymentInvoiceSummary{
		ID:            si.ID,
		Code:          si.Code,
		InvoiceNumber: si.InvoiceNumber,
		Type:          string(si.Type),
		InvoiceDate:   si.InvoiceDate.Format("2006-01-02"),
		DueDate:       dueDate,
		Amount:        si.Amount,
		Status:        string(si.Status),
	}
}

func (m *SalesPaymentMapper) ToListResponse(p *models.SalesPayment) *dto.SalesPaymentListResponse {
	if p == nil {
		return nil
	}
	var bankSummary *dto.SalesPaymentBankAccountSummary
	if strings.TrimSpace(p.BankAccountNameSnapshot) != "" ||
		strings.TrimSpace(p.BankAccountNumberSnapshot) != "" ||
		strings.TrimSpace(p.BankAccountHolderSnapshot) != "" ||
		strings.TrimSpace(p.BankAccountCurrencySnapshot) != "" {
		bankSummary = &dto.SalesPaymentBankAccountSummary{
			ID:            p.BankAccountID,
			Name:          p.BankAccountNameSnapshot,
			AccountNumber: p.BankAccountNumberSnapshot,
			AccountHolder: p.BankAccountHolderSnapshot,
			Currency:      p.BankAccountCurrencySnapshot,
		}
	} else if p.BankAccount != nil {
		bankSummary = &dto.SalesPaymentBankAccountSummary{
			ID:            p.BankAccount.ID,
			Name:          p.BankAccount.Name,
			AccountNumber: p.BankAccount.AccountNumber,
			AccountHolder: p.BankAccount.AccountHolder,
			Currency:      p.BankAccount.Currency,
		}
	}
	return &dto.SalesPaymentListResponse{
		ID:          p.ID,
		Invoice:     m.toInvoiceSummary(p.CustomerInvoice),
		BankAccount: bankSummary,
		PaymentDate: p.PaymentDate,
		Amount:      p.Amount,
		TenderAmount: p.TenderAmount,
		ChangeAmount: p.ChangeAmount,
		Method:      string(p.Method),
		Status:      string(p.Status),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func (m *SalesPaymentMapper) ToListResponseList(items []*models.SalesPayment) []*dto.SalesPaymentListResponse {
	out := make([]*dto.SalesPaymentListResponse, 0, len(items))
	for _, it := range items {
		out = append(out, m.ToListResponse(it))
	}
	return out
}

func (m *SalesPaymentMapper) ToDetailResponse(p *models.SalesPayment) *dto.SalesPaymentDetailResponse {
	if p == nil {
		return nil
	}
	base := m.ToListResponse(p)
	return &dto.SalesPaymentDetailResponse{
		SalesPaymentListResponse: *base,
		ReferenceNumber:          p.ReferenceNumber,
		Notes:                    p.Notes,
	}
}
