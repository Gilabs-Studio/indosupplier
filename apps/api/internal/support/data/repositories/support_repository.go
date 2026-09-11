package repositories

import (
	"context"

	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	supportModels "github.com/gilabs/indosupplier/api/internal/support/data/models"
)

type SupportRepository interface {
	GetBuyerProfileByUserID(ctx context.Context, userID string) (*buyerModels.BuyerProfile, error)
	ListTicketsByReporter(ctx context.Context, reporterType string, reporterID string) ([]supportModels.SupportTicket, error)
	GetTicketByNumber(ctx context.Context, reporterType string, reporterID string, ticketNumber string) (*supportModels.SupportTicket, error)
	ListMessagesByTicketID(ctx context.Context, ticketID string) ([]supportModels.SupportTicketMessage, error)
	CreateTicket(ctx context.Context, ticket *supportModels.SupportTicket) error
	CreateMessage(ctx context.Context, message *supportModels.SupportTicketMessage) error
	CreateTicketWithMessage(ctx context.Context, ticket *supportModels.SupportTicket, message *supportModels.SupportTicketMessage) error
	SaveTicket(ctx context.Context, ticket *supportModels.SupportTicket) error
}

type supportRepository struct {
	db *gorm.DB
}

func NewSupportRepository(db *gorm.DB) SupportRepository {
	return &supportRepository{db: db}
}

func (r *supportRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *supportRepository) GetBuyerProfileByUserID(ctx context.Context, userID string) (*buyerModels.BuyerProfile, error) {
	var profile buyerModels.BuyerProfile
	if err := r.getDB(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *supportRepository) ListTicketsByReporter(ctx context.Context, reporterType string, reporterID string) ([]supportModels.SupportTicket, error) {
	var tickets []supportModels.SupportTicket
	if err := r.getDB(ctx).
		Where("reporter_type = ? AND reporter_id = ?", reporterType, reporterID).
		Order("updated_at DESC").
		Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *supportRepository) GetTicketByNumber(ctx context.Context, reporterType string, reporterID string, ticketNumber string) (*supportModels.SupportTicket, error) {
	var ticket supportModels.SupportTicket
	if err := r.getDB(ctx).
		Where("ticket_number = ? AND reporter_type = ? AND reporter_id = ?", ticketNumber, reporterType, reporterID).
		First(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *supportRepository) ListMessagesByTicketID(ctx context.Context, ticketID string) ([]supportModels.SupportTicketMessage, error) {
	var messages []supportModels.SupportTicketMessage
	if err := r.getDB(ctx).
		Where("support_ticket_id = ? AND is_internal_note = ?", ticketID, false).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *supportRepository) CreateTicket(ctx context.Context, ticket *supportModels.SupportTicket) error {
	return r.getDB(ctx).Create(ticket).Error
}

func (r *supportRepository) CreateMessage(ctx context.Context, message *supportModels.SupportTicketMessage) error {
	return r.getDB(ctx).Create(message).Error
}

func (r *supportRepository) CreateTicketWithMessage(ctx context.Context, ticket *supportModels.SupportTicket, message *supportModels.SupportTicketMessage) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ticket).Error; err != nil {
			return err
		}
		message.SupportTicketID = ticket.ID
		return tx.Create(message).Error
	})
}

func (r *supportRepository) SaveTicket(ctx context.Context, ticket *supportModels.SupportTicket) error {
	return r.getDB(ctx).Save(ticket).Error
}
