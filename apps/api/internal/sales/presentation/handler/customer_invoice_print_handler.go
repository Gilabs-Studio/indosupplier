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

// CustomerInvoicePrintHandler generates enterprise-grade PDF documents for customer invoices
type CustomerInvoicePrintHandler struct {
	invoiceUC usecase.CustomerInvoiceUsecase
	db        *gorm.DB
}

// NewCustomerInvoicePrintHandler creates a new print handler instance
func NewCustomerInvoicePrintHandler(invoiceUC usecase.CustomerInvoiceUsecase, db *gorm.DB) *CustomerInvoicePrintHandler {
	return &CustomerInvoicePrintHandler{invoiceUC: invoiceUC, db: db}
}

// customerInvoicePDFData holds all data required for HTML template rendering
type customerInvoicePDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	Code           string
	InvoiceNumber  string
	InvoiceType    string
	Status         string
	InvoiceDate    string
	DueDate        string
	PaymentTerms   string
	SalesOrderCode string
	Notes          string

	CustomerName    string
	CustomerContact string
	CustomerPhone   string
	CustomerEmail   string

	Subtotal          float64
	TaxRate           float64
	TaxAmount         float64
	DeliveryCost      float64
	OtherCost         float64
	DownPaymentAmount float64
	Amount            float64

	Items     []invoicePDFItem
	PrintDate string
}

// invoicePDFItem represents a single line item in the PDF output
type invoicePDFItem struct {
	ProductName string
	ProductCode string
	Quantity    float64
	Price       float64
	Discount    float64
	Subtotal    float64
}

// PrintInvoice generates a PDF for the customer invoice and serves it inline.
func (h *CustomerInvoicePrintHandler) PrintInvoice(c *gin.Context) {
	id := c.Param("id")

	invoice, err := h.invoiceUC.GetByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load invoice")
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

	// Map line items
	items := make([]invoicePDFItem, 0, len(invoice.Items))
	for _, item := range invoice.Items {
		pi := invoicePDFItem{
			ProductName: "Product",
			Quantity:    item.Quantity,
			Price:       item.Price,
			Discount:    item.Discount,
			Subtotal:    item.Subtotal,
		}
		if item.Product != nil {
			pi.ProductName = item.Product.Name
			pi.ProductCode = item.Product.Code
		}
		items = append(items, pi)
	}

	// Resolve document type label
	invoiceTypeLabel := "Invoice"
	if invoice.Type == "down_payment" {
		invoiceTypeLabel = "Down Payment Invoice"
	} else if invoice.Type == "proforma" {
		invoiceTypeLabel = "Proforma Invoice"
	}

	data := &customerInvoicePDFData{
		CompanyName:    company.Name,
		CompanyAddress: company.Address,
		CompanyPhone:   company.Phone,
		CompanyEmail:   company.Email,

		Code:        invoice.Code,
		InvoiceType: invoiceTypeLabel,
		Status:      strings.ToUpper(invoice.Status[:1]) + invoice.Status[1:],
		Notes:       invoice.Notes,

		Subtotal:          invoice.Subtotal,
		TaxRate:           invoice.TaxRate,
		TaxAmount:         invoice.TaxAmount,
		DeliveryCost:      invoice.DeliveryCost,
		OtherCost:         invoice.OtherCost,
		DownPaymentAmount: invoice.DownPaymentAmount,
		Amount:            invoice.Amount,

		Items:     items,
		PrintDate: apptime.Now().Format("02 January 2006"),
	}

	if invoice.InvoiceNumber != nil {
		data.InvoiceNumber = *invoice.InvoiceNumber
	}

	// Parse dates
	if invoice.InvoiceDate != "" {
		if t, e := time.Parse("2006-01-02", invoice.InvoiceDate); e == nil {
			data.InvoiceDate = t.Format("02 January 2006")
		} else {
			data.InvoiceDate = invoice.InvoiceDate
		}
	}
	if invoice.DueDate != nil && *invoice.DueDate != "" {
		if t, e := time.Parse("2006-01-02", *invoice.DueDate); e == nil {
			data.DueDate = t.Format("02 January 2006")
		} else {
			data.DueDate = *invoice.DueDate
		}
	}

	// Resolve relation names
	if invoice.PaymentTerms != nil {
		data.PaymentTerms = invoice.PaymentTerms.Name
	}
	if invoice.SalesOrder != nil {
		data.SalesOrderCode = invoice.SalesOrder.Code
	}

	// Generate PDF and serve it inline — browser opens the built-in PDF viewer directly
	pdfBytes, err := buildCustomerInvoicePDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	filename := fmt.Sprintf("invoice-%s.pdf", data.Code)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ===== PDF Generation =====

func buildCustomerInvoicePDF(d *customerInvoicePDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	ciDrawHeader(pdf, d)
	ciDrawMeta(pdf, d)
	ciDrawItems(pdf, d)
	ciDrawTotals(pdf, d)
	ciDrawNotes(pdf, d)
	ciDrawFooter(pdf, d)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ciDrawHeader(pdf *fpdf.Fpdf, d *customerInvoicePDFData) {
	pdf.SetFont("Helvetica", "I", 22)
	pdf.SetTextColor(122, 122, 122)
	pdf.CellFormat(pdfCW/2, 10, d.InvoiceType, "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(pdfCW/2, 10, strings.ToUpper(d.CompanyName), "", 1, "R", false, 0, "")

	pdf.Ln(1)
	pdf.SetFillColor(74, 94, 138)
	pdf.Rect(pdfML, pdf.GetY(), pdfCW, 1.5, "F")
	pdf.Ln(5)
}

func ciDrawMeta(pdf *fpdf.Fpdf, d *customerInvoicePDFData) {
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
	pdf.Ln(4)

	const labelW = 38.0
	const lineH = 5.0

	metaRow := func(label, value string) {
		pdf.SetX(leftX)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(102, 102, 102)
		pdf.CellFormat(labelW, lineH, label, "", 0, "L", false, 0, "")
		pdf.CellFormat(3, lineH, ":", "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(leftW-labelW-3, lineH, value, "", 1, "L", false, 0, "")
	}

	metaRow("INVOICE NO", "# "+d.Code)
	if d.InvoiceNumber != "" {
		metaRow("TAX INVOICE NO", d.InvoiceNumber)
	}
	metaRow("INVOICE DATE", d.InvoiceDate)
	if d.DueDate != "" {
		metaRow("DUE DATE", d.DueDate)
	}
	if d.PaymentTerms != "" {
		metaRow("PAYMENT TERMS", d.PaymentTerms)
	}
	if d.SalesOrderCode != "" {
		metaRow("SALES ORDER", d.SalesOrderCode)
	}

	leftEndY := pdf.GetY()

	// Right: customer info (from linked SO if available)
	pdf.SetXY(rightX, startY)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(rightW, 4.5, "BILL TO:", "", 1, "L", false, 0, "")

	writeRight := func(text string, bold bool) {
		if text == "" {
			return
		}
		pdf.SetX(rightX)
		style := ""
		if bold {
			style = "B"
		}
		pdf.SetFont("Helvetica", style, 9)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(rightW, 4.5, text, "", 1, "L", false, 0, "")
	}

	writeRight(d.CustomerName, true)
	writeRight(d.CustomerContact, false)
	writeRight(d.CustomerPhone, false)
	writeRight(d.CustomerEmail, false)

	rightEndY := pdf.GetY()
	pdf.SetY(max(leftEndY, rightEndY) + 6)
}

func ciDrawItems(pdf *fpdf.Fpdf, d *customerInvoicePDFData) {
	cNo := 12.0
	cDesc := 72.0
	cQty := 25.0
	cPrice := 32.5
	cTotal := 32.5
	hdrH := 7.0
	rowH := 6.0

	drawHeader := func() {
		pdf.SetFillColor(74, 94, 138)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.CellFormat(cNo, hdrH, "NO", "", 0, "C", true, 0, "")
		pdf.CellFormat(cDesc, hdrH, "DESCRIPTION", "", 0, "L", true, 0, "")
		pdf.CellFormat(cQty, hdrH, "QUANTITY", "", 0, "R", true, 0, "")
		pdf.CellFormat(cPrice, hdrH, "UNIT PRICE", "", 0, "R", true, 0, "")
		pdf.CellFormat(cTotal, hdrH, "TOTAL", "", 0, "R", true, 0, "")
		pdf.Ln(-1)
	}

	drawHeader()

	for i, item := range d.Items {
		need := rowH
		if item.ProductCode != "" {
			need += 4
		}
		if pdf.GetY()+need > 283 {
			pdf.AddPage()
			drawHeader()
		}

		pdf.SetTextColor(51, 51, 51)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(cNo, rowH, fmt.Sprintf("%d", i+1), "", 0, "C", false, 0, "")
		pdf.CellFormat(cDesc, rowH, item.ProductName, "", 0, "L", false, 0, "")
		pdf.CellFormat(cQty, rowH, fmtQty(item.Quantity), "", 0, "R", false, 0, "")
		pdf.CellFormat(cPrice, rowH, fmtMoney(item.Price), "", 0, "R", false, 0, "")
		pdf.CellFormat(cTotal, rowH, fmtMoney(item.Subtotal), "", 0, "R", false, 0, "")
		pdf.Ln(-1)

		pdf.SetDrawColor(221, 221, 221)
		y := pdf.GetY()
		pdf.Line(pdfML, y, pdfRE, y)

		if item.ProductCode != "" {
			pdf.SetFont("Helvetica", "", 7.5)
			pdf.SetTextColor(153, 153, 153)
			pdf.CellFormat(cNo, 4, "", "", 0, "", false, 0, "")
			pdf.CellFormat(cDesc, 4, item.ProductCode, "", 0, "L", false, 0, "")
			pdf.Ln(-1)
		}
	}

	pdf.Ln(4)
}

func ciDrawTotals(pdf *fpdf.Fpdf, d *customerInvoicePDFData) {
	tx := pdfRE - 80
	lw := 45.0
	vw := 35.0
	tw := lw + vw

	row := func(label, value string, bold, red bool) {
		pdf.SetX(tx)
		s := ""
		if bold {
			s = "B"
		}
		pdf.SetFont("Helvetica", s, 9)
		pdf.SetTextColor(51, 51, 51)
		pdf.CellFormat(lw, 5, label, "", 0, "L", false, 0, "")
		if red {
			pdf.SetTextColor(204, 0, 0)
		}
		pdf.CellFormat(vw, 5, value, "", 1, "R", false, 0, "")
	}

	separator := func() {
		pdf.SetDrawColor(238, 238, 238)
		pdf.SetLineWidth(0.2)
		y := pdf.GetY()
		pdf.Line(tx, y, tx+tw, y)
	}

	if pdf.GetY()+50 > 283 {
		pdf.AddPage()
	}

	row("SUB TOTAL", fmtMoney(d.Subtotal), false, false)
	separator()

	row(fmt.Sprintf("TAX %s%%", fmtTax(d.TaxRate)), fmtMoney(d.TaxAmount), false, false)
	separator()

	if d.DeliveryCost > 0 {
		row("DELIVERY COST", fmtMoney(d.DeliveryCost), false, false)
		separator()
	}

	if d.OtherCost > 0 {
		row("OTHER COST", fmtMoney(d.OtherCost), false, false)
		separator()
	}

	if d.DownPaymentAmount > 0 {
		row("DOWN PAYMENT", "-"+fmtMoney(d.DownPaymentAmount), false, true)
		separator()
	}

	// Grand total with double blue border
	pdf.Ln(1)
	pdf.SetDrawColor(74, 94, 138)
	pdf.SetLineWidth(0.4)
	y := pdf.GetY()
	pdf.Line(tx, y, tx+tw, y)
	pdf.Ln(1)

	pdf.SetX(tx)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(lw, 7, "TOTAL DUE", "", 0, "L", false, 0, "")
	pdf.CellFormat(vw, 7, fmtMoney(d.Amount), "", 1, "R", false, 0, "")

	y = pdf.GetY()
	pdf.Line(tx, y, tx+tw, y)
	pdf.Line(tx, y+0.6, tx+tw, y+0.6)
	pdf.SetLineWidth(0.2)

	pdf.Ln(8)
}

func ciDrawNotes(pdf *fpdf.Fpdf, d *customerInvoicePDFData) {
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

func ciDrawFooter(pdf *fpdf.Fpdf, d *customerInvoicePDFData) {
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
		contact := "If you have any questions about this invoice, please contact us at " + strings.Join(parts, ", ")
		pdf.MultiCell(pdfCW, 4, contact, "", "C", false)
		pdf.Ln(4)
	}

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(74, 94, 138)
	pdf.CellFormat(pdfCW, 6, "THANK YOU FOR YOUR BUSINESS!", "", 1, "C", false, 0, "")
	pdf.Ln(6)

	pdf.SetDrawColor(204, 204, 204)
	pdf.Line(pdfML, pdf.GetY(), pdfRE, pdf.GetY())
	pdf.Ln(3)

	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(153, 153, 153)
	pdf.CellFormat(pdfCW, 4, fmt.Sprintf("This document was generated on %s  -  %s", d.PrintDate, d.CompanyName), "", 1, "C", false, 0, "")
}
