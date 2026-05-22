package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/middleware"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/mapper"
	"gorm.io/gorm"
)

// journalRefTypeAliases maps compact / legacy spellings to canonical keys used for lookups.
var journalRefTypeAliases = map[string]string{
	"GR":                "GOODS_RECEIPT",
	"DO":                "DELIVERY_ORDER",
	"SO":                "SALES_ORDER",
	"PO":                "PURCHASE_ORDER",
	"GOODSRECEIPT":      "GOODS_RECEIPT",
	"GOODS_RECEIPT":     "GOODS_RECEIPT",
	"DELIVERYORDER":     "DELIVERY_ORDER",
	"DELIVERY_ORDER":    "DELIVERY_ORDER",
	"SALESORDER":        "SALES_ORDER",
	"SALES_ORDER":       "SALES_ORDER",
	"PURCHASEORDER":     "PURCHASE_ORDER",
	"PURCHASE_ORDER":    "PURCHASE_ORDER",
	"SALESINVOICE":      "SALES_INVOICE",
	"SALES_INVOICE":     "SALES_INVOICE",
	"SALESINVOICEDP":    "SALES_INVOICE_DP",
	"SALES_INVOICE_DP":  "SALES_INVOICE_DP",
	"SUPPLIERINVOICE":   "SUPPLIER_INVOICE",
	"SUPPLIER_INVOICE":  "SUPPLIER_INVOICE",
	"SUPPLIERINVOICEDP": "SUPPLIER_INVOICE_DP",
	"SUPPLIER_INVOICE_DP": "SUPPLIER_INVOICE_DP",
	"SALESPAYMENT":      "SALES_PAYMENT",
	"SALES_PAYMENT":     "SALES_PAYMENT",
	"PURCHASEPAYMENT":   "PURCHASE_PAYMENT",
	"PURCHASE_PAYMENT":  "PURCHASE_PAYMENT",
	"STOCKOPNAME":       "STOCK_OPNAME",
	"STOCK_OPNAME":      "STOCK_OPNAME",
	"OPNAME":            "STOCK_OPNAME",
	"INVENTORYADJUSTMENT": "INVENTORY_ADJUSTMENT",
	"INVENTORY_ADJUSTMENT": "INVENTORY_ADJUSTMENT",
	"CASHBANK":          "CASH_BANK",
	"CASH_BANK":         "CASH_BANK",
	"CASHIN":            "CASH_BANK",
	"CASHOUT":           "CASH_BANK",
	"TRANSFER":          "CASH_BANK",
	"TRF":               "CASH_BANK",
	"YEAR_END_CLOSING":  "PERIOD_CLOSING",
	"PERIOD_CLOSING":    "PERIOD_CLOSING",
}

func compactJournalRefType(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			continue
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}

func normalizeJournalRefType(refType *string) string {
	if refType == nil {
		return ""
	}
	s := strings.TrimSpace(*refType)
	if s == "" {
		return ""
	}
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, " ", "_")
	if v, ok := journalRefTypeAliases[s]; ok {
		return v
	}
	c := compactJournalRefType(s)
	if v, ok := journalRefTypeAliases[c]; ok {
		return v
	}
	return s
}

type idCodeRow struct {
	ID   string `gorm:"column:id"`
	Code string `gorm:"column:code"`
}

func appendUnique(dst []string, v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return dst
	}
	for _, x := range dst {
		if x == v {
			return dst
		}
	}
	return append(dst, v)
}

func applyQualifiedTenantFilter(ctx context.Context, query *gorm.DB, qualifiedColumns ...string) *gorm.DB {
	if query == nil {
		return query
	}

	if middleware.IsSystemAdmin(ctx) {
		return query
	}

	tenantID := strings.TrimSpace(middleware.TenantFromContext(ctx))
	if tenantID == "" {
		return query
	}

	for _, col := range qualifiedColumns {
		col = strings.TrimSpace(col)
		if col == "" {
			continue
		}
		query = query.Where(col+" = ?", tenantID)
	}

	return query
}

// BatchResolveJournalReferenceCodes resolves human-readable business codes for journal rows (per page).
func BatchResolveJournalReferenceCodes(ctx context.Context, db *gorm.DB, entries []financeModels.JournalEntry) map[string]string {
	out := make(map[string]string, len(entries))
	if len(entries) == 0 || db == nil {
		return out
	}

	type keyed struct {
		journalID string
		refID     string
		kind      string
	}
	var jobs []keyed
	idsByKind := map[string][]string{}

	for i := range entries {
		je := &entries[i]
		if je.ReferenceType == nil || je.ReferenceID == nil {
			continue
		}
		kind := normalizeJournalRefType(je.ReferenceType)
		refID := strings.TrimSpace(*je.ReferenceID)
		if kind == "" || refID == "" {
			continue
		}
		jobs = append(jobs, keyed{journalID: je.ID, refID: refID, kind: kind})
		idsByKind[kind] = appendUnique(idsByKind[kind], refID)
	}

	codeByKindAndID := map[string]map[string]string{}

	resolveSimple := func(kind, table string, ids []string) {
		if len(ids) == 0 {
			return
		}
		var rows []idCodeRow
		if err := db.WithContext(ctx).Table(table).Select("id", "code").Where("id IN ?", ids).Scan(&rows).Error; err != nil {
			return
		}
		m := codeByKindAndID[kind]
		if m == nil {
			m = map[string]string{}
			codeByKindAndID[kind] = m
		}
		for _, r := range rows {
			if strings.TrimSpace(r.Code) != "" {
				m[r.ID] = strings.TrimSpace(r.Code)
			}
		}
	}

	// Document entities with .code
	for _, kind := range []string{
		"SALES_INVOICE", "SALES_INVOICE_DP",
		"SUPPLIER_INVOICE", "SUPPLIER_INVOICE_DP",
	} {
		table := "customer_invoices"
		if strings.HasPrefix(kind, "SUPPLIER") {
			table = "supplier_invoices"
		}
		resolveSimple(kind, table, idsByKind[kind])
	}

	resolveSimple("GOODS_RECEIPT", "goods_receipts", idsByKind["GOODS_RECEIPT"])
	resolveSimple("DELIVERY_ORDER", "delivery_orders", idsByKind["DELIVERY_ORDER"])
	resolveSimple("SALES_ORDER", "sales_orders", idsByKind["SALES_ORDER"])
	resolveSimple("PURCHASE_ORDER", "purchase_orders", idsByKind["PURCHASE_ORDER"])

	// Stock opname uses opname_number
	if ids := idsByKind["STOCK_OPNAME"]; len(ids) > 0 {
		type opRow struct {
			ID           string `gorm:"column:id"`
			OpnameNumber string `gorm:"column:opname_number"`
		}
		var rows []opRow
		if err := db.WithContext(ctx).Table("stock_opnames").Select("id", "opname_number").Where("id IN ?", ids).Scan(&rows).Error; err == nil {
			m := codeByKindAndID["STOCK_OPNAME"]
			if m == nil {
				m = map[string]string{}
				codeByKindAndID["STOCK_OPNAME"] = m
			}
			for _, r := range rows {
				if strings.TrimSpace(r.OpnameNumber) != "" {
					m[r.ID] = strings.TrimSpace(r.OpnameNumber)
				}
			}
		}
	}

	// Valuation runs (reference_id is journal's reference_id pointing at valuation run id)
	for _, kind := range []string{
		"INVENTORY_VALUATION", "CURRENCY_REVALUATION", "COST_ADJUSTMENT", "DEPRECIATION_VALUATION",
	} {
		if ids := idsByKind[kind]; len(ids) > 0 {
			type valRow struct {
				ID          string `gorm:"column:id"`
				ReferenceID string `gorm:"column:reference_id"`
			}
			var rows []valRow
			if err := db.WithContext(ctx).Table("valuation_runs").Select("id", "reference_id").Where("id IN ?", ids).Scan(&rows).Error; err == nil {
				m := codeByKindAndID[kind]
				if m == nil {
					m = map[string]string{}
					codeByKindAndID[kind] = m
				}
				for _, r := range rows {
					code := strings.TrimSpace(r.ReferenceID)
					if code == "" {
						code = "VAL-" + shortID(r.ID)
					} else {
						code = "VAL-" + code
					}
					m[r.ID] = code
				}
			}
		}
	}

	// Finance payments (no code column)
	if ids := idsByKind["PAYMENT"]; len(ids) > 0 {
		m := codeByKindAndID["PAYMENT"]
		if m == nil {
			m = map[string]string{}
			codeByKindAndID["PAYMENT"] = m
		}
		for _, id := range ids {
			m[id] = "FPAY-" + shortID(id)
		}
	}

	// Cash & bank journals
	if ids := idsByKind["CASH_BANK"]; len(ids) > 0 {
		m := codeByKindAndID["CASH_BANK"]
		if m == nil {
			m = map[string]string{}
			codeByKindAndID["CASH_BANK"] = m
		}
		for _, id := range ids {
			m[id] = "CB-" + shortID(id)
		}
	}

	// Financial closing — use period end as readable key
	if ids := idsByKind["PERIOD_CLOSING"]; len(ids) > 0 {
		type clRow struct {
			ID            string    `gorm:"column:id"`
			PeriodEndDate time.Time `gorm:"column:period_end_date"`
		}
		var rows []clRow
		if err := db.WithContext(ctx).Table("financial_closings").Select("id", "period_end_date").Where("id IN ?", ids).Scan(&rows).Error; err == nil {
			m := codeByKindAndID["PERIOD_CLOSING"]
			if m == nil {
				m = map[string]string{}
				codeByKindAndID["PERIOD_CLOSING"] = m
			}
			for _, r := range rows {
				if r.PeriodEndDate.IsZero() {
					continue
				}
				m[r.ID] = "CLOSE-" + r.PeriodEndDate.Format("2006-01-02")
			}
		}
	}

	// Sales payments — prefer bank reference, else invoice business code
	if ids := idsByKind["SALES_PAYMENT"]; len(ids) > 0 {
		type spRow struct {
			ID              string  `gorm:"column:id"`
			ReferenceNumber *string `gorm:"column:reference_number"`
			InvoiceCode     *string `gorm:"column:invoice_code"`
		}
		var rows []spRow
		query := db.WithContext(ctx).Session(&gorm.Session{NewDB: true}).Table("sales_payments sp").
			Select("sp.id, sp.reference_number, ci.code AS invoice_code").
			Joins("LEFT JOIN customer_invoices ci ON ci.id = sp.customer_invoice_id").
			Where("sp.id IN ?", ids)
		query = applyQualifiedTenantFilter(ctx, query, "sp.tenant_id", "ci.tenant_id")
		err := query.Scan(&rows).Error
		if err == nil {
			m := codeByKindAndID["SALES_PAYMENT"]
			if m == nil {
				m = map[string]string{}
				codeByKindAndID["SALES_PAYMENT"] = m
			}
			for _, r := range rows {
				var code string
				if r.ReferenceNumber != nil && strings.TrimSpace(*r.ReferenceNumber) != "" {
					code = strings.TrimSpace(*r.ReferenceNumber)
				} else if r.InvoiceCode != nil && strings.TrimSpace(*r.InvoiceCode) != "" {
					code = "PAY-" + strings.TrimSpace(*r.InvoiceCode)
				} else {
					code = "PAY-" + shortID(r.ID)
				}
				m[r.ID] = code
			}
		}
	}

	// Purchase payments
	if ids := idsByKind["PURCHASE_PAYMENT"]; len(ids) > 0 {
		type ppRow struct {
			ID              string  `gorm:"column:id"`
			ReferenceNumber *string `gorm:"column:reference_number"`
			InvoiceCode     *string `gorm:"column:invoice_code"`
		}
		var rows []ppRow
		query := db.WithContext(ctx).Session(&gorm.Session{NewDB: true}).Table("purchase_payments pp").
			Select("pp.id, pp.reference_number, si.code AS invoice_code").
			Joins("LEFT JOIN supplier_invoices si ON si.id = pp.supplier_invoice_id").
			Where("pp.id IN ?", ids)
		query = applyQualifiedTenantFilter(ctx, query, "pp.tenant_id", "si.tenant_id")
		err := query.Scan(&rows).Error
		if err == nil {
			m := codeByKindAndID["PURCHASE_PAYMENT"]
			if m == nil {
				m = map[string]string{}
				codeByKindAndID["PURCHASE_PAYMENT"] = m
			}
			for _, r := range rows {
				var code string
				if r.ReferenceNumber != nil && strings.TrimSpace(*r.ReferenceNumber) != "" {
					code = strings.TrimSpace(*r.ReferenceNumber)
				} else if r.InvoiceCode != nil && strings.TrimSpace(*r.InvoiceCode) != "" {
					code = "PAY-" + strings.TrimSpace(*r.InvoiceCode)
				} else {
					code = "PAY-" + shortID(r.ID)
				}
				m[r.ID] = code
			}
		}
	}

	for _, job := range jobs {
		byID := codeByKindAndID[job.kind]
		if byID == nil {
			continue
		}
		if code, ok := byID[job.refID]; ok && strings.TrimSpace(code) != "" {
			out[job.journalID] = code
		}
	}

	// Fallback: synthetic code from type + id (same as mapper) when unresolved
	for i := range entries {
		je := &entries[i]
		if _, ok := out[je.ID]; ok {
			continue
		}
		fallback := mapper.BuildJournalReferenceCodeForExport(je.ReferenceType, je.ReferenceID)
		if fallback != nil {
			out[je.ID] = *fallback
		}
	}

	return out
}

func shortID(uuid string) string {
	s := strings.ReplaceAll(strings.TrimSpace(uuid), "-", "")
	if len(s) < 8 {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(s[:8])
}
