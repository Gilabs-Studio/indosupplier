package database

import (
	"fmt"

	buyer "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	core "github.com/gilabs/indosupplier/api/internal/core/data/models"
	discovery "github.com/gilabs/indosupplier/api/internal/discovery/data/models"
	monetization "github.com/gilabs/indosupplier/api/internal/monetization/data/models"
	refreshToken "github.com/gilabs/indosupplier/api/internal/refresh_token/data/models"
	rfq "github.com/gilabs/indosupplier/api/internal/rfq/data/models"
	supplier "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	support "github.com/gilabs/indosupplier/api/internal/support/data/models"
	sysadmin "github.com/gilabs/indosupplier/api/internal/sysadmin/data/models"
	trust "github.com/gilabs/indosupplier/api/internal/trust/data/models"
	user "github.com/gilabs/indosupplier/api/internal/user/data/models"
	verification "github.com/gilabs/indosupplier/api/internal/verification/data/models"
	waitingList "github.com/gilabs/indosupplier/api/internal/waiting_list/data/models"
)

// AutoMigrate runs minimal migrations for the cleaned baseline project.
func AutoMigrate() error {
	if err := DB.AutoMigrate(
		&user.User{},
		&refreshToken.RefreshToken{},
		&core.AuditLog{},
		&core.TimeZone{},
		&core.Country{},
		&waitingList.WaitingList{},
		&sysadmin.SystemAdmin{},
		&buyer.BuyerProfile{},
		&buyer.BuyerDocument{},
		&buyer.Bookmark{},
		&buyer.ComparisonSession{},
		&buyer.ComparisonSessionItem{},
		&supplier.Category{},
		&supplier.SupplierProfile{},
		&supplier.SupplierCategory{},
		&supplier.SupplierProduct{},
		&supplier.SupplierProductPhoto{},
		&supplier.SupplierProductTag{},
		&supplier.SupplierPhoto{},
		&supplier.Certification{},
		&supplier.SupplierCertification{},
		&supplier.SupplierDocument{},
		&discovery.AISearchLog{},
		&discovery.SearchBoostCampaign{},
		&rfq.RFQ{},
		&rfq.RFQRecipient{},
		&rfq.RFQMessage{},
		&rfq.RFQAttachment{},
		&trust.SupplierReview{},
		&trust.Notification{},
		&monetization.AdProduct{},
		&monetization.AdCampaign{},
		&monetization.AuctionSession{},
		&monetization.AuctionBid{},
		&monetization.SubscriptionPlan{},
		&monetization.SupplierSubscription{},
		&monetization.Payment{},
		&monetization.Invoice{},
		&monetization.Refund{},
		&verification.VerificationRequest{},
		&verification.SiteVisit{},
		&support.SupportTicket{},
		&support.SupportTicketMessage{},
		&support.SupportTicketAttachment{},
		&support.FAQArticle{},
		&support.AbuseReport{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
