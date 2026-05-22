package mapper

import (
	"testing"
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

func TestCashBankJournalMapper_ToResponse_ShouldBuildReferenceFromPaymentLine(t *testing.T) {
	m := NewCashBankJournalMapper(NewChartOfAccountMapper())

	refType := "PAYMENT"
	refID := "7fd6d5ae-5a4f-4db8-a04f-25922a5a1e6e"
	item := &financeModels.CashBankJournal{
		ID:              "e8ddf5a2-2e4b-4fcd-aafa-86e66da0e121",
		TransactionDate: time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
		Type:            financeModels.CashBankTypeCashOut,
		Status:          financeModels.CashBankStatusPosted,
		Lines: []financeModels.CashBankJournalLine{
			{
				ReferenceType: &refType,
				ReferenceID:   &refID,
			},
		},
	}

	res := m.ToResponse(item)
	if res.ReferenceType != "PAY" {
		t.Fatalf("expected PAY reference_type, got %s", res.ReferenceType)
	}
	if res.ReferenceID != refID {
		t.Fatalf("expected reference_id %s, got %s", refID, res.ReferenceID)
	}
	if res.ReferenceCode != "PAY-7FD6D5AE" {
		t.Fatalf("expected reference_code PAY-7FD6D5AE, got %s", res.ReferenceCode)
	}
}

func TestCashBankJournalMapper_ToResponse_ShouldDefaultTransferReference(t *testing.T) {
	m := NewCashBankJournalMapper(NewChartOfAccountMapper())

	item := &financeModels.CashBankJournal{
		ID:              "a8d95f0f-299a-4b77-9f26-efca39a15be0",
		TransactionDate: time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC),
		Type:            financeModels.CashBankTypeTransfer,
		Status:          financeModels.CashBankStatusPosted,
	}

	res := m.ToResponse(item)
	if res.ReferenceType != "TRF" {
		t.Fatalf("expected TRF reference_type, got %s", res.ReferenceType)
	}
	if res.ReferenceID != item.ID {
		t.Fatalf("expected reference_id %s, got %s", item.ID, res.ReferenceID)
	}
	if res.ReferenceCode != "TRF-A8D95F0F" {
		t.Fatalf("expected reference_code TRF-A8D95F0F, got %s", res.ReferenceCode)
	}
}
