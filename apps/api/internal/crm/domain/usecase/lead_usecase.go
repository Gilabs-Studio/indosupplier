package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreRepos "github.com/gilabs/gims/api/internal/core/data/repositories"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/mapper"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	orgDto "github.com/gilabs/gims/api/internal/organization/domain/dto"
	"github.com/google/uuid"
)

// phoneRegex validates phone number format:
// optional leading +, then digits, spaces, hyphens, and parentheses only.
var phoneRegex = regexp.MustCompile(`^\+?[0-9\s\-()+]*$`)

// validatePhone returns an error if the phone string contains invalid characters.
// An empty phone is always considered valid (field is optional).
func validatePhone(phone string) error {
	if phone == "" {
		return nil
	}
	if !phoneRegex.MatchString(phone) {
		return errors.New("phone number may only contain digits, +, -, spaces, and parentheses")
	}
	return nil
}

// LeadUsecase defines the interface for lead business logic
type LeadUsecase interface {
	Create(ctx context.Context, req dto.CreateLeadRequest, createdBy string) (dto.LeadResponse, error)
	GetByID(ctx context.Context, id string) (dto.LeadResponse, error)
	List(ctx context.Context, params repositories.LeadListParams) ([]dto.LeadResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateLeadRequest, updatedBy string) (dto.LeadResponse, error)
	Delete(ctx context.Context, id string) error
	Convert(ctx context.Context, id string, req dto.ConvertLeadRequest, convertedBy string) (dto.LeadResponse, error)
	BulkUpsert(ctx context.Context, req dto.BulkUpsertLeadRequest, createdBy string) (*dto.BulkUpsertLeadResponse, error)
	GetUnprocessed(ctx context.Context, limit int) ([]dto.LeadResponse, error)
	GetFormData(ctx context.Context) (*dto.LeadFormDataResponse, error)
	GetAnalytics(ctx context.Context) (*repositories.LeadAnalytics, error)
	GetProductItems(ctx context.Context, leadID string) ([]dto.LeadProductItemResponse, error)
}

var ErrCannotManuallySetConvertedStatus = errors.New("cannot manually set converted status, use the convert endpoint")

const (
	leadSourceGoogleMapsID = "cb000001-0000-0000-0000-000000000006"
	leadSourceLinkedInID   = "cb000001-0000-0000-0000-000000000007"
)

type leadUsecase struct {
	leadRepo          repositories.LeadRepository
	leadStatusRepo    repositories.LeadStatusRepository
	leadSourceRepo    repositories.LeadSourceRepository
	contactRoleRepo   repositories.ContactRoleRepository
	dealRepo          repositories.DealRepository
	pipelineStageRepo repositories.PipelineStageRepository
	activityRepo      repositories.ActivityRepository
	taskRepo          repositories.TaskRepository
	employeeRepo      orgRepos.EmployeeRepository
	businessTypeRepo  orgRepos.BusinessTypeRepository
	areaRepo          orgRepos.AreaRepository
	paymentTermsRepo  coreRepos.PaymentTermsRepository
}

// NewLeadUsecase creates a new lead usecase
func NewLeadUsecase(
	leadRepo repositories.LeadRepository,
	leadStatusRepo repositories.LeadStatusRepository,
	leadSourceRepo repositories.LeadSourceRepository,
	contactRoleRepo repositories.ContactRoleRepository,
	dealRepo repositories.DealRepository,
	pipelineStageRepo repositories.PipelineStageRepository,
	activityRepo repositories.ActivityRepository,
	taskRepo repositories.TaskRepository,
	employeeRepo orgRepos.EmployeeRepository,
	businessTypeRepo orgRepos.BusinessTypeRepository,
	areaRepo orgRepos.AreaRepository,
	paymentTermsRepo coreRepos.PaymentTermsRepository,
) LeadUsecase {
	return &leadUsecase{
		leadRepo:          leadRepo,
		leadStatusRepo:    leadStatusRepo,
		leadSourceRepo:    leadSourceRepo,
		contactRoleRepo:   contactRoleRepo,
		dealRepo:          dealRepo,
		pipelineStageRepo: pipelineStageRepo,
		activityRepo:      activityRepo,
		taskRepo:          taskRepo,
		employeeRepo:      employeeRepo,
		businessTypeRepo:  businessTypeRepo,
		areaRepo:          areaRepo,
		paymentTermsRepo:  paymentTermsRepo,
	}
}

func (u *leadUsecase) Create(ctx context.Context, req dto.CreateLeadRequest, createdBy string) (dto.LeadResponse, error) {
	// Validate lead source if provided
	if req.LeadSourceID != nil && *req.LeadSourceID != "" {
		_, err := u.leadSourceRepo.FindByID(ctx, *req.LeadSourceID)
		if err != nil {
			return dto.LeadResponse{}, errors.New("lead source not found")
		}
	}

	// Assign default lead status if not provided
	var statusID *string
	if req.LeadStatusID != nil && *req.LeadStatusID != "" {
		_, err := u.leadStatusRepo.FindByID(ctx, *req.LeadStatusID)
		if err != nil {
			return dto.LeadResponse{}, errors.New("lead status not found")
		}
		statusID = req.LeadStatusID
	} else {
		// Use the default status (e.g., "New")
		defaultStatus, err := u.leadStatusRepo.FindDefault(ctx)
		if err == nil && defaultStatus != nil {
			statusID = &defaultStatus.ID
		}
	}

	// Validate assigned employee if provided
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		_, err := u.employeeRepo.FindByID(ctx, *req.AssignedTo)
		if err != nil {
			return dto.LeadResponse{}, errors.New("assigned employee not found")
		}
	}

	// Validate contact role if provided
	if req.ContactRoleID != nil && *req.ContactRoleID != "" {
		_, err := u.contactRoleRepo.FindByID(ctx, *req.ContactRoleID)
		if err != nil {
			return dto.LeadResponse{}, errors.New("contact role not found")
		}
	}

	// Validate phone format
	if err := validatePhone(req.Phone); err != nil {
		return dto.LeadResponse{}, err
	}

	// Parse time expected
	var timeExpected *time.Time
	if req.TimeExpected != nil && *req.TimeExpected != "" {
		t, err := time.Parse("2006-01-02", *req.TimeExpected)
		if err != nil {
			return dto.LeadResponse{}, errors.New("invalid time_expected format, use YYYY-MM-DD")
		}
		timeExpected = &t
	}

	// Extract tenant_id from context
	tenantID, _ := ctx.Value("tenant_id").(string)

	lead := &models.Lead{
		ID:                   uuid.New().String(),
		FirstName:            req.FirstName,
		LastName:             req.LastName,
		CompanyName:          req.CompanyName,
		Email:                req.Email,
		Phone:                req.Phone,
		ContactRoleID:        req.ContactRoleID,
		JobTitle:             req.JobTitle,
		Address:              req.Address,
		City:                 req.City,
		Province:             req.Province,
		ProvinceID:           req.ProvinceID,
		CityID:               req.CityID,
		DistrictID:           req.DistrictID,
		VillageName:          req.VillageName,
		LeadSourceID:         req.LeadSourceID,
		LeadStatusID:         statusID,
		EstimatedValue:       req.EstimatedValue,
		Probability:          req.Probability,
		Website:              req.Website,
		BankAccountID:        req.BankAccountID,
		BankAccountReference: req.BankAccountReference,
		Latitude:             req.Latitude,
		Longitude:            req.Longitude,
		BudgetConfirmed:      req.BudgetConfirmed,
		BudgetAmount:         req.BudgetAmount,
		AuthConfirmed:        req.AuthConfirmed,
		AuthPerson:           req.AuthPerson,
		NeedConfirmed:        req.NeedConfirmed,
		NeedDescription:      req.NeedDescription,
		TimeConfirmed:        req.TimeConfirmed,
		TimeExpected:         timeExpected,
		AssignedTo:           req.AssignedTo,
		Notes:                req.Notes,
		BusinessTypeID:       req.BusinessTypeID,
		AreaID:               req.AreaID,
		PaymentTermsID:       req.PaymentTermsID,
		CreatedBy:            &createdBy,
	}
	lead.TenantID = tenantID

	// Calculate lead score after setting all fields
	// Need to preload status for score calculation
	if statusID != nil {
		status, _ := u.leadStatusRepo.FindByID(ctx, *statusID)
		if status != nil {
			lead.LeadStatus = status
		}
	}
	lead.LeadScore = lead.CalculateLeadScore()

	if err := u.leadRepo.Create(ctx, lead); err != nil {
		return dto.LeadResponse{}, fmt.Errorf("failed to create lead: %w", err)
	}

	// Reload with preloaded relations
	created, err := u.leadRepo.FindByID(ctx, lead.ID)
	if err != nil {
		return dto.LeadResponse{}, err
	}

	return mapper.ToLeadResponse(created), nil
}

func (u *leadUsecase) GetByID(ctx context.Context, id string) (dto.LeadResponse, error) {
	if !security.CheckRecordScopeAccess(database.DB, ctx, &models.Lead{}, id, security.MixedOwnershipScopeQueryOptions("assigned_to")) {
		return dto.LeadResponse{}, errors.New("lead not found")
	}

	lead, err := u.leadRepo.FindByID(ctx, id)
	if err != nil {
		return dto.LeadResponse{}, errors.New("lead not found")
	}
	return mapper.ToLeadResponse(lead), nil
}

func (u *leadUsecase) List(ctx context.Context, params repositories.LeadListParams) ([]dto.LeadResponse, int64, error) {
	leads, total, err := u.leadRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToLeadResponseList(leads), total, nil
}

func (u *leadUsecase) Update(ctx context.Context, id string, req dto.UpdateLeadRequest, updatedBy string) (dto.LeadResponse, error) {
	lead, err := u.leadRepo.FindByID(ctx, id)
	if err != nil {
		return dto.LeadResponse{}, errors.New("lead not found")
	}

	// Track fields that changed so we can log an activity entry for auditing.
	changedFields := make([]string, 0)
	statusChanged := false
	oldStatusName := ""
	newStatusName := ""
	if lead.LeadStatus != nil {
		oldStatusName = lead.LeadStatus.Name
	}
	oldStatusID := lead.LeadStatusID

	addStringChange := func(field, oldVal, newVal string) {
		if oldVal != newVal {
			changedFields = append(changedFields, field)
		}
	}

	addBoolChange := func(field string, oldVal, newVal bool) {
		if oldVal != newVal {
			changedFields = append(changedFields, field)
		}
	}

	// Prevent updates on converted leads; geographical coordinates are the only allowed exception
	// since an active deal may need to pin the customer's location.
	// Also check DealID to catch leads whose ConvertedAt was not persisted due to a prior error.
	if lead.ConvertedAt != nil || lead.DealID != nil {
		if req.Latitude == nil && req.Longitude == nil {
			return dto.LeadResponse{}, errors.New("cannot update a converted lead")
		}
		// Coordinate-only update path: apply and return early
		if req.Latitude != nil {
			lead.Latitude = req.Latitude
		}
		if req.Longitude != nil {
			lead.Longitude = req.Longitude
		}
		// Nil associations to prevent GORM FK side-effects
		lead.LeadSource = nil
		lead.LeadStatus = nil
		lead.AssignedEmployee = nil
		lead.Customer = nil
		lead.Contact = nil
		lead.Deal = nil
		lead.BusinessType = nil
		lead.Area = nil
		lead.PaymentTerms = nil
		lead.Activities = nil
		if err := u.leadRepo.Update(ctx, lead); err != nil {
			return dto.LeadResponse{}, fmt.Errorf("failed to update lead location: %w", err)
		}
		updated, err := u.leadRepo.FindByID(ctx, id)
		if err != nil {
			return dto.LeadResponse{}, err
		}
		return mapper.ToLeadResponse(updated), nil
	}

	// Validate lead source if changing
	if req.LeadSourceID != nil && *req.LeadSourceID != "" {
		_, err := u.leadSourceRepo.FindByID(ctx, *req.LeadSourceID)
		if err != nil {
			return dto.LeadResponse{}, errors.New("lead source not found")
		}
		lead.LeadSourceID = req.LeadSourceID
	}

	// Validate lead status if changing
	if req.LeadStatusID != nil && *req.LeadStatusID != "" {
		status, err := u.leadStatusRepo.FindByID(ctx, *req.LeadStatusID)
		if err != nil {
			return dto.LeadResponse{}, errors.New("lead status not found")
		}
		// Prevent manual assignment of converted status
		if status.IsConverted {
			return dto.LeadResponse{}, ErrCannotManuallySetConvertedStatus
		}

		// Track status changes for activity logging
		if oldStatusID == nil || *oldStatusID != *req.LeadStatusID {
			statusChanged = true
			newStatusName = status.Name
		}

		lead.LeadStatusID = req.LeadStatusID
		lead.LeadStatus = status
	}

	// Validate assigned employee if changing
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		_, err := u.employeeRepo.FindByID(ctx, *req.AssignedTo)
		if err != nil {
			return dto.LeadResponse{}, errors.New("assigned employee not found")
		}
		lead.AssignedTo = req.AssignedTo
	}

	// Validate contact role if changing
	if req.ContactRoleID != nil && *req.ContactRoleID != "" {
		_, err := u.contactRoleRepo.FindByID(ctx, *req.ContactRoleID)
		if err != nil {
			return dto.LeadResponse{}, errors.New("contact role not found")
		}
	}

	// Apply partial updates
	if req.FirstName != nil {
		addStringChange("first_name", lead.FirstName, *req.FirstName)
		lead.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		addStringChange("last_name", lead.LastName, *req.LastName)
		lead.LastName = *req.LastName
	}
	if req.CompanyName != nil {
		addStringChange("company_name", lead.CompanyName, *req.CompanyName)
		lead.CompanyName = *req.CompanyName
	}
	if req.Email != nil {
		addStringChange("email", lead.Email, *req.Email)
		lead.Email = *req.Email
	}
	if req.Phone != nil {
		if err := validatePhone(*req.Phone); err != nil {
			return dto.LeadResponse{}, err
		}
		addStringChange("phone", lead.Phone, *req.Phone)
		lead.Phone = *req.Phone
	}
	if req.ContactRoleID != nil {
		old := ""
		if lead.ContactRoleID != nil {
			old = *lead.ContactRoleID
		}
		newVal := ""
		if *req.ContactRoleID != "" {
			newVal = *req.ContactRoleID
		}
		addStringChange("contact_role_id", old, newVal)
		lead.ContactRoleID = req.ContactRoleID
	}
	if req.JobTitle != nil {
		addStringChange("job_title", lead.JobTitle, *req.JobTitle)
		lead.JobTitle = *req.JobTitle
	}
	if req.Address != nil {
		addStringChange("address", lead.Address, *req.Address)
		lead.Address = *req.Address
	}
	if req.City != nil {
		addStringChange("city", lead.City, *req.City)
		lead.City = *req.City
	}
	if req.Province != nil {
		addStringChange("province", lead.Province, *req.Province)
		lead.Province = *req.Province
	}
	if req.ProvinceID != nil {
		// Only record if changed (nil vs non-nil or different value)
		old := ""
		if lead.ProvinceID != nil {
			old = *lead.ProvinceID
		}
		newVal := ""
		if *req.ProvinceID != "" {
			newVal = *req.ProvinceID
		}
		addStringChange("province_id", old, newVal)
		lead.ProvinceID = req.ProvinceID
	}
	if req.CityID != nil {
		old := ""
		if lead.CityID != nil {
			old = *lead.CityID
		}
		newVal := ""
		if *req.CityID != "" {
			newVal = *req.CityID
		}
		addStringChange("city_id", old, newVal)
		lead.CityID = req.CityID
	}
	if req.DistrictID != nil {
		old := ""
		if lead.DistrictID != nil {
			old = *lead.DistrictID
		}
		newVal := ""
		if *req.DistrictID != "" {
			newVal = *req.DistrictID
		}
		addStringChange("district_id", old, newVal)
		lead.DistrictID = req.DistrictID
	}
	if req.VillageName != nil {
		addStringChange("village_name", lead.VillageName, *req.VillageName)
		lead.VillageName = *req.VillageName
	}
	if req.Website != nil {
		addStringChange("website", lead.Website, *req.Website)
		lead.Website = *req.Website
	}
	if req.BankAccountID != nil {
		old := ""
		if lead.BankAccountID != nil {
			old = *lead.BankAccountID
		}
		newVal := ""
		if *req.BankAccountID != "" {
			newVal = *req.BankAccountID
		}
		addStringChange("bank_account_id", old, newVal)
		lead.BankAccountID = req.BankAccountID
	}
	if req.BankAccountReference != nil {
		addStringChange("bank_account_reference", lead.BankAccountReference, *req.BankAccountReference)
		lead.BankAccountReference = *req.BankAccountReference
	}
	if req.EstimatedValue != nil {
		if lead.EstimatedValue != *req.EstimatedValue {
			changedFields = append(changedFields, "estimated_value")
		}
		lead.EstimatedValue = *req.EstimatedValue
	}
	if req.Probability != nil {
		if lead.Probability != *req.Probability {
			changedFields = append(changedFields, "probability")
		}
		lead.Probability = *req.Probability
	}
	if req.BudgetConfirmed != nil {
		addBoolChange("budget_confirmed", lead.BudgetConfirmed, *req.BudgetConfirmed)
		lead.BudgetConfirmed = *req.BudgetConfirmed
	}
	if req.BudgetAmount != nil {
		if lead.BudgetAmount != *req.BudgetAmount {
			changedFields = append(changedFields, "budget_amount")
		}
		lead.BudgetAmount = *req.BudgetAmount
	}
	if req.AuthConfirmed != nil {
		addBoolChange("auth_confirmed", lead.AuthConfirmed, *req.AuthConfirmed)
		lead.AuthConfirmed = *req.AuthConfirmed
	}
	if req.AuthPerson != nil {
		addStringChange("auth_person", lead.AuthPerson, *req.AuthPerson)
		lead.AuthPerson = *req.AuthPerson
	}
	if req.NeedConfirmed != nil {
		addBoolChange("need_confirmed", lead.NeedConfirmed, *req.NeedConfirmed)
		lead.NeedConfirmed = *req.NeedConfirmed
	}
	if req.NeedDescription != nil {
		addStringChange("need_description", lead.NeedDescription, *req.NeedDescription)
		lead.NeedDescription = *req.NeedDescription
	}
	if req.TimeConfirmed != nil {
		addBoolChange("time_confirmed", lead.TimeConfirmed, *req.TimeConfirmed)
		lead.TimeConfirmed = *req.TimeConfirmed
	}
	if req.TimeExpected != nil && *req.TimeExpected != "" {
		t, err := time.Parse("2006-01-02", *req.TimeExpected)
		if err != nil {
			return dto.LeadResponse{}, errors.New("invalid time_expected format, use YYYY-MM-DD")
		}
		if lead.TimeExpected == nil || !lead.TimeExpected.Equal(t) {
			changedFields = append(changedFields, "time_expected")
		}
		lead.TimeExpected = &t
	}
	if req.Notes != nil {
		addStringChange("notes", lead.Notes, *req.Notes)
		lead.Notes = *req.Notes
	}
	if req.NPWP != nil {
		addStringChange("npwp", lead.NPWP, *req.NPWP)
		lead.NPWP = *req.NPWP
	}
	if req.Latitude != nil {
		if lead.Latitude == nil || *lead.Latitude != *req.Latitude {
			changedFields = append(changedFields, "latitude")
		}
		lead.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		if lead.Longitude == nil || *lead.Longitude != *req.Longitude {
			changedFields = append(changedFields, "longitude")
		}
		lead.Longitude = req.Longitude
	}
	if req.BusinessTypeID != nil {
		old := ""
		if lead.BusinessTypeID != nil {
			old = *lead.BusinessTypeID
		}
		newVal := ""
		if *req.BusinessTypeID != "" {
			newVal = *req.BusinessTypeID
		}
		addStringChange("business_type_id", old, newVal)
		lead.BusinessTypeID = req.BusinessTypeID
	}
	if req.AreaID != nil {
		old := ""
		if lead.AreaID != nil {
			old = *lead.AreaID
		}
		newVal := ""
		if *req.AreaID != "" {
			newVal = *req.AreaID
		}
		addStringChange("area_id", old, newVal)
		lead.AreaID = req.AreaID
	}
	if req.PaymentTermsID != nil {
		old := ""
		if lead.PaymentTermsID != nil {
			old = *lead.PaymentTermsID
		}
		newVal := ""
		if *req.PaymentTermsID != "" {
			newVal = *req.PaymentTermsID
		}
		addStringChange("payment_terms_id", old, newVal)
		lead.PaymentTermsID = req.PaymentTermsID
	}

	// Recalculate lead score after updates
	lead.LeadScore = lead.CalculateLeadScore()

	// Nil out preloaded associations to prevent GORM FullSaveAssociations from
	// attempting to upsert associated records (BelongsTo FK override, HasMany upsert).
	lead.LeadSource = nil
	lead.LeadStatus = nil
	lead.ContactRole = nil
	lead.AssignedEmployee = nil
	lead.Customer = nil
	lead.Contact = nil
	lead.Deal = nil
	lead.BusinessType = nil
	lead.Area = nil
	lead.PaymentTerms = nil
	lead.Activities = nil
	lead.Tasks = nil
	lead.ProductItems = nil

	if err := u.leadRepo.Update(ctx, lead); err != nil {
		return dto.LeadResponse{}, fmt.Errorf("failed to update lead: %w", err)
	}

	// Log activity for lead edits and status transitions (best-effort)
	// This keeps account of changes in the audit trail.
	if updatedBy == "" {
		updatedBy = "system"
	}

	// Determine description for activity
	var activityDescription string
	if len(changedFields) > 0 {
		activityDescription = fmt.Sprintf("Updated lead fields: %s", strings.Join(changedFields, ", "))
	}
	if statusChanged {
		statusDesc := fmt.Sprintf("Status changed from %s to %s", oldStatusName, newStatusName)
		if activityDescription != "" {
			activityDescription = fmt.Sprintf("%s; %s", activityDescription, statusDesc)
		} else {
			activityDescription = statusDesc
		}
	}

	if activityDescription != "" {
		activityType := "lead_update"
		if statusChanged {
			activityType = "lead_status_change"
		}

		activity := &models.Activity{
			Type:           activityType,
			ActivityTypeID: strPtr(activityTypeFollowUpID),
			LeadID:         &lead.ID,
			EmployeeID:     updatedBy,
			Description:    activityDescription,
			Timestamp:      apptime.Now(),
		}
		_ = u.activityRepo.Create(ctx, activity)
	}

	// Reload with preloaded relations
	updated, err := u.leadRepo.FindByID(ctx, id)
	if err != nil {
		return dto.LeadResponse{}, err
	}

	return mapper.ToLeadResponse(updated), nil
}

func (u *leadUsecase) Delete(ctx context.Context, id string) error {
	lead, err := u.leadRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("lead not found")
	}

	// Prevent deletion of converted leads
	if lead.ConvertedAt != nil {
		return errors.New("cannot delete a converted lead")
	}

	// Cascade soft-delete related deals
	if err := database.DB.WithContext(ctx).Where("lead_id = ?", id).Delete(&models.Deal{}).Error; err != nil {
		return err
	}

	return u.leadRepo.Delete(ctx, id)
}

// Convert transforms a lead into a deal in the pipeline, updating the lead's conversion fields
func (u *leadUsecase) Convert(ctx context.Context, id string, req dto.ConvertLeadRequest, convertedBy string) (dto.LeadResponse, error) {
	lead, err := u.leadRepo.FindByID(ctx, id)
	if err != nil {
		return dto.LeadResponse{}, errors.New("lead not found")
	}

	// Prevent double conversion: check the lead's own fields first, then scan for an
	// orphaned deal (handles broken-state leads where the deal was created but the lead
	// record was not updated before a previous request failure).
	if lead.ConvertedAt != nil || lead.DealID != nil {
		return dto.LeadResponse{}, errors.New("lead already converted")
	}
	orphanExists, _ := u.dealRepo.ExistsByLeadID(ctx, lead.ID)
	if orphanExists {
		return dto.LeadResponse{}, errors.New("lead already converted")
	}

	// Prevent conversion of lost leads
	if lead.LeadStatus != nil && lead.LeadStatus.Code == "LOST" {
		return dto.LeadResponse{}, errors.New("cannot convert a lost lead")
	}

	// Resolve pipeline stage: use provided stage or default to lowest-probability active pipeline stage
	var pipelineStageID string
	if req.PipelineStageID != nil && *req.PipelineStageID != "" {
		stage, err := u.pipelineStageRepo.FindByID(ctx, *req.PipelineStageID)
		if err != nil {
			return dto.LeadResponse{}, errors.New("pipeline stage not found")
		}
		pipelineStageID = stage.ID
	} else {
		stages, _, err := u.pipelineStageRepo.List(ctx, repositories.ListParams{Limit: 100})
		if err != nil {
			return dto.LeadResponse{}, fmt.Errorf("failed to load pipeline stages: %w", err)
		}
		if len(stages) == 0 {
			return dto.LeadResponse{}, errors.New("no pipeline stages available")
		}

		// Select the active stage with the highest probability that does not exceed
		// the lead's probability. This mirrors where the lead currently stands in the
		// sales cycle. For example: lead at 70% with stages 20%, 30%, 80%, 100%
		// → stage at 30% is chosen (highest that is still ≤ 70%).
		// If no stage probability is at or below the lead's, fall back to the first
		// active non-won/non-lost stage by order.
		var defaultStage *models.PipelineStage
		for i := range stages {
			stage := &stages[i]
			if stage.IsWon || stage.IsLost || !stage.IsActive {
				continue
			}
			if stage.Probability <= lead.Probability {
				if defaultStage == nil || stage.Probability > defaultStage.Probability {
					defaultStage = stage
				}
			}
		}
		// Fallback: no stage probability ≤ lead's (e.g. all stages higher) — pick first by order.
		if defaultStage == nil {
			for i := range stages {
				stage := &stages[i]
				if !stage.IsWon && !stage.IsLost && stage.IsActive {
					defaultStage = stage
					break
				}
			}
		}

		if defaultStage == nil {
			// Fallback: take first stage if no non-won/non-lost stage exists
			defaultStage = &stages[0]
		}

		pipelineStageID = defaultStage.ID
	}

	// Build deal title from request or lead data
	dealTitle := req.DealTitle
	if dealTitle == "" {
		if lead.CompanyName != "" {
			dealTitle = lead.CompanyName
		} else {
			dealTitle = lead.FirstName
			if lead.LastName != "" {
				dealTitle += " " + lead.LastName
			}
		}
	}

	// Determine deal value
	dealValue := lead.EstimatedValue
	if req.DealValue != nil {
		dealValue = *req.DealValue
	}

	now := apptime.Now()

	// Create deal from lead data
	newDeal := &models.Deal{
		ID:              uuid.New().String(),
		Title:           dealTitle,
		Status:          models.DealStatusOpen,
		PipelineStageID: pipelineStageID,
		Value:           dealValue,
		Probability:     lead.Probability,
		LeadID:          &lead.ID,
		AssignedTo:      lead.AssignedTo,
		BudgetConfirmed: lead.BudgetConfirmed,
		BudgetAmount:    lead.BudgetAmount,
		AuthConfirmed:   lead.AuthConfirmed,
		AuthPerson:      lead.AuthPerson,
		NeedConfirmed:   lead.NeedConfirmed,
		NeedDescription: lead.NeedDescription,
		TimeConfirmed:   lead.TimeConfirmed,
		// Use the lead's original notes so deal mirrors lead data exactly
		Notes:     lead.Notes,
		CreatedBy: &convertedBy,
	}

	if lead.TimeExpected != nil {
		newDeal.ExpectedCloseDate = lead.TimeExpected
	}

	if err := u.dealRepo.Create(ctx, newDeal); err != nil {
		return dto.LeadResponse{}, fmt.Errorf("failed to create deal from lead: %w", err)
	}

	// Create initial stage history entry so the Information tab is never empty
	initialHistory := &models.DealHistory{
		DealID:        newDeal.ID,
		ToStageID:     pipelineStageID,
		ToProbability: newDeal.Probability,
		ChangedAt:     now,
		Reason:        fmt.Sprintf("Converted from lead %s", lead.Code),
	}
	// best-effort — do not block conversion if history creation fails
	_ = u.dealRepo.CreateHistory(ctx, initialHistory)

	// Copy lead product items to deal product items, preserving interest level and eliminated state
	leadProducts, lpErr := u.leadRepo.ListProductItems(ctx, lead.ID)
	if lpErr == nil && len(leadProducts) > 0 {
		dealItems := make([]models.DealProductItem, 0, len(leadProducts))
		eliminatedIDs := make([]string, 0)
		for _, lp := range leadProducts {
			itemID := uuid.New().String()
			item := models.DealProductItem{
				ID:            itemID,
				DealID:        newDeal.ID,
				ProductID:     lp.ProductID,
				ProductName:   lp.ProductName,
				ProductSKU:    lp.ProductSKU,
				InterestLevel: lp.InterestLevel,
				UnitPrice:     lp.UnitPrice,
				Quantity:      lp.Quantity,
				Notes:         lp.Notes,
			}
			item.Subtotal = item.UnitPrice * float64(item.Quantity)
			dealItems = append(dealItems, item)
			// Track items that were eliminated in the lead (soft-deleted)
			if lp.DeletedAt.Valid {
				eliminatedIDs = append(eliminatedIDs, itemID)
			}
		}
		// best-effort — do not block conversion if product copy fails
		_ = u.dealRepo.CreateItems(ctx, dealItems)
		// Preserve eliminated state from lead items in the deal
		for _, id := range eliminatedIDs {
			_ = u.dealRepo.SoftDeleteItemByID(ctx, id, newDeal.ID)
		}
	}

	// Associate lead tasks with the new deal so they appear in the deal task tab.
	// best-effort — do not block conversion if this fails
	_ = u.taskRepo.UpdateDealIDByLeadID(ctx, lead.ID, newDeal.ID)

	// Create a special immutable activity recording the conversion — best-effort, never blocks conversion
	if convertedBy != "" {
		conversionActivity := &models.Activity{
			Type:        "conversion",
			DealID:      &newDeal.ID,
			LeadID:      &lead.ID,
			EmployeeID:  convertedBy,
			Description: fmt.Sprintf("Lead %s (%s %s) dikonversi ke deal pipeline %s", lead.Code, lead.FirstName, lead.LastName, newDeal.Code),
			Timestamp:   now,
		}
		_ = u.activityRepo.Create(ctx, conversionActivity)
	}

	// Update lead with conversion data
	lead.DealID = &newDeal.ID
	lead.ConvertedAt = &now
	lead.ConvertedBy = &convertedBy

	// Set lead status to "Converted"
	convertedStatus, err := u.leadStatusRepo.FindConverted(ctx)
	if err == nil && convertedStatus != nil {
		lead.LeadStatusID = &convertedStatus.ID
		lead.LeadStatus = convertedStatus
	}

	lead.LeadScore = lead.CalculateLeadScore()

	// Nil out preloaded associations to prevent GORM FullSaveAssociations from
	// attempting to upsert associated records (BelongsTo FK override, HasMany upsert).
	lead.LeadSource = nil
	lead.LeadStatus = nil
	lead.ContactRole = nil
	lead.AssignedEmployee = nil
	lead.Customer = nil
	lead.Contact = nil
	lead.Deal = nil
	lead.BusinessType = nil
	lead.Area = nil
	lead.PaymentTerms = nil
	lead.Activities = nil
	lead.Tasks = nil
	lead.ProductItems = nil

	if err := u.leadRepo.Update(ctx, lead); err != nil {
		return dto.LeadResponse{}, fmt.Errorf("failed to update lead conversion: %w", err)
	}

	// Reload with all preloaded relations
	converted, err := u.leadRepo.FindByID(ctx, lead.ID)
	if err != nil {
		return dto.LeadResponse{}, err
	}

	return mapper.ToLeadResponse(converted), nil
}

// BulkUpsert creates or updates leads in bulk, using email as the deduplication key.
// Designed for automation workflows (e.g., n8n) that scrape leads from external sources.
func (u *leadUsecase) BulkUpsert(ctx context.Context, req dto.BulkUpsertLeadRequest, createdBy string) (*dto.BulkUpsertLeadResponse, error) {
	result := &dto.BulkUpsertLeadResponse{
		Items: make([]dto.LeadResponse, 0, len(req.Leads)),
	}

	// Resolve default status once for all new leads
	var defaultStatusID *string
	defaultStatus, err := u.leadStatusRepo.FindDefault(ctx)
	if err != nil || defaultStatus == nil {
		fallbackStatus, fallbackErr := u.leadStatusRepo.FindByCode(ctx, "NEW")
		if fallbackErr == nil && fallbackStatus != nil {
			defaultStatus = fallbackStatus
			err = nil
		}
	}
	if err == nil && defaultStatus != nil {
		defaultStatusID = &defaultStatus.ID
	}

	for _, item := range req.Leads {
		// Resolve lead source defensively so list column doesn't show empty source for automation leads.
		resolvedLeadSourceID := item.LeadSourceID
		if resolvedLeadSourceID == nil || *resolvedLeadSourceID == "" {
			inferred := inferLeadSourceID(item)
			if inferred != "" {
				resolvedLeadSourceID = &inferred
			}
		}
		if resolvedLeadSourceID != nil && *resolvedLeadSourceID != "" {
			if _, sourceErr := u.leadSourceRepo.FindByID(ctx, *resolvedLeadSourceID); sourceErr != nil {
				inferred := inferLeadSourceID(item)
				if inferred != "" {
					resolvedLeadSourceID = &inferred
				} else {
					resolvedLeadSourceID = nil
				}
			}
		}

		// Try to find existing lead by place_id, cid, email, phone, or company_name (deduplication key)
		existing, findErr := u.leadRepo.FindDuplicate(ctx, item.Email, item.Phone, item.CompanyName, item.PlaceID, item.CID)

		typesStr := ""
		if item.Types != nil {
			if s, ok := item.Types.(string); ok {
				typesStr = s
			} else {
				b, _ := json.Marshal(item.Types)
				typesStr = string(b)
			}
		}

		openingHoursStr := ""
		if item.OpeningHours != nil {
			if s, ok := item.OpeningHours.(string); ok {
				openingHoursStr = s
			} else {
				b, _ := json.Marshal(item.OpeningHours)
				openingHoursStr = string(b)
			}
		}

		if findErr == nil && existing != nil {
			// Update existing lead with new data (merge non-empty fields)
			if defaultStatusID != nil && (existing.LeadStatusID == nil || *existing.LeadStatusID == "") {
				existing.LeadStatusID = defaultStatusID
			}
			if item.FirstName != "" {
				existing.FirstName = item.FirstName
			}
			if item.LastName != "" {
				existing.LastName = item.LastName
			}
			if item.CompanyName != "" {
				existing.CompanyName = item.CompanyName
			}
			if item.Phone != "" {
				existing.Phone = item.Phone
			}
			if item.JobTitle != "" {
				existing.JobTitle = item.JobTitle
			}
			if item.Address != "" {
				existing.Address = item.Address
			}
			if item.City != "" {
				existing.City = item.City
			}
			if item.Province != "" {
				existing.Province = item.Province
			}
			if item.ProvinceID != nil && *item.ProvinceID != "" {
				existing.ProvinceID = item.ProvinceID
			}
			if item.CityID != nil && *item.CityID != "" {
				existing.CityID = item.CityID
			}
			if item.DistrictID != nil && *item.DistrictID != "" {
				existing.DistrictID = item.DistrictID
			}
			if item.VillageName != "" {
				existing.VillageName = item.VillageName
			}
			if item.EstimatedValue > 0 {
				existing.EstimatedValue = item.EstimatedValue
			}
			if resolvedLeadSourceID != nil && *resolvedLeadSourceID != "" {
				existing.LeadSourceID = resolvedLeadSourceID
			}
			if item.Latitude != nil {
				existing.Latitude = item.Latitude
			}
			if item.Longitude != nil {
				existing.Longitude = item.Longitude
			}
			if item.Rating != nil {
				existing.Rating = item.Rating
			}
			if item.RatingCount != nil {
				existing.RatingCount = item.RatingCount
			}
			if typesStr != "" {
				existing.Types = typesStr
			}
			if openingHoursStr != "" {
				existing.OpeningHours = openingHoursStr
			}
			if item.ThumbnailURL != "" {
				existing.ThumbnailURL = item.ThumbnailURL
			}
			if item.CID != "" {
				existing.CID = item.CID
			}
			if item.PlaceID != "" {
				existing.PlaceID = item.PlaceID
			}
			if item.Website != "" {
				existing.Website = item.Website
			}
			if item.Notes != "" {
				existing.Notes = existing.Notes + "\n---\n" + item.Notes
			}

			// Mark as processed by n8n
			now := apptime.Now()
			existing.ProcessedFromN8N = true
			existing.ProcessedAt = &now

			existing.LeadScore = existing.CalculateLeadScore()

			if updateErr := u.leadRepo.Update(ctx, existing); updateErr != nil {
				result.Errors++
				continue
			}

			reloaded, reloadErr := u.leadRepo.FindByID(ctx, existing.ID)
			if reloadErr != nil {
				result.Errors++
				continue
			}

			result.Updated++
			result.Items = append(result.Items, mapper.ToLeadResponse(reloaded))
		} else {
			// Create new lead
			now := apptime.Now()
			lead := &models.Lead{
				ID:               uuid.New().String(),
				FirstName:        item.FirstName,
				LastName:         item.LastName,
				CompanyName:      item.CompanyName,
				Email:            item.Email,
				Phone:            item.Phone,
				JobTitle:         item.JobTitle,
				Address:          item.Address,
				City:             item.City,
				Province:         item.Province,
				ProvinceID:       item.ProvinceID,
				CityID:           item.CityID,
				DistrictID:       item.DistrictID,
				VillageName:      item.VillageName,
				LeadSourceID:     resolvedLeadSourceID,
				LeadStatusID:     defaultStatusID,
				EstimatedValue:   item.EstimatedValue,
				Latitude:         item.Latitude,
				Longitude:        item.Longitude,
				Rating:           item.Rating,
				RatingCount:      item.RatingCount,
				Types:            typesStr,
				OpeningHours:     openingHoursStr,
				ThumbnailURL:     item.ThumbnailURL,
				CID:              item.CID,
				PlaceID:          item.PlaceID,
				Website:          item.Website,
				Notes:            item.Notes,
				CreatedBy:        &createdBy,
				ProcessedFromN8N: true,
				ProcessedAt:      &now,
			}

			// Load status for score calculation
			if defaultStatusID != nil && defaultStatus != nil {
				lead.LeadStatus = defaultStatus
			}
			lead.LeadScore = lead.CalculateLeadScore()

			if createErr := u.leadRepo.Create(ctx, lead); createErr != nil {
				fmt.Printf("Error creating new lead %s - %s: %v\n", lead.CompanyName, lead.FirstName, createErr)
				result.Errors++
				continue
			}

			reloaded, reloadErr := u.leadRepo.FindByID(ctx, lead.ID)
			if reloadErr != nil {
				result.Errors++
				continue
			}

			result.Created++
			result.Items = append(result.Items, mapper.ToLeadResponse(reloaded))
		}
	}

	return result, nil
}

func (u *leadUsecase) GetFormData(ctx context.Context) (*dto.LeadFormDataResponse, error) {
	// Fetch employees for assignment dropdown
	employees, err := u.employeeRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %w", err)
	}

	employeeOptions := make([]dto.LeadEmployeeOption, 0, len(employees))
	for _, emp := range employees {
		empID, err := uuid.Parse(emp.ID)
		if err != nil {
			continue
		}
		employeeOptions = append(employeeOptions, dto.LeadEmployeeOption{
			ID:           empID,
			EmployeeCode: emp.EmployeeCode,
			Name:         emp.Name,
		})
	}

	// Fetch lead sources
	sources, _, err := u.leadSourceRepo.List(ctx, repositories.ListParams{Limit: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lead sources: %w", err)
	}

	sourceOptions := make([]dto.LeadSourceOption, 0, len(sources))
	for _, s := range sources {
		sourceOptions = append(sourceOptions, dto.LeadSourceOption{
			ID:   s.ID,
			Name: s.Name,
			Code: s.Code,
		})
	}

	// Fetch lead statuses
	statuses, _, err := u.leadStatusRepo.List(ctx, repositories.ListParams{Limit: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lead statuses: %w", err)
	}

	statusOptions := make([]dto.LeadStatusOption, 0, len(statuses))
	for _, s := range statuses {
		if !s.IsActive {
			continue
		}
		statusOptions = append(statusOptions, dto.LeadStatusOption{
			ID:          s.ID,
			Name:        s.Name,
			Code:        s.Code,
			Color:       s.Color,
			IsDefault:   s.IsDefault,
			IsConverted: s.IsConverted,
		})
	}

	// Fetch pipeline stages for conversion dropdown
	stages, _, err := u.pipelineStageRepo.List(ctx, repositories.ListParams{Limit: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pipeline stages: %w", err)
	}

	stageOptions := make([]dto.LeadPipelineStageOption, 0, len(stages))
	for _, s := range stages {
		if !s.IsActive {
			continue
		}
		stageOptions = append(stageOptions, dto.LeadPipelineStageOption{
			ID:          s.ID,
			Name:        s.Name,
			Code:        s.Code,
			Order:       s.Order,
			Probability: s.Probability,
		})
	}

	// Fetch business types
	businessTypes, _, err := u.businessTypeRepo.List(ctx, &orgDto.ListBusinessTypesRequest{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch business types: %w", err)
	}

	businessTypeOptions := make([]dto.LeadBusinessTypeOption, 0, len(businessTypes))
	for _, bt := range businessTypes {
		businessTypeOptions = append(businessTypeOptions, dto.LeadBusinessTypeOption{
			ID:   bt.ID,
			Name: bt.Name,
		})
	}

	// Fetch areas
	areas, _, err := u.areaRepo.List(ctx, &orgDto.ListAreasRequest{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch areas: %w", err)
	}

	areaOptions := make([]dto.LeadAreaOption, 0, len(areas))
	for _, a := range areas {
		areaOptions = append(areaOptions, dto.LeadAreaOption{
			ID:       a.ID,
			Name:     a.Name,
			Province: a.Province,
		})
	}

	// Fetch payment terms
	paymentTermsList, _, err := u.paymentTermsRepo.List(ctx, coreRepos.ListParams{Limit: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment terms: %w", err)
	}

	paymentTermsOptions := make([]dto.LeadPaymentTermsOption, 0, len(paymentTermsList))
	for _, pt := range paymentTermsList {
		paymentTermsOptions = append(paymentTermsOptions, dto.LeadPaymentTermsOption{
			ID:   pt.ID,
			Name: pt.Name,
			Code: pt.Code,
			Days: pt.Days,
		})
	}

	return &dto.LeadFormDataResponse{
		Employees:      employeeOptions,
		LeadSources:    sourceOptions,
		LeadStatuses:   statusOptions,
		PipelineStages: stageOptions,
		BusinessTypes:  businessTypeOptions,
		Areas:          areaOptions,
		PaymentTerms:   paymentTermsOptions,
	}, nil
}

func (u *leadUsecase) GetAnalytics(ctx context.Context) (*repositories.LeadAnalytics, error) {
	return u.leadRepo.GetAnalytics(ctx)
}

// GetUnprocessed retrieves unprocessed leads for n8n automation (prevents duplicate processing)
func (u *leadUsecase) GetUnprocessed(ctx context.Context, limit int) ([]dto.LeadResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50 // Default and max limit
	}

	leads, err := u.leadRepo.FindUnprocessed(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unprocessed leads: %w", err)
	}

	responses := mapper.ToLeadResponseList(leads)
	return responses, nil
}

func inferLeadSourceID(item dto.UpsertLeadItem) string {
	placeIDLower := strings.ToLower(strings.TrimSpace(item.PlaceID))
	websiteLower := strings.ToLower(strings.TrimSpace(item.Website))
	notesLower := strings.ToLower(strings.TrimSpace(item.Notes))

	if strings.Contains(placeIDLower, "linkedin.com") || strings.Contains(websiteLower, "linkedin.com") || strings.Contains(notesLower, "linkedin") {
		return leadSourceLinkedInID
	}

	if strings.Contains(placeIDLower, "google.com/maps") || strings.Contains(notesLower, "google maps") || item.CID != "" || item.Latitude != nil || item.Longitude != nil {
		return leadSourceGoogleMapsID
	}

	return ""
}

func (u *leadUsecase) GetProductItems(ctx context.Context, leadID string) ([]dto.LeadProductItemResponse, error) {
	items, err := u.leadRepo.ListProductItems(ctx, leadID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.LeadProductItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, dto.LeadProductItemResponse{
			ID:                  item.ID,
			LeadID:              item.LeadID,
			ProductID:           item.ProductID,
			ProductName:         item.ProductName,
			ProductSKU:          item.ProductSKU,
			InterestLevel:       item.InterestLevel,
			Quantity:            item.Quantity,
			UnitPrice:           item.UnitPrice,
			Notes:               item.Notes,
			SourceVisitReportID: item.SourceVisitReportID,
			LastSurveyAnswers:   item.LastSurveyAnswers,
			IsDeleted:           item.DeletedAt.Valid,
			CreatedAt:           item.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
		})
	}
	return result, nil
}
