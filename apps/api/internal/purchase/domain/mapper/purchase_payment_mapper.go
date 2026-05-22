package mapper

import (
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
)

type PurchasePaymentMapper struct{}

// formatDateToISO formats a time.Time to ISO8601 date string (YYYY-MM-DD)
func formatDateToISO_PurchasePayment(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func NewPurchasePaymentMapper() *PurchasePaymentMapper {
	return &PurchasePaymentMapper{}
}

func (m *PurchasePaymentMapper) toInvoiceSummary(si *models.SupplierInvoice) *dto.PurchasePaymentInvoiceSummary {
	if si == nil {
		return nil
	}
	return &dto.PurchasePaymentInvoiceSummary{
		ID:              si.ID,
		Code:            si.Code,
		InvoiceNumber:   si.InvoiceNumber,
		InvoiceDate:     formatDateToISO_PurchasePayment(si.InvoiceDate),
		DueDate:         formatDateToISO_PurchasePayment(si.DueDate),
		TaxRate:         si.TaxRate,
		TaxAmount:       si.TaxAmount,
		Amount:          si.Amount,
		RemainingAmount: si.RemainingAmount,
		Status:          string(si.Status),
		Notes:           si.Notes,
	}
}

func (m *PurchasePaymentMapper) ToListResponse(p *models.PurchasePayment) *dto.PurchasePaymentListResponse {
	if p == nil {
		return nil
	}
	var bankSummary *dto.PurchasePaymentBankAccountSummary
	if strings.TrimSpace(p.BankAccountNameSnapshot) != "" ||
		strings.TrimSpace(p.BankAccountNumberSnapshot) != "" ||
		strings.TrimSpace(p.BankAccountHolderSnapshot) != "" ||
		strings.TrimSpace(p.BankAccountCurrencySnapshot) != "" {
		id := ""
		if p.BankAccountID != nil {
			id = strings.TrimSpace(*p.BankAccountID)
		}
		bankSummary = &dto.PurchasePaymentBankAccountSummary{
			ID:            id,
			Name:          p.BankAccountNameSnapshot,
			AccountNumber: p.BankAccountNumberSnapshot,
			AccountHolder: p.BankAccountHolderSnapshot,
			Currency:      p.BankAccountCurrencySnapshot,
		}
	} else if p.BankAccount != nil {
		bankSummary = &dto.PurchasePaymentBankAccountSummary{
			ID:            p.BankAccount.ID,
			Name:          p.BankAccount.Name,
			AccountNumber: p.BankAccount.AccountNumber,
			AccountHolder: p.BankAccount.AccountHolder,
			Currency:      p.BankAccount.Currency,
		}
	}
	return &dto.PurchasePaymentListResponse{
		ID:           p.ID,
		CompanyID:    &p.CompanyID,
		FiscalYearID: p.FiscalYearID,
		Invoice:      m.toInvoiceSummary(p.SupplierInvoice),
		BankAccount:  bankSummary,
		PaymentDate:  formatDateToISO_PurchasePayment(p.PaymentDate),
		Amount:       p.Amount,
		Method:       string(p.Method),
		Status:       string(p.Status),
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func (m *PurchasePaymentMapper) ToListResponseList(items []*models.PurchasePayment) []*dto.PurchasePaymentListResponse {
	out := make([]*dto.PurchasePaymentListResponse, 0, len(items))
	for _, it := range items {
		out = append(out, m.ToListResponse(it))
	}
	return out
}

func (m *PurchasePaymentMapper) ToDetailResponse(p *models.PurchasePayment) *dto.PurchasePaymentDetailResponse {
	if p == nil {
		return nil
	}
	base := m.ToListResponse(p)
	return &dto.PurchasePaymentDetailResponse{
		PurchasePaymentListResponse: *base,
		ReferenceNumber:             p.ReferenceNumber,
		Notes:                       p.Notes,
		CashBankTransactionID:       p.CashBankTransactionID,
	}
}
