package mapper

import (
	"fmt"
	"strings"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

type JournalEntryMapper struct {
	coaMapper *ChartOfAccountMapper
}

func buildJournalReferenceCode(referenceType, referenceID *string) *string {
	if referenceType == nil || strings.TrimSpace(*referenceType) == "" {
		return nil
	}

	typeCode := strings.ToUpper(strings.TrimSpace(*referenceType))
	prefix := "REF"
	switch {
	case strings.Contains(typeCode, "SALES_INVOICE"):
		prefix = "INV"
	case strings.Contains(typeCode, "SUPPLIER_INVOICE"):
		prefix = "PINV"
	case strings.Contains(typeCode, "PAYMENT"):
		prefix = "PAY"
	case strings.Contains(typeCode, "CASH_BANK"):
		prefix = "CB"
	case strings.Contains(typeCode, "ADJUST"):
		prefix = "ADJ"
	case strings.Contains(typeCode, "VALUATION"):
		prefix = "VAL"
	default:
		if len(typeCode) >= 3 {
			prefix = typeCode[:3]
		}
	}

	suffix := "N/A"
	if referenceID != nil {
		trimmed := strings.TrimSpace(*referenceID)
		if trimmed != "" {
			segments := strings.Split(trimmed, "-")
			suffix = strings.ToUpper(segments[0])
			if len(suffix) > 10 {
				suffix = suffix[:10]
			}
		}
	}

	code := fmt.Sprintf("%s-%s", prefix, suffix)
	return &code
}

// BuildJournalReferenceCodeForExport exposes the synthetic reference code used when no source document code is found.
func BuildJournalReferenceCodeForExport(referenceType, referenceID *string) *string {
	return buildJournalReferenceCode(referenceType, referenceID)
}

func NewJournalEntryMapper(coaMapper *ChartOfAccountMapper) *JournalEntryMapper {
	return &JournalEntryMapper{coaMapper: coaMapper}
}

func (m *JournalEntryMapper) ToResponse(item *financeModels.JournalEntry) dto.JournalEntryResponse {
	if item == nil {
		return dto.JournalEntryResponse{}
	}

	lines := make([]dto.JournalLineResponse, 0, len(item.Lines))
	var calcDebit float64
	var calcCredit float64
	for _, ln := range item.Lines {
		calcDebit += ln.Debit
		calcCredit += ln.Credit

		var coaResp *dto.ChartOfAccountResponse
		if strings.TrimSpace(ln.ChartOfAccountCodeSnapshot) != "" || strings.TrimSpace(ln.ChartOfAccountNameSnapshot) != "" || strings.TrimSpace(ln.ChartOfAccountTypeSnapshot) != "" {
			coaResp = &dto.ChartOfAccountResponse{
				ID:        ln.ChartOfAccountID,
				Code:      ln.ChartOfAccountCodeSnapshot,
				Name:      ln.ChartOfAccountNameSnapshot,
				Type:      financeModels.AccountType(ln.ChartOfAccountTypeSnapshot),
				ParentID:  nil,
				IsActive:  true,
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			}
		} else if ln.ChartOfAccount != nil {
			v := m.coaMapper.ToResponse(ln.ChartOfAccount)
			coaResp = &v
		}
		lines = append(lines, dto.JournalLineResponse{
			ID:               ln.ID,
			ChartOfAccountID: ln.ChartOfAccountID,
			ChartOfAccount:   coaResp,
			Debit:            ln.Debit,
			Credit:           ln.Credit,
			Memo:             ln.Memo,
		})
	}

	// Task 7: Fallback for totals if empty (legacy data)
	finalDebit := item.DebitTotal
	finalCredit := item.CreditTotal
	if finalDebit == 0 && finalCredit == 0 && (calcDebit != 0 || calcCredit != 0) {
		finalDebit = calcDebit
		finalCredit = calcCredit
	}

	res := dto.JournalEntryResponse{
		ID:                item.ID,
		CompanyID:         item.CompanyID,
		FiscalYearID:      item.FiscalYearID,
		JournalNumber:     item.JournalNumber,
		EntryDate:         item.EntryDate,
		Reference:         item.Reference,
		Description:       item.Description,
		ReferenceType:     item.ReferenceType,
		ReferenceID:       item.ReferenceID,
		ReferenceCode:     buildJournalReferenceCode(item.ReferenceType, item.ReferenceID),
		Status:            item.Status,
		JournalType:       string(item.JournalType),
		PostedAt:          item.PostedAt,
		PostedBy:          item.PostedBy,
		CreatedBy:         item.CreatedBy,
		ReversedAt:        item.ReversedAt,
		ReversedBy:        item.ReversedBy,
		ReversalReason:    item.ReversalReason,
		IsSystemGenerated: item.IsSystemGenerated,
		SourceDocumentURL: item.SourceDocumentURL,
		Lines:             lines,
		DebitTotal:        finalDebit,
		CreditTotal:       finalCredit,
		CurrencyCode:      item.CurrencyCode,
		ExchangeRate:      item.ExchangeRate,
		IsReversal:        item.IsReversal,
		ReversedFrom:      item.ReversedFrom,
		IsValuation:       item.IsValuation,
		Source:            string(item.Source),
		ValuationRunID:    item.ValuationRunID,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}

	// Task 1: Audit names
	if item.CreatedByUser != nil {
		res.CreatedByName = &item.CreatedByUser.Name
		res.CreatedByEmail = &item.CreatedByUser.Email
	}
	if item.PostedByUser != nil {
		res.PostedByName = &item.PostedByUser.Name
		res.PostedByEmail = &item.PostedByUser.Email
	}
	if item.ReversedByUser != nil {
		res.ReversedByName = &item.ReversedByUser.Name
		res.ReversedByEmail = &item.ReversedByUser.Email
	}

	return res
}

func (m *JournalEntryMapper) ToSummaryResponse(item *financeModels.JournalEntry) dto.JournalEntryResponse {
	if item == nil {
		return dto.JournalEntryResponse{}
	}

	var calcDebit float64
	var calcCredit float64
	for _, ln := range item.Lines {
		calcDebit += ln.Debit
		calcCredit += ln.Credit
	}

	// Task 7: Fallback for totals if empty (legacy data)
	finalDebit := item.DebitTotal
	finalCredit := item.CreditTotal
	if finalDebit == 0 && finalCredit == 0 && (calcDebit != 0 || calcCredit != 0) {
		finalDebit = calcDebit
		finalCredit = calcCredit
	}

	res := dto.JournalEntryResponse{
		ID:                item.ID,
		CompanyID:         item.CompanyID,
		FiscalYearID:      item.FiscalYearID,
		JournalNumber:     item.JournalNumber,
		EntryDate:         item.EntryDate,
		Reference:         item.Reference,
		Description:       item.Description,
		ReferenceType:     item.ReferenceType,
		ReferenceID:       item.ReferenceID,
		ReferenceCode:     buildJournalReferenceCode(item.ReferenceType, item.ReferenceID),
		Status:            item.Status,
		JournalType:       string(item.JournalType),
		PostedAt:          item.PostedAt,
		PostedBy:          item.PostedBy,
		CreatedBy:         item.CreatedBy,
		ReversedAt:        item.ReversedAt,
		ReversedBy:        item.ReversedBy,
		ReversalReason:    item.ReversalReason,
		IsSystemGenerated: item.IsSystemGenerated,
		SourceDocumentURL: item.SourceDocumentURL,
		Lines:             nil, // Summary should not include lines for performance
		DebitTotal:        finalDebit,
		CreditTotal:       finalCredit,
		CurrencyCode:      item.CurrencyCode,
		ExchangeRate:      item.ExchangeRate,
		IsReversal:        item.IsReversal,
		ReversedFrom:      item.ReversedFrom,
		IsValuation:       item.IsValuation,
		Source:            string(item.Source),
		ValuationRunID:    item.ValuationRunID,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}

	// Task 1: Audit names
	if item.CreatedByUser != nil {
		res.CreatedByName = &item.CreatedByUser.Name
		res.CreatedByEmail = &item.CreatedByUser.Email
	}
	if item.PostedByUser != nil {
		res.PostedByName = &item.PostedByUser.Name
		res.PostedByEmail = &item.PostedByUser.Email
	}
	if item.ReversedByUser != nil {
		res.ReversedByName = &item.ReversedByUser.Name
		res.ReversedByEmail = &item.ReversedByUser.Email
	}

	return res
}
