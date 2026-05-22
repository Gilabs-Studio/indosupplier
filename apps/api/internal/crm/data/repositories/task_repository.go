package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TaskListParams defines filters for listing tasks
type TaskListParams struct {
	Search     string
	SortBy     string
	SortDir    string
	Limit      int
	Offset     int
	Status     string
	Priority   string
	Type       string
	AssignedTo string
	CustomerID string
	DealID     string
	LeadID     string
	DueDateFrom string
	DueDateTo   string
	IsOverdue  *bool
}

// TaskRepository defines data access for CRM tasks
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	FindByID(ctx context.Context, id string) (*models.Task, error)
	List(ctx context.Context, params TaskListParams) ([]models.Task, int64, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
	// UpdateDealIDByLeadID associates all tasks linked to the given lead with a new deal.
	UpdateDealIDByLeadID(ctx context.Context, leadID, dealID string) error
}

type taskRepository struct {
	db *gorm.DB
}

func taskScopeQueryOptions() security.ScopeQueryOptions {
	return security.ScopeQueryOptions{
		OwnerEmployeeIDColumn: "assigned_to",
		DivisionJoinSQL:       "assigned_to IN (SELECT id FROM employees WHERE division_id = ?)",
	}
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) error {
	return database.GetDB(ctx, r.db).Create(task).Error
}

func (r *taskRepository) FindByID(ctx context.Context, id string) (*models.Task, error) {
	var task models.Task
	err := database.GetDB(ctx, r.db).
		Preload("AssignedEmployee").
		Preload("AssignerEmployee").
		Preload("Customer").
		Preload("Contact").
		Preload("Deal").
		Preload("Lead").
		Preload("Reminders").
		First(&task, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) List(ctx context.Context, params TaskListParams) ([]models.Task, int64, error) {
	query := database.GetDB(ctx, r.db).Model(&models.Task{})
	query = security.ApplyScopeFilter(query, ctx, taskScopeQueryOptions())

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Priority != "" {
		query = query.Where("priority = ?", params.Priority)
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.AssignedTo != "" {
		query = query.Where("assigned_to = ?", params.AssignedTo)
	}
	if params.CustomerID != "" {
		query = query.Where("customer_id = ?", params.CustomerID)
	}
	if params.DealID != "" {
		query = query.Where("deal_id = ?", params.DealID)
	}
	if params.LeadID != "" {
		query = query.Where("lead_id = ?", params.LeadID)
	}
	if params.DueDateFrom != "" {
		query = query.Where("due_date >= ?", params.DueDateFrom)
	}
	if params.DueDateTo != "" {
		query = query.Where("due_date <= ?", params.DueDateTo)
	}
	if params.IsOverdue != nil && *params.IsOverdue {
		query = query.Where("due_date < NOW() AND status NOT IN ('completed', 'cancelled')")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy, sortDir := "created_at", "DESC"
	allowedSorts := map[string]bool{
		"title": true, "status": true, "priority": true,
		"due_date": true, "created_at": true, "updated_at": true,
	}
	if params.SortBy != "" && allowedSorts[params.SortBy] {
		sortBy = params.SortBy
	}
	if params.SortDir != "" && (strings.EqualFold(params.SortDir, "ASC") || strings.EqualFold(params.SortDir, "DESC")) {
		sortDir = strings.ToUpper(params.SortDir)
	}

	var tasks []models.Task
	err := query.
		Preload("AssignedEmployee").
		Preload("Customer").
		Preload("Lead").
		Preload("Reminders").
		Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: sortDir == "DESC"}).
		Limit(params.Limit).Offset(params.Offset).
		Find(&tasks).Error
	return tasks, total, err
}

func (r *taskRepository) Update(ctx context.Context, task *models.Task) error {
	return database.GetDB(ctx, r.db).Save(task).Error
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.Task{}).Error
}

func (r *taskRepository) UpdateDealIDByLeadID(ctx context.Context, leadID, dealID string) error {
	return database.GetDB(ctx, r.db).Model(&models.Task{}).
		Where("lead_id = ? AND deal_id IS NULL", leadID).
		Update("deal_id", dealID).Error
}
