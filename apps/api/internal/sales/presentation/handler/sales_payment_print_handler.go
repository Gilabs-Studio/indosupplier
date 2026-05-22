package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/sales/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"gorm.io/gorm"
)

// SalesPaymentPrintHandler generates enterprise-grade PDF payment receipts
type SalesPaymentPrintHandler struct {
	paymentUC usecase.SalesPaymentUsecase
	db        *gorm.DB
}

// NewSalesPaymentPrintHandler creates a new payment print handler instance
func NewSalesPaymentPrintHandler(paymentUC usecase.SalesPaymentUsecase, db *gorm.DB) *SalesPaymentPrintHandler {
	return &SalesPaymentPrintHandler{paymentUC: paymentUC, db: db}
}

// paymentReceiptPDFData holds all data required for HTML template rendering
type paymentReceiptPDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	PaymentID       string
	PaymentDate     string
	Amount          float64
	Method          string
	Status          string
	ReferenceNumber string
	Notes           string

	InvoiceCode    string
	InvoiceNumber  string
	InvoiceType    string
	InvoiceDate    string
	InvoiceDueDate string
	InvoiceAmount  float64

	BankAccountName   string
	BankAccountNumber string
	AccountHolder     string
	Currency          string

	PrintDate string
}

// PrintPayment generates a PDF payment receipt and serves it inline.
func (h *SalesPaymentPrintHandler) PrintPayment(c *gin.Context) {
	id := c.Param("id")

	payment, err := h.paymentUC.GetByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load payment")
		return
	}

	// Resolve company
	var company orgModels.Company
	if companyID := c.Query("company_id"); companyID != "" {
		err := database.GetDB(c.Request.Context(), h.db).
			Where("id = ?", strings.TrimSpace(companyID)).
			First(&company).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.String(http.StatusNotFound, "Company not found")
				return
			}
			c.String(http.StatusInternalServerError, "Failed to load company")
			return
		}
	} else {
		err := database.GetDB(c.Request.Context(), h.db).
			Where("is_active = ?", true).
			Order("created_at ASC").
			First(&company).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			c.String(http.StatusInternalServerError, "Failed to load company")
			return
		}
	}
	if company.Name == "" {
		company.Name = "Gilabs"
	}

	status := payment.Status
	if len(status) > 0 {
		status = strings.ToUpper(status[:1]) + status[1:]
	}

	data := &paymentReceiptPDFData{
		CompanyName:    company.Name,
		CompanyAddress: company.Address,
		CompanyPhone:   company.Phone,
		CompanyEmail:   company.Email,

		PaymentID: payment.ID,
		Amount:    payment.Amount,
		Method:    formatPaymentMethod(payment.Method),
		Status:    status,
		PrintDate: apptime.Now().Format("02 January 2006"),
	}

	if payment.PaymentDate != "" {
		if t, e := time.Parse("2006-01-02", payment.PaymentDate); e == nil {
			data.PaymentDate = t.Format("02 January 2006")
		} else {
			data.PaymentDate = payment.PaymentDate
		}
	}
	if payment.ReferenceNumber != nil {
		data.ReferenceNumber = *payment.ReferenceNumber
	}
	if payment.Notes != nil {
		data.Notes = *payment.Notes
	}

	if payment.Invoice != nil {
		inv := payment.Invoice
		data.InvoiceCode = inv.Code
		if inv.InvoiceNumber != nil {
			data.InvoiceNumber = *inv.InvoiceNumber
		}
		data.InvoiceType = inv.Type
		data.InvoiceAmount = inv.Amount

		if inv.InvoiceDate != "" {
			if t, e := time.Parse("2006-01-02", inv.InvoiceDate); e == nil {
				data.InvoiceDate = t.Format("02 January 2006")
			} else {
				data.InvoiceDate = inv.InvoiceDate
			}
		}
		if inv.DueDate != nil && *inv.DueDate != "" {
			if t, e := time.Parse("2006-01-02", *inv.DueDate); e == nil {
				data.InvoiceDueDate = t.Format("02 January 2006")
			} else {
				data.InvoiceDueDate = *inv.DueDate
			}
		}
	}

	if payment.BankAccount != nil {
		ba := payment.BankAccount
		data.BankAccountName = ba.Name
		data.BankAccountNumber = ba.AccountNumber
		data.AccountHolder = ba.AccountHolder
		data.Currency = ba.Currency
	}

	// Generate PDF and serve it inline — browser opens the built-in PDF viewer directly
	pdfBytes, err := buildPaymentReceiptPDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	filename := fmt.Sprintf("payment-%s.pdf", data.PaymentID)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// formatPaymentMethod converts a raw payment method value to a human-readable label.
func formatPaymentMethod(method string) string {
	switch strings.ToLower(method) {
	case "bank_transfer":
		return "Bank Transfer"
	case "cash":
		return "Cash"
	case "credit_card":
		return "Credit Card"
	case "check":
		return "Check"
	case "giro":
		return "Giro"
	default:
		if method == "" {
			return "-"
		}
		words := strings.Split(strings.ReplaceAll(method, "_", " "), " ")
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
			}
		}
		return strings.Join(words, " ")
	}
}

// ===== PDF Generation =====

func buildPaymentReceiptPDF(d *paymentReceiptPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	prDrawHeader(pdf, d)
	prDrawMeta(pdf, d)
	prDrawPaymentDetails(pdf, d)
	prDrawInvoiceRef(pdf, d)
	prDrawNotes(pdf, d)
	prDrawFooter(pdf, d)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func prDrawHeader(pdf *fpdf.Fpdf, d *paymentReceiptPDFData) {
	pdf.SetFont("Helvetica", "I", 22)
	pdf.SetTextColor(122, 122, 122)
	pdf.CellFormat(pdfCW/2, 10, "Payment Receipt", "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(pdfCW/2, 10, strings.ToUpper(d.CompanyName), "", 1, "R", false, 0, "")

	pdf.Ln(1)
	pdf.SetFillColor(74, 94, 138)
	pdf.Rect(pdfML, pdf.GetY(), pdfCW, 1.5, "F")
	pdf.Ln(5)
}

func prDrawMeta(pdf *fpdf.Fpdf, d *paymentReceiptPDFData) {
	startY := pdf.GetY()
	leftX := pdfML
	leftW := 82.0
	rightX := pdfML + leftW + 10
	rightW := pdfCW - leftW - 10

	pdf.SetXY(leftX, startY)
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(85, 85, 85)

	if d.CompanyAddress != "" {
		pdf.MultiCell(leftW, 4, d.CompanyAddress, "", "L", false)
	}
	if d.CompanyPhone != "" {
		pdf.SetX(leftX)
		pdf.CellFormat(leftW, 4, "Tel: "+d.CompanyPhone, "", 1, "L", false, 0, "")
	}
	if d.CompanyEmail != "" {
		pdf.SetX(leftX)
		pdf.CellFormat(leftW, 4, d.CompanyEmail, "", 1, "L", false, 0, "")
	}

	leftEndY := pdf.GetY()

	// Right: Receipt metadata
	const metaLabelW = 38.0
	const metaLineH = 5.0

	metaRow := func(label, value string) {
		pdf.SetX(rightX)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(102, 102, 102)
		pdf.CellFormat(metaLabelW, metaLineH, label, "", 0, "L", false, 0, "")
		pdf.CellFormat(3, metaLineH, ":", "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(rightW-metaLabelW-3, metaLineH, value, "", 1, "R", false, 0, "")
	}

	pdf.SetXY(rightX, startY)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(rightW, 4.5, "RECEIPT DETAILS", "", 1, "L", false, 0, "")

	metaRow("PAYMENT DATE", d.PaymentDate)
	metaRow("PAYMENT METHOD", d.Method)
	if d.ReferenceNumber != "" {
		metaRow("REFERENCE NO", d.ReferenceNumber)
	}
	metaRow("STATUS", d.Status)

	rightEndY := pdf.GetY()
	pdf.SetY(max(leftEndY, rightEndY) + 6)
}

func prDrawPaymentDetails(pdf *fpdf.Fpdf, d *paymentReceiptPDFData) {
	tx := pdfRE - 80
	lw := 45.0
	vw := 35.0
	tw := lw + vw

	if pdf.GetY()+25 > 283 {
		pdf.AddPage()
	}

	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(0, 4, "PAYMENT AMOUNT", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	pdf.SetDrawColor(74, 94, 138)
	pdf.SetLineWidth(0.4)
	y := pdf.GetY()
	pdf.Line(tx, y, tx+tw, y)
	pdf.Ln(1)

	// Currency prefix
	currency := "IDR"
	if d.Currency != "" {
		currency = d.Currency
	}

	pdf.SetX(tx)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(lw, 8, "AMOUNT PAID   "+currency, "", 0, "L", false, 0, "")
	pdf.CellFormat(vw, 8, fmtMoney(d.Amount), "", 1, "R", false, 0, "")

	y = pdf.GetY()
	pdf.Line(tx, y, tx+tw, y)
	pdf.Line(tx, y+0.6, tx+tw, y+0.6)
	pdf.SetLineWidth(0.2)

	pdf.Ln(8)

	// Bank account details (if present)
	if d.BankAccountName != "" {
		pdf.SetFont("Helvetica", "B", 8.5)
		pdf.SetTextColor(51, 51, 51)
		pdf.CellFormat(0, 4, "TRANSFERRED TO", "", 1, "L", false, 0, "")
		pdf.Ln(1)

		brow := func(label, value string) {
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(102, 102, 102)
			pdf.CellFormat(45, 5, label, "", 0, "L", false, 0, "")
			pdf.CellFormat(3, 5, ":", "", 0, "L", false, 0, "")
			pdf.SetFont("Helvetica", "B", 9)
			pdf.SetTextColor(0, 0, 0)
			pdf.CellFormat(0, 5, value, "", 1, "L", false, 0, "")
		}
		brow("Bank", d.BankAccountName)
		if d.BankAccountNumber != "" {
			brow("Account Number", d.BankAccountNumber)
		}
		if d.AccountHolder != "" {
			brow("Account Holder", d.AccountHolder)
		}
		pdf.Ln(4)
	}
}

func prDrawInvoiceRef(pdf *fpdf.Fpdf, d *paymentReceiptPDFData) {
	if d.InvoiceCode == "" {
		return
	}

	if pdf.GetY()+30 > 283 {
		pdf.AddPage()
	}

	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(0, 4, "INVOICE REFERENCE", "", 1, "L", false, 0, "")
	pdf.Ln(1)

	irow := func(label, value string) {
		if value == "" {
			return
		}
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(102, 102, 102)
		pdf.CellFormat(45, 5, label, "", 0, "L", false, 0, "")
		pdf.CellFormat(3, 5, ":", "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(0, 5, value, "", 1, "L", false, 0, "")
	}

	irow("Invoice No", "# "+d.InvoiceCode)
	irow("Tax Invoice No", d.InvoiceNumber)
	irow("Invoice Date", d.InvoiceDate)
	irow("Due Date", d.InvoiceDueDate)
	if d.InvoiceAmount > 0 {
		irow("Invoice Amount", fmtMoney(d.InvoiceAmount))
	}

	pdf.Ln(6)
}

func prDrawNotes(pdf *fpdf.Fpdf, d *paymentReceiptPDFData) {
	if d.Notes == "" {
		return
	}
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(0, 4, "NOTES", "", 1, "L", false, 0, "")
	pdf.Ln(1)
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(85, 85, 85)
	pdf.MultiCell(pdfCW, 4, d.Notes, "", "L", false)
	pdf.Ln(6)
}

func prDrawFooter(pdf *fpdf.Fpdf, d *paymentReceiptPDFData) {
	parts := make([]string, 0, 2)
	if d.CompanyPhone != "" {
		parts = append(parts, d.CompanyPhone)
	}
	if d.CompanyEmail != "" {
		parts = append(parts, d.CompanyEmail)
	}
	if len(parts) > 0 {
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(85, 85, 85)
		contact := "If you have any questions about this receipt, please contact us at " + strings.Join(parts, ", ")
		pdf.MultiCell(pdfCW, 4, contact, "", "C", false)
		pdf.Ln(4)
	}

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(74, 94, 138)
	pdf.CellFormat(pdfCW, 6, "THANK YOU FOR YOUR PAYMENT!", "", 1, "C", false, 0, "")
	pdf.Ln(6)

	pdf.SetDrawColor(204, 204, 204)
	pdf.Line(pdfML, pdf.GetY(), pdfRE, pdf.GetY())
	pdf.Ln(3)

	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(153, 153, 153)
	pdf.CellFormat(pdfCW, 4, fmt.Sprintf("This document was generated on %s  -  %s", d.PrintDate, d.CompanyName), "", 1, "C", false, 0, "")
}
