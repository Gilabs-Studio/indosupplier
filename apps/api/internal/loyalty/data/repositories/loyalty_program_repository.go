package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/loyalty/data/models"
)

type LoyaltyProgramRepository interface {
	Create(ctx context.Context, program *models.LoyaltyProgram) error
	GetByID(ctx context.Context, id string) (*models.LoyaltyProgram, error)
	List(ctx context.Context, page, perPage int, search string) ([]models.LoyaltyProgram, int64, error)
	// ListWithOutletFilter returns programs visible to an outlet manager:
	// global programs (outlet_id IS NULL) plus programs owned by the given outlet IDs.
	ListWithOutletFilter(ctx context.Context, outletIDs []string, page, perPage int, search string) ([]models.LoyaltyProgram, int64, error)
	// FindActiveForOutlet returns the active program for a given outlet,
	// falling back to the global program (outlet_id IS NULL) if no outlet-specific one is found.
	FindActiveForOutlet(ctx context.Context, outletID string) (*models.LoyaltyProgram, error)
	Update(ctx context.Context, program *models.LoyaltyProgram) error
	Delete(ctx context.Context, id string) error
}
