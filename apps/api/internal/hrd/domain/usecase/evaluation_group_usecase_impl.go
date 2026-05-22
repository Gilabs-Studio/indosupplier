package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/mapper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type evaluationGroupUsecase struct {
	db           *gorm.DB
	groupRepo    repositories.EvaluationGroupRepository
	criteriaRepo repositories.EvaluationCriteriaRepository
	auditService audit.AuditService
}

const errEvaluationGroupNotFound = "evaluation group not found"

type evaluationAuditRow struct {
	ID             string    `gorm:"column:id"`
	ActorID        string    `gorm:"column:actor_id"`
	PermissionCode string    `gorm:"column:permission_code"`
	TargetID       string    `gorm:"column:target_id"`
	Action         string    `gorm:"column:action"`
	Metadata       string    `gorm:"column:metadata"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	ActorEmail     *string   `gorm:"column:actor_email"`
	ActorName      *string   `gorm:"column:actor_name"`
}

// NewEvaluationGroupUsecase creates a new instance of EvaluationGroupUsecase
func NewEvaluationGroupUsecase(
	db *gorm.DB,
	groupRepo repositories.EvaluationGroupRepository,
	criteriaRepo repositories.EvaluationCriteriaRepository,
	auditService audit.AuditService,
) EvaluationGroupUsecase {
	return &evaluationGroupUsecase{
		db:           db,
		groupRepo:    groupRepo,
		criteriaRepo: criteriaRepo,
		auditService: auditService,
	}
}

func (u *evaluationGroupUsecase) GetAll(ctx context.Context, page, perPage int, search string, isActive *bool) ([]*dto.EvaluationGroupResponse, *response.PaginationMeta, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	groups, total, err := u.groupRepo.FindAll(ctx, page, perPage, search, isActive)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch evaluation groups: %w", err)
	}

	responses := mapper.ToEvaluationGroupResponseList(groups)

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	meta := &response.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      int(total),
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	return responses, meta, nil
}

func (u *evaluationGroupUsecase) GetByID(ctx context.Context, id string) (*dto.EvaluationGroupResponse, error) {
	// Fetch group with criteria preloaded
	group, err := u.groupRepo.FindByIDWithCriteria(ctx, id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New(errEvaluationGroupNotFound)
	}

	return mapper.ToEvaluationGroupResponse(group), nil
}

func (u *evaluationGroupUsecase) Create(ctx context.Context, req *dto.CreateEvaluationGroupRequest) (*dto.EvaluationGroupResponse, error) {
	id := uuid.New().String()
	group := mapper.ToEvaluationGroupModel(req, id)

	// Set tenant_id from request context
	group.TenantID = middleware.TenantFromContext(ctx)

	if err := u.groupRepo.Create(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to create evaluation group: %w", err)
	}

	u.auditService.Log(ctx, "evaluation_group.create", id, map[string]interface{}{"after": group})

	return mapper.ToEvaluationGroupResponse(group), nil
}

func (u *evaluationGroupUsecase) Update(ctx context.Context, id string, req *dto.UpdateEvaluationGroupRequest) (*dto.EvaluationGroupResponse, error) {
	group, err := u.groupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New(errEvaluationGroupNotFound)
	}
	before := *group

	mapper.UpdateEvaluationGroupModel(group, req)

	if err := u.groupRepo.Update(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to update evaluation group: %w", err)
	}

	// Re-fetch with criteria for full response
	updatedGroup, err := u.groupRepo.FindByIDWithCriteria(ctx, id)
	if err != nil {
		return nil, err
	}

	u.auditService.Log(ctx, "evaluation_group.update", id, map[string]interface{}{
		"before": before,
		"after":  updatedGroup,
	})

	return mapper.ToEvaluationGroupResponse(updatedGroup), nil
}

func (u *evaluationGroupUsecase) Delete(ctx context.Context, id string) error {
	group, err := u.groupRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New(errEvaluationGroupNotFound)
	}

	if err := u.groupRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete evaluation group: %w", err)
	}

	u.auditService.Log(ctx, "evaluation_group.delete", id, map[string]interface{}{"before": group})

	return nil
}

func (u *evaluationGroupUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.EvaluationAuditTrailEntry, int64, error) {
	if u.db == nil {
		return nil, 0, fmt.Errorf("db is nil")
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	tx := u.db.WithContext(ctx).Model(&coreModels.AuditLog{}).
		Where("audit_logs.target_id = ?", id).
		Where("(audit_logs.permission_code LIKE ? OR audit_logs.permission_code LIKE ?)", "evaluation_group.%", "evaluation_criteria.%")

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]evaluationAuditRow, 0)
	if err := tx.
		Select("audit_logs.id, audit_logs.actor_id, audit_logs.permission_code, audit_logs.target_id, audit_logs.action, audit_logs.metadata, audit_logs.created_at, users.email as actor_email, users.name as actor_name").
		Joins("LEFT JOIN users ON users.id = audit_logs.actor_id").
		Order("audit_logs.created_at DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return mapEvaluationAuditEntries(rows), total, nil
}

func mapEvaluationAuditEntries(rows []evaluationAuditRow) []dto.EvaluationAuditTrailEntry {
	entries := make([]dto.EvaluationAuditTrailEntry, 0, len(rows))
	for _, row := range rows {
		metadata := map[string]interface{}{}
		if strings.TrimSpace(row.Metadata) != "" {
			_ = json.Unmarshal([]byte(row.Metadata), &metadata)
		}

		var user *dto.EvaluationAuditTrailUser
		if row.ActorID != "" {
			email := ""
			name := ""
			if row.ActorEmail != nil {
				email = *row.ActorEmail
			}
			if row.ActorName != nil {
				name = *row.ActorName
			}
			user = &dto.EvaluationAuditTrailUser{ID: row.ActorID, Email: email, Name: name}
		}

		entries = append(entries, dto.EvaluationAuditTrailEntry{
			ID:             row.ID,
			Action:         row.Action,
			PermissionCode: row.PermissionCode,
			TargetID:       row.TargetID,
			Metadata:       metadata,
			User:           user,
			CreatedAt:      row.CreatedAt,
		})
	}

	return entries
}
