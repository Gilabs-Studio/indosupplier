package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/hrd/data/models"
	"github.com/gilabs/gims/api/internal/hrd/data/repositories"
	"github.com/gilabs/gims/api/internal/hrd/domain/dto"
	orgDTO "github.com/gilabs/gims/api/internal/organization/domain/dto"
	orgUsecase "github.com/gilabs/gims/api/internal/organization/domain/usecase"
	"gorm.io/datatypes"
)

// validateLinkedInURL validates that the URL is a valid LinkedIn profile URL and extracts the username
// Returns the extracted username or empty string if invalid
func validateLinkedInURL(url string) (string, error) {
	if url == "" {
		return "", nil // Empty is valid (optional field)
	}

	// Normalize URL
	url = strings.TrimSpace(url)

	// Remove protocol if present for easier parsing
	urlWithoutProtocol := url
	if trimmedURL, ok := strings.CutPrefix(url, "http://"); ok {
		urlWithoutProtocol = trimmedURL
	} else if trimmedURL, ok := strings.CutPrefix(url, "https://"); ok {
		urlWithoutProtocol = trimmedURL
	}

	// Remove www. if present
	urlWithoutProtocol = strings.TrimPrefix(urlWithoutProtocol, "www.")

	// LinkedIn URL patterns to match
	// linkedin.com/in/username
	// linkedin.com/pub/username
	// linkedin.com/profile/view?id=...

	// Pattern for /in/username
	inPattern := regexp.MustCompile(`^linkedin\.com/in/([a-zA-Z0-9\-]+)/?$`)
	if matches := inPattern.FindStringSubmatch(urlWithoutProtocol); len(matches) > 1 {
		return matches[1], nil
	}

	// Pattern for /pub/username
	pubPattern := regexp.MustCompile(`^linkedin\.com/pub/([a-zA-Z0-9\-]+)/?$`)
	if matches := pubPattern.FindStringSubmatch(urlWithoutProtocol); len(matches) > 1 {
		return matches[1], nil
	}

	// Pattern for /profile/view?id=...
	profilePattern := regexp.MustCompile(`^linkedin\.com/profile/view\?id=[0-9]+/?$`)
	if profilePattern.MatchString(urlWithoutProtocol) {
		return "", errors.New("INVALID_LINKEDIN_URL: Profile ID format not supported. Please use the /in/username format")
	}

	return "", errors.New("INVALID_LINKEDIN_URL: Please enter a valid LinkedIn profile URL (e.g., linkedin.com/in/username)")
}

// strPtrOrNil returns a pointer to the string if not empty, otherwise nil
func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type recruitmentApplicantUsecase struct {
	applicantRepo   repositories.RecruitmentApplicantRepository
	stageRepo       repositories.ApplicantStageRepository
	activityRepo    repositories.ApplicantActivityRepository
	recruitmentRepo repositories.RecruitmentRequestRepository
	employeeUsecase orgUsecase.EmployeeUsecase
}

// NewRecruitmentApplicantUsecase creates a new instance of RecruitmentApplicantUsecase
func NewRecruitmentApplicantUsecase(
	applicantRepo repositories.RecruitmentApplicantRepository,
	stageRepo repositories.ApplicantStageRepository,
	activityRepo repositories.ApplicantActivityRepository,
	recruitmentRepo repositories.RecruitmentRequestRepository,
	employeeUsecase orgUsecase.EmployeeUsecase,
) RecruitmentApplicantUsecase {
	return &recruitmentApplicantUsecase{
		applicantRepo:   applicantRepo,
		stageRepo:       stageRepo,
		activityRepo:    activityRepo,
		recruitmentRepo: recruitmentRepo,
		employeeUsecase: employeeUsecase,
	}
}

func (u *recruitmentApplicantUsecase) GetAll(ctx context.Context, params dto.ListApplicantsParams) ([]*dto.RecruitmentApplicantResponse, *response.PaginationMeta, error) {
	page := params.Page
	perPage := params.PerPage

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	applicants, total, err := u.applicantRepo.FindAll(ctx, page, perPage, params.Search, params.RecruitmentRequestID, params.StageID, params.Source)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch applicants: %w", err)
	}

	responses := make([]*dto.RecruitmentApplicantResponse, 0, len(applicants))
	for _, a := range applicants {
		responses = append(responses, toApplicantResponse(&a))
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

func (u *recruitmentApplicantUsecase) GetByID(ctx context.Context, id string) (*dto.RecruitmentApplicantResponse, error) {
	if !security.CheckRecordScopeAccess(database.DB, ctx, &models.RecruitmentApplicant{}, id, security.HRDScopeQueryOptions()) {
		return nil, errors.New("applicant not found")
	}
	applicant, err := u.applicantRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if applicant == nil {
		return nil, errors.New("applicant not found")
	}

	return toApplicantResponse(applicant), nil
}

func (u *recruitmentApplicantUsecase) Create(ctx context.Context, req *dto.CreateRecruitmentApplicantDTO, userID string) (*dto.RecruitmentApplicantResponse, error) {
	// Validate recruitment request exists
	recruitment, err := u.recruitmentRepo.FindByID(ctx, req.RecruitmentRequestID)
	if err != nil {
		return nil, err
	}
	if recruitment == nil {
		return nil, errors.New("recruitment request not found")
	}

	// Only allow adding applicants to open recruitment requests
	if recruitment.Status != models.RecruitmentStatusOpen {
		return nil, errors.New("applicants can only be added to open recruitment requests")
	}

	// Validate stage exists
	stage, err := u.stageRepo.FindByID(ctx, req.StageID)
	if err != nil {
		return nil, err
	}
	if stage == nil {
		return nil, errors.New("stage not found")
	}

	// Validate source
	if !models.IsValidSource(req.Source) {
		return nil, errors.New("invalid applicant source")
	}

	// Validate LinkedIn URL and extract username
	linkedinUsername := ""
	if req.LinkedinURL != nil && *req.LinkedinURL != "" {
		username, err := validateLinkedInURL(*req.LinkedinURL)
		if err != nil {
			return nil, err
		}
		linkedinUsername = username
	}

	now := apptime.Now()
	applicant := &models.RecruitmentApplicant{
		RecruitmentRequestID: req.RecruitmentRequestID,
		StageID:              req.StageID,
		FullName:             req.FullName,
		Email:                req.Email,
		Phone:                req.Phone,
		Source:               req.Source,
		Notes:                req.Notes,
		ResumeURL:            req.ResumeURL,
		LinkedinURL:          strPtrOrNil(linkedinUsername),
		AppliedAt:            now,
		LastActivityAt:       now,
		CreatedBy:            &userID,
	}

	if err := u.applicantRepo.Create(ctx, applicant); err != nil {
		return nil, fmt.Errorf("failed to create applicant: %w", err)
	}

	// Create activity record
	activity := &models.ApplicantActivity{
		ApplicantID: applicant.ID,
		Type:        models.ActivityTypeCreated,
		Description: fmt.Sprintf("Applicant %s was added to %s", applicant.FullName, stage.Name),
		CreatedBy:   &userID,
	}
	_ = u.activityRepo.Create(ctx, activity) // Non-critical, ignore error

	// Reload with stage
	return u.GetByID(ctx, applicant.ID)
}

func (u *recruitmentApplicantUsecase) Update(ctx context.Context, id string, req *dto.UpdateRecruitmentApplicantDTO, userID string) (*dto.RecruitmentApplicantResponse, error) {
	applicant, err := u.applicantRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if applicant == nil {
		return nil, errors.New("applicant not found")
	}

	// Cannot edit applicant after conversion to employee (#296)
	if applicant.EmployeeID != nil && *applicant.EmployeeID != "" {
		return nil, errors.New("APPLICANT_ALREADY_CONVERTED: Cannot edit applicant after conversion to employee")
	}

	// Apply updates
	if req.FullName != nil {
		applicant.FullName = *req.FullName
	}
	if req.Email != nil {
		applicant.Email = *req.Email
	}
	if req.Phone != nil {
		applicant.Phone = req.Phone
	}
	if req.Source != nil {
		if !models.IsValidSource(*req.Source) {
			return nil, errors.New("invalid applicant source")
		}
		applicant.Source = *req.Source
	}
	if req.Notes != nil {
		applicant.Notes = req.Notes
	}
	if req.ResumeURL != nil {
		applicant.ResumeURL = req.ResumeURL
	}
	if req.LinkedinURL != nil {
		username, err := validateLinkedInURL(*req.LinkedinURL)
		if err != nil {
			return nil, err
		}
		applicant.LinkedinURL = strPtrOrNil(username)
	}
	if req.Rating != nil {
		oldRating := applicant.Rating
		applicant.Rating = req.Rating

		// Create activity for rating change
		if oldRating == nil || *oldRating != *req.Rating {
			activity := &models.ApplicantActivity{
				ApplicantID: applicant.ID,
				Type:        models.ActivityTypeRatingChanged,
				Description: fmt.Sprintf("Rating changed to %d stars", *req.Rating),
				Metadata:    toJSONMetadata(map[string]any{"old_rating": oldRating, "new_rating": *req.Rating}),
				CreatedBy:   &userID,
			}
			_ = u.activityRepo.Create(ctx, activity)
		}
	}

	applicant.UpdatedBy = &userID

	if err := u.applicantRepo.Update(ctx, applicant); err != nil {
		return nil, fmt.Errorf("failed to update applicant: %w", err)
	}

	// Create activity record
	activity := &models.ApplicantActivity{
		ApplicantID: applicant.ID,
		Type:        models.ActivityTypeUpdated,
		Description: fmt.Sprintf("Applicant %s information was updated", applicant.FullName),
		CreatedBy:   &userID,
	}
	_ = u.activityRepo.Create(ctx, activity)

	return u.GetByID(ctx, applicant.ID)
}

func (u *recruitmentApplicantUsecase) Delete(ctx context.Context, id string) error {
	applicant, err := u.applicantRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if applicant == nil {
		return errors.New("applicant not found")
	}

	// Cannot delete applicant after conversion to employee (#296)
	if applicant.EmployeeID != nil && *applicant.EmployeeID != "" {
		return errors.New("APPLICANT_ALREADY_CONVERTED: Cannot delete applicant after conversion to employee")
	}

	return u.applicantRepo.Delete(ctx, id)
}

func (u *recruitmentApplicantUsecase) MoveStage(ctx context.Context, id string, req *dto.MoveApplicantStageDTO, userID string) (*dto.RecruitmentApplicantResponse, error) {
	applicant, err := u.applicantRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if applicant == nil {
		return nil, errors.New("applicant not found")
	}

	// Validate: Cannot move stage if applicant has been converted to employee
	if applicant.EmployeeID != nil && *applicant.EmployeeID != "" {
		return nil, errors.New("APPLICANT_ALREADY_CONVERTED: Cannot change status after applicant has been converted to employee")
	}

	// Get current and target stages
	fromStage := applicant.Stage
	targetStageID := req.TargetStageID()
	toStage, err := u.stageRepo.FindByID(ctx, targetStageID)
	if err != nil {
		return nil, err
	}
	if toStage == nil {
		return nil, errors.New("target stage not found")
	}

	// Move the applicant
	if err := u.applicantRepo.MoveStage(ctx, id, targetStageID); err != nil {
		return nil, err
	}

	// Determine activity type
	activityType := models.ActivityTypeStageChange
	if toStage.IsWon {
		activityType = models.ActivityTypeHired
	} else if toStage.IsLost {
		activityType = models.ActivityTypeRejected
	}

	description := fmt.Sprintf("Moved from %s to %s", fromStage.Name, toStage.Name)
	if req.Reason != nil && *req.Reason != "" {
		description = fmt.Sprintf("%s. Reason: %s", description, *req.Reason)
	}

	// Create activity record
	activity := &models.ApplicantActivity{
		ApplicantID: id,
		Type:        activityType,
		Description: description,
		Metadata:    toJSONMetadata(map[string]any{"from_stage_id": fromStage.ID, "to_stage_id": toStage.ID, "notes": req.Notes}),
		CreatedBy:   &userID,
	}
	_ = u.activityRepo.Create(ctx, activity)

	// Update filled count based on stage movement
	// Allow updates for all non-terminal recruitment statuses (not DRAFT or CANCELLED)
	recruitment, _ := u.recruitmentRepo.FindByID(ctx, applicant.RecruitmentRequestID)
	if recruitment != nil && recruitment.Status != models.RecruitmentStatusDraft && recruitment.Status != models.RecruitmentStatusCancelled {
		// Moving to Hired (Won) stage - increment filled count
		if toStage.IsWon && !fromStage.IsWon {
			recruitment.FilledCount++
			_ = u.recruitmentRepo.Update(ctx, recruitment)
		}
		// Moving from Hired (Won) to non-Won stage - decrement filled count
		if fromStage.IsWon && !toStage.IsWon {
			if recruitment.FilledCount > 0 {
				recruitment.FilledCount--
				_ = u.recruitmentRepo.Update(ctx, recruitment)
			}
		}
	}

	return u.GetByID(ctx, id)
}

func (u *recruitmentApplicantUsecase) GetByStage(ctx context.Context, params dto.ListApplicantsByStageParams) (map[string]*dto.ApplicantsByStageResponse, error) {
	page := params.Page
	perPage := params.PerPage

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	// Get all active stages
	stages, err := u.stageRepo.FindAllActive(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*dto.ApplicantsByStageResponse)

	for _, stage := range stages {
		applicants, total, err := u.applicantRepo.FindByStage(ctx, stage.ID, params.RecruitmentRequestID, page, perPage)
		if err != nil {
			continue // Skip this stage on error
		}

		applicantResponses := make([]*dto.RecruitmentApplicantResponse, 0, len(applicants))
		for _, a := range applicants {
			applicantResponses = append(applicantResponses, toApplicantResponse(&a))
		}

		result[stage.ID] = &dto.ApplicantsByStageResponse{
			StageID:    stage.ID,
			StageName:  stage.Name,
			StageColor: stage.Color,
			Order:      stage.Order,
			Applicants: applicantResponses,
			Total:      total,
		}
	}

	return result, nil
}

func (u *recruitmentApplicantUsecase) GetStages(ctx context.Context) ([]*dto.ApplicantStageResponse, error) {
	stages, err := u.stageRepo.FindAllActive(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ApplicantStageResponse, 0, len(stages))
	for _, s := range stages {
		responses = append(responses, &dto.ApplicantStageResponse{
			ID:       s.ID,
			Name:     s.Name,
			Color:    s.Color,
			Order:    s.Order,
			IsWon:    s.IsWon,
			IsLost:   s.IsLost,
			IsActive: s.IsActive,
		})
	}

	return responses, nil
}

func (u *recruitmentApplicantUsecase) GetActivities(ctx context.Context, applicantID string, page, perPage int) ([]*dto.ApplicantActivityResponse, *response.PaginationMeta, error) {
	// Verify applicant exists
	applicant, err := u.applicantRepo.FindByID(ctx, applicantID)
	if err != nil {
		return nil, nil, err
	}
	if applicant == nil {
		return nil, nil, errors.New("applicant not found")
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	activities, total, err := u.activityRepo.FindByApplicant(ctx, applicantID, page, perPage)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]*dto.ApplicantActivityResponse, 0, len(activities))
	for _, a := range activities {
		var metadata map[string]any
		if a.Metadata != nil {
			_ = json.Unmarshal(*a.Metadata, &metadata)
		}

		// Note: Creator name lookup would require N+1 queries
		// Frontend can look up user names separately using created_by ID
		var createdByName *string = nil

		responses = append(responses, &dto.ApplicantActivityResponse{
			ID:            a.ID,
			ApplicantID:   a.ApplicantID,
			Type:          a.Type,
			Description:   a.Description,
			Metadata:      metadata,
			CreatedBy:     a.CreatedBy,
			CreatedByName: createdByName,
			CreatedAt:     a.CreatedAt,
		})
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

func (u *recruitmentApplicantUsecase) AddActivity(ctx context.Context, applicantID string, req *dto.CreateApplicantActivityDTO, userID string) (*dto.ApplicantActivityResponse, error) {
	// Verify applicant exists
	applicant, err := u.applicantRepo.FindByID(ctx, applicantID)
	if err != nil {
		return nil, err
	}
	if applicant == nil {
		return nil, errors.New("applicant not found")
	}

	activity := &models.ApplicantActivity{
		ApplicantID: applicantID,
		Type:        req.Type,
		Description: req.Description,
		CreatedBy:   &userID,
	}

	if req.Metadata != nil {
		metadata, _ := json.Marshal(req.Metadata)
		activity.Metadata = (*datatypes.JSON)(&metadata)
	}

	if err := u.activityRepo.Create(ctx, activity); err != nil {
		return nil, err
	}

	return &dto.ApplicantActivityResponse{
		ID:          activity.ID,
		ApplicantID: activity.ApplicantID,
		Type:        activity.Type,
		Description: activity.Description,
		Metadata:    req.Metadata,
		CreatedBy:   &userID,
		CreatedAt:   activity.CreatedAt,
	}, nil
}

func (u *recruitmentApplicantUsecase) GetByRecruitmentRequest(ctx context.Context, recruitmentRequestID string, page, perPage int) ([]*dto.RecruitmentApplicantResponse, *response.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	applicants, total, err := u.applicantRepo.FindByRecruitmentRequest(ctx, recruitmentRequestID, page, perPage)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]*dto.RecruitmentApplicantResponse, 0, len(applicants))
	for _, a := range applicants {
		responses = append(responses, toApplicantResponse(&a))
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

// Helper functions

func toApplicantResponse(a *models.RecruitmentApplicant) *dto.RecruitmentApplicantResponse {
	if a == nil {
		return nil
	}

	resp := &dto.RecruitmentApplicantResponse{
		ID:                   a.ID,
		RecruitmentRequestID: a.RecruitmentRequestID,
		StageID:              a.StageID,
		FullName:             a.FullName,
		Email:                a.Email,
		Phone:                a.Phone,
		ResumeURL:            a.ResumeURL,
		LinkedinURL:          a.LinkedinURL,
		Source:               a.Source,
		AppliedAt:            a.AppliedAt,
		LastActivityAt:       a.LastActivityAt,
		Rating:               a.Rating,
		Notes:                a.Notes,
		EmployeeID:           a.EmployeeID,
		CreatedAt:            a.CreatedAt,
		UpdatedAt:            a.UpdatedAt,
	}

	// Stage info - check if Stage is loaded
	if a.Stage.ID != "" {
		resp.Stage = &dto.ApplicantStageResponse{
			ID:       a.Stage.ID,
			Name:     a.Stage.Name,
			Color:    a.Stage.Color,
			Order:    a.Stage.Order,
			IsWon:    a.Stage.IsWon,
			IsLost:   a.Stage.IsLost,
			IsActive: a.Stage.IsActive,
		}
	}

	// Employee info - check if EmployeeID is set
	if a.EmployeeID != nil && *a.EmployeeID != "" {
		resp.Employee = &dto.EmployeeSummaryResponse{
			ID: *a.EmployeeID,
		}
		// If employee data is loaded, populate the details
		if a.Employee != nil {
			resp.Employee.Name = a.Employee.Name
			resp.Employee.Email = a.Employee.Email
			resp.Employee.EmployeeCode = a.Employee.EmployeeCode
		}
	}

	return resp
}

func toJSONMetadata(data map[string]any) *datatypes.JSON {
	b, _ := json.Marshal(data)
	return (*datatypes.JSON)(&b)
}

func (u *recruitmentApplicantUsecase) ConvertToEmployee(
	ctx context.Context,
	applicantID string,
	req *dto.ConvertApplicantToEmployeeDTO,
	userID string,
) (*dto.RecruitmentApplicantResponse, error) {
	// 1. Get applicant with recruitment request for division/position info
	applicant, err := u.applicantRepo.FindByID(ctx, applicantID)
	if err != nil {
		return nil, errors.New("applicant not found")
	}

	// 2. Check if already converted
	if applicant.EmployeeID != nil {
		return nil, errors.New("applicant already converted to employee")
	}

	// 3. Get current stage to verify it's a "won" stage (Hired)
	stage, err := u.stageRepo.FindByID(ctx, applicant.StageID)
	if err != nil {
		return nil, errors.New("stage not found")
	}
	if !stage.IsWon {
		return nil, errors.New("applicant must be in hired stage to convert")
	}

	// 4. Get recruitment request for division and position info
	recruitment, _ := u.recruitmentRepo.FindByID(ctx, applicant.RecruitmentRequestID)

	// 5. Use applicant data as defaults for empty fields
	name := req.Name
	if name == "" {
		name = applicant.FullName
	}

	email := req.Email
	if email == "" {
		email = applicant.Email
	}

	var phone *string
	if req.Phone != "" {
		phone = &req.Phone
	} else if applicant.Phone != nil && *applicant.Phone != "" {
		phone = applicant.Phone
	}

	// Use recruitment request's division and position if available
	var divisionID, jobPositionID *string
	if req.DivisionID != nil && *req.DivisionID != "" {
		divisionID = req.DivisionID
	} else if recruitment != nil {
		divisionID = &recruitment.DivisionID
	}
	if req.JobPositionID != nil && *req.JobPositionID != "" {
		jobPositionID = req.JobPositionID
	} else if recruitment != nil {
		jobPositionID = &recruitment.PositionID
	}

	// 6. Parse date of birth
	var dob *time.Time
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return nil, errors.New("invalid date_of_birth format, expected YYYY-MM-DD")
		}
		dob = &parsed
	}

	// 7. Create employee using organization module
	isActive := true

	// Convert phone pointer to string value
	var phoneStr string
	if phone != nil {
		phoneStr = *phone
	}

	employeeReq := orgDTO.CreateEmployeeRequest{
		Name:          name,
		Email:         email,
		Phone:         phoneStr,
		DivisionID:    divisionID,
		JobPositionID: jobPositionID,
		NIK:           req.NIK,
		DateOfBirth:   dob,
		PlaceOfBirth:  req.PlaceOfBirth,
		Gender:        req.Gender,
		Religion:      req.Religion,
		Address:       req.Address,
		VillageID:     req.VillageID,
		IsActive:      &isActive,
	}

	// Only add initial contract if contract type is provided
	if req.ContractType != "" {
		employeeReq.InitialContract = &orgDTO.EmployeeContractInput{
			ContractNumber: req.ContractNumber,
			ContractType:   req.ContractType,
			StartDate:      req.StartDate,
			EndDate:        req.EndDate,
		}
	}

	employeeResp, err := u.employeeUsecase.Create(ctx, employeeReq, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	// 8. Update applicant with employee reference
	applicant.EmployeeID = &employeeResp.ID
	if err := u.applicantRepo.Update(ctx, applicant); err != nil {
		// Note: Employee was created but linking failed - log error for manual cleanup
		return nil, fmt.Errorf("failed to link applicant to employee: %w", err)
	}

	// 9. Create activity log
	activity := &models.ApplicantActivity{
		ApplicantID: applicantID,
		Type:        models.ActivityTypeConverted,
		Description: fmt.Sprintf("Converted to employee %s (%s)", employeeResp.EmployeeCode, employeeResp.Name),
		CreatedBy:   &userID,
	}
	_ = u.activityRepo.Create(ctx, activity)

	// 10. Return updated applicant
	return u.GetByID(ctx, applicantID)
}

func (u *recruitmentApplicantUsecase) CanConvertToEmployee(
	ctx context.Context,
	applicantID string,
) (bool, string, error) {
	applicant, err := u.applicantRepo.FindByID(ctx, applicantID)
	if err != nil {
		return false, "", err
	}

	if applicant.EmployeeID != nil {
		return false, "Applicant already converted to employee", nil
	}

	stage, err := u.stageRepo.FindByID(ctx, applicant.StageID)
	if err != nil {
		return false, "", err
	}

	if !stage.IsWon {
		return false, "Applicant must be in hired stage", nil
	}

	return true, "", nil
}
