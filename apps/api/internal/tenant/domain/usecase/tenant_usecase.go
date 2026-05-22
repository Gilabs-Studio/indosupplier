package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/tenant/data/repositories"
	"github.com/gilabs/gims/api/internal/tenant/domain/dto"
	"gorm.io/gorm"
)

var ErrTenantDeletionNotScheduled = errors.New("tenant deletion is not scheduled")

// TenantUsecase handles tenant management (system admin only)
type TenantUsecase interface {
	ListTenants(ctx context.Context) ([]dto.TenantListResponse, error)
	ListTenantsPaginated(ctx context.Context, params dto.TenantListParams) ([]dto.TenantListResponse, *response.PaginationMeta, error)
	GetTenantDetail(ctx context.Context, id string) (*dto.TenantListResponse, error)
	RecoverTenantDeletion(ctx context.Context, id string) (*dto.TenantListResponse, error)
}

type tenantUsecase struct {
	repo repositories.TenantRepository
}

// NewTenantUsecase creates a new TenantUsecase
func NewTenantUsecase(repo repositories.TenantRepository) TenantUsecase {
	return &tenantUsecase{repo: repo}
}

func toTenantDTO(t repositories.TenantWithCounts) dto.TenantListResponse {
	ownerID := ""
	if t.OwnerUserID != nil {
		ownerID = *t.OwnerUserID
	}
	deletionRequestedBy := ""
	if t.DeletionRequestedBy != nil {
		deletionRequestedBy = *t.DeletionRequestedBy
	}
	var createdAt *string
	if !t.CreatedAt.IsZero() {
		s := t.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		createdAt = &s
	}
	var deletionRequestedAt *string
	if t.DeletionRequestedAt != nil {
		s := t.DeletionRequestedAt.Format("2006-01-02T15:04:05Z07:00")
		deletionRequestedAt = &s
	}
	var deletionScheduledAt *string
	if t.DeletionScheduledAt != nil {
		s := t.DeletionScheduledAt.Format("2006-01-02T15:04:05Z07:00")
		deletionScheduledAt = &s
	}
	deletionPreviousStatus := ""
	if t.DeletionPreviousStatus != nil {
		deletionPreviousStatus = *t.DeletionPreviousStatus
	}
	return dto.TenantListResponse{
		ID:                     t.ID,
		Name:                   t.Name,
		Slug:                   t.Slug,
		Status:                 t.Status,
		Plan:                   t.Plan,
		MaxUsers:               t.MaxUsers,
		CurrentUsers:           int(t.CurrentUsers),
		CompanyCount:           int(t.CompanyCount),
		OutletCount:            int(t.OutletCount),
		WarehouseCount:         int(t.WarehouseCount),
		OwnerUserID:            ownerID,
		OwnerName:              t.OwnerName,
		OwnerEmail:             t.OwnerEmail,
		DeletionRequestedAt:    deletionRequestedAt,
		DeletionScheduledAt:    deletionScheduledAt,
		DeletionRequestedBy:    deletionRequestedBy,
		DeletionPreviousStatus: deletionPreviousStatus,
		CreatedAt:              createdAt,
	}
}

func (u *tenantUsecase) ListTenants(ctx context.Context) ([]dto.TenantListResponse, error) {
	list, _, err := u.ListTenantsPaginated(ctx, dto.TenantListParams{Page: 1, PerPage: 100})
	return list, err
}

func (u *tenantUsecase) ListTenantsPaginated(ctx context.Context, params dto.TenantListParams) ([]dto.TenantListResponse, *response.PaginationMeta, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 10
	}

	tenants, total, err := u.repo.FindPaginatedWithCounts(ctx, params.Search, params.Page, params.PerPage)
	if err != nil {
		return nil, nil, err
	}

	result := make([]dto.TenantListResponse, 0, len(tenants))
	for _, t := range tenants {
		result = append(result, toTenantDTO(t))
	}

	pagination := response.NewPaginationMeta(params.Page, params.PerPage, int(total))
	return result, pagination, nil
}

func (u *tenantUsecase) GetTenantDetail(ctx context.Context, id string) (*dto.TenantListResponse, error) {
	tenant, err := u.repo.FindByIDWithCounts(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toTenantDTO(*tenant)
	return &resp, nil
}

func (u *tenantUsecase) RecoverTenantDeletion(ctx context.Context, id string) (*dto.TenantListResponse, error) {
	tenant, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	if !strings.EqualFold(strings.TrimSpace(tenant.Status), "pending_deletion") {
		return nil, ErrTenantDeletionNotScheduled
	}

	now := apptime.Now()
	nextStatus := "active"
	if tenant.DeletionPreviousStatus != nil {
		candidate := strings.TrimSpace(*tenant.DeletionPreviousStatus)
		if candidate != "" && !strings.EqualFold(candidate, "pending_deletion") {
			nextStatus = candidate
		}
	}

	tenant.Status = nextStatus
	tenant.DeletionRecoveredAt = &now
	tenant.DeletionRequestedAt = nil
	tenant.DeletionScheduledAt = nil
	tenant.DeletionRequestedBy = nil
	tenant.DeletionPreviousStatus = nil

	if err := u.repo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	// Re-activate tenant users after explicit system-admin recovery.
	_ = database.DB.WithContext(ctx).
		Table("users").
		Where("tenant_id = ? AND deleted_at IS NULL", tenant.ID).
		Updates(map[string]interface{}{"status": "active", "updated_at": now}).Error

	full, err := u.repo.FindByIDWithCounts(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toTenantDTO(*full)
	return &resp, nil
}
