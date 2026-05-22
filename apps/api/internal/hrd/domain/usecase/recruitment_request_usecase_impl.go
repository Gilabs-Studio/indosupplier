package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	"github.com/gilabs/gims/api/internal/hrd/domain/mapper"
	notificationService "github.com/gilabs/gims/api/internal/notification/service"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	orgDTO "github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/google/uuid"
)

type recruitmentRequestUsecase struct {
	recruitmentRepo repositories.RecruitmentRequestRepository
	employeeRepo    orgRepos.EmployeeRepository
	divisionRepo    orgRepos.DivisionRepository
	positionRepo    orgRepos.JobPositionRepository
}

// NewRecruitmentRequestUsecase creates a new instance of RecruitmentRequestUsecase
func NewRecruitmentRequestUsecase(
	recruitmentRepo repositories.RecruitmentRequestRepository,
	employeeRepo orgRepos.EmployeeRepository,
	divisionRepo orgRepos.DivisionRepository,
	positionRepo orgRepos.JobPositionRepository,
) RecruitmentRequestUsecase {
	return &recruitmentRequestUsecase{
		recruitmentRepo: recruitmentRepo,
		employeeRepo:    employeeRepo,
		divisionRepo:    divisionRepo,
		positionRepo:    positionRepo,
	}
}

func (u *recruitmentRequestUsecase) GetAll(ctx context.Context, page, perPage int, search string, status *string, divisionID, positionID *string, priority *string) ([]*dto.RecruitmentRequestResponse, *response.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	// Convert string filters to typed filters
	var statusFilter *models.RecruitmentStatus
	if status != nil && *status != "" {
		s := models.RecruitmentStatus(*status)
		statusFilter = &s
	}

	var priorityFilter *models.RecruitmentPriority
	if priority != nil && *priority != "" {
		p := models.RecruitmentPriority(*priority)
		priorityFilter = &p
	}

	requests, total, err := u.recruitmentRepo.FindAll(ctx, page, perPage, search, statusFilter, divisionID, positionID, priorityFilter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch recruitment requests: %w", err)
	}

	responses := mapper.ToRecruitmentRequestResponseList(requests)

	// Batch-fetch related entities for enrichment
	employeeIDs := make([]string, 0)
	divisionIDs := make([]string, 0)
	positionIDs := make([]string, 0)
	for _, req := range requests {
		employeeIDs = append(employeeIDs, req.RequestedByID)
		if req.ApprovedByID != nil {
			employeeIDs = append(employeeIDs, *req.ApprovedByID)
		}
		divisionIDs = append(divisionIDs, req.DivisionID)
		positionIDs = append(positionIDs, req.PositionID)
	}

	employeeMap := u.buildEmployeeMap(ctx, employeeIDs)
	divisionMap := u.buildDivisionMap(ctx, divisionIDs)
	positionMap := u.buildPositionMap(ctx, positionIDs)

	for _, resp := range responses {
		mapper.EnrichRecruitmentResponse(resp, employeeMap, divisionMap, positionMap)
	}

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

func (u *recruitmentRequestUsecase) GetByID(ctx context.Context, id string) (*dto.RecruitmentRequestResponse, error) {
	if !security.CheckRecordScopeAccess(database.DB, ctx, &models.RecruitmentRequest{}, id, security.HRDScopeQueryOptions()) {
		return nil, errors.New("recruitment request not found")
	}
	req, err := u.recruitmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("recruitment request not found")
	}

	resp := mapper.ToRecruitmentRequestResponse(req)

	// Enrich with related data
	employeeIDs := []string{req.RequestedByID}
	if req.ApprovedByID != nil {
		employeeIDs = append(employeeIDs, *req.ApprovedByID)
	}

	employeeMap := u.buildEmployeeMap(ctx, employeeIDs)
	divisionMap := u.buildDivisionMap(ctx, []string{req.DivisionID})
	positionMap := u.buildPositionMap(ctx, []string{req.PositionID})

	mapper.EnrichRecruitmentResponse(resp, employeeMap, divisionMap, positionMap)

	return resp, nil
}

func (u *recruitmentRequestUsecase) Create(ctx context.Context, reqDTO *dto.CreateRecruitmentRequestDTO, userID string) (*dto.RecruitmentRequestResponse, error) {
	// Find the employee by user ID to get the requester
	employee, err := u.employeeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("employee not found for current user")
	}

	// Validate division exists
	division, err := u.divisionRepo.FindByID(ctx, reqDTO.DivisionID)
	if err != nil || division == nil {
		return nil, errors.New("division not found")
	}

	// Validate position exists
	position, err := u.positionRepo.FindByID(ctx, reqDTO.PositionID)
	if err != nil || position == nil {
		return nil, errors.New("position not found")
	}

	// Validate salary range
	if reqDTO.SalaryRangeMin != nil && reqDTO.SalaryRangeMax != nil {
		if *reqDTO.SalaryRangeMin > *reqDTO.SalaryRangeMax {
			return nil, errors.New("INVALID_SALARY_RANGE: salary_range_min must not exceed salary_range_max")
		}
	}

	// Generate request code
	requestCode, err := u.recruitmentRepo.GenerateRequestCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate request code: %w", err)
	}

	id := uuid.New().String()
	model, err := mapper.ToRecruitmentRequestModel(reqDTO, id, requestCode, employee.ID)
	if err != nil {
		return nil, fmt.Errorf("INVALID_DATE_FORMAT: %w", err)
	}

	if err := u.recruitmentRepo.Create(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to create recruitment request: %w", err)
	}

	return u.GetByID(ctx, id)
}

func (u *recruitmentRequestUsecase) Update(ctx context.Context, id string, reqDTO *dto.UpdateRecruitmentRequestDTO, userID string) (*dto.RecruitmentRequestResponse, error) {
	existing, err := u.recruitmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("recruitment request not found")
	}

	// WHY: Only DRAFT and REJECTED requests can be edited; REJECTED allows fixing before resubmission
	if !existing.IsEditable() {
		return nil, errors.New("RECRUITMENT_NOT_EDITABLE: only DRAFT and REJECTED requests can be edited")
	}

	// Validate division if changed
	if reqDTO.DivisionID != nil {
		division, err := u.divisionRepo.FindByID(ctx, *reqDTO.DivisionID)
		if err != nil || division == nil {
			return nil, errors.New("division not found")
		}
	}

	// Validate position if changed
	if reqDTO.PositionID != nil {
		position, err := u.positionRepo.FindByID(ctx, *reqDTO.PositionID)
		if err != nil || position == nil {
			return nil, errors.New("position not found")
		}
	}

	if err := mapper.ApplyRecruitmentUpdateDTO(existing, reqDTO); err != nil {
		return nil, fmt.Errorf("INVALID_DATE_FORMAT: %w", err)
	}

	existing.UpdatedBy = &userID

	if err := u.recruitmentRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update recruitment request: %w", err)
	}

	return u.GetByID(ctx, id)
}

func (u *recruitmentRequestUsecase) Delete(ctx context.Context, id string) error {
	existing, err := u.recruitmentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("recruitment request not found")
	}

	// WHY: Only DRAFT requests can be deleted to preserve audit trail of submitted/approved requests
	if !existing.IsEditable() {
		return errors.New("RECRUITMENT_NOT_EDITABLE: only DRAFT requests can be deleted")
	}

	return u.recruitmentRepo.Delete(ctx, id)
}

func (u *recruitmentRequestUsecase) UpdateStatus(ctx context.Context, id string, reqDTO *dto.UpdateRecruitmentStatusDTO, userID string) (*dto.RecruitmentRequestResponse, error) {
	existing, err := u.recruitmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("recruitment request not found")
	}

	newStatus := models.RecruitmentStatus(reqDTO.Status)

	if !existing.IsValidTransition(newStatus) {
		return nil, fmt.Errorf("INVALID_STATUS_TRANSITION: cannot transition from %s to %s", existing.Status, newStatus)
	}

	now := apptime.Now()

	switch newStatus {
	case models.RecruitmentStatusPending:
		// Submit for approval — no special fields needed
	case models.RecruitmentStatusApproved:
		existing.ApprovedByID = &userID
		existing.ApprovedAt = &now
	case models.RecruitmentStatusRejected:
		existing.RejectedByID = &userID
		existing.RejectedAt = &now
		existing.RejectionNotes = reqDTO.Notes
	case models.RecruitmentStatusOpen:
		// Opening for hiring
	case models.RecruitmentStatusClosed:
		existing.ClosedAt = &now
	case models.RecruitmentStatusCancelled:
		// Cancellation
	}

	existing.Status = newStatus
	existing.UpdatedBy = &userID

	if err := u.recruitmentRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update recruitment status: %w", err)
	}

	if newStatus == models.RecruitmentStatusPending {
		if err := notificationService.CreateApprovalNotification(ctx, database.DB, notificationService.ApprovalNotificationParams{
			PermissionCode: "recruitment.approve",
			EntityType:     "recruitment",
			EntityID:       existing.ID,
			Title:          "Recruitment Request Approval",
			Message:        "A recruitment request has been submitted and requires your approval.",
			ActorUserID:    userID,
		}); err != nil {
			log.Printf("warning: failed to create recruitment notification: %v", err)
		}
	}

	return u.GetByID(ctx, id)
}

func (u *recruitmentRequestUsecase) UpdateFilledCount(ctx context.Context, id string, reqDTO *dto.UpdateFilledCountDTO) (*dto.RecruitmentRequestResponse, error) {
	existing, err := u.recruitmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("recruitment request not found")
	}

	if existing.Status != models.RecruitmentStatusOpen {
		return nil, errors.New("RECRUITMENT_NOT_OPEN: can only update filled count for OPEN requests")
	}

	if reqDTO.FilledCount > existing.RequiredCount {
		return nil, errors.New("FILLED_EXCEEDS_REQUIRED: filled count cannot exceed required count")
	}

	existing.FilledCount = reqDTO.FilledCount

	if err := u.recruitmentRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update filled count: %w", err)
	}

	return u.GetByID(ctx, id)
}

func (u *recruitmentRequestUsecase) GetFormData(ctx context.Context) (*dto.RecruitmentFormDataResponse, error) {
	// Fetch employees
	employees, err := u.employeeRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %w", err)
	}
	employeeOptions := make([]dto.EmployeeFormOption, 0, len(employees))
	for _, emp := range employees {
		parsedID, err := uuid.Parse(emp.ID)
		if err != nil {
			continue
		}
		employeeOptions = append(employeeOptions, dto.EmployeeFormOption{
			ID:           parsedID,
			EmployeeCode: emp.EmployeeCode,
			Name:         emp.Name,
		})
	}

	// Fetch divisions
	divListReq := &orgDTO.ListDivisionsRequest{Page: 1, PerPage: 100}
	divisions, _, err := u.divisionRepo.List(ctx, divListReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch divisions: %w", err)
	}
	divisionOptions := make([]dto.DivisionFormOption, 0, len(divisions))
	for _, div := range divisions {
		if div.IsActive {
			divisionOptions = append(divisionOptions, dto.DivisionFormOption{
				ID:   div.ID,
				Name: div.Name,
			})
		}
	}

	// Fetch job positions
	posListReq := &orgDTO.ListJobPositionsRequest{Page: 1, PerPage: 100}
	positions, _, err := u.positionRepo.List(ctx, posListReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job positions: %w", err)
	}
	positionOptions := make([]dto.JobPositionFormOption, 0, len(positions))
	for _, pos := range positions {
		if pos.IsActive {
			positionOptions = append(positionOptions, dto.JobPositionFormOption{
				ID:   pos.ID,
				Name: pos.Name,
			})
		}
	}

	// Employment types
	employmentTypes := []dto.RecruitmentEmploymentTypeOption{
		{Value: "FULL_TIME", Label: "Full Time"},
		{Value: "PART_TIME", Label: "Part Time"},
		{Value: "CONTRACT", Label: "Contract"},
		{Value: "INTERN", Label: "Intern"},
	}

	// Priorities
	priorities := []dto.RecruitmentPriorityOption{
		{Value: "LOW", Label: "Low"},
		{Value: "MEDIUM", Label: "Medium"},
		{Value: "HIGH", Label: "High"},
		{Value: "URGENT", Label: "Urgent"},
	}

	// Statuses
	statuses := []dto.RecruitmentStatusOption{
		{Value: "DRAFT", Label: "Draft"},
		{Value: "PENDING", Label: "Pending Approval"},
		{Value: "APPROVED", Label: "Approved"},
		{Value: "REJECTED", Label: "Rejected"},
		{Value: "OPEN", Label: "Open"},
		{Value: "CLOSED", Label: "Closed"},
		{Value: "CANCELLED", Label: "Cancelled"},
	}

	return &dto.RecruitmentFormDataResponse{
		Employees:       employeeOptions,
		Divisions:       divisionOptions,
		JobPositions:    positionOptions,
		EmploymentTypes: employmentTypes,
		Priorities:      priorities,
		Statuses:        statuses,
	}, nil
}

// buildEmployeeMap fetches employees by IDs and returns a map keyed by ID
func (u *recruitmentRequestUsecase) buildEmployeeMap(ctx context.Context, ids []string) map[string]*orgModels.Employee {
	m := make(map[string]*orgModels.Employee)
	if len(ids) == 0 {
		return m
	}

	// Deduplicate
	unique := make(map[string]bool)
	dedupIDs := make([]string, 0)
	for _, id := range ids {
		if !unique[id] {
			unique[id] = true
			dedupIDs = append(dedupIDs, id)
		}
	}

	employees, err := u.employeeRepo.FindByIDs(ctx, dedupIDs)
	if err != nil {
		return m
	}
	for i := range employees {
		m[employees[i].ID] = &employees[i]
	}
	return m
}

// buildDivisionMap fetches divisions by IDs and returns a map
func (u *recruitmentRequestUsecase) buildDivisionMap(ctx context.Context, ids []string) map[string]*orgModels.Division {
	m := make(map[string]*orgModels.Division)
	if len(ids) == 0 {
		return m
	}

	unique := make(map[string]bool)
	for _, id := range ids {
		if unique[id] {
			continue
		}
		unique[id] = true
		div, err := u.divisionRepo.FindByID(ctx, id)
		if err == nil && div != nil {
			m[div.ID] = div
		}
	}
	return m
}

// buildPositionMap fetches job positions by IDs and returns a map
func (u *recruitmentRequestUsecase) buildPositionMap(ctx context.Context, ids []string) map[string]*orgModels.JobPosition {
	m := make(map[string]*orgModels.JobPosition)
	if len(ids) == 0 {
		return m
	}

	unique := make(map[string]bool)
	for _, id := range ids {
		if unique[id] {
			continue
		}
		unique[id] = true
		pos, err := u.positionRepo.FindByID(ctx, id)
		if err == nil && pos != nil {
			m[pos.ID] = pos
		}
	}
	return m
}
