package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
	depreciationsvc "github.com/gilabs/gims/api/internal/finance/domain/service"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FieldGroup represents the classification of an editable field
type FieldGroup string

const (
	GroupA FieldGroup = "A" // Non-accounting (free to change)
	GroupB FieldGroup = "B" // Accounting-sensitive (need recalc & confirmation)
	GroupC FieldGroup = "C" // Immutable after active status
)

// FieldClassification tracks which group a field belongs to
var fieldToGroup = map[string]FieldGroup{
	// Group A: Non-accounting fields
	"name":                    GroupA,
	"description":             GroupA,
	"status":                  GroupA,
	"asset_tag":               GroupA,
	"location_id":             GroupA,
	"department_id":           GroupA,
	"assigned_to_employee_id":  GroupA,
	"custodian_user_id":       GroupA,
	"serial_number":           GroupA,
	"barcode":                 GroupA,
	"vendor_id":               GroupA,
	"purchase_invoice_id":     GroupA,
	"warranty_start":          GroupA,
	"warranty_end":            GroupA,
	"warranty_provider":       GroupA,
	"warranty_terms":          GroupA,
	"insurance_policy_number": GroupA,
	"insurance_provider":      GroupA,
	"insurance_start":         GroupA,
	"insurance_end":           GroupA,
	"insurance_value":         GroupA,
	"shipping_cost":           GroupA,
	"installation_cost":       GroupA,
	"tax_amount":              GroupA,
	"other_costs":             GroupA,

	// Group B: Accounting-sensitive
	"salvage_value":        GroupB,
	"useful_life_months":   GroupB,
	"depreciation_method":  GroupB,
	"category_id":          GroupB,
	"acquisition_cost":     GroupB,

	// Group C: Immutable after active
	"code":              GroupC,
	"acquisition_date":  GroupC,
	"asset_type_id":     GroupC,
}

// FieldChanges holds the changes classified by group
type FieldChanges struct {
	GroupA map[string]interface{}
	GroupB map[string]interface{}
	GroupC map[string]interface{}
}

// NewFieldChanges creates an empty field changes map
func NewFieldChanges() *FieldChanges {
	return &FieldChanges{
		GroupA: make(map[string]interface{}),
		GroupB: make(map[string]interface{}),
		GroupC: make(map[string]interface{}),
	}
}

// ClassifyChanges compares old and new asset values and classifies them
func ClassifyChanges(existing *financeModels.Asset, req *dto.EditAssetRequest) (*FieldChanges, error) {
	changes := NewFieldChanges()

	// Helper function to check if field changed
	checkAndClassify := func(fieldName string, oldVal interface{}, newVal interface{}) {
		oldVal = normalizeComparableValue(oldVal)
		newVal = normalizeComparableValue(newVal)

		// Compare values
		changed := false
		if oldVal != newVal {
			// Additional null check
			oldIsNil := oldVal == nil || (fmt.Sprintf("%v", oldVal) == "")
			newIsNil := newVal == nil || (fmt.Sprintf("%v", newVal) == "")
			if oldIsNil != newIsNil {
				changed = true
			} else if !oldIsNil && !newIsNil {
				changed = fmt.Sprintf("%v", oldVal) != fmt.Sprintf("%v", newVal)
			}
		}

		if !changed {
			return
		}

		group := fieldToGroup[fieldName]
		switch group {
		case GroupA:
			changes.GroupA[fieldName] = map[string]interface{}{
				"old": oldVal,
				"new": newVal,
			}
		case GroupB:
			changes.GroupB[fieldName] = map[string]interface{}{
				"old": oldVal,
				"new": newVal,
			}
		case GroupC:
			changes.GroupC[fieldName] = map[string]interface{}{
				"old": oldVal,
				"new": newVal,
			}
		}
	}

	// Check all fields
	checkAndClassify("code", existing.Code, req.Code)
	checkAndClassify("name", existing.Name, req.Name)
	checkAndClassify("description", existing.Description, req.Description)
	checkAndClassify("status", existing.Status, req.Status)
	checkAndClassify("asset_type_id", existing.AssetTypeID, &req.AssetTypeID)
	checkAndClassify("category_id", existing.CategoryID, req.CategoryID)
	checkAndClassify("location_id", existing.LocationID, req.LocationID)
	checkAndClassify("acquisition_date", existing.AcquisitionDate.Format("2006-01-02"), req.AcquisitionDate)
	checkAndClassify("acquisition_cost", existing.AcquisitionCost, req.AcquisitionCost)
	checkAndClassify("salvage_value", existing.SalvageValue, req.SalvageValue)
	checkAndClassify("useful_life_months", existing.UsefulLifeMonths, req.UsefulLifeMonths)
	checkAndClassify("depreciation_method", existing.DepreciationMethod, req.DepreciationMethod)
	checkAndClassify("vendor_id", existing.SupplierID, req.VendorID)
	checkAndClassify("purchase_invoice_id", existing.SupplierInvoiceID, req.PurchaseInvoiceID)
	checkAndClassify("department_id", existing.DepartmentID, req.DepartmentID)
	checkAndClassify("assigned_to_employee_id", existing.AssignedToEmployeeID, req.AssignedToEmployeeID)
	checkAndClassify("custodian_user_id", existing.CustodianUserID, req.CustodianUserID)
	checkAndClassify("serial_number", existing.SerialNumber, req.SerialNumber)
	checkAndClassify("barcode", existing.Barcode, req.Barcode)
	checkAndClassify("asset_tag", existing.AssetTag, req.AssetTag)
	checkAndClassify("shipping_cost", existing.ShippingCost, req.ShippingCost)
	checkAndClassify("installation_cost", existing.InstallationCost, req.InstallationCost)
	checkAndClassify("tax_amount", existing.TaxAmount, req.TaxAmount)
	checkAndClassify("other_costs", existing.OtherCosts, req.OtherCosts)
	checkAndClassify("warranty_start", existing.WarrantyStart, parseStringToTime(req.WarrantyStart))
	checkAndClassify("warranty_end", existing.WarrantyEnd, parseStringToTime(req.WarrantyEnd))
	checkAndClassify("warranty_provider", existing.WarrantyProvider, req.WarrantyProvider)
	checkAndClassify("warranty_terms", existing.WarrantyTerms, req.WarrantyTerms)
	checkAndClassify("insurance_policy_number", existing.InsurancePolicyNumber, req.InsurancePolicyNumber)
	checkAndClassify("insurance_provider", existing.InsuranceProvider, req.InsuranceProvider)
	checkAndClassify("insurance_start", existing.InsuranceStart, parseStringToTime(req.InsuranceStart))
	checkAndClassify("insurance_end", existing.InsuranceEnd, parseStringToTime(req.InsuranceEnd))
	checkAndClassify("insurance_value", existing.InsuranceValue, req.InsuranceValue)

	return changes, nil
}

func normalizeComparableValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case *string:
		if typed == nil {
			return nil
		}
		return strings.TrimSpace(*typed)
	case *int:
		if typed == nil {
			return nil
		}
		return *typed
	case *float64:
		if typed == nil {
			return nil
		}
		return *typed
	case *time.Time:
		if typed == nil || typed.IsZero() {
			return nil
		}
		return typed.Format("2006-01-02")
	case financeModels.AssetStatus:
		return string(typed)
	case financeModels.AssetLifecycleStage:
		return string(typed)
	default:
		return value
	}
}

func deriveEditAcquisitionCost(existing *financeModels.Asset, req *dto.EditAssetRequest) float64 {
	if existing == nil || req == nil {
		return 0
	}
	shippingCost := existing.ShippingCost
	if req.ShippingCost != nil {
		shippingCost = *req.ShippingCost
	}
	installationCost := existing.InstallationCost
	if req.InstallationCost != nil {
		installationCost = *req.InstallationCost
	}
	taxAmount := existing.TaxAmount
	if req.TaxAmount != nil {
		taxAmount = *req.TaxAmount
	}
	otherCosts := existing.OtherCosts
	if req.OtherCosts != nil {
		otherCosts = *req.OtherCosts
	}
	basePurchasePrice := round2(existing.AcquisitionCost - existing.ShippingCost - existing.InstallationCost - existing.TaxAmount - existing.OtherCosts)
	if basePurchasePrice < 0 {
		basePurchasePrice = 0
	}
	return round2(basePurchasePrice + shippingCost + installationCost + taxAmount + otherCosts)
}

// isUserAdmin checks if the user is an admin based on their role in context
func isUserAdmin(ctx context.Context) bool {
	// Check system admin flag first
	if isSystemAdmin, ok := ctx.Value("is_system_admin").(bool); ok && isSystemAdmin {
		return true
	}

	// Check user role
	if userRole, ok := ctx.Value("user_role").(string); ok {
		return userRole == "admin" || userRole == "system_admin" || userRole == "owner"
	}

	return false
}

// ValidateGroupCChanges ensures Group C fields are not changed if asset is active
// Admins can override this restriction
func ValidateGroupCChanges(ctx context.Context, existing *financeModels.Asset, changes *FieldChanges) error {
	if len(changes.GroupC) == 0 {
		return nil
	}

	// Group C is immutable for active/in_use assets, UNLESS user is admin
	if existing.Status == financeModels.AssetStatusActive ||
		existing.Status == financeModels.AssetStatusInUse {
		// Admin users can edit immutable fields
		if isUserAdmin(ctx) {
			return nil
		}

		fields := make([]string, 0, len(changes.GroupC))
		for field := range changes.GroupC {
			fields = append(fields, field)
		}
		return fmt.Errorf("immutable fields cannot be changed for active assets: %v", fields)
	}

	return nil
}

// DepreciationRecalcInfo holds the depreciation recalculation details
type DepreciationRecalcInfo struct {
	OldMonthlyAmount    float64
	NewMonthlyAmount    float64
	RemainingMonths     int
	OldTotalRemaining   float64
	NewTotalRemaining   float64
	ImpactAmount        float64
	NewBookValue        float64
	EntriesToRegenerate int
	FirstAffectedPeriod string
	DepreciationMethod  string
}

// CalculateDepreciationImpact calculates the impact of Group B changes
func CalculateDepreciationImpact(
	ctx context.Context,
	existing *financeModels.Asset,
	changes *FieldChanges,
	db *gorm.DB,
) (*DepreciationRecalcInfo, error) {
	if len(changes.GroupB) == 0 {
		return nil, nil
	}

	// Get category to resolve defaults
	var category financeModels.AssetCategory
	if err := db.WithContext(ctx).First(&category, "id = ?", existing.CategoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("asset category not found")
		}
		return nil, err
	}

	// Determine new values (use existing if not changed)
	newSalvageValue := existing.SalvageValue
	if val, ok := changes.GroupB["salvage_value"]; ok {
		newSalvageValue = val.(map[string]interface{})["new"].(float64)
	}

	newUsefulLifeMonths := existing.UsefulLifeMonths
	if val, ok := changes.GroupB["useful_life_months"]; ok {
		monthsVal := val.(map[string]interface{})["new"]
		if monthsVal != nil {
			newUsefulLifeMonths = convertToIntPtr(monthsVal)
		}
	}

	newDepMethod := existing.DepreciationMethod
	if val, ok := changes.GroupB["depreciation_method"]; ok {
		methodVal := val.(map[string]interface{})["new"]
		if methodVal != nil {
			newDepMethod = convertToStringPtr(methodVal)
		}
	}

	// Calculate old values
	oldUsefulLife := 0
	if existing.UsefulLifeMonths != nil {
		oldUsefulLife = *existing.UsefulLifeMonths
	} else {
		oldUsefulLife = int(category.UsefulLifeMonths)
	}

	oldMethod := financeModels.DepreciationMethodStraightLine
	if existing.DepreciationMethod != nil {
		oldMethod = financeModels.DepreciationMethod(*existing.DepreciationMethod)
	} else {
		oldMethod = category.DepreciationMethod
	}

	// Calculate monthly depreciation for old and new
	oldMonthly := 0.0
	if oldUsefulLife > 0 {
		oldMonthly = round2(math.Max(0, existing.AcquisitionCost-existing.SalvageValue) / float64(oldUsefulLife))
	}

	newMonthly := 0.0
	newUsefulLifeVal := 0
	if newUsefulLifeMonths != nil {
		newUsefulLifeVal = *newUsefulLifeMonths
	} else {
		newUsefulLifeVal = int(category.UsefulLifeMonths)
	}

	if newUsefulLifeVal > 0 {
		newMonthly = round2(math.Max(0, existing.AcquisitionCost-newSalvageValue) / float64(newUsefulLifeVal))
	}

	// Count remaining months (prospective from now)
	now := apptime.Now()
	remainingMonths := calculateRemainingMonths(existing.AcquisitionDate, now, newUsefulLifeVal)

	oldTotalRemaining := round2(oldMonthly * float64(remainingMonths))
	newTotalRemaining := round2(newMonthly * float64(remainingMonths))
	impactAmount := round2(newTotalRemaining - oldTotalRemaining)

	newBookValue := round2(existing.AcquisitionCost - existing.AccumulatedDepreciation - newTotalRemaining)

	// Count how many depreciation entries need to be regenerated
	entriesToRegen := int64(0)
	var firstUnpostedPeriod string
	var firstUnpostedDep financeModels.AssetDepreciationSchedule
	assetUUID, parseErr := uuid.Parse(existing.ID)
	if parseErr != nil {
		return nil, parseErr
	}
	tenantID := tenantIDFromContext(ctx)
	if err := db.WithContext(ctx).
		Where("asset_id = ? AND tenant_id = ? AND is_posted = false AND deleted_at IS NULL", assetUUID, tenantID).
		Order("period_month ASC").
		First(&firstUnpostedDep).Error; err == nil {
		firstUnpostedPeriod = firstUnpostedDep.PeriodStartDate.Format("2006-01")
	}

	// Count total unposted entries
	db.WithContext(ctx).
		Model(&financeModels.AssetDepreciationSchedule{}).
		Where("asset_id = ? AND tenant_id = ? AND is_posted = false AND deleted_at IS NULL", assetUUID, tenantID).
		Count(&entriesToRegen)

	if firstUnpostedPeriod == "" {
		now := apptime.Now()
		firstUnpostedPeriod = ymPeriod(now.AddDate(0, 1, 0)) // Next month
	}

	depMethod := string(oldMethod)
	if newDepMethod != nil {
		depMethod = *newDepMethod
	}

	return &DepreciationRecalcInfo{
		OldMonthlyAmount:    oldMonthly,
		NewMonthlyAmount:    newMonthly,
		RemainingMonths:     remainingMonths,
		OldTotalRemaining:   oldTotalRemaining,
		NewTotalRemaining:   newTotalRemaining,
		ImpactAmount:        impactAmount,
		NewBookValue:        newBookValue,
		EntriesToRegenerate: int(entriesToRegen),
		FirstAffectedPeriod: firstUnpostedPeriod,
		DepreciationMethod:  depMethod,
	}, nil
}

// RecalculateDepreciation regenerates depreciation schedule after Group B changes
// Returns the count of new entries created
func RecalculateDepreciation(
	ctx context.Context,
	db *gorm.DB,
	asset *financeModels.Asset,
	tenantID string,
	newSalvageValue float64,
	newUsefulLifeMonths *int,
	newDepMethod *string,
) (int, error) {
	if asset == nil {
		return 0, errors.New("asset is required")
	}
	assetUUID, err := uuid.Parse(asset.ID)
	if err != nil {
		return 0, err
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return 0, errors.New("tenant is required")
	}
	// 1. Find first unposted entry (if any)
	var firstUnposted financeModels.AssetDepreciationSchedule
	hasUnposted := db.WithContext(ctx).
		Where("asset_id = ? AND tenant_id = ? AND is_posted = false AND deleted_at IS NULL", assetUUID, tenantID).
		Order("period_month ASC").
		First(&firstUnposted).Error == nil

	if !hasUnposted {
		// All entries are posted, cannot recalculate (immutable)
		return 0, errors.New("all depreciation entries are posted, cannot recalculate")
	}

	// 2. Calculate affected period
	acqDate := asset.AcquisitionDate

	// 3. Soft-delete unposted entries from fromPeriod forward
	if err := db.WithContext(ctx).
		Where("asset_id = ? AND tenant_id = ? AND is_posted = false AND deleted_at IS NULL AND period_start_date >= ?", assetUUID, tenantID, firstUnposted.PeriodStartDate).
		Delete(&financeModels.AssetDepreciationSchedule{}).Error; err != nil {
		return 0, err
	}

	// 4. Resolve new depreciation method
	var category financeModels.AssetCategory
	if err := db.WithContext(ctx).First(&category, "id = ?", asset.CategoryID).Error; err != nil {
		return 0, err
	}

	newMethod := asset.DepreciationMethod
	if newDepMethod != nil {
		newMethod = newDepMethod
	} else if asset.DepreciationMethod == nil {
		resolved := string(category.DepreciationMethod)
		newMethod = &resolved
	}

	// 5. Calculate remaining useful life
	usefulLife := 0
	if newUsefulLifeMonths != nil {
		usefulLife = *newUsefulLifeMonths
	} else if asset.UsefulLifeMonths != nil {
		usefulLife = *asset.UsefulLifeMonths
	} else {
		usefulLife = int(category.UsefulLifeMonths)
	}

	// 6. Regenerate schedule from fromPeriod
	method := financeModels.DepreciationMethodStraightLine
	if newMethod != nil && strings.TrimSpace(*newMethod) != "" {
		method = financeModels.DepreciationMethod(strings.TrimSpace(*newMethod))
	}
	engine, err := depreciationsvc.NewDepreciationEngine(method, asset.AcquisitionCost, newSalvageValue, usefulLife, asset.AcquisitionDate, acqDate)
	if err != nil {
		return 0, err
	}
	allSchedules, err := engine.GenerateSchedule(acqDate.AddDate(0, usefulLife, 0))
	if err != nil {
		return 0, err
	}

	// Count how many entries we regenerate
	entriesCreated := 0
	for _, schedule := range allSchedules {
		if schedule.PeriodStartDate.Before(firstUnposted.PeriodStartDate) {
			continue
		}
		row := financeModels.AssetDepreciationSchedule{
			TenantID:                tenantID,
			AssetID:                 assetUUID,
			PeriodStartDate:         schedule.PeriodStartDate,
			PeriodEndDate:           schedule.PeriodEndDate,
			PeriodMonth:             schedule.PeriodMonth,
			DepreciationAmount:      schedule.DepreciationAmount,
			AccumulatedDepreciation: schedule.AccumulatedDepreciation,
			BookValue:               schedule.BookValue,
		}
		if err := db.WithContext(ctx).Create(&row).Error; err != nil {
			return entriesCreated, err
		}
		entriesCreated++
	}

	return entriesCreated, nil
}

// Helper functions

func parseStringToTime(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

func calculateRemainingMonths(acqDate time.Time, now time.Time, usefulLife int) int {
	depStartDate := acqDate
	monthsElapsed := (now.Year()-depStartDate.Year())*12 + int(now.Month()) - int(depStartDate.Month())
	remaining := usefulLife - monthsElapsed
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

func convertToIntPtr(v interface{}) *int {
	switch t := v.(type) {
	case *int:
		return t
	case int:
		return &t
	case float64:
		i := int(t)
		return &i
	case *float64:
		if t == nil {
			return nil
		}
		i := int(*t)
		return &i
	case string:
		if t == "" {
			return nil
		}
		// try parse as int
		var i int
		_, err := fmt.Sscanf(t, "%d", &i)
		if err != nil {
			return nil
		}
		return &i
	default:
		return nil
	}
}

func convertToStringPtr(v interface{}) *string {
	switch t := v.(type) {
	case *string:
		return t
	case string:
		s := strings.TrimSpace(t)
		return &s
	default:
		// fallback to fmt.Sprintf
		s := fmt.Sprintf("%v", v)
		s = strings.TrimSpace(s)
		if s == "<nil>" || s == "" {
			return nil
		}
		return &s
	}
}

func closeCurrentAssignmentTx(tx *gorm.DB, assetID, tenantID string, returnAt time.Time) error {
	return tx.Model(&financeModels.AssetAssignmentHistory{}).
		Where("asset_id = ? AND tenant_id = ? AND returned_at IS NULL", assetID, tenantID).
		Updates(map[string]interface{}{
			"returned_at":   returnAt,
			"return_reason": "reassigned",
		}).Error
}

func insertAssignmentHistoryTx(tx *gorm.DB, assetID, tenantID string, req *dto.EditAssetRequest, actorID string, assignedAt time.Time) error {
	if req == nil || req.AssignedToEmployeeID == nil || strings.TrimSpace(*req.AssignedToEmployeeID) == "" {
		return nil
	}

	employeeUUID, err := uuid.Parse(strings.TrimSpace(*req.AssignedToEmployeeID))
	if err != nil {
		return err
	}
	assetUUID, err := uuid.Parse(assetID)
	if err != nil {
		return err
	}
	actorUUID, err := uuid.Parse(strings.TrimSpace(actorID))
	if err != nil {
		return err
	}

	var tenantUUID *uuid.UUID
	if parsed, err := uuid.Parse(strings.TrimSpace(tenantID)); err == nil {
		tenantUUID = &parsed
	}
	history := financeModels.AssetAssignmentHistory{
		TenantID:   tenantUUID,
		AssetID:    assetUUID,
		EmployeeID: &employeeUUID,
		AssignedAt: assignedAt,
		AssignedBy: &actorUUID,
	}
	if req.DepartmentID != nil && strings.TrimSpace(*req.DepartmentID) != "" {
		departmentValue := strings.TrimSpace(*req.DepartmentID)
		if departmentUUID := parseUUIDPtr(&departmentValue); departmentUUID != nil {
			history.DepartmentID = departmentUUID
		}
	}
	if strings.TrimSpace(req.LocationID) != "" {
		locationValue := strings.TrimSpace(req.LocationID)
		if locationUUID := parseUUIDPtr(&locationValue); locationUUID != nil {
			history.LocationID = locationUUID
		}
	}
	return tx.Create(&history).Error
}

func cascadeStatusToChildrenTx(tx *gorm.DB, parentID, tenantID string, status financeModels.AssetStatus, lifecycle financeModels.AssetLifecycleStage) error {
	return tx.Model(&financeModels.Asset{}).
		Where("tenant_id = ? AND parent_asset_id = ?", tenantID, parentID).
		Updates(map[string]interface{}{
			"status":          status,
			"lifecycle_stage": lifecycle,
		}).Error
}
