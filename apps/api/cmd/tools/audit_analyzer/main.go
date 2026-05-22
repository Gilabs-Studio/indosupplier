package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	coreModels "github.com/gilabs/gims/api/internal/core/data/models"
	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/gilabs/gims/api/internal/finance/domain/reference"
	purchaseModels "github.com/gilabs/gims/api/internal/purchase/data/models"
	salesModels "github.com/gilabs/gims/api/internal/sales/data/models"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type IntegrityIssue struct {
	Level         string
	ReferenceType string
	ReferenceID   string
	Description   string
	CreatedAt     time.Time
}

func main() {
	_ = godotenv.Load()

	if err := config.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := database.Connect(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	db := database.DB

	fmt.Println("🔍 Starting ERP Accounting Integrity Analysis...")
	fmt.Println("================================================")

	issues := analyzeIntegrity(db)

	if len(issues) == 0 {
		fmt.Println("✅ No integrity issues found. System is compliant.")
	} else {
		fmt.Printf("❌ Found %d integrity issues:\n", len(issues))
		for _, issue := range issues {
			fmt.Printf("[%s] %s (%s): %s\n", issue.Level, issue.ReferenceType, issue.ReferenceID, issue.Description)
		}
	}

	fmt.Println("================================================")
	fmt.Println("Analysis completed at:", time.Now().Format(time.RFC1123))
}

func analyzeIntegrity(db *gorm.DB) []IntegrityIssue {
	var issues []IntegrityIssue

	// 1. Check for Orphan Journals
	var orphanJournals []financeModels.JournalEntry
	db.Where("reference_type NOT IN ?", []string{"OPENING_BALANCE", "MANUAL_ADJUSTMENT"}).
		Find(&orphanJournals)

	for _, j := range orphanJournals {
		if j.ReferenceID == nil || *j.ReferenceID == "" {
			issues = append(issues, IntegrityIssue{
				Level:         "CRITICAL",
				ReferenceType: "JOURNAL",
				ReferenceID:   j.ID,
				Description:   "Journal entry has no reference ID",
				CreatedAt:     j.CreatedAt,
			})
			continue
		}

		exists := false
		switch *j.ReferenceType {
		case reference.RefTypeSupplierInvoice:
			var si purchaseModels.SupplierInvoice
			if err := db.First(&si, "id = ?", *j.ReferenceID).Error; err == nil {
				exists = true
			}
		case reference.RefTypeSalesInvoice:
			var ci salesModels.CustomerInvoice
			if err := db.First(&ci, "id = ?", *j.ReferenceID).Error; err == nil {
				exists = true
			}
		case reference.RefTypePurchasePayment:
			var p purchaseModels.PurchasePayment
			if err := db.First(&p, "id = ?", *j.ReferenceID).Error; err == nil {
				exists = true
			}
		case reference.RefTypeSalesPayment:
			var p salesModels.SalesPayment
			if err := db.First(&p, "id = ?", *j.ReferenceID).Error; err == nil {
				exists = true
			}
		default:
			// Custom reference types or unknown
			exists = true
		}

		if !exists {
			issues = append(issues, IntegrityIssue{
				Level:         "CRITICAL",
				ReferenceType: "JOURNAL",
				ReferenceID:   j.ID,
				Description:   fmt.Sprintf("Orphan journal entry: reference %s (%s) not found in source module", *j.ReferenceType, *j.ReferenceID),
				CreatedAt:     j.CreatedAt,
			})
		}
	}

	// 2. Check for Missing Journals (Source Documents with Posted/Paid status but no journal)
	// Supplier Invoices
	var missingSI []purchaseModels.SupplierInvoice
	db.Where("status IN ?", []string{
		string(purchaseModels.SupplierInvoiceStatusUnpaid),
		string(purchaseModels.SupplierInvoiceStatusPartial),
		string(purchaseModels.SupplierInvoiceStatusPaid),
	}).Find(&missingSI)

	for _, si := range missingSI {
		var j financeModels.JournalEntry
		if err := db.Where("reference_type = ? AND reference_id = ?", reference.RefTypeSupplierInvoice, si.ID).First(&j).Error; err == gorm.ErrRecordNotFound {
			issues = append(issues, IntegrityIssue{
				Level:         "ERROR",
				ReferenceType: reference.RefTypeSupplierInvoice,
				ReferenceID:   si.ID,
				Description:   fmt.Sprintf("Confirmed supplier invoice %s has no corresponding journal entry", si.Code),
				CreatedAt:     si.CreatedAt,
			})
		}
	}

	// Sales Invoices
	var missingCI []salesModels.CustomerInvoice
	db.Where("status IN ?", []string{
		string(salesModels.CustomerInvoiceStatusApproved),
		string(salesModels.CustomerInvoiceStatusUnpaid),
		string(salesModels.CustomerInvoiceStatusPartial),
		string(salesModels.CustomerInvoiceStatusPaid),
	}).Find(&missingCI)

	for _, ci := range missingCI {
		var j financeModels.JournalEntry
		if err := db.Where("reference_type = ? AND reference_id = ?", reference.RefTypeSalesInvoice, ci.ID).First(&j).Error; err == gorm.ErrRecordNotFound {
			issues = append(issues, IntegrityIssue{
				Level:         "ERROR",
				ReferenceType: reference.RefTypeSalesInvoice,
				ReferenceID:   ci.ID,
				Description:   fmt.Sprintf("Approved customer invoice %s has no corresponding journal entry", ci.Code),
				CreatedAt:     ci.CreatedAt,
			})
		}
	}

	// 3. Check for Audit Trail consistency
	var reversalJournals []financeModels.JournalEntry
	db.Where("description LIKE ?", "%Reversal%").Find(&reversalJournals)

	for _, j := range reversalJournals {
		var logs []coreModels.AuditLog
		if err := db.Where("target_id = ? AND action LIKE ?", j.ID, "%reverse%").Find(&logs).Error; err == nil && len(logs) == 0 {
			// Check if source document has audit log
			if j.ReferenceID != nil {
				if err := db.Where("target_id = ? AND action LIKE ?", *j.ReferenceID, "%reverse%").Find(&logs).Error; err == nil && len(logs) == 0 {
					issues = append(issues, IntegrityIssue{
						Level:         "WARNING",
						ReferenceType: "JOURNAL",
						ReferenceID:   j.ID,
						Description:   "Reversal journal exists but no 'reverse' action found in audit logs for journal or source document",
						CreatedAt:     j.CreatedAt,
					})
				}
			}
		}
	}

	return issues
}
