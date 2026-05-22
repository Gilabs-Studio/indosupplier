package mapper

import (
	"github.com/gilabs/gims/api/internal/pos/data/models"
	"github.com/gilabs/gims/api/internal/pos/domain/dto"
)

// ToPOSPaymentResponse maps a POSPayment model to the DTO response
func ToPOSPaymentResponse(p *models.POSPayment) *dto.POSPaymentResponse {
	return &dto.POSPaymentResponse{
		ID:              p.ID,
		OrderID:         p.OrderID,
		Method:          string(p.Method),
		Status:          string(p.Status),
		Amount:          p.Amount,
		TenderAmount:    p.TenderAmount,
		ChangeAmount:    p.ChangeAmount,
		ReferenceNumber: p.ReferenceNumber,
		TransactionID:   p.TransactionID,
		PaymentType:     p.PaymentType,
		VaNumber:        p.VaNumber,
		QrCode:          p.QrCode,
		PaymentURL:      p.PaymentURL,
		ChannelCode:     p.PaymentType,
		ExpiresAt:       p.ExpiresAt,
		PaidAt:          p.PaidAt,
		Notes:           p.Notes,
		CreatedAt:       p.CreatedAt,
	}
}
