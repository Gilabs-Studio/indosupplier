package usecase

import (
	"context"
	"fmt"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/notification/data/models"
	"github.com/gilabs/gims/api/internal/notification/data/repositories"
	"github.com/gilabs/gims/api/internal/notification/domain/dto"
)

type entityLinkRule struct {
	basePath string
	queryKey string
}

var entityLinkRules = map[string]entityLinkRule{
	"company":              {basePath: "/master-data/company", queryKey: "open_company"},
	"employee":             {basePath: "/master-data/employees", queryKey: "open_employee"},
	"supplier":             {basePath: "/master-data/suppliers", queryKey: "open_supplier"},
	"customer":             {basePath: "/master-data/customers", queryKey: "open_customer"},
	"product":              {basePath: "/master-data/products", queryKey: "open_product"},
	"sales_quotation":      {basePath: "/sales/quotations", queryKey: "open_quotation"},
	"sales_order":          {basePath: "/sales/orders", queryKey: "open_order"},
	"delivery_order":       {basePath: "/sales/delivery-orders", queryKey: "open_delivery_order"},
	"customer_invoice":     {basePath: "/sales/invoices", queryKey: "open_customer_invoice"},
	"customer_invoice_dp":  {basePath: "/sales/customer-invoice-down-payments", queryKey: "open_customer_invoice_dp"},
	"purchase_requisition": {basePath: "/purchase/purchase-requisitions", queryKey: "open_purchase_requisition"},
	"purchase_order":       {basePath: "/purchase/purchase-orders", queryKey: "open_purchase_order"},
	"goods_receipt":        {basePath: "/purchase/goods-receipt", queryKey: "open_goods_receipt"},
	"supplier_invoice":     {basePath: "/purchase/supplier-invoices", queryKey: "open_supplier_invoice"},
	"supplier_invoice_dp":  {basePath: "/purchase/supplier-invoice-down-payments", queryKey: "open_supplier_invoice_dp"},
	"stock_opname":         {basePath: "/stock/opname", queryKey: "open_stock_opname"},
	"payment":              {basePath: "/finance/payments", queryKey: "open_payment"},
	"non_trade_payable":    {basePath: "/finance/non-trade-payables", queryKey: "open_non_trade_payable"},
	"budget":               {basePath: "/finance/budget", queryKey: "open_budget"},
	"financial_closing":    {basePath: "/finance/closing", queryKey: "open_financial_closing"},
	"asset_maintenance":    {basePath: "/finance/asset-maintenance", queryKey: "open_asset_maintenance"},
	"travel_plan":          {basePath: "/travel/travel-planner", queryKey: "open_trip"},
	"salary":               {basePath: "/hrd/salary-structures", queryKey: "open_salary"},
	"leave_request":        {basePath: "/hrd/leave-requests", queryKey: "open_leave_request"},
	"overtime":             {basePath: "/hrd/overtime", queryKey: "open_overtime"},
	"recruitment":          {basePath: "/hrd/recruitment", queryKey: "open_recruitment"},
	"crm_visit":            {basePath: "/crm/visits", queryKey: "open_crm_visit"},
	"pos_self_order":       {basePath: "/pos/fb/live-table", queryKey: "outlet_id"},
}

type NotificationUsecase interface {
	List(ctx context.Context, userID string, page, perPage int, notifType, entityType string, isRead *bool) ([]dto.NotificationResponse, int64, error)
	GetUnreadCount(ctx context.Context, userID string) (*dto.UnreadCountResponse, error)
	MarkAsRead(ctx context.Context, userID, id string) (*dto.NotificationResponse, error)
	MarkAllAsRead(ctx context.Context, userID string) (*dto.MarkAllAsReadResponse, error)
}

type notificationUsecase struct {
	repo repositories.NotificationRepository
}

func NewNotificationUsecase(repo repositories.NotificationRepository) NotificationUsecase {
	return &notificationUsecase{repo: repo}
}

func (u *notificationUsecase) List(ctx context.Context, userID string, page, perPage int, notifType, entityType string, isRead *bool) ([]dto.NotificationResponse, int64, error) {
	items, total, err := u.repo.List(ctx, repositories.ListParams{
		UserID:     userID,
		Type:       notifType,
		EntityType: entityType,
		IsRead:     isRead,
		Limit:      perPage,
		Offset:     (page - 1) * perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	res := make([]dto.NotificationResponse, 0, len(items))
	for _, item := range items {
		res = append(res, mapToResponse(item))
	}
	return res, total, nil
}

func (u *notificationUsecase) GetUnreadCount(ctx context.Context, userID string) (*dto.UnreadCountResponse, error) {
	total, err := u.repo.CountUnread(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.UnreadCountResponse{UnreadCount: total}, nil
}

func (u *notificationUsecase) MarkAsRead(ctx context.Context, userID, id string) (*dto.NotificationResponse, error) {
	item, err := u.repo.MarkAsRead(ctx, userID, id, apptime.Now())
	if err != nil {
		return nil, err
	}
	res := mapToResponse(*item)
	return &res, nil
}

func (u *notificationUsecase) MarkAllAsRead(ctx context.Context, userID string) (*dto.MarkAllAsReadResponse, error) {
	marked, err := u.repo.MarkAllAsRead(ctx, userID, apptime.Now())
	if err != nil {
		return nil, err
	}

	return &dto.MarkAllAsReadResponse{Marked: marked}, nil
}

func mapToResponse(item models.Notification) dto.NotificationResponse {
	return dto.NotificationResponse{
		ID:         item.ID,
		UserID:     item.UserID,
		Type:       item.Type,
		Title:      item.Title,
		Message:    item.Message,
		EntityType: item.EntityType,
		EntityID:   item.EntityID,
		EntityLink: buildEntityLink(item.EntityType, item.EntityID),
		IsRead:     item.IsRead,
		ReadAt:     item.ReadAt,
		CreatedAt:  item.CreatedAt,
	}
}

func buildEntityLink(entityType, entityID string) string {
	rule, ok := entityLinkRules[entityType]
	if !ok {
		return ""
	}

	if entityID == "" || rule.queryKey == "" {
		return rule.basePath
	}

	return fmt.Sprintf("%s?%s=%s", rule.basePath, rule.queryKey, entityID)
}
