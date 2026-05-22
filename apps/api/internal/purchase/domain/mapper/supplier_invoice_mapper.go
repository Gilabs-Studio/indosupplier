package mapper

import (
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
)

type SupplierInvoiceMapper struct{}

// formatDateToISO formats a time.Time to ISO8601 date string (YYYY-MM-DD)
func formatDateToISO(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// resolveSupplierName returns the best available supplier name, falling back through
// the snapshot hierarchy to handle legacy records that predate snapshot columns.
func resolveSupplierName(si *models.SupplierInvoice) string {
	if sn := strings.TrimSpace(si.SupplierNameSnapshot); sn != "" {
		return sn
	}
	if si.PurchaseOrder != nil {
		if sns := strings.TrimSpace(si.PurchaseOrder.SupplierNameSnapshot); sns != "" {
			return sns
		}
		if si.PurchaseOrder.Supplier != nil {
			if name := strings.TrimSpace(si.PurchaseOrder.Supplier.Name); name != "" {
				return name
			}
		}
	}
	return ""
}

// resolveSupplierID mirrors the name fallback: prefers the invoice's own SupplierID,
// then falls back to the referenced PO's supplier identifier.
func resolveSupplierID(si *models.SupplierInvoice) string {
	if si.SupplierID != "" {
		return si.SupplierID
	}
	if si.PurchaseOrder != nil {
		if si.PurchaseOrder.SupplierID != nil && *si.PurchaseOrder.SupplierID != "" {
			return *si.PurchaseOrder.SupplierID
		}
		if si.PurchaseOrder.Supplier != nil {
			return si.PurchaseOrder.Supplier.ID
		}
	}
	return ""
}

func NewSupplierInvoiceMapper() *SupplierInvoiceMapper {
	return &SupplierInvoiceMapper{}
}

func (m *SupplierInvoiceMapper) ToListResponse(si *models.SupplierInvoice) *dto.SupplierInvoiceListResponse {
	if si == nil {
		return nil
	}

	var poMini *dto.SupplierInvoicePurchaseOrderMini
	if si.PurchaseOrder != nil {
		poMini = &dto.SupplierInvoicePurchaseOrderMini{ID: si.PurchaseOrder.ID, Code: si.PurchaseOrder.Code}
	}

	var grMini *dto.SupplierInvoiceGoodsReceiptMini
	if si.GoodsReceipt != nil {
		grMini = &dto.SupplierInvoiceGoodsReceiptMini{ID: si.GoodsReceipt.ID, Code: si.GoodsReceipt.Code}
	}

	var ptMini *dto.SupplierInvoicePaymentTermsMini
	if strings.TrimSpace(si.PaymentTermsNameSnapshot) != "" || si.PaymentTermsDaysSnapshot != nil {
		id := ""
		if si.PaymentTermsID != nil {
			id = strings.TrimSpace(*si.PaymentTermsID)
		}
		ptMini = &dto.SupplierInvoicePaymentTermsMini{ID: id, Name: strings.TrimSpace(si.PaymentTermsNameSnapshot), Days: si.PaymentTermsDaysSnapshot}
	} else if si.PaymentTerms != nil {
		days := si.PaymentTerms.Days
		ptMini = &dto.SupplierInvoicePaymentTermsMini{ID: si.PaymentTerms.ID, Name: si.PaymentTerms.Name, Days: &days}
	}

	var dpMini *dto.SupplierInvoiceAddDownPaymentMini
	if si.DownPaymentInvoice != nil {
		dpMini = &dto.SupplierInvoiceAddDownPaymentMini{
			ID:            si.DownPaymentInvoice.ID,
			Code:          si.DownPaymentInvoice.Code,
			InvoiceNumber: si.DownPaymentInvoice.InvoiceNumber,
			InvoiceDate:   formatDateToISO(si.DownPaymentInvoice.InvoiceDate),
			DueDate:       formatDateToISO(si.DownPaymentInvoice.DueDate),
			Amount:        si.DownPaymentInvoice.Amount,
			Status:        string(si.DownPaymentInvoice.Status),
		}
		if si.DownPaymentInvoice.Notes != nil {
			dpMini.Notes = si.DownPaymentInvoice.Notes
		}
	}

	return &dto.SupplierInvoiceListResponse{
		ID:                 si.ID,
		CompanyID:          si.CompanyID,
		FiscalYearID:       si.FiscalYearID,
		PurchaseOrder:      poMini,
		GoodsReceipt:       grMini,
		PaymentTerms:       ptMini,
		Type:               string(si.Type),
		Code:               si.Code,
		InvoiceNumber:      si.InvoiceNumber,
		InvoiceDate:        formatDateToISO(si.InvoiceDate),
		DueDate:            formatDateToISO(si.DueDate),
		SupplierID:         resolveSupplierID(si),
		SupplierName:       resolveSupplierName(si),
		TaxRate:            si.TaxRate,
		TaxAmount:          si.TaxAmount,
		DeliveryCost:       si.DeliveryCost,
		OtherCost:          si.OtherCost,
		SubTotal:           si.SubTotal,
		Amount:             si.Amount,
		PaidAmount:         si.PaidAmount,
		RemainingAmount:    si.RemainingAmount,
		DownPaymentAmount:  si.DownPaymentAmount,
		DownPaymentInvoice: dpMini,
		IsPosted:           si.IsPosted,
		JournalEntryID:     si.JournalEntryID,
		Status:             string(si.Status),
		Notes:              si.Notes,
		CreatedBy:          si.CreatedBy,
		SubmittedAt:        si.SubmittedAt,
		ApprovedAt:         si.ApprovedAt,
		RejectedAt:         si.RejectedAt,
		CancelledAt:        si.CancelledAt,
	}
}

func (m *SupplierInvoiceMapper) ToListResponseList(items []*models.SupplierInvoice) []*dto.SupplierInvoiceListResponse {
	out := make([]*dto.SupplierInvoiceListResponse, 0, len(items))
	for _, it := range items {
		out = append(out, m.ToListResponse(it))
	}
	return out
}

func (m *SupplierInvoiceMapper) ToDetailResponse(si *models.SupplierInvoice) *dto.SupplierInvoiceDetailResponse {
	if si == nil {
		return nil
	}

	var poMini *dto.SupplierInvoicePurchaseOrderMini
	if si.PurchaseOrder != nil {
		poMini = &dto.SupplierInvoicePurchaseOrderMini{ID: si.PurchaseOrder.ID, Code: si.PurchaseOrder.Code}
	}

	var grMini *dto.SupplierInvoiceGoodsReceiptMini
	if si.GoodsReceipt != nil {
		grMini = &dto.SupplierInvoiceGoodsReceiptMini{ID: si.GoodsReceipt.ID, Code: si.GoodsReceipt.Code}
	}

	var ptMini *dto.SupplierInvoicePaymentTermsMini
	if strings.TrimSpace(si.PaymentTermsNameSnapshot) != "" || si.PaymentTermsDaysSnapshot != nil {
		id := ""
		if si.PaymentTermsID != nil {
			id = strings.TrimSpace(*si.PaymentTermsID)
		}
		ptMini = &dto.SupplierInvoicePaymentTermsMini{ID: id, Name: strings.TrimSpace(si.PaymentTermsNameSnapshot), Days: si.PaymentTermsDaysSnapshot}
	} else if si.PaymentTerms != nil {
		days := si.PaymentTerms.Days
		ptMini = &dto.SupplierInvoicePaymentTermsMini{ID: si.PaymentTerms.ID, Name: si.PaymentTerms.Name, Days: &days}
	}

	var dpMini *dto.SupplierInvoiceAddDownPaymentMini
	if si.DownPaymentInvoice != nil {
		dpMini = &dto.SupplierInvoiceAddDownPaymentMini{
			ID:            si.DownPaymentInvoice.ID,
			Code:          si.DownPaymentInvoice.Code,
			InvoiceNumber: si.DownPaymentInvoice.InvoiceNumber,
			InvoiceDate:   formatDateToISO(si.DownPaymentInvoice.InvoiceDate),
			DueDate:       formatDateToISO(si.DownPaymentInvoice.DueDate),
			Amount:        si.DownPaymentInvoice.Amount,
			Status:        string(si.DownPaymentInvoice.Status),
		}
		if si.DownPaymentInvoice.Notes != nil {
			dpMini.Notes = si.DownPaymentInvoice.Notes
		}
	}

	items := make([]dto.SupplierInvoiceItemResponse, 0, len(si.Items))
	for _, it := range si.Items {
		productObj := any(it.Product)
		if strings.TrimSpace(it.ProductNameSnapshot) != "" || strings.TrimSpace(it.ProductCodeSnapshot) != "" {
			productObj = &struct {
				ID   string `json:"id"`
				Code string `json:"code"`
				Name string `json:"name"`
			}{
				ID:   strings.TrimSpace(it.ProductID),
				Code: strings.TrimSpace(it.ProductCodeSnapshot),
				Name: strings.TrimSpace(it.ProductNameSnapshot),
			}
		}
		items = append(items, dto.SupplierInvoiceItemResponse{
			ID:                  it.ID,
			SupplierInvoiceID:   it.SupplierInvoiceID,
			ProductID:           it.ProductID,
			Product:             productObj,
			Quantity:            it.Quantity,
			Price:               it.Price,
			Discount:            it.Discount,
			SubTotal:            it.SubTotal,
			PurchaseOrderItemID: it.PurchaseOrderItemID,
			CreatedAt:           it.CreatedAt,
			UpdatedAt:           it.UpdatedAt,
		})
	}

	return &dto.SupplierInvoiceDetailResponse{
		ID:                 si.ID,
		CompanyID:          si.CompanyID,
		FiscalYearID:       si.FiscalYearID,
		PurchaseOrder:      poMini,
		GoodsReceipt:       grMini,
		PaymentTerms:       ptMini,
		Type:               string(si.Type),
		Code:               si.Code,
		InvoiceNumber:      si.InvoiceNumber,
		InvoiceDate:        formatDateToISO(si.InvoiceDate),
		DueDate:            formatDateToISO(si.DueDate),
		SupplierID:         resolveSupplierID(si),
		SupplierName:       resolveSupplierName(si),
		TaxRate:            si.TaxRate,
		TaxAmount:          si.TaxAmount,
		DeliveryCost:       si.DeliveryCost,
		OtherCost:          si.OtherCost,
		SubTotal:           si.SubTotal,
		Amount:             si.Amount,
		PaidAmount:         si.PaidAmount,
		RemainingAmount:    si.RemainingAmount,
		DownPaymentAmount:  si.DownPaymentAmount,
		DownPaymentInvoice: dpMini,
		IsPosted:           si.IsPosted,
		JournalEntryID:     si.JournalEntryID,
		Status:             string(si.Status),
		Notes:              si.Notes,
		Items:              items,
		CreatedBy:          si.CreatedBy,
		SubmittedAt:        si.SubmittedAt,
		ApprovedAt:         si.ApprovedAt,
		RejectedAt:         si.RejectedAt,
		CancelledAt:        si.CancelledAt,
	}
}

func (m *SupplierInvoiceMapper) ToDownPaymentListResponse(si *models.SupplierInvoice) *dto.SupplierInvoiceDownPaymentListResponse {
	if si == nil {
		return nil
	}
	var poMini *dto.SupplierInvoicePurchaseOrderMini
	if si.PurchaseOrder != nil {
		poMini = &dto.SupplierInvoicePurchaseOrderMini{ID: si.PurchaseOrder.ID, Code: si.PurchaseOrder.Code}
	}
	regulars := []dto.SupplierInvoiceDownPaymentRegularInvoiceMini{}
	for _, reg := range si.RegularInvoices {
		regulars = append(regulars, dto.SupplierInvoiceDownPaymentRegularInvoiceMini{
			ID:   reg.ID,
			Code: reg.Code,
		})
	}
	return &dto.SupplierInvoiceDownPaymentListResponse{
		ID:              si.ID,
		CompanyID:       si.CompanyID,
		FiscalYearID:    si.FiscalYearID,
		PurchaseOrder:   poMini,
		SupplierID:      resolveSupplierID(si),
		SupplierName:    resolveSupplierName(si),
		Code:            si.Code,
		InvoiceNumber:   si.InvoiceNumber,
		InvoiceDate:     formatDateToISO(si.InvoiceDate),
		DueDate:         formatDateToISO(si.DueDate),
		Amount:          si.Amount,
		PaidAmount:      si.PaidAmount,
		RemainingAmount: si.RemainingAmount,
		Status:          string(si.Status),
		Notes:           si.Notes,
		RegularInvoices: regulars,
		SubmittedAt:     si.SubmittedAt,
		ApprovedAt:      si.ApprovedAt,
		RejectedAt:      si.RejectedAt,
		CancelledAt:     si.CancelledAt,
		CreatedAt:       si.CreatedAt,
		UpdatedAt:       si.UpdatedAt,
	}
}

func (m *SupplierInvoiceMapper) ToDownPaymentListResponseList(items []*models.SupplierInvoice) []*dto.SupplierInvoiceDownPaymentListResponse {
	out := make([]*dto.SupplierInvoiceDownPaymentListResponse, 0, len(items))
	for _, it := range items {
		out = append(out, m.ToDownPaymentListResponse(it))
	}
	return out
}

func (m *SupplierInvoiceMapper) ToDownPaymentDetailResponse(si *models.SupplierInvoice) *dto.SupplierInvoiceDownPaymentDetailResponse {
	if si == nil {
		return nil
	}
	var poMini *dto.SupplierInvoicePurchaseOrderMini
	if si.PurchaseOrder != nil {
		poMini = &dto.SupplierInvoicePurchaseOrderMini{ID: si.PurchaseOrder.ID, Code: si.PurchaseOrder.Code}
	}
	regulars := []dto.SupplierInvoiceDownPaymentRegularInvoiceMini{}
	for _, reg := range si.RegularInvoices {
		regulars = append(regulars, dto.SupplierInvoiceDownPaymentRegularInvoiceMini{
			ID:   reg.ID,
			Code: reg.Code,
		})
	}
	return &dto.SupplierInvoiceDownPaymentDetailResponse{
		ID:              si.ID,
		CompanyID:       si.CompanyID,
		FiscalYearID:    si.FiscalYearID,
		PurchaseOrder:   poMini,
		SupplierID:      resolveSupplierID(si),
		SupplierName:    resolveSupplierName(si),
		Code:            si.Code,
		InvoiceNumber:   si.InvoiceNumber,
		InvoiceDate:     formatDateToISO(si.InvoiceDate),
		DueDate:         formatDateToISO(si.DueDate),
		Amount:          si.Amount,
		PaidAmount:      si.PaidAmount,
		RemainingAmount: si.RemainingAmount,
		Status:          string(si.Status),
		Notes:           si.Notes,
		RegularInvoices: regulars,
		SubmittedAt:     si.SubmittedAt,
		ApprovedAt:      si.ApprovedAt,
		RejectedAt:      si.RejectedAt,
		CancelledAt:     si.CancelledAt,
		CreatedBy:       si.CreatedBy,
		CreatedAt:       si.CreatedAt,
		UpdatedAt:       si.UpdatedAt,
	}
}
