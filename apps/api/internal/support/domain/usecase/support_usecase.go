package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/utils"
	"github.com/gilabs/indosupplier/api/internal/support/data/repositories"
	supportModels "github.com/gilabs/indosupplier/api/internal/support/data/models"
	"github.com/gilabs/indosupplier/api/internal/support/domain/dto"
)

var (
	ErrBuyerProfileNotFound = errors.New("buyer profile not found")
	ErrSupportTicketNotFound = errors.New("support ticket not found")
)

type SupportUsecase interface {
	ListBuyerTickets(ctx context.Context, userID string) ([]dto.SupportTicketListItem, error)
	GetBuyerTicketDetail(ctx context.Context, userID string, ticketNumber string) (*dto.SupportTicketDetail, error)
	CreateBuyerTicket(ctx context.Context, userID string, req *dto.CreateSupportTicketRequest) (*dto.SupportTicketListItem, error)
	ReplyBuyerTicket(ctx context.Context, userID string, ticketNumber string, req *dto.CreateSupportReplyRequest) (*dto.SupportMessage, error)
}

type supportUsecase struct {
	repo repositories.SupportRepository
}

func NewSupportUsecase(repo repositories.SupportRepository) SupportUsecase {
	return &supportUsecase{repo: repo}
}

func (u *supportUsecase) getBuyerProfile(ctx context.Context, userID string) (*models.BuyerProfile, error) {
	profile, err := u.repo.GetBuyerProfileByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBuyerProfileNotFound
		}
		return nil, err
	}
	return profile, nil
}

func supportStatusLabel(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "closed", "resolved":
		return "Closed"
	default:
		return "Open"
	}
}

func supportSenderLabel(senderType string) string {
	switch strings.ToLower(strings.TrimSpace(senderType)) {
	case "support", "agent", "sysadmin":
		return "support"
	default:
		return "buyer"
	}
}

func supportSenderName(profile *models.BuyerProfile, senderType string) string {
	if supportSenderLabel(senderType) == "support" {
		return "IndoSupplier Support"
	}
	if strings.TrimSpace(profile.FullName) != "" {
		return strings.TrimSpace(profile.FullName)
	}
	if strings.TrimSpace(profile.CompanyName) != "" {
		return strings.TrimSpace(profile.CompanyName)
	}
	return "Buyer"
}

func mapTicketItem(ticket supportModels.SupportTicket) dto.SupportTicketListItem {
	return dto.SupportTicketListItem{
		ID:      ticket.TicketNumber,
		Subject: ticket.Subject,
		Date:    utils.FormatDisplayDate(ticket.CreatedAt),
		Status:  supportStatusLabel(ticket.Status),
	}
}

func (u *supportUsecase) ListBuyerTickets(ctx context.Context, userID string) ([]dto.SupportTicketListItem, error) {
	profile, err := u.getBuyerProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	tickets, err := u.repo.ListTicketsByReporter(ctx, "buyer", profile.ID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.SupportTicketListItem, 0, len(tickets))
	for _, ticket := range tickets {
		items = append(items, mapTicketItem(ticket))
	}
	return items, nil
}

func (u *supportUsecase) GetBuyerTicketDetail(ctx context.Context, userID string, ticketNumber string) (*dto.SupportTicketDetail, error) {
	profile, err := u.getBuyerProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	ticket, err := u.repo.GetTicketByNumber(ctx, "buyer", profile.ID, ticketNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupportTicketNotFound
		}
		return nil, err
	}

	messages, err := u.repo.ListMessagesByTicketID(ctx, ticket.ID)
	if err != nil {
		return nil, err
	}

	response := &dto.SupportTicketDetail{
		ID:       ticket.TicketNumber,
		Subject:  ticket.Subject,
		Date:     utils.FormatDisplayDate(ticket.CreatedAt),
		Status:   supportStatusLabel(ticket.Status),
		Messages: make([]dto.SupportMessage, 0, len(messages)),
	}

	for _, message := range messages {
		response.Messages = append(response.Messages, dto.SupportMessage{
			Sender: supportSenderLabel(message.SenderType),
			Name:   supportSenderName(profile, message.SenderType),
			Text:   message.Body,
			Time:   message.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	return response, nil
}

func newTicketNumber(nowTimeUnix int64, year int) string {
	return fmt.Sprintf("TK-%d-%04d", year, nowTimeUnix%10000)
}

func (u *supportUsecase) CreateBuyerTicket(ctx context.Context, userID string, req *dto.CreateSupportTicketRequest) (*dto.SupportTicketListItem, error) {
	profile, err := u.getBuyerProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := apptime.Now()
	slaDeadline := now.AddDate(0, 0, 2)
	ticket := &supportModels.SupportTicket{
		TicketNumber:  newTicketNumber(now.UnixNano(), now.Year()),
		ReporterType:  "buyer",
		ReporterID:    profile.ID,
		Category:      "general",
		Subject:       strings.TrimSpace(req.Subject),
		Description:   strings.TrimSpace(req.Message),
		Priority:      "normal",
		Status:        "open",
		SLADeadlineAt: &slaDeadline,
	}

	message := &supportModels.SupportTicketMessage{
		SenderType: "buyer",
		SenderID:   profile.ID,
		Body:       strings.TrimSpace(req.Message),
	}

	if err := u.repo.CreateTicketWithMessage(ctx, ticket, message); err != nil {
		return nil, err
	}

	item := mapTicketItem(*ticket)
	return &item, nil
}

func (u *supportUsecase) ReplyBuyerTicket(ctx context.Context, userID string, ticketNumber string, req *dto.CreateSupportReplyRequest) (*dto.SupportMessage, error) {
	profile, err := u.getBuyerProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	ticket, err := u.repo.GetTicketByNumber(ctx, "buyer", profile.ID, ticketNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupportTicketNotFound
		}
		return nil, err
	}

	message := &supportModels.SupportTicketMessage{
		SupportTicketID: ticket.ID,
		SenderType:      "buyer",
		SenderID:        profile.ID,
		Body:            strings.TrimSpace(req.Text),
	}
	if err := u.repo.CreateMessage(ctx, message); err != nil {
		return nil, err
	}

	if strings.EqualFold(ticket.Status, "closed") || strings.EqualFold(ticket.Status, "resolved") {
		ticket.Status = "open"
	}
	if err := u.repo.SaveTicket(ctx, ticket); err != nil {
		return nil, err
	}

	response := &dto.SupportMessage{
		Sender: "buyer",
		Name:   supportSenderName(profile, "buyer"),
		Text:   message.Body,
		Time:   message.CreatedAt.Format("2006-01-02 15:04"),
	}
	return response, nil
}
