package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/hrd/data/models"
)

// RecruitmentRequestRepository defines the interface for recruitment request data operations
type RecruitmentRequestRepository interface {
	// Basic CRUD
	Create(ctx context.Context, req *models.RecruitmentRequest) error
	FindByID(ctx context.Context, id string) (*models.RecruitmentRequest, error)
	Update(ctx context.Context, req *models.RecruitmentRequest) error
	Delete(ctx context.Context, id string) error

	// List with filters and pagination
	FindAll(ctx context.Context, page, perPage int, search string, status *models.RecruitmentStatus, divisionID, positionID *string, priority *models.RecruitmentPriority) ([]models.RecruitmentRequest, int64, error)

	// Generate unique request code
	GenerateRequestCode(ctx context.Context) (string, error)
}
