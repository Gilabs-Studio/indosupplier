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

// CustomerInvoiceDPPrintHandler generates enterprise-grade PDF documents for down payment invoices
type CustomerInvoiceDPPrintHandler struct {
	invoiceDpUC usecase.CustomerInvoiceDownPaymentUsecase
	db          *gorm.DB
}

// NewCustomerInvoiceDPPrintHandler creates a new print handler instance
func NewCustomerInvoiceDPPrintHandler(invoiceDpUC usecase.CustomerInvoiceDownPaymentUsecase, db *gorm.DB) *CustomerInvoiceDPPrintHandler {
	return &CustomerInvoiceDPPrintHandler{invoiceDpUC: invoiceDpUC, db: db}
}

// dpInvoicePDFData holds all data required for HTML template rendering
type dpInvoicePDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	Code               string
	InvoiceNumber      string
	InvoiceDate        string
	DueDate            string
	SalesOrderCode     string
	RelatedInvoiceCode string
	Notes              string
	Status             string

	CustomerName string

	Amount          float64
	RemainingAmount float64

	PrintDate string
}

// PrintDownPaymentInvoice generates a PDF for the down payment invoice and serves it inline.
func (h *CustomerInvoiceDPPrintHandler) PrintDownPaymentInvoice(c *gin.Context) {
	id := c.Param("id")

	dp, err := h.invoiceDpUC.GetByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load down payment invoice")
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

	data := &dpInvoicePDFData{
		CompanyName:    company.Name,
		CompanyAddress: company.Address,
		CompanyPhone:   company.Phone,
		CompanyEmail:   company.Email,

		Code:            dp.Code,
		Status:          strings.ToUpper(dp.Status[:1]) + dp.Status[1:],
		Amount:          dp.Amount,
		RemainingAmount: dp.RemainingAmount,
		PrintDate:       apptime.Now().Format("02 January 2006"),
	}

	if dp.InvoiceNumber != nil {
		data.InvoiceNumber = *dp.InvoiceNumber
	}
	if dp.Notes != nil {
		data.Notes = *dp.Notes
	}
	if dp.RelatedInvoiceCode != nil {
		data.RelatedInvoiceCode = *dp.RelatedInvoiceCode
	}
	if dp.SalesOrder != nil {
		data.SalesOrderCode = dp.SalesOrder.Code
	}

	// Parse dates
	if dp.InvoiceDate != "" {
		if t, e := time.Parse("2006-01-02", dp.InvoiceDate); e == nil {
			data.InvoiceDate = t.Format("02 January 2006")
		} else {
			data.InvoiceDate = dp.InvoiceDate
		}
	}
	if dp.DueDate != nil && *dp.DueDate != "" {
		if t, e := time.Parse("2006-01-02", *dp.DueDate); e == nil {
			data.DueDate = t.Format("02 January 2006")
		} else {
			data.DueDate = *dp.DueDate
		}
	}

	// Generate PDF and serve it inline — browser opens the built-in PDF viewer directly
	pdfBytes, err := buildDPInvoicePDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	filename := fmt.Sprintf("dp-invoice-%s.pdf", data.Code)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ===== PDF Generation =====

func buildDPInvoicePDF(d *dpInvoicePDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	dpDrawHeader(pdf, d)
	dpDrawMeta(pdf, d)
	dpDrawAmount(pdf, d)
	dpDrawNotes(pdf, d)
	dpDrawSignature(pdf, d)
	dpDrawFooter(pdf, d)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func dpDrawHeader(pdf *fpdf.Fpdf, d *dpInvoicePDFData) {
	pdf.SetFont("Helvetica", "I", 22)
	pdf.SetTextColor(122, 122, 122)
	pdf.CellFormat(pdfCW/2, 10, "Down Payment Invoice", "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(pdfCW/2, 10, strings.ToUpper(d.CompanyName), "", 1, "R", false, 0, "")

	pdf.Ln(1)
	pdf.SetFillColor(74, 94, 138)
	pdf.Rect(pdfML, pdf.GetY(), pdfCW, 1.5, "F")
	pdf.Ln(5)
}

func dpDrawMeta(pdf *fpdf.Fpdf, d *dpInvoicePDFData) {
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
	if d.SalesOrderCode != "" {
		metaRow("SALES ORDER", d.SalesOrderCode)
	}
	if d.RelatedInvoiceCode != "" {
		metaRow("FINAL INVOICE", d.RelatedInvoiceCode)
	}
	metaRow("STATUS", d.Status)

	leftEndY := pdf.GetY()

	// Right: customer section (placeholder — DP links via SO)
	pdf.SetXY(rightX, startY)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(rightW, 4.5, "BILL TO:", "", 1, "L", false, 0, "")

	if d.CustomerName != "" {
		pdf.SetX(rightX)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(rightW, 4.5, d.CustomerName, "", 1, "L", false, 0, "")
	}

	rightEndY := pdf.GetY()
	pdf.SetY(max(leftEndY, rightEndY) + 6)
}

func dpDrawAmount(pdf *fpdf.Fpdf, d *dpInvoicePDFData) {
	tx := pdfRE - 80
	lw := 45.0
	vw := 35.0
	tw := lw + vw

	if pdf.GetY()+30 > 283 {
		pdf.AddPage()
	}

	// Amount section header
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(0, 4, "PAYMENT DETAILS", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	rowLine := func(label, value string, bold bool) {
		pdf.SetX(tx)
		s := ""
		if bold {
			s = "B"
		}
		pdf.SetFont("Helvetica", s, 9)
		pdf.SetTextColor(51, 51, 51)
		pdf.CellFormat(lw, 5, label, "", 0, "L", false, 0, "")
		pdf.CellFormat(vw, 5, value, "", 1, "R", false, 0, "")
		pdf.SetDrawColor(238, 238, 238)
		pdf.SetLineWidth(0.2)
		y := pdf.GetY()
		pdf.Line(tx, y, tx+tw, y)
	}

	rowLine("DOWN PAYMENT AMOUNT", fmtMoney(d.Amount), false)

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
	pdf.CellFormat(lw, 7, "AMOUNT DUE", "", 0, "L", false, 0, "")
	pdf.CellFormat(vw, 7, fmtMoney(d.Amount), "", 1, "R", false, 0, "")

	y = pdf.GetY()
	pdf.Line(tx, y, tx+tw, y)
	pdf.Line(tx, y+0.6, tx+tw, y+0.6)
	pdf.SetLineWidth(0.2)

	pdf.Ln(8)
}

func dpDrawNotes(pdf *fpdf.Fpdf, d *dpInvoicePDFData) {
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

func dpDrawSignature(pdf *fpdf.Fpdf, d *dpInvoicePDFData) {
	pdf.SetFont("Helvetica", "I", 8.5)
	pdf.SetTextColor(85, 85, 85)
	pdf.CellFormat(0, 4, "To confirm this down payment, please sign here and return: __________________________", "", 1, "L", false, 0, "")
	pdf.Ln(6)
}

func dpDrawFooter(pdf *fpdf.Fpdf, d *dpInvoicePDFData) {
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
