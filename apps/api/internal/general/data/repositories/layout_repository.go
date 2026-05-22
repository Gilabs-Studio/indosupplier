package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"errors"

	generalModels "github.com/gilabs/gims/api/internal/general/data/models"
	"gorm.io/gorm"
)

// DashboardLayoutRepository handles persistence for user dashboard layout preferences.
type DashboardLayoutRepository interface {
	GetLayout(ctx context.Context, userID, dashboardType string) (string, error)
	SaveLayout(ctx context.Context, userID, dashboardType, layoutJSON string) error
}

type dashboardLayoutRepository struct {
	db *gorm.DB
}

// NewDashboardLayoutRepository creates a new DashboardLayoutRepository.
func NewDashboardLayoutRepository(db *gorm.DB) DashboardLayoutRepository {
	return &dashboardLayoutRepository{db: db}
}

// GetLayout retrieves the saved layout JSON for a user's dashboard. Returns gorm.ErrRecordNotFound if none exists.
func (r *dashboardLayoutRepository) GetLayout(ctx context.Context, userID, dashboardType string) (string, error) {
	var layout generalModels.DashboardLayout
	err := database.GetDB(ctx, r.db).
		Where("user_id = ? AND dashboard_type = ? AND deleted_at IS NULL", userID, dashboardType).
		First(&layout).Error
	if err != nil {
		return "", err
	}
	return layout.LayoutJSON, nil
}

// SaveLayout upserts the layout JSON for a user's dashboard using ON CONFLICT DO UPDATE.
func (r *dashboardLayoutRepository) SaveLayout(ctx context.Context, userID, dashboardType, layoutJSON string) error {
	var layout generalModels.DashboardLayout
	err := database.GetDB(ctx, r.db).
		Where("user_id = ? AND dashboard_type = ? AND deleted_at IS NULL", userID, dashboardType).
		First(&layout).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new layout record
		layout = generalModels.DashboardLayout{
			UserID:        userID,
			DashboardType: dashboardType,
			LayoutJSON:    layoutJSON,
		}
		return database.GetDB(ctx, r.db).Create(&layout).Error
	} else if err != nil {
		return err
	}

	// Update existing record
	return database.GetDB(ctx, r.db).Model(&layout).
		Update("layout_json", layoutJSON).Error
}
