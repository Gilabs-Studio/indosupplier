package dto

import "time"

type AgingFinanceQuery struct {
	AsOfDate       time.Time
	Search         string
	PartnerID      string
	MinAmount      float64
	IncludeCurrent bool
}

type AgingBucketDefinition struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	MinDays *int   `json:"min_days,omitempty"`
	MaxDays *int   `json:"max_days,omitempty"`
}

type AgingBuckets struct {
	Current    float64            `json:"current"`
	Days1To30  float64            `json:"days_1_30"`
	Days31To60 float64            `json:"days_31_60"`
	Days61To90 float64            `json:"days_61_90"`
	Over90     float64            `json:"over_90"`
	Dynamic    map[string]float64 `json:"dynamic,omitempty"`
}

type ARAgingInvoiceRow struct {
	InvoiceID       string       `json:"invoice_id"`
	SourceType      string       `json:"source_type,omitempty"`
	Code            string       `json:"code"`
	InvoiceNumber   *string      `json:"invoice_number"`
	CustomerID      string       `json:"customer_id,omitempty"`
	CustomerName    string       `json:"customer_name,omitempty"`
	InvoiceDate     time.Time    `json:"invoice_date"`
	DueDate         time.Time    `json:"due_date"`
	DaysPastDue     int          `json:"days_past_due"`
	Amount          float64      `json:"amount"`
	RemainingAmount float64      `json:"remaining_amount"`
	Buckets         AgingBuckets `json:"buckets"`
}

type APAgingInvoiceRow struct {
	InvoiceID       string       `json:"invoice_id"`
	SourceType      string       `json:"source_type,omitempty"`
	Code            string       `json:"code"`
	InvoiceNumber   string       `json:"invoice_number"`
	InvoiceDate     time.Time    `json:"invoice_date"`
	DueDate         time.Time    `json:"due_date"`
	DaysPastDue     int          `json:"days_past_due"`
	SupplierID      string       `json:"supplier_id"`
	SupplierName    string       `json:"supplier_name"`
	Amount          float64      `json:"amount"`
	PaidAmount      float64      `json:"paid_amount"`
	RemainingAmount float64      `json:"remaining_amount"`
	Buckets         AgingBuckets `json:"buckets"`
}

type AgingSummary struct {
	PartnerCount     int          `json:"partner_count"`
	InvoiceCount     int          `json:"invoice_count"`
	TotalOutstanding float64      `json:"total_outstanding"`
	TotalOverdue     float64      `json:"total_overdue"`
	TotalCurrent     float64      `json:"total_current"`
	Buckets          AgingBuckets `json:"buckets"`
}

type ARAgingPartnerGroup struct {
	CustomerID       string              `json:"customer_id,omitempty"`
	CustomerName     string              `json:"customer_name"`
	InvoiceCount     int                 `json:"invoice_count"`
	TotalOutstanding float64             `json:"total_outstanding"`
	Buckets          AgingBuckets        `json:"buckets"`
	Invoices         []ARAgingInvoiceRow `json:"invoices"`
}

type APAgingPartnerGroup struct {
	SupplierID       string              `json:"supplier_id,omitempty"`
	SupplierName     string              `json:"supplier_name"`
	InvoiceCount     int                 `json:"invoice_count"`
	TotalOutstanding float64             `json:"total_outstanding"`
	Buckets          AgingBuckets        `json:"buckets"`
	Invoices         []APAgingInvoiceRow `json:"invoices"`
}

type AgingTotals struct {
	Count     int          `json:"count"`
	Remaining float64      `json:"remaining"`
	Buckets   AgingBuckets `json:"buckets"`
}

type ARAgingReportResponse struct {
	AsOfDate     time.Time               `json:"as_of_date"`
	BucketConfig []AgingBucketDefinition `json:"bucket_config,omitempty"`
	Rows         []ARAgingInvoiceRow     `json:"rows"`
	Totals       AgingTotals             `json:"totals"`
	Summary      *AgingSummary           `json:"summary,omitempty"`
	Customers    []ARAgingPartnerGroup   `json:"customers,omitempty"`
}

type APAgingReportResponse struct {
	AsOfDate     time.Time               `json:"as_of_date"`
	BucketConfig []AgingBucketDefinition `json:"bucket_config,omitempty"`
	Rows         []APAgingInvoiceRow     `json:"rows"`
	Totals       AgingTotals             `json:"totals"`
	Summary      *AgingSummary           `json:"summary,omitempty"`
	Suppliers    []APAgingPartnerGroup   `json:"suppliers,omitempty"`
}
