package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	"github.com/gilabs/gims/api/internal/core/infrastructure/audit"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
	"gorm.io/gorm"
)

type salesAuditRow struct {
	ID             string    `gorm:"column:id"`
	ActorID        string    `gorm:"column:actor_id"`
	PermissionCode string    `gorm:"column:permission_code"`
	TargetID       string    `gorm:"column:target_id"`
	Action         string    `gorm:"column:action"`
	Metadata       string    `gorm:"column:metadata"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	ActorEmail     *string   `gorm:"column:actor_email"`
	ActorName      *string   `gorm:"column:actor_name"`
}

func listAuditTrailEntries(
	ctx context.Context,
	db *gorm.DB,
	targetID string,
	permissionPrefix string,
	page int,
	perPage int,
) ([]dto.CustomerInvoiceAuditTrailEntry, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	tx := db.WithContext(ctx).Model(&coreModels.AuditLog{}).
		Where("audit_logs.target_id = ?", targetID).
		Where("audit_logs.permission_code LIKE ?", permissionPrefix+"%")

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]salesAuditRow, 0)
	if err := tx.
		Select("audit_logs.id, audit_logs.actor_id, audit_logs.permission_code, audit_logs.target_id, audit_logs.action, audit_logs.metadata, audit_logs.created_at, users.email as actor_email, COALESCE(users.name, employees.name, users.email, audit_logs.actor_id::text) as actor_name").
		Joins("LEFT JOIN employees ON employees.id = audit_logs.actor_id OR employees.user_id = audit_logs.actor_id").
		Joins("LEFT JOIN users ON users.id = audit_logs.actor_id OR users.id = employees.user_id").
		Order("audit_logs.created_at DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	entries := make([]dto.CustomerInvoiceAuditTrailEntry, 0, len(rows))
	refCache := make(map[string]string)
	for _, row := range rows {
		metadata := parseAuditMetadata(ctx, db, row.Metadata, refCache)
		user := buildAuditTrailUser(row.ActorID, row.ActorEmail, row.ActorName)

		entries = append(entries, dto.CustomerInvoiceAuditTrailEntry{
			ID:             row.ID,
			Action:         row.Action,
			PermissionCode: row.PermissionCode,
			TargetID:       row.TargetID,
			Metadata:       metadata,
			User:           user,
			CreatedAt:      row.CreatedAt,
		})
	}

	return entries, total, nil
}

func parseAuditMetadata(ctx context.Context, db *gorm.DB, raw string, refCache map[string]string) map[string]interface{} {
	metadata := map[string]interface{}{}
	if strings.TrimSpace(raw) == "" {
		return metadata
	}

	_ = json.Unmarshal([]byte(raw), &metadata)
	enrichAuditMetadataReferences(ctx, db, metadata, refCache)
	return metadata
}

func enrichAuditMetadataReferences(ctx context.Context, db *gorm.DB, metadata map[string]interface{}, refCache map[string]string) {
	enrichSnapshotReferenceNames(ctx, db, metadata, refCache)

	if before, ok := metadata["before"].(map[string]interface{}); ok {
		enrichSnapshotReferenceNames(ctx, db, before, refCache)
	}

	if after, ok := metadata["after"].(map[string]interface{}); ok {
		enrichSnapshotReferenceNames(ctx, db, after, refCache)
	}
}

func enrichSnapshotReferenceNames(ctx context.Context, db *gorm.DB, snapshot map[string]interface{}, refCache map[string]string) {
	type refMap struct {
		idKey   string
		nameKey string
	}

	references := []refMap{
		{idKey: "payment_terms_id", nameKey: "payment_terms_name"},
		{idKey: "supplier_id", nameKey: "supplier_name"},
		{idKey: "customer_id", nameKey: "customer_name"},
		{idKey: "business_unit_id", nameKey: "business_unit_name"},
		{idKey: "business_type_id", nameKey: "business_type_name"},
		{idKey: "delivery_area_id", nameKey: "delivery_area_name"},
		{idKey: "employee_id", nameKey: "employee_name"},
		{idKey: "sales_rep_id", nameKey: "sales_rep_name"},
		{idKey: "sales_quotation_id", nameKey: "sales_quotation_code"},
		{idKey: "purchase_requisitions_id", nameKey: "purchase_requisition_code"},
		{idKey: "sales_order_id", nameKey: "sales_order_code"},
	}

	for _, ref := range references {
		id := stringValue(snapshot[ref.idKey])
		if id == "" {
			continue
		}

		if stringValue(snapshot[ref.nameKey]) != "" {
			continue
		}

		if resolved := lookupReferenceLabel(ctx, db, ref.idKey, id, refCache); resolved != "" {
			snapshot[ref.nameKey] = resolved
		}
	}
}

func stringValue(value interface{}) string {
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(str)
}

func lookupReferenceLabel(ctx context.Context, db *gorm.DB, idKey, id string, refCache map[string]string) string {
	cacheKey := idKey + "|" + id
	if cached, ok := refCache[cacheKey]; ok {
		return cached
	}

	table, column := lookupSourceByReferenceKey(idKey)
	if table == "" || column == "" {
		refCache[cacheKey] = ""
		return ""
	}

	queryDB := db.WithContext(ctx)
	if !isGlobalSalesReferenceTable(table) {
		queryDB = database.GetDB(ctx, db)
	}

	row := queryDB.Table(table).Select(column).Where("id = ?", id).Limit(1).Row()
	var value string
	if err := row.Scan(&value); err != nil {
		refCache[cacheKey] = ""
		return ""
	}

	trimmed := strings.TrimSpace(value)
	refCache[cacheKey] = trimmed
	return trimmed
}

func lookupSourceByReferenceKey(idKey string) (string, string) {
	sources := map[string]struct {
		table  string
		column string
	}{
		"payment_terms_id":         {table: "payment_terms", column: "name"},
		"supplier_id":              {table: "suppliers", column: "name"},
		"customer_id":              {table: "customers", column: "name"},
		"business_unit_id":         {table: "business_units", column: "name"},
		"business_type_id":         {table: "business_types", column: "name"},
		"delivery_area_id":         {table: "areas", column: "name"},
		"employee_id":              {table: "employees", column: "name"},
		"sales_rep_id":             {table: "employees", column: "name"},
		"sales_quotation_id":       {table: "sales_quotations", column: "code"},
		"purchase_requisitions_id": {table: "purchase_requisitions", column: "code"},
		"sales_order_id":           {table: "sales_orders", column: "code"},
	}

	source, ok := sources[idKey]
	if !ok {
		return "", ""
	}

	if strings.TrimSpace(source.table) == "" || strings.TrimSpace(source.column) == "" {
		return "", ""
	}

	if strings.Contains(source.table, " ") || strings.Contains(source.column, " ") {
		return "", ""
	}

	return source.table, source.column
}

func isGlobalSalesReferenceTable(table string) bool {
	switch strings.TrimSpace(table) {
	case "currencies":
		return true
	default:
		return false
	}
}

func buildAuditTrailUser(actorID string, actorEmail, actorName *string) *dto.AuditTrailUser {
	if actorID == "" {
		return nil
	}

	user := &dto.AuditTrailUser{ID: actorID, Email: "", Name: ""}
	if actorEmail != nil {
		user.Email = *actorEmail
	}
	if actorName != nil {
		user.Name = *actorName
	}
	if user.Name == "" && user.Email != "" {
		user.Name = user.Email
	}

	return user
}

func logSalesAudit(auditService audit.AuditService, ctx context.Context, action string, targetID string, metadata map[string]interface{}) {
	if auditService == nil || strings.TrimSpace(action) == "" || strings.TrimSpace(targetID) == "" {
		return
	}
	auditService.Log(ctx, action, targetID, metadata)
}

func shouldLogSnapshotChange(before, after map[string]interface{}) bool {
	if before == nil && after == nil {
		return false
	}

	beforeJSON, beforeErr := json.Marshal(before)
	afterJSON, afterErr := json.Marshal(after)
	if beforeErr != nil || afterErr != nil {
		return true
	}

	return string(beforeJSON) != string(afterJSON)
}
