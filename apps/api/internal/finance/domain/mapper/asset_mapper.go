package mapper

import (
	"time"
	"strings"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

type AssetMapper struct {
	categoryMapper *AssetCategoryMapper
	locationMapper *AssetLocationMapper
}

func NewAssetMapper(categoryMapper *AssetCategoryMapper, locationMapper *AssetLocationMapper) *AssetMapper {
	return &AssetMapper{categoryMapper: categoryMapper, locationMapper: locationMapper}
}

// Helper function to safely dereference string pointers
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func displayStatusForAsset(status financeModels.AssetStatus, lifecycle financeModels.AssetLifecycleStage) string {
	normalizedStatus := strings.ToLower(strings.TrimSpace(string(status)))
	normalizedLifecycle := strings.ToLower(strings.TrimSpace(string(lifecycle)))

	switch normalizedStatus {
	case "sold", "transferred", "written_off", "disposed":
		return normalizedStatus
	}

	if normalizedLifecycle != "" {
		return normalizedLifecycle
	}
	if normalizedStatus != "" {
		return normalizedStatus
	}
	return "draft"
}

func warrantyExpiryDays(warrantyEnd *time.Time) *int {
	if warrantyEnd == nil {
		return nil
	}
	location := warrantyEnd.Location()
	if location == nil {
		location = time.Local
	}
	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	end := time.Date(warrantyEnd.Year(), warrantyEnd.Month(), warrantyEnd.Day(), 0, 0, 0, 0, location)
	days := int(end.Sub(today).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return &days
}

func (m *AssetMapper) ToResponse(item *financeModels.Asset, withDetails bool) dto.AssetResponse {
	if item == nil {
		return dto.AssetResponse{}
	}

	resp := dto.AssetResponse{
		// Core
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,

		// Identity
		SerialNumber: item.SerialNumber,
		Barcode:      item.Barcode,
		QRCode:       item.QRCode,
		AssetTag:     item.AssetTag,

		// Category / Location
		AssetTypeID: getStringValue(item.AssetTypeID),
		CategoryID:  item.CategoryID,
		LocationID:  item.LocationID,

		// Organization
		CompanyID:       item.CompanyID,
		BusinessUnitID:  item.BusinessUnitID,
		DepartmentID:    item.DepartmentID,
		CustodianUserID: item.CustodianUserID,
		VendorID:        item.SupplierID,
		PurchaseInvoiceID: item.SupplierInvoiceID,

		// Assignment
		AssignedToEmployeeID: item.AssignedToEmployeeID,
		AssignmentDate:       item.AssignmentDate,

		// Acquisition
		AcquisitionDate: item.AcquisitionDate,
		AcquisitionCost: item.AcquisitionCost,
		SalvageValue:    item.SalvageValue,

		// Cost Breakdown
		ShippingCost:     item.ShippingCost,
		InstallationCost: item.InstallationCost,
		TaxAmount:        item.TaxAmount,
		OtherCosts:       item.OtherCosts,
		TotalCost:        item.TotalAcquisitionCost(),

		// Depreciation
		AccumulatedDepreciation: item.AccumulatedDepreciation,
		BookValue:               item.BookValue,

		// Depreciation Config
		DepreciationMethod:    item.DepreciationMethod,
		UsefulLifeMonths:      item.UsefulLifeMonths,
		DepreciationStartDate: item.DepreciationStartDate,

		// Status / Lifecycle
		Status:               item.Status,
		LifecycleStage:       item.LifecycleStage,
		DisplayStatus:        displayStatusForAsset(item.Status, item.LifecycleStage),
		IsCapitalized:        item.IsCapitalized,
		IsDepreciable:        item.IsDepreciable,
		IsFullyDepreciated:   item.IsFullyDepreciated,
		DisposedAt:           item.DisposedAt,
		DepreciationProgress: item.DepreciationProgress(),
		AgeInMonths:          item.AgeInMonths(),

		// Parent/Child
		ParentAssetID: item.ParentAssetID,
		IsParent:      item.IsParent,

		// Warranty
		WarrantyStart:         item.WarrantyStart,
		WarrantyEnd:           item.WarrantyEnd,
		WarrantyProvider:      item.WarrantyProvider,
		WarrantyTerms:         item.WarrantyTerms,
		IsUnderWarranty:       item.IsUnderWarranty(),
		WarrantyDaysRemaining: item.WarrantyDaysRemaining(),
		WarrantyExpiryDays:    warrantyExpiryDays(item.WarrantyEnd),

		// Insurance
		InsurancePolicyNumber: item.InsurancePolicyNumber,
		InsuranceProvider:     item.InsuranceProvider,
		InsuranceStart:        item.InsuranceStart,
		InsuranceEnd:          item.InsuranceEnd,
		InsuranceValue:        item.InsuranceValue,
		IsInsured:             item.IsInsured(),

		// Audit
		CreatedBy:  item.CreatedBy,
		ApprovedBy: item.ApprovedBy,
		ApprovedAt: item.ApprovedAt,

		// Timestamps
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}

	if item.Category != nil {
		cat := m.categoryMapper.ToResponse(item.Category)
		resp.Category = &cat
		resp.CategoryName = item.Category.Name
	}
	if item.Location != nil {
		loc := m.locationMapper.ToResponse(item.Location)
		resp.Location = &loc
		resp.LocationName = item.Location.Name
	}
	if item.Department != nil {
		resp.DeptName = item.Department.Name
	}
	if item.AssetTypeID != nil {
		resp.AssetType = getStringValue(item.AssetTypeID)
	}

	if withDetails {
		// Depreciations
		deps := make([]dto.AssetDepreciationResponse, 0, len(item.Depreciations))
		for i := range item.Depreciations {
			d := item.Depreciations[i]
			deps = append(deps, dto.AssetDepreciationResponse{
				ID:               d.ID,
				AssetID:          d.AssetID,
				Period:           d.Period,
				DepreciationDate: d.DepreciationDate,
				Method:           d.Method,
				Amount:           d.Amount,
				Accumulated:      d.Accumulated,
				BookValue:        d.BookValue,
				JournalEntryID:   d.JournalEntryID,
				CreatedAt:        d.CreatedAt,
			})
		}
		resp.Depreciations = deps

		// Transactions
		txs := make([]dto.AssetTransactionResponse, 0, len(item.Transactions))
		for i := range item.Transactions {
			t := item.Transactions[i]
			txs = append(txs, dto.AssetTransactionResponse{
				ID:                     t.ID,
				AssetID:                t.AssetID,
				Type:                   t.Type,
				TransactionDate:        t.TransactionDate,
				Amount:                 t.Amount,
				Description:            t.Description,
				Status:                 string(t.Status),
				ReferenceType:          t.ReferenceType,
				ReferenceID:            t.ReferenceID,
				ProceedsAmount:         t.ProceedsAmount,
				BankAccountID:          t.BankAccountID,
				BookValueAtTransaction: t.BookValueAtTransaction,
				GainLossAmount:         t.GainLossAmount,
				GainLossAccountID:      t.GainLossAccountID,
				CreatedAt:              t.CreatedAt,
			})
		}
		resp.Transactions = txs

		// Attachments
		attachments := make([]dto.AssetAttachmentResponse, 0, len(item.Attachments))
		for i := range item.Attachments {
			a := item.Attachments[i]
			attachments = append(attachments, m.ToAttachmentResponse(&a))
		}
		resp.Attachments = attachments

		// Audit Logs
		auditLogs := make([]dto.AssetAuditLogResponse, 0, len(item.AuditLogs))
		for i := range item.AuditLogs {
			al := item.AuditLogs[i]
			auditLogs = append(auditLogs, m.ToAuditLogResponse(&al))
		}
		resp.AuditLogs = auditLogs

		// Assignment Histories
		assignments := make([]dto.AssetAssignmentHistoryResponse, 0, len(item.AssignmentHistories))
		for i := range item.AssignmentHistories {
			ah := item.AssignmentHistories[i]
			assignments = append(assignments, m.ToAssignmentHistoryResponse(&ah))
		}
		resp.AssignmentHistories = assignments
	}

	return resp
}

func (m *AssetMapper) ToAttachmentResponse(a *financeModels.AssetAttachment) dto.AssetAttachmentResponse {
	resp := dto.AssetAttachmentResponse{
		ID:          a.ID.String(),
		AssetID:     a.AssetID.String(),
		FileName:    a.FileName,
		FilePath:    a.FilePath,
		FileURL:     a.FileURL,
		FileType:    a.FileType,
		FileSize:    a.FileSize,
		MimeType:    a.MimeType,
		Description: a.Description,
		UploadedBy:  dto.UuidPtrToStringPtr(a.UploadedBy),
		UploadedAt:  a.UploadedAt,
		CreatedAt:   a.CreatedAt,
	}
	return resp
}

func (m *AssetMapper) ToAuditLogResponse(al *financeModels.AssetAuditLog) dto.AssetAuditLogResponse {
	changes := make([]dto.AuditChangeResponse, 0, len(al.Changes))
	for _, c := range al.Changes {
		changes = append(changes, dto.AuditChangeResponse{
			Field:    c.Field,
			OldValue: c.OldValue,
			NewValue: c.NewValue,
		})
	}

	resp := dto.AssetAuditLogResponse{
		ID:          al.ID.String(),
		AssetID:     al.AssetID.String(),
		Action:      al.Action,
		Changes:     changes,
		PerformedBy: dto.UuidPtrToStringPtr(al.PerformedBy),
		PerformedAt: al.PerformedAt,
		IPAddress:   al.IPAddress,
		Metadata:    al.Metadata,
		CreatedAt:   al.CreatedAt,
	}
	return resp
}

func (m *AssetMapper) ToAssignmentHistoryResponse(ah *financeModels.AssetAssignmentHistory) dto.AssetAssignmentHistoryResponse {
	resp := dto.AssetAssignmentHistoryResponse{
		ID:           ah.ID.String(),
		AssetID:      ah.AssetID.String(),
		EmployeeID:   dto.UuidPtrToStringPtr(ah.EmployeeID),
		DepartmentID: dto.UuidPtrToStringPtr(ah.DepartmentID),
		LocationID:   dto.UuidPtrToStringPtr(ah.LocationID),
		AssignedAt:   ah.AssignedAt,
		AssignedBy:   dto.UuidPtrToStringPtr(ah.AssignedBy),
		ReturnedAt:   ah.ReturnedAt,
		ReturnReason: ah.ReturnReason,
		Notes:        ah.Notes,
		CreatedAt:    ah.CreatedAt,
	}

	// Populate employee name if relation loaded
	if ah.Employee != nil {
		name := ah.Employee.Name
		resp.EmployeeName = &name
	}

	return resp
}
