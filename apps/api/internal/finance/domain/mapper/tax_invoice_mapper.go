package mapper

import (
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/dto"
)

type TaxInvoiceMapper struct{}

func NewTaxInvoiceMapper() *TaxInvoiceMapper {
	return &TaxInvoiceMapper{}
}

func (m *TaxInvoiceMapper) ToResponse(item *financeModels.TaxInvoice) dto.TaxInvoiceResponse {
	if item == nil {
		return dto.TaxInvoiceResponse{}
	}
	return dto.TaxInvoiceResponse{
		ID: item.ID,
		TaxInvoiceNumber: item.TaxInvoiceNumber,
		TaxInvoiceDate: item.TaxInvoiceDate,
		CustomerInvoiceID: item.CustomerInvoiceID,
		SupplierInvoiceID: item.SupplierInvoiceID,
		DPPAmount: item.DPPAmount,
		VATAmount: item.VATAmount,
		TotalAmount: item.TotalAmount,
		Notes: item.Notes,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
