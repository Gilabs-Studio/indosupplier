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

// SalesOrderPrintHandler generates enterprise-grade PDF documents for sales orders
type SalesOrderPrintHandler struct {
	orderUC usecase.SalesOrderUsecase
	db      *gorm.DB
}

// NewSalesOrderPrintHandler creates a new print handler instance
func NewSalesOrderPrintHandler(orderUC usecase.SalesOrderUsecase, db *gorm.DB) *SalesOrderPrintHandler {
	return &SalesOrderPrintHandler{orderUC: orderUC, db: db}
}

// salesOrderPDFData holds all data required for HTML template rendering
type salesOrderPDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	Code         string
	Status       string
	StatusLabel  string
	OrderDate    string
	PaymentTerms string
	BusinessUnit string
	BusinessType string
	SalesRep     string
	Notes        string

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

	Items     []salesOrderPDFItem
	PrintDate string
}

// salesOrderPDFItem represents a single line item in the PDF output
type salesOrderPDFItem struct {
	ProductName string
	ProductCode string
	Quantity    float64
	Price       float64
	Discount    float64
	Subtotal    float64
}

// PrintOrder generates a PDF for the sales order and serves it inline.
func (h *SalesOrderPrintHandler) PrintOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.orderUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSalesOrderNotFound {
			c.String(http.StatusNotFound, "Sales order not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to load sales order")
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
	items := make([]salesOrderPDFItem, 0, len(order.Items))
	for _, item := range order.Items {
		pi := salesOrderPDFItem{
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

	data := &salesOrderPDFData{
		CompanyName:    company.Name,
		CompanyAddress: company.Address,
		CompanyPhone:   company.Phone,
		CompanyEmail:   company.Email,

		Code:        order.Code,
		Status:      order.Status,
		StatusLabel: strings.ToUpper(order.Status[:1]) + order.Status[1:],
		Notes:       order.Notes,

		CustomerName:    order.CustomerName,
		CustomerContact: order.CustomerContact,
		CustomerPhone:   order.CustomerPhone,
		CustomerEmail:   order.CustomerEmail,

		Subtotal:       order.Subtotal,
		DiscountAmount: order.DiscountAmount,
		TaxRate:        order.TaxRate,
		TaxAmount:      order.TaxAmount,
		DeliveryCost:   order.DeliveryCost,
		OtherCost:      order.OtherCost,
		TotalAmount:    order.TotalAmount,

		Items:     items,
		PrintDate: apptime.Now().Format("02 January 2006"),
	}

	// Parse order date
	if order.OrderDate != "" {
		if t, e := time.Parse("2006-01-02", order.OrderDate); e == nil {
			data.OrderDate = t.Format("02 January 2006")
		} else {
			data.OrderDate = order.OrderDate
		}
	}

	// Resolve relation names
	if order.PaymentTerms != nil {
		data.PaymentTerms = order.PaymentTerms.Name
	}
	if order.BusinessUnit != nil {
		data.BusinessUnit = order.BusinessUnit.Name
	}
	if order.BusinessType != nil {
		data.BusinessType = order.BusinessType.Name
	}
	if order.SalesRep != nil {
		data.SalesRep = order.SalesRep.Name
	}

	// Generate PDF and serve it inline — browser opens the built-in PDF viewer directly
	pdfBytes, err := buildSalesOrderPDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	filename := fmt.Sprintf("sales-order-%s.pdf", data.Code)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ===== PDF Generation =====

func buildSalesOrderPDF(d *salesOrderPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	soDrawHeader(pdf, d)
	soDrawMeta(pdf, d)
	soDrawItems(pdf, d)
	soDrawTotals(pdf, d)
	soDrawNotes(pdf, d)
	soDrawFooter(pdf, d)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func soDrawHeader(pdf *fpdf.Fpdf, d *salesOrderPDFData) {
	// Left: document title in italic grey
	pdf.SetFont("Helvetica", "I", 22)
	pdf.SetTextColor(122, 122, 122)
	pdf.CellFormat(pdfCW/2, 10, "Sales Order", "", 0, "L", false, 0, "")

	// Right: company name bold uppercase
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(pdfCW/2, 10, strings.ToUpper(d.CompanyName), "", 1, "R", false, 0, "")

	// Blue accent bar
	pdf.Ln(1)
	pdf.SetFillColor(74, 94, 138)
	pdf.Rect(pdfML, pdf.GetY(), pdfCW, 1.5, "F")
	pdf.Ln(5)
}

func soDrawMeta(pdf *fpdf.Fpdf, d *salesOrderPDFData) {
	startY := pdf.GetY()
	leftX := pdfML
	leftW := 82.0
	rightX := pdfML + leftW + 10
	rightW := pdfCW - leftW - 10

	// Left: company address + order details
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

	metaRow("ORDER NO", "# "+d.Code)
	metaRow("ORDER DATE", d.OrderDate)
	metaRow("STATUS", d.StatusLabel)
	if d.PaymentTerms != "" {
		metaRow("PAYMENT TERMS", d.PaymentTerms)
	}
	if d.BusinessUnit != "" {
		metaRow("BUSINESS UNIT", d.BusinessUnit)
	}

	leftEndY := pdf.GetY()

	// Right: customer info
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
	pdf.SetY(max(leftEndY, rightEndY) + 6)
}

func soDrawItems(pdf *fpdf.Fpdf, d *salesOrderPDFData) {
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

func soDrawTotals(pdf *fpdf.Fpdf, d *salesOrderPDFData) {
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
	pdf.SetLineWidth(0.2)

	pdf.Ln(8)
}

func soDrawNotes(pdf *fpdf.Fpdf, d *salesOrderPDFData) {
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

func soDrawFooter(pdf *fpdf.Fpdf, d *salesOrderPDFData) {
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
		contact := "If you have any questions about this order, please contact " + strings.Join(parts, ", ")
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
