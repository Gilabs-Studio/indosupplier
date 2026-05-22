package usecase

import (
	"context"
	"errors"

	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/events"
	coreDB "github.com/gilabs/gims/api/internal/core/infrastructure/database"
	infraEvents "github.com/gilabs/gims/api/internal/core/infrastructure/events"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"
	"github.com/gilabs/gims/api/internal/core/middleware"
	"github.com/gilabs/gims/api/internal/role/data/models"
	"github.com/gilabs/gims/api/internal/role/data/repositories"
	"github.com/gilabs/gims/api/internal/role/domain/dto"
	"github.com/gilabs/gims/api/internal/role/domain/mapper"
	tenantRepos "github.com/gilabs/gims/api/internal/tenant/data/repositories"
	tenantPolicy "github.com/gilabs/gims/api/internal/tenant/domain/policy"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound                = errors.New("role not found")
	ErrRoleAlreadyExists           = errors.New("role already exists")
	ErrRoleProtected               = errors.New("role is protected and cannot be deleted or modified")
	ErrRoleInUse                   = errors.New("role is in use by users and cannot be deleted")
	ErrLastAdminCannotDelete       = errors.New("cannot delete the last admin role")
	ErrLastAdminCannotDisable      = errors.New("cannot disable the last admin role")
	ErrPermissionNotAllowedForPlan = errors.New("one or more permissions are not allowed for the tenant plan")
)

const (
	cacheRoleByIDKeyLegacy    = "roles:id:%s"
	cacheRoleListKeyFmtLegacy = "roles:list:page:%d:limit:%d"
	cacheRoleListPage1Limit10 = "roles:list:page:1:limit:10"
	cacheRoleListPage1Limit20 = "roles:list:page:1:limit:20"
	masterDataMenuPrefix      = "/master-data"
)

func roleCacheScope(ctx context.Context) string {
	if middleware.IsSystemAdmin(ctx) {
		return "system_admin"
	}
	if tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx)); tenantID != "" {
		return tenantID
	}
	return "public"
}

func roleListCacheKey(ctx context.Context, page, limit int, search string) string {
	return fmt.Sprintf("roles:list:scope:%s:page:%d:limit:%d:search:%s", roleCacheScope(ctx), page, limit, search)
}

func roleByIDCacheKey(ctx context.Context, roleID string) string {
	return fmt.Sprintf("roles:id:scope:%s:id:%s", roleCacheScope(ctx), roleID)
}

var alwaysSafePermissionPrefixes = []string{
	"profile.",
	"setting",
	"billing.",
	"user.",
	"role.",
	"permission.",
	"company.",
	"warehouse.",
	"product.",
	"customer.",
	"supplier.",
	"employee.",
}

var mandatoryPermissionCodes = []string{"user.read", "role.read", "permission.read"}
var systemAdminOnlyPermissionCodes = []string{"currency.create", "currency.update", "currency.delete"}

type RoleUsecase interface {
	List(ctx context.Context, page, limit int, search string) ([]dto.RoleResponse, int64, error)
	GetByID(ctx context.Context, id string) (*dto.RoleResponse, error)
	Create(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error)
	Update(ctx context.Context, id string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	Delete(ctx context.Context, id string) error
	AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error
	AssignPermissionsWithScope(ctx context.Context, roleID string, assignments []dto.PermissionAssignment) error
	// UpdatePermissionsDiff applies differential permission updates (added/removed only).
	// Only validates added permissions against plan; leaves existing unchanged untouched.
	UpdatePermissionsDiff(ctx context.Context, roleID string, diff *dto.DiffPermissionsRequest) error
	GetMenuAccess(ctx context.Context, roleID string) ([]dto.RoleMenuAccessResponse, error)
	UpdateMenuAccess(ctx context.Context, roleID string, assignments []dto.RoleMenuAccessAssignment) error
	FilterAllowedPermissions(ctx context.Context, permissionIDs []string) (allowed []string, blocked []string, err error)
	ValidateUserRole(ctx context.Context, userID string, roleID string) (bool, error)
}

type roleUsecase struct {
	roleRepo       repositories.RoleRepository
	eventPublisher infraEvents.EventPublisher
	redis          *redis.Client
	permService    security.PermissionService
	db             *gorm.DB
	planRepo       tenantRepos.SubscriptionPlanRepository
}

func NewRoleUsecase(roleRepo repositories.RoleRepository, eventPublisher infraEvents.EventPublisher, redis *redis.Client, permService security.PermissionService, db *gorm.DB, planRepo tenantRepos.SubscriptionPlanRepository) RoleUsecase {
	return &roleUsecase{
		roleRepo:       roleRepo,
		eventPublisher: eventPublisher,
		redis:          redis,
		permService:    permService,
		db:             db,
		planRepo:       planRepo,
	}
}

func (u *roleUsecase) List(ctx context.Context, page, limit int, search string) ([]dto.RoleResponse, int64, error) {
	cacheKey := roleListCacheKey(ctx, page, limit, search)

	// Try to get from cache
	val, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedResult struct {
			Roles []dto.RoleResponse `json:"roles"`
			Total int64              `json:"total"`
		}
		if err := json.Unmarshal([]byte(val), &cachedResult); err == nil {
			return cachedResult.Roles, cachedResult.Total, nil
		}
	}

	// Get from DB
	roles, total, err := u.roleRepo.List(ctx, page, limit, search)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.RoleResponse, len(roles))
	for i, r := range roles {
		responses[i] = *mapper.ToRoleResponse(&r)
	}

	// Cache result (TTL 5 minutes)
	cacheData := struct {
		Roles []dto.RoleResponse `json:"roles"`
		Total int64              `json:"total"`
	}{
		Roles: responses,
		Total: total,
	}

	if data, err := json.Marshal(cacheData); err == nil {
		u.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return responses, total, nil
}

func (u *roleUsecase) GetByID(ctx context.Context, id string) (*dto.RoleResponse, error) {
	cacheKey := roleByIDCacheKey(ctx, id)

	// Try to get from cache
	val, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedRole dto.RoleResponse
		if err := json.Unmarshal([]byte(val), &cachedRole); err == nil {
			return &cachedRole, nil
		}
	}

	r, err := u.roleRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	resp := mapper.ToRoleResponse(r)

	// Cache result (TTL 15 minutes)
	if data, err := json.Marshal(resp); err == nil {
		u.redis.Set(ctx, cacheKey, data, 15*time.Minute)
	}

	return resp, nil
}

func (u *roleUsecase) Create(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	// Check if code already exists
	_, err := u.roleRepo.FindByCode(ctx, req.Code)
	if err == nil {
		return nil, ErrRoleAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Set default status
	status := req.Status
	if status == "" {
		status = "active"
	}

	// Invalidate cache
	// We scan only a few keys or just expire them. Since we don't know all pages,
	// specific invalidation is hard without maintaining a set of keys.
	// For now, let's just invalidate the specific role ID cache.
	// List cache uses 5 min TTL, so it will expire eventually.
	// For immediate consistency, we could use keys pattern matching but it's expensive.
	// Enterprise solution: Use a "version" key for lists or just accept eventual consistency (5 mins).
	// Let's implement partial invalidation for common first page.
	u.redis.Del(ctx,
		roleListCacheKey(ctx, 1, 10, ""),
		roleListCacheKey(ctx, 1, 20, ""),
		cacheRoleListPage1Limit10,
		cacheRoleListPage1Limit20,
	)

	// Create role
	r := &models.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      status,
		IsProtected: false, // Always false for user-created roles
	}

	if err := u.roleRepo.Create(ctx, r); err != nil {
		return nil, err
	}

	// Reload with permissions
	createdRole, err := u.roleRepo.FindByID(ctx, r.ID)
	if err != nil {
		return nil, err
	}

	// Publish event (async, fire-and-forget)
	u.eventPublisher.PublishAsync(ctx, events.NewRoleCreatedEvent(ctx, events.RoleCreatedPayload{
		RoleID:      r.ID,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
	}))

	return mapper.ToRoleResponse(createdRole), nil
}

func (u *roleUsecase) Update(ctx context.Context, id string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	// Find role
	r, err := u.roleRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	// Validate protected role changes in helper to keep complexity low
	if err := u.validateProtectedRoleChange(ctx, r, req); err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		r.Name = req.Name
	}

	if req.Code != "" {
		// Check if code already exists (excluding current role)
		existingRole, err := u.roleRepo.FindByCode(ctx, req.Code)
		if err == nil && existingRole.ID != id {
			return nil, ErrRoleAlreadyExists
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.Code = req.Code
	}

	if req.Description != "" {
		r.Description = req.Description
	}

	if req.Status != "" {
		r.Status = req.Status
	}

	if err := u.roleRepo.Update(ctx, r); err != nil {
		return nil, err
	}

	// Invalidate cache (tenant-scoped + legacy keys)
	u.redis.Del(ctx,
		roleByIDCacheKey(ctx, id),
		fmt.Sprintf(cacheRoleByIDKeyLegacy, id),
		roleListCacheKey(ctx, 1, 10, ""),
		roleListCacheKey(ctx, 1, 20, ""),
		cacheRoleListPage1Limit10,
		cacheRoleListPage1Limit20,
	)

	// Reload with permissions
	updatedRole, err := u.roleRepo.FindByID(ctx, r.ID)
	if err != nil {
		return nil, err
	}

	// Publish event (async, fire-and-forget)
	u.eventPublisher.PublishAsync(ctx, events.NewRoleUpdatedEvent(ctx, events.RoleUpdatedPayload{
		RoleID:      r.ID,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		Status:      r.Status,
		UpdatedAt:   r.UpdatedAt,
	}))

	return mapper.ToRoleResponse(updatedRole), nil
}

func (u *roleUsecase) validateProtectedRoleChange(ctx context.Context, r *models.Role, req *dto.UpdateRoleRequest) error {
	if !r.IsProtected {
		return nil
	}

	// Protected roles cannot have status changed to inactive if it's the last admin
	if req.Status == "inactive" && r.Code == "admin" {
		adminCount, err := u.roleRepo.CountAdmins(ctx)
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return ErrLastAdminCannotDisable
		}
	}

	// Protected roles cannot have code changed
	if req.Code != "" && req.Code != r.Code {
		return ErrRoleProtected
	}

	return nil
}

func (u *roleUsecase) Delete(ctx context.Context, id string) error {
	// Check if role exists
	r, err := u.roleRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	// Check if role is protected
	if r.IsProtected {
		return ErrRoleProtected
	}

	// Check if role is admin and it's the last admin
	if r.Code == "admin" {
		adminCount, err := u.roleRepo.CountAdmins(ctx)
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return ErrLastAdminCannotDelete
		}
	}

	// Check if role is in use by users
	userCount, err := u.roleRepo.CountUsersByRoleID(ctx, id)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return ErrRoleInUse
	}

	if err := u.roleRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache (tenant-scoped + legacy keys)
	u.redis.Del(ctx,
		roleByIDCacheKey(ctx, id),
		fmt.Sprintf(cacheRoleByIDKeyLegacy, id),
		roleListCacheKey(ctx, 1, 10, ""),
		roleListCacheKey(ctx, 1, 20, ""),
		cacheRoleListPage1Limit10,
		cacheRoleListPage1Limit20,
	)

	// Publish event (async, fire-and-forget)
	u.eventPublisher.PublishAsync(ctx, events.NewRoleDeletedEvent(ctx, events.RoleDeletedPayload{
		RoleID:    id,
		Code:      r.Code,
		DeletedAt: apptime.Now(),
	}))

	return nil
}

func (u *roleUsecase) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	// Check if role exists
	role, err := u.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	permissionIDs, err = u.withPreservedAlwaysSafePermissionIDs(ctx, role.Code, roleID, permissionIDs)
	if err != nil {
		return err
	}

	if err := u.ensureSystemAdminOnlyPermissions(ctx, permissionIDs); err != nil {
		return err
	}

	if err := u.validatePlanPermissionBoundary(ctx, permissionIDs); err != nil {
		return err
	}

	if err := u.roleRepo.AssignPermissions(ctx, roleID, permissionIDs); err != nil {
		return err
	}

	u.invalidatePermissionCaches(ctx, roleID, role.Code)

	// Publish event (async, fire-and-forget)
	u.eventPublisher.PublishAsync(ctx, events.NewRolePermissionsAssignedEvent(ctx, events.RolePermissionsAssignedPayload{
		RoleID:        roleID,
		PermissionIDs: permissionIDs,
		AssignedAt:    apptime.Now(),
	}))

	return nil
}

func (u *roleUsecase) AssignPermissionsWithScope(ctx context.Context, roleID string, assignments []dto.PermissionAssignment) error {
	// Validate role exists
	role, err := u.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	// Validate all scopes
	for _, a := range assignments {
		if !models.IsValidScope(a.Scope) {
			return fmt.Errorf("invalid scope '%s' for permission '%s'", a.Scope, a.PermissionID)
		}
	}

	// Convert to model-level assignments
	rolePerms := make([]models.RolePermission, 0, len(assignments))
	permIDs := make([]string, 0, len(assignments))
	for _, a := range assignments {
		scope := a.Scope
		if scope == "" {
			scope = models.ScopeAll
		}
		permIDs = append(permIDs, a.PermissionID)
		rolePerms = append(rolePerms, models.RolePermission{
			RoleID:       roleID,
			PermissionID: a.PermissionID,
			Scope:        scope,
		})
	}

	rolePerms, err = u.withPreservedAlwaysSafeRolePermissions(ctx, role.Code, roleID, rolePerms)
	if err != nil {
		return err
	}
	permIDs = make([]string, 0, len(rolePerms))
	for _, rp := range rolePerms {
		permIDs = append(permIDs, rp.PermissionID)
	}

	if err := u.ensureSystemAdminOnlyPermissions(ctx, permIDs); err != nil {
		return err
	}

	if err := u.validatePlanPermissionBoundary(ctx, permIDs); err != nil {
		return err
	}

	if err := u.roleRepo.AssignPermissionsWithScope(ctx, roleID, rolePerms); err != nil {
		return err
	}

	u.invalidatePermissionCaches(ctx, roleID, role.Code)

	// Publish event with permission IDs
	u.eventPublisher.PublishAsync(ctx, events.NewRolePermissionsAssignedEvent(ctx, events.RolePermissionsAssignedPayload{
		RoleID:        role.ID,
		PermissionIDs: permIDs,
		AssignedAt:    apptime.Now(),
	}))

	return nil
}

// UpdatePermissionsDiff applies differential permission updates.
// Only validates added permissions against plan; existing unchanged permissions are left untouched.
// This prevents the "plan restriction" error when bulk-updating with unchanged permissions included.
func (u *roleUsecase) UpdatePermissionsDiff(ctx context.Context, roleID string, diff *dto.DiffPermissionsRequest) error {
	tenantID := middleware.TenantFromContext(ctx)
	log.Printf("[RoleUsecase] UpdatePermissionsDiff tenant_id=%s role_id=%s added=%d removed=%d added_with_scope=%d", tenantID, roleID, len(diff.Added), len(diff.Removed), len(diff.AddedWithScope))

	// Validate role exists and get existing permissions
	role, err := u.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	// Get current permissions from role.RolePermissions (already preloaded from FindByID)
	existingSet := make(map[string]models.RolePermission)
	for _, rp := range role.RolePermissions {
		existingSet[rp.PermissionID] = rp
	}

	// Build next state: keep existing, remove specified, add specified
	nextSet := make(map[string]models.RolePermission)

	// Start with existing permissions
	for permID, rp := range existingSet {
		nextSet[permID] = rp
	}

	// Remove specified permission IDs
	for _, permID := range diff.Removed {
		delete(nextSet, permID)
	}

	// Add new permissions (validate these against plan)
	addedToValidate := make([]string, 0)

	// Handle simple added permissions
	for _, permID := range diff.Added {
		if _, exists := nextSet[permID]; !exists {
			nextSet[permID] = models.RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
				Scope:        models.ScopeAll,
			}
			addedToValidate = append(addedToValidate, permID)
		}
	}

	// Handle scope-aware added permissions
	for _, a := range diff.AddedWithScope {
		if !models.IsValidScope(a.Scope) {
			return fmt.Errorf("invalid scope '%s' for permission '%s'", a.Scope, a.PermissionID)
		}
		if _, exists := nextSet[a.PermissionID]; !exists {
			nextSet[a.PermissionID] = models.RolePermission{
				RoleID:       roleID,
				PermissionID: a.PermissionID,
				Scope:        a.Scope,
			}
			addedToValidate = append(addedToValidate, a.PermissionID)
		}
	}

	// Only validate added permissions against plan boundary
	if len(addedToValidate) > 0 {
		if err := u.validatePlanPermissionBoundary(ctx, addedToValidate); err != nil {
			return err
		}
	}

	// Preserve always-safe permissions (conditionally based on role type)
	// Special roles (owner, admin, managers) must keep protected permissions.
	// Regular roles can freely remove them.
	var attemptedToRemoveProtected []string
	if isSpecialRole(role.Code) {
		// For special roles, preserve and protect always-safe permissions
		for permID, rp := range existingSet {
			if rp.Permission != nil && isAlwaysSafePermissionCode(rp.Permission.Code) {
				// Check if user tried to remove this protected permission
				for _, removedID := range diff.Removed {
					if removedID == permID {
						attemptedToRemoveProtected = append(attemptedToRemoveProtected, rp.Permission.Code)
						break
					}
				}
				nextSet[permID] = rp
			}
		}
	}
	// For non-special roles, allow removal of always-safe permissions (don't re-add them to nextSet)

	// If user tried to remove protected permissions, return error
	if len(attemptedToRemoveProtected) > 0 {
		sort.Strings(attemptedToRemoveProtected)
		log.Printf("[RoleUsecase] UpdatePermissionsDiff attempt_remove_protected tenant_id=%s role_id=%s role_code=%s codes=%v", tenantID, roleID, role.Code, attemptedToRemoveProtected)
		return fmt.Errorf("this role (%s) requires certain permissions to remain enabled: %s", role.Code, strings.Join(attemptedToRemoveProtected, ", "))
	}

	// Guarantee minimum required permissions for special roles only (query database if needed)
	// Non-special roles can remove all permissions, even mandatory ones.
	if isSpecialRole(role.Code) {
		minimumCodes := map[string]bool{
			"user.read":       true,
			"role.read":       true,
			"permission.read": true,
		}

		// Get existing permission IDs for minimum codes from current set
		for _, rp := range nextSet {
			if rp.Permission != nil && minimumCodes[rp.Permission.Code] {
				delete(minimumCodes, rp.Permission.Code) // Already have it
			}
		}

		// If we're missing any minimum permissions, query database to find their IDs.
		// Must use u.db directly (not GetDB) because permissions is a global/shared table with no tenant_id.
		if len(minimumCodes) > 0 {
			var perms []struct {
				ID   string
				Code string
			}

			codesToFind := make([]string, 0, len(minimumCodes))
			for code := range minimumCodes {
				codesToFind = append(codesToFind, code)
			}

			if err := u.db.WithContext(ctx).
				Table("permissions").
				Select("id, code").
				Where("code IN ?", codesToFind).
				Scan(&perms).Error; err == nil {
				for _, p := range perms {
					nextSet[p.ID] = models.RolePermission{
						RoleID:       roleID,
						PermissionID: p.ID,
						Scope:        models.ScopeAll,
					}
				}
			}
		}
	}

	// Revalidate final permission IDs against source-of-truth permissions table.
	// This removes stale IDs from legacy rows while still rejecting explicitly invalid user additions.
	finalIDs := make([]string, 0, len(nextSet))
	for permID := range nextSet {
		finalIDs = append(finalIDs, permID)
	}

	if len(finalIDs) > 0 {
		var existingIDs []string
		// Must use u.db directly — permissions is a global/shared table with no tenant_id column.
		// Using GetDB here would inject WHERE tenant_id=? causing all IDs to appear non-existent,
		// which would silently delete the entire permission set.
		if err := u.db.WithContext(ctx).
			Table("permissions").
			Where("id IN ? AND deleted_at IS NULL", finalIDs).
			Pluck("id", &existingIDs).Error; err != nil {
			return err
		}

		existingIDSet := make(map[string]bool, len(existingIDs))
		for _, id := range existingIDs {
			existingIDSet[id] = true
		}

		addedSet := make(map[string]bool, len(diff.Added)+len(diff.AddedWithScope))
		for _, id := range diff.Added {
			addedSet[id] = true
		}
		for _, a := range diff.AddedWithScope {
			addedSet[a.PermissionID] = true
		}

		missingAdded := make([]string, 0)
		filteredStale := make([]string, 0)
		for _, id := range finalIDs {
			if existingIDSet[id] {
				continue
			}
			if addedSet[id] {
				missingAdded = append(missingAdded, id)
				continue
			}
			filteredStale = append(filteredStale, id)
			delete(nextSet, id)
		}

		if len(missingAdded) > 0 {
			sort.Strings(missingAdded)
			log.Printf("[RoleUsecase] UpdatePermissionsDiff invalid_added_permissions tenant_id=%s role_id=%s permission_ids=%v", tenantID, roleID, missingAdded)
			return fmt.Errorf("invalid permission IDs in diff.added: %v", missingAdded)
		}

		if len(filteredStale) > 0 {
			sort.Strings(filteredStale)
			log.Printf("[RoleUsecase] UpdatePermissionsDiff filtered_stale_permissions tenant_id=%s role_id=%s permission_ids=%v", tenantID, roleID, filteredStale)
		}
	}

	// Convert map back to slice
	final := make([]models.RolePermission, 0, len(nextSet))
	for _, rp := range nextSet {
		final = append(final, rp)
	}

	finalIDs = make([]string, 0, len(final))
	for _, rp := range final {
		finalIDs = append(finalIDs, rp.PermissionID)
	}

	if err := u.ensureSystemAdminOnlyPermissions(ctx, finalIDs); err != nil {
		return err
	}

	// Apply the change
	if err := u.roleRepo.AssignPermissionsWithScope(ctx, roleID, final); err != nil {
		log.Printf("[RoleUsecase] UpdatePermissionsDiff assign_failed tenant_id=%s role_id=%s err=%v", tenantID, roleID, err)
		return err
	}

	u.invalidatePermissionCaches(ctx, roleID, role.Code)

	// Publish event with all final permission IDs
	permIDs := make([]string, 0, len(final))
	for _, rp := range final {
		permIDs = append(permIDs, rp.PermissionID)
	}

	u.eventPublisher.PublishAsync(ctx, events.NewRolePermissionsAssignedEvent(ctx, events.RolePermissionsAssignedPayload{
		RoleID:        role.ID,
		PermissionIDs: permIDs,
		AssignedAt:    apptime.Now(),
	}))

	return nil
}

func (u *roleUsecase) GetMenuAccess(ctx context.Context, roleID string) ([]dto.RoleMenuAccessResponse, error) {
	if _, err := u.roleRepo.FindByID(ctx, roleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	rows := make([]models.RoleMenuAccess, 0)
	err := coreDB.GetDB(ctx, u.db).
		Where("role_id = ? AND is_enabled = true", roleID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]dto.RoleMenuAccessResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.RoleMenuAccessResponse{
			MenuID:    row.MenuID,
			Scope:     row.Scope,
			IsEnabled: row.IsEnabled,
		})
	}

	return result, nil
}

func (u *roleUsecase) UpdateMenuAccess(ctx context.Context, roleID string, assignments []dto.RoleMenuAccessAssignment) error {
	role, err := u.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	normalizedAssignments := make([]dto.RoleMenuAccessAssignment, 0, len(assignments))
	seen := make(map[string]struct{}, len(assignments))
	menuIDs := make([]string, 0, len(assignments))
	for _, assignment := range assignments {
		menuID := strings.TrimSpace(assignment.MenuID)
		if menuID == "" {
			continue
		}
		if _, exists := seen[menuID]; exists {
			continue
		}
		scope := strings.ToUpper(strings.TrimSpace(assignment.Scope))
		if scope == "" {
			scope = models.ScopeAll
		}
		if !models.IsValidScope(scope) {
			return fmt.Errorf("invalid scope '%s' for menu '%s'", assignment.Scope, menuID)
		}

		seen[menuID] = struct{}{}
		normalizedAssignments = append(normalizedAssignments, dto.RoleMenuAccessAssignment{
			MenuID: menuID,
			Scope:  scope,
		})
		menuIDs = append(menuIDs, menuID)
	}

	if len(menuIDs) > 0 {
		var validCount int64
		err = u.db.WithContext(ctx).
			Table("menus").
			Where("id IN ? AND deleted_at IS NULL", menuIDs).
			Count(&validCount).Error
		if err != nil {
			return err
		}
		if int(validCount) != len(menuIDs) {
			return repositories.ErrInvalidPermissionIDs
		}
	}

	return coreDB.GetDB(ctx, u.db).Transaction(func(tx *gorm.DB) error {
		txCtx := coreDB.WithTx(ctx, tx)

		if err := tx.Exec("DELETE FROM role_menu_access WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}

		if len(normalizedAssignments) > 0 {
			rows := make([]models.RoleMenuAccess, 0, len(normalizedAssignments))
			for _, assignment := range normalizedAssignments {
				rows = append(rows, models.RoleMenuAccess{
					RoleID:    roleID,
					MenuID:    assignment.MenuID,
					Scope:     assignment.Scope,
					IsEnabled: true,
					TenantID:  role.TenantID,
				})
			}
			if err := tx.Create(&rows).Error; err != nil {
				return err
			}
		}

		roleAssignments, permissionIDs, err := u.deriveRolePermissionAssignmentsFromMenus(txCtx, roleID, normalizedAssignments)
		if err != nil {
			return err
		}

		if err := u.validatePlanPermissionBoundary(txCtx, permissionIDs); err != nil {
			return err
		}

		if err := u.ensureSystemAdminOnlyPermissions(txCtx, permissionIDs); err != nil {
			return err
		}

		if err := u.roleRepo.AssignPermissionsWithScope(txCtx, roleID, roleAssignments); err != nil {
			return err
		}

		// Invalidate caches after successful update
		u.invalidatePermissionCaches(txCtx, roleID, role.Code)

		return nil
	})
}

func (u *roleUsecase) ensureSystemAdminOnlyPermissions(ctx context.Context, permissionIDs []string) error {
	if len(permissionIDs) == 0 || middleware.IsSystemAdmin(ctx) {
		return nil
	}

	var blocked []string
	if err := u.db.WithContext(ctx).
		Table("permissions").
		Where("id IN ? AND code IN ?", permissionIDs, systemAdminOnlyPermissionCodes).
		Pluck("code", &blocked).Error; err != nil {
		return err
	}

	if len(blocked) == 0 {
		return nil
	}

	sort.Strings(blocked)
	return fmt.Errorf("permissions reserved for system admins: %s", strings.Join(blocked, ", "))
}

func (u *roleUsecase) validatePlanPermissionBoundary(ctx context.Context, permissionIDs []string) error {
	if len(permissionIDs) == 0 || u.db == nil || u.planRepo == nil {
		return nil
	}

	// System admins are platform-wide operators and are not constrained by tenant plans.
	if middleware.IsSystemAdmin(ctx) {
		return nil
	}

	tenantID := middleware.TenantFromContext(ctx)
	if tenantID == "" {
		return nil
	}

	planSlug, err := u.resolveTenantPlanSlug(ctx, tenantID)
	if err != nil || planSlug == "" {
		log.Printf("[RoleUsecase] validatePlanPermissionBoundary no_active_plan tenant_id=%s err=%v", tenantID, err)
		return ErrPermissionNotAllowedForPlan
	}

	modules, err := u.planRepo.GetEnabledModules(ctx, planSlug)
	if err != nil {
		log.Printf("[RoleUsecase] validatePlanPermissionBoundary modules_load_failed tenant_id=%s plan=%s err=%v", tenantID, planSlug, err)
		return ErrPermissionNotAllowedForPlan
	}
	log.Printf("[RoleUsecase] validatePlanPermissionBoundary tenant_id=%s plan=%s requested_permissions=%d enabled_modules=%d", tenantID, planSlug, len(permissionIDs), len(modules))
	allowedModules := make(map[string]struct{}, len(modules))
	for _, m := range modules {
		allowedModules[strings.ToLower(strings.TrimSpace(m))] = struct{}{}
	}

	permissionMeta, err := u.fetchPermissionMeta(ctx, permissionIDs)
	if err != nil {
		return err
	}

	policy, err := u.loadPlanPermissionPolicy(ctx, planSlug)
	if err != nil {
		return err
	}

	if len(permissionMeta) == 0 {
		return nil
	}

	for _, p := range permissionMeta {
		codeLower := strings.ToLower(p.Code)
		if strings.HasPrefix(codeLower, "dashboard.") ||
			strings.HasPrefix(codeLower, "profile.") ||
			strings.HasPrefix(codeLower, "setting") ||
			strings.HasPrefix(codeLower, "billing.") ||
			codeLower == "pos.payment.manage" ||
			strings.HasPrefix(codeLower, "user.") ||
			strings.HasPrefix(codeLower, "role.") ||
			strings.HasPrefix(codeLower, "permission.") ||
			strings.HasPrefix(codeLower, "company.") ||
			strings.HasPrefix(codeLower, "warehouse.") ||
			strings.HasPrefix(codeLower, "product.") ||
			strings.HasPrefix(codeLower, "customer.") ||
			strings.HasPrefix(codeLower, "supplier.") ||
			strings.HasPrefix(codeLower, "employee.") {
			continue
		}

		module := normalizeMenuModule(p.MenuModule)
		if module == "" {
			module = moduleFromURL(p.MenuURL)
		}
		if module == "" {
			continue
		}
		if _, ok := allowedModules[module]; !ok {
			log.Printf("[RoleUsecase] validatePlanPermissionBoundary blocked_by_module tenant_id=%s plan=%s permission_code=%s module=%s", tenantID, planSlug, p.Code, module)
			return ErrPermissionNotAllowedForPlan
		}
		if policy.hasRules && !isPermissionAllowedByPolicy(p, policy) {
			log.Printf("[RoleUsecase] validatePlanPermissionBoundary blocked_by_policy tenant_id=%s plan=%s permission_code=%s module=%s", tenantID, planSlug, p.Code, module)
			return ErrPermissionNotAllowedForPlan
		}
	}

	return nil
}

func (u *roleUsecase) resolveTenantPlanSlug(ctx context.Context, tenantID string) (string, error) {
	var planSlug string
	err := u.db.WithContext(ctx).
		Table("tenant_subscriptions").
		Select("plan").
		Where("tenant_id = ? AND status IN ('active','trial') AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Limit(1).
		Row().Scan(&planSlug)
	if err != nil {
		return "", err
	}
	return normalizeTenantPlanSlug(planSlug), nil
}

func normalizeTenantPlanSlug(planSlug string) string {
	normalized := strings.ToLower(strings.TrimSpace(planSlug))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")

	switch normalized {
	case "pos", "pos_modular":
		return "pos_growth"
	case "erp", "erp_modular":
		return "erp_pro"
	case "crm", "crm_modular":
		return "crm_growth"
	case "hr", "hr_modular":
		return "hr_growth"
	default:
		return normalized
	}
}

type permissionMetaRow struct {
	Code       string
	MenuURL    string
	MenuModule string
}

type permissionMetaRowWithID struct {
	ID         string
	Code       string
	MenuURL    string
	MenuModule string
}

type planPermissionPolicy struct {
	codeSet      map[string]struct{}
	menuPrefixes []string
	hasRules     bool
}

func (u *roleUsecase) fetchPermissionMeta(ctx context.Context, permissionIDs []string) ([]permissionMetaRow, error) {
	uniq := make([]string, 0, len(permissionIDs))
	seen := make(map[string]struct{}, len(permissionIDs))
	for _, id := range permissionIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}

	if len(uniq) == 0 {
		return []permissionMetaRow{}, nil
	}

	rows := make([]permissionMetaRow, 0, len(uniq))
	err := u.db.WithContext(ctx).
		Table("permissions p").
		Select("p.code AS code, COALESCE(m.url, '') AS menu_url, COALESCE(m.module, '') AS menu_module").
		Joins("LEFT JOIN menus m ON m.id = p.menu_id").
		Where("p.id IN ? AND p.deleted_at IS NULL", uniq).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (u *roleUsecase) fetchPermissionMetaWithID(ctx context.Context, permissionIDs []string) ([]permissionMetaRowWithID, error) {
	uniq := make([]string, 0, len(permissionIDs))
	seen := make(map[string]struct{}, len(permissionIDs))
	for _, id := range permissionIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}

	if len(uniq) == 0 {
		return []permissionMetaRowWithID{}, nil
	}

	rows := make([]permissionMetaRowWithID, 0, len(uniq))
	err := u.db.WithContext(ctx).
		Table("permissions p").
		Select("p.id AS id, p.code AS code, COALESCE(m.url, '') AS menu_url, COALESCE(m.module, '') AS menu_module").
		Joins("LEFT JOIN menus m ON m.id = p.menu_id").
		Where("p.id IN ? AND p.deleted_at IS NULL", uniq).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (u *roleUsecase) loadPlanPermissionPolicy(ctx context.Context, planSlug string) (planPermissionPolicy, error) {
	policy := planPermissionPolicy{
		codeSet:      map[string]struct{}{},
		menuPrefixes: []string{},
		hasRules:     false,
	}
	normalizedPlanSlug := normalizeTenantPlanSlug(planSlug)

	rows := make([]struct {
		PermissionCode string
		MenuURL        string
	}, 0)

	err := u.db.WithContext(ctx).
		Table("plan_permission_entitlements").
		Select("permission_code, menu_url").
		Where("plan_slug = ? AND is_enabled = true", planSlug).
		Scan(&rows).Error
	if err != nil {
		return policy, err
	}
	defaultRows := tenantPolicy.DefaultPlanEntitlementRows(planSlug)
	if len(defaultRows) > 0 {
		merged := make([]struct {
			PermissionCode string
			MenuURL        string
		}, 0, len(rows)+len(defaultRows))
		merged = append(merged, rows...)
		for _, row := range defaultRows {
			merged = append(merged, struct {
				PermissionCode string
				MenuURL        string
			}{
				PermissionCode: row.PermissionCode,
				MenuURL:        row.MenuURL,
			})
		}
		rows = merged
	}

	for _, row := range rows {
		code := strings.ToLower(strings.TrimSpace(row.PermissionCode))
		menuURL := strings.ToLower(strings.TrimSpace(row.MenuURL))
		if isBlockedPOSBroadMenuURL(normalizedPlanSlug, menuURL) {
			continue
		}
		if code != "" {
			policy.codeSet[code] = struct{}{}
			policy.hasRules = true
		}
		if menuURL != "" {
			policy.menuPrefixes = append(policy.menuPrefixes, menuURL)
			policy.hasRules = true
		}
	}

	return policy, nil
}

func isBlockedPOSBroadMenuURL(normalizedPlanSlug, menuURL string) bool {
	if normalizedPlanSlug != "pos_growth" {
		return false
	}
	switch menuURL {
	case "/sales", "/purchase", "/stock", "/finance":
		return true
	default:
		return false
	}
}

func isPermissionAllowedByPolicy(p permissionMetaRow, policy planPermissionPolicy) bool {
	code := strings.ToLower(strings.TrimSpace(p.Code))
	if code != "" {
		if _, ok := policy.codeSet[code]; ok {
			return true
		}
	}

	menuURL := strings.ToLower(strings.TrimSpace(p.MenuURL))
	if menuURL == "" {
		return true
	}

	for _, prefix := range policy.menuPrefixes {
		// "/" means allow all URLs — used by ultimate_suite and enterprise plans.
		if prefix == "/" {
			return true
		}
		if menuURL == prefix || strings.HasPrefix(menuURL, prefix+"/") {
			return true
		}
	}

	return false
}

func normalizeMenuModule(module string) string {
	m := strings.ToLower(strings.TrimSpace(module))
	switch m {
	case "hrd", "human-resources", "human_resources":
		return "hr"
	case "master_data", "master-data":
		return "core"
	case "fb", "feedback", "loyalty":
		return "pos"
	case "stock":
		return "inventory"
	default:
		return m
	}
}

func moduleFromURL(url string) string {
	p := strings.ToLower(strings.TrimSpace(url))
	if p == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(p, "/hrd"):
		return "hr"
	case strings.HasPrefix(p, "/stock"):
		return "inventory"
	case strings.HasPrefix(p, "/pos"):
		return "pos"
	case strings.HasPrefix(p, "/crm"):
		return "crm"
	case strings.HasPrefix(p, "/sales"):
		return "sales"
	case strings.HasPrefix(p, "/purchase"):
		return "purchase"
	case strings.HasPrefix(p, "/finance"):
		return "finance"
	case strings.HasPrefix(p, "/master-data"):
		return "core"
	default:
		return ""
	}
}

// invalidatePermissionCaches clears role and permission caches after assignment changes.
// Both L1 (in-memory via PermissionService) and L2 (Redis) caches are invalidated
// to ensure permission/scope changes take effect immediately.
func (u *roleUsecase) invalidatePermissionCaches(ctx context.Context, roleID string, roleCode string) {
	u.redis.Del(ctx,
		roleByIDCacheKey(ctx, roleID),
		fmt.Sprintf(cacheRoleByIDKeyLegacy, roleID),
		roleListCacheKey(ctx, 1, 10, ""),
		roleListCacheKey(ctx, 1, 20, ""),
		cacheRoleListPage1Limit10,
		cacheRoleListPage1Limit20,
	)

	// Invalidate L1 in-memory cache in PermissionService (immediate effect)
	if u.permService != nil {
		if err := u.permService.InvalidateCache(roleCode); err != nil {
			log.Printf("[RoleUsecase] failed to invalidate L1 permission cache for role '%s': %v", roleCode, err)
		}
	}

	// Invalidate L2 Redis permission caches (correct key patterns)
	for _, pattern := range []string{"permissions:*", "permissions_scope:*"} {
		iter := u.redis.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			u.redis.Del(ctx, iter.Val())
		}
	}
}

func (u *roleUsecase) ValidateUserRole(ctx context.Context, userID string, roleID string) (bool, error) {
	// Check if role exists and is active
	r, err := u.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// Check if role is active
	if r.Status != "active" {
		return false, nil
	}

	return true, nil
}

// FilterAllowedPermissions returns which permission IDs are allowed and which are blocked by the plan.
// This is used for tolerant mode assignment where we drop blocked IDs instead of rejecting the entire request.
func (u *roleUsecase) FilterAllowedPermissions(ctx context.Context, permissionIDs []string) (allowed []string, blocked []string, err error) {
	if len(permissionIDs) == 0 {
		return []string{}, []string{}, nil
	}

	// System admins are not constrained by plan
	if middleware.IsSystemAdmin(ctx) {
		return permissionIDs, []string{}, nil
	}

	tenantID := middleware.TenantFromContext(ctx)
	if tenantID == "" {
		// If no tenant context, all permissions are allowed (fallback)
		return permissionIDs, []string{}, nil
	}

	planSlug, err := u.resolveTenantPlanSlug(ctx, tenantID)
	if err != nil || planSlug == "" {
		// If we can't resolve plan, block all
		return []string{}, permissionIDs, nil
	}

	modules, err := u.planRepo.GetEnabledModules(ctx, planSlug)
	if err != nil {
		// If we can't get modules, block all
		return []string{}, permissionIDs, nil
	}
	allowedModules := make(map[string]struct{}, len(modules))
	for _, m := range modules {
		allowedModules[strings.ToLower(strings.TrimSpace(m))] = struct{}{}
	}

	permissionMeta, err := u.fetchPermissionMetaWithID(ctx, permissionIDs)
	if err != nil {
		// If we can't fetch metadata, allow all (assume permissions are valid)
		return permissionIDs, []string{}, nil
	}

	policy, err := u.loadPlanPermissionPolicy(ctx, planSlug)
	if err != nil {
		// If we can't load policy, allow all (assume no strict policy)
		return permissionIDs, []string{}, nil
	}

	allowedList := make([]string, 0)
	blockedList := make([]string, 0)

	for _, meta := range permissionMeta {
		// Check if permission is in always-allowed list
		codeLower := strings.ToLower(meta.Code)
		if strings.HasPrefix(codeLower, "dashboard.") ||
			strings.HasPrefix(codeLower, "profile.") ||
			strings.HasPrefix(codeLower, "setting") ||
			strings.HasPrefix(codeLower, "billing.") ||
			codeLower == "pos.payment.manage" ||
			strings.HasPrefix(codeLower, "user.") ||
			strings.HasPrefix(codeLower, "role.") ||
			strings.HasPrefix(codeLower, "permission.") ||
			strings.HasPrefix(codeLower, "company.") ||
			strings.HasPrefix(codeLower, "warehouse.") ||
			strings.HasPrefix(codeLower, "product.") ||
			strings.HasPrefix(codeLower, "customer.") ||
			strings.HasPrefix(codeLower, "supplier.") ||
			strings.HasPrefix(codeLower, "employee.") {
			allowedList = append(allowedList, meta.ID)
			continue
		}

		// Check module entitlement
		module := normalizeMenuModule(meta.MenuModule)
		if module == "" {
			module = moduleFromURL(meta.MenuURL)
		}
		if module == "" {
			// No module detected, allow it
			allowedList = append(allowedList, meta.ID)
			continue
		}

		if _, ok := allowedModules[module]; !ok {
			// Module not allowed
			blockedList = append(blockedList, meta.ID)
			continue
		}

		// Check granular policy
		if policy.hasRules && !isPermissionAllowedByPolicy(permissionMetaRow{
			Code:       meta.Code,
			MenuURL:    meta.MenuURL,
			MenuModule: meta.MenuModule,
		}, policy) {
			blockedList = append(blockedList, meta.ID)
			continue
		}

		// Permission is allowed
		allowedList = append(allowedList, meta.ID)
	}

	return allowedList, blockedList, nil
}

func (u *roleUsecase) deriveRolePermissionAssignmentsFromMenus(ctx context.Context, roleID string, menuAssignments []dto.RoleMenuAccessAssignment) ([]models.RolePermission, []string, error) {
	type menuRow struct {
		ID       string
		ParentID *string
	}
	type permissionRow struct {
		ID     string
		Code   string
		MenuID *string
	}

	menuRows := make([]menuRow, 0)
	if err := u.db.WithContext(ctx).
		Table("menus").
		Select("id, parent_id").
		Where("deleted_at IS NULL").
		Scan(&menuRows).Error; err != nil {
		return nil, nil, err
	}

	parentByMenuID := make(map[string]string, len(menuRows))
	childrenByMenuID := make(map[string][]string, len(menuRows))
	for _, row := range menuRows {
		if row.ParentID != nil && *row.ParentID != "" {
			parentByMenuID[row.ID] = *row.ParentID
			childrenByMenuID[*row.ParentID] = append(childrenByMenuID[*row.ParentID], row.ID)
		}
	}

	scopeBySelectedMenuID := make(map[string]string, len(menuAssignments))
	for _, assignment := range menuAssignments {
		scopeBySelectedMenuID[assignment.MenuID] = assignment.Scope
	}

	allowedMenuIDs := make(map[string]struct{}, len(menuAssignments))
	queue := make([]string, 0, len(menuAssignments))
	for _, assignment := range menuAssignments {
		if _, exists := allowedMenuIDs[assignment.MenuID]; exists {
			continue
		}
		allowedMenuIDs[assignment.MenuID] = struct{}{}
		queue = append(queue, assignment.MenuID)
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, childID := range childrenByMenuID[current] {
			if _, exists := allowedMenuIDs[childID]; exists {
				continue
			}
			allowedMenuIDs[childID] = struct{}{}
			queue = append(queue, childID)
		}
	}

	permissionRows := make([]permissionRow, 0)
	if err := u.db.WithContext(ctx).
		Table("permissions").
		Select("id, code, menu_id").
		Where("deleted_at IS NULL").
		Scan(&permissionRows).Error; err != nil {
		return nil, nil, err
	}

	permissionScopeMap := make(map[string]string)
	for _, permissionRow := range permissionRows {
		codeLower := strings.ToLower(strings.TrimSpace(permissionRow.Code))
		if isAlwaysSafePermissionCode(codeLower) {
			permissionScopeMap[permissionRow.ID] = models.ScopeAll
			continue
		}
		if permissionRow.MenuID == nil || *permissionRow.MenuID == "" {
			continue
		}
		if _, allowed := allowedMenuIDs[*permissionRow.MenuID]; !allowed {
			continue
		}

		scope := resolveScopeForMenu(*permissionRow.MenuID, parentByMenuID, scopeBySelectedMenuID)
		permissionScopeMap[permissionRow.ID] = scope
	}

	roleAssignments := make([]models.RolePermission, 0, len(permissionScopeMap))
	permissionIDs := make([]string, 0, len(permissionScopeMap))
	for permissionID, scope := range permissionScopeMap {
		roleAssignments = append(roleAssignments, models.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
			Scope:        scope,
		})
		permissionIDs = append(permissionIDs, permissionID)
	}

	return roleAssignments, permissionIDs, nil
}

func resolveScopeForMenu(menuID string, parentByMenuID map[string]string, scopeBySelectedMenuID map[string]string) string {
	current := menuID
	for current != "" {
		if scope, exists := scopeBySelectedMenuID[current]; exists && scope != "" {
			return scope
		}
		parentID, exists := parentByMenuID[current]
		if !exists {
			break
		}
		current = parentID
	}
	return models.ScopeAll
}

func isAlwaysSafePermissionCode(code string) bool {
	if code == "pos.payment.manage" {
		return true
	}
	for _, prefix := range alwaysSafePermissionPrefixes {
		if strings.HasPrefix(code, prefix) {
			return true
		}
	}
	return false
}

// isSpecialRole returns true if the role should have protected "always safe" permissions.
// Special roles (owner, admin, managers) must keep mandatory permissions enabled.
// Regular roles can freely remove permissions.
func isSpecialRole(roleCode string) bool {
	specialRoles := map[string]bool{
		"tenant_owner": true, // Account owner
		"admin":        true, // Full system admin
	}
	return specialRoles[roleCode]
}

func (u *roleUsecase) withPreservedAlwaysSafePermissionIDs(ctx context.Context, roleCode, roleID string, permissionIDs []string) ([]string, error) {
	seen := make(map[string]struct{}, len(permissionIDs))
	result := make([]string, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		if permissionID == "" {
			continue
		}
		if _, exists := seen[permissionID]; exists {
			continue
		}
		seen[permissionID] = struct{}{}
		result = append(result, permissionID)
	}

	if !isSpecialRole(roleCode) {
		return result, nil
	}

	alwaysSafeRows, err := u.loadExistingAlwaysSafeRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	mandatoryPermissionIDs, err := u.loadMandatoryPermissionIDs(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range alwaysSafeRows {
		if row.PermissionID == "" {
			continue
		}
		if _, exists := seen[row.PermissionID]; exists {
			continue
		}
		seen[row.PermissionID] = struct{}{}
		result = append(result, row.PermissionID)
	}
	for _, permissionID := range mandatoryPermissionIDs {
		if permissionID == "" {
			continue
		}
		if _, exists := seen[permissionID]; exists {
			continue
		}
		seen[permissionID] = struct{}{}
		result = append(result, permissionID)
	}

	return result, nil
}

func (u *roleUsecase) withPreservedAlwaysSafeRolePermissions(ctx context.Context, roleCode, roleID string, rolePerms []models.RolePermission) ([]models.RolePermission, error) {
	byPermissionID := make(map[string]models.RolePermission, len(rolePerms))
	for _, rolePerm := range rolePerms {
		if rolePerm.PermissionID == "" {
			continue
		}
		if rolePerm.Scope == "" {
			rolePerm.Scope = models.ScopeAll
		}
		byPermissionID[rolePerm.PermissionID] = rolePerm
	}

	if !isSpecialRole(roleCode) {
		result := make([]models.RolePermission, 0, len(byPermissionID))
		for _, rolePerm := range byPermissionID {
			result = append(result, rolePerm)
		}
		return result, nil
	}

	alwaysSafeRows, err := u.loadExistingAlwaysSafeRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	mandatoryPermissionIDs, err := u.loadMandatoryPermissionIDs(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range alwaysSafeRows {
		if row.PermissionID == "" {
			continue
		}
		if _, exists := byPermissionID[row.PermissionID]; exists {
			continue
		}
		scope := row.Scope
		if scope == "" {
			scope = models.ScopeAll
		}
		byPermissionID[row.PermissionID] = models.RolePermission{
			RoleID:       roleID,
			PermissionID: row.PermissionID,
			Scope:        scope,
		}
	}
	for _, permissionID := range mandatoryPermissionIDs {
		if permissionID == "" {
			continue
		}
		if _, exists := byPermissionID[permissionID]; exists {
			continue
		}
		byPermissionID[permissionID] = models.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
			Scope:        models.ScopeAll,
		}
	}

	result := make([]models.RolePermission, 0, len(byPermissionID))
	for _, rolePerm := range byPermissionID {
		result = append(result, rolePerm)
	}

	return result, nil
}

func (u *roleUsecase) loadExistingAlwaysSafeRolePermissions(ctx context.Context, roleID string) ([]models.RolePermission, error) {
	type rolePermissionRow struct {
		PermissionID string
		Scope        string
		Code         string
	}

	rows := make([]rolePermissionRow, 0)
	err := u.db.WithContext(ctx).
		Table("role_permissions rp").
		Select("rp.permission_id, rp.scope, p.code").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Where("rp.role_id = ? AND p.deleted_at IS NULL", roleID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]models.RolePermission, 0, len(rows))
	for _, row := range rows {
		if !isAlwaysSafePermissionCode(strings.ToLower(strings.TrimSpace(row.Code))) {
			continue
		}
		scope := strings.ToUpper(strings.TrimSpace(row.Scope))
		if scope == "" {
			scope = models.ScopeAll
		}
		result = append(result, models.RolePermission{
			RoleID:       roleID,
			PermissionID: row.PermissionID,
			Scope:        scope,
		})
	}

	return result, nil
}

func (u *roleUsecase) loadMandatoryPermissionIDs(ctx context.Context) ([]string, error) {
	ids := make([]string, 0, len(mandatoryPermissionCodes))
	err := u.db.WithContext(ctx).
		Table("permissions").
		Where("deleted_at IS NULL AND code IN ?", mandatoryPermissionCodes).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}
