package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"

	"github.com/gilabs/gims/api/internal/pos/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PosTableStatusRepository manages real-time table occupancy during a POS session
type PosTableStatusRepository interface {
	UpsertByTableObject(ctx context.Context, ts *models.PosTableStatusRecord) error
	UpdateStatus(ctx context.Context, sessionID, tableObjectID string, status models.PosTableStatus, orderID *string, guestCount int) error
	FindBySession(ctx context.Context, sessionID string) ([]models.PosTableStatusRecord, error)
}

type posTableStatusRepository struct {
	db *gorm.DB
}

func NewPosTableStatusRepository(db *gorm.DB) PosTableStatusRepository {
	return &posTableStatusRepository{db: db}
}

func (r *posTableStatusRepository) UpsertByTableObject(ctx context.Context, ts *models.PosTableStatusRecord) error {
	return database.GetDB(ctx, r.db).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "session_id"}, {Name: "table_object_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"status", "occupied_since", "current_order_id", "guest_count", "floor_plan_id", "table_label", "updated_at",
		}),
	}).Create(ts).Error
}

func (r *posTableStatusRepository) UpdateStatus(ctx context.Context, sessionID, tableObjectID string, status models.PosTableStatus, orderID *string, guestCount int) error {
	updates := map[string]interface{}{
		"status":      status,
		"guest_count": guestCount,
	}
	if orderID != nil {
		updates["current_order_id"] = orderID
	}
	if status == models.PosTableStatusAvailable || status == models.PosTableStatusCleaning {
		updates["occupied_since"] = nil
	}

	return database.GetDB(ctx, r.db).
		Model(&models.PosTableStatusRecord{}).
		Where("session_id = ? AND table_object_id = ?", sessionID, tableObjectID).
		Updates(updates).Error
}

func (r *posTableStatusRepository) FindBySession(ctx context.Context, sessionID string) ([]models.PosTableStatusRecord, error) {
	var records []models.PosTableStatusRecord
	err := database.GetDB(ctx, r.db).
		Where("session_id = ? AND deleted_at IS NULL", sessionID).
		Find(&records).Error
	return records, err
}
