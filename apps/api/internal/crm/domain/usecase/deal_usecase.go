package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/mapper"
	customerModels "github.com/gilabs/gims/api/internal/customer/data/models"
	customerRepos "github.com/gilabs/gims/api/internal/customer/data/repositories"
	orgRepos "github.com/gilabs/gims/api/internal/organization/data/repositories"
	productRepos "github.com/gilabs/gims/api/internal/product/data/repositories"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	salesRepos "github.com/gilabs/gims/api/internal/sales/data/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DealUsecase defines the interface for deal business logic
type DealUsecase interface {
	Create(ctx context.Context, req dto.CreateDealRequest, createdBy string) (dto.DealResponse, error)
	GetByID(ctx context.Context, id string) (dto.DealResponse, error)
	List(ctx context.Context, params repositories.DealListParams) ([]dto.DealResponse, int64, error)
	ListByStage(ctx context.Context, params repositories.DealsByStageParams) ([]dto.DealResponse, int64, error)
	Update(ctx context.Context, id string, req dto.UpdateDealRequest) (dto.DealResponse, error)
	Delete(ctx context.Context, id string) error
	MoveStage(ctx context.Context, id string, req dto.MoveDealStageRequest, changedBy string) (dto.MoveDealStageResponse, error)
	GetHistory(ctx context.Context, dealID string) ([]dto.DealHistoryResponse, error)
	GetFormData(ctx context.Context) (*dto.DealFormDataResponse, error)
	GetPipelineSummary(ctx context.Context) (dto.DealPipelineSummaryResponse, error)
	GetForecast(ctx context.Context) (dto.DealForecastResponse, error)
	ConvertToQuotation(ctx context.Context, dealID string, req dto.ConvertToQuotationRequest, userID string) (dto.ConvertToQuotationResponse, error)
	StockCheck(ctx context.Context, dealID string) (dto.StockCheckResponse, error)
	SoftDeleteItem(ctx context.Context, dealID, itemID string) error
	RestoreItem(ctx context.Context, dealID, itemID string) error
	GetProductItems(ctx context.Context, dealID string) ([]dto.LeadProductItemResponse, error)
}

type dealUsecase struct {
	dealRepo           repositories.DealRepository
	stageRepo          repositories.PipelineStageRepository
	customerRepo       customerRepos.CustomerRepository
	contactRepo        repositories.ContactRepository
	employeeRepo       orgRepos.EmployeeRepository
	productRepo        productRepos.ProductRepository
	leadRepo           repositories.LeadRepository
	activityRepo       repositories.ActivityRepository
	salesQuotationRepo salesRepos.SalesQuotationRepository
	db                 *gorm.DB
}

// NewDealUsecase creates a new deal usecase
func NewDealUsecase(
	dealRepo repositories.DealRepository,
	stageRepo repositories.PipelineStageRepository,
	customerRepo customerRepos.CustomerRepository,
	contactRepo repositories.ContactRepository,
	employeeRepo orgRepos.EmployeeRepository,
	productRepo productRepos.ProductRepository,
	leadRepo repositories.LeadRepository,
	activityRepo repositories.ActivityRepository,
	salesQuotationRepo salesRepos.SalesQuotationRepository,
	db *gorm.DB,
) DealUsecase {
	return &dealUsecase{
		dealRepo:           dealRepo,
		stageRepo:          stageRepo,
		customerRepo:       customerRepo,
		contactRepo:        contactRepo,
		employeeRepo:       employeeRepo,
		productRepo:        productRepo,
		leadRepo:           leadRepo,
		activityRepo:       activityRepo,
		salesQuotationRepo: salesQuotationRepo,
		db:                 db,
	}
}

func (u *dealUsecase) Create(ctx context.Context, req dto.CreateDealRequest, createdBy string) (dto.DealResponse, error) {
	// Validate pipeline stage
	stage, err := u.stageRepo.FindByID(ctx, req.PipelineStageID)
	if err != nil {
		return dto.DealResponse{}, errors.New("pipeline stage not found")
	}

	// Validate customer if provided
	if req.CustomerID != nil && *req.CustomerID != "" {
		_, err := u.customerRepo.FindByID(ctx, *req.CustomerID)
		if err != nil {
			return dto.DealResponse{}, errors.New("customer not found")
		}
	}

	// Validate contact if provided
	if req.ContactID != nil && *req.ContactID != "" {
		_, err := u.contactRepo.FindByID(ctx, *req.ContactID)
		if err != nil {
			return dto.DealResponse{}, errors.New("contact not found")
		}
	}

	// Validate assigned employee if provided
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		_, err := u.employeeRepo.FindByID(ctx, *req.AssignedTo)
		if err != nil {
			return dto.DealResponse{}, errors.New("assigned employee not found")
		}
	}

	// Validate lead if provided
	if req.LeadID != nil && *req.LeadID != "" {
		_, err := u.leadRepo.FindByID(ctx, *req.LeadID)
		if err != nil {
			return dto.DealResponse{}, errors.New("lead not found")
		}
	}

	// Parse expected close date
	var expectedCloseDate *time.Time
	if req.ExpectedCloseDate != nil && *req.ExpectedCloseDate != "" {
		t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate)
		if err != nil {
			return dto.DealResponse{}, errors.New("invalid expected_close_date format, use YYYY-MM-DD")
		}
		expectedCloseDate = &t
	}

	// Build product items with snapshot data
	items := make([]models.DealProductItem, 0, len(req.Items))
	for _, itemDTO := range req.Items {
		item := models.DealProductItem{
			ID:              uuid.New().String(),
			ProductID:       itemDTO.ProductID,
			ProductName:     itemDTO.ProductName,
			ProductSKU:      itemDTO.ProductSKU,
			UnitPrice:       itemDTO.UnitPrice,
			Quantity:        itemDTO.Quantity,
			DiscountPercent: itemDTO.DiscountPercent,
			DiscountAmount:  itemDTO.DiscountAmount,
			Notes:           itemDTO.Notes,
		}

		// Snapshot product data if product ID is provided
		if itemDTO.ProductID != nil && *itemDTO.ProductID != "" {
			product, err := u.productRepo.FindByID(ctx, *itemDTO.ProductID)
			if err != nil {
				return dto.DealResponse{}, fmt.Errorf("product not found: %s", *itemDTO.ProductID)
			}
			if item.ProductName == "" {
				item.ProductName = product.Name
			}
			if item.ProductSKU == "" {
				item.ProductSKU = product.Sku
			}
			if item.UnitPrice == 0 {
				item.UnitPrice = product.SellingPrice
			}
		}

		// Calculate subtotal
		item.Subtotal = item.CalculateSubtotal()
		items = append(items, item)
	}

	// Calculate value from items if items exist and no manual value
	dealValue := req.Value
	if len(items) > 0 && dealValue == 0 {
		for _, item := range items {
			dealValue += item.Subtotal
		}
	}

	// Extract tenant_id from context
	tenantID, _ := ctx.Value("tenant_id").(string)

	deal := &models.Deal{
		ID:                uuid.New().String(),
		TenantID:          tenantID,
		Title:             req.Title,
		Description:       req.Description,
		Status:            models.DealStatusOpen,
		PipelineStageID:    req.PipelineStageID,
		Value:             dealValue,
		Probability:       stage.Probability,
		ExpectedCloseDate: expectedCloseDate,
		CustomerID:        req.CustomerID,
		ContactID:         req.ContactID,
		AssignedTo:        req.AssignedTo,
		LeadID:            req.LeadID,
		BudgetConfirmed:   req.BudgetConfirmed,
		BudgetAmount:      req.BudgetAmount,
		AuthConfirmed:     req.AuthConfirmed,
		AuthPerson:        req.AuthPerson,
		NeedConfirmed:     req.NeedConfirmed,
		NeedDescription:   req.NeedDescription,
		TimeConfirmed:     req.TimeConfirmed,
		Notes:             req.Notes,
		CreatedBy:         &createdBy,
		Items:             items,
	}

	for i := range deal.Items {
		deal.Items[i].TenantID = tenantID
	}

	if err := u.dealRepo.Create(ctx, deal); err != nil {
		return dto.DealResponse{}, fmt.Errorf("failed to create deal: %w", err)
	}

	// Reload with preloaded relations
	created, err := u.dealRepo.FindByID(ctx, deal.ID)
	if err != nil {
		return dto.DealResponse{}, err
	}

	return mapper.ToDealResponse(created), nil
}

func (u *dealUsecase) GetByID(ctx context.Context, id string) (dto.DealResponse, error) {
	if !security.CheckRecordScopeAccess(u.db, ctx, &models.Deal{}, id, security.MixedOwnershipScopeQueryOptions("assigned_to")) {
		return dto.DealResponse{}, errors.New("deal not found")
	}

	deal, err := u.dealRepo.FindByID(ctx, id)
	if err != nil {
		return dto.DealResponse{}, errors.New("deal not found")
	}
	return mapper.ToDealResponse(deal), nil
}

func (u *dealUsecase) List(ctx context.Context, params repositories.DealListParams) ([]dto.DealResponse, int64, error) {
	deals, total, err := u.dealRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToDealResponseList(deals), total, nil
}

func (u *dealUsecase) ListByStage(ctx context.Context, params repositories.DealsByStageParams) ([]dto.DealResponse, int64, error) {
	deals, total, err := u.dealRepo.ListByStage(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	// Ensure lead is populated for each deal when repository did not preload it.
	// This can happen in some DB/ORM configs; fetching missing lead ensures
	// the mapper can create snapshot customer/contact from lead data.
	for i := range deals {
		if deals[i].Lead == nil && deals[i].LeadID != nil && *deals[i].LeadID != "" {
			if lead, lerr := u.leadRepo.FindByID(ctx, *deals[i].LeadID); lerr == nil && lead != nil {
				deals[i].Lead = lead
			}
		}
	}
	return mapper.ToDealResponseList(deals), total, nil
}

func (u *dealUsecase) Update(ctx context.Context, id string, req dto.UpdateDealRequest) (dto.DealResponse, error) {
	deal, err := u.dealRepo.FindByID(ctx, id)
	if err != nil {
		return dto.DealResponse{}, errors.New("deal not found")
	}

	// Prevent updates on closed deals
	if deal.Status != models.DealStatusOpen {
		return dto.DealResponse{}, errors.New("deal already closed")
	}

	// Validate stage if changing
	if req.PipelineStageID != nil && *req.PipelineStageID != "" {
		stage, err := u.stageRepo.FindByID(ctx, *req.PipelineStageID)
		if err != nil {
			return dto.DealResponse{}, errors.New("pipeline stage not found")
		}
		deal.PipelineStageID = *req.PipelineStageID
		deal.Probability = stage.Probability
	}

	// Validate customer if changing
	if req.CustomerID != nil && *req.CustomerID != "" {
		_, err := u.customerRepo.FindByID(ctx, *req.CustomerID)
		if err != nil {
			return dto.DealResponse{}, errors.New("customer not found")
		}
		deal.CustomerID = req.CustomerID
	}

	// Validate contact if changing
	if req.ContactID != nil && *req.ContactID != "" {
		_, err := u.contactRepo.FindByID(ctx, *req.ContactID)
		if err != nil {
			return dto.DealResponse{}, errors.New("contact not found")
		}
		deal.ContactID = req.ContactID
	}

	// Validate assigned employee if changing
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		_, err := u.employeeRepo.FindByID(ctx, *req.AssignedTo)
		if err != nil {
			return dto.DealResponse{}, errors.New("assigned employee not found")
		}
		deal.AssignedTo = req.AssignedTo
	}
	if req.LeadID != nil {
		if *req.LeadID != "" {
			_, err := u.leadRepo.FindByID(ctx, *req.LeadID)
			if err != nil {
				return dto.DealResponse{}, errors.New("lead not found")
			}
		}
		deal.LeadID = req.LeadID
	}

	// Apply partial updates
	if req.Title != nil {
		deal.Title = *req.Title
	}
	if req.Description != nil {
		deal.Description = *req.Description
	}
	if req.Value != nil {
		deal.Value = *req.Value
	}
	if req.ExpectedCloseDate != nil && *req.ExpectedCloseDate != "" {
		t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate)
		if err != nil {
			return dto.DealResponse{}, errors.New("invalid expected_close_date format, use YYYY-MM-DD")
		}
		deal.ExpectedCloseDate = &t
	}

	// BANT updates
	if req.BudgetConfirmed != nil {
		deal.BudgetConfirmed = *req.BudgetConfirmed
	}
	if req.BudgetAmount != nil {
		deal.BudgetAmount = *req.BudgetAmount
	}
	if req.AuthConfirmed != nil {
		deal.AuthConfirmed = *req.AuthConfirmed
	}
	if req.AuthPerson != nil {
		deal.AuthPerson = *req.AuthPerson
	}
	if req.NeedConfirmed != nil {
		deal.NeedConfirmed = *req.NeedConfirmed
	}
	if req.NeedDescription != nil {
		deal.NeedDescription = *req.NeedDescription
	}
	if req.TimeConfirmed != nil {
		deal.TimeConfirmed = *req.TimeConfirmed
	}
	if req.Notes != nil {
		deal.Notes = *req.Notes
	}

	// Handle product items replacement
	if req.Items != nil {
		// Delete existing items
		if err := u.dealRepo.DeleteItemsByDealID(ctx, id); err != nil {
			return dto.DealResponse{}, fmt.Errorf("failed to clear deal items: %w", err)
		}

		// Create new items
		items := make([]models.DealProductItem, 0, len(*req.Items))
		for _, itemDTO := range *req.Items {
			item := models.DealProductItem{
				ID:              uuid.New().String(),
				DealID:          id,
				ProductID:       itemDTO.ProductID,
				ProductName:     itemDTO.ProductName,
				ProductSKU:      itemDTO.ProductSKU,
				UnitPrice:       itemDTO.UnitPrice,
				Quantity:        itemDTO.Quantity,
				DiscountPercent: itemDTO.DiscountPercent,
				DiscountAmount:  itemDTO.DiscountAmount,
				Notes:           itemDTO.Notes,
			}

			// Snapshot product data if product ID is provided.
			// During update, if the product is no longer found (e.g. archived),
			// use the data already supplied in the item DTO as-is.
			if itemDTO.ProductID != nil && *itemDTO.ProductID != "" {
				product, err := u.productRepo.FindByID(ctx, *itemDTO.ProductID)
				if err == nil {
					if item.ProductName == "" {
						item.ProductName = product.Name
					}
					if item.ProductSKU == "" {
						item.ProductSKU = product.Sku
					}
					if item.UnitPrice == 0 {
						item.UnitPrice = product.SellingPrice
					}
				}
			}

			item.Subtotal = item.CalculateSubtotal()
			items = append(items, item)
		}

		if len(items) > 0 {
			if err := u.dealRepo.CreateItems(ctx, items); err != nil {
				return dto.DealResponse{}, fmt.Errorf("failed to create deal items: %w", err)
			}

			// Recalculate deal value from items if value not explicitly set
			if req.Value == nil {
				totalValue := 0.0
				for _, item := range items {
					totalValue += item.Subtotal
				}
				deal.Value = totalValue
			}
		}
	}

	// Nil out preloaded association pointers so GORM cannot use stale BelongsTo
	// data to override FK columns during Save.
	deal.PipelineStage = nil
	deal.Customer = nil
	deal.Contact = nil
	deal.AssignedEmployee = nil
	deal.Lead = nil
	deal.Items = nil

	if err := u.dealRepo.Update(ctx, deal); err != nil {
		return dto.DealResponse{}, fmt.Errorf("failed to update deal: %w", err)
	}

	// Reload with preloaded relations
	updated, err := u.dealRepo.FindByID(ctx, id)
	if err != nil {
		return dto.DealResponse{}, err
	}

	return mapper.ToDealResponse(updated), nil
}

func (u *dealUsecase) Delete(ctx context.Context, id string) error {
	_, err := u.dealRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("deal not found")
	}
	return u.dealRepo.Delete(ctx, id)
}

func (u *dealUsecase) MoveStage(ctx context.Context, id string, req dto.MoveDealStageRequest, changedBy string) (dto.MoveDealStageResponse, error) {
	deal, err := u.dealRepo.FindByID(ctx, id)
	if err != nil {
		return dto.MoveDealStageResponse{}, errors.New("deal not found")
	}

	// Block stage movement for deals that have already been converted to a quotation
	if deal.ConvertedToQuotationID != nil && *deal.ConvertedToQuotationID != "" {
		return dto.MoveDealStageResponse{}, errors.New("deal already converted")
	}

	// Validate target stage first, so we know if we're re-opening or staying closed
	toStage, err := u.stageRepo.FindByID(ctx, req.ToStageID)
	if err != nil {
		return dto.MoveDealStageResponse{}, errors.New("invalid pipeline stage")
	}

	// Require close reason for won/lost stages
	if (toStage.IsWon || toStage.IsLost) && req.CloseReason == "" {
		if toStage.IsLost {
			return dto.MoveDealStageResponse{}, errors.New("close reason required for lost deals")
		}
	}

	// Calculate days in previous stage
	daysInPrevStage := 0
	lastHistory, err := u.dealRepo.GetLastHistoryByDealID(ctx, id)
	if err == nil && lastHistory != nil {
		daysInPrevStage = int(math.Ceil(time.Since(lastHistory.ChangedAt).Hours() / 24))
	} else {
		// First stage transition, calculate from deal creation
		daysInPrevStage = int(math.Ceil(time.Since(deal.CreatedAt).Hours() / 24))
	}

	// Record previous stage info
	fromStageID := deal.PipelineStageID
	fromStageName := ""
	fromProbability := deal.Probability
	if deal.PipelineStage != nil {
		fromStageName = deal.PipelineStage.Name
	}

	// Resolve user ID to employee ID to satisfy the FK constraint on crm_deal_history.
	// The JWT carries a user_id, but the FK references the employees table.
	var employeeID *string
	if changedBy != "" {
		emp, empErr := u.employeeRepo.FindByUserID(ctx, changedBy)
		if empErr == nil && emp != nil {
			empID := emp.ID
			employeeID = &empID
		}
		// If no matching employee exists (e.g. system/admin account), ChangedBy stays nil.
	}

	// Create history record
	history := &models.DealHistory{
		ID:              uuid.New().String(),
			TenantID:        deal.TenantID,
		DealID:          id,
		FromStageID:     &fromStageID,
		FromStageName:   fromStageName,
		ToStageID:       req.ToStageID,
		ToStageName:     toStage.Name,
		FromProbability: fromProbability,
		ToProbability:   toStage.Probability,
		DaysInPrevStage: daysInPrevStage,
		ChangedBy:       employeeID,
		ChangedAt:       apptime.Now(),
		Reason:          req.Reason,
		Notes:           req.Notes,
	}

	if err := u.dealRepo.CreateHistory(ctx, history); err != nil {
		return dto.MoveDealStageResponse{}, fmt.Errorf("failed to create deal history: %w", err)
	}

	// Update deal fields
	deal.PipelineStageID = req.ToStageID
	deal.Probability = toStage.Probability

	if toStage.IsWon {
		deal.Status = models.DealStatusWon
		now := apptime.Now()
		deal.ActualCloseDate = &now
		if req.CloseReason != "" {
			deal.CloseReason = req.CloseReason
		}
	} else if toStage.IsLost {
		deal.Status = models.DealStatusLost
		now := apptime.Now()
		deal.ActualCloseDate = &now
		deal.CloseReason = req.CloseReason
	} else {
		// Moving to an open (non-closing) stage - re-open the deal if it was previously closed
		deal.Status = models.DealStatusOpen
		deal.ActualCloseDate = nil
		deal.CloseReason = ""
	}

	// Capture lead data before nil-ing preloaded associations
	var leadData *models.Lead
	if deal.Lead != nil {
		l := *deal.Lead
		leadData = &l
	}

	// Nil out preloaded associations before Save to prevent GORM BelongsTo FK override
	deal.PipelineStage = nil
	deal.Customer = nil
	deal.Contact = nil
	deal.AssignedEmployee = nil
	deal.Lead = nil
	deal.Items = nil

	if err := u.dealRepo.Update(ctx, deal); err != nil {
		return dto.MoveDealStageResponse{}, fmt.Errorf("failed to update deal stage: %w", err)
	}

	// Auto-create customer + contact when deal is won and no customer assigned yet
	if toStage.IsWon && deal.CustomerID == nil {
		if err := u.autoCreateCustomerFromDeal(ctx, deal, leadData, changedBy); err != nil {
			log.Printf("Warning: auto-create customer failed for deal %s: %v", id, err)
		}
	}

	// Reload with preloaded relations
	updated, err := u.dealRepo.FindByID(ctx, id)
	if err != nil {
		return dto.MoveDealStageResponse{}, err
	}

	// Log activity for stage movement (best-effort, non-blocking)
	activityEmployeeID := changedBy
	if activityEmployeeID == "" {
		activityEmployeeID = "system"
	}
	stageActivity := &models.Activity{
		Type:           "deal_stage_change",
		ActivityTypeID: strPtr(activityTypeFollowUpID),
		DealID:         &updated.ID,
		LeadID:         updated.LeadID,
		EmployeeID:     activityEmployeeID,
		Description:    fmt.Sprintf("Pipeline stage moved from %s to %s", fromStageName, toStage.Name),
		Timestamp:      apptime.Now(),
	}
	_ = u.activityRepo.Create(ctx, stageActivity)

	response := dto.MoveDealStageResponse{
		Deal: mapper.ToDealResponse(updated),
	}

	// Auto-convert to quotation if requested and stage is won
	if req.ConvertToQuotation && toStage.IsWon {
		convReq := dto.ConvertToQuotationRequest{}
		convResult, convErr := u.ConvertToQuotation(ctx, id, convReq, changedBy)
		if convErr != nil {
			// Conversion failed but stage move succeeded — return the deal with error info
			return response, fmt.Errorf("stage moved successfully but conversion failed: %w", convErr)
		}
		response.Conversion = &convResult

		// Reload deal again to reflect conversion fields
		finalDeal, reloadErr := u.dealRepo.FindByID(ctx, id)
		if reloadErr == nil {
			response.Deal = mapper.ToDealResponse(finalDeal)
		}
	}

	return response, nil
}

func (u *dealUsecase) GetHistory(ctx context.Context, dealID string) ([]dto.DealHistoryResponse, error) {
	// Verify deal exists
	_, err := u.dealRepo.FindByID(ctx, dealID)
	if err != nil {
		return nil, errors.New("deal not found")
	}

	history, err := u.dealRepo.GetHistory(ctx, dealID)
	if err != nil {
		return nil, err
	}

	return mapper.ToDealHistoryResponseList(history), nil
}

func (u *dealUsecase) GetFormData(ctx context.Context) (*dto.DealFormDataResponse, error) {
	// Employees
	employees, err := u.employeeRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %w", err)
	}
	employeeOptions := make([]dto.DealEmployeeOption, 0, len(employees))
	for _, emp := range employees {
		empID, err := uuid.Parse(emp.ID)
		if err != nil {
			continue
		}
		employeeOptions = append(employeeOptions, dto.DealEmployeeOption{
			ID:           empID,
			EmployeeCode: emp.EmployeeCode,
			Name:         emp.Name,
		})
	}

	// Customers
	customers, _, err := u.customerRepo.List(ctx, customerRepos.CustomerListParams{
		ListParams: customerRepos.ListParams{Limit: 500},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customers: %w", err)
	}
	customerOptions := make([]dto.DealCustomerOption, 0, len(customers))
	for _, c := range customers {
		customerOptions = append(customerOptions, dto.DealCustomerOption{
			ID:   c.ID,
			Code: c.Code,
			Name: c.Name,
		})
	}

	// Contacts
	contacts, _, err := u.contactRepo.List(ctx, repositories.ContactListParams{ListParams: repositories.ListParams{Limit: 500}})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contacts: %w", err)
	}
	contactOptions := make([]dto.DealContactOption, 0, len(contacts))
	for _, c := range contacts {
		contactOptions = append(contactOptions, dto.DealContactOption{
			ID:         c.ID,
			Name:       c.Name,
			Phone:      c.Phone,
			Email:      c.Email,
			CustomerID: c.CustomerID,
		})
	}

	// Pipeline stages
	stages, _, err := u.stageRepo.List(ctx, repositories.ListParams{Limit: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pipeline stages: %w", err)
	}
	stageOptions := make([]dto.DealPipelineStageOption, 0, len(stages))
	for _, s := range stages {
		if !s.IsActive {
			continue
		}
		stageOptions = append(stageOptions, dto.DealPipelineStageOption{
			ID:          s.ID,
			Name:        s.Name,
			Code:        s.Code,
			Color:       s.Color,
			Order:       s.Order,
			Probability: s.Probability,
			IsWon:       s.IsWon,
			IsLost:      s.IsLost,
			IsActive:    s.IsActive,
		})
	}

	// Products
	products, _, err := u.productRepo.List(ctx, productRepos.ProductListParams{
		ListParams: productRepos.ListParams{Limit: 500},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}
	productOptions := make([]dto.DealProductOption, 0, len(products))
	for _, p := range products {
		productOptions = append(productOptions, dto.DealProductOption{
			ID:           p.ID,
			Code:         p.Code,
			Name:         p.Name,
			SKU:          p.Sku,
			SellingPrice: p.SellingPrice,
		})
	}

	// Leads (qualified and not yet converted, for deal creation from lead)
	leads, _, err := u.leadRepo.List(ctx, repositories.LeadListParams{Limit: 500})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch leads: %w", err)
	}
	leadOptions := make([]dto.DealLeadOption, 0, len(leads))
	for _, l := range leads {
		if l.IsConverted() {
			continue
		}
		if l.LeadStatus == nil || !strings.EqualFold(l.LeadStatus.Code, "QUALIFIED") {
			continue
		}
		leadOptions = append(leadOptions, dto.DealLeadOption{
			ID:                       l.ID,
			Code:                     l.Code,
			FirstName:                l.FirstName,
			LastName:                 l.LastName,
			CompanyName:              l.CompanyName,
			IsConverted:              l.IsConverted(),
			IsQualifiedForConversion: true,
		})
	}

	return &dto.DealFormDataResponse{
		Employees:      employeeOptions,
		Customers:      customerOptions,
		Contacts:       contactOptions,
		PipelineStages: stageOptions,
		Products:       productOptions,
		Leads:          leadOptions,
	}, nil
}

func (u *dealUsecase) GetPipelineSummary(ctx context.Context) (dto.DealPipelineSummaryResponse, error) {
	data, err := u.dealRepo.GetPipelineSummary(ctx)
	if err != nil {
		return dto.DealPipelineSummaryResponse{}, err
	}
	return mapper.ToPipelineSummaryResponse(data), nil
}

func (u *dealUsecase) GetForecast(ctx context.Context) (dto.DealForecastResponse, error) {
	data, err := u.dealRepo.GetForecast(ctx)
	if err != nil {
		return dto.DealForecastResponse{}, err
	}
	return mapper.ToForecastResponse(data), nil
}

// ConvertToQuotation converts a won deal into a Sales Quotation
func (u *dealUsecase) ConvertToQuotation(ctx context.Context, dealID string, req dto.ConvertToQuotationRequest, userID string) (dto.ConvertToQuotationResponse, error) {
	deal, err := u.dealRepo.FindByID(ctx, dealID)
	if err != nil {
		return dto.ConvertToQuotationResponse{}, errors.New("deal not found")
	}

	// Validate deal status must be "won"
	if deal.Status != models.DealStatusWon {
		return dto.ConvertToQuotationResponse{}, errors.New("deal not won")
	}

	// Validate deal has not been converted already
	if deal.ConvertedToQuotationID != nil && *deal.ConvertedToQuotationID != "" {
		return dto.ConvertToQuotationResponse{}, errors.New("deal already converted")
	}

	// Resolve active source items for conversion:
	// - lead-linked deals: authoritative source is crm_lead_product_items
	// - standalone deals: fallback to crm_deal_product_items snapshot
	var leadItems []models.LeadProductItem
	if deal.LeadID != nil && *deal.LeadID != "" {
		leadItems, err = u.leadRepo.ListProductItems(ctx, *deal.LeadID)
		if err != nil {
			return dto.ConvertToQuotationResponse{}, fmt.Errorf("failed to fetch lead product items: %w", err)
		}

		hasActiveLeadItems := false
		for _, item := range leadItems {
			if !item.DeletedAt.Valid {
				hasActiveLeadItems = true
				break
			}
		}
		if !hasActiveLeadItems {
			return dto.ConvertToQuotationResponse{}, errors.New("deal has no items")
		}
	} else {
		// Validate standalone deal has product items
		hasActiveDealItems := false
		for _, item := range deal.Items {
			if !item.DeletedAt.Valid {
				hasActiveDealItems = true
				break
			}
		}
		if !hasActiveDealItems {
			return dto.ConvertToQuotationResponse{}, errors.New("deal has no items")
		}
	}

	// Auto-create customer from lead when no customer is linked yet or linked customer doesn't exist
	customerExists := true
	if deal.CustomerID == nil || *deal.CustomerID == "" {
		customerExists = false
	} else {
		_, findErr := u.customerRepo.FindByID(ctx, *deal.CustomerID)
		if findErr != nil {
			customerExists = false
		}
	}

	if !customerExists {
		if deal.Lead == nil {
			return dto.ConvertToQuotationResponse{}, errors.New("deal customer required")
		}
		customerCode, codeErr := u.customerRepo.GetNextCode(ctx)
		if codeErr != nil {
			return dto.ConvertToQuotationResponse{}, fmt.Errorf("failed to generate customer code: %w", codeErr)
		}
		customerName := deal.Lead.CompanyName
		if customerName == "" {
			customerName = deal.Lead.FirstName
			if deal.Lead.LastName != "" {
				customerName += " " + deal.Lead.LastName
			}
		}
		newCustomer := &customerModels.Customer{
			ID:                    uuid.New().String(),
			Code:                  customerCode,
			Name:                  customerName,
			Address:               deal.Lead.Address,
			Email:                 deal.Lead.Email,
			Website:               deal.Lead.Website,
			NPWP:                  deal.Lead.NPWP,
			ContactPerson:         deal.Lead.FirstName + " " + deal.Lead.LastName,
			Latitude:              deal.Lead.Latitude,
			Longitude:             deal.Lead.Longitude,
			ProvinceID:            deal.Lead.ProvinceID,
			DefaultBusinessTypeID: deal.Lead.BusinessTypeID,
			DefaultAreaID:         deal.Lead.AreaID,
			DefaultSalesRepID:     deal.AssignedTo,
			DefaultPaymentTermsID: deal.Lead.PaymentTermsID,
			CreatedBy:             &userID,
		}
		if err := u.customerRepo.Create(ctx, newCustomer); err != nil {
			return dto.ConvertToQuotationResponse{}, fmt.Errorf("failed to create customer from lead: %w", err)
		}

		phoneToUse := deal.Lead.Phone
		nameToUse := strings.TrimSpace(deal.Lead.FirstName + " " + deal.Lead.LastName)
		emailToUse := deal.Lead.Email
		labelToUse := deal.Lead.JobTitle
		if deal.Contact != nil {
			if deal.Contact.Phone != "" {
				phoneToUse = deal.Contact.Phone
			}
			if deal.Contact.Name != "" {
				nameToUse = deal.Contact.Name
			}
			if deal.Contact.Email != "" {
				emailToUse = deal.Contact.Email
			}
			if deal.Contact.Position != "" {
				labelToUse = deal.Contact.Position
			}
		}

		if nameToUse != "" {
			newContact := &models.Contact{
				ID:         uuid.New().String(),
				CustomerID: newCustomer.ID,
				Name:       nameToUse,
				Phone:      phoneToUse,
				Email:      emailToUse,
				Position:   labelToUse,
				IsActive:   true,
			}
			if userID != "" {
				newContact.CreatedBy = &userID
			}
			_ = u.contactRepo.Create(ctx, newContact)
		}
		deal.CustomerID = &newCustomer.ID
		// Best-effort: link the lead to the new customer
		if deal.Lead != nil {
			deal.Lead.CustomerID = &newCustomer.ID
			deal.Lead.Customer = nil
			_ = u.leadRepo.Update(ctx, deal.Lead)
		}
	}

	// Snapshot customer data
	customer, err := u.customerRepo.FindByID(ctx, *deal.CustomerID)
	if err != nil {
		return dto.ConvertToQuotationResponse{}, fmt.Errorf("customer not found: %w", err)
	}

	// Generate quotation code
	now := apptime.Now()
	prefix := "QUO"
	codePrefix := fmt.Sprintf("%s-%s", prefix, now.Format("200601"))
	quotationCode, err := u.salesQuotationRepo.GetNextQuotationNumber(ctx, codePrefix)
	if err != nil {
		return dto.ConvertToQuotationResponse{}, fmt.Errorf("failed to generate quotation code: %w", err)
	}

	// Build quotation items from resolved active source items
	var subtotal float64
	quotationItems := make([]salesModels.SalesQuotationItem, 0)
	if deal.LeadID != nil && *deal.LeadID != "" {
		quotationItems = make([]salesModels.SalesQuotationItem, 0, len(leadItems))
		for _, leadItem := range leadItems {
			if leadItem.DeletedAt.Valid {
				continue // skip soft-deleted items
			}
			item := salesModels.SalesQuotationItem{
				ID:       uuid.New().String(),
				Quantity: float64(leadItem.Quantity),
				Price:    leadItem.UnitPrice,
				Discount: 0,
			}

			if leadItem.ProductID != nil && *leadItem.ProductID != "" {
				item.ProductID = *leadItem.ProductID
			}

			item.CalculateSubtotal()
			subtotal += item.Subtotal
			quotationItems = append(quotationItems, item)
		}
	} else {
		quotationItems = make([]salesModels.SalesQuotationItem, 0, len(deal.Items))
		for _, dealItem := range deal.Items {
			if dealItem.DeletedAt.Valid {
				continue // skip soft-deleted items
			}
			item := salesModels.SalesQuotationItem{
				ID:       uuid.New().String(),
				Quantity: float64(dealItem.Quantity),
				Price:    dealItem.UnitPrice,
				Discount: dealItem.DiscountAmount,
			}

			if dealItem.ProductID != nil && *dealItem.ProductID != "" {
				item.ProductID = *dealItem.ProductID
			}

			item.CalculateSubtotal()
			subtotal += item.Subtotal
			quotationItems = append(quotationItems, item)
		}
	}

	// Calculate tax (11% PPN)
	const defaultTaxRate = 11.0
	taxAmount := subtotal * (defaultTaxRate / 100)
	totalAmount := subtotal + taxAmount

	// Build quotation
	dealIDRef := dealID
	quotation := &salesModels.SalesQuotation{
		ID:            uuid.New().String(),
		Code:          quotationCode,
		QuotationDate: now,
		CustomerID:    deal.CustomerID,
		CustomerName:  customer.Name,
		SalesRepID:    deal.AssignedTo,
		SourceDealID:  &dealIDRef,
		Subtotal:      subtotal,
		TaxRate:       defaultTaxRate,
		TaxAmount:     taxAmount,
		TotalAmount:   totalAmount,
		Status:        salesModels.SalesQuotationStatusDraft,
		CreatedBy:     &userID,
		Items:         quotationItems,
	}

	// Apply optional overrides
	if req.PaymentTermsID != nil && *req.PaymentTermsID != "" {
		quotation.PaymentTermsID = req.PaymentTermsID
	}
	if req.BusinessUnitID != nil && *req.BusinessUnitID != "" {
		quotation.BusinessUnitID = req.BusinessUnitID
	}
	if req.BusinessTypeID != nil && *req.BusinessTypeID != "" {
		quotation.BusinessTypeID = req.BusinessTypeID
	}
	if req.Notes != "" {
		quotation.Notes = req.Notes
	}

	// Snapshot customer contact info
	quotation.CustomerContact = customer.ContactPerson
	quotation.CustomerEmail = customer.Email

	// Create the quotation
	if err := u.salesQuotationRepo.Create(ctx, quotation); err != nil {
		return dto.ConvertToQuotationResponse{}, fmt.Errorf("failed to create quotation: %w", err)
	}

	// Update deal with conversion reference
	quotationID := quotation.ID
	deal.ConvertedToQuotationID = &quotationID
	deal.ConvertedAt = &now

	if err := u.dealRepo.Update(ctx, deal); err != nil {
		return dto.ConvertToQuotationResponse{}, fmt.Errorf("failed to update deal conversion: %w", err)
	}

	return dto.ConvertToQuotationResponse{
		DealID:        deal.ID,
		QuotationID:   quotation.ID,
		QuotationCode: quotation.Code,
	}, nil
}

// stockRow holds the aggregated stock data for a product
type stockRow struct {
	ProductID      string  `gorm:"column:product_id"`
	AvailableStock float64 `gorm:"column:available_stock"`
	ReservedStock  float64 `gorm:"column:reserved_stock"`
}

// StockCheck queries ERP inventory for stock availability per deal product item
func (u *dealUsecase) StockCheck(ctx context.Context, dealID string) (dto.StockCheckResponse, error) {
	deal, err := u.dealRepo.FindByID(ctx, dealID)
	if err != nil {
		return dto.StockCheckResponse{}, errors.New("deal not found")
	}

	if len(deal.Items) == 0 {
		return dto.StockCheckResponse{
			DealID:        deal.ID,
			Items:         []dto.StockCheckItemResponse{},
			AllSufficient: true,
		}, nil
	}

	// Collect product IDs that have a valid reference
	productIDs := make([]string, 0, len(deal.Items))
	for _, item := range deal.Items {
		if item.ProductID != nil && *item.ProductID != "" {
			productIDs = append(productIDs, *item.ProductID)
		}
	}

	// Query aggregated stock from inventory_batches
	stockMap := make(map[string]stockRow)
	if len(productIDs) > 0 {
		var rows []stockRow
		err := u.db.WithContext(ctx).
			Table("inventory_batches").
			Select(`
				product_id,
				COALESCE(SUM(current_quantity - reserved_quantity), 0) as available_stock,
				COALESCE(SUM(reserved_quantity), 0) as reserved_stock
			`).
			Where("product_id IN ? AND is_active = ? AND deleted_at IS NULL", productIDs, true).
			Group("product_id").
			Scan(&rows).Error
		if err != nil {
			return dto.StockCheckResponse{}, errors.New("stock check failed")
		}
		for _, r := range rows {
			stockMap[r.ProductID] = r
		}
	}

	// Build response items
	allSufficient := true
	items := make([]dto.StockCheckItemResponse, 0, len(deal.Items))
	for _, dealItem := range deal.Items {
		respItem := dto.StockCheckItemResponse{
			ProductName:       dealItem.ProductName,
			RequestedQuantity: dealItem.Quantity,
		}

		if dealItem.ProductID != nil && *dealItem.ProductID != "" {
			respItem.ProductID = *dealItem.ProductID
			if stock, ok := stockMap[*dealItem.ProductID]; ok {
				respItem.AvailableStock = stock.AvailableStock
				respItem.ReservedStock = stock.ReservedStock
			}
		}

		respItem.IsSufficient = respItem.AvailableStock >= float64(respItem.RequestedQuantity)
		if !respItem.IsSufficient {
			allSufficient = false
		}

		items = append(items, respItem)
	}

	return dto.StockCheckResponse{
		DealID:        deal.ID,
		Items:         items,
		AllSufficient: allSufficient,
	}, nil
}

// autoCreateCustomerFromDeal creates a customer and contact from the deal's associated lead
// when a deal moves to a Won stage and has no customer assigned yet.
func (u *dealUsecase) autoCreateCustomerFromDeal(ctx context.Context, deal *models.Deal, leadData *models.Lead, changedBy string) error {
	// Determine customer name from lead data or deal title
	customerName := deal.Title
	if leadData != nil {
		if leadData.CompanyName != "" {
			customerName = leadData.CompanyName
		} else if leadData.FirstName != "" {
			customerName = strings.TrimSpace(leadData.FirstName + " " + leadData.LastName)
		}
	}

	code, err := u.customerRepo.GetNextCode(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate customer code: %w", err)
	}

	newCustomer := &customerModels.Customer{
		ID:       uuid.New().String(),
		Code:     code,
		Name:     customerName,
		IsActive: true,
	}

	// Map lead fields to customer if available
	if leadData != nil {
		newCustomer.Email = leadData.Email
		newCustomer.ContactPerson = strings.TrimSpace(leadData.FirstName + " " + leadData.LastName)
		newCustomer.Address = leadData.Address
		newCustomer.ProvinceID = leadData.ProvinceID
		newCustomer.CityID = leadData.CityID
		newCustomer.DistrictID = leadData.DistrictID
		newCustomer.NPWP = leadData.NPWP
		newCustomer.Latitude = leadData.Latitude
		newCustomer.Longitude = leadData.Longitude
		newCustomer.Website = leadData.Website
		newCustomer.DefaultBusinessTypeID = leadData.BusinessTypeID
		newCustomer.DefaultAreaID = leadData.AreaID
		newCustomer.DefaultPaymentTermsID = leadData.PaymentTermsID

		if leadData.VillageName != "" {
			newCustomer.VillageName = &leadData.VillageName
		}
		if leadData.AssignedTo != nil {
			newCustomer.DefaultSalesRepID = leadData.AssignedTo
		}
		if changedBy != "" {
			newCustomer.CreatedBy = &changedBy
		}
	}

	if err := u.customerRepo.Create(ctx, newCustomer); err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	// Create contact from lead person info
	var newContact *models.Contact
	if leadData != nil && leadData.FirstName != "" {
		contactName := strings.TrimSpace(leadData.FirstName + " " + leadData.LastName)
		newContact = &models.Contact{
			ID:         uuid.New().String(),
			CustomerID: newCustomer.ID,
			Name:       contactName,
			Phone:      leadData.Phone,
			Email:      leadData.Email,
			Position:   leadData.JobTitle,
			IsActive:   true,
		}
		if changedBy != "" {
			newContact.CreatedBy = &changedBy
		}

		if err := u.contactRepo.Create(ctx, newContact); err != nil {
			return fmt.Errorf("failed to create contact: %w", err)
		}
	}

	// Update deal with customer and contact references
	customerID := newCustomer.ID
	deal.CustomerID = &customerID
	if newContact != nil {
		contactID := newContact.ID
		deal.ContactID = &contactID
	}
	if err := u.dealRepo.Update(ctx, deal); err != nil {
		return fmt.Errorf("failed to update deal with customer: %w", err)
	}

	// Update lead with customer and contact references
	if leadData != nil {
		leadData.CustomerID = &customerID
		if newContact != nil {
			contactID := newContact.ID
			leadData.ContactID = &contactID
		}
		// Nil out preloaded associations to prevent GORM FK override
		leadData.LeadSource = nil
		leadData.LeadStatus = nil
		leadData.AssignedEmployee = nil
		leadData.Customer = nil
		leadData.Contact = nil
		leadData.Deal = nil
		leadData.BusinessType = nil
		leadData.Area = nil
		leadData.PaymentTerms = nil
		leadData.Activities = nil

		if err := u.leadRepo.Update(ctx, leadData); err != nil {
			log.Printf("Warning: failed to update lead %s with customer reference: %v", leadData.ID, err)
		}
	}

	return nil
}

// SoftDeleteItem soft-deletes a single deal product item by ID (marks it as deleted without removing it).
func (u *dealUsecase) SoftDeleteItem(ctx context.Context, dealID, itemID string) error {
	if err := u.dealRepo.SoftDeleteItemByID(ctx, itemID, dealID); err != nil {
		return fmt.Errorf("failed to remove deal item: %w", err)
	}
	return nil
}

// RestoreItem restores a previously soft-deleted deal product item.
func (u *dealUsecase) RestoreItem(ctx context.Context, dealID, itemID string) error {
	if err := u.dealRepo.RestoreItemByID(ctx, itemID, dealID); err != nil {
		return fmt.Errorf("failed to restore deal item: %w", err)
	}
	return nil
}

// GetProductItems returns the product interest items for a deal.
// For lead-linked deals it reads from crm_lead_product_items (the live source of truth).
// For standalone deals it falls back to the deal's own crm_deal_product_items snapshot.
func (u *dealUsecase) GetProductItems(ctx context.Context, dealID string) ([]dto.LeadProductItemResponse, error) {
	deal, err := u.dealRepo.FindByID(ctx, dealID)
	if err != nil {
		return nil, errors.New("deal not found")
	}

	// Lead-linked deal: use lead items as the authoritative source
	if deal.LeadID != nil && *deal.LeadID != "" {
		items, err := u.leadRepo.ListProductItems(ctx, *deal.LeadID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch lead product items: %w", err)
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

	// Standalone deal: map from the deal's own product items snapshot
	result := make([]dto.LeadProductItemResponse, 0, len(deal.Items))
	for _, item := range deal.Items {
		result = append(result, dto.LeadProductItemResponse{
			ID:            item.ID,
			LeadID:        "",
			ProductID:     item.ProductID,
			ProductName:   item.ProductName,
			ProductSKU:    item.ProductSKU,
			InterestLevel: item.InterestLevel,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			Notes:         item.Notes,
			IsDeleted:     item.DeletedAt.Valid,
			CreatedAt:     item.CreatedAt.Format("2006-01-02T15:04:05+07:00"),
		})
	}
	return result, nil
}
