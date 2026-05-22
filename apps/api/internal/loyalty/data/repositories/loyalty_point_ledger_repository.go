package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/loyalty/data/models"
)

type LedgerListParams struct {
	MemberID string
	Page     int
	PerPage  int
}

type LoyaltyPointLedgerRepository interface {
	Create(ctx context.Context, entry *models.LoyaltyPointLedger) error
	ListByMember(ctx context.Context, params LedgerListParams) ([]models.LoyaltyPointLedger, int64, error)
	// FindByTransaction checks whether points have already been awarded for a given transaction.
	FindByTransaction(ctx context.Context, memberID, transactionID, entryType string) (*models.LoyaltyPointLedger, error)
}
