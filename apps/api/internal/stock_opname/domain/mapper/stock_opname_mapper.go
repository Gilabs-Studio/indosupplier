package mapper

import (
	"time"

	employeeModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/stock_opname/data/models"
	"github.com/gilabs/gims/api/internal/stock_opname/domain/dto"
)

func ToStockOpnameResponse(m *models.StockOpname) dto.StockOpnameResponse {
	totalNegativeVariance, totalPositiveVariance := calculateVarianceSummary(m.Items)

	return dto.StockOpnameResponse{
		ID:                       m.ID,
		OpnameNumber:             m.OpnameNumber,
		WarehouseID:              m.WarehouseID,
		WarehouseName:            getWarehouseName(m),
		JournalID:                m.JournalID,
		Date:                     m.Date,
		Status:                   mapStatusForResponse(m.Status),
		Description:              m.Description,
		TotalItems:               m.TotalItems,
		TotalVarianceQty:         m.TotalVarianceQty,
		TotalNegativeVarianceQty: totalNegativeVariance,
		TotalPositiveVarianceQty: totalPositiveVariance,
		OrderedByID:              m.OrderedByID,
		OrderedByName:            getEmployeeName(m.OrderedBy),
		AssignedToID:             m.AssignedToID,
		AssignedToName:           getEmployeeName(m.AssignedTo),
		CreatedBy:                m.CreatedBy,
		CreatedAt:                m.CreatedAt,
		UpdatedAt:                m.UpdatedAt,
	}
}

func mapStatusForResponse(status models.StockOpnameStatus) dto.StockOpnameStatus {
	switch status {
	case models.StockOpnameStatusPending:
		return dto.StockOpnameStatusPendingApproval
	case models.StockOpnameStatusPosted:
		return dto.StockOpnameStatusCompleted
	default:
		return dto.StockOpnameStatus(status)
	}
}

func calculateVarianceSummary(items []models.StockOpnameItem) (float64, float64) {
	var totalNegativeVariance float64
	var totalPositiveVariance float64

	for _, item := range items {
		if item.VarianceQty < 0 {
			totalNegativeVariance += item.VarianceQty
			continue
		}

		if item.VarianceQty > 0 {
			totalPositiveVariance += item.VarianceQty
		}
	}

	return totalNegativeVariance, totalPositiveVariance
}

func getWarehouseName(m *models.StockOpname) string {
	if m.Warehouse != nil {
		return m.Warehouse.Name
	}
	return ""
}

func getEmployeeName(e *employeeModels.Employee) *string {
	if e == nil {
		return nil
	}
	name := e.Name
	return &name
}

func ToStockOpnameItemResponse(m *models.StockOpnameItem) dto.StockOpnameItemResponse {
	return dto.StockOpnameItemResponse{
		ID:               m.ID,
		StockOpnameID:    m.StockOpnameID,
		ProductID:        m.ProductID,
		ProductName:      m.Product.Name,
		ProductCode:      m.Product.Code,
		ProductImageURL:  m.Product.ImageURL,
		SystemQty:        m.SystemQty,
		PhysicalQty:      m.PhysicalQty,
		VarianceQty:      m.VarianceQty,
		UnitCost:         m.Product.CurrentHpp,
		Notes:            m.Notes,
		InventoryBatchID: m.InventoryBatchID,
		BatchNumber:      m.BatchNumber,
		BatchQty:         m.BatchQty,
	}
}

func getProductName(m *models.StockOpnameItem) string {
	if m.Product != nil {
		return m.Product.Name
	}
	return ""
}

func getProductCode(m *models.StockOpnameItem) string {
	if m.Product != nil {
		return m.Product.Code
	}
	return ""
}

func ToStockOpnameModel(req *dto.CreateStockOpnameRequest, opnameNumber string, createdBy *string) (*models.StockOpname, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}
	return &models.StockOpname{
		OpnameNumber: opnameNumber,
		WarehouseID:  req.WarehouseID,
		Date:         date,
		Description:  req.Description,
		Status:       models.StockOpnameStatusDraft,
		OrderedByID:  req.OrderedByID,
		AssignedToID: req.AssignedToID,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
	}, nil
}
