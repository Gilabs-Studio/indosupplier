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

// LeadListParams defines filtering/sorting/pagination parameters for lead queries
type LeadListParams struct {
	Search       string
	SortBy       string
	SortDir      string
	Limit        int
	Offset       int
	LeadStatusID string
	LeadSourceID string
	AssignedTo   string
	ScoreMin     *int
	ScoreMax     *int
	DateFrom     string
	DateTo       string
	IsConverted  *bool
}

// LeadRepository defines data access methods for leads
type LeadRepository interface {
	Create(ctx context.Context, lead *models.Lead) error
	FindByID(ctx context.Context, id string) (*models.Lead, error)
	FindByEmail(ctx context.Context, email string) (*models.Lead, error)
	FindDuplicate(ctx context.Context, email, phone, companyName, placeID, cid string) (*models.Lead, error)
	FindUnprocessed(ctx context.Context, limit int) ([]models.Lead, error)
	List(ctx context.Context, params LeadListParams) ([]models.Lead, int64, error)
	Update(ctx context.Context, lead *models.Lead) error
	Delete(ctx context.Context, id string) error
	ExistsByCode(ctx context.Context, code string) (bool, error)
	GetAnalytics(ctx context.Context) (*LeadAnalytics, error)
	// Product items
	ListProductItems(ctx context.Context, leadID string) ([]models.LeadProductItem, error)
	UpsertProductItems(ctx context.Context, leadID string, items []models.LeadProductItem) error
}

// LeadAnalytics holds aggregated lead statistics
type LeadAnalytics struct {
	TotalLeads     int64                `json:"total_leads"`
	ByStatus       []LeadCountByField   `json:"by_status"`
	BySource       []LeadCountByField   `json:"by_source"`
	ConversionRate float64              `json:"conversion_rate"`
	AvgScore       float64              `json:"avg_score"`
}

// LeadCountByField holds count grouped by a field
type LeadCountByField struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
	Count int64  `json:"count"`
}

type leadRepository struct {
	db *gorm.DB
}

// NewLeadRepository creates a new lead repository instance
func NewLeadRepository(db *gorm.DB) LeadRepository {
	return &leadRepository{db: db}
}

func (r *leadRepository) scopedLeadQuery(ctx context.Context) *gorm.DB {
	query := database.GetDB(ctx, r.db).Model(&models.Lead{})
	return security.ApplyScopeFilter(query, ctx, security.MixedOwnershipScopeQueryOptions("assigned_to"))
}

func (r *leadRepository) Create(ctx context.Context, lead *models.Lead) error {
	return database.GetDB(ctx, r.db).Create(lead).Error
}

func (r *leadRepository) FindByID(ctx context.Context, id string) (*models.Lead, error) {
	var lead models.Lead
	err := r.scopedLeadQuery(ctx).
		Preload("LeadSource").
		Preload("LeadStatus").
		Preload("ContactRole").
		Preload("AssignedEmployee").
		Preload("Customer").
		Preload("Contact").
		Preload("BusinessType").
		Preload("Area").
		Preload("Deal").
		Preload("Deal.PipelineStage").
		Preload("ProductItems").
		Preload("ProductItems.Product").
		Preload("Activities", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(50)
		}).
		Preload("Activities.ActivityType").
		Preload("Activities.Employee").
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Order("CASE WHEN status IN ('pending','in_progress') THEN 0 ELSE 1 END, due_date ASC NULLS LAST").Limit(20)
		}).
		Preload("Tasks.AssignedEmployee").
		First(&lead, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

// FindByEmail looks up an unconverted lead by email
func (r *leadRepository) FindByEmail(ctx context.Context, email string) (*models.Lead, error) {
	var lead models.Lead
	err := r.scopedLeadQuery(ctx).
		Preload("LeadSource").
		Preload("LeadStatus").
		Preload("ContactRole").
		Preload("AssignedEmployee").
		Where("email = ? AND converted_at IS NULL", email).
		First(&lead).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

// FindDuplicate looks up an unconverted lead by either place_id, cid, email, phone, or company name for deduplication during upsert
func (r *leadRepository) FindDuplicate(ctx context.Context, email, phone, companyName, placeID, cid string) (*models.Lead, error) {
	var lead models.Lead
	query := r.scopedLeadQuery(ctx).
		Preload("LeadSource").
		Preload("LeadStatus").
		Preload("ContactRole").
		Preload("AssignedEmployee").
		Where("converted_at IS NULL")

	if placeID != "" {
		query = query.Where("place_id = ?", placeID)
	} else if cid != "" {
		query = query.Where("cid = ?", cid)
	} else if email != "" {
		query = query.Where("email = ?", email)
	} else if phone != "" {
		query = query.Where("phone = ?", phone)
	} else if companyName != "" && !strings.EqualFold(companyName, "Unknown Company") && !strings.EqualFold(companyName, "N/A") {
		query = query.Where("company_name = ?", companyName)
	} else {
		// Nothing to match against
		return nil, gorm.ErrRecordNotFound
	}

	err := query.First(&lead).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

// FindUnprocessed retrieves leads that haven't been processed by n8n yet
// ordered by creation date to process in FIFO order
func (r *leadRepository) FindUnprocessed(ctx context.Context, limit int) ([]models.Lead, error) {
	var leads []models.Lead
	err := r.scopedLeadQuery(ctx).
		Preload("LeadSource").
		Preload("LeadStatus").
		Where("processed_from_n8n = ?", false).
		Order("created_at ASC").
		Limit(limit).
		Find(&leads).Error
	return leads, err
}

func (r *leadRepository) List(ctx context.Context, params LeadListParams) ([]models.Lead, int64, error) {
	query := r.scopedLeadQuery(ctx)

	// Search filter (prefix search for indexed columns)
	if params.Search != "" {
		search := strings.TrimSpace(params.Search)
		if search != "" {
			searchTerm := "%" + search + "%"
			query = query.Where(
				`(
					first_name ILIKE ? OR
					last_name ILIKE ? OR
					company_name ILIKE ? OR
					CONCAT_WS(' ', first_name, last_name) ILIKE ? OR
					code ILIKE ? OR
					email ILIKE ?
				)`,
				searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
			)
		}
	}

	// Status filter
	if params.LeadStatusID != "" {
		query = query.Where("lead_status_id = ?", params.LeadStatusID)
	}

	// Source filter
	if params.LeadSourceID != "" {
		query = query.Where("lead_source_id = ?", params.LeadSourceID)
	}

	// Assignment filter
	if params.AssignedTo != "" {
		query = query.Where("assigned_to = ?", params.AssignedTo)
	}

	// Score range filter
	if params.ScoreMin != nil {
		query = query.Where("lead_score >= ?", *params.ScoreMin)
	}
	if params.ScoreMax != nil {
		query = query.Where("lead_score <= ?", *params.ScoreMax)
	}

	// Date range filter
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo+" 23:59:59")
	}

	// Conversion filter
	if params.IsConverted != nil {
		if *params.IsConverted {
			query = query.Where("converted_at IS NOT NULL")
		} else {
			query = query.Where("converted_at IS NULL")
		}
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	sortBy := "created_at"
	sortDir := "DESC"
	allowedSorts := map[string]bool{
		"code": true, "first_name": true, "last_name": true, "company_name": true,
		"lead_score": true, "probability": true, "estimated_value": true,
		"created_at": true, "updated_at": true,
	}
	if params.SortBy != "" && allowedSorts[params.SortBy] {
		sortBy = params.SortBy
	}
	if params.SortDir != "" && (strings.EqualFold(params.SortDir, "ASC") || strings.EqualFold(params.SortDir, "DESC")) {
		sortDir = strings.ToUpper(params.SortDir)
	}

	var leads []models.Lead
	err := query.
		Preload("LeadSource").
		Preload("LeadStatus").
		Preload("ContactRole").
		Preload("AssignedEmployee").
		Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: sortDir == "DESC"}).
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&leads).Error

	return leads, total, err
}

func (r *leadRepository) Update(ctx context.Context, lead *models.Lead) error {
	var existing models.Lead
	if err := r.scopedLeadQuery(ctx).Select("id").Where("id = ?", lead.ID).First(&existing).Error; err != nil {
		return err
	}
	return database.GetDB(ctx, r.db).Save(lead).Error
}

func (r *leadRepository) Delete(ctx context.Context, id string) error {
	return r.scopedLeadQuery(ctx).Where("id = ?", id).Delete(&models.Lead{}).Error
}

func (r *leadRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var count int64
	err := database.GetDB(ctx, r.db).Model(&models.Lead{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

func (r *leadRepository) GetAnalytics(ctx context.Context) (*LeadAnalytics, error) {
	analytics := &LeadAnalytics{}
	baseQuery := r.scopedLeadQuery(ctx)

	// Total leads
	if err := baseQuery.Count(&analytics.TotalLeads).Error; err != nil {
		return nil, err
	}

	// By status
	if err := r.scopedLeadQuery(ctx).
		Table("crm_leads").
		Select("crm_lead_statuses.id, crm_lead_statuses.name, crm_lead_statuses.color, COUNT(*) as count").
		Joins("LEFT JOIN crm_lead_statuses ON crm_leads.lead_status_id = crm_lead_statuses.id").
		Where("crm_leads.deleted_at IS NULL").
		Group("crm_lead_statuses.id, crm_lead_statuses.name, crm_lead_statuses.color").
		Scan(&analytics.ByStatus).Error; err != nil {
		return nil, err
	}

	// By source
	if err := r.scopedLeadQuery(ctx).
		Table("crm_leads").
		Select("crm_lead_sources.id, crm_lead_sources.name, COUNT(*) as count").
		Joins("LEFT JOIN crm_lead_sources ON crm_leads.lead_source_id = crm_lead_sources.id").
		Where("crm_leads.deleted_at IS NULL").
		Group("crm_lead_sources.id, crm_lead_sources.name").
		Scan(&analytics.BySource).Error; err != nil {
		return nil, err
	}

	// Conversion rate
	var convertedCount int64
	if err := r.scopedLeadQuery(ctx).Where("converted_at IS NOT NULL").Count(&convertedCount).Error; err != nil {
		return nil, err
	}
	if analytics.TotalLeads > 0 {
		analytics.ConversionRate = float64(convertedCount) / float64(analytics.TotalLeads) * 100
	}

	// Average score
	if err := r.scopedLeadQuery(ctx).Select("COALESCE(AVG(lead_score), 0)").Scan(&analytics.AvgScore).Error; err != nil {
		return nil, err
	}

	return analytics, nil
}

func (r *leadRepository) ListProductItems(ctx context.Context, leadID string) ([]models.LeadProductItem, error) {
	if _, err := r.FindByID(ctx, leadID); err != nil {
		return nil, err
	}

	var items []models.LeadProductItem
	err := r.db.Unscoped().WithContext(ctx).
		Preload("Product").
		Where("lead_id = ?", leadID).
		Order("created_at ASC").
		Find(&items).Error
	return items, err
}

// UpsertProductItems syncs product items for a lead using soft-delete semantics:
// items in the new list are created or restored+updated; items absent from the new list are soft-deleted.
func (r *leadRepository) UpsertProductItems(ctx context.Context, leadID string, items []models.LeadProductItem) error {
	if _, err := r.FindByID(ctx, leadID); err != nil {
		return err
	}

	return database.GetDB(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// Collect product IDs that are active in the new list
		activeProductIDs := make([]string, 0, len(items))
		for _, item := range items {
			if item.ProductID != nil {
				activeProductIDs = append(activeProductIDs, *item.ProductID)
			}
		}

		// Soft-delete items NOT in the new list
		q := tx.Where("lead_id = ?", leadID)
		if len(activeProductIDs) > 0 {
			q = q.Where("product_id NOT IN ?", activeProductIDs)
		}
		if err := q.Delete(&models.LeadProductItem{}).Error; err != nil {
			return err
		}

		if len(items) == 0 {
			return nil
		}

		// Upsert items in the new list: restore soft-deleted ones if they have an ID, else create
		for i := range items {
			items[i].LeadID = leadID
			items[i].DeletedAt = gorm.DeletedAt{} // ensure restored
			if items[i].ID != "" {
				if err := tx.Unscoped().Save(&items[i]).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Create(&items[i]).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
