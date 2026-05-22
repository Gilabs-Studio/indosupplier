package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/inventory/data/models"
	"github.com/gilabs/gims/api/internal/inventory/domain/dto"
)

func (r *inventoryRepository) GetProductLedgers(ctx context.Context, productID string, req *dto.GetProductStockLedgersRequest) ([]models.StockLedger, int64, error) {
	items := make([]models.StockLedger, 0)

	type transactionPageRow struct {
		TransactionID string    `gorm:"column:transaction_id"`
		LastCreated   time.Time `gorm:"column:last_created"`
	}

	// Build base query with filters
	baseQuery := r.DB(ctx).Model(&models.StockLedger{}).Where("product_id = ?", productID)
	if req.TransactionType != "" {
		baseQuery = baseQuery.Where("transaction_type = ?", strings.ToUpper(strings.TrimSpace(req.TransactionType)))
	}
	if req.ParsedDateFrom != nil {
		baseQuery = baseQuery.Where("DATE(created_at) >= ?", req.ParsedDateFrom.Format("2006-01-02"))
	}
	if req.ParsedDateTo != nil {
		baseQuery = baseQuery.Where("DATE(created_at) <= ?", req.ParsedDateTo.Format("2006-01-02"))
	}

	// Count distinct transactions for pagination
	var totalDistinct int64
	if err := baseQuery.Distinct("transaction_id").Count(&totalDistinct).Error; err != nil {
		return nil, 0, err
	}

	// Enforce maximum limit to prevent excessive memory allocation (DoS mitigation)
	const MaxLimit = 1000
	if req.Limit > MaxLimit {
		req.Limit = MaxLimit
	}

	// Get transaction IDs for current page ordered by latest created_at
	offset := (req.Page - 1) * req.Limit
	pageRows := make([]transactionPageRow, 0, req.Limit)
	if err := baseQuery.
		Select("transaction_id, MAX(created_at) as last_created").
		Group("transaction_id").
		Order("MAX(created_at) DESC").
		Limit(req.Limit).
		Offset(offset).
		Scan(&pageRows).Error; err != nil {
		return nil, 0, err
	}

	txIDs := make([]string, 0, len(pageRows))
	for _, row := range pageRows {
		if row.TransactionID != "" {
			txIDs = append(txIDs, row.TransactionID)
		}
	}

	if len(txIDs) == 0 {
		return []models.StockLedger{}, totalDistinct, nil
	}

	// Fetch ledger rows for these transactions so we can aggregate
	if err := r.DB(ctx).Model(&models.StockLedger{}).
		Where("product_id = ?", productID).
		Where("transaction_id IN ?", txIDs).
		Order("transaction_id, created_at DESC, id DESC").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	// Aggregate by transaction_id: sum qty, keep representative latest cost and running values
	aggMap := make(map[string]models.StockLedger)
	for _, row := range items {
		a, ok := aggMap[row.TransactionID]
		if !ok {
			a = models.StockLedger{
				ID:              row.ID,
				TenantID:        row.TenantID,
				ProductID:       row.ProductID,
				TransactionID:   row.TransactionID,
				TransactionType: row.TransactionType,
				Qty:             0,
				UnitCost:        row.UnitCost,
				AverageCost:     row.AverageCost,
				StockValue:      row.StockValue,
				RunningQty:      row.RunningQty,
				CreatedAt:       row.CreatedAt,
			}
		}
		a.Qty += row.Qty
		// keep latest representative numeric fields
		if row.CreatedAt.After(a.CreatedAt) {
			a.CreatedAt = row.CreatedAt
			a.UnitCost = row.UnitCost
			a.AverageCost = row.AverageCost
			a.StockValue = row.StockValue
			a.RunningQty = row.RunningQty
		}
		aggMap[row.TransactionID] = a
	}

	// Preserve order according to txIDs
	aggregated := make([]models.StockLedger, 0, len(txIDs))
	for _, tx := range txIDs {
		if v, ok := aggMap[tx]; ok {
			aggregated = append(aggregated, v)
		}
	}

	return aggregated, totalDistinct, nil
}
