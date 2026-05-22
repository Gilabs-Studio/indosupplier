package mapper

import (
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
)

type GoodsReceiptMapper struct{}

func NewGoodsReceiptMapper() *GoodsReceiptMapper {
	return &GoodsReceiptMapper{}
}

func (m *GoodsReceiptMapper) ToListResponse(gr *models.GoodsReceipt) *dto.GoodsReceiptListResponse {
	if gr == nil {
		return nil
	}
	var receiptDate *string
	if gr.ReceiptDate != nil {
		s := gr.ReceiptDate.Format(time.RFC3339)
		receiptDate = &s
	}

	var totalItemsReceived float64
	for _, item := range gr.Items {
		totalItemsReceived += item.QuantityReceived
	}

	resp := &dto.GoodsReceiptListResponse{
		ID:                 gr.ID,
		Code:               gr.Code,
		CompanyID:          gr.CompanyID,
		FiscalYearID:       gr.FiscalYearID,
		ReceiptDate:        receiptDate,
		WarehouseID:        gr.WarehouseID,
		Notes:              gr.Notes,
		ProofImageURL:      gr.ProofImageURL,
		Status:             string(gr.Status),
		CreatedBy:          gr.CreatedBy,
		TotalItemsReceived: totalItemsReceived,

		SubmittedAt:                  gr.SubmittedAt,
		ApprovedAt:                   gr.ApprovedAt,
		ClosedAt:                     gr.ClosedAt,
		RejectedAt:                   gr.RejectedAt,
		ConvertedAt:                  gr.ConvertedAt,
		ConvertedToSupplierInvoiceID: gr.ConvertedToSupplierInvoiceID,
		JournalEntryID:               gr.JournalEntryID,
	}

	if gr.PurchaseOrder != nil {
		resp.PurchaseOrder = &dto.GoodsReceiptPurchaseOrderMini{ID: gr.PurchaseOrder.ID, Code: gr.PurchaseOrder.Code}
	}
	if gr.Warehouse != nil {
		resp.Warehouse = &dto.GoodsReceiptWarehouseMini{ID: gr.Warehouse.ID, Name: gr.Warehouse.Name}
	}
	if strings.TrimSpace(gr.SupplierNameSnapshot) != "" || strings.TrimSpace(gr.SupplierCodeSnapshot) != "" {
		resp.Supplier = &dto.GoodsReceiptSupplierMini{ID: gr.SupplierID, Name: strings.TrimSpace(gr.SupplierNameSnapshot)}
	} else if gr.Supplier != nil {
		resp.Supplier = &dto.GoodsReceiptSupplierMini{ID: gr.Supplier.ID, Name: gr.Supplier.Name}
	}

	return resp
}

func (m *GoodsReceiptMapper) ToListResponseList(items []*models.GoodsReceipt) []*dto.GoodsReceiptListResponse {
	res := make([]*dto.GoodsReceiptListResponse, 0, len(items))
	for _, it := range items {
		res = append(res, m.ToListResponse(it))
	}
	return res
}

func (m *GoodsReceiptMapper) ToDetailResponse(gr *models.GoodsReceipt) *dto.GoodsReceiptDetailResponse {
	if gr == nil {
		return nil
	}
	var receiptDate *string
	if gr.ReceiptDate != nil {
		s := gr.ReceiptDate.Format(time.RFC3339)
		receiptDate = &s
	}

	resp := &dto.GoodsReceiptDetailResponse{
		ID:            gr.ID,
		Code:          gr.Code,
		CompanyID:     gr.CompanyID,
		FiscalYearID:  gr.FiscalYearID,
		ReceiptDate:   receiptDate,
		WarehouseID:   gr.WarehouseID,
		Notes:         gr.Notes,
		ProofImageURL: gr.ProofImageURL,
		Status:        string(gr.Status),
		CreatedBy:     gr.CreatedBy,
		Items:         make([]dto.GoodsReceiptItemResponse, 0, len(gr.Items)),

		SubmittedAt:                  gr.SubmittedAt,
		ApprovedAt:                   gr.ApprovedAt,
		ClosedAt:                     gr.ClosedAt,
		RejectedAt:                   gr.RejectedAt,
		ConvertedAt:                  gr.ConvertedAt,
		ConvertedToSupplierInvoiceID: gr.ConvertedToSupplierInvoiceID,
		JournalEntryID:               gr.JournalEntryID,
	}

	if gr.PurchaseOrder != nil {
		resp.PurchaseOrder = &dto.GoodsReceiptPurchaseOrderDetail{ID: gr.PurchaseOrder.ID, Code: gr.PurchaseOrder.Code, Status: string(gr.PurchaseOrder.Status)}
	}
	if gr.Warehouse != nil {
		resp.Warehouse = &dto.GoodsReceiptWarehouseMini{ID: gr.Warehouse.ID, Name: gr.Warehouse.Name}
	}
	if strings.TrimSpace(gr.SupplierNameSnapshot) != "" || strings.TrimSpace(gr.SupplierCodeSnapshot) != "" {
		resp.Supplier = &dto.GoodsReceiptSupplierMini{ID: gr.SupplierID, Name: strings.TrimSpace(gr.SupplierNameSnapshot)}
	} else if gr.Supplier != nil {
		resp.Supplier = &dto.GoodsReceiptSupplierMini{ID: gr.Supplier.ID, Name: gr.Supplier.Name}
	}

	for _, it := range gr.Items {
		item := dto.GoodsReceiptItemResponse{
			ID:                  it.ID,
			PurchaseOrderItemID: it.PurchaseOrderItemID,
			QuantityReceived:    it.QuantityReceived,
			Notes:               it.Notes,
		}
		if strings.TrimSpace(it.ProductNameSnapshot) != "" || strings.TrimSpace(it.ProductCodeSnapshot) != "" {
			sku := (*string)(nil)
			if it.Product != nil && it.Product.Sku != "" {
				s := it.Product.Sku
				sku = &s
			}
			item.Product = &dto.ProductMini{ID: it.ProductID, Name: strings.TrimSpace(it.ProductNameSnapshot), SKU: sku}
		} else if it.Product != nil {
			sku := (*string)(nil)
			if it.Product.Sku != "" {
				s := it.Product.Sku
				sku = &s
			}
			item.Product = &dto.ProductMini{ID: it.Product.ID, Name: it.Product.Name, SKU: sku}
		}
		resp.Items = append(resp.Items, item)
	}

	return resp
}
