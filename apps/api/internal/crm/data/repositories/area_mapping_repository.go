package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/crm/domain/dto"
	"gorm.io/gorm"
)

// AreaMappingRepository defines operations for area mapping data
type AreaMappingRepository interface {
	// GetLeadsForMapping returns all leads with their activity metrics
	GetLeadsForMapping(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]dto.AreaMappingLeadData, error)

	// GetPipelinesForMapping returns all pipeline deals with mappable locations
	GetPipelinesForMapping(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]dto.AreaMappingPipelineData, error)
}

// NewAreaMappingRepository creates a new area mapping repository
func NewAreaMappingRepository(db *gorm.DB) AreaMappingRepository {
	return &areaMappingRepository{db: db}
}

type areaMappingRepository struct {
	db *gorm.DB
}

func buildInPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}

	placeholders := "?"
	for i := 1; i < count; i++ {
		placeholders += ", ?"
	}

	return placeholders
}

func appendStringsAsArgs(args []interface{}, values []string) []interface{} {
	for _, value := range values {
		args = append(args, value)
	}
	return args
}

func buildCRMScopeFilter(ctx context.Context, tableAlias string) (string, []interface{}) {
	scope, _ := ctx.Value("permission_scope").(string)
	if scope == "" || scope == "ALL" {
		return "", nil
	}

	userID, _ := ctx.Value("scope_user_id").(string)
	employeeID, _ := ctx.Value("scope_employee_id").(string)
	divisionID, _ := ctx.Value("scope_division_id").(string)
	areaIDs, _ := ctx.Value("scope_area_ids").([]string)

	ownerClause := func() (string, []interface{}) {
		conditions := ""
		args := make([]interface{}, 0, 2)

		if employeeID != "" {
			conditions += tableAlias + ".assigned_to = ?"
			args = append(args, employeeID)
		}
		if userID != "" {
			if conditions != "" {
				conditions += " OR "
			}
			conditions += tableAlias + ".created_by = ?"
			args = append(args, userID)
		}

		if conditions == "" {
			return " AND 1 = 0", nil
		}

		return " AND (" + conditions + ")", args
	}

	switch scope {
	case "OWN":
		return ownerClause()
	case "DIVISION":
		if divisionID == "" {
			return ownerClause()
		}

		clause := " AND ((" + tableAlias + ".assigned_to IN (SELECT id FROM employees WHERE division_id = ? AND deleted_at IS NULL)) OR (" + tableAlias + ".created_by IN (SELECT user_id FROM employees WHERE division_id = ? AND user_id IS NOT NULL AND deleted_at IS NULL)))"
		return clause, []interface{}{divisionID, divisionID}
	case "AREA":
		if len(areaIDs) == 0 {
			return ownerClause()
		}

		placeholders := buildInPlaceholders(len(areaIDs))
		clause := " AND ((" + tableAlias + ".assigned_to IN (SELECT employee_id FROM employee_areas WHERE area_id IN (" + placeholders + ") AND deleted_at IS NULL)) OR (" + tableAlias + ".created_by IN (SELECT user_id FROM employees WHERE id IN (SELECT employee_id FROM employee_areas WHERE area_id IN (" + placeholders + ") AND deleted_at IS NULL) AND user_id IS NOT NULL AND deleted_at IS NULL)))"

		args := make([]interface{}, 0, len(areaIDs)*2)
		args = appendStringsAsArgs(args, areaIDs)
		args = appendStringsAsArgs(args, areaIDs)
		return clause, args
	default:
		return ownerClause()
	}
}

func resolveTenantScope(ctx context.Context) (tenantID string, enabled bool) {
	if v := ctx.Value("is_system_admin"); v != nil {
		if isSystemAdmin, ok := v.(bool); ok && isSystemAdmin {
			return "", false
		}
	}

	if v := ctx.Value("tenant_id"); v != nil {
		if tid, ok := v.(string); ok && tid != "" {
			return tid, true
		}
	}

	return "", false
}

func (r *areaMappingRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

// GetLeadsForMapping returns potential customers (leads) with activity metrics.
func (r *areaMappingRepository) GetLeadsForMapping(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]dto.AreaMappingLeadData, error) {
	var results []dto.AreaMappingLeadData

	activityJoin := "LEFT JOIN crm_activities a ON l.id = a.lead_id"
	taskJoin := "LEFT JOIN crm_tasks t ON l.id = t.lead_id AND t.deleted_at IS NULL"
	dealJoin := "LEFT JOIN crm_deals d ON l.id = d.lead_id AND d.deleted_at IS NULL"
	tenantWhereFilter := ""
	scopeWhereFilter := ""
	whereDateFilter := ""
	args := make([]interface{}, 0)
	tenantID, tenantScoped := resolveTenantScope(ctx)
	scopeClause, scopeArgs := buildCRMScopeFilter(ctx, "l")

	if tenantScoped {
		activityJoin += " AND a.tenant_id = l.tenant_id"
		taskJoin += " AND t.tenant_id = l.tenant_id"
		dealJoin += " AND d.tenant_id = l.tenant_id"
		tenantWhereFilter = " AND l.tenant_id = ?"
		args = append(args, tenantID)
	}

	if scopeClause != "" {
		scopeWhereFilter = scopeClause
		args = append(args, scopeArgs...)
	}

	if startDate != nil && endDate != nil {
		activityJoin += " AND a.created_at >= ? AND a.created_at < ?"
		args = append(args, *startDate, *endDate)
		taskJoin += " AND t.created_at >= ? AND t.created_at < ?"
		args = append(args, *startDate, *endDate)
		dealJoin += " AND d.created_at >= ? AND d.created_at < ?"
		args = append(args, *startDate, *endDate)
		whereDateFilter = " AND l.created_at >= ? AND l.created_at < ?"
		args = append(args, *startDate, *endDate)
	}

	query := fmt.Sprintf(`
		SELECT
			l.id,
			l.code,
			CONCAT(l.first_name, ' ', COALESCE(l.last_name, '')) AS name,
			'lead' AS type,
			l.latitude,
			l.longitude,
			COALESCE(l.province, '') AS province,
			COALESCE(l.city, '') AS city,
			COALESCE(ls.name, '') AS lead_status,
			COALESCE(l.lead_score, 0) AS lead_score,
			COALESCE(l.estimated_value, 0) AS estimated_value,
			l.assigned_to,
			e.name AS assigned_name,
			COALESCE(COUNT(DISTINCT a.id), 0) AS activity_count,
			COALESCE(COUNT(DISTINCT t.id), 0) AS task_count,
			COALESCE(COUNT(DISTINCT d.id), 0) AS pipeline_deal_count,
			MAX(a.created_at) AS last_activity_at,
			CASE
				WHEN (COALESCE(COUNT(DISTINCT a.id), 0) + COALESCE(COUNT(DISTINCT t.id), 0) + COALESCE(COUNT(DISTINCT d.id), 0) + COALESCE(l.lead_score, 0)) = 0 THEN 10
				ELSE LEAST(100, (COALESCE(COUNT(DISTINCT a.id), 0) * 5 + COALESCE(COUNT(DISTINCT t.id), 0) * 3 + COALESCE(COUNT(DISTINCT d.id), 0) * 10 + COALESCE(l.lead_score, 0) / 2))
			END AS intensity_score
		FROM crm_leads l
		LEFT JOIN crm_lead_statuses ls ON l.lead_status_id = ls.id AND (ls.tenant_id = l.tenant_id OR ls.tenant_id IS NULL)
		LEFT JOIN employees e ON l.assigned_to = e.id AND e.tenant_id = l.tenant_id
		%s
		%s
		%s
		WHERE l.deleted_at IS NULL
			AND l.converted_at IS NULL
			AND l.latitude IS NOT NULL
			AND l.longitude IS NOT NULL
			%s
			%s
			%s
		GROUP BY l.id, l.code, l.first_name, l.last_name, l.latitude, l.longitude, l.province, l.city, ls.name, l.lead_score, l.estimated_value, l.assigned_to, e.name
		ORDER BY intensity_score DESC
	`, activityJoin, taskJoin, dealJoin, tenantWhereFilter, scopeWhereFilter, whereDateFilter)

	if err := r.getDB(ctx).Raw(query, args...).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// GetPipelinesForMapping returns pipeline deals with location from lead/customer.
func (r *areaMappingRepository) GetPipelinesForMapping(ctx context.Context, startDate *time.Time, endDate *time.Time) ([]dto.AreaMappingPipelineData, error) {
	var results []dto.AreaMappingPipelineData

	tenantWhereFilter := ""
	scopeWhereFilter := ""
	whereDateFilter := ""
	args := make([]interface{}, 0)
	tenantID, tenantScoped := resolveTenantScope(ctx)
	scopeClause, scopeArgs := buildCRMScopeFilter(ctx, "d")
	if tenantScoped {
		tenantWhereFilter = " AND d.tenant_id = ?"
		args = append(args, tenantID)
	}

	if scopeClause != "" {
		scopeWhereFilter = scopeClause
		args = append(args, scopeArgs...)
	}

	if startDate != nil && endDate != nil {
		whereDateFilter = " AND d.created_at >= ? AND d.created_at < ?"
		args = append(args, *startDate, *endDate)
	}

	query := fmt.Sprintf(`
		SELECT
			d.id,
			d.code,
			d.title,
			'pipeline' AS type,
			COALESCE(l.latitude, c.latitude) AS latitude,
			COALESCE(l.longitude, c.longitude) AS longitude,
			COALESCE(NULLIF(l.province, ''), p.name, '') AS province,
			COALESCE(NULLIF(l.city, ''), ci.name, '') AS city,
			d.pipeline_stage_id,
			COALESCE(ps.name, '') AS pipeline_stage_name,
			COALESCE(CAST(d.status AS varchar), '') AS status,
			COALESCE(d.value, 0) AS value,
			COALESCE(d.probability, 0) AS probability,
			d.expected_close_date,
			d.assigned_to,
			e.name AS assigned_name,
			d.lead_id,
			CASE
				WHEN l.id IS NULL THEN NULL
				ELSE CONCAT(l.first_name, ' ', COALESCE(l.last_name, ''))
			END AS lead_name,
			CASE
				WHEN COALESCE(d.value, 0) <= 0 AND COALESCE(d.probability, 0) <= 0 THEN 10
				ELSE LEAST(100, COALESCE(d.probability, 0) + CASE WHEN COALESCE(d.value, 0) > 0 THEN 20 ELSE 0 END)
			END AS intensity_score
		FROM crm_deals d
		LEFT JOIN crm_leads l ON d.lead_id = l.id AND l.deleted_at IS NULL AND l.tenant_id = d.tenant_id
		LEFT JOIN customers c ON d.customer_id = c.id AND c.deleted_at IS NULL AND c.tenant_id = d.tenant_id
		LEFT JOIN crm_pipeline_stages ps ON d.pipeline_stage_id = ps.id AND (ps.tenant_id = d.tenant_id OR ps.tenant_id IS NULL)
		LEFT JOIN employees e ON d.assigned_to = e.id AND e.tenant_id = d.tenant_id
		LEFT JOIN provinces p ON c.province_id = p.id
		LEFT JOIN cities ci ON c.city_id = ci.id
		WHERE d.deleted_at IS NULL
			AND COALESCE(l.latitude, c.latitude) IS NOT NULL
			AND COALESCE(l.longitude, c.longitude) IS NOT NULL
			%s
			%s
			%s
		ORDER BY intensity_score DESC, d.created_at DESC
	`, tenantWhereFilter, scopeWhereFilter, whereDateFilter)

	if err := r.getDB(ctx).Raw(query, args...).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
