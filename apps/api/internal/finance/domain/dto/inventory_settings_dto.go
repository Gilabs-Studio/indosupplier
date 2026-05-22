package dto

import (
	"time"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
)

type UpdateInventorySettingsRequest struct {
	CompanyID       string                                 `json:"company_id" binding:"required,uuid"`
	ValuationMethod financeModels.InventoryValuationMethod `json:"valuation_method" binding:"required,oneof=average_cost fifo specific_identification"`
}

type InventorySettingsResponse struct {
	ID              string                                 `json:"id"`
	CompanyID       string                                 `json:"company_id"`
	ValuationMethod financeModels.InventoryValuationMethod `json:"valuation_method"`
	IsLocked        bool                                   `json:"is_locked"`
	CreatedAt       time.Time                              `json:"created_at"`
	UpdatedAt       time.Time                              `json:"updated_at"`
}

type InventoryAverageCostResponse struct {
	CompanyID     string    `json:"company_id"`
	ProductID     string    `json:"product_id"`
	AverageCost   float64   `json:"average_cost"`
	TotalQuantity float64   `json:"total_quantity"`
	TotalValue    float64   `json:"total_value"`
	LastUpdated   time.Time `json:"last_updated"`
}
