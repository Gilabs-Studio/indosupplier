package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/data/repositories"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	"gorm.io/gorm"
)

var (
	ErrMappingNotConfigured = errors.New("system account mapping not configured")
	ErrMappingValidation    = errors.New("invalid system account mapping request")
	ErrAccountNotPostable   = errors.New("selected account is not postable")
	ErrAccountInactive      = errors.New("selected account is inactive")
)

type systemAccountMappingAuditLogger interface {
	Log(ctx context.Context, action string, targetID string, metadata map[string]interface{})
	LogWithReason(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{})
	LogWithChanges(ctx context.Context, action string, targetID string, metadata map[string]interface{}, changes interface{})
	LogWithChangesFull(ctx context.Context, action string, targetID string, reason string, metadata map[string]interface{}, changes interface{})
}

type SystemAccountMappingUsecase interface {
	GetAccountID(ctx context.Context, key string, companyID *string) (string, error)
	List(ctx context.Context, companyID *string) ([]dto.SystemAccountMappingResponse, error)
	GetByKey(ctx context.Context, key string, companyID *string) (*dto.SystemAccountMappingResponse, error)
	Upsert(ctx context.Context, key, coaCode, label string, companyID *string) (*dto.SystemAccountMappingResponse, error)
	Delete(ctx context.Context, key string, companyID *string) error
}

type systemAccountMappingUsecase struct {
	repo  repositories.SystemAccountMappingRepository
	coaUC ChartOfAccountUsecase
	audit systemAccountMappingAuditLogger
}

func NewSystemAccountMappingUsecase(repo repositories.SystemAccountMappingRepository, coaUC ChartOfAccountUsecase, audits ...systemAccountMappingAuditLogger) SystemAccountMappingUsecase {
	uc := &systemAccountMappingUsecase{repo: repo, coaUC: coaUC}
	if len(audits) > 0 {
		uc.audit = audits[0]
	}
	return uc
}

func (uc *systemAccountMappingUsecase) GetAccountID(ctx context.Context, key string, companyID *string) (string, error) {
	code, err := uc.repo.GetByKey(ctx, key, companyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("%w: key %s", ErrMappingNotConfigured, key)
		}
		return "", fmt.Errorf("failed to resolve mapping key %s: %w", key, err)
	}

	coa, err := uc.coaUC.GetByCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("account with code %s not found for mapping %s: %w", code, key, err)
	}

	return coa.ID, nil
}

func (uc *systemAccountMappingUsecase) List(ctx context.Context, companyID *string) ([]dto.SystemAccountMappingResponse, error) {
	rows, err := uc.repo.ListMappings(ctx, companyID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.SystemAccountMappingResponse, 0, len(rows))
	for _, row := range rows {
		mapped, mapErr := uc.toResponse(ctx, row)
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, *mapped)
	}

	return result, nil
}

func (uc *systemAccountMappingUsecase) GetByKey(ctx context.Context, key string, companyID *string) (*dto.SystemAccountMappingResponse, error) {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return nil, fmt.Errorf("%w: key is required", ErrMappingValidation)
	}

	mapping, err := uc.repo.GetMappingByKey(ctx, trimmedKey, companyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: key %s", ErrMappingNotConfigured, trimmedKey)
		}
		return nil, err
	}

	return uc.toResponse(ctx, *mapping)
}

func (uc *systemAccountMappingUsecase) Upsert(ctx context.Context, key, coaCode, label string, companyID *string) (*dto.SystemAccountMappingResponse, error) {
	trimmedKey := strings.TrimSpace(key)
	trimmedCOACode := strings.TrimSpace(coaCode)
	trimmedLabel := strings.TrimSpace(label)

	if trimmedKey == "" {
		return nil, fmt.Errorf("%w: key is required", ErrMappingValidation)
	}
	if trimmedCOACode == "" {
		return nil, fmt.Errorf("%w: coa_code is required", ErrMappingValidation)
	}
	if trimmedLabel == "" {
		trimmedLabel = trimmedKey
	}

	coa, err := uc.coaUC.GetByCode(ctx, trimmedCOACode)
	if err != nil {
		return nil, fmt.Errorf("%w: coa_code %s not found", ErrMappingValidation, trimmedCOACode)
	}
	if !coa.IsPostable {
		return nil, fmt.Errorf("%w: coa_code %s", ErrAccountNotPostable, trimmedCOACode)
	}
	if !coa.IsActive {
		return nil, fmt.Errorf("%w: coa_code %s", ErrAccountInactive, trimmedCOACode)
	}

	before, _ := uc.repo.GetExactMappingByKey(ctx, trimmedKey, companyID)

	m := &models.SystemAccountMapping{
		Key:       trimmedKey,
		COACode:   trimmedCOACode,
		Label:     trimmedLabel,
		CompanyID: companyID,
	}
	if err := uc.repo.Upsert(ctx, m); err != nil {
		return nil, err
	}

	after, err := uc.repo.GetExactMappingByKey(ctx, trimmedKey, companyID)
	if err != nil {
		return nil, err
	}

	if uc.audit != nil {
		action := "account_mappings.update"
		if before == nil {
			action = "account_mappings.create"
		}

		changes := map[string]interface{}{
			"before": before,
			"after":  after,
		}
		metadata := map[string]interface{}{
			"key":        trimmedKey,
			"company_id": after.CompanyID,
		}
		uc.audit.LogWithChanges(ctx, action, after.ID, metadata, changes)
	}

	return uc.toResponse(ctx, *after)
}

func (uc *systemAccountMappingUsecase) Delete(ctx context.Context, key string, companyID *string) error {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return fmt.Errorf("%w: key is required", ErrMappingValidation)
	}

	existing, err := uc.repo.GetExactMappingByKey(ctx, trimmedKey, companyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("%w: key %s", ErrMappingNotConfigured, trimmedKey)
		}
		return err
	}

	if err := uc.repo.DeleteByKey(ctx, trimmedKey, companyID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("%w: key %s", ErrMappingNotConfigured, trimmedKey)
		}
		return err
	}

	if uc.audit != nil {
		metadata := map[string]interface{}{
			"key":        existing.Key,
			"company_id": existing.CompanyID,
		}
		uc.audit.LogWithChanges(ctx, "account_mappings.delete", existing.ID, metadata, map[string]interface{}{
			"before": existing,
			"after":  nil,
		})
	}

	return nil
}

func (uc *systemAccountMappingUsecase) toResponse(ctx context.Context, mapping models.SystemAccountMapping) (*dto.SystemAccountMappingResponse, error) {
	coa, err := uc.coaUC.GetByCode(ctx, mapping.COACode)
	if err != nil {
		return nil, fmt.Errorf("mapped chart of account %s not found for key %s: %w", mapping.COACode, mapping.Key, err)
	}

	return &dto.SystemAccountMappingResponse{
		ID:        mapping.ID,
		Key:       mapping.Key,
		CompanyID: mapping.CompanyID,
		COACode:   mapping.COACode,
		Label:     mapping.Label,
		COA: dto.SystemAccountMappingCOAInfo{
			ID:         coa.ID,
			Code:       coa.Code,
			Name:       coa.Name,
			IsPostable: coa.IsPostable,
			IsActive:   coa.IsActive,
		},
		CreatedAt: mapping.CreatedAt,
		UpdatedAt: mapping.UpdatedAt,
	}, nil
}
