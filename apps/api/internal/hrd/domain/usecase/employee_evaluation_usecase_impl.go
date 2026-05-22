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
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/mapper"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type employeeEvaluationUsecase struct {
	db             *gorm.DB
	evaluationRepo repositories.EmployeeEvaluationRepository
	groupRepo      repositories.EvaluationGroupRepository
	criteriaRepo   repositories.EvaluationCriteriaRepository
	employeeRepo   orgRepos.EmployeeRepository
	auditService   audit.AuditService
}

const (
	errEmployeeEvaluationNotFound = "employee evaluation not found"
	evaluationDateLayout          = "2006-01-02"
)

type employeeEvaluationAuditRow struct {
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

// NewEmployeeEvaluationUsecase creates a new instance of EmployeeEvaluationUsecase
func NewEmployeeEvaluationUsecase(
	db *gorm.DB,
	evaluationRepo repositories.EmployeeEvaluationRepository,
	groupRepo repositories.EvaluationGroupRepository,
	criteriaRepo repositories.EvaluationCriteriaRepository,
	employeeRepo orgRepos.EmployeeRepository,
	auditService audit.AuditService,
) EmployeeEvaluationUsecase {
	return &employeeEvaluationUsecase{
		db:             db,
		evaluationRepo: evaluationRepo,
		groupRepo:      groupRepo,
		criteriaRepo:   criteriaRepo,
		employeeRepo:   employeeRepo,
		auditService:   auditService,
	}
}

func (u *employeeEvaluationUsecase) GetAll(ctx context.Context, page, perPage int, search, employeeID, evaluationGroupID, evaluationType string) ([]*dto.EmployeeEvaluationResponse, *response.PaginationMeta, error) {
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

	evaluations, total, err := u.evaluationRepo.FindAll(ctx, page, perPage, search, employeeID, evaluationGroupID, evaluationType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch employee evaluations: %w", err)
	}

	// Collect unique employee IDs (employees + evaluators)
	employeeIDSet := make(map[string]struct{})
	for _, eval := range evaluations {
		employeeIDSet[eval.EmployeeID] = struct{}{}
		employeeIDSet[eval.EvaluatorID] = struct{}{}
	}
	employeeIDs := make([]string, 0, len(employeeIDSet))
	for id := range employeeIDSet {
		employeeIDs = append(employeeIDs, id)
	}

	// Batch fetch employees
	employeeMap := make(map[string]dto.EmployeeSimpleResponse)
	if len(employeeIDs) > 0 {
		employees, err := u.employeeRepo.FindByIDs(ctx, employeeIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch employees: %w", err)
		}
		for _, emp := range employees {
			empID, _ := uuid.Parse(emp.ID)
			employeeMap[emp.ID] = dto.EmployeeSimpleResponse{
				ID:           empID,
				EmployeeCode: emp.EmployeeCode,
				Name:         emp.Name,
			}
		}
	}

	responses := mapper.ToEmployeeEvaluationResponseList(evaluations, employeeMap)

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

func (u *employeeEvaluationUsecase) GetByID(ctx context.Context, id string) (*dto.EmployeeEvaluationResponse, error) {
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.EmployeeEvaluation{}, id, security.HRDScopeQueryOptions()) {
		return nil, errors.New(errEmployeeEvaluationNotFound)
	}
	eval, err := u.evaluationRepo.FindByIDWithDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	if eval == nil {
		return nil, errors.New(errEmployeeEvaluationNotFound)
	}

	// Fetch employee and evaluator data
	employeeIDs := []string{eval.EmployeeID, eval.EvaluatorID}
	employees, err := u.employeeRepo.FindByIDs(ctx, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %w", err)
	}

	employeeMap := make(map[string]dto.EmployeeSimpleResponse)
	for _, emp := range employees {
		empID, _ := uuid.Parse(emp.ID)
		employeeMap[emp.ID] = dto.EmployeeSimpleResponse{
			ID:           empID,
			EmployeeCode: emp.EmployeeCode,
			Name:         emp.Name,
		}
	}

	// Build criteria name map from the evaluation group's criteria
	criteriaMap := make(map[string]string)
	if eval.EvaluationGroupID != "" {
		criteria, err := u.criteriaRepo.FindByGroupID(ctx, eval.EvaluationGroupID)
		if err == nil {
			for _, c := range criteria {
				criteriaMap[c.ID] = c.Name
			}
		}
	}

	return mapper.ToEmployeeEvaluationResponse(eval, employeeMap, criteriaMap), nil
}

func (u *employeeEvaluationUsecase) GetFormData(ctx context.Context) (*dto.EmployeeEvaluationFormDataResponse, error) {
	// Fetch all employees
	employees, err := listScopedActiveEmployees(ctx, u.db)
	if err != nil {
		return nil, err
	}

	employeeOptions := make([]dto.EmployeeFormOption, 0, len(employees))
	for _, emp := range employees {
		employeeID, err := uuid.Parse(emp.ID)
		if err != nil {
			continue
		}
		employeeOptions = append(employeeOptions, dto.EmployeeFormOption{
			ID:           employeeID,
			EmployeeCode: emp.EmployeeCode,
			Name:         emp.Name,
		})
	}

	// Fetch active evaluation groups
	isActive := true
	groups, _, err := u.groupRepo.FindAll(ctx, 1, 100, "", &isActive)
	if err != nil {
		return nil, err
	}

	groupOptions := make([]dto.EvaluationGroupSimpleResponse, 0, len(groups))
	for _, g := range groups {
		groupOptions = append(groupOptions, dto.EvaluationGroupSimpleResponse{
			ID:   g.ID,
			Name: g.Name,
		})
	}

	// Evaluation types
	evaluationTypes := []dto.EvaluationTypeOption{
		{Value: "SELF", Label: "Self Evaluation"},
		{Value: "MANAGER", Label: "Manager Evaluation"},
	}

	return &dto.EmployeeEvaluationFormDataResponse{
		Employees:        employeeOptions,
		EvaluationGroups: groupOptions,
		EvaluationTypes:  evaluationTypes,
	}, nil
}

func (u *employeeEvaluationUsecase) Create(ctx context.Context, req *dto.CreateEmployeeEvaluationRequest) (*dto.EmployeeEvaluationResponse, error) {
	// Validate employee exists
	employee, err := u.employeeRepo.FindByID(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, errors.New("employee not found")
	}

	// Validate evaluator exists
	evaluator, err := u.employeeRepo.FindByID(ctx, req.EvaluatorID)
	if err != nil {
		return nil, err
	}
	if evaluator == nil {
		return nil, errors.New("evaluator not found")
	}

	// Validate evaluation group exists and is active
	group, err := u.groupRepo.FindByID(ctx, req.EvaluationGroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("evaluation group not found")
	}
	if !group.IsActive {
		return nil, errors.New("evaluation group is not active")
	}

	// Parse dates
	periodStart, err := time.Parse(evaluationDateLayout, req.PeriodStart)
	if err != nil {
		return nil, errors.New("invalid period_start format, must be YYYY-MM-DD")
	}
	periodEnd, err := time.Parse(evaluationDateLayout, req.PeriodEnd)
	if err != nil {
		return nil, errors.New("invalid period_end format, must be YYYY-MM-DD")
	}

	// Validate period_end is after period_start
	if periodEnd.Before(periodStart) {
		return nil, errors.New("period_end must be after period_start")
	}

	id := uuid.New().String()
	evaluation := &models.EmployeeEvaluation{
		ID:                id,
		EmployeeID:        req.EmployeeID,
		EvaluationGroupID: req.EvaluationGroupID,
		EvaluatorID:       req.EvaluatorID,
		EvaluationType:    models.EvaluationType(req.EvaluationType),
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		OverallScore:      0,
		Notes:             req.Notes,
	}

	if err := u.evaluationRepo.Create(ctx, evaluation); err != nil {
		return nil, fmt.Errorf("failed to create employee evaluation: %w", err)
	}

	// Save criteria scores if provided
	if len(req.CriteriaScores) > 0 {
		// Build weight map from group criteria
		criteria, err := u.criteriaRepo.FindByGroupID(ctx, req.EvaluationGroupID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch criteria: %w", err)
		}
		criteriaWeightMap := make(map[string]float64)
		for _, c := range criteria {
			criteriaWeightMap[c.ID] = c.Weight
		}

		// Validate all criteria IDs belong to the group
		for _, score := range req.CriteriaScores {
			if _, ok := criteriaWeightMap[score.EvaluationCriteriaID]; !ok {
				return nil, fmt.Errorf("criteria %s does not belong to the evaluation group", score.EvaluationCriteriaID)
			}
		}

		scoreModels := mapper.ToEvaluationCriteriaScoreModels(id, req.CriteriaScores, criteriaWeightMap)

		if err := u.evaluationRepo.SaveCriteriaScores(ctx, id, scoreModels); err != nil {
			return nil, fmt.Errorf("failed to save criteria scores: %w", err)
		}

		// Calculate and update overall score
		evaluation.CriteriaScores = scoreModels
		evaluation.CalculateOverallScore()
		if err := u.evaluationRepo.Update(ctx, evaluation); err != nil {
			return nil, fmt.Errorf("failed to update overall score: %w", err)
		}
	}

	// Return full response
	out, err := u.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	u.auditService.Log(ctx, "employee_evaluation.create", id, map[string]interface{}{"after": out})

	return out, nil
}

func (u *employeeEvaluationUsecase) Update(ctx context.Context, id string, req *dto.UpdateEmployeeEvaluationRequest) (*dto.EmployeeEvaluationResponse, error) {
	eval, err := u.evaluationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if eval == nil {
		return nil, errors.New(errEmployeeEvaluationNotFound)
	}
	before := *eval

	// Update fields
	if req.EvaluatorID != nil {
		// Validate evaluator exists
		evaluator, err := u.employeeRepo.FindByID(ctx, *req.EvaluatorID)
		if err != nil {
			return nil, err
		}
		if evaluator == nil {
			return nil, errors.New("evaluator not found")
		}
		eval.EvaluatorID = *req.EvaluatorID
	}
	if req.EvaluationType != nil {
		eval.EvaluationType = models.EvaluationType(*req.EvaluationType)
	}
	if req.PeriodStart != nil {
		periodStart, err := time.Parse(evaluationDateLayout, *req.PeriodStart)
		if err != nil {
			return nil, errors.New("invalid period_start format, must be YYYY-MM-DD")
		}
		eval.PeriodStart = periodStart
	}
	if req.PeriodEnd != nil {
		periodEnd, err := time.Parse(evaluationDateLayout, *req.PeriodEnd)
		if err != nil {
			return nil, errors.New("invalid period_end format, must be YYYY-MM-DD")
		}
		eval.PeriodEnd = periodEnd
	}
	if req.Notes != nil {
		eval.Notes = req.Notes
	}

	// Validate period_end is after period_start
	if eval.PeriodEnd.Before(eval.PeriodStart) {
		return nil, errors.New("period_end must be after period_start")
	}

	// Update criteria scores if provided
	if len(req.CriteriaScores) > 0 {
		criteria, err := u.criteriaRepo.FindByGroupID(ctx, eval.EvaluationGroupID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch criteria: %w", err)
		}
		criteriaWeightMap := make(map[string]float64)
		for _, c := range criteria {
			criteriaWeightMap[c.ID] = c.Weight
		}

		// Validate all criteria IDs belong to the group
		for _, score := range req.CriteriaScores {
			if _, ok := criteriaWeightMap[score.EvaluationCriteriaID]; !ok {
				return nil, fmt.Errorf("criteria %s does not belong to the evaluation group", score.EvaluationCriteriaID)
			}
		}

		scoreModels := mapper.ToEvaluationCriteriaScoreModels(id, req.CriteriaScores, criteriaWeightMap)

		if err := u.evaluationRepo.SaveCriteriaScores(ctx, id, scoreModels); err != nil {
			return nil, fmt.Errorf("failed to save criteria scores: %w", err)
		}

		// Recalculate overall score
		eval.CriteriaScores = scoreModels
		eval.CalculateOverallScore()
	}

	if err := u.evaluationRepo.Update(ctx, eval); err != nil {
		return nil, fmt.Errorf("failed to update employee evaluation: %w", err)
	}

	out, err := u.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	u.auditService.Log(ctx, "employee_evaluation.update", id, map[string]interface{}{
		"before": before,
		"after":  out,
	})

	return out, nil
}

func (u *employeeEvaluationUsecase) Delete(ctx context.Context, id string) error {
	eval, err := u.evaluationRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if eval == nil {
		return errors.New(errEmployeeEvaluationNotFound)
	}

	// Delete criteria scores first
	before := *eval
	if err := u.evaluationRepo.DeleteCriteriaScores(ctx, id); err != nil {
		return fmt.Errorf("failed to delete criteria scores: %w", err)
	}

	if err := u.evaluationRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete employee evaluation: %w", err)
	}

	u.auditService.Log(ctx, "employee_evaluation.delete", id, map[string]interface{}{"before": before})

	return nil
}

func (u *employeeEvaluationUsecase) ListAuditTrail(ctx context.Context, id string, page, perPage int) ([]dto.EvaluationAuditTrailEntry, int64, error) {
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
		Where("audit_logs.permission_code LIKE ?", "employee_evaluation.%")

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]employeeEvaluationAuditRow, 0)
	if err := tx.
		Select("audit_logs.id, audit_logs.actor_id, audit_logs.permission_code, audit_logs.target_id, audit_logs.action, audit_logs.metadata, audit_logs.created_at, users.email as actor_email, users.name as actor_name").
		Joins("LEFT JOIN users ON users.id = audit_logs.actor_id").
		Order("audit_logs.created_at DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return mapEmployeeEvaluationAuditEntries(rows), total, nil
}

func mapEmployeeEvaluationAuditEntries(rows []employeeEvaluationAuditRow) []dto.EvaluationAuditTrailEntry {
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
