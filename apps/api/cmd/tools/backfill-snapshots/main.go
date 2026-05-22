package main

import (
	"embed"
	"log"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/logger"
)

type backfillStep struct {
	name string
	path string
}

//go:embed sql/*.sql
var sqlFS embed.FS

func mustReadSQL(path string) string {
	b, err := sqlFS.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read SQL file (%s): %v", path, err)
	}
	return strings.TrimSpace(string(b))
}

func main() {
	logger.Init()

	if err := config.Load(); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	steps := []backfillStep{
		{
			name: "purchase_orders.supplier snapshots",
			path: "sql/001_purchase_orders_supplier.sql",
		},
		{
			name: "purchase_orders.payment_terms snapshots",
			path: "sql/002_purchase_orders_payment_terms.sql",
		},
		{
			name: "purchase_orders.business_unit snapshot",
			path: "sql/003_purchase_orders_business_unit.sql",
		},
		{
			name: "purchase_order_items.product snapshots",
			path: "sql/004_purchase_order_items_product.sql",
		},
		{
			name: "goods_receipts.supplier snapshots",
			path: "sql/005_goods_receipts_supplier.sql",
		},
		{
			name: "goods_receipt_items.product snapshots",
			path: "sql/006_goods_receipt_items_product.sql",
		},
		{
			name: "supplier_invoices.supplier snapshots",
			path: "sql/007_supplier_invoices_supplier.sql",
		},
		{
			name: "supplier_invoices.payment_terms snapshots",
			path: "sql/008_supplier_invoices_payment_terms.sql",
		},
		{
			name: "supplier_invoice_items.product snapshots",
			path: "sql/009_supplier_invoice_items_product.sql",
		},
		{
			name: "purchase_payments.bank_account snapshots",
			path: "sql/010_purchase_payments_bank_account.sql",
		},
		{
			name: "purchase_requisitions.supplier snapshots",
			path: "sql/011_purchase_requisitions_supplier.sql",
		},
		{
			name: "purchase_requisitions.payment_terms snapshots",
			path: "sql/012_purchase_requisitions_payment_terms.sql",
		},
		{
			name: "purchase_requisitions.business_unit snapshot",
			path: "sql/013_purchase_requisitions_business_unit.sql",
		},
		{
			name: "purchase_requisition_items.product snapshots",
			path: "sql/014_purchase_requisition_items_product.sql",
		},
		{
			name: "journal_lines.coa snapshots",
			path: "sql/015_journal_lines_coa.sql",
		},
		{
			name: "cash_bank_journals.bank_account snapshots",
			path: "sql/016_cash_bank_journals_bank_account.sql",
		},
		{
			name: "cash_bank_journal_lines.coa snapshots",
			path: "sql/017_cash_bank_journal_lines_coa.sql",
		},
		{
			name: "payments.bank_account snapshots",
			path: "sql/018_payments_bank_account.sql",
		},
		{
			name: "payment_allocations.coa snapshots",
			path: "sql/019_payment_allocations_coa.sql",
		},
		{
			name: "budget_items.coa snapshots",
			path: "sql/020_budget_items_coa.sql",
		},
		{
			name: "non_trade_payables.coa snapshots",
			path: "sql/021_non_trade_payables_coa.sql",
		},
	}

	for _, step := range steps {
		sql := mustReadSQL(step.path)
		if sql == "" {
			log.Printf("Skip empty SQL (%s)", step.name)
			continue
		}
		res := database.DB.Exec(sql)
		if res.Error != nil {
			log.Fatalf("Backfill failed (%s): %v", step.name, res.Error)
		}
		log.Printf("Backfill OK (%s): rows_affected=%d", step.name, res.RowsAffected)
	}

	log.Println("Backfill snapshots completed")
}
