package repositories

import (
	"context"
	"errors"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"gorm.io/gorm"
)

// EmployeeSignatureRepository defines the interface for employee signature data access
type EmployeeSignatureRepository interface {
	// GetByEmployeeID retrieves the active signature for an employee
	GetByEmployeeID(ctx context.Context, employeeID string) (*models.EmployeeSignature, error)

	// GetAnyByEmployeeID retrieves any signature (including soft deleted) for an employee
	GetAnyByEmployeeID(ctx context.Context, employeeID string) (*models.EmployeeSignature, error)

	// Create creates a new signature record
	Create(ctx context.Context, signature *models.EmployeeSignature) error

	// SoftDelete performs a soft delete of a signature (for history keeping)
	SoftDelete(ctx context.Context, id string) error

	// HardDelete permanently deletes a signature (used when replacing)
	HardDelete(ctx context.Context, id string) error
}

// employeeSignatureRepository implements EmployeeSignatureRepository
type employeeSignatureRepository struct {
	db *gorm.DB
}

// NewEmployeeSignatureRepository creates a new instance of EmployeeSignatureRepository
func NewEmployeeSignatureRepository(db *gorm.DB) EmployeeSignatureRepository {
	return &employeeSignatureRepository{db: db}
}

// GetByEmployeeID retrieves the active signature for an employee
func (r *employeeSignatureRepository) GetByEmployeeID(ctx context.Context, employeeID string) (*models.EmployeeSignature, error) {
	var signature models.EmployeeSignature
	err := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		First(&signature).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &signature, nil
}

// GetAnyByEmployeeID retrieves any signature (including soft deleted) for an employee
func (r *employeeSignatureRepository) GetAnyByEmployeeID(ctx context.Context, employeeID string) (*models.EmployeeSignature, error) {
	var signature models.EmployeeSignature
	err := r.db.WithContext(ctx).
		Unscoped().
		Where("employee_id = ?", employeeID).
		First(&signature).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &signature, nil
}

// Create creates a new signature record
func (r *employeeSignatureRepository) Create(ctx context.Context, signature *models.EmployeeSignature) error {
	return r.db.WithContext(ctx).Create(signature).Error
}

// SoftDelete performs a soft delete of a signature
func (r *employeeSignatureRepository) SoftDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.EmployeeSignature{}).Error
}

// HardDelete permanently deletes a signature (used when replacing)
func (r *employeeSignatureRepository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&models.EmployeeSignature{}).Error
}
