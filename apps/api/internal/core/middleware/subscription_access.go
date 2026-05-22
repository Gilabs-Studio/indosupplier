package middleware

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	coreErrors "github.com/gilabs/gims/api/internal/core/errors"
	coreConfig "github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	tenantModels "github.com/gilabs/gims/api/internal/tenant/data/models"
	"github.com/gin-gonic/gin"
)

const subscriptionAccessCacheTTL = 60 * time.Second
const billingPathPrefix = "/api/v1/billing"
const authLogoutPath = "/api/v1/auth/logout"

type subscriptionAccessSnapshot struct {
	State                string
	Enforcement          string
	Status               string
	TenantStatus         string
	DaysOverdue          int
	GracePeriodDays      int
	ForceBillingRedirect bool
	AllowRead            bool
	AllowWrite           bool
	Message              string
	BillingPath          string
}

type subscriptionAccessCacheEntry struct {
	Snapshot  subscriptionAccessSnapshot
	ExpiresAt time.Time
}

var subscriptionAccessCache = struct {
	sync.RWMutex
	entries map[string]subscriptionAccessCacheEntry
}{entries: map[string]subscriptionAccessCacheEntry{}}

type tenantSubscriptionRow struct {
	Status        string
	NextBillingAt *time.Time
	ExpiresAt     *time.Time
}

type tenantDeletionRow struct {
	Status              string
	DeletionScheduledAt *time.Time
}

func enforceSubscriptionAccess(c *gin.Context, userRole, tenantID string) {
	if tenantID == "" {
		return
	}

	snapshot := resolveSubscriptionAccessSnapshot(c.Request.Context(), tenantID)
	c.Set("subscription_access", snapshot)

	reqCtx := c.Request.Context()
	reqCtx = context.WithValue(reqCtx, "subscription_access", snapshot)
	c.Request = c.Request.WithContext(reqCtx)

	isOwnerOrAdmin := isBillingPrivilegedRole(userRole)
	isBillingPath := strings.HasPrefix(c.Request.URL.Path, billingPathPrefix)

	// Logout must always be allowed regardless of subscription state so that
	// suspended users can sign out without being intercepted by payment gates.
	isLogoutPath := c.Request.URL.Path == authLogoutPath

	if snapshot.State == "suspended" && !isLogoutPath {
		if !isOwnerOrAdmin {
			coreErrors.ErrorResponse(c, "ACCOUNT_SUSPENDED", map[string]interface{}{
				"subscription": snapshot,
			}, nil)
			c.Abort()
			return
		}

		if !isBillingPath {
			coreErrors.ErrorResponse(c, "PAYMENT_REQUIRED", map[string]interface{}{
				"subscription": snapshot,
			}, nil)
			c.Abort()
			return
		}
	}

}

func resolveSubscriptionAccessSnapshot(ctx context.Context, tenantID string) subscriptionAccessSnapshot {
	now := apptime.Now()

	cached, ok := getSubscriptionAccessCache(tenantID, now)
	if ok {
		return cached
	}

	snapshot := defaultSubscriptionAccessSnapshot()
	snapshot.GracePeriodDays = currentGracePeriodDays()

	var tenantRow tenantDeletionRow
	err := database.DB.WithContext(ctx).
		Table("tenants").
		Select("status, deletion_scheduled_at").
		Where("id = ? AND deleted_at IS NULL", tenantID).
		Limit(1).
		Scan(&tenantRow).Error
	if err != nil && isMissingTenantDeletionScheduledAtColumnError(err) {
		// Backward compatibility for databases that haven't added deletion_scheduled_at yet.
		_ = database.DB.WithContext(ctx).
			Table("tenants").
			Select("status").
			Where("id = ? AND deleted_at IS NULL", tenantID).
			Limit(1).
			Scan(&tenantRow).Error
	}
	if err == nil || tenantRow.Status != "" {
		snapshot.TenantStatus = strings.ToLower(strings.TrimSpace(tenantRow.Status))
	}

	if snapshot.TenantStatus == "pending_deletion" {
		snapshot.State = "suspended"
		snapshot.Enforcement = "hard_lock"
		snapshot.AllowRead = false
		snapshot.AllowWrite = false
		snapshot.ForceBillingRedirect = false
		if tenantRow.DeletionScheduledAt != nil {
			snapshot.Message = "Account deletion has been requested. This tenant is locked until recovery or permanent deletion. Scheduled deletion at " + tenantRow.DeletionScheduledAt.Format(time.RFC3339)
		} else {
			snapshot.Message = "Account deletion has been requested. This tenant is locked until recovery or permanent deletion."
		}
		setSubscriptionAccessCache(tenantID, snapshot, now)
		return snapshot
	}

	var row tenantSubscriptionRow
	err = database.DB.WithContext(ctx).
		Table("tenant_subscriptions").
		Select("status, next_billing_at, expires_at").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("starts_at DESC").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		setSubscriptionAccessCache(tenantID, snapshot, now)
		return snapshot
	}

	status := strings.ToLower(strings.TrimSpace(row.Status))
	if status == "" {
		setSubscriptionAccessCache(tenantID, snapshot, now)
		return snapshot
	}

	snapshot.Status = status
	dueAt := row.NextBillingAt
	if dueAt == nil {
		dueAt = row.ExpiresAt
	}

	daysOverdue := 0
	if dueAt != nil && now.After(*dueAt) {
		daysOverdue = int(now.Sub(*dueAt).Hours()/24) + 1
	}
	snapshot.DaysOverdue = daysOverdue

	graceDays := snapshot.GracePeriodDays

	if daysOverdue > 0 && (status == string(tenantModels.SubscriptionActive) || status == string(tenantModels.SubscriptionTrial)) {
		_ = database.DB.WithContext(ctx).
			Table("tenant_subscriptions").
			Where("tenant_id = ? AND deleted_at IS NULL AND status IN ?", tenantID, []string{string(tenantModels.SubscriptionActive), string(tenantModels.SubscriptionTrial)}).
			Updates(map[string]interface{}{"status": tenantModels.SubscriptionPastDue}).Error
		status = string(tenantModels.SubscriptionPastDue)
		snapshot.Status = status
	}

	if daysOverdue <= 0 && status == string(tenantModels.SubscriptionPastDue) {
		_ = database.DB.WithContext(ctx).
			Table("tenant_subscriptions").
			Where("tenant_id = ? AND deleted_at IS NULL AND status = ?", tenantID, tenantModels.SubscriptionPastDue).
			Updates(map[string]interface{}{"status": tenantModels.SubscriptionActive}).Error
		status = string(tenantModels.SubscriptionActive)
		snapshot.Status = status
	}

	if daysOverdue > graceDays && status != string(tenantModels.SubscriptionSuspended) {
		_ = database.DB.WithContext(ctx).
			Table("tenant_subscriptions").
			Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
			Updates(map[string]interface{}{"status": tenantModels.SubscriptionSuspended}).Error
		status = string(tenantModels.SubscriptionSuspended)
		snapshot.Status = status
	}

	snapshot = applyAccessState(snapshot, status, daysOverdue, graceDays)
	setSubscriptionAccessCache(tenantID, snapshot, now)
	return snapshot
}

func applyAccessState(snapshot subscriptionAccessSnapshot, status string, daysOverdue, graceDays int) subscriptionAccessSnapshot {
	snapshot.State = "active"
	snapshot.Enforcement = "full_access"
	snapshot.AllowRead = true
	snapshot.AllowWrite = true
	snapshot.ForceBillingRedirect = false
	snapshot.Message = "Subscription is active."

	if status == string(tenantModels.SubscriptionSuspended) || daysOverdue > graceDays {
		snapshot.State = "suspended"
		snapshot.Enforcement = "hard_lock"
		snapshot.AllowRead = false
		snapshot.AllowWrite = false
		snapshot.ForceBillingRedirect = true
		snapshot.Message = "Subscription suspended. Resolve billing to restore access."
		return snapshot
	}

	if status == string(tenantModels.SubscriptionPastDue) {
		if daysOverdue <= 0 {
			return snapshot
		}

		snapshot.State = "grace_period"
		snapshot.Enforcement = "full_access"
		snapshot.AllowRead = true
		snapshot.AllowWrite = true
		snapshot.ForceBillingRedirect = false
		snapshot.Message = "Subscription payment is overdue. Please pay before write access is restricted."
	}

	return snapshot
}

func defaultSubscriptionAccessSnapshot() subscriptionAccessSnapshot {
	return subscriptionAccessSnapshot{
		State:                "active",
		Enforcement:          "full_access",
		Status:               string(tenantModels.SubscriptionActive),
		TenantStatus:         "active",
		DaysOverdue:          0,
		GracePeriodDays:      7,
		ForceBillingRedirect: false,
		AllowRead:            true,
		AllowWrite:           true,
		Message:              "Subscription is active.",
		BillingPath:          "/billing/subscription",
	}
}

func currentGracePeriodDays() int {
	if coreConfig.AppConfig == nil || coreConfig.AppConfig.Subscription.GracePeriodDays < 1 {
		return 7
	}
	return coreConfig.AppConfig.Subscription.GracePeriodDays
}

func getSubscriptionAccessCache(tenantID string, now time.Time) (subscriptionAccessSnapshot, bool) {
	subscriptionAccessCache.RLock()
	entry, ok := subscriptionAccessCache.entries[tenantID]
	subscriptionAccessCache.RUnlock()
	if !ok || now.After(entry.ExpiresAt) {
		return subscriptionAccessSnapshot{}, false
	}
	return entry.Snapshot, true
}

func isMissingTenantDeletionScheduledAtColumnError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return (strings.Contains(errMsg, "sqlstate 42703") && strings.Contains(errMsg, "deletion_scheduled_at")) ||
		strings.Contains(errMsg, "column \"deletion_scheduled_at\" does not exist")
}

func setSubscriptionAccessCache(tenantID string, snapshot subscriptionAccessSnapshot, now time.Time) {
	subscriptionAccessCache.Lock()
	subscriptionAccessCache.entries[tenantID] = subscriptionAccessCacheEntry{
		Snapshot:  snapshot,
		ExpiresAt: now.Add(subscriptionAccessCacheTTL),
	}
	subscriptionAccessCache.Unlock()
}

func isWriteHTTPMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

func isBillingPrivilegedRole(role string) bool {
	normalized := strings.ToLower(strings.TrimSpace(role))
	if normalized == "admin" || normalized == "superadmin" {
		return true
	}
	return strings.HasPrefix(normalized, "tenant_owner_") || normalized == "owner" || strings.HasSuffix(normalized, "_owner")
}
