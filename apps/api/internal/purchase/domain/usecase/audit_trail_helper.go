package usecase

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"gorm.io/gorm"
)

func parsePurchaseAuditMetadata(ctx context.Context, db *gorm.DB, raw string, refCache map[string]string) map[string]interface{} {
	metadata := map[string]interface{}{}
	if strings.TrimSpace(raw) == "" {
		return metadata
	}

	_ = json.Unmarshal([]byte(raw), &metadata)
	enrichPurchaseAuditMetadataReferences(ctx, db, metadata, refCache)
	return metadata
}

func enrichPurchaseAuditMetadataReferences(ctx context.Context, db *gorm.DB, metadata map[string]interface{}, refCache map[string]string) {
	enrichPurchaseSnapshotReferenceNames(ctx, db, metadata, refCache)

	if before, ok := metadata["before"].(map[string]interface{}); ok {
		enrichPurchaseSnapshotReferenceNames(ctx, db, before, refCache)
	}

	if after, ok := metadata["after"].(map[string]interface{}); ok {
		enrichPurchaseSnapshotReferenceNames(ctx, db, after, refCache)
	}
}

func enrichPurchaseSnapshotReferenceNames(ctx context.Context, db *gorm.DB, snapshot map[string]interface{}, refCache map[string]string) {
	type refMap struct {
		idKey   string
		nameKey string
	}

	references := []refMap{
		{idKey: "payment_terms_id", nameKey: "payment_terms_name"},
		{idKey: "supplier_id", nameKey: "supplier_name"},
		{idKey: "customer_id", nameKey: "customer_name"},
		{idKey: "business_unit_id", nameKey: "business_unit_name"},
		{idKey: "business_type_id", nameKey: "business_type_name"},
		{idKey: "delivery_area_id", nameKey: "delivery_area_name"},
		{idKey: "employee_id", nameKey: "employee_name"},
		{idKey: "sales_rep_id", nameKey: "sales_rep_name"},
		{idKey: "sales_quotation_id", nameKey: "sales_quotation_code"},
		{idKey: "purchase_requisition_id", nameKey: "purchase_requisition_code"},
		{idKey: "purchase_requisitions_id", nameKey: "purchase_requisition_code"},
		{idKey: "sales_order_id", nameKey: "sales_order_code"},
		{idKey: "purchase_order_id", nameKey: "purchase_order_code"},
		{idKey: "goods_receipt_id", nameKey: "goods_receipt_code"},
		{idKey: "supplier_invoice_id", nameKey: "supplier_invoice_code"},
		{idKey: "down_payment_invoice_id", nameKey: "down_payment_invoice_code"},
		{idKey: "bank_account_id", nameKey: "bank_account_name"},
		{idKey: "currency_id", nameKey: "currency_code"},
	}

	for _, ref := range references {
		id := stringValue(ref.idKey, snapshot)
		if id == "" {
			continue
		}

		if stringValue(ref.nameKey, snapshot) != "" {
			continue
		}

		if resolved := lookupPurchaseReferenceLabel(ctx, db, ref.idKey, id, refCache); resolved != "" {
			snapshot[ref.nameKey] = resolved
		}
	}
}

func stringValue(key string, data map[string]interface{}) string {
	str, ok := data[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(str)
}

func lookupPurchaseReferenceLabel(ctx context.Context, db *gorm.DB, idKey, id string, refCache map[string]string) string {
	cacheKey := idKey + "|" + id
	if cached, ok := refCache[cacheKey]; ok {
		return cached
	}

	table, column := lookupPurchaseSourceByReferenceKey(idKey)
	if table == "" || column == "" {
		refCache[cacheKey] = ""
		return ""
	}

	queryDB := db.WithContext(ctx)
	if !isGlobalReferenceTable(table) {
		queryDB = database.GetDB(ctx, db)
	}

	row := queryDB.Table(table).Select(column).Where("id = ?", id).Limit(1).Row()
	var value string
	if err := row.Scan(&value); err != nil {
		refCache[cacheKey] = ""
		return ""
	}

	trimmed := strings.TrimSpace(value)
	refCache[cacheKey] = trimmed
	return trimmed
}

func lookupPurchaseSourceByReferenceKey(idKey string) (string, string) {
	sources := map[string]struct {
		table  string
		column string
	}{
		"payment_terms_id":         {table: "payment_terms", column: "name"},
		"supplier_id":              {table: "suppliers", column: "name"},
		"customer_id":              {table: "customers", column: "name"},
		"business_unit_id":         {table: "business_units", column: "name"},
		"business_type_id":         {table: "business_types", column: "name"},
		"delivery_area_id":         {table: "areas", column: "name"},
		"employee_id":              {table: "employees", column: "name"},
		"sales_rep_id":             {table: "employees", column: "name"},
		"sales_quotation_id":       {table: "sales_quotations", column: "code"},
		"purchase_requisition_id":  {table: "purchase_requisitions", column: "code"},
		"purchase_requisitions_id": {table: "purchase_requisitions", column: "code"},
		"sales_order_id":           {table: "sales_orders", column: "code"},
		"purchase_order_id":        {table: "purchase_orders", column: "code"},
		"goods_receipt_id":         {table: "goods_receipts", column: "code"},
		"supplier_invoice_id":      {table: "supplier_invoices", column: "code"},
		"down_payment_invoice_id":  {table: "supplier_invoices", column: "code"},
		"bank_account_id":          {table: "bank_accounts", column: "name"},
		"currency_id":              {table: "currencies", column: "code"},
	}

	source, ok := sources[idKey]
	if !ok {
		return "", ""
	}

	if strings.Contains(source.table, " ") || strings.Contains(source.column, " ") {
		return "", ""
	}

	return source.table, source.column
}

func isGlobalReferenceTable(table string) bool {
	switch strings.TrimSpace(table) {
	case "currencies":
		return true
	default:
		return false
	}
}
