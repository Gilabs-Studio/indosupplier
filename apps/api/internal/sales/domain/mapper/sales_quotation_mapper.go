package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/dto"
)

// ToSalesQuotationResponse converts a SalesQuotation model to response DTO
func ToSalesQuotationResponse(m *salesModels.SalesQuotation) dto.SalesQuotationResponse {
	response := dto.SalesQuotationResponse{
		ID:                  m.ID,
		Code:                m.Code,
		QuotationDate:       m.QuotationDate.Format("2006-01-02"),
		Subtotal:            m.Subtotal,
		DiscountAmount:      m.DiscountAmount,
		TaxRate:             m.TaxRate,
		TaxAmount:           m.TaxAmount,
		DeliveryCost:        m.DeliveryCost,
		OtherCost:           m.OtherCost,
		TotalAmount:         m.TotalAmount,
		Status:              string(m.Status),
		CustomerID:          m.CustomerID,
		CustomerContactID:   m.CustomerContactID,
		CustomerName:        m.CustomerName,
		CustomerContact:     m.CustomerContact,
		CustomerPhone:       m.CustomerPhone,
		CustomerEmail:       m.CustomerEmail,
		Notes:               m.Notes,
		CreatedAt:           m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           m.UpdatedAt.Format(time.RFC3339),
	}

	if m.Customer != nil {
		response.Customer = &dto.CustomerResponse{
			ID:             m.Customer.ID,
			Code:           m.Customer.Code,
			Name:           m.Customer.Name,
			CustomerTypeID: m.Customer.CustomerTypeID,
			Address:        m.Customer.Address,
			Email:          m.Customer.Email,
			ContactPerson:  m.Customer.ContactPerson,
		}
	}

	if m.ValidUntil != nil {
		validUntil := m.ValidUntil.Format("2006-01-02")
		response.ValidUntil = &validUntil
	}

	if m.PaymentTermsID != nil {
		response.PaymentTermsID = m.PaymentTermsID
		if m.PaymentTerms != nil {
			response.PaymentTerms = &dto.PaymentTermsResponse{
				ID:          m.PaymentTerms.ID,
				Code:        m.PaymentTerms.Code,
				Name:        m.PaymentTerms.Name,
				Description: m.PaymentTerms.Description,
				Days:        m.PaymentTerms.Days,
			}
		}
	}

	if m.SalesRepID != nil {
		response.SalesRepID = m.SalesRepID
		if m.SalesRep != nil {
			response.SalesRep = &dto.EmployeeResponse{
				ID:           m.SalesRep.ID,
				EmployeeCode: m.SalesRep.EmployeeCode,
				Name:         m.SalesRep.Name,
				Email:        m.SalesRep.Email,
				Phone:        m.SalesRep.Phone,
			}
		}
	}

	if m.BusinessUnitID != nil {
		response.BusinessUnitID = m.BusinessUnitID
		if m.BusinessUnit != nil {
			response.BusinessUnit = &dto.BusinessUnitResponse{
				ID:          m.BusinessUnit.ID,
				Name:        m.BusinessUnit.Name,
				Description: m.BusinessUnit.Description,
			}
		}
	}

	if m.BusinessTypeID != nil {
		response.BusinessTypeID = m.BusinessTypeID
		if m.BusinessType != nil {
			response.BusinessType = &dto.BusinessTypeResponse{
				ID:          m.BusinessType.ID,
				Name:        m.BusinessType.Name,
				Description: m.BusinessType.Description,
			}
		}
	}

	if m.CreatedBy != nil {
		response.CreatedBy = m.CreatedBy
	}

	if m.ApprovedBy != nil {
		response.ApprovedBy = m.ApprovedBy
		if m.ApprovedAt != nil {
			approvedAt := m.ApprovedAt.Format(time.RFC3339)
			response.ApprovedAt = &approvedAt
		}
	}

	if m.RejectedBy != nil {
		response.RejectedBy = m.RejectedBy
		if m.RejectedAt != nil {
			rejectedAt := m.RejectedAt.Format(time.RFC3339)
			response.RejectedAt = &rejectedAt
		}
		response.RejectionReason = m.RejectionReason
	}

	if m.SourceDealID != nil {
		response.SourceDealID = m.SourceDealID
	}

	if m.ConvertedToSalesOrderID != nil {
		response.ConvertedToSalesOrderID = m.ConvertedToSalesOrderID
		if m.ConvertedAt != nil {
			convertedAt := m.ConvertedAt.Format(time.RFC3339)
			response.ConvertedAt = &convertedAt
		}
	}

	// Map items
	if len(m.Items) > 0 {
		response.Items = make([]dto.SalesQuotationItemResponse, len(m.Items))
		for i, item := range m.Items {
			response.Items[i] = ToSalesQuotationItemResponse(&item)
		}
	}

	return response
}

// ToSalesQuotationItemResponse converts a SalesQuotationItem model to response DTO
func ToSalesQuotationItemResponse(m *salesModels.SalesQuotationItem) dto.SalesQuotationItemResponse {
	response := dto.SalesQuotationItemResponse{
		ID:               m.ID,
		SalesQuotationID: m.SalesQuotationID,
		ProductID:        m.ProductID,
		Quantity:         m.Quantity,
		Price:            m.Price,
		Discount:         m.Discount,
		Subtotal:         m.Subtotal,
		CreatedAt:        m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        m.UpdatedAt.Format(time.RFC3339),
	}

	if m.Product != nil {
		response.Product = &dto.ProductResponse{
			ID:           m.Product.ID,
			Code:         m.Product.Code,
			Name:         m.Product.Name,
			SellingPrice: m.Product.SellingPrice,
			ImageURL:     m.Product.ImageURL,
		}
	}

	return response
}

// ToSalesQuotationModel converts a CreateSalesQuotationRequest to SalesQuotation model
func ToSalesQuotationModel(req *dto.CreateSalesQuotationRequest, code string, createdBy *string) (*salesModels.SalesQuotation, error) {
	quotationDate, err := time.Parse("2006-01-02", req.QuotationDate)
	if err != nil {
		return nil, err
	}

	paymentTermsID := req.PaymentTermsID
	salesRepID := req.SalesRepID
	businessUnitID := req.BusinessUnitID

	quotation := &salesModels.SalesQuotation{
		Code:            code,
		QuotationDate:   quotationDate,
		PaymentTermsID:  &paymentTermsID,
		SalesRepID:      &salesRepID,
		BusinessUnitID:  &businessUnitID,
		BusinessTypeID:  req.BusinessTypeID,
		CustomerID:      req.CustomerID,
		CustomerContactID: req.CustomerContactID,
		CustomerName:    req.CustomerName,
		CustomerContact: req.CustomerContact,
		CustomerPhone:   req.CustomerPhone,
		CustomerEmail:   req.CustomerEmail,
		TaxRate:         req.TaxRate,
		DeliveryCost:    req.DeliveryCost,
		OtherCost:       req.OtherCost,
		DiscountAmount:  req.DiscountAmount,
		Notes:           req.Notes,
		Status:          salesModels.SalesQuotationStatusDraft,
		CreatedBy:       createdBy,
		CreatedAt:       apptime.Now(),
		UpdatedAt:       apptime.Now(),
	}

	if req.ValidUntil != nil {
		validUntil, err := time.Parse("2006-01-02", *req.ValidUntil)
		if err == nil {
			quotation.ValidUntil = &validUntil
		}
	}

	// Default tax rate to 11% if not provided
	if req.TaxRate == 0 {
		quotation.TaxRate = 11.00
	}

	// Map items
	if len(req.Items) > 0 {
		quotation.Items = make([]salesModels.SalesQuotationItem, len(req.Items))
		for i, itemReq := range req.Items {
			quotation.Items[i] = salesModels.SalesQuotationItem{
				ProductID: itemReq.ProductID,
				Quantity:  itemReq.Quantity,
				Price:     itemReq.Price,
				Discount:  itemReq.Discount,
				CreatedAt: apptime.Now(),
				UpdatedAt: apptime.Now(),
			}
			quotation.Items[i].CalculateSubtotal()
		}
	}

	return quotation, nil
}

// UpdateSalesQuotationModel updates a SalesQuotation model from UpdateSalesQuotationRequest
func UpdateSalesQuotationModel(m *salesModels.SalesQuotation, req *dto.UpdateSalesQuotationRequest) error {
	if req.QuotationDate != nil {
		quotationDate, err := time.Parse("2006-01-02", *req.QuotationDate)
		if err != nil {
			return err
		}
		m.QuotationDate = quotationDate
	}

	if req.ValidUntil != nil {
		validUntil, err := time.Parse("2006-01-02", *req.ValidUntil)
		if err == nil {
			m.ValidUntil = &validUntil
		}
	}

	if req.PaymentTermsID != nil {
		m.PaymentTermsID = req.PaymentTermsID
	}

	if req.SalesRepID != nil {
		m.SalesRepID = req.SalesRepID
	}

	if req.BusinessUnitID != nil {
		m.BusinessUnitID = req.BusinessUnitID
	}

	if req.BusinessTypeID != nil {
		m.BusinessTypeID = req.BusinessTypeID
	}

	if req.CustomerID != nil {
		m.CustomerID = req.CustomerID
	}

	if req.CustomerContactID != nil {
		m.CustomerContactID = req.CustomerContactID
	}

	if req.CustomerName != nil {
		m.CustomerName = *req.CustomerName
	}

	if req.CustomerContact != nil {
		m.CustomerContact = *req.CustomerContact
	}

	if req.CustomerPhone != nil {
		m.CustomerPhone = *req.CustomerPhone
	}

	if req.CustomerEmail != nil {
		m.CustomerEmail = *req.CustomerEmail
	}

	if req.TaxRate != nil {
		m.TaxRate = *req.TaxRate
	}

	if req.DeliveryCost != nil {
		m.DeliveryCost = *req.DeliveryCost
	}

	if req.OtherCost != nil {
		m.OtherCost = *req.OtherCost
	}

	if req.DiscountAmount != nil {
		m.DiscountAmount = *req.DiscountAmount
	}

	if req.Notes != nil {
		m.Notes = *req.Notes
	}

	// Update items if provided
	if req.Items != nil && len(*req.Items) > 0 {
		m.Items = make([]salesModels.SalesQuotationItem, len(*req.Items))
		for i, itemReq := range *req.Items {
			m.Items[i] = salesModels.SalesQuotationItem{
				ProductID: itemReq.ProductID,
				Quantity:  itemReq.Quantity,
				Price:     itemReq.Price,
				Discount:  itemReq.Discount,
				UpdatedAt: apptime.Now(),
			}
			m.Items[i].CalculateSubtotal()
		}
	}

	m.UpdatedAt = apptime.Now()
	return nil
}
