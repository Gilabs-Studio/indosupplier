package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/travel_planner/data/models"
	"gorm.io/gorm"
)

type TravelPlanListParams struct {
	Search    string
	PlanType  *models.TravelPlanType
	Mode      *models.TravelMode
	Status    *models.TravelPlanStatus
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

type TravelPlanRepository interface {
	GenerateCode(ctx context.Context, now time.Time) (string, error)
	Create(ctx context.Context, plan *models.TravelPlan) error
	Update(ctx context.Context, plan *models.TravelPlan) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string, withRelations bool) (*models.TravelPlan, error)
	List(ctx context.Context, params TravelPlanListParams) ([]models.TravelPlan, int64, error)
	ReplaceDays(ctx context.Context, planID string, days []models.TravelPlanDay) error
	ListExpenses(ctx context.Context, planID string) ([]models.TravelPlanExpense, error)
	CreateExpense(ctx context.Context, expense *models.TravelPlanExpense) error
	DeleteExpense(ctx context.Context, planID string, expenseID string) error
}

type travelPlanRepository struct {
	db *gorm.DB
}

func NewTravelPlanRepository(db *gorm.DB) TravelPlanRepository {
	return &travelPlanRepository{db: db}
}

func (r *travelPlanRepository) GenerateCode(ctx context.Context, now time.Time) (string, error) {
	prefix := fmt.Sprintf("TPL-%s", now.Format("200601"))

	var count int64
	err := database.GetDB(ctx, r.db).
		Model(&models.TravelPlan{}).
		Where("code LIKE ?", prefix+"-%").
		Count(&count).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%04d", prefix, count+1), nil
}

func (r *travelPlanRepository) Create(ctx context.Context, plan *models.TravelPlan) error {
	return database.GetDB(ctx, r.db).Create(plan).Error
}

func (r *travelPlanRepository) Update(ctx context.Context, plan *models.TravelPlan) error {
	return database.GetDB(ctx, r.db).
		Model(&models.TravelPlan{}).
		Where("id = ?", plan.ID).
		Updates(map[string]any{
			"title":         plan.Title,
			"plan_type":     plan.PlanType,
			"mode":          plan.Mode,
			"start_date":    plan.StartDate,
			"end_date":      plan.EndDate,
			"status":        plan.Status,
			"budget_amount": plan.BudgetAmount,
			"notes":         plan.Notes,
		}).Error
}

func (r *travelPlanRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.TravelPlan{}, "id = ?", id).Error
}

func (r *travelPlanRepository) FindByID(ctx context.Context, id string, withRelations bool) (*models.TravelPlan, error) {
	q := database.GetDB(ctx, r.db).Model(&models.TravelPlan{})

	if withRelations {
		q = q.
			Preload("Days", func(db *gorm.DB) *gorm.DB {
				return db.Order("day_index ASC")
			}).
			Preload("Days.Stops", func(db *gorm.DB) *gorm.DB {
				return db.Order("order_index ASC")
			}).
			Preload("Days.Notes", func(db *gorm.DB) *gorm.DB {
				return db.Order("order_index ASC")
			}).
			Preload("Expenses", func(db *gorm.DB) *gorm.DB {
				return db.Order("expense_date ASC").Order("created_at ASC")
			})
	}

	var plan models.TravelPlan
	if err := q.Where("id = ?", id).First(&plan).Error; err != nil {
		return nil, err
	}

	return &plan, nil
}

func (r *travelPlanRepository) List(ctx context.Context, params TravelPlanListParams) ([]models.TravelPlan, int64, error) {
	q := database.GetDB(ctx, r.db).Model(&models.TravelPlan{})

	if scope, _ := ctx.Value("permission_scope").(string); scope != "" {
		actorID, _ := ctx.Value("user_id").(string)
		scope = strings.ToUpper(strings.TrimSpace(scope))
		switch scope {
		case "OWN", "DIVISION", "AREA":
			if actorID != "" {
				q = q.Where("created_by = ?", actorID)
			}
		}
	}

	if strings.TrimSpace(params.Search) != "" {
		like := "%" + strings.TrimSpace(params.Search) + "%"
		q = q.Where("code ILIKE ? OR title ILIKE ? OR notes ILIKE ?", like, like, like)
	}
	if params.PlanType != nil {
		q = q.Where("plan_type = ?", *params.PlanType)
	}
	if params.Mode != nil {
		q = q.Where("mode = ?", *params.Mode)
	}
	if params.Status != nil {
		q = q.Where("status = ?", *params.Status)
	}
	if params.StartDate != nil {
		q = q.Where("start_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		q = q.Where("end_date <= ?", *params.EndDate)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var plans []models.TravelPlan
	if err := q.
		Order("created_at DESC").
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&plans).Error; err != nil {
		return nil, 0, err
	}

	return plans, total, nil
}

func (r *travelPlanRepository) ReplaceDays(ctx context.Context, planID string, days []models.TravelPlanDay) error {
	return database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM travel_plan_days WHERE travel_plan_id = ?", planID).Error; err != nil {
			return err
		}

		if len(days) == 0 {
			return nil
		}

		for i := range days {
			days[i].TravelPlanID = planID
			if err := tx.Create(&days[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *travelPlanRepository) ListExpenses(ctx context.Context, planID string) ([]models.TravelPlanExpense, error) {
	expenses := make([]models.TravelPlanExpense, 0)
	// travel_plan_expenses has no tenant_id column; use plain db to avoid
	// GetDB injecting WHERE tenant_id = ? which causes SQLSTATE 42703.
	// Access is already tenant-scoped: the caller verifies the parent plan
	// belongs to the tenant before calling this method.
	err := r.db.WithContext(ctx).
		Where("travel_plan_id = ?", planID).
		Order("expense_date ASC").
		Order("created_at ASC").
		Find(&expenses).Error
	if err != nil {
		return nil, err
	}
	return expenses, nil
}

func (r *travelPlanRepository) CreateExpense(ctx context.Context, expense *models.TravelPlanExpense) error {
	// travel_plan_expenses has no tenant_id; use plain db to avoid SQLSTATE 42703.
	return r.db.WithContext(ctx).Create(expense).Error
}

func (r *travelPlanRepository) DeleteExpense(ctx context.Context, planID string, expenseID string) error {
	// travel_plan_expenses has no tenant_id; use plain db to avoid SQLSTATE 42703.
	return r.db.WithContext(ctx).
		Where("id = ? AND travel_plan_id = ?", expenseID, planID).
		Delete(&models.TravelPlanExpense{}).Error
}
