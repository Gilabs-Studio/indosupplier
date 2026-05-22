package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/loyalty/data/models"
	"gorm.io/gorm"
)

type loyaltyMemberRepo struct {
	db *gorm.DB
}

func NewLoyaltyMemberRepository(db *gorm.DB) LoyaltyMemberRepository {
	return &loyaltyMemberRepo{db: db}
}

func (r *loyaltyMemberRepo) Create(ctx context.Context, member *models.LoyaltyMember) error {
	return database.GetDB(ctx, r.db).Create(member).Error
}

func (r *loyaltyMemberRepo) GetByID(ctx context.Context, id string) (*models.LoyaltyMember, error) {
	var member models.LoyaltyMember
	err := database.GetDB(ctx, r.db).Where("id = ?", id).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &member, err
}

func (r *loyaltyMemberRepo) GetByCustomerID(ctx context.Context, customerID string) (*models.LoyaltyMember, error) {
	var member models.LoyaltyMember
	err := database.GetDB(ctx, r.db).Where("customer_id = ?", customerID).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &member, err
}

// LookupByNameAndOutlet performs case-insensitive name matching by joining
// the customers table. It matches members whose program is active for the outlet.
func (r *loyaltyMemberRepo) LookupByNameAndOutlet(ctx context.Context, name, outletID string) (*models.LoyaltyMember, error) {
	var member models.LoyaltyMember
	err := database.GetDB(ctx, r.db).
		Joins("JOIN customers ON customers.id = loyalty_members.customer_id").
		Joins("JOIN loyalty_programs ON loyalty_programs.id = loyalty_members.program_id").
		Where("loyalty_members.deleted_at IS NULL").
		Where("customers.deleted_at IS NULL").
		Where("LOWER(customers.name) LIKE LOWER(?)", "%"+name+"%").
		Where("loyalty_programs.is_active = true").
		Where("loyalty_programs.outlet_id = ? OR loyalty_programs.outlet_id IS NULL", outletID).
		Order("loyalty_programs.outlet_id DESC NULLS LAST").
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &member, err
}

func (r *loyaltyMemberRepo) List(ctx context.Context, params MemberListParams) ([]models.LoyaltyMember, int64, error) {
	var members []models.LoyaltyMember
	var total int64

	// Use WithContext (not GetDB) because ProgramOutletIDs filter JOINs loyalty_programs.
	// GetDB's unqualified WHERE tenant_id=? causes PG ambiguity on JOIN queries.
	q := r.db.WithContext(ctx).Model(&models.LoyaltyMember{})

	// When searching by name, join customers table to filter by customer name or member_code.
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		q = q.Joins("JOIN customers c ON c.id = loyalty_members.customer_id AND c.deleted_at IS NULL").
			Where("c.name ILIKE ? OR loyalty_members.member_code ILIKE ?", searchPattern, searchPattern)
	}

	// Outlet-scope filter: only members whose program belongs to the given outlets (or is global).
	if len(params.ProgramOutletIDs) > 0 {
		q = q.Joins("JOIN loyalty_programs lp ON lp.id = loyalty_members.program_id AND lp.deleted_at IS NULL").
			Where("lp.outlet_id IS NULL OR lp.outlet_id IN ?", params.ProgramOutletIDs)
	}

	if params.Tier != "" {
		q = q.Where("loyalty_members.current_tier = ?", params.Tier)
	}
	if params.OutletID != "" {
		q = q.Where("loyalty_members.enrolled_outlet_id = ?", params.OutletID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PerPage
	err := q.Order("loyalty_members.lifetime_points DESC").Offset(offset).Limit(params.PerPage).Find(&members).Error
	return members, total, err
}

func (r *loyaltyMemberRepo) Update(ctx context.Context, member *models.LoyaltyMember) error {
	return database.GetDB(ctx, r.db).Save(member).Error
}

func (r *loyaltyMemberRepo) GetNextMemberCode(ctx context.Context) (string, error) {
	var count int64
	if err := database.GetDB(ctx, r.db).Model(&models.LoyaltyMember{}).Unscoped().Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("MBR-%05d", count+1), nil
}
