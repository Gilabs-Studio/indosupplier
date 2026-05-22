package service

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/gilabs/gims/api/internal/notification/data/models"
	notificationWS "github.com/gilabs/gims/api/internal/notification/infrastructure/ws"
	"gorm.io/gorm"
)

type ApprovalNotificationParams struct {
	PermissionCode string
	EntityType     string
	EntityID       string
	Title          string
	Message        string
	ActorUserID    string
}

type approvalRecipient struct {
	UserID     string
	Email      string
	Scope      string
	EmployeeID string
	DivisionID string
	AreaIDs    []string
}

type targetScopeContext struct {
	EntityType      string
	EntityID        string
	CreatorUserID   string
	OwnerEmployeeID string
	DivisionID      string
	AreaIDs         []string
}

// CreateApprovalNotification sends approval notifications to all active users
// who own the required approval permission code.
func CreateApprovalNotification(ctx context.Context, db *gorm.DB, params ApprovalNotificationParams) error {
	if db == nil {
		return nil
	}

	traceSQL := isApprovalNotificationTraceEnabled()
	if traceSQL {
		db = db.Debug()
	}

	permissionCode := strings.TrimSpace(params.PermissionCode)
	entityType := strings.TrimSpace(params.EntityType)
	entityID := strings.TrimSpace(params.EntityID)
	title := strings.TrimSpace(params.Title)
	message := strings.TrimSpace(params.Message)
	actorUserID := strings.TrimSpace(params.ActorUserID)

	if permissionCode == "" && entityType != "" {
		permissionCode = entityType + ".approve"
	}

	if permissionCode == "" || entityType == "" || entityID == "" || title == "" || message == "" {
		if traceSQL {
			log.Printf("[approval_notification] skipped: missing required fields permission=%q entityType=%q entityID=%q", permissionCode, entityType, entityID)
		}
		return nil
	}

	targetCtx, err := resolveTargetScopeContext(ctx, db, entityType, entityID)
	if err != nil {
		if traceSQL {
			log.Printf("[approval_notification] resolve target scope failed: permission=%s entityType=%s entityID=%s err=%v", permissionCode, entityType, entityID, err)
		}
		return err
	}

	recipients, err := findApprovalRecipients(ctx, db, permissionCode, entityType)
	if err != nil {
		if traceSQL {
			log.Printf("[approval_notification] find recipients failed: permission=%s err=%v", permissionCode, err)
		}
		return err
	}

	notifications := make([]models.Notification, 0, len(recipients))
	scopedRecipientCount := 0
	for _, recipient := range recipients {
		if strings.TrimSpace(recipient.UserID) == "" {
			continue
		}
		if !isRecipientAllowedByScope(recipient, targetCtx) {
			continue
		}
		scopedRecipientCount++

		notifications = append(notifications, models.Notification{
			UserID:     recipient.UserID,
			Type:       models.NotificationTypeApprovalRequest,
			Title:      title,
			Message:    message,
			EntityType: entityType,
			EntityID:   entityID,
			IsRead:     false,
		})
	}

	fallbackUsed := false
	if len(notifications) == 0 {
		fallbackUsed = true
		for _, recipient := range recipients {
			if strings.TrimSpace(recipient.UserID) == "" {
				continue
			}

			notifications = append(notifications, models.Notification{
				UserID:     recipient.UserID,
				Type:       models.NotificationTypeApprovalRequest,
				Title:      title,
				Message:    message,
				EntityType: entityType,
				EntityID:   entityID,
				IsRead:     false,
			})
		}
	}

	if len(notifications) == 0 {
		if traceSQL {
			log.Printf("[approval_notification] no recipient after filtering: permission=%s entityType=%s entityID=%s actor=%s recipients=%d scoped=%d fallback=%t", permissionCode, entityType, entityID, actorUserID, len(recipients), scopedRecipientCount, fallbackUsed)
		}
		return nil
	}

	if traceSQL {
		log.Printf("[approval_notification] create notifications: permission=%s entityType=%s entityID=%s actor=%s recipients=%d scoped=%d final=%d fallback=%t", permissionCode, entityType, entityID, actorUserID, len(recipients), scopedRecipientCount, len(notifications), fallbackUsed)
	}

	if err := db.WithContext(ctx).Create(&notifications).Error; err != nil {
		return err
	}

	for _, notification := range notifications {
		notificationWS.DefaultNotificationHub().PublishCreated(notification)
	}

	return nil
}

func isApprovalNotificationTraceEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("APPROVAL_NOTIFICATION_TRACE_SQL")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func findApprovalRecipients(ctx context.Context, db *gorm.DB, permissionCode, entityType string) ([]approvalRecipient, error) {
	rows := make([]struct {
		UserID     string
		Email      sql.NullString
		Scope      sql.NullString
		EmployeeID sql.NullString
		DivisionID sql.NullString
	}, 0)

	buildBaseQuery := func() *gorm.DB {
		return db.WithContext(ctx).
			Table("users").
			Select("users.id AS user_id, users.email AS email, COALESCE(NULLIF(rp.scope, ''), 'ALL') AS scope, e.id AS employee_id, e.division_id AS division_id").
			Joins("JOIN role_permissions rp ON rp.role_id = users.role_id").
			Joins("JOIN permissions p ON p.id = rp.permission_id").
			Joins("LEFT JOIN employees e ON e.user_id = users.id AND e.deleted_at IS NULL").
			Where("users.deleted_at IS NULL").
			Where("users.status = ?", "active")
	}

	err := buildBaseQuery().
		Where("p.code = ?", permissionCode).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 && strings.TrimSpace(entityType) != "" {
		err = buildBaseQuery().
			Where("p.resource = ?", entityType).
			Where("p.action = ?", "APPROVE").
			Scan(&rows).Error
		if err != nil {
			return nil, err
		}
	}

	recipients := make([]approvalRecipient, 0, len(rows))
	byUserID := make(map[string]int, len(rows))
	employeeIDs := make([]string, 0, len(rows))
	unresolvedUserIDs := make([]string, 0)
	unresolvedEmails := make([]string, 0)
	for _, row := range rows {
		recipient := approvalRecipient{
			UserID: row.UserID,
			Scope:  normalizeScope(row.Scope.String),
		}
		if row.Email.Valid {
			recipient.Email = strings.ToLower(strings.TrimSpace(row.Email.String))
		}
		if row.EmployeeID.Valid {
			recipient.EmployeeID = row.EmployeeID.String
			employeeIDs = append(employeeIDs, row.EmployeeID.String)
		} else {
			unresolvedUserIDs = append(unresolvedUserIDs, row.UserID)
			if recipient.Email != "" {
				unresolvedEmails = append(unresolvedEmails, recipient.Email)
			}
		}
		if row.DivisionID.Valid {
			recipient.DivisionID = row.DivisionID.String
		}
		byUserID[recipient.UserID] = len(recipients)
		recipients = append(recipients, recipient)
	}

	if len(unresolvedUserIDs) > 0 || len(unresolvedEmails) > 0 {
		type employeeFallbackRow struct {
			EmployeeID string
			DivisionID sql.NullString
			UserID     sql.NullString
			Email      sql.NullString
		}

		fallbackRows := make([]employeeFallbackRow, 0)
		q := db.WithContext(ctx).
			Table("employees").
			Select("employees.id AS employee_id, employees.division_id AS division_id, employees.user_id AS user_id, employees.email AS email").
			Where("employees.deleted_at IS NULL")

		if len(unresolvedUserIDs) > 0 && len(unresolvedEmails) > 0 {
			q = q.Where("employees.user_id IN ? OR LOWER(employees.email) IN ?", unresolvedUserIDs, unresolvedEmails)
		} else if len(unresolvedUserIDs) > 0 {
			q = q.Where("employees.user_id IN ?", unresolvedUserIDs)
		} else {
			q = q.Where("LOWER(employees.email) IN ?", unresolvedEmails)
		}

		if err := q.Scan(&fallbackRows).Error; err != nil {
			return nil, err
		}

		byEmail := make(map[string]employeeFallbackRow)
		for _, fr := range fallbackRows {
			if fr.UserID.Valid {
				if idx, ok := byUserID[fr.UserID.String]; ok {
					recipients[idx].EmployeeID = fr.EmployeeID
					if fr.DivisionID.Valid {
						recipients[idx].DivisionID = fr.DivisionID.String
					}
					employeeIDs = append(employeeIDs, fr.EmployeeID)
				}
			}
			if fr.Email.Valid {
				email := strings.ToLower(strings.TrimSpace(fr.Email.String))
				if email != "" {
					if _, exists := byEmail[email]; !exists {
						byEmail[email] = fr
					}
				}
			}
		}

		for i := range recipients {
			if recipients[i].EmployeeID != "" || recipients[i].Email == "" {
				continue
			}
			if fr, exists := byEmail[recipients[i].Email]; exists {
				recipients[i].EmployeeID = fr.EmployeeID
				if fr.DivisionID.Valid {
					recipients[i].DivisionID = fr.DivisionID.String
				}
				employeeIDs = append(employeeIDs, fr.EmployeeID)
			}
		}
	}

	if len(employeeIDs) == 0 {
		return recipients, nil
	}

	areaRows := make([]struct {
		EmployeeID string
		AreaID     string
	}, 0)
	if err := db.WithContext(ctx).
		Table("employee_areas").
		Select("employee_id, area_id").
		Where("employee_id IN ?", employeeIDs).
		Scan(&areaRows).Error; err != nil {
		return nil, err
	}

	areaMap := make(map[string][]string)
	for _, row := range areaRows {
		if row.EmployeeID == "" || row.AreaID == "" {
			continue
		}
		areaMap[row.EmployeeID] = append(areaMap[row.EmployeeID], row.AreaID)
	}

	for i := range recipients {
		if recipients[i].EmployeeID != "" {
			recipients[i].AreaIDs = areaMap[recipients[i].EmployeeID]
		}
	}

	return recipients, nil
}

func resolveTargetScopeContext(ctx context.Context, db *gorm.DB, entityType, entityID string) (targetScopeContext, error) {
	target := targetScopeContext{EntityType: entityType, EntityID: entityID}

	switch entityType {
	case "sales_quotation":
		row := struct {
			CreatedBy  sql.NullString
			SalesRepID sql.NullString
		}{}
		err := db.WithContext(ctx).
			Table("sales_quotations").
			Select("created_by, sales_rep_id").
			Where("id = ? AND deleted_at IS NULL", entityID).
			Take(&row).Error
		if err != nil {
			return target, err
		}

		if row.CreatedBy.Valid {
			target.CreatorUserID = row.CreatedBy.String
		}
		if row.SalesRepID.Valid {
			target.OwnerEmployeeID = row.SalesRepID.String
		}
	case "purchase_requisition":
		row := struct {
			EmployeeID sql.NullString
		}{}
		err := db.WithContext(ctx).
			Table("purchase_requisitions").
			Select("employee_id").
			Where("id = ? AND deleted_at IS NULL", entityID).
			Take(&row).Error
		if err != nil {
			return target, err
		}

		if row.EmployeeID.Valid {
			target.OwnerEmployeeID = row.EmployeeID.String
			empByID := struct {
				UserID sql.NullString
			}{}
			if err := db.WithContext(ctx).
				Table("employees").
				Select("user_id").
				Where("id = ? AND deleted_at IS NULL", row.EmployeeID.String).
				Take(&empByID).Error; err == nil {
				if empByID.UserID.Valid {
					target.CreatorUserID = empByID.UserID.String
				}
			}
		}
	default:
		tableName, hasTable := resolveEntityTableName(entityType)
		if !hasTable {
			return target, nil
		}

		row := struct {
			CreatedBy sql.NullString
		}{}
		err := db.WithContext(ctx).
			Table(tableName).
			Select("created_by").
			Where("id = ? AND deleted_at IS NULL", entityID).
			Take(&row).Error
		if err != nil {
			if isMissingCreatedByColumnError(err) {
				return target, nil
			}
			return target, err
		}
		if row.CreatedBy.Valid {
			target.CreatorUserID = row.CreatedBy.String
		}
	}

	if target.OwnerEmployeeID == "" && target.CreatorUserID != "" {
		empByUser := struct {
			ID         string
			DivisionID sql.NullString
		}{}
		err := db.WithContext(ctx).
			Table("employees").
			Select("id, division_id").
			Where("user_id = ? AND deleted_at IS NULL", target.CreatorUserID).
			Take(&empByUser).Error
		if err == nil {
			target.OwnerEmployeeID = empByUser.ID
			if empByUser.DivisionID.Valid {
				target.DivisionID = empByUser.DivisionID.String
			}
		}
	}

	if target.OwnerEmployeeID != "" {
		empRow := struct {
			DivisionID sql.NullString
		}{}
		if err := db.WithContext(ctx).
			Table("employees").
			Select("division_id").
			Where("id = ? AND deleted_at IS NULL", target.OwnerEmployeeID).
			Take(&empRow).Error; err == nil {
			if empRow.DivisionID.Valid {
				target.DivisionID = empRow.DivisionID.String
			}
		}

		var areaIDs []string
		if err := db.WithContext(ctx).
			Table("employee_areas").
			Where("employee_id = ?", target.OwnerEmployeeID).
			Pluck("area_id", &areaIDs).Error; err == nil {
			target.AreaIDs = areaIDs
		}
	}

	if target.CreatorUserID != "" && (target.DivisionID == "" || len(target.AreaIDs) == 0) {
		creatorEmp := struct {
			ID         string
			DivisionID sql.NullString
		}{}
		if err := db.WithContext(ctx).
			Table("employees").
			Select("id, division_id").
			Where("user_id = ? AND deleted_at IS NULL", target.CreatorUserID).
			Take(&creatorEmp).Error; err == nil {
			if target.DivisionID == "" && creatorEmp.DivisionID.Valid {
				target.DivisionID = creatorEmp.DivisionID.String
			}
			if len(target.AreaIDs) == 0 && creatorEmp.ID != "" {
				var creatorAreaIDs []string
				if err := db.WithContext(ctx).
					Table("employee_areas").
					Where("employee_id = ?", creatorEmp.ID).
					Pluck("area_id", &creatorAreaIDs).Error; err == nil {
					target.AreaIDs = creatorAreaIDs
				}
			}
		}
	}

	return target, nil
}

func isMissingCreatedByColumnError(err error) bool {
	if err == nil {
		return false
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "column") && strings.Contains(errText, "created_by") && strings.Contains(errText, "does not exist")
}

func resolveEntityTableName(entityType string) (string, bool) {
	switch entityType {
	case "company":
		return "companies", true
	case "purchase_requisition":
		return "purchase_requisitions", true
	case "sales_order":
		return "sales_orders", true
	case "customer_invoice":
		return "customer_invoices", true
	case "customer_invoice_dp":
		return "customer_invoices", true
	case "purchase_order":
		return "purchase_orders", true
	case "goods_receipt":
		return "goods_receipts", true
	case "supplier_invoice":
		return "supplier_invoices", true
	case "supplier_invoice_dp":
		return "supplier_invoices", true
	case "non_trade_payable":
		return "non_trade_payables", true
	case "travel_plan":
		return "travel_plans", true
	case "leave_request":
		return "leave_requests", true
	case "recruitment":
		return "recruitment_requests", true
	case "overtime":
		return "overtime_requests", true
	case "stock_opname":
		return "stock_opnames", true
	case "crm_visit":
		return "visit_reports", true
	default:
		return "", false
	}
}

func normalizeScope(scope string) string {
	s := strings.ToUpper(strings.TrimSpace(scope))
	if s == "" {
		return "ALL"
	}
	return s
}

// POSOrderNotificationParams holds the data needed to fan-out a self-order notification.
type POSOrderNotificationParams struct {
	TenantID   string
	OutletID   string
	OutletName string
	TableLabel string
	OrderID    string
}

// CreatePOSOrderNotification creates in-app notifications for all staff who have
// pos.order.read permission scoped to the given outlet.  It mirrors the
// CreateApprovalNotification fan-out pattern but filters recipients by
// outlet-scope assignment instead of entity-level scope.
//
// Recipients are selected by one of two criteria:
//   - Their role_permission scope is ALL (unrestricted), or
//   - They are explicitly assigned to the outlet via employee_outlets.
//
// The function uses the raw db handle (not database.GetDB) because the shared
// permissions table may have NULL tenant_id.
func CreatePOSOrderNotification(ctx context.Context, db *gorm.DB, params POSOrderNotificationParams) error {
	if db == nil {
		return nil
	}

	tenantID := strings.TrimSpace(params.TenantID)
	outletID := strings.TrimSpace(params.OutletID)
	if tenantID == "" || outletID == "" {
		return nil
	}

	tableLabel := strings.TrimSpace(params.TableLabel)
	if tableLabel == "" {
		tableLabel = "Meja"
	}

	outletName := strings.TrimSpace(params.OutletName)
	if outletName == "" {
		outletName = "Outlet"
	}

	title := "Order Baru"
	message := tableLabel + " — " + outletName

	type recipientRow struct {
		UserID string
	}

	var rows []recipientRow

	// Query users with pos.order.read permission for this tenant+outlet.
	// We LEFT JOIN employee_outlets so that scope=ALL users are included
	// unconditionally, while scope=OUTLET users are filtered to this outlet.
	err := db.WithContext(ctx).
		Table("users u").
		Select("DISTINCT u.id AS user_id").
		Joins("JOIN role_permissions rp ON rp.role_id = u.role_id").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Joins("LEFT JOIN employees e ON e.user_id = u.id AND e.deleted_at IS NULL").
		Joins("LEFT JOIN employee_outlets eo ON eo.employee_id = e.id").
		Where("u.deleted_at IS NULL").
		Where("u.status = ?", "active").
		Where("u.tenant_id = ?", tenantID).
		Where("p.code = ?", "pos.order.read").
		Where("COALESCE(NULLIF(rp.scope, ''), 'ALL') = 'ALL' OR eo.outlet_id = ?", outletID).
		Scan(&rows).Error
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		return nil
	}

	notifications := make([]models.Notification, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.UserID) == "" {
			continue
		}
		notifications = append(notifications, models.Notification{
			TenantID:   tenantID,
			UserID:     row.UserID,
			Type:       models.NotificationTypeInfo,
			Title:      title,
			Message:    message,
			EntityType: "pos_self_order",
			EntityID:   outletID,
			IsRead:     false,
		})
	}

	if len(notifications) == 0 {
		return nil
	}

	if err := db.WithContext(ctx).Create(&notifications).Error; err != nil {
		return err
	}

	for _, n := range notifications {
		notificationWS.DefaultNotificationHub().PublishCreated(n)
	}

	return nil
}

func isRecipientAllowedByScope(recipient approvalRecipient, target targetScopeContext) bool {
	if target.EntityID == "" {
		return false
	}

	scope := normalizeScope(recipient.Scope)
	switch scope {
	case "ALL":
		return true
	case "OWN":
		if target.CreatorUserID == "" {
			return false
		}
		return recipient.UserID == target.CreatorUserID
	case "DIVISION":
		if recipient.DivisionID == "" || target.DivisionID == "" {
			// Fallback for legacy/missing employee-division mapping.
			// Keep strict match only when both division IDs are available.
			return true
		}
		return recipient.DivisionID == target.DivisionID
	case "AREA":
		if len(recipient.AreaIDs) == 0 || len(target.AreaIDs) == 0 {
			return false
		}
		targetAreaSet := make(map[string]struct{}, len(target.AreaIDs))
		for _, areaID := range target.AreaIDs {
			targetAreaSet[areaID] = struct{}{}
		}
		for _, recipientAreaID := range recipient.AreaIDs {
			if _, exists := targetAreaSet[recipientAreaID]; exists {
				return true
			}
		}
		return false
	default:
		return false
	}
}
