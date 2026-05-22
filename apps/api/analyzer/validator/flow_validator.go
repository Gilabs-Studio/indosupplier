package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/gilabs/gims/api/analyzer"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
)

// FlowValidator checks ERP flow chains: Sales and Purchase
type FlowValidator struct{}

func NewFlowValidator() *FlowValidator { return &FlowValidator{} }

func (v *FlowValidator) Name() string { return "ERP Flow Validator (Sales / Purchase)" }

func (v *FlowValidator) Run(cfg *analyzer.Config) []analyzer.Finding {
	var findings []analyzer.Finding
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if cfg.ShouldRunModule("purchase") {
		findings = append(findings, v.checkPurchaseFlow(ctx, cfg)...)
	}
	if cfg.ShouldRunModule("sales") {
		findings = append(findings, v.checkSalesFlow(ctx, cfg)...)
	}
	return findings
}

// checkPurchaseFlow validates PO → GR → Supplier Invoice → Payment chain
func (v *FlowValidator) checkPurchaseFlow(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB
	var findings []analyzer.Finding

	// 1. Approved POs without any Goods Receipt
	var poNoGR int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM purchase_orders po
		WHERE po.status = 'approved' AND po.deleted_at IS NULL
		  AND po.order_date >= ? AND po.order_date <= ?
		  AND NOT EXISTS (
		    SELECT 1 FROM goods_receipts gr WHERE gr.purchase_order_id = po.id AND gr.deleted_at IS NULL
		  )
	`, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&poNoGR)

	if poNoGR > 0 {
		findings = append(findings, analyzer.Finding{
			Code:           "FLOW-P01",
			Severity:       analyzer.SeverityWarning,
			Module:         "purchase",
			Entity:         "purchase_order",
			Flow:           "PO → GR",
			Message:        fmt.Sprintf("%d approved POs without goods receipts", poNoGR),
			Recommendation: "Create goods receipts for pending POs or check if they are intentionally on hold.",
		})
	} else {
		findings = append(findings, analyzer.Finding{
			Code:     "FLOW-P01",
			Severity: analyzer.SeverityPass,
			Module:   "purchase",
			Flow:     "PO → GR",
			Message:  "All approved POs have goods receipts",
		})
	}

	// 2. Goods Receipts without Supplier Invoice
	var grNoSI int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM goods_receipts gr
		WHERE gr.status = 'completed' AND gr.deleted_at IS NULL
		  AND gr.receipt_date >= ? AND gr.receipt_date <= ?
		  AND NOT EXISTS (
		    SELECT 1 FROM supplier_invoices si
		    WHERE si.purchase_order_id = gr.purchase_order_id AND si.deleted_at IS NULL
		  )
	`, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&grNoSI)

	if grNoSI > 0 {
		findings = append(findings, analyzer.Finding{
			Code:           "FLOW-P02",
			Severity:       analyzer.SeverityWarning,
			Module:         "purchase",
			Entity:         "goods_receipt",
			Flow:           "GR → SI",
			Message:        fmt.Sprintf("%d completed GRs without supplier invoices", grNoSI),
			Recommendation: "Create supplier invoices for received goods.",
		})
	} else {
		findings = append(findings, analyzer.Finding{
			Code:     "FLOW-P02",
			Severity: analyzer.SeverityPass,
			Module:   "purchase",
			Flow:     "GR → SI",
			Message:  "All completed GRs have supplier invoices",
		})
	}

	// 3. Approved Supplier Invoices without Payment
	var siNoPay int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM supplier_invoices si
		WHERE si.status = 'approved' AND si.deleted_at IS NULL
		  AND si.invoice_date >= ? AND si.invoice_date <= ?
		  AND NOT EXISTS (
		    SELECT 1 FROM purchase_payments pp
		    WHERE pp.supplier_invoice_id = si.id AND pp.deleted_at IS NULL
		  )
	`, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&siNoPay)

	if siNoPay > 0 {
		findings = append(findings, analyzer.Finding{
			Code:     "FLOW-P03",
			Severity: analyzer.SeverityInfo,
			Module:   "purchase",
			Entity:   "supplier_invoice",
			Flow:     "SI → Payment",
			Message:  fmt.Sprintf("%d approved supplier invoices without payments (may be outstanding)", siNoPay),
		})
	} else {
		findings = append(findings, analyzer.Finding{
			Code:     "FLOW-P03",
			Severity: analyzer.SeverityPass,
			Module:   "purchase",
			Flow:     "SI → Payment",
			Message:  "All approved supplier invoices have payments",
		})
	}

	return findings
}

// checkSalesFlow validates SO → DO → Customer Invoice → Payment chain
func (v *FlowValidator) checkSalesFlow(ctx context.Context, cfg *analyzer.Config) []analyzer.Finding {
	db := database.DB
	var findings []analyzer.Finding

	// 1. Confirmed SOs without Delivery Orders
	var soNoDO int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM sales_orders so
		WHERE so.status = 'confirmed' AND so.deleted_at IS NULL
		  AND so.order_date >= ? AND so.order_date <= ?
		  AND NOT EXISTS (
		    SELECT 1 FROM delivery_orders d WHERE d.sales_order_id = so.id AND d.deleted_at IS NULL
		  )
	`, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&soNoDO)

	if soNoDO > 0 {
		findings = append(findings, analyzer.Finding{
			Code:           "FLOW-S01",
			Severity:       analyzer.SeverityWarning,
			Module:         "sales",
			Entity:         "sales_order",
			Flow:           "SO → DO",
			Message:        fmt.Sprintf("%d confirmed SOs without delivery orders", soNoDO),
			Recommendation: "Create delivery orders for confirmed SOs or check hold status.",
		})
	} else {
		findings = append(findings, analyzer.Finding{
			Code:     "FLOW-S01",
			Severity: analyzer.SeverityPass,
			Module:   "sales",
			Flow:     "SO → DO",
			Message:  "All confirmed SOs have delivery orders",
		})
	}

	// 2. Delivered orders without Customer Invoice
	var doNoInv []string
	db.WithContext(ctx).Raw(`
		SELECT d.delivery_number FROM delivery_orders d
		WHERE d.status = 'delivered' AND d.deleted_at IS NULL
		  AND d.delivery_date >= ? AND d.delivery_date <= ?
		  AND NOT EXISTS (
		    SELECT 1 FROM customer_invoices ci
		    WHERE ci.sales_order_id = d.sales_order_id AND ci.deleted_at IS NULL
		  )
	`, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&doNoInv)

	if len(doNoInv) > 0 {
		evidence := "Missing invoice for DO: "
		for i, v := range doNoInv {
			if i > 2 {
				evidence += fmt.Sprintf("... and %d more", len(doNoInv)-3)
				break
			}
			evidence += v + ", "
		}

		findings = append(findings, analyzer.Finding{
			Code:           "FLOW-S02",
			Severity:       analyzer.SeverityWarning,
			Module:         "sales",
			Entity:         "delivery_order",
			Flow:           "DO → Invoice",
			Message:        fmt.Sprintf("%d delivered DOs without customer invoices", len(doNoInv)),
			Evidence:       evidence,
			Recommendation: "Create customer invoices for delivered orders.",
		})
	} else {
		findings = append(findings, analyzer.Finding{
			Code:     "FLOW-S02",
			Severity: analyzer.SeverityPass,
			Module:   "sales",
			Flow:     "DO → Invoice",
			Message:  "All delivered DOs have customer invoices",
		})
	}

	// 3. Approved Customer Invoices without Payment
	var invNoPay int64
	db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM customer_invoices ci
		WHERE ci.status = 'approved' AND ci.deleted_at IS NULL
		  AND ci.invoice_date >= ? AND ci.invoice_date <= ?
		  AND NOT EXISTS (
		    SELECT 1 FROM sales_payments sp
		    WHERE sp.customer_invoice_id = ci.id AND sp.deleted_at IS NULL
		  )
	`, cfg.FromDateStr(), cfg.ToDateStr()).Scan(&invNoPay)

	if invNoPay > 0 {
		findings = append(findings, analyzer.Finding{
			Code:     "FLOW-S03",
			Severity: analyzer.SeverityInfo,
			Module:   "sales",
			Entity:   "customer_invoice",
			Flow:     "Invoice → Payment",
			Message:  fmt.Sprintf("%d approved customer invoices without payments (may be outstanding)", invNoPay),
		})
	} else {
		findings = append(findings, analyzer.Finding{
			Code:     "FLOW-S03",
			Severity: analyzer.SeverityPass,
			Module:   "sales",
			Flow:     "Invoice → Payment",
			Message:  "All approved customer invoices have payments",
		})
	}

	return findings
}
