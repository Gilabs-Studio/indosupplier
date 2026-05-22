package usecase

import (
	"testing"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
)

func TestValuationTotals_ShouldSplitGainAndLoss(t *testing.T) {
	items := []ValuationItem{
		{Delta: 120.5},
		{Delta: -30.25},
		{Delta: 0},
		{Delta: -10.75},
	}

	gain := totalGain(items)
	loss := totalLoss(items)

	if gain != 120.5 {
		t.Fatalf("expected total gain 120.5, got %v", gain)
	}
	if loss != 41.0 {
		t.Fatalf("expected total loss 41.0, got %v", loss)
	}
}

func TestMapValuationToRefType_ShouldReturnExpectedReferenceType(t *testing.T) {
	tests := []struct {
		name          string
		valuationType string
		expected      string
	}{
		{name: "inventory", valuationType: "inventory", expected: reference.RefTypeInventoryValuation},
		{name: "fx", valuationType: "fx", expected: reference.RefTypeCurrencyRevaluation},
		{name: "depreciation", valuationType: "depreciation", expected: reference.RefTypeDepreciationValuation},
		{name: "fallback", valuationType: "unknown", expected: reference.RefTypeInventoryValuation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := mapValuationToRefType(tt.valuationType)
			if actual != tt.expected {
				t.Fatalf("expected ref type %s, got %s", tt.expected, actual)
			}
		})
	}
}

func TestDetailsToResult_ShouldAggregateItemsAndDelta(t *testing.T) {
	productID := "00000000-0000-0000-0000-000000000001"
	details := []financeModels.ValuationRunDetail{
		{
			ReferenceID: "REF-1",
			ProductID:   &productID,
			Qty:         2,
			BookValue:   100,
			ActualValue: 120,
			Delta:       20,
			Direction:   financeModels.ValuationDirectionGain,
		},
		{
			ReferenceID: "REF-2",
			Qty:         1,
			BookValue:   80,
			ActualValue: 70,
			Delta:       -10,
			Direction:   financeModels.ValuationDirectionLoss,
		},
	}

	result := detailsToResult(details)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.TotalDelta != 10 {
		t.Fatalf("expected total delta 10, got %v", result.TotalDelta)
	}
}
