package mapper

import (
	"strings"

	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
)

func safePtrString(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

type PurchaseOrderMapper struct{}

func NewPurchaseOrderMapper() *PurchaseOrderMapper {
	return &PurchaseOrderMapper{}
}

func (m *PurchaseOrderMapper) ToListResponse(po *models.PurchaseOrder) *dto.PurchaseOrderListResponse {
	if po == nil {
		return nil
	}

	var supplierSummary *dto.PurchaseOrderPartySummary
	if strings.TrimSpace(po.SupplierNameSnapshot) != "" || strings.TrimSpace(po.SupplierCodeSnapshot) != "" {
		supplierSummary = &dto.PurchaseOrderPartySummary{
			ID:   safePtrString(po.SupplierID),
			Code: strings.TrimSpace(po.SupplierCodeSnapshot),
			Name: strings.TrimSpace(po.SupplierNameSnapshot),
		}
	} else if po.Supplier != nil {
		// Fall back to the preloaded relation when snapshot fields were not captured.
		supplierSummary = &dto.PurchaseOrderPartySummary{
			ID:   po.Supplier.ID,
			Code: po.Supplier.Code,
			Name: po.Supplier.Name,
		}
	}

	var prRef *dto.PurchaseOrderRequisitionRef
	if po.PurchaseRequisition != nil {
		prRef = &dto.PurchaseOrderRequisitionRef{
			ID:   po.PurchaseRequisition.ID,
			Code: po.PurchaseRequisition.Code,
		}
	}

	grSummaries := make([]dto.GoodsReceiptSummary, 0, len(po.GoodsReceipts))
	for _, gr := range po.GoodsReceipts {
		var totalReceivedForGR float64
		for _, gi := range gr.Items {
			totalReceivedForGR += gi.QuantityReceived
		}
		grSummaries = append(grSummaries, dto.GoodsReceiptSummary{
			ID:                  gr.ID,
			Code:                gr.Code,
			Status:              string(gr.Status),
			TotalItems:          len(gr.Items),
			TotalItemsReceived:  totalReceivedForGR,
			CreatedAt:           gr.CreatedAt,
		})
	}

	siSummaries := make([]dto.SupplierInvoiceSummary, 0, len(po.SupplierInvoices))
	for _, si := range po.SupplierInvoices {
		var grID *string
		var grCode *string
		if si.GoodsReceipt != nil {
			grID = &si.GoodsReceipt.ID
			grCode = &si.GoodsReceipt.Code
		}
		siSummaries = append(siSummaries, dto.SupplierInvoiceSummary{
			ID:               si.ID,
			Code:             si.Code,
			Status:           string(si.Status),
			Amount:           si.Amount,
			PaidAmount:       si.PaidAmount,
			GoodsReceiptID:   grID,
			GoodsReceiptCode: grCode,
			CreatedAt:        si.CreatedAt,
		})
	}

	// Compute receipt fulfillment only for active APPROVED POs.
	var fulfillment *dto.POFulfillmentSummary
	status := strings.TrimSpace(string(po.Status))
	if strings.EqualFold(status, string(models.PurchaseOrderStatusApproved)) && len(po.Items) > 0 {
		var totalOrdered, totalReceived, totalPending float64
		for _, item := range po.Items {
			totalOrdered += item.Quantity
		}
		for _, gr := range po.GoodsReceipts {
			grStatus := strings.ToUpper(strings.TrimSpace(string(gr.Status)))
			for _, grItem := range gr.Items {
				switch grStatus {
				// Items are physically received once the GR is approved/partial/closed/confirmed (legacy).
				case string(models.GoodsReceiptStatusApproved),
					string(models.GoodsReceiptStatusPartial),
					string(models.GoodsReceiptStatusClosed),
					string(models.GoodsReceiptStatusConfirmed):
					totalReceived += grItem.QuantityReceived
				// Items are in-flight when the GR is still a draft or awaiting approval.
				case string(models.GoodsReceiptStatusDraft),
					string(models.GoodsReceiptStatusSubmitted):
					totalPending += grItem.QuantityReceived
				// REJECTED GRs do not contribute to fulfillment.
				}
			}
		}
		totalRemaining := totalOrdered - totalReceived - totalPending
		if totalRemaining < 0 {
			totalRemaining = 0
		}
		fulfillment = &dto.POFulfillmentSummary{
			TotalOrdered:   totalOrdered,
			TotalReceived:  totalReceived,
			TotalPending:   totalPending,
			TotalRemaining: totalRemaining,
		}
	}

	return &dto.PurchaseOrderListResponse{
		ID:                  po.ID,
		Code:                po.Code,
		CompanyID:           po.CompanyID,
		FiscalYearID:        po.FiscalYearID,
		OrderDate:           po.OrderDate,
		DueDate:             po.DueDate,
		Status:              string(po.Status),
		TotalAmount:         po.TotalAmount,
		Supplier:            supplierSummary,
		PurchaseRequisition: prRef,
		GoodsReceipts:       grSummaries,
		SupplierInvoices:    siSummaries,
		Fulfillment:         fulfillment,
	}
}

func (m *PurchaseOrderMapper) ToListResponseList(items []*models.PurchaseOrder) []*dto.PurchaseOrderListResponse {
	out := make([]*dto.PurchaseOrderListResponse, 0, len(items))
	for _, it := range items {
		out = append(out, m.ToListResponse(it))
	}
	return out
}

func (m *PurchaseOrderMapper) ToDetailResponse(po *models.PurchaseOrder) *dto.PurchaseOrderDetailResponse {
	if po == nil {
		return nil
	}

	supplierObj := any(po.Supplier)
	if strings.TrimSpace(po.SupplierNameSnapshot) != "" || strings.TrimSpace(po.SupplierCodeSnapshot) != "" {
		supplierObj = &struct {
			ID   string `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		}{
			ID:   safePtrString(po.SupplierID),
			Code: strings.TrimSpace(po.SupplierCodeSnapshot),
			Name: strings.TrimSpace(po.SupplierNameSnapshot),
		}
	}

	paymentTermsObj := any(po.PaymentTerms)
	if strings.TrimSpace(po.PaymentTermsNameSnapshot) != "" || po.PaymentTermsDaysSnapshot != nil {
		paymentTermsObj = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Days *int   `json:"days,omitempty"`
		}{
			ID:   safePtrString(po.PaymentTermsID),
			Name: strings.TrimSpace(po.PaymentTermsNameSnapshot),
			Days: po.PaymentTermsDaysSnapshot,
		}
	}

	businessUnitObj := any(po.BusinessUnit)
	if strings.TrimSpace(po.BusinessUnitNameSnapshot) != "" {
		businessUnitObj = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   safePtrString(po.BusinessUnitID),
			Name: strings.TrimSpace(po.BusinessUnitNameSnapshot),
		}
	}

	items := make([]dto.PurchaseOrderItemResponse, 0, len(po.Items))
	for _, it := range po.Items {
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
		items = append(items, dto.PurchaseOrderItemResponse{
			ID:        it.ID,
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Price:     it.Price,
			Discount:  it.Discount,
			Subtotal:  it.Subtotal,
			Notes:     it.Notes,
			Product:   productObj,
		})
	}

	var prDetailObj interface{}
	if po.PurchaseRequisition != nil {
		prDetailObj = &struct {
			ID   string `json:"id"`
			Code string `json:"code"`
		}{
			ID:   po.PurchaseRequisition.ID,
			Code: po.PurchaseRequisition.Code,
		}
	}

	return &dto.PurchaseOrderDetailResponse{
		ID:                    po.ID,
		Code:                  po.Code,
		CompanyID:             po.CompanyID,
		FiscalYearID:          po.FiscalYearID,
		SupplierID:            po.SupplierID,
		PaymentTermsID:        po.PaymentTermsID,
		BusinessUnitID:        po.BusinessUnitID,
		CreatedBy:             po.CreatedBy,
		PurchaseRequisitionID: po.PurchaseRequisitionID,
		SalesOrderID:          po.SalesOrderID,
		OrderDate:             po.OrderDate,
		DueDate:               po.DueDate,
		RevisionComment:       po.RevisionComment,
		Notes:                 po.Notes,
		Status:                string(po.Status),
		TaxRate:               po.TaxRate,
		TaxAmount:             po.TaxAmount,
		DeliveryCost:          po.DeliveryCost,
		OtherCost:             po.OtherCost,
		SubTotal:              po.SubTotal,
		TotalAmount:           po.TotalAmount,
		Supplier:              supplierObj,
		PaymentTerms:          paymentTermsObj,
		BusinessUnit:          businessUnitObj,
		Creator:               po.Creator,
		PurchaseRequisition:   prDetailObj,
		Items:                 items,
		CreatedAt:             po.CreatedAt,
		UpdatedAt:             po.UpdatedAt,
		SubmittedAt:           po.SubmittedAt,
		ApprovedAt:            po.ApprovedAt,
		ClosedAt:              po.ClosedAt,
	}
}
