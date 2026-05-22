package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SalesVisitRepository defines the interface for sales visit data access
const (
	queryByID           = "id = ?"
	queryBySalesVisitID = "sales_visit_id = ?"
)

type SalesVisitRepository interface {
	FindByID(ctx context.Context, id string) (*models.SalesVisit, error)
	FindByCode(ctx context.Context, code string) (*models.SalesVisit, error)
	List(ctx context.Context, req *dto.ListSalesVisitsRequest) ([]models.SalesVisit, int64, error)
	ListDetails(ctx context.Context, visitID string, req *dto.ListSalesVisitDetailsRequest) ([]models.SalesVisitDetail, int64, error)
	ListProgressHistory(ctx context.Context, visitID string, req *dto.ListSalesVisitProgressHistoryRequest) ([]models.SalesVisitProgressHistory, int64, error)
	Create(ctx context.Context, visit *models.SalesVisit) error
	Update(ctx context.Context, visit *models.SalesVisit) error
	Delete(ctx context.Context, id string) error
	GetNextVisitNumber(ctx context.Context, prefix string) (string, error)
	UpdateStatus(ctx context.Context, id string, status models.SalesVisitStatus, notes string, userID *string) error
	CheckIn(ctx context.Context, id string, latitude, longitude *float64, checkInAt time.Time) error
	CheckOut(ctx context.Context, id string, checkOutAt time.Time, result string) error
	CreateProgressHistory(ctx context.Context, history *models.SalesVisitProgressHistory) error
	GetCalendarSummary(ctx context.Context, req *dto.GetCalendarSummaryRequest) ([]dto.CalendarDaySummary, error)
	ListInterestQuestions(ctx context.Context) ([]models.SalesVisitInterestQuestion, error)
}

type salesVisitRepository struct {
	db *gorm.DB
}

// NewSalesVisitRepository creates a new SalesVisitRepository
func NewSalesVisitRepository(db *gorm.DB) SalesVisitRepository {
	return &salesVisitRepository{db: db}
}

func (r *salesVisitRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *salesVisitRepository) FindByID(ctx context.Context, id string) (*models.SalesVisit, error) {
	var visit models.SalesVisit
	err := r.getDB(ctx).
		Preload("Employee").
		Preload("Company").
		Preload("Village.District.City.Province").
		Preload("Details.Product").
		Preload("Details.Answers.Question").
		Preload("Details.Answers.Option").
		Where(queryByID, id).
		First(&visit).Error
	if err != nil {
		return nil, err
	}
	return &visit, nil
}

func (r *salesVisitRepository) FindByCode(ctx context.Context, code string) (*models.SalesVisit, error) {
	var visit models.SalesVisit
	err := r.getDB(ctx).
		Preload("Employee").
		Preload("Company").
		Preload("Village.District.City.Province").
		Preload("Details.Product").
		Preload("Details.Answers.Question").
		Preload("Details.Answers.Option").
		Where("code = ?", code).
		First(&visit).Error
	if err != nil {
		return nil, err
	}
	return &visit, nil
}

func (r *salesVisitRepository) List(ctx context.Context, req *dto.ListSalesVisitsRequest) ([]models.SalesVisit, int64, error) {
	var visits []models.SalesVisit
	var total int64
	var query *gorm.DB

	// Apply search filter
	if s := strings.TrimSpace(req.Search); s != "" {
		search := "%" + s + "%"
		query = r.db.WithContext(ctx).Model(&models.SalesVisit{}).
			Joins("LEFT JOIN employees ON employees.id = sales_visits.employee_id").
			Joins("LEFT JOIN companies ON companies.id = sales_visits.company_id")
		
		// Apply tenant filter manually since we are using joins
		var err error
		query, err = applyTenantFilter(ctx, query, "sales_visits.tenant_id", "employees.tenant_id", "companies.tenant_id")
		if err != nil {
			return nil, 0, err
		}

		query = query.Where("companies.name ILIKE ? OR employees.name ILIKE ? OR sales_visits.contact_person ILIKE ? OR sales_visits.code ILIKE ? OR sales_visits.purpose ILIKE ? OR sales_visits.notes ILIKE ?", search, search, search, search, search, search)
	} else {
		query = r.db.WithContext(ctx).Model(&models.SalesVisit{})
		var err error
		query, err = applyTenantFilter(ctx, query, "sales_visits.tenant_id")
		if err != nil {
			return nil, 0, err
		}
	}

	// Apply scope-based data filtering (OWN/DIVISION/AREA/ALL)
	query = security.ApplyScopeFilter(query, ctx, security.HRDScopeQueryOptions())

	// Apply status filter
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Apply employee filter
	if req.EmployeeID != "" {
		query = query.Where("employee_id = ?", req.EmployeeID)
	}

	// Apply company filter
	if req.CompanyID != "" {
		query = query.Where("company_id = ?", req.CompanyID)
	}

	// Apply date range filter
	if req.DateFrom != "" {
		query = query.Where("visit_date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("visit_date <= ?", req.DateTo)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	// Apply sorting with whitelist and clause builder to prevent SQL injection
	allowedSortColumns := map[string]string{
		"code":       "code",
		"visit_date": "visit_date",
		"status":     "status",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}

	sortByFromReq := strings.ToLower(strings.TrimSpace(req.SortBy))
	sortBy := allowedSortColumns[sortByFromReq]
	if sortBy == "" {
		sortBy = "visit_date"
	}

	isDesc := strings.ToLower(strings.TrimSpace(req.SortDir)) == "desc"
	query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: sortBy}, Desc: isDesc})

	// Execute query with preloads
	err := query.
		Preload("Employee").
		Preload("Company").
		Limit(perPage).
		Offset(offset).
		Find(&visits).Error
	if err != nil {
		return nil, 0, err
	}

	return visits, total, nil
}

func (r *salesVisitRepository) ListDetails(ctx context.Context, visitID string, req *dto.ListSalesVisitDetailsRequest) ([]models.SalesVisitDetail, int64, error) {
	var details []models.SalesVisitDetail
	var total int64

	// Set defaults
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	// Count total
	if err := r.getDB(ctx).Model(&models.SalesVisitDetail{}).
		Where(queryBySalesVisitID, visitID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated items
	offset := (page - 1) * perPage
	err := r.getDB(ctx).
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "code", "name", "selling_price", "image_url")
		}).
		Preload("Answers.Question").
		Preload("Answers.Option").
		Where(queryBySalesVisitID, visitID).
		Order("created_at ASC").
		Limit(perPage).
		Offset(offset).
		Find(&details).Error

	if err != nil {
		return nil, 0, err
	}

	return details, total, nil
}

func (r *salesVisitRepository) ListProgressHistory(ctx context.Context, visitID string, req *dto.ListSalesVisitProgressHistoryRequest) ([]models.SalesVisitProgressHistory, int64, error) {
	var history []models.SalesVisitProgressHistory
	var total int64

	// Set defaults
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	// Count total
	if err := r.getDB(ctx).Model(&models.SalesVisitProgressHistory{}).
		Where(queryBySalesVisitID, visitID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated history
	offset := (page - 1) * perPage
	err := r.getDB(ctx).
		Where(queryBySalesVisitID, visitID).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&history).Error

	if err != nil {
		return nil, 0, err
	}

	return history, total, nil
}

func (r *salesVisitRepository) Create(ctx context.Context, visit *models.SalesVisit) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Store details temporarily
		details := visit.Details
		visit.Details = nil

		// Create visit without details
		if err := tx.Create(visit).Error; err != nil {
			return err
		}

		// Create details with the visit ID
		if len(details) > 0 {
			for i := range details {
				details[i].SalesVisitID = visit.ID
				if err := tx.Create(&details[i]).Error; err != nil {
					return err
				}
			}
		}

		// Create initial progress history
		initialHistory := models.SalesVisitProgressHistory{
			SalesVisitID: visit.ID,
			FromStatus:   "",
			ToStatus:     visit.Status,
			Notes:        "Visit created",
			ChangedBy:    visit.CreatedBy,
			CreatedAt:    apptime.Now(),
		}
		if err := tx.Create(&initialHistory).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *salesVisitRepository) Update(ctx context.Context, visit *models.SalesVisit) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Update visit WITHOUT associations
		if err := tx.Omit("Details", "ProgressHistory").Save(visit).Error; err != nil {
			return err
		}

		// Delete existing details
		if err := tx.Where(queryBySalesVisitID, visit.ID).Delete(&models.SalesVisitDetail{}).Error; err != nil {
			return err
		}

		// Create new details
		if len(visit.Details) > 0 {
			for i := range visit.Details {
				visit.Details[i].SalesVisitID = visit.ID
				visit.Details[i].CreatedAt = apptime.Now()
				visit.Details[i].UpdatedAt = apptime.Now()
				if err := tx.Create(&visit.Details[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *salesVisitRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete details first
		if err := tx.Where(queryBySalesVisitID, id).Delete(&models.SalesVisitDetail{}).Error; err != nil {
			return err
		}

		// Delete progress history
		if err := tx.Where(queryBySalesVisitID, id).Delete(&models.SalesVisitProgressHistory{}).Error; err != nil {
			return err
		}

		// Delete visit
		return tx.Delete(&models.SalesVisit{}, queryByID, id).Error
	})
}

func (r *salesVisitRepository) GetNextVisitNumber(ctx context.Context, prefix string) (string, error) {
	var lastVisit models.SalesVisit
	var sequence int

	// Find the last visit with the same prefix
	err := r.getDB(ctx).
		Where("code LIKE ?", prefix+"%").
		Order("code DESC").
		First(&lastVisit).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			sequence = 1
		} else {
			return "", err
		}
	} else {
		var count int64
		r.getDB(ctx).Model(&models.SalesVisit{}).
			Where("code LIKE ?", prefix+"%").
			Count(&count)
		sequence = int(count) + 1
	}

	// Generate new code: PREFIX-YYYYMMDD-XXXX (e.g., VIS-20240115-0001)
	now := database.GetDB(ctx, r.db).NowFunc()
	dateStr := now.Format("20060102")
	code := prefix + "-" + dateStr + "-" + fmt.Sprintf("%04d", sequence)

	return code, nil
}

func (r *salesVisitRepository) UpdateStatus(ctx context.Context, id string, status models.SalesVisitStatus, notes string, userID *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.SalesVisitStatusCancelled {
		updates["cancelled_by"] = userID
		updates["cancelled_at"] = database.GetDB(ctx, r.db).NowFunc()
	}

	return r.getDB(ctx).Model(&models.SalesVisit{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *salesVisitRepository) CheckIn(ctx context.Context, id string, latitude, longitude *float64, checkInAt time.Time) error {
	updates := map[string]interface{}{
		"check_in_at": checkInAt,
		"status":      models.SalesVisitStatusInProgress,
		"actual_time": checkInAt.Format("15:04:05"),
	}

	if latitude != nil {
		updates["latitude"] = *latitude
	}
	if longitude != nil {
		updates["longitude"] = *longitude
	}

	return r.getDB(ctx).Model(&models.SalesVisit{}).
		Where(queryByID, id).
		Updates(updates).Error
}

func (r *salesVisitRepository) CheckOut(ctx context.Context, id string, checkOutAt time.Time, result string) error {
	updates := map[string]interface{}{
		"check_out_at": checkOutAt,
		"status":       models.SalesVisitStatusCompleted,
	}

	if result != "" {
		updates["result"] = result
	}

	return r.getDB(ctx).Model(&models.SalesVisit{}).
		Where(queryByID, id).
		Updates(updates).Error
}

func (r *salesVisitRepository) CreateProgressHistory(ctx context.Context, history *models.SalesVisitProgressHistory) error {
	return r.getDB(ctx).Create(history).Error
}

func (r *salesVisitRepository) GetCalendarSummary(ctx context.Context, req *dto.GetCalendarSummaryRequest) ([]dto.CalendarDaySummary, error) {
	var summaries []dto.CalendarDaySummary

	// 1. Get daily counts
	query := r.getDB(ctx).
		Table("sales_visits").
		Select(`
			visit_date::date as date,
			COUNT(*) as total_count,
			SUM(CASE WHEN status = 'planned' THEN 1 ELSE 0 END) as planned,
			SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled
		`).
		Where("visit_date >= ? AND visit_date <= ?", req.DateFrom, req.DateTo)

	if req.EmployeeID != "" {
		query = query.Where("employee_id = ?", req.EmployeeID)
	}

	if req.CompanyID != "" {
		query = query.Where("company_id = ?", req.CompanyID)
	}

	err := query.Group("visit_date::date").Scan(&summaries).Error
	if err != nil {
		return nil, err
	}

	// 2. Get preview items (Top 3 per day)
	// Using CTE and Window Function for efficient fetching per day
	// Priority: In Progress > Upcoming (today/future) > Past
	// Sort by: Status priority, then Scheduled Time
	
	// Note: GORM raw query for complex window function
	rawQuery := `
		WITH prioritized_visits AS (
			SELECT 
				sv.id,
				sv.code,
				sv.visit_date,
				sv.scheduled_time,
				sv.status,
				COALESCE(c.name, sv.contact_person) as customer_name,
				ROW_NUMBER() OVER (
					PARTITION BY sv.visit_date 
					ORDER BY 
						CASE 
							WHEN sv.status = 'in_progress' THEN 1 
							WHEN sv.status = 'planned' AND (sv.visit_date + sv.scheduled_time::time) >= NOW() THEN 2
							WHEN sv.status = 'planned' AND (sv.visit_date + sv.scheduled_time::time) < NOW() THEN 3
							ELSE 4 
						END ASC,
						sv.scheduled_time ASC
				) as rn
			FROM sales_visits sv
			LEFT JOIN companies c ON sv.company_id = c.id
			WHERE sv.tenant_id = ? 
			AND sv.visit_date >= ? AND sv.visit_date <= ?
			AND sv.deleted_at IS NULL
			`
	
	tenantID := ""
	if tid, ok := ctx.Value("tenant_id").(string); ok {
		tenantID = tid
	}
	args := []interface{}{tenantID, req.DateFrom, req.DateTo}

	if req.EmployeeID != "" {
		rawQuery += " AND sv.employee_id = ?"
		args = append(args, req.EmployeeID)
	}
	if req.CompanyID != "" {
		rawQuery += " AND sv.company_id = ?"
		args = append(args, req.CompanyID)
	}

	rawQuery += `
		)
		SELECT 
			visit_date::text as date,
			id,
			code,
			to_char(scheduled_time, 'HH24:MI') as scheduled_time,
			customer_name,
			status
		FROM prioritized_visits
		WHERE rn <= 3
	`

	var previews []struct {
		Date          string
		ID            string
		Code          string
		ScheduledTime string
		CustomerName  string
		Status        string
	}

	if err := r.getDB(ctx).Raw(rawQuery, args...).Scan(&previews).Error; err != nil {
		return nil, err
	}

	// 3. Merge previews into summaries
	// Map previews by date for O(1) lookup
	previewMap := make(map[string][]dto.CalendarPreviewItem)
	for _, p := range previews {
		// Go's time format from DB might include T00:00:00Z, parse just YYYY-MM-DD
		// Actually visit_date::text from Postgres usually returns YYYY-MM-DD
		dateKey := p.Date[0:10] 
		previewMap[dateKey] = append(previewMap[dateKey], dto.CalendarPreviewItem{
			ID:            p.ID,
			Code:          p.Code,
			ScheduledTime: p.ScheduledTime,
			CustomerName:  p.CustomerName,
			Status:        p.Status,
		})
	}

	// Assign to summaries
	for i := range summaries {
		dateKey := summaries[i].Date[0:10]
		if items, ok := previewMap[dateKey]; ok {
			summaries[i].PreviewItems = items
		} else {
			summaries[i].PreviewItems = []dto.CalendarPreviewItem{}
		}
	}

	return summaries, nil
}

func (r *salesVisitRepository) ListInterestQuestions(ctx context.Context) ([]models.SalesVisitInterestQuestion, error) {
	var questions []models.SalesVisitInterestQuestion
	err := r.getDB(ctx).
		Preload("Options").
		Order("sequence ASC").
		Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}
