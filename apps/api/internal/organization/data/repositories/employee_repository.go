package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EmployeeListParams defines the parameters for listing employees
type EmployeeListParams struct {
	Page          int
	PerPage       int
	Search        string
	DivisionID    string
	JobPositionID string
	AreaID        string
	CompanyID     string
	Status        string
	IsActive      *bool
	SortBy        string
	SortDir       string
}

// EmployeeRepository defines the interface for employee data operations
type EmployeeRepository interface {
	Create(ctx context.Context, employee *models.Employee) error
	FindByID(ctx context.Context, id string) (*models.Employee, error)
	FindByIDs(ctx context.Context, ids []string) ([]models.Employee, error)
	FindByCode(ctx context.Context, code string) (*models.Employee, error)
	FindByUserID(ctx context.Context, userID string) (*models.Employee, error)
	FindAll(ctx context.Context) ([]models.Employee, error)
	FindByRoleDataScope(ctx context.Context, dataScope string) ([]models.Employee, error)
	List(ctx context.Context, params EmployeeListParams) ([]models.Employee, int64, error)
	GetLastEmployeeCode(ctx context.Context) (string, error)
	Update(ctx context.Context, employee *models.Employee) error
	Delete(ctx context.Context, id string) error
}

type employeeRepository struct {
	db *gorm.DB
}

// NewEmployeeRepository creates a new EmployeeRepository instance
func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, employee *models.Employee) error {
	return database.GetDB(ctx, r.db).Create(employee).Error
}

func (r *employeeRepository) FindByID(ctx context.Context, id string) (*models.Employee, error) {
	var employee models.Employee
	err := database.GetDB(ctx, r.db).
		Preload("Division").
		Preload("JobPosition").
		Preload("Company").
		Preload("Village.District.City.Province").
		Preload("User").
		Preload("ReplacementFor").
		Preload("Areas.Area").
		Preload("Outlets.Outlet").
		Preload("Warehouses.Warehouse").
		First(&employee, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) FindByIDs(ctx context.Context, ids []string) ([]models.Employee, error) {
	var employees []models.Employee
	err := database.GetDB(ctx, r.db).
		Preload("Division").
		Preload("JobPosition").
		Preload("Company").
		Where("id IN ?", ids).
		Find(&employees).Error
	if err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *employeeRepository) FindByCode(ctx context.Context, code string) (*models.Employee, error) {
	var employee models.Employee
	err := database.GetDB(ctx, r.db).
		Preload("Division").
		Preload("JobPosition").
		Preload("Company").
		First(&employee, "employee_code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) FindByUserID(ctx context.Context, userID string) (*models.Employee, error) {
	var employee models.Employee
	err := database.GetDB(ctx, r.db).
		Preload("Division").
		Preload("JobPosition").
		Preload("Company").
		First(&employee, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) FindAll(ctx context.Context) ([]models.Employee, error) {
	var employees []models.Employee
	err := database.GetDB(ctx, r.db).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&employees).Error
	if err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *employeeRepository) List(ctx context.Context, params EmployeeListParams) ([]models.Employee, int64, error) {
	var employees []models.Employee
	var total int64

	// Use WithContext instead of GetDB: when AreaID filter triggers a JOIN on employee_areas,
	// both joined tables have tenant_id. GetDB adds unqualified WHERE tenant_id=? which
	// causes PG ambiguity. The GORM callback (scope.go) adds the qualified table.tenant_id filter.
	query := r.db.WithContext(ctx).Model(&models.Employee{})

	// Apply search filter
	if searchTerm := strings.TrimSpace(params.Search); searchTerm != "" {
		searchPattern := "%" + searchTerm + "%"
		query = query.Where(
			"name ILIKE ? OR employee_code ILIKE ? OR email ILIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Apply division filter
	if params.DivisionID != "" {
		query = query.Where("division_id = ?", params.DivisionID)
	}

	// Apply job position filter
	if params.JobPositionID != "" {
		query = query.Where("job_position_id = ?", params.JobPositionID)
	}

	// Apply company filter
	if params.CompanyID != "" {
		query = query.Where("company_id = ?", params.CompanyID)
	}

	// Apply status filter
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// Apply is_active filter
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	// Apply area filter (via join)
	if params.AreaID != "" {
		query = query.Joins("INNER JOIN employee_areas ON employee_areas.employee_id = employees.id").
			Where("employee_areas.area_id = ?", params.AreaID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortDir := params.SortDir
	if sortDir == "" {
		sortDir = "desc"
	}

	query = query.Order("is_active DESC").
		Order(clause.OrderByColumn{
			Column: clause.Column{Name: sortBy},
			Desc:   sortDir == "desc",
		})

	// Apply pagination
	if params.PerPage > 0 {
		offset := (params.Page - 1) * params.PerPage
		if offset < 0 {
			offset = 0
		}
		query = query.Offset(offset).Limit(params.PerPage)
	}

	// Preload relations
	err := query.
		Preload("Division").
		Preload("JobPosition").
		Preload("Company").
		Preload("Areas.Area").
		Find(&employees).Error
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

func (r *employeeRepository) GetLastEmployeeCode(ctx context.Context) (string, error) {
	var lastCode string
	err := database.GetDB(ctx, r.db).Unscoped().Model(&models.Employee{}).
		Select("employee_code").
		Where("employee_code ~ '^EMP-[0-9]+$'").
		Order("CAST(SUBSTRING(employee_code FROM 5) AS INTEGER) DESC").
		Limit(1).
		Pluck("employee_code", &lastCode).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "EMP-000", nil
		}
		return "", err
	}

	if lastCode == "" {
		return "EMP-000", nil
	}

	return lastCode, nil
}

func (r *employeeRepository) Update(ctx context.Context, employee *models.Employee) error {
	return database.GetDB(ctx, r.db).Omit(clause.Associations).Save(employee).Error
}

func (r *employeeRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Delete(&models.Employee{}, "id = ?", id).Error
}

// FindByRoleDataScope returns active employees whose linked user has a role with the given data_scope.
// Used to populate outlet manager dropdowns with only eligible users.
// Returns employees matching ANY of:
// - Role with data_scope matching the parameter (e.g., OUTLET)
// - Role name containing "manager" (case-insensitive)
// - Role data_scope is OUTLET, DIVISION, AREA (all management levels)
func (r *employeeRepository) FindByRoleDataScope(ctx context.Context, dataScope string) ([]models.Employee, error) {
	var employees []models.Employee
	// Use WithContext (not GetDB) so that the unqualified WHERE tenant_id=? from GetDB is
	// not added alongside the JOINed tables. The GORM callback adds employees.tenant_id=? (qualified).
	err := r.db.WithContext(ctx).
		Joins("INNER JOIN users ON users.id = employees.user_id AND users.deleted_at IS NULL").
		Joins("INNER JOIN roles ON roles.id = users.role_id AND roles.deleted_at IS NULL").
		Where("employees.is_active = ?", true).
		Where("users.status = ?", "active").
		Where("roles.status = ?", "active").
		Where("(roles.data_scope = ? OR roles.name ILIKE ? OR roles.data_scope IN (?, ?, ?))",
			dataScope, "%manager%", "OUTLET", "DIVISION", "AREA").
		Order("CASE WHEN roles.name ILIKE '%manager%' THEN 0 ELSE 1 END, employees.name ASC").
		Find(&employees).Error
	if err != nil {
		return nil, err
	}
	
	// Remove duplicates if an employee appears multiple times (unlikely but defensive)
	seen := make(map[string]bool)
	var unique []models.Employee
	for _, emp := range employees {
		if !seen[emp.ID] {
			seen[emp.ID] = true
			unique = append(unique, emp)
		}
	}
	
	return unique, nil
}
