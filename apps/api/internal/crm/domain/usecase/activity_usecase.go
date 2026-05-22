package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"github.com/gilabs/gims/api/internal/crm/data/repositories"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"github.com/gilabs/gims/api/internal/crm/domain/mapper"
	"github.com/google/uuid"
)

// ActivityUsecase defines business logic for CRM activities
type ActivityUsecase interface {
	Create(ctx context.Context, req dto.CreateActivityRequest, employeeID string) (dto.ActivityResponse, error)
	GetByID(ctx context.Context, id string) (dto.ActivityResponse, error)
	List(ctx context.Context, params repositories.ActivityListParams) ([]dto.ActivityResponse, int64, error)
	Timeline(ctx context.Context, params repositories.ActivityListParams) ([]dto.ActivityResponse, int64, error)
}

type activityUsecase struct {
	activityRepo     repositories.ActivityRepository
	activityTypeRepo repositories.ActivityTypeRepository
	leadRepo         repositories.LeadRepository
	dealRepo         repositories.DealRepository
}

// NewActivityUsecase creates a new activity usecase
func NewActivityUsecase(
	activityRepo repositories.ActivityRepository,
	activityTypeRepo repositories.ActivityTypeRepository,
	leadRepo repositories.LeadRepository,
	dealRepo repositories.DealRepository,
) ActivityUsecase {
	return &activityUsecase{
		activityRepo:     activityRepo,
		activityTypeRepo: activityTypeRepo,
		leadRepo:         leadRepo,
		dealRepo:         dealRepo,
	}
}

func (u *activityUsecase) Create(ctx context.Context, req dto.CreateActivityRequest, employeeID string) (dto.ActivityResponse, error) {
	// Validate activity type if provided
	if req.ActivityTypeID != nil && *req.ActivityTypeID != "" {
		_, err := u.activityTypeRepo.FindByID(ctx, *req.ActivityTypeID)
		if err != nil {
			return dto.ActivityResponse{}, errors.New("activity type not found")
		}
	}

	// Parse timestamp or use current time
	timestamp := apptime.Now()
	if req.Timestamp != nil && *req.Timestamp != "" {
		t, err := time.Parse(time.RFC3339, *req.Timestamp)
		if err != nil {
			return dto.ActivityResponse{}, errors.New("invalid timestamp format, use ISO 8601")
		}
		timestamp = t
	}

	activity := &models.Activity{
		ID:             uuid.New().String(),
		Type:           req.Type,
		ActivityTypeID: req.ActivityTypeID,
		CustomerID:     req.CustomerID,
		ContactID:      req.ContactID,
		DealID:         req.DealID,
		LeadID:         req.LeadID,
		VisitReportID:  req.VisitReportID,
		EmployeeID:     employeeID,
		Description:    req.Description,
		Timestamp: timestamp,
		Metadata: func() *string {
			if len(req.Metadata) == 0 || string(req.Metadata) == "null" {
				return nil
			}
			s := string(req.Metadata)
			return &s
		}(),
	}

	if err := u.activityRepo.Create(ctx, activity); err != nil {
		return dto.ActivityResponse{}, fmt.Errorf("failed to create activity: %w", err)
	}

	// Reload with preloaded relations
	created, err := u.activityRepo.FindByID(ctx, activity.ID)
	if err != nil {
		return dto.ActivityResponse{}, err
	}

	// Sync product items from metadata to lead and deal (non-blocking — failures are logged, not surfaced)
	if req.LeadID != nil && *req.LeadID != "" && activity.Metadata != nil {
		dealID := ""
		if req.DealID != nil {
			dealID = *req.DealID
		}
		go u.syncProductItemsFromMetadata(context.Background(), *req.LeadID, dealID, *activity.Metadata, activity.ID)
	}

	return mapper.ToActivityResponse(created), nil
}

// syncProductItemsFromMetadata parses activity metadata products and merges them into the lead's
// product items and (if dealID is provided) the deal's product items.
func (u *activityUsecase) syncProductItemsFromMetadata(ctx context.Context, leadID string, dealID string, metadataJSON string, activityID string) {
	type answerInfo struct {
		QuestionID string `json:"question_id"`
		OptionID   string `json:"option_id"`
		Answer     bool   `json:"answer,omitempty"`
	}

	type metaProduct struct {
		ProductID     string       `json:"product_id"`
		ProductName   string       `json:"product_name"`
		ProductSKU    string       `json:"product_sku"`
		InterestLevel int          `json:"interest_level"`
		Quantity      *int         `json:"quantity"`
		UnitPrice     float64      `json:"unit_price"`
		Notes         string       `json:"notes"`
		Answers       []answerInfo `json:"answers"`
	}

	type activityMeta struct {
		Products []metaProduct `json:"products"`
	}

	var meta activityMeta
	if err := json.Unmarshal([]byte(metadataJSON), &meta); err != nil || len(meta.Products) == 0 {
		return
	}

	// Fetch existing product items to merge (preserve items from other sources)
	existing, err := u.leadRepo.ListProductItems(ctx, leadID)
	if err != nil {
		fmt.Printf("[WARN] activity sync: failed to fetch lead product items for lead %s: %v\n", leadID, err)
		return
	}

	existingByProductID := make(map[string]*models.LeadProductItem)
	for i := range existing {
		if existing[i].ProductID != nil {
			existingByProductID[*existing[i].ProductID] = &existing[i]
		}
	}

	var items []models.LeadProductItem
	seen := make(map[string]bool)

	for _, p := range meta.Products {
		if p.ProductID == "" || seen[p.ProductID] {
			continue
		}
		seen[p.ProductID] = true

		pid := p.ProductID
		item := models.LeadProductItem{
			LeadID:        leadID,
			ProductID:     &pid,
			ProductName:   p.ProductName,
			ProductSKU:    p.ProductSKU,
			InterestLevel: p.InterestLevel,
			Notes:         p.Notes,
		}

		if p.Quantity != nil {
			item.Quantity = *p.Quantity
		}

		if p.UnitPrice > 0 {
			item.UnitPrice = p.UnitPrice
		}

		if len(p.Answers) > 0 {
			if b, err := json.Marshal(p.Answers); err == nil {
				s := string(b)
				item.LastSurveyAnswers = &s
			}
		}

		// Preserve existing item ID to update rather than create duplicate
		if existingItem, ok := existingByProductID[pid]; ok {
			item.ID = existingItem.ID
		}

		items = append(items, item)
	}

	if err := u.leadRepo.UpsertProductItems(ctx, leadID, items); err != nil {
		fmt.Printf("[WARN] activity sync: failed to upsert product items for lead %s: %v\n", leadID, err)
	}

	// If the activity is also linked to a deal, mirror the product items there too.
	if dealID == "" {
		return
	}
	dealItems := make([]models.DealProductItem, 0, len(items))
	for _, li := range items {
		if li.ProductID == nil {
			continue
		}
		di := models.DealProductItem{
			DealID:        dealID,
			ProductID:     li.ProductID,
			ProductName:   li.ProductName,
			ProductSKU:    li.ProductSKU,
			InterestLevel: li.InterestLevel,
			UnitPrice:     li.UnitPrice,
			Quantity:      li.Quantity,
			Notes:         li.Notes,
		}
		di.Subtotal = di.UnitPrice * float64(di.Quantity)
		dealItems = append(dealItems, di)
	}
	if err := u.dealRepo.UpsertProductItemsFromVisit(ctx, dealID, dealItems); err != nil {
		fmt.Printf("[WARN] activity sync: failed to upsert product items for deal %s: %v\n", dealID, err)
	}
}


func (u *activityUsecase) GetByID(ctx context.Context, id string) (dto.ActivityResponse, error) {
	if !security.CheckRecordScopeAccess(database.DB, ctx, &models.Activity{}, id, security.MixedOwnershipScopeQueryOptions("employee_id")) {
		return dto.ActivityResponse{}, errors.New("activity not found")
	}
	activity, err := u.activityRepo.FindByID(ctx, id)
	if err != nil {
		return dto.ActivityResponse{}, errors.New("activity not found")
	}
	return mapper.ToActivityResponse(activity), nil
}

func (u *activityUsecase) List(ctx context.Context, params repositories.ActivityListParams) ([]dto.ActivityResponse, int64, error) {
	activities, total, err := u.activityRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToActivityResponseList(activities), total, nil
}

func (u *activityUsecase) Timeline(ctx context.Context, params repositories.ActivityListParams) ([]dto.ActivityResponse, int64, error) {
	activities, total, err := u.activityRepo.Timeline(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	return mapper.ToActivityResponseList(activities), total, nil
}
