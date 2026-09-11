package database

import (
	"fmt"

	buyer "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	chat "github.com/gilabs/indosupplier/api/internal/chat/data/models"
	content "github.com/gilabs/indosupplier/api/internal/content/data/models"
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
	// Drop legacy unique index if exists to allow bookmarking both supplier and products under idx_bookmark_buyer_item
	DB.Exec("DROP INDEX IF EXISTS idx_bookmark_buyer_supplier")

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
		&buyer.SupplierFollowing{},
		&buyer.ComparisonSession{},
		&buyer.ComparisonSessionItem{},
		&buyer.ComparisonProductSessionItem{},
		&buyer.PurchaseOrder{},
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
		&chat.ChatRoom{},
		&chat.ChatMessage{},
		&support.FAQArticle{},
		&support.AbuseReport{},
		&content.ContentArticle{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	DB.Exec("ALTER TABLE supplier_products ALTER COLUMN currency DROP DEFAULT")
	DB.Exec("ALTER TABLE payments ALTER COLUMN currency DROP DEFAULT")

	// Extensions and performance indexes
	DB.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_supplier_profiles_name_trgm ON supplier_profiles USING gin (company_name gin_trgm_ops)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_supplier_products_name_trgm ON supplier_products USING gin (name gin_trgm_ops)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_po_buyer_status_created ON purchase_orders (buyer_profile_id, status, created_at DESC)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_notifications_recipient_channel_created ON notifications (recipient_id, channel, created_at DESC)")
	DB.Exec("UPDATE supplier_profiles SET slug = LOWER(REGEXP_REPLACE(company_name, '[^a-zA-Z0-9]+', '-', 'g')) WHERE slug IS NULL OR slug = ''")

	return nil
}
