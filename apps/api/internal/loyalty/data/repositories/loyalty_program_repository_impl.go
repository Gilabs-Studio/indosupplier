package repositories

import (
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/loyalty/data/models"
	"gorm.io/gorm"
)

type loyaltyProgramRepo struct {
	db *gorm.DB
}

func NewLoyaltyProgramRepository(db *gorm.DB) LoyaltyProgramRepository {
	return &loyaltyProgramRepo{db: db}
}

func (r *loyaltyProgramRepo) Create(ctx context.Context, program *models.LoyaltyProgram) error {
	return database.GetDB(ctx, r.db).Create(program).Error
}

func (r *loyaltyProgramRepo) GetByID(ctx context.Context, id string) (*models.LoyaltyProgram, error) {
	var program models.LoyaltyProgram
	err := database.GetDB(ctx, r.db).Where("id = ?", id).First(&program).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &program, err
}

func (r *loyaltyProgramRepo) List(ctx context.Context, page, perPage int, search string) ([]models.LoyaltyProgram, int64, error) {
	var programs []models.LoyaltyProgram
	var total int64

	offset := (page - 1) * perPage
	q := database.GetDB(ctx, r.db).Model(&models.LoyaltyProgram{})
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Select("loyalty_programs.*, (SELECT count(id) FROM loyalty_members lm WHERE lm.program_id = loyalty_programs.id AND lm.deleted_at IS NULL) as member_count").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&programs).Error
	return programs, total, err
}

// ListWithOutletFilter returns global programs and programs owned by any of the given outlet IDs.
func (r *loyaltyProgramRepo) ListWithOutletFilter(ctx context.Context, outletIDs []string, page, perPage int, search string) ([]models.LoyaltyProgram, int64, error) {
	var programs []models.LoyaltyProgram
	var total int64

	offset := (page - 1) * perPage
	q := database.GetDB(ctx, r.db).Model(&models.LoyaltyProgram{}).
		Where("outlet_id IS NULL OR outlet_id IN ?", outletIDs)
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Select("loyalty_programs.*, (SELECT count(id) FROM loyalty_members lm WHERE lm.program_id = loyalty_programs.id AND lm.deleted_at IS NULL) as member_count").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&programs).Error
	return programs, total, err
}

func (r *loyaltyProgramRepo) FindActiveForOutlet(ctx context.Context, outletID string) (*models.LoyaltyProgram, error) {
	var program models.LoyaltyProgram
	// Prefer outlet-specific program; fall back to global.
	err := database.GetDB(ctx, r.db).
		Where("is_active = true AND (outlet_id = ? OR outlet_id IS NULL)", outletID).
		Order("outlet_id DESC NULLS LAST").
		First(&program).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &program, err
}

func (r *loyaltyProgramRepo) Update(ctx context.Context, program *models.LoyaltyProgram) error {
	return database.GetDB(ctx, r.db).Save(program).Error
}

func (r *loyaltyProgramRepo) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.LoyaltyProgram{}).Error
}
