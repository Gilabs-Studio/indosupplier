package mapper

import (
	"encoding/json"

	"github.com/gilabs/gims/api/internal/loyalty/data/models"
	"github.com/gilabs/gims/api/internal/loyalty/domain/dto"
)

func ToLoyaltyProgramResponse(m *models.LoyaltyProgram) *dto.LoyaltyProgramResponse {
	var cfg dto.LoyaltyConfigJSON
	_ = json.Unmarshal(m.ConfigJSON, &cfg)

	return &dto.LoyaltyProgramResponse{
		ID:          m.ID,
		OutletID:    m.OutletID,
		Name:        m.Name,
		Description: m.Description,
		Config:      cfg,
		MemberCount: m.MemberCount,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func ToLoyaltyMemberResponse(m *models.LoyaltyMember, customerName string, customerPhone *string, enrolledOutletName *string) *dto.LoyaltyMemberResponse {
	return &dto.LoyaltyMemberResponse{
		ID:                m.ID,
		CustomerID:        m.CustomerID,
		ProgramID:         m.ProgramID,
		MemberCode:        m.MemberCode,
		EnrolledOutletID:  m.EnrolledOutletID,
		EnrolledOutletName: enrolledOutletName,
		CurrentTier:       m.CurrentTier,
		TierBadgeColor:    m.TierBadgeColor,
		LifetimePoints:    m.LifetimePoints,
		PointBalance:      m.PointBalance,
		JoinedAt:          m.JoinedAt,
		LastTransactionAt: m.LastTransactionAt,
		TotalTransactions: m.TotalTransactions,
		CustomerName:      customerName,
		CustomerPhone:     customerPhone,
	}
}

func ToPointLedgerResponse(e *models.LoyaltyPointLedger) *dto.PointLedgerResponse {
	return &dto.PointLedgerResponse{
		ID:              e.ID,
		MemberID:        e.MemberID,
		EntryType:       string(e.EntryType),
		TransactionType: e.TransactionType,
		TransactionID:   e.TransactionID,
		Points:          e.Points,
		BalanceAfter:    e.BalanceAfter,
		LifetimeAfter:   e.LifetimeAfter,
		Multiplier:      e.Multiplier,
		RewardID:        e.RewardID,
		RewardName:      e.RewardName,
		Notes:           e.Notes,
		ExpiresAt:       e.ExpiresAt,
		CreatedAt:       e.CreatedAt,
	}
}
