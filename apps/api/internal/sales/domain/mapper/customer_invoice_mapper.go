package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
)

// MapCustomerInvoiceToResponse maps CustomerInvoice model to response DTO
func MapCustomerInvoiceToResponse(invoice *models.CustomerInvoice) *dto.CustomerInvoiceResponse {
	if invoice == nil {
		return nil
	}

	resp := &dto.CustomerInvoiceResponse{
		ID:                invoice.ID,
		Code:              invoice.Code,
		InvoiceNumber:     invoice.InvoiceNumber,
		Type:              string(invoice.Type),
		InvoiceDate:       invoice.InvoiceDate.Format("2006-01-02"),
		SalesOrderID:      invoice.SalesOrderID,
		DeliveryOrderID:   invoice.DeliveryOrderID,
		PaymentTermsID:    invoice.PaymentTermsID,
		Subtotal:          invoice.Subtotal,
		TaxRate:           invoice.TaxRate,
		TaxAmount:         invoice.TaxAmount,
		DeliveryCost:      invoice.DeliveryCost,
		OtherCost:         invoice.OtherCost,
		DownPaymentAmount: invoice.DownPaymentAmount,
		Amount:            invoice.Amount,
		PaidAmount:        invoice.PaidAmount,
		RemainingAmount:   invoice.RemainingAmount,
		Status:            string(invoice.Status),
		Notes:             invoice.Notes,
		IsPosted:          invoice.IsPosted,
		JournalEntryID:    invoice.JournalEntryID,
		CreatedBy:         invoice.CreatedBy,
		CancelledBy:       invoice.CancelledBy,
		CreatedAt:         invoice.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         invoice.UpdatedAt.Format(time.RFC3339),
	}

	if invoice.DownPaymentInvoiceID != nil {
		resp.DownPaymentInvoiceID = invoice.DownPaymentInvoiceID
	}
	if invoice.DownPaymentInvoice != nil {
		resp.DownPaymentInvoiceCode = &invoice.DownPaymentInvoice.Code
	}

	// Map optional date fields
	if invoice.DueDate != nil {
		dueDate := invoice.DueDate.Format("2006-01-02")
		resp.DueDate = &dueDate
	}

	if invoice.PaymentAt != nil {
		paymentAt := invoice.PaymentAt.Format(time.RFC3339)
		resp.PaymentAt = &paymentAt
	}

	if invoice.CancelledAt != nil {
		cancelledAt := invoice.CancelledAt.Format(time.RFC3339)
		resp.CancelledAt = &cancelledAt
	}

	// Map PaymentTerms
	if invoice.PaymentTerms != nil {
		resp.PaymentTerms = &dto.PaymentTermsResponse{
			ID:          invoice.PaymentTerms.ID,
			Code:        invoice.PaymentTerms.Code,
			Name:        invoice.PaymentTerms.Name,
			Description: invoice.PaymentTerms.Description,
			Days:        invoice.PaymentTerms.Days,
		}
	}

	// Map SalesOrder
	if invoice.SalesOrder != nil {
		so := invoice.SalesOrder
		resp.SalesOrder = &dto.SalesOrderBriefResponse{
			ID:            so.ID,
			Code:          so.Code,
			CustomerID:    so.CustomerID,
			CustomerName:  so.CustomerName,
			CustomerPhone: so.CustomerPhone,
			CustomerEmail: so.CustomerEmail,
		}
		// Map nested Customer if preloaded
		if so.Customer != nil {
			resp.SalesOrder.Customer = &dto.CustomerResponse{
				ID:            so.Customer.ID,
				Code:          so.Customer.Code,
				Name:          so.Customer.Name,
				CustomerTypeID: so.Customer.CustomerTypeID,
				Address:       so.Customer.Address,
				Email:         so.Customer.Email,
				ContactPerson: so.Customer.ContactPerson,
			}
		}
	}

	// Map DeliveryOrder
	if invoice.DeliveryOrder != nil {
		resp.DeliveryOrder = &dto.DeliveryOrderBriefResponse{
			ID:   invoice.DeliveryOrder.ID,
			Code: invoice.DeliveryOrder.Code,
		}
	}

	// Map Items
	if len(invoice.Items) > 0 {
		resp.Items = make([]dto.CustomerInvoiceItemResponse, len(invoice.Items))
		for i, item := range invoice.Items {
			resp.Items[i] = *MapCustomerInvoiceItemToResponse(&item)
		}
	}

	return resp
}

// MapCustomerInvoiceItemToResponse maps CustomerInvoiceItem model to response DTO
func MapCustomerInvoiceItemToResponse(item *models.CustomerInvoiceItem) *dto.CustomerInvoiceItemResponse {
	if item == nil {
		return nil
	}

	resp := &dto.CustomerInvoiceItemResponse{
		ID:                  item.ID,
		CustomerInvoiceID:   item.CustomerInvoiceID,
		ProductID:           item.ProductID,
		SalesOrderItemID:    item.SalesOrderItemID,
		DeliveryOrderItemID: item.DeliveryOrderItemID,
		Quantity:            item.Quantity,
		Price:             item.Price,
		Discount:          item.Discount,
		Subtotal:          item.Subtotal,
		HPPAmount:         item.HPPAmount,
		CreatedAt:         item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         item.UpdatedAt.Format(time.RFC3339),
	}

	// Map Product
	if item.Product != nil {
		resp.Product = &dto.ProductResponse{
			ID:           item.Product.ID,
			Code:         item.Product.Code,
			Name:         item.Product.Name,
			SellingPrice: item.Product.SellingPrice,
			ImageURL:     item.Product.ImageURL,
		}
	}

	return resp
}

// MapCustomerInvoicesToResponse maps a slice of CustomerInvoice models to response DTOs
func MapCustomerInvoicesToResponse(invoices []models.CustomerInvoice) []dto.CustomerInvoiceResponse {
	result := make([]dto.CustomerInvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		result[i] = *MapCustomerInvoiceToResponse(&invoice)
	}
	return result
}

// MapCustomerInvoiceItemsToResponse maps a slice of CustomerInvoiceItem models to response DTOs
func MapCustomerInvoiceItemsToResponse(items []models.CustomerInvoiceItem) []dto.CustomerInvoiceItemResponse {
	result := make([]dto.CustomerInvoiceItemResponse, len(items))
	for i, item := range items {
		result[i] = *MapCustomerInvoiceItemToResponse(&item)
	}
	return result
}
