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

// Layout constants for A4 page (210x297mm) — shared across all print handlers in this package.
const (
	pdfML = 18.0  // margin left
	pdfMT = 18.0  // margin top
	pdfMR = 18.0  // margin right
	pdfMB = 14.0  // margin bottom
	pdfPW = 210.0 // page width
	pdfCW = 174.0 // content width (210 - 18 - 18)
	pdfRE = 192.0 // right edge (210 - 18)
)

// quotationPDFData holds all data required for HTML template rendering
type quotationPDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	Code          string
	Status        string
	StatusLabel   string
	QuotationDate string
	ValidUntil    string
	PaymentTerms  string
	BusinessUnit  string
	BusinessType  string
	SalesRep      string
	Notes         string

	CustomerName    string
	CustomerContact string
	CustomerPhone   string
	CustomerEmail   string

	Subtotal       float64
	DiscountAmount float64
	TaxRate        float64
	TaxAmount      float64
	DeliveryCost   float64
	OtherCost      float64
	TotalAmount    float64

	Items     []quotationPDFItem
	PrintDate string
}

// quotationPDFItem represents a single line item in the PDF output
type quotationPDFItem struct {
	ProductName string
	ProductCode string
	Quantity    float64
	Price       float64
	Discount    float64
	Subtotal    float64
}

// SalesQuotationPrintHandler generates enterprise-grade PDF documents for sales quotations
type SalesQuotationPrintHandler struct {
	quotationUC usecase.SalesQuotationUsecase
	db          *gorm.DB
}

// NewSalesQuotationPrintHandler creates a new print handler instance
func NewSalesQuotationPrintHandler(quotationUC usecase.SalesQuotationUsecase, db *gorm.DB) *SalesQuotationPrintHandler {
	return &SalesQuotationPrintHandler{quotationUC: quotationUC, db: db}
}

// PrintQuotation generates a PDF for the sales quotation and serves it inline.
// The browser's built-in PDF viewer handles display and printing.
func (h *SalesQuotationPrintHandler) PrintQuotation(c *gin.Context) {
	id := c.Param("id")

	quotation, err := h.quotationUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesQuotationNotFound {
			c.String(http.StatusNotFound, "Quotation not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to load quotation")
		return
	}

	// Resolve company: prefer explicit company_id, fallback to first active
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
	items := make([]quotationPDFItem, 0, len(quotation.Items))
	for _, item := range quotation.Items {
		pi := quotationPDFItem{
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

	data := &quotationPDFData{
		CompanyName:    company.Name,
		CompanyAddress: company.Address,
		CompanyPhone:   company.Phone,
		CompanyEmail:   company.Email,

		Code:        quotation.Code,
		Status:      quotation.Status,
		StatusLabel: strings.ToUpper(quotation.Status[:1]) + quotation.Status[1:],
		Notes:       quotation.Notes,

		CustomerName:    quotation.CustomerName,
		CustomerContact: quotation.CustomerContact,
		CustomerPhone:   quotation.CustomerPhone,
		CustomerEmail:   quotation.CustomerEmail,

		Subtotal:       quotation.Subtotal,
		DiscountAmount: quotation.DiscountAmount,
		TaxRate:        quotation.TaxRate,
		TaxAmount:      quotation.TaxAmount,
		DeliveryCost:   quotation.DeliveryCost,
		OtherCost:      quotation.OtherCost,
		TotalAmount:    quotation.TotalAmount,

		Items:     items,
		PrintDate: apptime.Now().Format("02 January 2006"),
	}

	// Parse formatted dates
	if quotation.QuotationDate != "" {
		if t, e := time.Parse("2006-01-02", quotation.QuotationDate); e == nil {
			data.QuotationDate = t.Format("02 January 2006")
		} else {
			data.QuotationDate = quotation.QuotationDate
		}
	}
	if quotation.ValidUntil != nil && *quotation.ValidUntil != "" {
		if t, e := time.Parse("2006-01-02", *quotation.ValidUntil); e == nil {
			data.ValidUntil = t.Format("02 January 2006")
		} else {
			data.ValidUntil = *quotation.ValidUntil
		}
	}

	// Resolve relation names
	if quotation.PaymentTerms != nil {
		data.PaymentTerms = quotation.PaymentTerms.Name
	}
	if quotation.BusinessUnit != nil {
		data.BusinessUnit = quotation.BusinessUnit.Name
	}
	if quotation.BusinessType != nil {
		data.BusinessType = quotation.BusinessType.Name
	}
	if quotation.SalesRep != nil {
		data.SalesRep = quotation.SalesRep.Name
	}

	// Generate PDF and serve it inline — browser opens the built-in PDF viewer directly
	pdfBytes, err := buildQuotationPDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	filename := fmt.Sprintf("quotation-%s.pdf", data.Code)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ===== PDF Generation =====

func buildQuotationPDF(d *quotationPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	pdfDrawHeader(pdf, d)
	pdfDrawMeta(pdf, d)
	pdfDrawItems(pdf, d)
	pdfDrawTotals(pdf, d)
	pdfDrawTerms(pdf, d)
	pdfDrawFooter(pdf, d)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ----- Header: document title + company name + blue accent bar -----

func pdfDrawHeader(pdf *fpdf.Fpdf, d *quotationPDFData) {
	// Left: "Sales Quotation" in italic grey
	pdf.SetFont("Helvetica", "I", 22)
	pdf.SetTextColor(122, 122, 122)
	pdf.CellFormat(pdfCW/2, 10, "Sales Quotation", "", 0, "L", false, 0, "")

	// Right: company name in bold uppercase
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(pdfCW/2, 10, strings.ToUpper(d.CompanyName), "", 1, "R", false, 0, "")

	// Blue accent bar
	pdf.Ln(1)
	pdf.SetFillColor(74, 94, 138)
	pdf.Rect(pdfML, pdf.GetY(), pdfCW, 1.5, "F")
	pdf.Ln(5)
}

// ----- Meta: two-column layout with vertically aligned ":" -----

func pdfDrawMeta(pdf *fpdf.Fpdf, d *quotationPDFData) {
	startY := pdf.GetY()
	leftX := pdfML
	leftW := 82.0
	rightX := pdfML + leftW + 10
	rightW := pdfCW - leftW - 10

	// --- Left column: company address then quotation details ---
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

	// Quotation detail rows with aligned ":"
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

	metaRow("QUOTATION NO", "# "+d.Code)
	metaRow("QUOTATION DATE", d.QuotationDate)
	if d.ValidUntil != "" {
		metaRow("VALID UNTIL", d.ValidUntil)
	}
	if d.PaymentTerms != "" {
		metaRow("PAYMENT TERMS", d.PaymentTerms)
	}

	leftEndY := pdf.GetY()

	// --- Right column: customer info ---
	pdf.SetXY(rightX, startY)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(rightW, 4.5, "CUSTOMER:", "", 1, "L", false, 0, "")

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

	// Advance past the taller column
	pdf.SetY(max(leftEndY, rightEndY) + 6)
}

// ----- Items table with blue header -----

func pdfDrawItems(pdf *fpdf.Fpdf, d *quotationPDFData) {
	// Column widths (sum = 174mm)
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
		pdf.CellFormat(cNo, hdrH, "ITEMS", "", 0, "C", true, 0, "")
		pdf.CellFormat(cDesc, hdrH, "DESCRIPTION", "", 0, "L", true, 0, "")
		pdf.CellFormat(cQty, hdrH, "QUANTITY", "", 0, "R", true, 0, "")
		pdf.CellFormat(cPrice, hdrH, "UNIT PRICE", "", 0, "R", true, 0, "")
		pdf.CellFormat(cTotal, hdrH, "TOTAL", "", 0, "R", true, 0, "")
		pdf.Ln(-1)
	}

	drawHeader()

	for i, item := range d.Items {
		// Page-break check: ensure enough space for row + optional code line
		need := rowH
		if item.ProductCode != "" {
			need += 4
		}
		if pdf.GetY()+need > 283 {
			pdf.AddPage()
			drawHeader()
		}

		// Item row
		pdf.SetTextColor(51, 51, 51)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(cNo, rowH, fmt.Sprintf("%d", i+1), "", 0, "C", false, 0, "")
		pdf.CellFormat(cDesc, rowH, item.ProductName, "", 0, "L", false, 0, "")
		pdf.CellFormat(cQty, rowH, fmtQty(item.Quantity), "", 0, "R", false, 0, "")
		pdf.CellFormat(cPrice, rowH, fmtMoney(item.Price), "", 0, "R", false, 0, "")
		pdf.CellFormat(cTotal, rowH, fmtMoney(item.Subtotal), "", 0, "R", false, 0, "")
		pdf.Ln(-1)

		// Row bottom border
		pdf.SetDrawColor(221, 221, 221)
		y := pdf.GetY()
		pdf.Line(pdfML, y, pdfRE, y)

		// Product code sub-line (smaller, grey)
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

// ----- Totals: right-aligned summary with double-border grand total -----

func pdfDrawTotals(pdf *fpdf.Fpdf, d *quotationPDFData) {
	tx := pdfRE - 80 // totals block starts 80mm from right edge
	lw := 45.0       // label width
	vw := 35.0       // value width
	tw := lw + vw    // total block width

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

	// Ensure enough vertical space for the totals block (~50mm)
	if pdf.GetY()+50 > 283 {
		pdf.AddPage()
	}

	row("SUB TOTAL", fmtMoney(d.Subtotal), false, false)
	separator()

	if d.DiscountAmount > 0 {
		row("DISCOUNT", "-"+fmtMoney(d.DiscountAmount), false, true)
		separator()
	}

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
	pdf.CellFormat(lw, 7, "GRAND TOTAL", "", 0, "L", false, 0, "")
	pdf.CellFormat(vw, 7, fmtMoney(d.TotalAmount), "", 1, "R", false, 0, "")

	y = pdf.GetY()
	pdf.Line(tx, y, tx+tw, y)
	pdf.Line(tx, y+0.6, tx+tw, y+0.6)
	pdf.SetLineWidth(0.2) // reset to default

	pdf.Ln(8)
}

// ----- Terms & Conditions + signature line -----

func pdfDrawTerms(pdf *fpdf.Fpdf, d *quotationPDFData) {
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(0, 4, "TERMS AND CONDITIONS", "", 1, "L", false, 0, "")
	pdf.Ln(1)

	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(85, 85, 85)

	if d.Notes != "" {
		pdf.MultiCell(pdfCW, 4, d.Notes, "", "L", false)
	} else {
		pdf.CellFormat(0, 4, "1. Total payment due in 30 days.", "", 1, "L", false, 0, "")
		pdf.CellFormat(0, 4, "2. Please include the quotation number on your cheque.", "", 1, "L", false, 0, "")
	}

	if d.ValidUntil != "" {
		pdf.Ln(6)
		pdf.SetFont("Helvetica", "I", 8.5)
		pdf.CellFormat(0, 4, "To accept this quotation please sign here and return: __________________________", "", 1, "L", false, 0, "")
	}

	pdf.Ln(6)
}

// ----- Footer: contact info + thank you + generation date -----

func pdfDrawFooter(pdf *fpdf.Fpdf, d *quotationPDFData) {
	// Contact info
	parts := make([]string, 0, 3)
	if d.SalesRep != "" {
		parts = append(parts, d.SalesRep)
	}
	if d.CompanyPhone != "" {
		parts = append(parts, d.CompanyPhone)
	}
	if d.CompanyEmail != "" {
		parts = append(parts, d.CompanyEmail)
	}
	if len(parts) > 0 {
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(85, 85, 85)
		contact := "If you have any questions about this quotation, please contact " + strings.Join(parts, ", ")
		pdf.MultiCell(pdfCW, 4, contact, "", "C", false)
		pdf.Ln(4)
	}

	// "Thank You" banner
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(74, 94, 138)
	pdf.CellFormat(pdfCW, 6, "THANK YOU FOR YOUR BUSINESS!", "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Separator line
	pdf.SetDrawColor(204, 204, 204)
	pdf.Line(pdfML, pdf.GetY(), pdfRE, pdf.GetY())
	pdf.Ln(3)

	// Generation date
	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(153, 153, 153)
	pdf.CellFormat(pdfCW, 4, fmt.Sprintf("This document was generated on %s  -  %s", d.PrintDate, d.CompanyName), "", 1, "C", false, 0, "")
}

// fmtMoney, fmtQty, fmtTax, and moneyFmt are defined in print_template_helpers.go (shared across this package).
