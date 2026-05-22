package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"time"

	"github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetAssignmentRepository defines the interface for asset assignment history data access
type AssetAssignmentRepository interface {
	// CRUD operations
	Create(ctx context.Context, assignment *models.AssetAssignmentHistory) error
	GetByID(ctx context.Context, id string) (*models.AssetAssignmentHistory, error)
	GetByAssetID(ctx context.Context, assetID string) ([]models.AssetAssignmentHistory, error)
	GetCurrentAssignment(ctx context.Context, assetID string) (*models.AssetAssignmentHistory, error)

	// Update operations
	MarkAsReturned(ctx context.Context, id string, returnDate time.Time, reason string) error

	// Search and filter
	GetByEmployeeID(ctx context.Context, employeeID string, activeOnly bool) ([]models.AssetAssignmentHistory, error)
	GetByDepartmentID(ctx context.Context, departmentID string) ([]models.AssetAssignmentHistory, error)

	// Statistics
	GetAssignmentCounts(ctx context.Context, params AssignmentCountParams) (map[string]int64, error)
	GetEmployeeAssetCount(ctx context.Context, employeeID string) (int64, error)
}

// AssignmentCountParams defines parameters for counting assignments
type AssignmentCountParams struct {
	AssetID      string
	EmployeeID   string
	DepartmentID string
	LocationID   string
	ActiveOnly   bool
	StartDate    *time.Time
	EndDate      *time.Time
}

// assetAssignmentRepository implements AssetAssignmentRepository
type assetAssignmentRepository struct {
	db *gorm.DB
}

// NewAssetAssignmentRepository creates a new instance of AssetAssignmentRepository
func NewAssetAssignmentRepository(db *gorm.DB) AssetAssignmentRepository {
	return &assetAssignmentRepository{db: db}
}

// Create creates a new assignment record
func (r *assetAssignmentRepository) Create(ctx context.Context, assignment *models.AssetAssignmentHistory) error {
	if assignment.ID == uuid.Nil {
		assignment.ID = uuid.New()
	}
	assignment.AssignedAt = time.Now()

	return database.GetDB(ctx, r.db).Create(assignment).Error
}

// GetByID retrieves an assignment by ID
func (r *assetAssignmentRepository) GetByID(ctx context.Context, id string) (*models.AssetAssignmentHistory, error) {
	var assignment models.AssetAssignmentHistory
	err := database.GetDB(ctx, r.db).
		Where("id = ?", id).
		First(&assignment).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &assignment, nil
}

// GetByAssetID retrieves assignment history for an asset
func (r *assetAssignmentRepository) GetByAssetID(ctx context.Context, assetID string) ([]models.AssetAssignmentHistory, error) {
	var assignments []models.AssetAssignmentHistory
	err := database.GetDB(ctx, r.db).
		Where("asset_id = ?", assetID).
		Order("assigned_at DESC").
		Find(&assignments).Error

	return assignments, err
}

// GetCurrentAssignment retrieves the current (not returned) assignment for an asset
func (r *assetAssignmentRepository) GetCurrentAssignment(ctx context.Context, assetID string) (*models.AssetAssignmentHistory, error) {
	var assignment models.AssetAssignmentHistory
	err := database.GetDB(ctx, r.db).
		Where("asset_id = ? AND returned_at IS NULL", assetID).
		Order("assigned_at DESC").
		First(&assignment).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &assignment, nil
}

// MarkAsReturned marks an assignment as returned
func (r *assetAssignmentRepository) MarkAsReturned(ctx context.Context, id string, returnDate time.Time, reason string) error {
	updates := map[string]interface{}{
		"returned_at":   returnDate,
		"return_reason": reason,
	}

	return database.GetDB(ctx, r.db).
		Model(&models.AssetAssignmentHistory{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetByEmployeeID retrieves assignments for an employee
func (r *assetAssignmentRepository) GetByEmployeeID(ctx context.Context, employeeID string, activeOnly bool) ([]models.AssetAssignmentHistory, error) {
	var assignments []models.AssetAssignmentHistory

	query := database.GetDB(ctx, r.db).
		Where("employee_id = ?", employeeID)

	if activeOnly {
		query = query.Where("returned_at IS NULL")
	}

	err := query.Order("assigned_at DESC").Find(&assignments).Error
	return assignments, err
}

// GetByDepartmentID retrieves assignments for a department
func (r *assetAssignmentRepository) GetByDepartmentID(ctx context.Context, departmentID string) ([]models.AssetAssignmentHistory, error) {
	var assignments []models.AssetAssignmentHistory
	err := database.GetDB(ctx, r.db).
		Where("department_id = ? AND returned_at IS NULL", departmentID).
		Order("assigned_at DESC").
		Find(&assignments).Error

	return assignments, err
}

// GetAssignmentCounts returns counts of assignments based on parameters
func (r *assetAssignmentRepository) GetAssignmentCounts(ctx context.Context, params AssignmentCountParams) (map[string]int64, error) {
	counts := make(map[string]int64)

	query := database.GetDB(ctx, r.db).Model(&models.AssetAssignmentHistory{})

	if params.AssetID != "" {
		query = query.Where("asset_id = ?", params.AssetID)
	}
	if params.EmployeeID != "" {
		query = query.Where("employee_id = ?", params.EmployeeID)
	}
	if params.DepartmentID != "" {
		query = query.Where("department_id = ?", params.DepartmentID)
	}
	if params.LocationID != "" {
		query = query.Where("location_id = ?", params.LocationID)
	}
	if params.StartDate != nil {
		query = query.Where("assigned_at >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		query = query.Where("assigned_at <= ?", params.EndDate)
	}

	// Total assignments
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	counts["total"] = total

	// Active assignments
	if params.ActiveOnly || !params.ActiveOnly {
		var active int64
		activeQuery := query.Where("returned_at IS NULL")
		if err := activeQuery.Count(&active).Error; err != nil {
			return nil, err
		}
		counts["active"] = active
		counts["returned"] = total - active
	}

	return counts, nil
}

// GetEmployeeAssetCount returns the number of assets currently assigned to an employee
func (r *assetAssignmentRepository) GetEmployeeAssetCount(ctx context.Context, employeeID string) (int64, error) {
	var count int64
	err := database.GetDB(ctx, r.db).
		Model(&models.AssetAssignmentHistory{}).
		Where("employee_id = ? AND returned_at IS NULL", employeeID).
		Count(&count).Error

	return count, err
}
