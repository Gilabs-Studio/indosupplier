package mapper

import (
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/purchase/data/models"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
)

type PurchaseRequisitionMapper struct{}

func NewPurchaseRequisitionMapper() *PurchaseRequisitionMapper {
	return &PurchaseRequisitionMapper{}
}

func (m *PurchaseRequisitionMapper) ToListResponse(pr *models.PurchaseRequisition) *dto.PurchaseRequisitionListResponse {
	if pr == nil {
		return nil
	}

	resp := &dto.PurchaseRequisitionListResponse{
		ID:                         pr.ID,
		Code:                       pr.Code,
		SupplierID:                 pr.SupplierID,
		PaymentTermsID:             pr.PaymentTermsID,
		BusinessUnitID:             pr.BusinessUnitID,
		CompanyID:                  pr.CompanyID,
		FiscalYearID:               pr.FiscalYearID,
		RequestedBy:                pr.EmployeeID,
		RequestDate:                pr.RequestDate,
		Status:                     string(pr.Status),
		Subtotal:                   pr.Subtotal,
		TaxRate:                    pr.TaxRate,
		TaxAmount:                  pr.TaxAmount,
		DeliveryCost:               pr.DeliveryCost,
		OtherCost:                  pr.OtherCost,
		TotalAmount:                pr.TotalAmount,
		Notes:                      pr.Notes,
		ConvertedToPurchaseOrderID: pr.ConvertedToPurchaseOrderID,
		CreatedAt:                  pr.CreatedAt.In(apptime.Location()).Format(time.RFC3339),
		UpdatedAt:                  pr.UpdatedAt.In(apptime.Location()).Format(time.RFC3339),
	}

	if pr.SubmittedAt != nil {
		s := pr.SubmittedAt.In(apptime.Location()).Format(time.RFC3339)
		resp.SubmittedAt = &s
	}
	if pr.ApprovedAt != nil {
		s := pr.ApprovedAt.In(apptime.Location()).Format(time.RFC3339)
		resp.ApprovedAt = &s
	}
	if pr.RejectedAt != nil {
		s := pr.RejectedAt.In(apptime.Location()).Format(time.RFC3339)
		resp.RejectedAt = &s
	}
	if pr.ConvertedAt != nil {
		s := pr.ConvertedAt.In(apptime.Location()).Format(time.RFC3339)
		resp.ConvertedAt = &s
	}

	if pr.SupplierID != nil && *pr.SupplierID != "" && (pr.SupplierCodeSnapshot != "" || pr.SupplierNameSnapshot != "") {
		resp.Supplier = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Code string `json:"code"`
		}{
			ID:   *pr.SupplierID,
			Name: pr.SupplierNameSnapshot,
			Code: pr.SupplierCodeSnapshot,
		}
	} else if pr.Supplier != nil {
		resp.Supplier = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Code string `json:"code"`
		}{
			ID:   pr.Supplier.ID,
			Name: pr.Supplier.Name,
			Code: pr.Supplier.Code,
		}
	}

	if pr.PaymentTermsID != nil && *pr.PaymentTermsID != "" && (pr.PaymentTermsNameSnapshot != "" || pr.PaymentTermsDaysSnapshot != nil) {
		resp.PaymentTerms = &dto.PurchaseRequisitionPaymentTermsMini{ID: *pr.PaymentTermsID, Name: pr.PaymentTermsNameSnapshot, Days: pr.PaymentTermsDaysSnapshot}
	} else if pr.PaymentTerms != nil {
		days := pr.PaymentTerms.Days
		resp.PaymentTerms = &dto.PurchaseRequisitionPaymentTermsMini{ID: pr.PaymentTerms.ID, Name: pr.PaymentTerms.Name, Days: &days}
	}

	if pr.BusinessUnitID != nil && *pr.BusinessUnitID != "" && pr.BusinessUnitNameSnapshot != "" {
		resp.BusinessUnit = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   *pr.BusinessUnitID,
			Name: pr.BusinessUnitNameSnapshot,
		}
	} else if pr.BusinessUnit != nil {
		resp.BusinessUnit = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   pr.BusinessUnit.ID,
			Name: pr.BusinessUnit.Name,
		}
	}

	if pr.Employee != nil && pr.Employee.User != nil {
		resp.User = &struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}{
			ID:    pr.Employee.User.ID,
			Email: pr.Employee.User.Email,
			Name:  pr.Employee.User.Name,
		}
	}

	if pr.Employee != nil {
		resp.Employee = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   pr.Employee.ID,
			Name: pr.Employee.Name,
		}
	}

	return resp
}

func (m *PurchaseRequisitionMapper) ToListResponseList(items []*models.PurchaseRequisition) []*dto.PurchaseRequisitionListResponse {
	res := make([]*dto.PurchaseRequisitionListResponse, 0, len(items))
	for _, it := range items {
		res = append(res, m.ToListResponse(it))
	}
	return res
}

func (m *PurchaseRequisitionMapper) ToDetailResponse(pr *models.PurchaseRequisition) *dto.PurchaseRequisitionDetailResponse {
	if pr == nil {
		return nil
	}

	resp := &dto.PurchaseRequisitionDetailResponse{
		ID:                         pr.ID,
		Code:                       pr.Code,
		SupplierID:                 pr.SupplierID,
		PaymentTermsID:             pr.PaymentTermsID,
		BusinessUnitID:             pr.BusinessUnitID,
		CompanyID:                  pr.CompanyID,
		FiscalYearID:               pr.FiscalYearID,
		EmployeeID:                 pr.EmployeeID,
		RequestDate:                pr.RequestDate,
		Address:                    pr.Address,
		Notes:                      pr.Notes,
		Status:                     string(pr.Status),
		Subtotal:                   pr.Subtotal,
		TaxRate:                    pr.TaxRate,
		TaxAmount:                  pr.TaxAmount,
		DeliveryCost:               pr.DeliveryCost,
		OtherCost:                  pr.OtherCost,
		TotalAmount:                pr.TotalAmount,
		ConvertedToPurchaseOrderID: pr.ConvertedToPurchaseOrderID,
		CreatedAt:                  pr.CreatedAt.In(apptime.Location()).Format(time.RFC3339),
		UpdatedAt:                  pr.UpdatedAt.In(apptime.Location()).Format(time.RFC3339),
		Items:                      make([]dto.PurchaseRequisitionItemResponse, 0, len(pr.Items)),
	}

	if pr.SubmittedAt != nil {
		s := pr.SubmittedAt.In(apptime.Location()).Format(time.RFC3339)
		resp.SubmittedAt = &s
	}
	if pr.ApprovedAt != nil {
		s := pr.ApprovedAt.In(apptime.Location()).Format(time.RFC3339)
		resp.ApprovedAt = &s
	}
	if pr.RejectedAt != nil {
		s := pr.RejectedAt.In(apptime.Location()).Format(time.RFC3339)
		resp.RejectedAt = &s
	}
	if pr.ConvertedAt != nil {
		s := pr.ConvertedAt.In(apptime.Location()).Format(time.RFC3339)
		resp.ConvertedAt = &s
	}

	if pr.SupplierID != nil && *pr.SupplierID != "" && (pr.SupplierCodeSnapshot != "" || pr.SupplierNameSnapshot != "") {
		resp.Supplier = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Code string `json:"code"`
		}{
			ID:   *pr.SupplierID,
			Name: pr.SupplierNameSnapshot,
			Code: pr.SupplierCodeSnapshot,
		}
	} else if pr.Supplier != nil {
		resp.Supplier = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Code string `json:"code"`
		}{
			ID:   pr.Supplier.ID,
			Name: pr.Supplier.Name,
			Code: pr.Supplier.Code,
		}
	}

	if pr.PaymentTermsID != nil && *pr.PaymentTermsID != "" && (pr.PaymentTermsNameSnapshot != "" || pr.PaymentTermsDaysSnapshot != nil) {
		resp.PaymentTerms = &dto.PurchaseRequisitionPaymentTermsMini{ID: *pr.PaymentTermsID, Name: pr.PaymentTermsNameSnapshot, Days: pr.PaymentTermsDaysSnapshot}
	} else if pr.PaymentTerms != nil {
		days := pr.PaymentTerms.Days
		resp.PaymentTerms = &dto.PurchaseRequisitionPaymentTermsMini{ID: pr.PaymentTerms.ID, Name: pr.PaymentTerms.Name, Days: &days}
	}

	if pr.BusinessUnitID != nil && *pr.BusinessUnitID != "" && pr.BusinessUnitNameSnapshot != "" {
		resp.BusinessUnit = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   *pr.BusinessUnitID,
			Name: pr.BusinessUnitNameSnapshot,
		}
	} else if pr.BusinessUnit != nil {
		resp.BusinessUnit = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   pr.BusinessUnit.ID,
			Name: pr.BusinessUnit.Name,
		}
	}

	if pr.Employee != nil {
		resp.Employee = &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   pr.Employee.ID,
			Name: pr.Employee.Name,
		}
	}

	if pr.Employee != nil && pr.Employee.User != nil {
		resp.User = &struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}{
			ID:    pr.Employee.User.ID,
			Email: pr.Employee.User.Email,
			Name:  pr.Employee.User.Name,
		}
	}

	for _, it := range pr.Items {
		itemResp := dto.PurchaseRequisitionItemResponse{
			ID:            it.ID,
			ProductID:     it.ProductID,
			Quantity:      it.Quantity,
			PurchasePrice: it.PurchasePrice,
			Discount:      it.Discount,
			Subtotal:      it.Subtotal,
			Notes:         it.Notes,
		}
		if it.ProductID != "" && (it.ProductCodeSnapshot != "" || it.ProductNameSnapshot != "") {
			itemResp.Product = &struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Code string `json:"code"`
			}{
				ID:   it.ProductID,
				Name: it.ProductNameSnapshot,
				Code: it.ProductCodeSnapshot,
			}
		} else if it.Product != nil {
			itemResp.Product = &struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Code string `json:"code"`
			}{
				ID:   it.Product.ID,
				Name: it.Product.Name,
				Code: it.Product.Code,
			}
		}
		resp.Items = append(resp.Items, itemResp)
	}

	return resp
}
