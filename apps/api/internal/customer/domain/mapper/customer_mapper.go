package mapper

import (
	coreDTO "github.com/gilabs/gims/api/internal/core/domain/dto"
	"github.com/gilabs/gims/api/internal/customer/data/models"
	"github.com/gilabs/gims/api/internal/customer/domain/dto"
	geographic "github.com/gilabs/gims/api/internal/geographic/data/models"
	supplierDTO "github.com/gilabs/gims/api/internal/supplier/domain/dto"
)

// ToCustomerResponse converts Customer model to response DTO
func ToCustomerResponse(m *models.Customer) dto.CustomerResponse {
	resp := dto.CustomerResponse{
		ID:             m.ID,
		Code:           m.Code,
		Name:           m.Name,
		CustomerTypeID: m.CustomerTypeID,
		Address:        m.Address,
		ProvinceID:     m.ProvinceID,
		CityID:         m.CityID,
		DistrictID:     m.DistrictID,
		VillageID:      m.VillageID,
		VillageName:    m.VillageName,
		Email:          m.Email,
		Website:        m.Website,
		NPWP:           m.NPWP,
		ContactPerson:  m.ContactPerson,
		Notes:          m.Notes,
		Latitude:       m.Latitude,
		Longitude:      m.Longitude,
		CreatedBy:      m.CreatedBy,
		IsActive:       m.IsActive,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		// Sales defaults — FK IDs
		DefaultBusinessTypeID: m.DefaultBusinessTypeID,
		DefaultAreaID:         m.DefaultAreaID,
		DefaultSalesRepID:     m.DefaultSalesRepID,
		DefaultPaymentTermsID: m.DefaultPaymentTermsID,
		DefaultTaxRate:        m.DefaultTaxRate,
		CreditLimit:           m.CreditLimit,
		CreditIsActive:        m.CreditIsActive,
	}

	// Map CustomerType if loaded
	if m.CustomerType != nil {
		customerType := ToCustomerTypeResponse(m.CustomerType)
		resp.CustomerType = &customerType
	}

	// Map Village with nested district/city/province if loaded
	if m.Village != nil {
		resp.Village = toVillageResponse(m.Village)
	}

	// Map other geographic relations if loaded directly
	if m.Province != nil {
		resp.Province = &dto.ProvinceResponse{ID: m.Province.ID, Name: m.Province.Name}
	}
	if m.City != nil {
		resp.City = &dto.CityResponse{ID: m.City.ID, Name: m.City.Name}
	}
	if m.District != nil {
		resp.District = &dto.DistrictResponse{ID: m.District.ID, Name: m.District.Name}
	}

	// Map bank accounts if loaded
	if len(m.BankAccounts) > 0 {
		resp.BankAccounts = toCustomerBankResponseList(m.BankAccounts)
	}

	// Map sales default relations if loaded
	if m.DefaultBusinessType != nil {
		resp.DefaultBusinessType = &dto.SalesDefaultOptionBrief{
			ID:   m.DefaultBusinessType.ID,
			Name: m.DefaultBusinessType.Name,
		}
	}
	if m.DefaultArea != nil {
		resp.DefaultArea = &dto.SalesDefaultOptionBrief{
			ID:   m.DefaultArea.ID,
			Name: m.DefaultArea.Name,
		}
	}
	if m.DefaultSalesRep != nil {
		resp.DefaultSalesRep = &dto.SalesRepBrief{
			ID:           m.DefaultSalesRep.ID,
			EmployeeCode: m.DefaultSalesRep.EmployeeCode,
			Name:         m.DefaultSalesRep.Name,
		}
	}
	if m.DefaultPaymentTerms != nil {
		resp.DefaultPaymentTerms = &dto.SalesDefaultOptionBrief{
			ID:   m.DefaultPaymentTerms.ID,
			Name: m.DefaultPaymentTerms.Name,
		}
	}

	return resp
}

// ToCustomerResponseList converts a slice of Customer models to response DTOs
func ToCustomerResponseList(models []models.Customer) []dto.CustomerResponse {
	responses := make([]dto.CustomerResponse, len(models))
	for i := range models {
		responses[i] = ToCustomerResponse(&models[i])
	}
	return responses
}

// toCustomerBankResponse converts a single CustomerBank model to DTO
func toCustomerBankResponse(m *models.CustomerBank) dto.CustomerBankResponse {
	resp := dto.CustomerBankResponse{
		ID:            m.ID,
		CustomerID:    m.CustomerID,
		BankID:        m.BankID,
		CurrencyID:    m.CurrencyID,
		AccountNumber: m.AccountNumber,
		AccountName:   m.AccountName,
		Branch:        m.Branch,
		IsPrimary:     m.IsPrimary,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}

	if m.Bank != nil {
		bank := supplierDTO.BankResponse{
			ID:        m.Bank.ID,
			Name:      m.Bank.Name,
			Code:      m.Bank.Code,
			SwiftCode: m.Bank.SwiftCode,
			IsActive:  m.Bank.IsActive,
			CreatedAt: m.Bank.CreatedAt,
			UpdatedAt: m.Bank.UpdatedAt,
		}
		resp.Bank = &bank
	}
	if m.Currency != nil {
		currency := coreDTO.CurrencyResponse{
			ID:            m.Currency.ID,
			Code:          m.Currency.Code,
			Name:          m.Currency.Name,
			Symbol:        m.Currency.Symbol,
			DecimalPlaces: m.Currency.DecimalPlaces,
			IsActive:      m.Currency.IsActive,
			CreatedAt:     m.Currency.CreatedAt,
			UpdatedAt:     m.Currency.UpdatedAt,
		}
		resp.Currency = &currency
	}

	return resp
}

func toCustomerBankResponseList(models []models.CustomerBank) []dto.CustomerBankResponse {
	responses := make([]dto.CustomerBankResponse, len(models))
	for i := range models {
		responses[i] = toCustomerBankResponse(&models[i])
	}
	return responses
}

// toVillageResponse maps nested geographic Village chain
func toVillageResponse(v *geographic.Village) *dto.VillageResponse {
	if v == nil {
		return nil
	}

	village := &dto.VillageResponse{
		ID:   v.ID,
		Name: v.Name,
	}

	if v.District != nil {
		village.District = &dto.DistrictResponse{
			ID:   v.District.ID,
			Name: v.District.Name,
		}
		if v.District.City != nil {
			village.District.City = &dto.CityResponse{
				ID:   v.District.City.ID,
				Name: v.District.City.Name,
			}
			if v.District.City.Province != nil {
				village.District.City.Province = &dto.ProvinceResponse{
					ID:   v.District.City.Province.ID,
					Name: v.District.City.Province.Name,
				}
			}
		}
	}

	return village
}
