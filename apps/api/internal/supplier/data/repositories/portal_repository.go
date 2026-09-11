package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	monetizationModels "github.com/gilabs/indosupplier/api/internal/monetization/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/data/models"
	"github.com/gilabs/indosupplier/api/internal/supplier/domain/dto"
	userModels "github.com/gilabs/indosupplier/api/internal/user/data/models"
)

type PortalRepository interface {
	GetProfileByUserID(ctx context.Context, userID string) (*models.SupplierProfile, error)
	UpdateProfile(ctx context.Context, profile *models.SupplierProfile) error
	GetActiveSubscription(ctx context.Context, supplierProfileID string) (*monetizationModels.SupplierSubscription, error)
	GetSubscriptionPlanByID(ctx context.Context, planID string) (*monetizationModels.SubscriptionPlan, error)
	GetSubscriptionPlanByCode(ctx context.Context, code string) (*monetizationModels.SubscriptionPlan, error)
	CreateSubscriptionPlan(ctx context.Context, plan *monetizationModels.SubscriptionPlan) error
	CreateSubscription(ctx context.Context, sub *monetizationModels.SupplierSubscription) error
	UpdateSubscription(ctx context.Context, sub *monetizationModels.SupplierSubscription) error
	GetInvoices(ctx context.Context, supplierProfileID string) ([]monetizationModels.Invoice, error)
	GetPaymentByID(ctx context.Context, paymentID string) (*monetizationModels.Payment, error)
	CreateInvoice(ctx context.Context, inv *monetizationModels.Invoice) error
	CreatePayment(ctx context.Context, pay *monetizationModels.Payment) error
	GetUserAvatarURL(ctx context.Context, userID string) (string, error)
	UpdateUserAvatar(ctx context.Context, userID string, avatarURL string) error
	GetSupplierCategoryIDs(ctx context.Context, supplierProfileID string) ([]string, error)
	GetTotalSalesMetrics(ctx context.Context, supplierProfileID string, currentStart, currentEnd, prevStart, prevEnd time.Time) (currentMonthAmount float64, prevMonthAmount float64, allTimePaidAmount float64, paidCount int64, err error)
	GetProductMetrics(ctx context.Context, supplierProfileID string) (total int64, active int64, draft int64, featured int64, err error)
	GetMatchingRFQMetrics(ctx context.Context, supplierProfileID string, categoryIDs []string, now time.Time) (totalOpen int64, expiringSoon int64, newThisWeek int64, err error)
	GetMonthlySalesPerformance(ctx context.Context, supplierProfileID string, year int) ([]dto.MonthlySalesPerformance, error)
	GetRecentMatchingRFQs(ctx context.Context, supplierProfileID string, categoryIDs []string, limit int) ([]dto.RecentRFQItem, error)
	GetSupplierReviewsRating(ctx context.Context, supplierProfileID string) (avgRating float64, reviewCount int, err error)
}

type portalRepository struct {
	db *gorm.DB
}

func NewPortalRepository(db *gorm.DB) PortalRepository {
	return &portalRepository{db: db}
}

func (r *portalRepository) getDB(ctx context.Context) *gorm.DB {
	return database.GetDB(ctx, r.db)
}

func (r *portalRepository) UpdateUserAvatar(ctx context.Context, userID string, avatarURL string) error {
	return r.getDB(ctx).
		Model(&userModels.User{}).
		Where("id = ?", userID).
		Update("avatar_url", avatarURL).
		Error
}

func (r *portalRepository) GetUserAvatarURL(ctx context.Context, userID string) (string, error) {
	var avatarURL string
	// Ignore errors, return empty string if scan fails
	_ = r.getDB(ctx).
		Table("users").
		Select("avatar_url").
		Where("id = ?", userID).
		Limit(1).
		Row().
		Scan(&avatarURL)
	return avatarURL, nil
}

func (r *portalRepository) GetProfileByUserID(ctx context.Context, userID string) (*models.SupplierProfile, error) {
	var profile models.SupplierProfile
	err := r.getDB(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *portalRepository) UpdateProfile(ctx context.Context, profile *models.SupplierProfile) error {
	return r.getDB(ctx).Save(profile).Error
}

func (r *portalRepository) GetActiveSubscription(ctx context.Context, supplierProfileID string) (*monetizationModels.SupplierSubscription, error) {
	var sub monetizationModels.SupplierSubscription
	err := r.getDB(ctx).Where("supplier_profile_id = ? AND status = ?", supplierProfileID, "active").First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *portalRepository) GetSubscriptionPlanByID(ctx context.Context, planID string) (*monetizationModels.SubscriptionPlan, error) {
	var plan monetizationModels.SubscriptionPlan
	err := r.getDB(ctx).Where("id = ?", planID).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *portalRepository) GetSubscriptionPlanByCode(ctx context.Context, code string) (*monetizationModels.SubscriptionPlan, error) {
	var plan monetizationModels.SubscriptionPlan
	err := r.getDB(ctx).Where("code = ? AND is_active = ?", code, true).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *portalRepository) CreateSubscriptionPlan(ctx context.Context, plan *monetizationModels.SubscriptionPlan) error {
	return r.getDB(ctx).Create(plan).Error
}

func (r *portalRepository) CreateSubscription(ctx context.Context, sub *monetizationModels.SupplierSubscription) error {
	return r.getDB(ctx).Create(sub).Error
}

func (r *portalRepository) UpdateSubscription(ctx context.Context, sub *monetizationModels.SupplierSubscription) error {
	return r.getDB(ctx).Save(sub).Error
}

func (r *portalRepository) GetInvoices(ctx context.Context, supplierProfileID string) ([]monetizationModels.Invoice, error) {
	var invoices []monetizationModels.Invoice
	// We need to join payments to fetch invoices related to this supplier
	err := r.getDB(ctx).
		Joins("JOIN payments ON payments.id = invoices.payment_id").
		Where("payments.supplier_profile_id = ?", supplierProfileID).
		Order("invoices.created_at DESC").
		Find(&invoices).Error
	if err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *portalRepository) GetPaymentByID(ctx context.Context, paymentID string) (*monetizationModels.Payment, error) {
	var payment monetizationModels.Payment
	if err := r.getDB(ctx).Where("id = ?", paymentID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *portalRepository) CreateInvoice(ctx context.Context, inv *monetizationModels.Invoice) error {
	return r.getDB(ctx).Create(inv).Error
}

func (r *portalRepository) CreatePayment(ctx context.Context, pay *monetizationModels.Payment) error {
	return r.getDB(ctx).Create(pay).Error
}

func (r *portalRepository) GetSupplierCategoryIDs(ctx context.Context, supplierProfileID string) ([]string, error) {
	var categoryIDs []string
	err := r.getDB(ctx).Table("supplier_categories").
		Where("supplier_profile_id = ?", supplierProfileID).
		Pluck("category_id", &categoryIDs).Error
	if err != nil {
		return nil, err
	}
	if len(categoryIDs) == 0 {
		_ = r.getDB(ctx).Table("supplier_products").
			Where("supplier_profile_id = ? AND deleted_at IS NULL AND category_id != ''", supplierProfileID).
			Distinct("category_id").
			Pluck("category_id", &categoryIDs).Error
	}
	return categoryIDs, nil
}

func (r *portalRepository) GetTotalSalesMetrics(ctx context.Context, supplierProfileID string, currentStart, currentEnd, prevStart, prevEnd time.Time) (currentMonthAmount float64, prevMonthAmount float64, allTimePaidAmount float64, paidCount int64, err error) {
	db := r.getDB(ctx)

	err = db.Table("purchase_orders").
		Where("supplier_profile_id = ? AND (payment_status = 'paid' OR status IN ('completed', 'processing'))", supplierProfileID).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&allTimePaidAmount)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	_ = db.Table("purchase_orders").
		Where("supplier_profile_id = ? AND (payment_status = 'paid' OR status IN ('completed', 'processing'))", supplierProfileID).
		Count(&paidCount).Error

	_ = db.Table("purchase_orders").
		Where("supplier_profile_id = ? AND (payment_status = 'paid' OR status IN ('completed', 'processing')) AND created_at >= ? AND created_at <= ?", supplierProfileID, currentStart, currentEnd).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&currentMonthAmount)

	_ = db.Table("purchase_orders").
		Where("supplier_profile_id = ? AND (payment_status = 'paid' OR status IN ('completed', 'processing')) AND created_at >= ? AND created_at <= ?", supplierProfileID, prevStart, prevEnd).
		Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&prevMonthAmount)

	return currentMonthAmount, prevMonthAmount, allTimePaidAmount, paidCount, nil
}

func (r *portalRepository) GetProductMetrics(ctx context.Context, supplierProfileID string) (total int64, active int64, draft int64, featured int64, err error) {
	db := r.getDB(ctx)

	err = db.Table("supplier_products").
		Where("supplier_profile_id = ? AND deleted_at IS NULL", supplierProfileID).
		Count(&total).Error
	if err != nil {
		return 0, 0, 0, 0, err
	}

	_ = db.Table("supplier_products").
		Where("supplier_profile_id = ? AND deleted_at IS NULL AND is_featured = true", supplierProfileID).
		Count(&featured).Error

	active = total
	draft = 0

	return total, active, draft, featured, nil
}

func (r *portalRepository) GetMatchingRFQMetrics(ctx context.Context, supplierProfileID string, categoryIDs []string, now time.Time) (totalOpen int64, expiringSoon int64, newThisWeek int64, err error) {
	db := r.getDB(ctx).Table("rfqs").Where("visibility_status = 'open' AND closed_at IS NULL")
	if len(categoryIDs) > 0 {
		db = db.Where("category_id IN ? OR id IN (SELECT rfq_id FROM rfq_recipients WHERE supplier_profile_id = ?)", categoryIDs, supplierProfileID)
	} else {
		db = db.Where("id IN (SELECT rfq_id FROM rfq_recipients WHERE supplier_profile_id = ?)", supplierProfileID)
	}

	err = db.Count(&totalOpen).Error
	if err != nil {
		return 0, 0, 0, err
	}

	sevenDaysAgo := now.AddDate(0, 0, -7)
	dbNew := r.getDB(ctx).Table("rfqs").Where("visibility_status = 'open' AND closed_at IS NULL AND created_at >= ?", sevenDaysAgo)
	if len(categoryIDs) > 0 {
		dbNew = dbNew.Where("category_id IN ? OR id IN (SELECT rfq_id FROM rfq_recipients WHERE supplier_profile_id = ?)", categoryIDs, supplierProfileID)
	} else {
		dbNew = dbNew.Where("id IN (SELECT rfq_id FROM rfq_recipients WHERE supplier_profile_id = ?)", supplierProfileID)
	}
	_ = dbNew.Count(&newThisWeek).Error

	twentyDaysAgo := now.AddDate(0, 0, -20)
	dbExp := r.getDB(ctx).Table("rfqs").Where("visibility_status = 'open' AND closed_at IS NULL AND created_at < ?", twentyDaysAgo)
	if len(categoryIDs) > 0 {
		dbExp = dbExp.Where("category_id IN ? OR id IN (SELECT rfq_id FROM rfq_recipients WHERE supplier_profile_id = ?)", categoryIDs, supplierProfileID)
	} else {
		dbExp = dbExp.Where("id IN (SELECT rfq_id FROM rfq_recipients WHERE supplier_profile_id = ?)", supplierProfileID)
	}
	_ = dbExp.Count(&expiringSoon).Error

	return totalOpen, expiringSoon, newThisWeek, nil
}

func (r *portalRepository) GetMonthlySalesPerformance(ctx context.Context, supplierProfileID string, year int) ([]dto.MonthlySalesPerformance, error) {
	type monthRow struct {
		Month int     `gorm:"column:month"`
		Total float64 `gorm:"column:total"`
		Count int64   `gorm:"column:count"`
	}

	var rows []monthRow
	err := r.getDB(ctx).Table("purchase_orders").
		Select("EXTRACT(MONTH FROM created_at)::int AS month, COALESCE(SUM(total_amount), 0) AS total, COUNT(id) AS count").
		Where("supplier_profile_id = ? AND (payment_status = 'paid' OR status IN ('completed', 'processing')) AND EXTRACT(YEAR FROM created_at) = ?", supplierProfileID, year).
		Group("month").
		Order("month ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	monthMap := make(map[int]monthRow)
	for _, row := range rows {
		monthMap[row.Month] = row
	}

	monthLetters := []string{"J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"}
	monthNames := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	result := make([]dto.MonthlySalesPerformance, 12)
	for m := 1; m <= 12; m++ {
		amount := 0.0
		var count int64 = 0
		if data, ok := monthMap[m]; ok {
			amount = data.Total
			count = data.Count
		}
		result[m-1] = dto.MonthlySalesPerformance{
			Month:           monthLetters[m-1],
			MonthName:       monthNames[m-1],
			MonthNumber:     m,
			Year:            year,
			Amount:          amount,
			FormattedAmount: fmt.Sprintf("Rp %s", formatRupiahNumber(amount)),
			OrderCount:      count,
		}
	}

	return result, nil
}

func (r *portalRepository) GetRecentMatchingRFQs(ctx context.Context, supplierProfileID string, categoryIDs []string, limit int) ([]dto.RecentRFQItem, error) {
	type rfqQueryRow struct {
		ID                  string    `gorm:"column:id"`
		Title               string    `gorm:"column:title"`
		CategoryName        string    `gorm:"column:category_name"`
		QuantityValue       float64   `gorm:"column:quantity_value"`
		QuantityUnit        string    `gorm:"column:quantity_unit"`
		DestinationLocation string    `gorm:"column:destination_location"`
		CreatedAt           time.Time `gorm:"column:created_at"`
		VisibilityStatus    string    `gorm:"column:visibility_status"`
	}

	var rows []rfqQueryRow
	query := r.getDB(ctx).Table("rfqs").
		Select("rfqs.id, rfqs.title, categories.name as category_name, rfqs.quantity_value, rfqs.quantity_unit, rfqs.destination_location, rfqs.created_at, rfqs.visibility_status").
		Joins("LEFT JOIN categories ON categories.id = rfqs.category_id").
		Where("rfqs.visibility_status = 'open' AND rfqs.closed_at IS NULL")

	if len(categoryIDs) > 0 {
		query = query.Where("rfqs.category_id IN ? OR rfqs.id IN (SELECT rfq_id FROM rfq_recipients WHERE supplier_profile_id = ?)", categoryIDs, supplierProfileID)
	} else {
		query = query.Where("rfqs.id IN (SELECT rfq_id FROM rfq_recipients WHERE supplier_profile_id = ?)", supplierProfileID)
	}

	if err := query.Order("rfqs.created_at DESC").Limit(limit).Scan(&rows).Error; err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []dto.RecentRFQItem{}, nil
	}

	rfqIDs := make([]string, len(rows))
	for i, row := range rows {
		rfqIDs[i] = row.ID
	}

	type replyCountRow struct {
		RFQID string `gorm:"column:rfq_id"`
		Count int64  `gorm:"column:count"`
	}
	var replyCounts []replyCountRow
	_ = r.getDB(ctx).Table("rfq_messages").
		Select("rfq_id, COUNT(id) as count").
		Where("rfq_id IN ?", rfqIDs).
		Group("rfq_id").
		Scan(&replyCounts)

	replyMap := make(map[string]int64)
	for _, rc := range replyCounts {
		replyMap[rc.RFQID] = rc.Count
	}

	items := make([]dto.RecentRFQItem, len(rows))
	for i, row := range rows {
		rfqNumber := row.ID
		if len(rfqNumber) > 8 {
			rfqNumber = "RFQ-" + rfqNumber[:8]
		}

		qty := fmt.Sprintf("%.0f %s", row.QuantityValue, row.QuantityUnit)
		if row.QuantityUnit == "" {
			qty = fmt.Sprintf("%.0f Units", row.QuantityValue)
		}

		items[i] = dto.RecentRFQItem{
			ID:               row.ID,
			RFQNumber:        rfqNumber,
			Product:          row.Title,
			Category:         row.CategoryName,
			Quantity:         qty,
			TargetPort:       row.DestinationLocation,
			Date:             row.CreatedAt.Format("2006-01-02"),
			Replies:          replyMap[row.ID],
			Status:           "Open",
			MatchingCategory: true,
		}
	}

	return items, nil
}

func (r *portalRepository) GetSupplierReviewsRating(ctx context.Context, supplierProfileID string) (avgRating float64, reviewCount int, err error) {
	var res struct {
		AvgRating   float64 `gorm:"column:avg_rating"`
		ReviewCount int     `gorm:"column:review_count"`
	}
	err = r.getDB(ctx).Table("supplier_reviews").
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(id) as review_count").
		Where("supplier_profile_id = ? AND status = 'approved'", supplierProfileID).
		Scan(&res).Error
	if err != nil {
		return 0, 0, err
	}
	return res.AvgRating, res.ReviewCount, nil
}

func formatRupiahNumber(val float64) string {
	intVal := int64(val)
	str := fmt.Sprintf("%d", intVal)
	if len(str) <= 3 {
		return str
	}
	var res []byte
	n := len(str)
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			res = append(res, '.')
		}
		res = append(res, str[i])
	}
	return string(res)
}
