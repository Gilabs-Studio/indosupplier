package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeEducationHistoryRepository interface {
	Create(ctx context.Context, education *models.EmployeeEducationHistory) error
	Update(ctx context.Context, education *models.EmployeeEducationHistory) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.EmployeeEducationHistory, error)
	FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*models.EmployeeEducationHistory, error)
	FindLatestByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeEducationHistory, error)
	FindOngoingByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeEducationHistory, error)
	CountByEmployee(ctx context.Context, employeeID uuid.UUID) (int64, error)
}

type employeeEducationHistoryRepository struct {
	db *gorm.DB
}

func NewEmployeeEducationHistoryRepository(db *gorm.DB) EmployeeEducationHistoryRepository {
	return &employeeEducationHistoryRepository{db: db}
}

func (r *employeeEducationHistoryRepository) Create(ctx context.Context, education *models.EmployeeEducationHistory) error {
	return r.db.WithContext(ctx).Create(education).Error
}

func (r *employeeEducationHistoryRepository) Update(ctx context.Context, education *models.EmployeeEducationHistory) error {
	return r.db.WithContext(ctx).Save(education).Error
}

func (r *employeeEducationHistoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.EmployeeEducationHistory{}, "id = ?", id).Error
}

func (r *employeeEducationHistoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.EmployeeEducationHistory, error) {
	var education models.EmployeeEducationHistory
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&education).Error
	if err != nil {
		return nil, err
	}
	return &education, nil
}

func (r *employeeEducationHistoryRepository) FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*models.EmployeeEducationHistory, error) {
	var educations []*models.EmployeeEducationHistory
	err := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		Order("start_date DESC").
		Find(&educations).Error
	if err != nil {
		return nil, err
	}
	return educations, nil
}

func (r *employeeEducationHistoryRepository) FindOngoingByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeEducationHistory, error) {
	var education models.EmployeeEducationHistory
	err := r.db.WithContext(ctx).
		Where("employee_id = ? AND end_date IS NULL", employeeID).
		Order("start_date DESC").
		First(&education).Error
	if err != nil {
		return nil, err
	}
	return &education, nil
}

func (r *employeeEducationHistoryRepository) FindLatestByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeEducationHistory, error) {
	var education models.EmployeeEducationHistory
	err := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		Order("start_date DESC").
		First(&education).Error
	if err != nil {
		return nil, err
	}
	return &education, nil
}

func (r *employeeEducationHistoryRepository) CountByEmployee(ctx context.Context, employeeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.EmployeeEducationHistory{}).
		Where("employee_id = ?", employeeID).
		Count(&count).Error
	return count, err
}
