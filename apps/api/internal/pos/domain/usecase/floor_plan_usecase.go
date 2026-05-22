package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gilabs/gims/api/internal/core/apptime"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	orgRepo "github.com/gilabs/gims/api/internal/organization/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/data/repositories"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
	"github.com/gilabs/gims/api/internal/pos/domain/mapper"
)

var (
	ErrFloorPlanNotFound         = errors.New("floor plan not found")
	ErrFloorPlanForbidden        = errors.New("forbidden: you do not have access to this floor plan")
	ErrFloorPlanAlreadyExists    = errors.New("floor plan already exists for this outlet")
	ErrFloorPlanAlreadyPublished = errors.New("floor plan is already in published state")
	ErrVersionNotFound           = errors.New("layout version not found")
	ErrTableTokenStorageNotReady = errors.New("table token storage is not ready")
)

// FloorPlanUsecase defines business operations
type FloorPlanUsecase interface {
	Create(ctx context.Context, req *dto.CreateFloorPlanRequest, userID string, userCompanyID string, isOwner bool) (*dto.FloorPlanResponse, error)
	GetByID(ctx context.Context, id string, userCompanyID string, isOwner bool) (*dto.FloorPlanResponse, error)
	List(ctx context.Context, params repositories.FloorPlanListParams, userCompanyID string, isOwner bool) ([]dto.FloorPlanResponse, int64, error)
	Update(ctx context.Context, id string, req *dto.UpdateFloorPlanRequest, userCompanyID string, isOwner bool) (*dto.FloorPlanResponse, error)
	SaveLayoutData(ctx context.Context, id string, req *dto.SaveLayoutDataRequest, userCompanyID string, isOwner bool) (*dto.FloorPlanResponse, error)
	Delete(ctx context.Context, id string, userCompanyID string, isOwner bool) error
	Publish(ctx context.Context, id string, userID string, userCompanyID string, isOwner bool) (*dto.FloorPlanResponse, error)
	ListVersions(ctx context.Context, floorPlanID string, userCompanyID string, isOwner bool) ([]dto.LayoutVersionResponse, error)
	GetFormData(ctx context.Context, userCompanyID string, isOwner bool) (*dto.FloorPlanFormDataResponse, error)
	// Table QR token management
	GenerateTableToken(ctx context.Context, floorPlanID, tableObjectID string, req *dto.GenerateTableTokenRequest, userCompanyID string, isOwner bool) (*dto.TableQRTokenResponse, error)
	ListTableTokens(ctx context.Context, floorPlanID string, userCompanyID string, isOwner bool) ([]dto.TableQRTokenResponse, error)
	RevokeTableToken(ctx context.Context, floorPlanID, tableObjectID string, userCompanyID string, isOwner bool) error
}

type floorPlanUsecase struct {
	repo         repositories.FloorPlanRepository
	outletRepo   orgRepo.OutletRepository
	qrTokenRepo  repositories.TableQRTokenRepository
}

// NewFloorPlanUsecase creates a new instance
func NewFloorPlanUsecase(repo repositories.FloorPlanRepository, outletRepo orgRepo.OutletRepository, qrTokenRepo repositories.TableQRTokenRepository) FloorPlanUsecase {
	return &floorPlanUsecase{repo: repo, outletRepo: outletRepo, qrTokenRepo: qrTokenRepo}
}

func (u *floorPlanUsecase) resolveCompanyID(ctx context.Context, plan *models.FloorPlan) string {
	if plan.CompanyID != nil && *plan.CompanyID != "" {
		return *plan.CompanyID
	}
	if plan.OutletID == "" {
		return ""
	}
	outlet, err := u.outletRepo.GetByID(ctx, plan.OutletID)
	if err != nil || outlet == nil || outlet.CompanyID == nil {
		return ""
	}
	return *outlet.CompanyID
}

func (u *floorPlanUsecase) getOutletForCreate(ctx context.Context, req *dto.CreateFloorPlanRequest, userCompanyID string, isOwner bool) (*orgModels.Outlet, error) {
	if req.OutletID != "" {
		outlet, err := u.outletRepo.GetByID(ctx, req.OutletID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrFloorPlanNotFound
			}
			return nil, err
		}
		return outlet, nil
	}

	companyID := req.CompanyID
	
	// For tenant owners without employee record: use their company ID or resolve from tenant's first outlet
	if companyID == "" && isOwner {
		if userCompanyID != "" {
			companyID = userCompanyID
		} else {
			// Tenant owner without company ID: resolve first outlet for this tenant (auto-scoped by GetDB)
			isActive := true
			list, _, err := u.outletRepo.List(ctx, orgRepo.OutletListParams{
				IsActive: &isActive,
				Limit:    1,
				Offset:   0,
			})
			if err != nil {
				return nil, err
			}
			if len(list) > 0 {
				return list[0], nil
			}
			// No outlets found for tenant
			return nil, ErrFloorPlanNotFound
		}
	}

	// If still no company ID, fail
	if companyID == "" {
		return nil, ErrFloorPlanForbidden
	}

	// Resolve first active outlet from company_id
	isActive := true
	list, _, err := u.outletRepo.List(ctx, orgRepo.OutletListParams{
		CompanyID: companyID,
		IsActive:  &isActive,
		Limit:     1,
		Offset:    0,
	})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrFloorPlanNotFound
	}
	return list[0], nil
}

func (u *floorPlanUsecase) Create(ctx context.Context, req *dto.CreateFloorPlanRequest, userID string, userCompanyID string, isOwner bool) (*dto.FloorPlanResponse, error) {
	outlet, err := u.getOutletForCreate(ctx, req, userCompanyID, isOwner)
	if err != nil {
		return nil, err
	}

	// If outlet has no company assigned, allow tenant owners to proceed.
	if outlet.CompanyID == nil {
		if !isOwner {
			// Non-owners must have an outlet.company association
			log.Printf("[floorplan] forbidden: outlet has no company outlet_id=%s tenant_id=%v user_id=%s isOwner=%v userCompanyID=%s", outlet.ID, outlet.TenantID, userID, isOwner, userCompanyID)
			return nil, ErrFloorPlanForbidden
		}
		// Tenant owner: allow creation for outlet without company. plan.CompanyID will be nil.
	} else {
		// Non-owner can only create for outlet under their own company.
		if !isOwner && *outlet.CompanyID != userCompanyID {
			// Debug log to help trace ownership mismatch
			log.Printf("[floorplan] forbidden: non-owner company mismatch outlet_id=%s outlet_company=%s user_company=%s user_id=%s", outlet.ID, func() string { if outlet.CompanyID != nil { return *outlet.CompanyID }; return "<nil>" }(), userCompanyID, userID)
			return nil, ErrFloorPlanForbidden
		}
	}

	existingPlans, _, err := u.repo.List(ctx, repositories.FloorPlanListParams{
		OutletID: outlet.ID,
		Limit:    1,
		Offset:   0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check existing floor plans: %w", err)
	}
	if len(existingPlans) > 0 {
		return nil, ErrFloorPlanAlreadyExists
	}

	// Extract tenant_id from context
	tenantID, _ := ctx.Value("tenant_id").(string)

	plan := &models.FloorPlan{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		OutletID:    outlet.ID,
		CompanyID:   outlet.CompanyID,
		Name:        req.Name,
		FloorNumber: req.FloorNumber,
		Status:      models.FloorPlanStatusDraft,
		LayoutData:  "[]",
		CreatedBy:   &userID,
	}

	if req.GridSize != nil {
		plan.GridSize = *req.GridSize
	}
	if req.SnapToGrid != nil {
		plan.SnapToGrid = *req.SnapToGrid
	}
	if req.Width != nil {
		plan.Width = *req.Width
	}
	if req.Height != nil {
		plan.Height = *req.Height
	}

	if err = u.repo.Create(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create floor plan: %w", err)
	}

	return mapper.ToFloorPlanResponse(plan), nil
}

func (u *floorPlanUsecase) GetByID(ctx context.Context, id string, userCompanyID string, isOwner bool) (*dto.FloorPlanResponse, error) {
	plan, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFloorPlanNotFound
		}
		return nil, fmt.Errorf("failed to get floor plan: %w", err)
	}

	if !isOwner && u.resolveCompanyID(ctx, plan) != userCompanyID {
		return nil, ErrFloorPlanForbidden
	}

	return mapper.ToFloorPlanResponse(plan), nil
}

func (u *floorPlanUsecase) List(ctx context.Context, params repositories.FloorPlanListParams, userCompanyID string, isOwner bool) ([]dto.FloorPlanResponse, int64, error) {
	// Non-owner always scoped to their company
	if !isOwner {
		params.CompanyID = userCompanyID
	}

	plans, total, err := u.repo.List(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list floor plans: %w", err)
	}

	return mapper.ToFloorPlanListResponse(plans), total, nil
}

func (u *floorPlanUsecase) Update(ctx context.Context, id string, req *dto.UpdateFloorPlanRequest, userCompanyID string, isOwner bool) (*dto.FloorPlanResponse, error) {
	plan, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFloorPlanNotFound
		}
		return nil, fmt.Errorf("failed to find floor plan: %w", err)
	}

	if !isOwner && u.resolveCompanyID(ctx, plan) != userCompanyID {
		return nil, ErrFloorPlanForbidden
	}

	if req.Name != nil {
		plan.Name = *req.Name
	}
	if req.FloorNumber != nil {
		plan.FloorNumber = *req.FloorNumber
	}
	if req.GridSize != nil {
		plan.GridSize = *req.GridSize
	}
	if req.SnapToGrid != nil {
		plan.SnapToGrid = *req.SnapToGrid
	}
	if req.Width != nil {
		plan.Width = *req.Width
	}
	if req.Height != nil {
		plan.Height = *req.Height
	}

	if err := u.repo.Update(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to update floor plan: %w", err)
	}

	return mapper.ToFloorPlanResponse(plan), nil
}

func (u *floorPlanUsecase) SaveLayoutData(ctx context.Context, id string, req *dto.SaveLayoutDataRequest, userCompanyID string, isOwner bool) (*dto.FloorPlanResponse, error) {
	plan, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFloorPlanNotFound
		}
		return nil, fmt.Errorf("failed to find floor plan: %w", err)
	}

	if !isOwner && u.resolveCompanyID(ctx, plan) != userCompanyID {
		return nil, ErrFloorPlanForbidden
	}

	plan.LayoutData = string(req.LayoutData)
	// Reset published status on layout change
	plan.Status = models.FloorPlanStatusDraft

	if err := u.repo.Update(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to save layout data: %w", err)
	}

	return mapper.ToFloorPlanResponse(plan), nil
}

func (u *floorPlanUsecase) Delete(ctx context.Context, id string, userCompanyID string, isOwner bool) error {
	plan, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFloorPlanNotFound
		}
		return fmt.Errorf("failed to find floor plan: %w", err)
	}

	if !isOwner && u.resolveCompanyID(ctx, plan) != userCompanyID {
		return ErrFloorPlanForbidden
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete floor plan: %w", err)
	}

	return nil
}

func (u *floorPlanUsecase) Publish(ctx context.Context, id string, userID string, userCompanyID string, isOwner bool) (*dto.FloorPlanResponse, error) {
	plan, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFloorPlanNotFound
		}
		return nil, fmt.Errorf("failed to find floor plan: %w", err)
	}

	if !isOwner && u.resolveCompanyID(ctx, plan) != userCompanyID {
		return nil, ErrFloorPlanForbidden
	}

	// Prevent re-publishing an already-published floor plan.
	if plan.Status == models.FloorPlanStatusPublished {
		return nil, ErrFloorPlanAlreadyPublished
	}

	// Extract tenant_id from context for version record
	tenantID, _ := ctx.Value("tenant_id").(string)

	now := apptime.Now()
	plan.Version++
	plan.Status = models.FloorPlanStatusPublished
	plan.PublishedAt = &now
	plan.PublishedBy = &userID

	// Create immutable version snapshot
	version := &models.LayoutVersion{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		FloorPlanID: plan.ID,
		Version:     plan.Version,
		LayoutData:  plan.LayoutData,
		PublishedAt: now,
		PublishedBy: userID,
	}

	if err := u.repo.CreateVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create layout version: %w", err)
	}

	if err := u.repo.Update(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to publish floor plan: %w", err)
	}

	return mapper.ToFloorPlanResponse(plan), nil
}

func (u *floorPlanUsecase) ListVersions(ctx context.Context, floorPlanID string, userCompanyID string, isOwner bool) ([]dto.LayoutVersionResponse, error) {
	// Verify access to the floor plan
	plan, err := u.repo.FindByID(ctx, floorPlanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFloorPlanNotFound
		}
		return nil, fmt.Errorf("failed to find floor plan: %w", err)
	}

	if !isOwner && u.resolveCompanyID(ctx, plan) != userCompanyID {
		return nil, ErrFloorPlanForbidden
	}

	versions, err := u.repo.ListVersions(ctx, floorPlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	return mapper.ToLayoutVersionListResponse(versions), nil
}

func (u *floorPlanUsecase) GetFormData(ctx context.Context, userCompanyID string, isOwner bool) (*dto.FloorPlanFormDataResponse, error) {
	var outlets []dto.OutletOption
	isActive := true
	params := orgRepo.OutletListParams{
		IsActive: &isActive,
		Limit:    100,
		Offset:   0,
	}
	if !isOwner {
		params.CompanyID = userCompanyID
	}

	floorPlanParams := repositories.FloorPlanListParams{}
	if !isOwner {
		floorPlanParams.CompanyID = userCompanyID
	}

	floorPlans, _, err := u.repo.List(ctx, floorPlanParams)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch floor plans: %w", err)
	}

	hasFloorPlanByOutlet := make(map[string]bool, len(floorPlans))
	for _, plan := range floorPlans {
		if plan.OutletID == "" {
			continue
		}
		hasFloorPlanByOutlet[plan.OutletID] = true
	}

	allOutlets, _, err := u.outletRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch outlets: %w", err)
	}

	for _, outlet := range allOutlets {
		outletUUID, parseErr := uuid.Parse(outlet.ID)
		if parseErr != nil {
			continue
		}
		outlets = append(outlets, dto.OutletOption{
			ID:           outletUUID,
			Name:         outlet.Name,
			HasFloorPlan: hasFloorPlanByOutlet[outlet.ID],
		})
	}

	if outlets == nil {
		outlets = []dto.OutletOption{}
	}

	return &dto.FloorPlanFormDataResponse{Outlets: outlets}, nil
}

// ─── Table QR Token methods ───────────────────────────────────────────────────

// GenerateTableToken creates (or rotates) a QR token for the given table object.
func (u *floorPlanUsecase) GenerateTableToken(ctx context.Context, floorPlanID, tableObjectID string, req *dto.GenerateTableTokenRequest, userCompanyID string, isOwner bool) (*dto.TableQRTokenResponse, error) {
	plan, err := u.repo.FindByID(ctx, floorPlanID)
	if err != nil {
		return nil, ErrFloorPlanNotFound
	}
	if !isOwner && u.resolveCompanyID(ctx, plan) != userCompanyID {
		return nil, ErrFloorPlanForbidden
	}

	token := &models.PosTableQRToken{
		TenantID:      plan.TenantID,
		OutletID:      plan.OutletID,
		FloorPlanID:   floorPlanID,
		TableObjectID: tableObjectID,
		TableLabel:    req.TableLabel,
		IsActive:      true,
	}

	created, err := u.qrTokenRepo.GenerateForTable(ctx, token)
	if err != nil {
		if errors.Is(err, repositories.ErrTableQRTokenTableNotReady) {
			return nil, ErrTableTokenStorageNotReady
		}
		return nil, fmt.Errorf("failed to generate table token: %w", err)
	}
	return toTableQRTokenResponse(created), nil
}

// ListTableTokens returns all active QR tokens for a floor plan.
func (u *floorPlanUsecase) ListTableTokens(ctx context.Context, floorPlanID string, userCompanyID string, isOwner bool) ([]dto.TableQRTokenResponse, error) {
	plan, err := u.repo.FindByID(ctx, floorPlanID)
	if err != nil {
		return nil, ErrFloorPlanNotFound
	}
	if !isOwner && u.resolveCompanyID(ctx, plan) != userCompanyID {
		return nil, ErrFloorPlanForbidden
	}

	tokens, err := u.qrTokenRepo.FindByFloorPlan(ctx, floorPlanID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.TableQRTokenResponse, 0, len(tokens))
	for i := range tokens {
		result = append(result, *toTableQRTokenResponse(&tokens[i]))
	}
	return result, nil
}

// RevokeTableToken deactivates a QR token so its URL is no longer valid.
func (u *floorPlanUsecase) RevokeTableToken(ctx context.Context, floorPlanID, tableObjectID string, userCompanyID string, isOwner bool) error {
	plan, err := u.repo.FindByID(ctx, floorPlanID)
	if err != nil {
		return ErrFloorPlanNotFound
	}
	if !isOwner && u.resolveCompanyID(ctx, plan) != userCompanyID {
		return ErrFloorPlanForbidden
	}
	return u.qrTokenRepo.Revoke(ctx, floorPlanID, tableObjectID)
}

func toTableQRTokenResponse(t *models.PosTableQRToken) *dto.TableQRTokenResponse {
	return &dto.TableQRTokenResponse{
		ID:            t.ID,
		FloorPlanID:   t.FloorPlanID,
		TableObjectID: t.TableObjectID,
		TableLabel:    t.TableLabel,
		Token:         t.Token,
		IsActive:      t.IsActive,
		CreatedAt:     t.CreatedAt.Format(time.RFC3339),
	}
}
