package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/loyalty/data/models"
)

type MemberListParams struct {
	Page    int
	PerPage int
	Tier     string
	OutletID string
	// ProgramOutletIDs filters members to those whose program belongs to one of the
	// given outlet IDs (or is a global program with outlet_id IS NULL).
	// When nil, no program-scope filtering is applied (admin view).
	ProgramOutletIDs []string
	// Search filters members by customer name or member code (case-insensitive).
	Search string
}

type LoyaltyMemberRepository interface {
	Create(ctx context.Context, member *models.LoyaltyMember) error
	GetByID(ctx context.Context, id string) (*models.LoyaltyMember, error)
	GetByCustomerID(ctx context.Context, customerID string) (*models.LoyaltyMember, error)
	// LookupByPhoneOrName performs a case-insensitive name match + optional phone lookup
	// via the customers table to find an existing member.
	LookupByNameAndOutlet(ctx context.Context, name, outletID string) (*models.LoyaltyMember, error)
	List(ctx context.Context, params MemberListParams) ([]models.LoyaltyMember, int64, error)
	Update(ctx context.Context, member *models.LoyaltyMember) error
	GetNextMemberCode(ctx context.Context) (string, error)
}
