package repositories

import (
	"context"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"

	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const dealQueryByPipelineStageID = "pipeline_stage_id = ?"

// DealListParams defines filtering/sorting/pagination parameters for deal queries
type DealListParams struct {
	Search          string
	SortBy          string
	SortDir         string
	Limit           int
	Offset          int
	Status          string
	PipelineStageID string
	CustomerID      string
	AssignedTo      string
	LeadID          string
	DateFrom        string
	DateTo          string
}

// DealsByStageParams defines parameters for the Kanban board view
type DealsByStageParams struct {
	StageID string
	Limit   int
	Offset  int
	Search  string
	Status  string
}

// StageSummary holds aggregated stage statistics for pipeline summary
type StageSummary struct {
	StageID    string  `json:"stage_id"`
	StageName  string  `json:"stage_name"`
	StageColor string  `json:"stage_color"`
	StageOrder int     `json:"stage_order"`
	DealCount  int64   `json:"deal_count"`
	TotalValue float64 `json:"total_value"`
}

// PipelineSummaryData holds the complete pipeline summary
type PipelineSummaryData struct {
	TotalDeals int64          `json:"total_deals"`
	TotalValue float64        `json:"total_value"`
	OpenDeals  int64          `json:"open_deals"`
	OpenValue  float64        `json:"open_value"`
	WonDeals   int64          `json:"won_deals"`
	WonValue   float64        `json:"won_value"`
	LostDeals  int64          `json:"lost_deals"`
	LostValue  float64        `json:"lost_value"`
	ByStage    []StageSummary `json:"by_stage"`
}

// ForecastData holds weighted deal forecast
type ForecastData struct {
	TotalWeightedValue float64         `json:"total_weighted_value"`
	TotalDeals         int64           `json:"total_deals"`
	ByStage            []StageForecast `json:"by_stage"`
}

// StageForecast holds forecast per stage
type StageForecast struct {
	StageID       string  `json:"stage_id"`
	StageName     string  `json:"stage_name"`
	DealCount     int64   `json:"deal_count"`
	TotalValue    float64 `json:"total_value"`
	Probability   int     `json:"probability"`
	WeightedValue float64 `json:"weighted_value"`
}

// DealRepository defines data access methods for deals
type DealRepository interface {
	Create(ctx context.Context, deal *models.Deal) error
	FindByID(ctx context.Context, id string) (*models.Deal, error)
	List(ctx context.Context, params DealListParams) ([]models.Deal, int64, error)
	ListByStage(ctx context.Context, params DealsByStageParams) ([]models.Deal, int64, error)
	Update(ctx context.Context, deal *models.Deal) error
	Delete(ctx context.Context, id string) error
	CreateHistory(ctx context.Context, history *models.DealHistory) error
	GetHistory(ctx context.Context, dealID string) ([]models.DealHistory, error)
	GetPipelineSummary(ctx context.Context) (*PipelineSummaryData, error)
	GetForecast(ctx context.Context) (*ForecastData, error)
	DeleteItemsByDealID(ctx context.Context, dealID string) error
	CreateItems(ctx context.Context, items []models.DealProductItem) error
	SoftDeleteItemByID(ctx context.Context, itemID, dealID string) error
	RestoreItemByID(ctx context.Context, itemID, dealID string) error
	GetLastHistoryByDealID(ctx context.Context, dealID string) (*models.DealHistory, error)
	UpdateProbabilityByStageID(ctx context.Context, stageID string, probability int) error
	// UpsertProductItemsFromVisit merges visit-sourced product items into a deal.
	// For each item, if a non-deleted record already exists with the same product_id it is updated;
	// otherwise a new row is created. Items already on the deal that are not in the visit are untouched.
	UpsertProductItemsFromVisit(ctx context.Context, dealID string, items []models.DealProductItem) error
	// ExistsByLeadID reports whether any non-deleted deal is linked to the given lead ID.
	// Uses tenant scoping only (no user-ownership scope) to catch orphaned deals regardless
	// of which employee the deal is assigned to.
	ExistsByLeadID(ctx context.Context, leadID string) (bool, error)
}

type dealRepository struct {
	db *gorm.DB
}

// NewDealRepository creates a new deal repository instance
func NewDealRepository(db *gorm.DB) DealRepository {
	return &dealRepository{db: db}
}

func (r *dealRepository) Create(ctx context.Context, deal *models.Deal) error {
	return database.GetDB(ctx, r.db).Create(deal).Error
}

func (r *dealRepository) FindByID(ctx context.Context, id string) (*models.Deal, error) {
	var deal models.Deal
	err := database.GetDB(ctx, r.db).
		Preload("PipelineStage").
		Preload("Customer").
		Preload("Contact").
		Preload("AssignedEmployee").
		Preload("Lead").
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Unscoped().Order("created_at ASC") }).
		Preload("Items.Product").
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Order("CASE WHEN status IN ('pending','in_progress') THEN 0 ELSE 1 END, due_date ASC NULLS LAST").Limit(20)
		}).
		Preload("Tasks.AssignedEmployee").
		First(&deal, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &deal, nil
}

func (r *dealRepository) List(ctx context.Context, params DealListParams) ([]models.Deal, int64, error) {
	query := database.GetDB(ctx, r.db).Model(&models.Deal{})
	query = security.ApplyScopeFilter(query, ctx, security.MixedOwnershipScopeQueryOptions("assigned_to"))

	// Search filter (prefix search for indexed columns)
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(
			"title ILIKE ? OR code ILIKE ?",
			searchTerm, searchTerm,
		)
	}

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.PipelineStageID != "" {
		query = query.Where(dealQueryByPipelineStageID, params.PipelineStageID)
	}
	if params.CustomerID != "" {
		query = query.Where("customer_id = ?", params.CustomerID)
	}
	if params.AssignedTo != "" {
		query = query.Where("assigned_to = ?", params.AssignedTo)
	}
	if params.LeadID != "" {
		query = query.Where("lead_id = ?", params.LeadID)
	}
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo+" 23:59:59")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	sortBy := "created_at"
	sortDir := "DESC"
	allowedSorts := map[string]bool{
		"code": true, "title": true, "value": true, "probability": true,
		"status": true, "created_at": true, "updated_at": true, "expected_close_date": true,
	}
	if params.SortBy != "" && allowedSorts[params.SortBy] {
		sortBy = params.SortBy
	}
	if params.SortDir != "" && (strings.EqualFold(params.SortDir, "ASC") || strings.EqualFold(params.SortDir, "DESC")) {
		sortDir = strings.ToUpper(params.SortDir)
	}

	var deals []models.Deal
	err := query.
		Preload("PipelineStage").
		Preload("Customer").
		Preload("Contact").
		Preload("AssignedEmployee").
		Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: sortDir == "DESC"}).
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&deals).Error

	return deals, total, err
}

func (r *dealRepository) ListByStage(ctx context.Context, params DealsByStageParams) ([]models.Deal, int64, error) {
	query := database.GetDB(ctx, r.db).Model(&models.Deal{}).Where(dealQueryByPipelineStageID, params.StageID)
	query = security.ApplyScopeFilter(query, ctx, security.MixedOwnershipScopeQueryOptions("assigned_to"))

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("title ILIKE ? OR code ILIKE ?", searchTerm, searchTerm)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var deals []models.Deal
	err := query.
		Preload("PipelineStage").
		Preload("Customer").
		Preload("Contact").
		Preload("AssignedEmployee").
		Preload("Items").
		Order("updated_at DESC").
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&deals).Error

	return deals, total, err
}

func (r *dealRepository) Update(ctx context.Context, deal *models.Deal) error {
	// Use explicit Select to only update scalar/FK columns, bypassing GORM association
	// handling entirely. This prevents BelongsTo callbacks from overriding FK values.
	return database.GetDB(ctx, r.db).
		Select(
			"pipeline_stage_id", "title", "description", "status",
			"value", "probability",
			"expected_close_date", "actual_close_date", "close_reason",
			"customer_id", "contact_id", "assigned_to", "lead_id",
			"budget_confirmed", "budget_amount",
			"auth_confirmed", "auth_person",
			"need_confirmed", "need_description",
			"time_confirmed", "notes",
			"converted_to_quotation_id", "converted_at",
		).
		Save(deal).Error
}

func (r *dealRepository) Delete(ctx context.Context, id string) error {
	return database.GetDB(ctx, r.db).Where("id = ?", id).Delete(&models.Deal{}).Error
}

func (r *dealRepository) CreateHistory(ctx context.Context, history *models.DealHistory) error {
	return database.GetDB(ctx, r.db).Create(history).Error
}

func (r *dealRepository) GetHistory(ctx context.Context, dealID string) ([]models.DealHistory, error) {
	var history []models.DealHistory
	err := database.GetDB(ctx, r.db).
		Where("deal_id = ?", dealID).
		Preload("FromStage").
		Preload("ToStage").
		Preload("ChangedByEmployee").
		Order("changed_at DESC").
		Find(&history).Error
	return history, err
}

func (r *dealRepository) GetLastHistoryByDealID(ctx context.Context, dealID string) (*models.DealHistory, error) {
	var history models.DealHistory
	err := database.GetDB(ctx, r.db).
		Where("deal_id = ?", dealID).
		Order("changed_at DESC").
		First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *dealRepository) UpdateProbabilityByStageID(ctx context.Context, stageID string, probability int) error {
	return database.GetDB(ctx, r.db).
		Model(&models.Deal{}).
		Where(dealQueryByPipelineStageID, stageID).
		Update("probability", probability).Error
}

func (r *dealRepository) DeleteItemsByDealID(ctx context.Context, dealID string) error {
	// Hard-delete previously soft-deleted items to prevent unbounded accumulation
	if err := r.db.Unscoped().WithContext(ctx).
		Where("deal_id = ? AND deleted_at IS NOT NULL", dealID).
		Delete(&models.DealProductItem{}).Error; err != nil {
		return err
	}
	// Soft-delete currently active items so they appear struck-through in the detail view
	return database.GetDB(ctx, r.db).Where("deal_id = ?", dealID).Delete(&models.DealProductItem{}).Error
}

func (r *dealRepository) CreateItems(ctx context.Context, items []models.DealProductItem) error {
	if len(items) == 0 {
		return nil
	}
	return database.GetDB(ctx, r.db).Create(&items).Error
}

func (r *dealRepository) SoftDeleteItemByID(ctx context.Context, itemID, dealID string) error {
	return database.GetDB(ctx, r.db).Where("id = ? AND deal_id = ?", itemID, dealID).Delete(&models.DealProductItem{}).Error
}

func (r *dealRepository) RestoreItemByID(ctx context.Context, itemID, dealID string) error {
	return r.db.Unscoped().WithContext(ctx).Model(&models.DealProductItem{}).
		Where("id = ? AND deal_id = ?", itemID, dealID).
		Update("deleted_at", nil).Error
}

// UpsertProductItemsFromVisit merges visit-sourced product items into crm_deal_product_items.
// Existing (non-deleted) rows matched by (deal_id, product_id) are updated in place;
// absent rows are created. No deletions are performed so manually-added items are preserved.
func (r *dealRepository) UpsertProductItemsFromVisit(ctx context.Context, dealID string, items []models.DealProductItem) error {
	if len(items) == 0 {
		return nil
	}

	// Load existing active items keyed by product_id for O(1) lookup.
	var existing []models.DealProductItem
	if err := r.db.WithContext(ctx).
		Where("deal_id = ?", dealID).
		Find(&existing).Error; err != nil {
		return err
	}
	existingByProductID := make(map[string]*models.DealProductItem, len(existing))
	for i := range existing {
		if existing[i].ProductID != nil {
			existingByProductID[*existing[i].ProductID] = &existing[i]
		}
	}

	db := r.db.WithContext(ctx)
	for i := range items {
		items[i].DealID = dealID
		if items[i].ProductID == nil {
			continue
		}
		productID := *items[i].ProductID
		if ex, ok := existingByProductID[productID]; ok {
			// Update mutable fields; preserve ID, DealID, and creation metadata.
			if err := db.Model(ex).Updates(map[string]interface{}{
				"product_name":   items[i].ProductName,
				"product_sku":    items[i].ProductSKU,
				"unit_price":     items[i].UnitPrice,
				"quantity":       items[i].Quantity,
				"subtotal":       items[i].Subtotal,
				"notes":          items[i].Notes,
				"interest_level": items[i].InterestLevel,
			}).Error; err != nil {
				return err
			}
		} else {
			if err := db.Create(&items[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *dealRepository) ExistsByLeadID(ctx context.Context, leadID string) (bool, error) {
	var count int64
	err := database.GetDB(ctx, r.db).
		Model(&models.Deal{}).
		Where("lead_id = ?", leadID).
		Count(&count).Error
	return count > 0, err
}

func (r *dealRepository) GetPipelineSummary(ctx context.Context) (*PipelineSummaryData, error) {
	summary := &PipelineSummaryData{}

	// Total deals
	database.GetDB(ctx, r.db).Model(&models.Deal{}).Count(&summary.TotalDeals)
	database.GetDB(ctx, r.db).Model(&models.Deal{}).Select("COALESCE(SUM(value), 0)").Scan(&summary.TotalValue)

	// Open deals
	database.GetDB(ctx, r.db).Model(&models.Deal{}).Where("status = ?", "open").Count(&summary.OpenDeals)
	database.GetDB(ctx, r.db).Model(&models.Deal{}).Where("status = ?", "open").Select("COALESCE(SUM(value), 0)").Scan(&summary.OpenValue)

	// Won deals
	database.GetDB(ctx, r.db).Model(&models.Deal{}).Where("status = ?", "won").Count(&summary.WonDeals)
	database.GetDB(ctx, r.db).Model(&models.Deal{}).Where("status = ?", "won").Select("COALESCE(SUM(value), 0)").Scan(&summary.WonValue)

	// Lost deals
	database.GetDB(ctx, r.db).Model(&models.Deal{}).Where("status = ?", "lost").Count(&summary.LostDeals)
	database.GetDB(ctx, r.db).Model(&models.Deal{}).Where("status = ?", "lost").Select("COALESCE(SUM(value), 0)").Scan(&summary.LostValue)

	// By stage
	database.GetDB(ctx, r.db).
		Table("crm_deals").
		Select("crm_pipeline_stages.id as stage_id, crm_pipeline_stages.name as stage_name, crm_pipeline_stages.color as stage_color, crm_pipeline_stages.\"order\" as stage_order, COUNT(*) as deal_count, COALESCE(SUM(crm_deals.value), 0) as total_value").
		Joins("LEFT JOIN crm_pipeline_stages ON crm_deals.pipeline_stage_id = crm_pipeline_stages.id").
		Where("crm_deals.deleted_at IS NULL").
		Group("crm_pipeline_stages.id, crm_pipeline_stages.name, crm_pipeline_stages.color, crm_pipeline_stages.\"order\"").
		Order("crm_pipeline_stages.\"order\" ASC").
		Scan(&summary.ByStage)

	return summary, nil
}

func (r *dealRepository) GetForecast(ctx context.Context) (*ForecastData, error) {
	forecast := &ForecastData{}

	// Only open deals contribute to forecast
	database.GetDB(ctx, r.db).Model(&models.Deal{}).Where("status = ?", "open").Count(&forecast.TotalDeals)

	// Weighted value = sum(value * probability / 100)
	database.GetDB(ctx, r.db).Model(&models.Deal{}).
		Where("status = ?", "open").
		Select("COALESCE(SUM(value * probability / 100.0), 0)").
		Scan(&forecast.TotalWeightedValue)

	// By stage
	database.GetDB(ctx, r.db).
		Table("crm_deals").
		Select("crm_pipeline_stages.id as stage_id, crm_pipeline_stages.name as stage_name, COUNT(*) as deal_count, COALESCE(SUM(crm_deals.value), 0) as total_value, crm_pipeline_stages.probability, COALESCE(SUM(crm_deals.value * crm_deals.probability / 100.0), 0) as weighted_value").
		Joins("LEFT JOIN crm_pipeline_stages ON crm_deals.pipeline_stage_id = crm_pipeline_stages.id").
		Where("crm_deals.deleted_at IS NULL AND crm_deals.status = ?", "open").
		Group("crm_pipeline_stages.id, crm_pipeline_stages.name, crm_pipeline_stages.probability").
		Order("crm_pipeline_stages.probability ASC").
		Scan(&forecast.ByStage)

	return forecast, nil
}
