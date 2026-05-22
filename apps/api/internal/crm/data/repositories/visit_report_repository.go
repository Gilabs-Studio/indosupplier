package repositories

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/crm/data/models"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	visitQueryByID            = "id = ?"
	visitQueryByVisitReportID = "visit_report_id = ?"
)

// VisitReportListParams defines filtering/sorting/pagination for visit report queries
type VisitReportListParams struct {
	Search            string
	SortBy            string
	SortDir           string
	Limit             int
	Offset            int
	CustomerID        string
	EmployeeID        string
	ContactID         string
	DealID            string
	LeadID            string
	TravelPlanID      string
	WithoutTravelPlan bool
	Outcome           string
	DateFrom          string
	DateTo            string
}

// VisitReportRepository defines data access methods for visit reports
type VisitReportRepository interface {
	FindByID(ctx context.Context, id string) (*models.VisitReport, error)
	FindByCode(ctx context.Context, code string) (*models.VisitReport, error)
	List(ctx context.Context, params *VisitReportListParams) ([]models.VisitReport, int64, error)
	Create(ctx context.Context, report *models.VisitReport) error
	Update(ctx context.Context, report *models.VisitReport) error
	Delete(ctx context.Context, id string) error
	GetNextCode(ctx context.Context) (string, error)
	CheckIn(ctx context.Context, id string, location string, checkInAt time.Time) error
	CheckOut(ctx context.Context, id string, location string, checkOutAt time.Time) error
	UpdateStatus(ctx context.Context, id string, status models.VisitReportStatus) error
	CreateProgressHistory(ctx context.Context, history *models.VisitReportProgressHistory) error
	ListProgressHistory(ctx context.Context, visitReportID string, limit, offset int) ([]models.VisitReportProgressHistory, int64, error)

	ListInterestQuestions(ctx context.Context) ([]salesModels.SalesVisitInterestQuestion, error)
	UpdatePhotos(ctx context.Context, id string, photos string) error
	// GetEmployeeSummary returns per-employee visit report counts and latest visit date, scope-filtered.
	GetEmployeeSummary(ctx context.Context, search string, limit, offset int) ([]EmployeeVisitSummary, int64, error)
}

// EmployeeVisitSummary is a raw aggregation result per employee
type EmployeeVisitSummary struct {
	EmployeeID   string
	EmployeeCode string
	EmployeeName string
	TotalReports int64
	LatestVisit  string
	Draft        int64
	Submitted    int64
	Approved     int64
	Rejected     int64
}

type visitReportRepository struct {
	db *gorm.DB
}

// NewVisitReportRepository creates a new visit report repository
func NewVisitReportRepository(db *gorm.DB) VisitReportRepository {
	return &visitReportRepository{db: db}
}

func (r *visitReportRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *visitReportRepository) FindByID(ctx context.Context, id string) (*models.VisitReport, error) {
	var report models.VisitReport
	err := r.getDB(ctx).
		Preload("Employee").
		Preload("Customer").
		Preload("Contact").
		Preload("Deal").
		Preload("Lead").
		Preload("Village.District.City.Province").
		Preload("Details.Product").
		Preload("Details.Answers.Question").
		Preload("Details.Answers.Option").
		Where(visitQueryByID, id).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *visitReportRepository) FindByCode(ctx context.Context, code string) (*models.VisitReport, error) {
	var report models.VisitReport
	err := r.getDB(ctx).
		Preload("Employee").
		Preload("Customer").
		Preload("Contact").
		Where("code = ?", code).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *visitReportRepository) List(ctx context.Context, params *VisitReportListParams) ([]models.VisitReport, int64, error) {
	var reports []models.VisitReport
	var total int64

	query := r.getDB(ctx).Model(&models.VisitReport{})

	// Scope-based data filtering
	query = security.ApplyScopeFilter(query, ctx, security.HRDScopeQueryOptions())

	// Search filter
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("code ILIKE ? OR contact_person ILIKE ? OR purpose ILIKE ? OR notes ILIKE ?",
			search, search, search, search)
	}

	// Customer filter
	if params.CustomerID != "" {
		query = query.Where("customer_id = ?", params.CustomerID)
	}

	// Employee filter
	if params.EmployeeID != "" {
		query = query.Where("employee_id = ?", params.EmployeeID)
	}

	// Contact filter
	if params.ContactID != "" {
		query = query.Where("contact_id = ?", params.ContactID)
	}

	// Deal filter
	if params.DealID != "" {
		query = query.Where("deal_id = ?", params.DealID)
	}

	// Lead filter
	if params.LeadID != "" {
		query = query.Where("lead_id = ?", params.LeadID)
	}

	// Travel plan linkage filter
	if params.WithoutTravelPlan {
		query = query.Where("travel_plan_id IS NULL")
	}
	if params.TravelPlanID != "" {
		query = query.Where("travel_plan_id = ?", params.TravelPlanID)
	}

	// Outcome filter
	if params.Outcome != "" {
		query = query.Where("outcome = ?", params.Outcome)
	}

	// Date range filter
	if params.DateFrom != "" {
		query = query.Where("visit_date >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("visit_date <= ?", params.DateTo)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "visit_date"
	}
	sortDir := params.SortDir
	if sortDir == "" {
		sortDir = "desc"
	}

	query = query.Order(clause.OrderByColumn{
		Column: clause.Column{Name: sortBy},
		Desc:   sortDir == "desc",
	})

	// Pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Execute with preloads
	err := query.
		Preload("Employee").
		Preload("Customer").
		Preload("Contact").
		Find(&reports).Error
	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (r *visitReportRepository) Create(ctx context.Context, report *models.VisitReport) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		details := report.Details
		report.Details = nil

		if err := tx.Create(report).Error; err != nil {
			return err
		}

		// Create details with answers
		if len(details) > 0 {
			for i := range details {
				details[i].VisitReportID = report.ID
				details[i].TenantID = report.TenantID
				answers := details[i].Answers
				details[i].Answers = nil
				if err := tx.Create(&details[i]).Error; err != nil {
					return err
				}
				// Create answers for this detail
				if len(answers) > 0 {
					for j := range answers {
						answers[j].VisitReportDetailID = details[i].ID
						answers[j].TenantID = report.TenantID
					}
					if err := tx.Create(&answers).Error; err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

func (r *visitReportRepository) Update(ctx context.Context, report *models.VisitReport) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Details", "ProgressHistory").Save(report).Error; err != nil {
			return err
		}

		// Replace details
		if err := tx.Where(visitQueryByVisitReportID, report.ID).Delete(&models.VisitReportInterestAnswer{}).Error; err != nil {
			// Answers are nested under details, delete via detail IDs
		}

		// Delete existing detail answers first
		var detailIDs []string
		tx.Model(&models.VisitReportDetail{}).Where(visitQueryByVisitReportID, report.ID).Pluck("id", &detailIDs)
		if len(detailIDs) > 0 {
			tx.Where("visit_report_detail_id IN ?", detailIDs).Delete(&models.VisitReportInterestAnswer{})
		}

		// Delete existing details
		if err := tx.Where(visitQueryByVisitReportID, report.ID).Delete(&models.VisitReportDetail{}).Error; err != nil {
			return err
		}

		// Re-create details
		if len(report.Details) > 0 {
			for i := range report.Details {
				report.Details[i].VisitReportID = report.ID
				report.Details[i].TenantID = report.TenantID
				answers := report.Details[i].Answers
				report.Details[i].Answers = nil
				report.Details[i].CreatedAt = apptime.Now()
				report.Details[i].UpdatedAt = apptime.Now()
				if err := tx.Create(&report.Details[i]).Error; err != nil {
					return err
				}
				if len(answers) > 0 {
					for j := range answers {
						answers[j].VisitReportDetailID = report.Details[i].ID
						answers[j].TenantID = report.TenantID
					}
					if err := tx.Create(&answers).Error; err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

func (r *visitReportRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete answers via detail IDs
		var detailIDs []string
		tx.Model(&models.VisitReportDetail{}).Where(visitQueryByVisitReportID, id).Pluck("id", &detailIDs)
		if len(detailIDs) > 0 {
			tx.Where("visit_report_detail_id IN ?", detailIDs).Delete(&models.VisitReportInterestAnswer{})
		}

		// Delete details
		if err := tx.Where(visitQueryByVisitReportID, id).Delete(&models.VisitReportDetail{}).Error; err != nil {
			return err
		}

		// Soft delete the report
		return tx.Delete(&models.VisitReport{}, visitQueryByID, id).Error
	})
}

func (r *visitReportRepository) GetNextCode(ctx context.Context) (string, error) {
	now := r.getDB(ctx).NowFunc()
	prefix := fmt.Sprintf("VISIT-%s-", now.Format("200601"))

	var lastCode string
	r.getDB(ctx).Model(&models.VisitReport{}).
		Where("code LIKE ?", prefix+"%").
		Order("code DESC").
		Limit(1).
		Pluck("code", &lastCode)

	seq := 1
	if lastCode != "" && len(lastCode) > len(prefix) {
		suffix := lastCode[len(prefix):]
		if n, err := strconv.Atoi(suffix); err == nil {
			seq = n + 1
		}
	}

	return fmt.Sprintf("%s%05d", prefix, seq), nil
}

func (r *visitReportRepository) CheckIn(ctx context.Context, id string, location string, checkInAt time.Time) error {
	updates := map[string]interface{}{
		"check_in_at":       checkInAt,
		"check_in_location": location,
		"actual_time":       checkInAt,
	}
	return r.getDB(ctx).Model(&models.VisitReport{}).
		Where(visitQueryByID, id).
		Updates(updates).Error
}

func (r *visitReportRepository) CheckOut(ctx context.Context, id string, location string, checkOutAt time.Time) error {
	updates := map[string]interface{}{
		"check_out_at":       checkOutAt,
		"check_out_location": location,
	}
	return r.getDB(ctx).Model(&models.VisitReport{}).
		Where(visitQueryByID, id).
		Updates(updates).Error
}

func (r *visitReportRepository) UpdateStatus(ctx context.Context, id string, status models.VisitReportStatus) error {
	return r.getDB(ctx).Model(&models.VisitReport{}).
		Where(visitQueryByID, id).
		Update("status", status).Error
}

func (r *visitReportRepository) CreateProgressHistory(ctx context.Context, history *models.VisitReportProgressHistory) error {
	return r.getDB(ctx).Create(history).Error
}

func (r *visitReportRepository) ListProgressHistory(ctx context.Context, visitReportID string, limit, offset int) ([]models.VisitReportProgressHistory, int64, error) {
	var history []models.VisitReportProgressHistory
	var total int64

	if err := r.getDB(ctx).Model(&models.VisitReportProgressHistory{}).
		Where(visitQueryByVisitReportID, visitReportID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.getDB(ctx).
		Where(visitQueryByVisitReportID, visitReportID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).Error
	if err != nil {
		return nil, 0, err
	}

	return history, total, nil
}

func (r *visitReportRepository) ListInterestQuestions(ctx context.Context) ([]salesModels.SalesVisitInterestQuestion, error) {
	var questions []salesModels.SalesVisitInterestQuestion
	err := r.getDB(ctx).
		Preload("Options").
		Where("is_active = ?", true).
		Order("sequence ASC").
		Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *visitReportRepository) UpdatePhotos(ctx context.Context, id string, photos string) error {
	return r.getDB(ctx).Model(&models.VisitReport{}).
		Where(visitQueryByID, id).
		Update("photos", photos).Error
}

// GetEmployeeSummary aggregates visit report counts per employee.
// RBAC scope is read directly from ctx to maintain consistency with HRD module scope rules.
func (r *visitReportRepository) GetEmployeeSummary(ctx context.Context, search string, limit, offset int) ([]EmployeeVisitSummary, int64, error) {
	db := r.getDB(ctx)

	// Read RBAC scope values injected by middleware
	scope, _ := ctx.Value("permission_scope").(string)
	employeeID, _ := ctx.Value("scope_employee_id").(string)
	divisionID, _ := ctx.Value("scope_division_id").(string)

	// Build scope clause that restricts which employees are visible
	scopeClause := ""
	var scopeArgs []interface{}
	switch scope {
	case "OWN":
		if employeeID != "" {
			scopeClause = "AND vr_scoped.employee_id = ?"
			scopeArgs = append(scopeArgs, employeeID)
		}
	case "DIVISION", "AREA":
		// HRD module uses division_id for both DIVISION and AREA scopes
		if divisionID != "" {
			scopeClause = "AND vr_scoped.employee_id IN (SELECT id FROM employees WHERE division_id = ? AND deleted_at IS NULL)"
			scopeArgs = append(scopeArgs, divisionID)
		} else if employeeID != "" {
			// Fallback to OWN when no division is set
			scopeClause = "AND vr_scoped.employee_id = ?"
			scopeArgs = append(scopeArgs, employeeID)
		}
		// default (ALL or empty): no restriction
	}

	// Employee name/code search predicate
	searchClause := ""
	var searchArgs []interface{}
	if search != "" {
		s := search + "%"
		searchClause = "AND (e.name ILIKE ? OR e.employee_code ILIKE ?)"
		searchArgs = append(searchArgs, s, s)
	}

	// Count distinct employees who have at least one visit report under the scope
	countSQL := fmt.Sprintf(`
		SELECT COUNT(DISTINCT e.id)
		FROM employees e
		INNER JOIN (
			SELECT DISTINCT employee_id FROM crm_visit_reports
			WHERE deleted_at IS NULL %s
		) vr_scoped ON vr_scoped.employee_id = e.id
		WHERE e.deleted_at IS NULL %s
	`, scopeClause, searchClause)

	countArgs := append(scopeArgs, searchArgs...)
	var total int64
	if err := db.Raw(countSQL, countArgs...).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("GetEmployeeSummary count: %w", err)
	}

	if total == 0 {
		return []EmployeeVisitSummary{}, 0, nil
	}

	// Aggregation query: per-employee status counts + latest visit date
	dataSQL := fmt.Sprintf(`
		SELECT
			e.id                                                                      AS employee_id,
			e.employee_code                                                           AS employee_code,
			e.name                                                                    AS employee_name,
			COUNT(vr.id)                                                              AS total_reports,
			COALESCE(MAX(vr.visit_date::text), '')                                    AS latest_visit,
			COUNT(vr.id) FILTER (WHERE vr.status = 'draft')                          AS draft,
			COUNT(vr.id) FILTER (WHERE vr.status = 'submitted')                      AS submitted,
			COUNT(vr.id) FILTER (WHERE vr.status = 'approved')                       AS approved,
			COUNT(vr.id) FILTER (WHERE vr.status = 'rejected')                       AS rejected
		FROM employees e
		INNER JOIN (
			SELECT DISTINCT employee_id FROM crm_visit_reports
			WHERE deleted_at IS NULL %s
		) vr_scoped ON vr_scoped.employee_id = e.id
		LEFT JOIN crm_visit_reports vr
			ON vr.employee_id = e.id AND vr.deleted_at IS NULL
		WHERE e.deleted_at IS NULL %s
		GROUP BY e.id, e.employee_code, e.name
		ORDER BY total_reports DESC, e.name ASC
		LIMIT ? OFFSET ?
	`, scopeClause, searchClause)

	dataArgs := append(append(scopeArgs, searchArgs...), limit, offset)

	var results []EmployeeVisitSummary
	if err := db.Raw(dataSQL, dataArgs...).Scan(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("GetEmployeeSummary data: %w", err)
	}

	return results, total, nil
}
