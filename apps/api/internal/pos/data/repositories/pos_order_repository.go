package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/infrastructure/security"

	"github.com/gilabs/gims/api/internal/pos/data/models"
	"gorm.io/gorm"
)

// PosOrderListParams defines filter/pagination options for listing POS orders (legacy)
type PosOrderListParams struct {
	SessionID string
	OutletID  string
	Status    string
	Page      int
	PerPage   int
}

// POSOrderListParams defines handler-level offset-based filter params for POS orders
type POSOrderListParams struct {
	SessionID string
	OutletID  string
	Status    string
	Limit     int
	Offset    int
}

// PosOrderRepository defines data access for POS orders
type PosOrderRepository interface {
	Create(ctx context.Context, order *models.PosOrder) error
	GetByID(ctx context.Context, id string) (*models.PosOrder, error)
	FindActiveByOutletAndTable(ctx context.Context, outletID, tableID string) (*models.PosOrder, error)
	FindActiveByOutletAndTableLabel(ctx context.Context, outletID, tableLabel string) (*models.PosOrder, error)
	Update(ctx context.Context, order *models.PosOrder) error
	List(ctx context.Context, params PosOrderListParams) ([]models.PosOrder, int64, error)
	ListByParams(ctx context.Context, params POSOrderListParams) ([]models.PosOrder, int64, error)
	GetNextOrderNumber(ctx context.Context, prefix string) (string, error)
	AddItem(ctx context.Context, item *models.PosOrderItem) error
	UpdateItem(ctx context.Context, item *models.PosOrderItem) error
	DeleteItem(ctx context.Context, itemID string) error
	GetItems(ctx context.Context, orderID string) ([]models.PosOrderItem, error)
}

type posOrderRepository struct {
	db *gorm.DB
}

func NewPosOrderRepository(db *gorm.DB) PosOrderRepository {
	return &posOrderRepository{db: db}
}

func (r *posOrderRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *posOrderRepository) Create(ctx context.Context, order *models.PosOrder) error {
	return r.getDB(ctx).Create(order).Error
}

func (r *posOrderRepository) GetByID(ctx context.Context, id string) (*models.PosOrder, error) {
	var order models.PosOrder
	err := r.getDB(ctx).
		Preload("Items").
		First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *posOrderRepository) FindActiveByOutletAndTable(ctx context.Context, outletID, tableID string) (*models.PosOrder, error) {
	var order models.PosOrder
	err := r.getDB(ctx).
		Preload("Items").
		Where("outlet_id = ? AND table_id = ? AND status IN ?", outletID, tableID, []models.PosOrderStatus{
			models.PosOrderStatusDraft,
			models.PosOrderStatusInProgress,
			models.PosOrderStatusReady,
		}).
		Order("updated_at DESC").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *posOrderRepository) FindActiveByOutletAndTableLabel(ctx context.Context, outletID, tableLabel string) (*models.PosOrder, error) {
	var order models.PosOrder
	normalizedLabel := tableLabel
	err := r.getDB(ctx).
		Preload("Items").
		Where("outlet_id = ? AND lower(trim(table_label)) = lower(trim(?)) AND status IN ?", outletID, normalizedLabel, []models.PosOrderStatus{
			models.PosOrderStatusDraft,
			models.PosOrderStatusInProgress,
			models.PosOrderStatusReady,
		}).
		Order("updated_at DESC").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *posOrderRepository) Update(ctx context.Context, order *models.PosOrder) error {
	return r.getDB(ctx).Save(order).Error
}

func (r *posOrderRepository) List(ctx context.Context, params PosOrderListParams) ([]models.PosOrder, int64, error) {
	var orders []models.PosOrder
	var total int64

	query := r.getDB(ctx).Model(&models.PosOrder{})

	// Apply dynamic scope filter
	query = security.ApplyScopeFilter(query, ctx, security.ScopeQueryOptions{
		OutletIDColumn: "outlet_id",
	})

	if params.SessionID != "" {
		query = query.Where("session_id = ?", params.SessionID)
	}
	if params.OutletID != "" {
		query = query.Where("outlet_id = ?", params.OutletID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	perPage := params.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	if err := query.Preload("Items").Order("created_at DESC").Offset(offset).Limit(perPage).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *posOrderRepository) GetNextOrderNumber(ctx context.Context, prefix string) (string, error) {
	var sequence int
	now := time.Now()
	datePrefix := fmt.Sprintf("%s-%s-", prefix, now.Format("20060102"))

	var last models.PosOrder
	err := r.getDB(ctx).
		Where("order_number LIKE ?", datePrefix+"%").
		Order("order_number DESC").
		First(&last).Error

	if err != nil {
		sequence = 1
	} else {
		_, _ = fmt.Sscanf(last.OrderNumber[len(datePrefix):], "%d", &sequence)
		sequence++
	}

	return fmt.Sprintf("%s%04d", datePrefix, sequence), nil
}

func (r *posOrderRepository) AddItem(ctx context.Context, item *models.PosOrderItem) error {
	return r.getDB(ctx).Create(item).Error
}

func (r *posOrderRepository) UpdateItem(ctx context.Context, item *models.PosOrderItem) error {
	return r.getDB(ctx).Save(item).Error
}

func (r *posOrderRepository) DeleteItem(ctx context.Context, itemID string) error {
	return r.getDB(ctx).Delete(&models.PosOrderItem{}, "id = ?", itemID).Error
}

func (r *posOrderRepository) GetItems(ctx context.Context, orderID string) ([]models.PosOrderItem, error) {
	var items []models.PosOrderItem
	err := r.getDB(ctx).Where("pos_order_id = ?", orderID).Find(&items).Error
	return items, err
}

// ListByParams uses offset-based pagination for handler-level queries
func (r *posOrderRepository) ListByParams(ctx context.Context, params POSOrderListParams) ([]models.PosOrder, int64, error) {
	var orders []models.PosOrder
	var total int64

	query := r.getDB(ctx).Model(&models.PosOrder{})

	// Apply dynamic scope filter
	query = security.ApplyScopeFilter(query, ctx, security.ScopeQueryOptions{
		OutletIDColumn: "outlet_id",
	})

	if params.SessionID != "" {
		query = query.Where("session_id = ?", params.SessionID)
	}
	if params.OutletID != "" {
		query = query.Where("outlet_id = ?", params.OutletID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := params.Limit
	if limit < 1 {
		limit = 20
	}

	if err := query.Preload("Items").Order("created_at DESC").Offset(params.Offset).Limit(limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
