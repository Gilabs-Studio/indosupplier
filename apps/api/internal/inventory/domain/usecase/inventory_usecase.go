package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/finance/domain/accounting"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	finUC "github.com/gilabs/gims/api/internal/finance/domain/usecase"
	"github.com/gilabs/gims/api/internal/inventory/data/models"
	"github.com/gilabs/gims/api/internal/inventory/domain/dto"
	"github.com/gilabs/gims/api/internal/inventory/domain/repository"
	"gorm.io/gorm"
)

var (
	ErrBatchNotFound          = errors.New("inventory batch not found")
	ErrInsufficientBatchStock = errors.New("insufficient stock in selected batch")
	ErrInvalidInventoryInput  = errors.New("invalid inventory request")
)

type InventoryUsecase interface {
	GetStockList(ctx context.Context, req *dto.GetInventoryListRequest) (*dto.GetInventoryListResponse, error)
	GetInventoryMetrics(ctx context.Context) (*dto.InventoryMetrics, error)

	// Tree View
	GetTreeWarehouses(ctx context.Context) ([]dto.GetInventoryTreeWarehousesResponse, error)
	GetTreeProducts(ctx context.Context, req *dto.GetInventoryTreeProductsRequest) (*dto.GetInventoryTreeProductsResponse, error)
	GetTreeBatches(ctx context.Context, req *dto.GetInventoryTreeBatchesRequest) (*dto.GetInventoryTreeBatchesResponse, error)
	GetProductLedgers(ctx context.Context, productID string, req *dto.GetProductStockLedgersRequest) (*dto.GetProductStockLedgersResponse, error)

	// Stock Management
	ReserveStock(ctx context.Context, productID string, quantity float64) error
	ReleaseStock(ctx context.Context, productID string, quantity float64) error
	DeductStock(ctx context.Context, batchID string, quantity float64, params ...interface{}) error
	SelectBatches(ctx context.Context, productID string, quantity float64, strategy string) ([]dto.BatchSelectionItem, error)
	CreateStockMovement(ctx context.Context, req *dto.StockMovementRequest) (string, error)
	CreateManualStockMovement(ctx context.Context, req *dto.CreateManualMovementRequest) error

	// Integration
	ReceiveStockFromGR(ctx context.Context, req *dto.ReceiveStockRequest) error
	AdjustStockFromOpname(ctx context.Context, req *dto.AdjustStockFromOpnameRequest) error

	// Batch-level Stock Reservation
	ValidateBatchStock(ctx context.Context, batchID string, requiredQty float64) error
	ReserveBatchStock(ctx context.Context, batchID string, quantity float64) error
	ReleaseBatchStock(ctx context.Context, batchID string, quantity float64) error

	// Audit/Reconciliation/Consolidation
	TriggerMovementJournal(ctx context.Context, movementID string) error
	TriggerDocumentJournal(ctx context.Context, tx *gorm.DB, refType, refID string) error
}

type inventoryUsecase struct {
	db        *gorm.DB
	repo      repository.InventoryRepository
	journalUC finUC.JournalEntryUsecase
	engine    accounting.AccountingEngine
}

func (u *inventoryUsecase) resolveCompanyIDFromActor(ctx context.Context) (string, error) {
	actorID, _ := ctx.Value("user_id").(string)
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return "", errors.New("user not authenticated")
	}

	var companyID string
	if err := database.GetDB(ctx, u.db).
		WithContext(ctx).
		Table("employees").
		Select("company_id").
		Where("user_id = ? AND deleted_at IS NULL", actorID).
		Limit(1).
		Scan(&companyID).Error; err != nil {
		return "", err
	}
	companyID = strings.TrimSpace(companyID)
	if companyID == "" {
		return "", errors.New("employee company not found")
	}

	return companyID, nil
}

func NewInventoryUsecase(db *gorm.DB, repo repository.InventoryRepository, journalUC finUC.JournalEntryUsecase, engine accounting.AccountingEngine) InventoryUsecase {
	return &inventoryUsecase{
		db:        db,
		repo:      repo,
		journalUC: journalUC,
		engine:    engine,
	}
}

func calculateNewAverageCost(currentQty, currentAvgCost, incomingQty, incomingUnitCost float64) float64 {
	totalValue := (currentQty * currentAvgCost) + (incomingQty * incomingUnitCost)
	totalQty := currentQty + incomingQty
	if totalQty == 0 {
		return 0
	}
	return totalValue / totalQty
}

func (u *inventoryUsecase) appendStockLedger(ctx context.Context, tx *gorm.DB, productID, transactionID, transactionType string, qty, unitCost, avgCost, stockValue, runningQty float64) error {
	ledger := &models.StockLedger{
		ProductID:       productID,
		TransactionID:   strings.TrimSpace(transactionID),
		TransactionType: strings.TrimSpace(transactionType),
		Qty:             qty,
		UnitCost:        unitCost,
		AverageCost:     avgCost,
		StockValue:      stockValue,
		RunningQty:      runningQty,
	}

	return tx.WithContext(ctx).Create(ledger).Error
}

func (u *inventoryUsecase) GetStockList(ctx context.Context, req *dto.GetInventoryListRequest) (*dto.GetInventoryListResponse, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 20
	}

	items, total, err := u.repo.GetStockList(ctx, req)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))

	return &dto.GetInventoryListResponse{
		Data: items,
		Meta: dto.PaginationMeta{
			Total:      total,
			Page:       req.Page,
			PerPage:    req.PerPage,
			TotalPages: totalPages,
			HasNext:    req.Page < totalPages,
			HasPrev:    req.Page > 1,
		},
	}, nil
}

func (u *inventoryUsecase) GetTreeWarehouses(ctx context.Context) ([]dto.GetInventoryTreeWarehousesResponse, error) {
	return u.repo.GetTreeWarehouses(ctx)
}

func (u *inventoryUsecase) GetTreeProducts(ctx context.Context, req *dto.GetInventoryTreeProductsRequest) (*dto.GetInventoryTreeProductsResponse, error) {
	// Defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 20
	}

	items, total, summary, err := u.repo.GetTreeProducts(ctx, req)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))

	return &dto.GetInventoryTreeProductsResponse{
		Data: items,
		Meta: dto.PaginationMeta{
			Total:      total,
			Page:       req.Page,
			PerPage:    req.PerPage,
			TotalPages: totalPages,
			HasNext:    req.Page < totalPages,
			HasPrev:    req.Page > 1,
		},
		Summary: summary,
	}, nil
}

func (u *inventoryUsecase) GetTreeBatches(ctx context.Context, req *dto.GetInventoryTreeBatchesRequest) (*dto.GetInventoryTreeBatchesResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}

	items, total, err := u.repo.GetTreeBatches(ctx, req)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PerPage)))

	return &dto.GetInventoryTreeBatchesResponse{
		Data: items,
		Meta: dto.PaginationMeta{
			Total:      total,
			Page:       req.Page,
			PerPage:    req.PerPage,
			TotalPages: totalPages,
			HasNext:    req.Page < totalPages,
			HasPrev:    req.Page > 1,
		},
	}, nil
}

func (u *inventoryUsecase) GetInventoryMetrics(ctx context.Context) (*dto.InventoryMetrics, error) {
	return u.repo.GetInventoryMetrics(ctx)
}

func (u *inventoryUsecase) GetProductLedgers(ctx context.Context, productID string, req *dto.GetProductStockLedgersRequest) (*dto.GetProductStockLedgersResponse, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("%w: product_id is required", ErrInvalidInventoryInput)
	}

	if req == nil {
		req = &dto.GetProductStockLedgersRequest{}
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	if req.DateFrom != "" {
		parsedDateFrom, err := time.Parse("2006-01-02", req.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("%w: date_from must be YYYY-MM-DD", ErrInvalidInventoryInput)
		}
		req.ParsedDateFrom = &parsedDateFrom
	}

	if req.DateTo != "" {
		parsedDateTo, err := time.Parse("2006-01-02", req.DateTo)
		if err != nil {
			return nil, fmt.Errorf("%w: date_to must be YYYY-MM-DD", ErrInvalidInventoryInput)
		}
		dateToInclusive := parsedDateTo.Add(24*time.Hour - time.Nanosecond)
		req.ParsedDateTo = &dateToInclusive
	}

	if req.ParsedDateFrom != nil && req.ParsedDateTo != nil && req.ParsedDateFrom.After(*req.ParsedDateTo) {
		return nil, fmt.Errorf("%w: date_from cannot be greater than date_to", ErrInvalidInventoryInput)
	}

	ledgers, total, err := u.repo.GetProductLedgers(ctx, productID, req)
	if err != nil {
		return nil, err
	}

	items := make([]dto.ProductStockLedgerItem, 0, len(ledgers))
	for _, ledger := range ledgers {
		items = append(items, dto.ProductStockLedgerItem{
			ID:                   ledger.ID,
			ProductID:            ledger.ProductID,
			TransactionID:        ledger.TransactionID,
			TransactionType:      ledger.TransactionType,
			TransactionTypeLabel: stockLedgerTypeLabel(ledger.TransactionType),
			Qty:                  ledger.Qty,
			UnitCost:             ledger.UnitCost,
			AverageCost:          ledger.AverageCost,
			StockValue:           ledger.StockValue,
			RunningQty:           ledger.RunningQty,
			CreatedAt:            ledger.CreatedAt,
		})
	}

	totalPages := 0
	if req.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(req.Limit)))
	}

	return &dto.GetProductStockLedgersResponse{
		Data: items,
		Meta: dto.PaginationMeta{
			Total:      total,
			Page:       req.Page,
			PerPage:    req.Limit,
			TotalPages: totalPages,
			HasNext:    req.Page < totalPages,
			HasPrev:    req.Page > 1,
		},
	}, nil
}

func stockLedgerTypeLabel(transactionType string) string {
	switch strings.ToUpper(strings.TrimSpace(transactionType)) {
	case "GR":
		return "Goods Receipt"
	case "GI":
		return "Goods Issue"
	case "OPNAME":
		return "Stock Opname"
	case "TRANSFER":
		return "Transfer"
	default:
		if strings.TrimSpace(transactionType) == "" {
			return "Unknown"
		}
		return transactionType
	}
}

func (u *inventoryUsecase) ReserveStock(ctx context.Context, productID string, quantity float64) error {
	return u.repo.UpdateProductReservedStock(ctx, productID, quantity)
}

func (u *inventoryUsecase) ReleaseStock(ctx context.Context, productID string, quantity float64) error {
	// Release is essentially negative reservation (reducing reserved count)
	// But we ensure quantity is positive for method clarity, repo handles sign
	return u.repo.UpdateProductReservedStock(ctx, productID, -quantity)
}

func (u *inventoryUsecase) DeductStock(ctx context.Context, batchID string, quantity float64, params ...interface{}) error {
	if tx := database.GetTx(ctx); tx != nil {
		return u.deductStockWithTx(ctx, tx, batchID, quantity, params...)
	}

	return database.RetryTx(u.db.WithContext(ctx), func(tx *gorm.DB) error {
		return u.deductStockWithTx(database.WithTx(ctx, tx), tx, batchID, quantity, params...)
	})
}

func (u *inventoryUsecase) deductStockWithTx(ctx context.Context, tx *gorm.DB, batchID string, quantity float64, params ...interface{}) error {
	// First fetch batch to get product ID for aggregate update
	batch, err := u.repo.GetBatchByID(ctx, batchID)
	if err != nil {
		return err
	}
	if batch == nil {
		return ErrBatchNotFound
	}

	currentHpp, currentStock, err := u.repo.GetProductCostInfo(ctx, batch.ProductID)
	if err != nil {
		return err
	}

	runningQty := currentStock - quantity
	if runningQty < 0 {
		return ErrInsufficientBatchStock
	}

	// Allow custom transaction type (e.g., "SO" for sales orders), default to "GI"
	ledgerType := "GI"
	transactionID := batchID // Default to batch ID
	if len(params) > 0 {
		if v, ok := params[0].(string); ok && v != "" {
			ledgerType = v
		}
	}
	if len(params) > 1 {
		if v, ok := params[1].(string); ok && v != "" {
			transactionID = v
		}
	}

	stockValue := runningQty * currentHpp
	if err := u.appendStockLedger(ctx, tx, batch.ProductID, transactionID, ledgerType, -quantity, currentHpp, currentHpp, stockValue, runningQty); err != nil {
		return err
	}

	// 1. Update Batch Quantity
	if err := u.repo.UpdateBatchQuantity(ctx, batchID, -quantity); err != nil {
		return err
	}

	// 2. Update Aggregate Product Stock and keep current average cost
	if err := u.repo.UpdateProductStock(ctx, batch.ProductID, -quantity); err != nil {
		return err
	}

	return u.repo.UpdateProductAverageCost(ctx, batch.ProductID, currentHpp)
}

func (u *inventoryUsecase) SelectBatches(ctx context.Context, productID string, quantity float64, strategy string) ([]dto.BatchSelectionItem, error) {
	batches, err := u.repo.GetBatchesByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Map to selection items
	var selectionItems []dto.BatchSelectionItem
	for _, b := range batches {
		selectionItems = append(selectionItems, dto.BatchSelectionItem{
			ID:          b.ID,
			BatchNumber: b.BatchNumber,
			Quantity:    b.CurrentQuantity, // Now float64 matching struct
			ExpiredAt:   *b.ExpiryDate,
			ReceivedAt:  *b.ReceivedAt,
		})
	}

	// Sort based on strategy (Simple bubble sort for now or defer to repo/sql ordering)
	// Here we just implement basic logic placeholders, assuming repo returns sorted or we sort here
	if strategy == "FEFO" {
		// Sort by ExpiredAt
		for i := 0; i < len(selectionItems)-1; i++ {
			for j := 0; j < len(selectionItems)-i-1; j++ {
				if selectionItems[j].ExpiredAt.After(selectionItems[j+1].ExpiredAt) {
					selectionItems[j], selectionItems[j+1] = selectionItems[j+1], selectionItems[j]
				}
			}
		}
	} else {
		// FIFO (Default) - Sort by ReceivedAt
		for i := 0; i < len(selectionItems)-1; i++ {
			for j := 0; j < len(selectionItems)-i-1; j++ {
				if selectionItems[j].ReceivedAt.After(selectionItems[j+1].ReceivedAt) {
					selectionItems[j], selectionItems[j+1] = selectionItems[j+1], selectionItems[j]
				}
			}
		}
	}

	// allocate logic can be here or in FE, but usually FE selects from list
	// UseCase just returns sorted available batches
	return selectionItems, nil
}

func (u *inventoryUsecase) CreateStockMovement(ctx context.Context, req *dto.StockMovementRequest) (string, error) {
	var movementID string
	if tx := database.GetTx(ctx); tx != nil {
		id, err := u.repo.CreateStockMovement(ctx, req)
		if err != nil {
			return "", err
		}
		movementID = id

		if req.SkipJournaling {
			log.Printf("[Inventory] Skipping individual journal for movement %s (Document level trigger expected)", id)
			return movementID, nil
		}

		if err := u.triggerInventoryJournal(ctx, tx, movementID); err != nil {
			return "", fmt.Errorf("triggered journal failed: %w", err)
		}

		return movementID, nil
	}

	// Use RetryTx for atomic operation. Failure in journal = Rollback stock movement.
	err := database.RetryTx(u.db.WithContext(ctx), func(tx *gorm.DB) error {
		txCtx := database.WithTx(ctx, tx)
		id, err := u.repo.CreateStockMovement(txCtx, req)
		if err != nil {
			return err
		}
		movementID = id

		// Auto-trigger journal for financial impact, unless skipped
		if req.SkipJournaling {
			log.Printf("[Inventory] Skipping individual journal for movement %s (Document level trigger expected)", id)
			return nil
		}

		if err := u.triggerInventoryJournal(txCtx, tx, movementID); err != nil {
			return fmt.Errorf("triggered journal failed: %w", err)
		}
		return nil
	})

	if err != nil {
		log.Printf("[Inventory] FAILED CreateStockMovement for %s %s: %v", req.ReferenceType, req.ReferenceNumber, err)
		return "", err
	}

	return movementID, nil
}

func (u *inventoryUsecase) ReceiveStockFromGR(ctx context.Context, req *dto.ReceiveStockRequest) error {
	return database.RetryTx(u.db.WithContext(ctx), func(tx *gorm.DB) error {
		ctx = database.WithTx(ctx, tx)

		for idx, item := range req.Items {
			// 1. Calculate New Average Cost (Weighted Average)
			currentHpp, currentStock, err := u.repo.GetProductCostInfo(ctx, item.ProductID)
			if err != nil {
				return err
			}

			runningQty := currentStock + item.Quantity
			if runningQty < 0 {
				return fmt.Errorf("invalid stock movement for product %s: running quantity would be negative", item.ProductID)
			}

			newHpp := calculateNewAverageCost(currentStock, currentHpp, item.Quantity, item.CostPrice)
			if runningQty == 0 && item.Quantity > 0 {
				newHpp = item.CostPrice
			}

			stockValue := runningQty * newHpp

			// 2. Persist immutable stock ledger row before mutating product aggregates.
			if err := u.appendStockLedger(ctx, tx, item.ProductID, req.SourceID, "GR", item.Quantity, item.CostPrice, newHpp, stockValue, runningQty); err != nil {
				return err
			}

			// 3. Update Product Cost and Aggregate Stock
			if err := u.repo.UpdateProductAverageCost(ctx, item.ProductID, newHpp); err != nil {
				return err
			}
			if err := u.repo.UpdateProductStock(ctx, item.ProductID, item.Quantity); err != nil {
				return err
			}

			// 4. Create Batch
			batchToken := sanitizeBatchToken(req.SourceNumber)
			if batchToken == "" {
				batchToken = sanitizeBatchToken(req.SourceID)
			}
			if batchToken == "" {
				batchToken = apptime.Now().Format("20060102150405")
			}

			batchNumber := fmt.Sprintf("GR-%s-%03d", batchToken, idx+1)
			if len(batchNumber) > 100 {
				batchNumber = batchNumber[:100]
			}
			if item.BatchNumber != nil && *item.BatchNumber != "" {
				batchNumber = *item.BatchNumber
			}

			batchParams := &dto.CreateBatchParams{
				ProductID:       item.ProductID,
				WarehouseID:     req.WarehouseID,
				BatchNumber:     batchNumber,
				ExpiryDate:      item.ExpiryDate,
				InitialQuantity: item.Quantity,
				CostPrice:       item.CostPrice,
				ReceivedAt:      req.ReceivedAt,
			}

			batchID, err := u.repo.CreateBatch(ctx, batchParams)
			if err != nil {
				return err
			}

			// 5. Create Stock Movement (IN)
			createdBy := req.ReceivedBy
			movementReq := &dto.StockMovementRequest{
				InventoryBatchID: batchID,
				ProductID:        item.ProductID,
				WarehouseID:      req.WarehouseID,
				Type:             "IN",
				Quantity:         item.Quantity,
				ReferenceType:    reference.RefTypeGoodsReceipt,
				ReferenceID:      req.SourceID,
				ReferenceNumber:  req.SourceNumber,
				Source:           req.Source,
				Cost:             item.CostPrice,
				Description:      req.Notes,
				CreatedBy:        &createdBy,
			}

			// Note: For GR, we usually trigger journal at GoodsReceipt closing
			// but individual item movement traceability still needs updating
			if _, err := u.repo.CreateStockMovement(ctx, movementReq); err != nil {
				return err
			}

			// 6. Movement is created, but we NO LONGER trigger journal here per item.
			// The caller (e.g. GoodsReceiptUsecase.Approve) is responsible for calling
			// TriggerDocumentJournal at the end of the transaction for consolidation.
		}
		return nil
	})
}

func sanitizeBatchToken(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		" ", "-",
		":", "-",
		".", "-",
	)

	normalized := strings.ToUpper(replacer.Replace(trimmed))
	normalized = strings.Trim(normalized, "-")
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}

	return normalized
}

func (u *inventoryUsecase) ValidateBatchStock(ctx context.Context, batchID string, requiredQty float64) error {
	batch, err := u.repo.GetBatchByID(ctx, batchID)
	if err != nil {
		return err
	}
	if batch == nil {
		return ErrBatchNotFound
	}
	if batch.Available < requiredQty {
		return ErrInsufficientBatchStock
	}
	return nil
}

func (u *inventoryUsecase) ReserveBatchStock(ctx context.Context, batchID string, quantity float64) error {
	return u.repo.UpdateBatchReservedQuantity(ctx, batchID, quantity)
}

func (u *inventoryUsecase) ReleaseBatchStock(ctx context.Context, batchID string, quantity float64) error {
	return u.repo.UpdateBatchReservedQuantity(ctx, batchID, -quantity)
}

// AdjustStockFromOpname creates ADJUST stock movements and updates batch/product quantities
// based on variance data from a posted Stock Opname.
// Positive variance = surplus (qty found more than system) → IN adjustment
// Negative variance = shortage (qty found less than system) → OUT adjustment
func (u *inventoryUsecase) AdjustStockFromOpname(ctx context.Context, req *dto.AdjustStockFromOpnameRequest) error {
	return database.RetryTx(u.db.WithContext(ctx), func(tx *gorm.DB) error {
		ctx = database.WithTx(ctx, tx)

		for _, item := range req.Items {
			if item.VarianceQty == 0 {
				continue // No adjustment needed for matching items
			}

			currentHpp, currentStock, err := u.repo.GetProductCostInfo(ctx, item.ProductID)
			if err != nil {
				return err
			}

			runningQty := currentStock + item.VarianceQty
			if runningQty < 0 {
				return fmt.Errorf("invalid stock opname adjustment for product %s: running quantity would be negative", item.ProductID)
			}

			avgCost := currentHpp
			stockValue := runningQty * avgCost
			if err := u.appendStockLedger(ctx, tx, item.ProductID, req.OpnameID, "OPNAME", item.VarianceQty, avgCost, avgCost, stockValue, runningQty); err != nil {
				return err
			}

			// Determine batch ID — use the specific batch from opname if provided
			batchID := ""
			if item.BatchID != nil && *item.BatchID != "" {
				batchID = *item.BatchID
			} else {
				batches, err := u.repo.GetBatchesByProductAndWarehouse(ctx, item.ProductID, req.WarehouseID)
				if err != nil {
					return err
				}
				if len(batches) > 0 {
					batchID = batches[0].ID
				}
			}

			// Create the ADJUST stock movement
			// Pass signed variance so the repo can determine QtyIn vs QtyOut
			movementReq := &dto.StockMovementRequest{
				InventoryBatchID: batchID,
				ProductID:        item.ProductID,
				WarehouseID:      req.WarehouseID,
				Type:             "ADJUST",
				Quantity:         item.VarianceQty,
				ReferenceType:    "STOCK_OPNAME",
				ReferenceID:      req.OpnameID,
				ReferenceNumber:  req.OpnameNumber,
				Description:      req.Notes,
				CreatedBy:        &req.PostedBy,
			}

			movementID, err := u.repo.CreateStockMovement(ctx, movementReq)
			if err != nil {
				return err
			}

			// Update batch quantity with the variance delta
			if batchID != "" {
				if err := u.repo.UpdateBatchQuantity(ctx, batchID, item.VarianceQty); err != nil {
					return err
				}
			}

			// Update aggregate product stock
			if err := u.repo.UpdateProductStock(ctx, item.ProductID, item.VarianceQty); err != nil {
				return err
			}
			if err := u.repo.UpdateProductAverageCost(ctx, item.ProductID, avgCost); err != nil {
				return err
			}

			// TRIGGER JOURNAL for the adjustment
			if err := u.triggerInventoryJournal(ctx, tx, movementID); err != nil {
				return fmt.Errorf("opname journal: %w", err)
			}
		}
		return nil
	})
}

var ErrTargetWarehouseRequired = errors.New("target warehouse is required for TRANSFER")
var ErrInsufficientStock = errors.New("insufficient stock for movement")

func (u *inventoryUsecase) CreateManualStockMovement(ctx context.Context, req *dto.CreateManualMovementRequest) error {
	return database.RetryTx(u.db.WithContext(ctx), func(tx *gorm.DB) error {
		ctx = database.WithTx(ctx, tx)

		zeroUUID := "00000000-0000-0000-0000-000000000000"

		if req.ReferenceNumber == "" {
			req.ReferenceNumber = "MANUAL-" + time.Now().Format("20060102-150405")
		}

		deductStock := func(warehouseID string, qty float64, movementType string) (string, error) {
			currentHpp, currentStock, err := u.repo.GetProductCostInfo(ctx, req.ProductID)
			if err != nil {
				return "", err
			}

			// Pre-check: ensure available stock covers the requested quantity (#283)
			if currentStock < qty {
				return "", ErrInsufficientStock
			}

			// Determine ledger transaction type based on movement type
			ledgerType := "GI"
			if strings.ToUpper(movementType) == "TRANSFER" {
				ledgerType = "TRANSFER"
			}

			// Determine reference type for stock movement record
			refType := "INVENTORY_ADJUSTMENT"
			if strings.ToUpper(movementType) == "TRANSFER" {
				refType = "TRANSFER"
			}

			// Build a unified list of (id, available) pairs to deduct from
			type batchDeductItem struct {
				id        string
				available float64
			}
			var batchesToDeduct []batchDeductItem

			// If a specific batch is requested, validate and use only that batch
			if req.BatchID != nil && *req.BatchID != "" {
				specificBatch, err := u.repo.GetBatchByID(ctx, *req.BatchID)
				if err != nil {
					return "", fmt.Errorf("failed to fetch specified batch: %w", err)
				}
				if specificBatch == nil {
					return "", ErrBatchNotFound
				}
				if specificBatch.Available < qty {
					return "", ErrInsufficientBatchStock
				}
				batchesToDeduct = []batchDeductItem{{id: specificBatch.ID, available: specificBatch.Available}}
			} else {
				allBatches, err := u.repo.GetBatchesByProductAndWarehouse(ctx, req.ProductID, warehouseID)
				if err != nil {
					return "", err
				}
				for _, b := range allBatches {
					batchesToDeduct = append(batchesToDeduct, batchDeductItem{id: b.ID, available: b.Available})
				}
			}

			remaining := qty
			var lastMovementID string
			for _, batch := range batchesToDeduct {
				if remaining <= 0 {
					break
				}

				if batch.available <= 0 {
					continue
				}

				toDeduct := math.Min(batch.available, remaining)

				movReq := &dto.StockMovementRequest{
					InventoryBatchID:  batch.id,
					ProductID:         req.ProductID,
					WarehouseID:       warehouseID,
					Type:              movementType,
					Quantity:          toDeduct,
					ReferenceType:     refType,
					ReferenceID:       zeroUUID,
					ReferenceNumber:   req.ReferenceNumber,
					Description:       req.Description,
					CreatedBy:         &req.CreatedBy,
					MovementDirection: "OUT",
				}
				movID, err := u.repo.CreateStockMovement(ctx, movReq)
				if err != nil {
					return "", err
				}
				lastMovementID = movID

				if err := u.repo.UpdateBatchQuantity(ctx, batch.id, -toDeduct); err != nil {
					return "", err
				}

				remaining -= toDeduct
			}

			if remaining > 0 {
				return "", ErrInsufficientStock
			}

			runningQty := currentStock - qty
			if runningQty < 0 {
				return "", ErrInsufficientStock
			}
			stockValue := runningQty * currentHpp
			if err := u.appendStockLedger(ctx, tx, req.ProductID, req.ReferenceNumber, ledgerType, -qty, currentHpp, currentHpp, stockValue, runningQty); err != nil {
				return "", err
			}

			if err := u.repo.UpdateProductStock(ctx, req.ProductID, -qty); err != nil {
				return "", err
			}
			if err := u.repo.UpdateProductAverageCost(ctx, req.ProductID, currentHpp); err != nil {
				return "", err
			}

			return lastMovementID, nil
		}

		addStock := func(warehouseID string, qty float64, movementType string) (string, error) {
			currentHpp, currentStock, err := u.repo.GetProductCostInfo(ctx, req.ProductID)
			if err != nil {
				return "", err
			}

			// Determine ledger transaction type based on movement type
			ledgerType := "GR"
			if strings.ToUpper(movementType) == "TRANSFER" {
				ledgerType = "TRANSFER"
			}

			// Determine reference type for stock movement record
			refType := "INVENTORY_ADJUSTMENT"
			if strings.ToUpper(movementType) == "TRANSFER" {
				refType = "TRANSFER"
			}

			newAvg := calculateNewAverageCost(currentStock, currentHpp, qty, currentHpp)
			runningQty := currentStock + qty
			stockValue := runningQty * newAvg
			if err := u.appendStockLedger(ctx, tx, req.ProductID, req.ReferenceNumber, ledgerType, qty, currentHpp, newAvg, stockValue, runningQty); err != nil {
				return "", err
			}

			now := apptime.Now()
			batchNumber := "MB-" + now.Format("060102150405")

			batchParams := &dto.CreateBatchParams{
				ProductID:       req.ProductID,
				WarehouseID:     warehouseID,
				BatchNumber:     batchNumber,
				InitialQuantity: qty,
				CostPrice:       currentHpp,
				ReceivedAt:      now,
			}

			batchID, err := u.repo.CreateBatch(ctx, batchParams)
			if err != nil {
				return "", err
			}

			movReq := &dto.StockMovementRequest{
				InventoryBatchID:  batchID,
				ProductID:         req.ProductID,
				WarehouseID:       warehouseID,
				Type:              movementType,
				Quantity:          qty,
				ReferenceType:     refType,
				ReferenceID:       zeroUUID,
				ReferenceNumber:   req.ReferenceNumber,
				Description:       req.Description,
				CreatedBy:         &req.CreatedBy,
				MovementDirection: "IN",
			}
			movID, err := u.repo.CreateStockMovement(ctx, movReq)
			if err != nil {
				return "", err
			}

			if err := u.repo.UpdateProductStock(ctx, req.ProductID, qty); err != nil {
				return "", err
			}
			if err := u.repo.UpdateProductAverageCost(ctx, req.ProductID, newAvg); err != nil {
				return "", err
			}

			return movID, nil
		}

		var movementID string
		var err error

		switch req.Type {
		case "IN":
			movementID, err = addStock(req.WarehouseID, req.Quantity, "IN")
		case "OUT":
			movementID, err = deductStock(req.WarehouseID, req.Quantity, "OUT")
		case "ADJUST":
			return errors.New("please use stock opname for adjustments")
		case "TRANSFER":
			if req.TargetWarehouseID == nil || *req.TargetWarehouseID == "" {
				return ErrTargetWarehouseRequired
			}
			if _, err := deductStock(req.WarehouseID, req.Quantity, "TRANSFER"); err != nil {
				return err
			}
			if _, err := addStock(*req.TargetWarehouseID, req.Quantity, "TRANSFER"); err != nil {
				return err
			}
			return nil // Skip journal for internal transfers
		}

		if err != nil {
			return err
		}

		// TRIGGER JOURNAL for manual adjustments
		if movementID != "" {
			if err := u.triggerInventoryJournal(ctx, tx, movementID); err != nil {
				return fmt.Errorf("manual adjustment journal: %w", err)
			}
		}

		return nil
	})
}

// triggerInventoryJournal handles the financial integration for a single stock movement.
// It resolves the correct posting profile (Gain/Loss), generates the journal via
// AccountingEngine, and posts it via JournalEntryUsecase.
func (u *inventoryUsecase) triggerInventoryJournal(ctx context.Context, tx *gorm.DB, movementID string) error {
	if u.journalUC == nil || u.engine == nil {
		return fmt.Errorf("journaling engine not configured: physical movement (ID: %s) cancelled", movementID)
	}

	companyID, err := u.resolveCompanyIDFromActor(ctx)
	if err != nil {
		return err
	}

	// 1. Fetch the movement with product info
	var movement models.StockMovement
	if err := tx.Preload("Product").First(&movement, "id = ?", movementID).Error; err != nil {
		return fmt.Errorf("failed to fetch movement for journal: %w", err)
	}

	// 2. Determine volume (positive = IN, negative = OUT)
	volume := movement.QtyIn
	if movement.QtyOut > 0 {
		volume = -movement.QtyOut
	}

	if volume == 0 {
		return nil // No value change
	}

	// 4. Resolve Posting Profile based on reference type and volume
	var profile accounting.PostingProfile
	refTypeCanonical := reference.Normalize(movement.RefType)

	// Skip journaling for internal transfers
	if refTypeCanonical == "TRANSFER" {
		return nil
	}

	switch refTypeCanonical {
	case reference.RefTypeGoodsReceipt:
		profile = accounting.ProfileGoodsReceipt
	case reference.RefTypeDeliveryOrder:
		profile = accounting.ProfileCOGS
	default:
		if volume > 0 {
			profile = accounting.ProfileInventoryGain
		} else {
			profile = accounting.ProfileInventoryLoss
		}
	}

	// 3. Resolve cost-per-unit (use HPP if not set on movement)
	cost := movement.Cost
	if cost == 0 && movement.Product != nil {
		cost = movement.Product.CurrentHpp
	}

	// 4. Prepare Data for Accounting Engine
	data := accounting.TransactionData{
		ReferenceType:   "INVENTORY_MOVEMENT", // Link journal directly to the movement ID for 1:1 traceability
		ReferenceID:     movement.ID,
		EntryDate:       movement.Date.Format("2006-01-02"),
		Description:     fmt.Sprintf("%s [%s]: %s", profile.DescriptionTemplate, movement.RefNumber, movement.Source),
		TotalAmount:     math.Abs(volume * cost),
		DescriptionArgs: []interface{}{movement.RefNumber},
	}

	// Skip journaling when cost is zero — no financial impact
	if data.TotalAmount == 0 {
		return nil
	}

	// 5. Generate Journal Request
	genReq, err := u.engine.GenerateJournal(ctx, profile, data)
	if err != nil {
		return fmt.Errorf("accounting engine: %w", err)
	}
	genReq.CompanyID = companyID

	// 6. Post Journal (Idempotent via reference_type + reference_id)
	// Passing tx via context to ensure the usecase uses the same transaction.
	journal, err := u.journalUC.PostOrUpdateJournal(database.WithTx(ctx, tx), genReq)
	if err != nil {
		return fmt.Errorf("journal post: %w", err)
	}

	// 7. Store reference back to movement
	if err := tx.Model(&movement).Update("journal_entry_id", journal.ID).Error; err != nil {
		return fmt.Errorf("traceability link: %w", err)
	}

	return nil
}
func (u *inventoryUsecase) TriggerMovementJournal(ctx context.Context, movementID string) error {
	return u.triggerInventoryJournal(ctx, u.db, movementID)
}

func (u *inventoryUsecase) TriggerDocumentJournal(ctx context.Context, tx *gorm.DB, refType, refID string) error {
	if tx == nil {
		tx = database.GetDB(ctx, u.db)
	}

	companyID, err := u.resolveCompanyIDFromActor(ctx)
	if err != nil {
		return err
	}

	// 1. Fetch all movements for this document
	refType = reference.Normalize(refType)
	var movements []models.StockMovement
	if err := tx.WithContext(ctx).
		Preload("Product").
		Where("ref_type = ? AND ref_id = ?", refType, refID).
		Find(&movements).Error; err != nil {
		return fmt.Errorf("fetch movements: %w", err)
	}

	if len(movements) == 0 {
		log.Printf("[Inventory] TriggerDocumentJournal: No movements found for %s %s. Skipping journal.", refType, refID)
		return nil
	}

	log.Printf("[Inventory] TriggerDocumentJournal: Processing %d movements for %s %s", len(movements), refType, refID)

	// 2. Group by product and sum up volume*cost
	// Actually, for a single journal header, we just need the total value if they use the same accounts.
	// But standard ERP journals might have separate lines per product/batch.
	// Let's create one consolidated journal per document with 1 Credit (Inventory) and 1 Debit (COGS/Accrual) summary
	// or per-item if detail is needed.
	// For now, let's keep it simple: 1 consolidated journal with 2 lines (Total Debit, Total Credit).

	var totalValue float64
	var refNumber string
	var source string
	var date time.Time

	for _, m := range movements {
		cost := m.Cost
		if cost == 0 && m.Product != nil {
			cost = m.Product.CurrentHpp
		}

		volume := m.QtyIn - m.QtyOut
		totalValue += math.Abs(volume * cost)

		if refNumber == "" {
			refNumber = m.RefNumber
		}
		if source == "" {
			source = m.Source
		}
		if date.IsZero() {
			date = m.Date
		}
	}

	if totalValue <= 0.001 {
		return nil
	}

	// 3. Resolve Posting Profile
	var profile accounting.PostingProfile
	refTypeCanonical := reference.Normalize(refType)

	switch refTypeCanonical {
	case reference.RefTypeGoodsReceipt:
		profile = accounting.ProfileGoodsReceipt
	case reference.RefTypeDeliveryOrder:
		profile = accounting.ProfileCOGS
	default:
		// Default to gain/loss based on net movement (In vs Out)
		// But usually documents like DO/GR are unidirectional.
		if movements[0].QtyIn > movements[0].QtyOut {
			profile = accounting.ProfileInventoryGain
		} else {
			profile = accounting.ProfileInventoryLoss
		}
	}

	// 4. Prepare Data for Accounting Engine
	data := accounting.TransactionData{
		ReferenceType:   refTypeCanonical,
		ReferenceID:     refID,
		EntryDate:       date.Format("2006-01-02"),
		Description:     fmt.Sprintf("%s [%s]: %s", profile.DescriptionTemplate, refNumber, source),
		TotalAmount:     totalValue,
		DescriptionArgs: []interface{}{refNumber},
	}

	// 5. Generate Journal Request
	genReq, err := u.engine.GenerateJournal(ctx, profile, data)
	if err != nil {
		return fmt.Errorf("accounting engine: %w", err)
	}
	genReq.CompanyID = companyID

	// 6. Post Journal
	journal, err := u.journalUC.PostOrUpdateJournal(database.WithTx(ctx, tx), genReq)
	if err != nil {
		return fmt.Errorf("journal post: %w", err)
	}

	// 7. Store reference back to ALL movements AND the source document
	err = tx.Transaction(func(tx *gorm.DB) error {
		// Update movements
		if err := tx.Model(&models.StockMovement{}).
			Where("ref_type = ? AND ref_id = ?", refType, refID).
			Update("journal_entry_id", journal.ID).Error; err != nil {
			return err
		}

		// Update source document (DO/GR)
		table := ""
		switch refTypeCanonical {
		case reference.RefTypeDeliveryOrder:
			table = "delivery_orders"
		case reference.RefTypeGoodsReceipt:
			table = "goods_receipts"
		}

		if table != "" {
			if err := tx.Table(table).Where("id = ?", refID).Update("journal_entry_id", journal.ID).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("traceability link: %w", err)
	}

	return nil
}
