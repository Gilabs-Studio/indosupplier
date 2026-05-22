package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gilabs/gims/api/internal/core/response"
	"github.com/gilabs/gims/api/internal/tenant/data/models"
	"github.com/gilabs/gims/api/internal/tenant/data/repositories"
	"github.com/gilabs/gims/api/internal/tenant/domain/dto"
	"gorm.io/gorm"
)

var ErrPaymentTransactionNotFound = errors.New("payment transaction not found")

// PaymentTransactionUsecase handles payment transaction operations.
type PaymentTransactionUsecase interface {
	Create(ctx context.Context, req *dto.CreatePaymentTransactionRequest) (*dto.PaymentTransactionResponse, error)
	GetByID(ctx context.Context, id string) (*dto.PaymentTransactionResponse, error)
	ListByTenant(ctx context.Context, tenantID string, params dto.PaymentHistoryListParams) ([]*dto.PaymentTransactionResponse, *response.PaginationMeta, error)
	ListBySubscription(ctx context.Context, subscriptionID string, params dto.PaymentHistoryListParams) ([]*dto.PaymentTransactionResponse, *response.PaginationMeta, error)
	UpdateStatus(ctx context.Context, id string, req *dto.UpdatePaymentTransactionRequest) (*dto.PaymentTransactionResponse, error)
}

type paymentTransactionUsecase struct {
	paymentTxnRepo repositories.PaymentTransactionRepository
	subRepo        repositories.SubscriptionRepository
}

// NewPaymentTransactionUsecase creates a new PaymentTransactionUsecase.
func NewPaymentTransactionUsecase(
	paymentTxnRepo repositories.PaymentTransactionRepository,
	subRepo repositories.SubscriptionRepository,
) PaymentTransactionUsecase {
	return &paymentTransactionUsecase{
		paymentTxnRepo: paymentTxnRepo,
		subRepo:        subRepo,
	}
}

func (u *paymentTransactionUsecase) Create(ctx context.Context, req *dto.CreatePaymentTransactionRequest) (*dto.PaymentTransactionResponse, error) {
	txn := &models.PaymentTransaction{
		TenantID:          req.TenantID,
		SubscriptionID:    req.SubscriptionID,
		Provider:          models.PaymentProvider(req.Provider),
		Status:            models.PaymentTransactionStatus(req.Status),
		PaymentMethod:     models.PaymentMethod(req.PaymentMethod),
		AmountIDR:         req.AmountIDR,
		ProviderInvoiceID: req.ProviderInvoiceID,
		ProviderPaymentID: req.ProviderPaymentID,
		ReceiptURL:        req.ReceiptURL,
		InvoiceURL:        req.InvoiceURL,
		Description:       req.Description,
		Metadata:          req.Metadata,
		Notes:             req.Notes,
	}

	if err := u.paymentTxnRepo.Create(ctx, txn); err != nil {
		return nil, fmt.Errorf("failed to create payment transaction: %w", err)
	}

	return paymentTxnModelToDTO(txn), nil
}

func (u *paymentTransactionUsecase) GetByID(ctx context.Context, id string) (*dto.PaymentTransactionResponse, error) {
	txn, err := u.paymentTxnRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentTransactionNotFound
		}
		return nil, err
	}
	return paymentTxnModelToDTO(txn), nil
}

func (u *paymentTransactionUsecase) ListByTenant(ctx context.Context, tenantID string, params dto.PaymentHistoryListParams) ([]*dto.PaymentTransactionResponse, *response.PaginationMeta, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}

	txns, total, err := u.paymentTxnRepo.ListByTenantID(ctx, tenantID, params.Page, params.PerPage)
	if err != nil {
		if !isMissingTenantIDColumnErr(err) {
			return nil, nil, err
		}
		txns = nil
		total = 0
	}
	if total > 0 {
		result := make([]*dto.PaymentTransactionResponse, 0, len(txns))
		for _, t := range txns {
			result = append(result, paymentTxnModelToDTO(t))
		}
		pagination := response.NewPaginationMeta(params.Page, params.PerPage, int(total))
		return result, pagination, nil
	}

	if u.subRepo == nil {
		pagination := response.NewPaginationMeta(params.Page, params.PerPage, 0)
		return []*dto.PaymentTransactionResponse{}, pagination, nil
	}

	subs, err := u.subRepo.ListByTenantID(ctx, tenantID)
	if err != nil {
		return nil, nil, err
	}

	result := make([]*dto.PaymentTransactionResponse, 0, len(subs))
	for _, sub := range subs {
		result = append(result, paymentTxnFromSubscription(sub))
	}

	pagination := response.NewPaginationMeta(params.Page, params.PerPage, len(subs))
	return result, pagination, nil
}

func isMissingTenantIDColumnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "column \"tenant_id\" does not exist") ||
		strings.Contains(msg, "column tenant_id does not exist")
}

func (u *paymentTransactionUsecase) ListBySubscription(ctx context.Context, subscriptionID string, params dto.PaymentHistoryListParams) ([]*dto.PaymentTransactionResponse, *response.PaginationMeta, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 || params.PerPage > 100 {
		params.PerPage = 20
	}

	txns, total, err := u.paymentTxnRepo.ListBySubscriptionID(ctx, subscriptionID, params.Page, params.PerPage)
	if err != nil {
		return nil, nil, err
	}

	result := make([]*dto.PaymentTransactionResponse, 0, len(txns))
	for _, t := range txns {
		result = append(result, paymentTxnModelToDTO(t))
	}

	pagination := response.NewPaginationMeta(params.Page, params.PerPage, int(total))
	return result, pagination, nil
}

func (u *paymentTransactionUsecase) UpdateStatus(ctx context.Context, id string, req *dto.UpdatePaymentTransactionRequest) (*dto.PaymentTransactionResponse, error) {
	updates := make(map[string]interface{})

	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.PaymentMethod != "" {
		updates["payment_method"] = req.PaymentMethod
	}
	if req.ProviderPaymentID != "" {
		updates["provider_payment_id"] = req.ProviderPaymentID
	}
	if req.ReceiptURL != "" {
		updates["receipt_url"] = req.ReceiptURL
	}
	if req.Notes != "" {
		updates["notes"] = req.Notes
	}

	if len(updates) == 0 {
		return u.GetByID(ctx, id)
	}

	if err := u.paymentTxnRepo.Update(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to update payment transaction: %w", err)
	}

	return u.GetByID(ctx, id)
}

// paymentTxnModelToDTO converts a PaymentTransaction model to its DTO representation.
func paymentTxnModelToDTO(t *models.PaymentTransaction) *dto.PaymentTransactionResponse {
	return &dto.PaymentTransactionResponse{
		ID:               t.ID,
		Provider:         string(t.Provider),
		Status:           string(t.Status),
		PaymentMethod:    string(t.PaymentMethod),
		AmountIDR:        t.AmountIDR,
		ProviderInvoiceID: t.ProviderInvoiceID,
		ReceiptURL:       t.ReceiptURL,
		InvoiceURL:       t.InvoiceURL,
		Description:      t.Description,
		PaidAt:           t.PaidAt,
		ExpiresAt:        t.ExpiresAt,
		CreatedAt:        t.CreatedAt,
	}
}

func paymentTxnFromSubscription(sub *models.TenantSubscription) *dto.PaymentTransactionResponse {
	if sub == nil {
		return nil
	}

	provider := "internal"
	if sub.XenditInvoiceID != nil && *sub.XenditInvoiceID != "" {
		provider = "xendit"
	}

	status := string(sub.Status)
	if status == "active" || status == "trial" {
		status = "paid"
	}
	if status == "cancelled" {
		status = "canceled"
	}
	if status == "" {
		status = "paid"
	}

	id := sub.ID
	if sub.XenditInvoiceID != nil && *sub.XenditInvoiceID != "" {
		id = *sub.XenditInvoiceID
	}

	paidAt := sub.StartsAt
	return &dto.PaymentTransactionResponse{
		ID:               id,
		Provider:         provider,
		Status:           status,
		PaymentMethod:    "invoice",
		AmountIDR:        sub.AmountPaidIDR,
		ProviderInvoiceID: id,
		ReceiptURL:       "",
		InvoiceURL:       "",
		Description:      sub.Notes,
		PaidAt:           &paidAt,
		ExpiresAt:        sub.ExpiresAt,
		CreatedAt:        sub.CreatedAt,
	}
}
