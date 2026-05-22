package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/organization/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EmployeeAreaRepository defines the interface for employee area assignment operations.
// IsSupervisor=true marks the employee as a supervisor of the area.
type EmployeeAreaRepository interface {
	// AssignAreas assigns the employee as a regular member of the given areas.
	AssignAreas(ctx context.Context, employeeID string, areaIDs []string) error
	// AssignAreaWithRole assigns the employee to an area with an explicit supervisor flag.
	AssignAreaWithRole(ctx context.Context, employeeID, areaID string, isSupervisor bool) error
	// RemoveAreas removes specific area assignments for an employee.
	RemoveAreas(ctx context.Context, employeeID string, areaIDs []string) error
	// RemoveAllAreas removes all area assignments for an employee.
	RemoveAllAreas(ctx context.Context, employeeID string) error
	// GetByEmployeeID returns all area assignments for an employee.
	GetByEmployeeID(ctx context.Context, employeeID string) ([]models.EmployeeArea, error)
	// GetByAreaID returns all employee assignments for an area.
	GetByAreaID(ctx context.Context, areaID string) ([]models.EmployeeArea, error)
	// GetSupervisedAreas returns only the areas where the employee is a supervisor.
	GetSupervisedAreas(ctx context.Context, employeeID string) ([]models.EmployeeArea, error)
	// AssignSupervisorAreas replaces all supervisor assignments for the employee.
	AssignSupervisorAreas(ctx context.Context, employeeID string, areaIDs []string) error
	// RemoveFromArea removes a specific employee-area assignment (regardless of role).
	RemoveFromArea(ctx context.Context, employeeID, areaID string) error
	// AssignMembersToArea assigns a list of employees as members of a given area.
	AssignMembersToArea(ctx context.Context, areaID string, employeeIDs []string) error
	// AssignSupervisorsToArea assigns a list of employees as supervisors of a given area.
	AssignSupervisorsToArea(ctx context.Context, areaID string, employeeIDs []string) error
	// GetSupervisorsForArea returns all supervisor assignments for an area.
	GetSupervisorsForArea(ctx context.Context, areaID string) ([]models.EmployeeArea, error)
	// GetMembersForArea returns all member assignments for an area.
	GetMembersForArea(ctx context.Context, areaID string) ([]models.EmployeeArea, error)
}

type employeeAreaRepository struct {
	db *gorm.DB
}

// NewEmployeeAreaRepository creates a new EmployeeAreaRepository instance.
func NewEmployeeAreaRepository(db *gorm.DB) EmployeeAreaRepository {
	return &employeeAreaRepository{db: db}
}

func (r *employeeAreaRepository) AssignAreas(ctx context.Context, employeeID string, areaIDs []string) error {
	if len(areaIDs) == 0 {
		return nil
	}

	employeeAreas := make([]models.EmployeeArea, 0, len(areaIDs))
	tenantID := middleware.TenantFromContext(ctx)
	for _, areaID := range areaIDs {
		employeeAreas = append(employeeAreas, models.EmployeeArea{
			TenantID:     tenantID,
			EmployeeID:   employeeID,
			AreaID:       areaID,
			IsSupervisor: false,
		})
	}

	return database.GetDB(ctx, r.db).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&employeeAreas).Error
}

func (r *employeeAreaRepository) AssignAreaWithRole(ctx context.Context, employeeID, areaID string, isSupervisor bool) error {
	employeeArea := models.EmployeeArea{
		EmployeeID:   employeeID,
		AreaID:       areaID,
		IsSupervisor: isSupervisor,
	}

	// Upsert: if record exists (same employee+area), update the is_supervisor flag.
	return database.GetDB(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "employee_id"}, {Name: "area_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"is_supervisor"}),
		}).
		Create(&employeeArea).Error
}

func (r *employeeAreaRepository) RemoveAreas(ctx context.Context, employeeID string, areaIDs []string) error {
	return database.GetDB(ctx, r.db).
		Where("employee_id = ? AND area_id IN ?", employeeID, areaIDs).
		Delete(&models.EmployeeArea{}).Error
}

func (r *employeeAreaRepository) RemoveAllAreas(ctx context.Context, employeeID string) error {
	return database.GetDB(ctx, r.db).
		Where("employee_id = ?", employeeID).
		Delete(&models.EmployeeArea{}).Error
}

func (r *employeeAreaRepository) GetByEmployeeID(ctx context.Context, employeeID string) ([]models.EmployeeArea, error) {
	var areas []models.EmployeeArea
	err := database.GetDB(ctx, r.db).
		Preload("Area").
		Where("employee_id = ?", employeeID).
		Find(&areas).Error
	return areas, err
}

func (r *employeeAreaRepository) GetByAreaID(ctx context.Context, areaID string) ([]models.EmployeeArea, error) {
	var areas []models.EmployeeArea
	err := database.GetDB(ctx, r.db).
		Preload("Employee").
		Preload("Employee.Division").
		Preload("Employee.JobPosition").
		Where("area_id = ?", areaID).
		Find(&areas).Error
	return areas, err
}

func (r *employeeAreaRepository) GetSupervisedAreas(ctx context.Context, employeeID string) ([]models.EmployeeArea, error) {
	var areas []models.EmployeeArea
	err := database.GetDB(ctx, r.db).
		Preload("Area").
		Where("employee_id = ? AND is_supervisor = true", employeeID).
		Find(&areas).Error
	return areas, err
}

func (r *employeeAreaRepository) AssignSupervisorAreas(ctx context.Context, employeeID string, areaIDs []string) error {
	// Remove all existing supervisor assignments for this employee, then re-assign.
	if err := database.GetDB(ctx, r.db).
		Where("employee_id = ? AND is_supervisor = true", employeeID).
		Delete(&models.EmployeeArea{}).Error; err != nil {
		return err
	}

	if len(areaIDs) == 0 {
		return nil
	}

	records := make([]models.EmployeeArea, 0, len(areaIDs))
	tenantID := middleware.TenantFromContext(ctx)
	for _, areaID := range areaIDs {
		records = append(records, models.EmployeeArea{
			TenantID:     tenantID,
			EmployeeID:   employeeID,
			AreaID:       areaID,
			IsSupervisor: true,
		})
	}

	return database.GetDB(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "employee_id"}, {Name: "area_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"is_supervisor"}),
		}).
		Create(&records).Error
}

func (r *employeeAreaRepository) RemoveFromArea(ctx context.Context, employeeID, areaID string) error {
	return database.GetDB(ctx, r.db).
		Where("employee_id = ? AND area_id = ?", employeeID, areaID).
		Delete(&models.EmployeeArea{}).Error
}

func (r *employeeAreaRepository) AssignMembersToArea(ctx context.Context, areaID string, employeeIDs []string) error {
	if len(employeeIDs) == 0 {
		return nil
	}

	records := make([]models.EmployeeArea, 0, len(employeeIDs))
	tenantID := middleware.TenantFromContext(ctx)
	for _, empID := range employeeIDs {
		records = append(records, models.EmployeeArea{
			TenantID:     tenantID,
			EmployeeID:   empID,
			AreaID:       areaID,
			IsSupervisor: false,
		})
	}

	return database.GetDB(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "employee_id"}, {Name: "area_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"is_supervisor"}),
		}).
		Create(&records).Error
}

func (r *employeeAreaRepository) AssignSupervisorsToArea(ctx context.Context, areaID string, employeeIDs []string) error {
	if len(employeeIDs) == 0 {
		return nil
	}

	records := make([]models.EmployeeArea, 0, len(employeeIDs))
	for _, empID := range employeeIDs {
		records = append(records, models.EmployeeArea{
			EmployeeID:   empID,
			AreaID:       areaID,
			IsSupervisor: true,
		})
	}

	return database.GetDB(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "employee_id"}, {Name: "area_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"is_supervisor"}),
		}).
		Create(&records).Error
}

func (r *employeeAreaRepository) GetSupervisorsForArea(ctx context.Context, areaID string) ([]models.EmployeeArea, error) {
	var assignments []models.EmployeeArea
	err := database.GetDB(ctx, r.db).
		Preload("Employee").
		Preload("Employee.Division").
		Preload("Employee.JobPosition").
		Where("area_id = ? AND is_supervisor = true", areaID).
		Find(&assignments).Error
	return assignments, err
}

func (r *employeeAreaRepository) GetMembersForArea(ctx context.Context, areaID string) ([]models.EmployeeArea, error) {
	var assignments []models.EmployeeArea
	err := database.GetDB(ctx, r.db).
		Preload("Employee").
		Preload("Employee.Division").
		Preload("Employee.JobPosition").
		Where("area_id = ? AND is_supervisor = false", areaID).
		Find(&assignments).Error
	return assignments, err
}
