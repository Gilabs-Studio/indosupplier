package repositories

import (
	"context"

	"github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmployeeContractRepository interface {
	Create(ctx context.Context, contract *models.EmployeeContract) error
	Update(ctx context.Context, contract *models.EmployeeContract) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.EmployeeContract, error)
	FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*models.EmployeeContract, error)
	FindActiveByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeContract, error)
	FindActiveByEmployeeIDs(ctx context.Context, employeeIDs []uuid.UUID) (map[uuid.UUID]*models.EmployeeContract, error)
	FindByContractNumber(ctx context.Context, contractNumber string) (*models.EmployeeContract, error)
	FindAll(ctx context.Context, employeeID *uuid.UUID, status *models.ContractStatus, contractType *models.ContractType, page, perPage int) ([]*models.EmployeeContract, int64, error)
	CountByEmployee(ctx context.Context, employeeID uuid.UUID) (int64, error)
	HasActiveContract(ctx context.Context, employeeID uuid.UUID) (bool, error)
}

type employeeContractRepository struct {
	db *gorm.DB
}

func NewEmployeeContractRepository(db *gorm.DB) EmployeeContractRepository {
	return &employeeContractRepository{db: db}
}

func (r *employeeContractRepository) Create(ctx context.Context, contract *models.EmployeeContract) error {
	return r.db.WithContext(ctx).Create(contract).Error
}

func (r *employeeContractRepository) Update(ctx context.Context, contract *models.EmployeeContract) error {
	return r.db.WithContext(ctx).Save(contract).Error
}

func (r *employeeContractRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.EmployeeContract{}, "id = ?", id).Error
}

func (r *employeeContractRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.EmployeeContract, error) {
	var contract models.EmployeeContract
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (r *employeeContractRepository) FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*models.EmployeeContract, error) {
	var contracts []*models.EmployeeContract
	err := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		Order("start_date DESC").
		Find(&contracts).Error
	return contracts, err
}

func (r *employeeContractRepository) FindActiveByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeContract, error) {
	var contract models.EmployeeContract
	err := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		Where("status = ?", models.ContractStatusActive).
		Order("start_date DESC").
		Limit(1).
		First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (r *employeeContractRepository) FindActiveByEmployeeIDs(ctx context.Context, employeeIDs []uuid.UUID) (map[uuid.UUID]*models.EmployeeContract, error) {
	result := make(map[uuid.UUID]*models.EmployeeContract)
	if len(employeeIDs) == 0 {
		return result, nil
	}

	var contracts []*models.EmployeeContract
	err := r.db.WithContext(ctx).
		Where("employee_id IN ?", employeeIDs).
		Where("status = ?", models.ContractStatusActive).
		Order("start_date DESC").
		Find(&contracts).Error
	if err != nil {
		return nil, err
	}

	// Because we ordered by start_date DESC, the first contract we see for an employee is the most recent active one
	for _, contract := range contracts {
		if _, exists := result[contract.EmployeeID]; !exists {
			result[contract.EmployeeID] = contract
		}
	}

	return result, nil
}

func (r *employeeContractRepository) FindByContractNumber(ctx context.Context, contractNumber string) (*models.EmployeeContract, error) {
	var contract models.EmployeeContract
	err := r.db.WithContext(ctx).
		Where("contract_number = ?", contractNumber).
		First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (r *employeeContractRepository) FindAll(ctx context.Context, employeeID *uuid.UUID, status *models.ContractStatus, contractType *models.ContractType, page, perPage int) ([]*models.EmployeeContract, int64, error) {
	var contracts []*models.EmployeeContract
	var total int64

	query := r.db.WithContext(ctx).Model(&models.EmployeeContract{})

	if employeeID != nil {
		query = query.Where("employee_id = ?", *employeeID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if contractType != nil {
		query = query.Where("contract_type = ?", *contractType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	err := query.
		Order("start_date DESC, created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&contracts).Error

	return contracts, total, err
}

func (r *employeeContractRepository) CountByEmployee(ctx context.Context, employeeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.EmployeeContract{}).
		Where("employee_id = ?", employeeID).
		Count(&count).Error
	return count, err
}

func (r *employeeContractRepository) HasActiveContract(ctx context.Context, employeeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.EmployeeContract{}).
		Where("employee_id = ?", employeeID).
		Where("status = ?", models.ContractStatusActive).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
