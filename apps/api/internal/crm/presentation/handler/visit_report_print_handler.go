package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/crm/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

// Layout constants for A4 page (210x297mm)
const (
	vrML = 18.0  // margin left
	vrMT = 18.0  // margin top
	vrMR = 18.0  // margin right
	vrPW = 210.0 // page width
	vrCW = 174.0 // content width (210 - 18 - 18)
	vrRE = 192.0 // right edge (210 - 18)
)

// accent color for headers and lines
const (
	vrAccentR = 74
	vrAccentG = 94
	vrAccentB = 138
)

// visitPDFData holds all data required for PDF generation
type visitPDFData struct {
	Code          string
	Status        string
	VisitDate     string
	ScheduledTime string
	CheckInAt     string
	CheckOutAt    string

	EmployeeName string
	EmployeeCode string

	CustomerName  string
	ContactPerson string
	ContactPhone  string
	Address       string

	DealCode  string
	DealTitle string
	LeadCode  string
	LeadName  string

	Purpose   string
	Notes     string
	Result    string
	Outcome   string
	NextSteps string

	Products  []visitPDFProduct
	PrintDate string
}

// visitPDFProduct represents a product interest line item
type visitPDFProduct struct {
	ProductName   string
	ProductCode   string
	InterestLevel int
	Quantity      float64
	Price         float64
	Notes         string
}

// VisitReportPrintHandler generates PDF documents for visit reports
type VisitReportPrintHandler struct {
	visitUC usecase.VisitReportUsecase
}

// NewVisitReportPrintHandler creates a new print handler instance
func NewVisitReportPrintHandler(visitUC usecase.VisitReportUsecase) *VisitReportPrintHandler {
	return &VisitReportPrintHandler{visitUC: visitUC}
}

// PrintVisitReport generates a PDF for the visit report and serves it inline
func (h *VisitReportPrintHandler) PrintVisitReport(c *gin.Context) {
	id := c.Param("id")

	visit, err := h.visitUC.GetByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load visit report")
		return
	}
	if visit == nil {
		c.String(http.StatusNotFound, "Visit report not found")
		return
	}

	// Map response to PDF data struct
	data := visitPDFData{
		Code:      visit.Code,
		Status:    "-",
		VisitDate: visit.VisitDate,
		PrintDate: apptime.Now().Format("02 Jan 2006 15:04"),
	}

	if visit.ScheduledTime != nil {
		data.ScheduledTime = *visit.ScheduledTime
	}
	if visit.CheckInAt != nil {
		data.CheckInAt = *visit.CheckInAt
	}
	if visit.CheckOutAt != nil {
		data.CheckOutAt = *visit.CheckOutAt
	}

	// Employee
	if visit.Employee != nil {
		data.EmployeeName = visit.Employee.Name
		data.EmployeeCode = visit.Employee.EmployeeCode
	}

	// Customer & contact
	if visit.Customer != nil {
		data.CustomerName = visit.Customer.Name
	}
	data.ContactPerson = visit.ContactPerson
	data.ContactPhone = visit.ContactPhone
	data.Address = visit.Address

	// Deal / Lead
	if visit.Deal != nil {
		data.DealCode = visit.Deal.Code
		data.DealTitle = visit.Deal.Title
	}
	if visit.Lead != nil {
		data.LeadCode = visit.Lead.Code
		data.LeadName = fmt.Sprintf("%s %s", visit.Lead.FirstName, visit.Lead.LastName)
	}

	// Visit content
	data.Purpose = visit.Purpose
	data.Notes = visit.Notes
	data.Result = visit.Result
	data.Outcome = visit.Outcome
	data.NextSteps = visit.NextSteps

	// Product interest details
	for _, d := range visit.Details {
		item := visitPDFProduct{
			InterestLevel: d.InterestLevel,
			Notes:         d.Notes,
		}
		if d.Product != nil {
			item.ProductName = d.Product.Name
			item.ProductCode = d.Product.Code
		}
		if d.Quantity != nil {
			item.Quantity = *d.Quantity
		}
		if d.Price != nil {
			item.Price = *d.Price
		}
		data.Products = append(data.Products, item)
	}

	// Build PDF
	pdf := buildVisitReportPDF(data)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	filename := fmt.Sprintf("Visit_Report_%s.pdf", data.Code)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "application/pdf", buf.Bytes())
}

func buildVisitReportPDF(d visitPDFData) *fpdf.Fpdf {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(vrML, vrMT, vrMR)
	pdf.SetAutoPageBreak(true, 14)
	pdf.AddPage()

	vrDrawHeader(pdf, d)
	vrDrawMeta(pdf, d)
	vrDrawContactSection(pdf, d)

	if len(d.Products) > 0 {
		vrDrawProducts(pdf, d)
	}

	vrDrawVisitContent(pdf, d)
	vrDrawFooter(pdf, d)

	return pdf
}

// vrDrawHeader renders the report title and status badge
func vrDrawHeader(pdf *fpdf.Fpdf, d visitPDFData) {
	// Accent bar
	pdf.SetFillColor(vrAccentR, vrAccentG, vrAccentB)
	pdf.Rect(vrML, vrMT, vrCW, 1.5, "F")
	pdf.Ln(6)

	// Title
	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetTextColor(vrAccentR, vrAccentG, vrAccentB)
	pdf.CellFormat(vrCW*0.7, 8, "VISIT REPORT", "", 0, "L", false, 0, "")

	// Status badge
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFillColor(vrAccentR, vrAccentG, vrAccentB)
	statusW := pdf.GetStringWidth(d.Status) + 8
	pdf.CellFormat(statusW, 6, d.Status, "", 0, "C", true, 0, "")
	pdf.Ln(4)

	// Visit code
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(vrCW, 6, d.Code, "", 1, "L", false, 0, "")
	pdf.Ln(4)
}

// vrDrawMeta renders the visit logistics info
func vrDrawMeta(pdf *fpdf.Fpdf, d visitPDFData) {
	vrSectionTitle(pdf, "Visit Information")
	colW := vrCW / 2

	rows := []struct{ left, right [2]string }{
		{[2]string{"Visit Date", d.VisitDate}, [2]string{"Scheduled Time", d.ScheduledTime}},
		{[2]string{"Check In", vrFormatDateTime(d.CheckInAt)}, [2]string{"Check Out", vrFormatDateTime(d.CheckOutAt)}},
		{[2]string{"Employee", d.EmployeeName}, [2]string{"Employee Code", d.EmployeeCode}},
	}

	if d.Outcome != "" {
		outcomeLabel := strings.ReplaceAll(d.Outcome, "_", " ")
		if len(outcomeLabel) > 0 {
			outcomeLabel = strings.ToUpper(outcomeLabel[:1]) + outcomeLabel[1:]
		}
		rows = append(rows, struct{ left, right [2]string }{
			[2]string{"Outcome", outcomeLabel},
			[2]string{"", ""},
		})
	}

	for _, row := range rows {
		if row.left[1] == "" && row.right[1] == "" {
			continue
		}
		vrDrawMetaRow(pdf, colW, row.left[0], row.left[1], row.right[0], row.right[1])
	}
	pdf.Ln(4)
}

// vrDrawContactSection renders customer and contact details
func vrDrawContactSection(pdf *fpdf.Fpdf, d visitPDFData) {
	if d.CustomerName == "" && d.ContactPerson == "" && d.Address == "" {
		return
	}
	vrSectionTitle(pdf, "Contact Information")
	colW := vrCW / 2

	rows := []struct{ left, right [2]string }{
		{[2]string{"Customer", d.CustomerName}, [2]string{"Contact Person", d.ContactPerson}},
		{[2]string{"Phone", d.ContactPhone}, [2]string{"Address", d.Address}},
	}
	if d.DealCode != "" {
		rows = append(rows, struct{ left, right [2]string }{
			[2]string{"Deal", fmt.Sprintf("%s - %s", d.DealCode, d.DealTitle)},
			[2]string{"Lead", fmt.Sprintf("%s - %s", d.LeadCode, d.LeadName)},
		})
	}

	for _, row := range rows {
		vrDrawMetaRow(pdf, colW, row.left[0], row.left[1], row.right[0], row.right[1])
	}
	pdf.Ln(4)
}

// vrDrawProducts renders the product interest table
func vrDrawProducts(pdf *fpdf.Fpdf, d visitPDFData) {
	vrSectionTitle(pdf, "Product Interest")

	// Table header
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetFillColor(vrAccentR, vrAccentG, vrAccentB)
	pdf.SetTextColor(255, 255, 255)

	colWidths := []float64{10, 50, 30, 22, 22, 40}
	headers := []string{"No", "Product", "Code", "Interest", "Qty", "Notes"}

	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 7, h, "", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(50, 50, 50)

	for i, p := range d.Products {
		if pdf.GetY() > 260 {
			pdf.AddPage()
		}

		fill := i%2 == 0
		if fill {
			pdf.SetFillColor(245, 245, 250)
		}

		pdf.CellFormat(colWidths[0], 6, fmt.Sprintf("%d", i+1), "", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[1], 6, vrTruncate(p.ProductName, 28), "", 0, "L", fill, 0, "")
		pdf.CellFormat(colWidths[2], 6, p.ProductCode, "", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[3], 6, fmt.Sprintf("%d/5", p.InterestLevel), "", 0, "C", fill, 0, "")
		pdf.CellFormat(colWidths[4], 6, vrFmtQty(p.Quantity), "", 0, "R", fill, 0, "")
		pdf.CellFormat(colWidths[5], 6, vrTruncate(p.Notes, 22), "", 0, "L", fill, 0, "")
		pdf.Ln(-1)
	}

	// Bottom border
	pdf.SetDrawColor(vrAccentR, vrAccentG, vrAccentB)
	pdf.Line(vrML, pdf.GetY(), vrRE, pdf.GetY())
	pdf.Ln(4)
}

// vrDrawVisitContent renders purpose, notes, result, next steps
func vrDrawVisitContent(pdf *fpdf.Fpdf, d visitPDFData) {
	sections := []struct {
		title   string
		content string
	}{
		{"Purpose", d.Purpose},
		{"Notes", d.Notes},
		{"Result", d.Result},
		{"Next Steps", d.NextSteps},
	}

	hasContent := false
	for _, s := range sections {
		if s.content != "" {
			hasContent = true
			break
		}
	}
	if !hasContent {
		return
	}

	vrSectionTitle(pdf, "Visit Details")

	for _, s := range sections {
		if s.content == "" {
			continue
		}
		if pdf.GetY() > 260 {
			pdf.AddPage()
		}

		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(vrAccentR, vrAccentG, vrAccentB)
		pdf.CellFormat(vrCW, 5, s.title, "", 1, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		pdf.MultiCell(vrCW, 4.5, s.content, "", "L", false)
		pdf.Ln(3)
	}
}

// vrDrawFooter renders the printed date at the bottom
func vrDrawFooter(pdf *fpdf.Fpdf, d visitPDFData) {
	pdf.Ln(6)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(vrML, pdf.GetY(), vrRE, pdf.GetY())
	pdf.Ln(3)

	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(vrCW, 4, fmt.Sprintf("Printed on %s | GIMS - GILABS Integrated Management System", d.PrintDate), "", 0, "C", false, 0, "")
}

// --- Helpers ---

func vrSectionTitle(pdf *fpdf.Fpdf, title string) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(vrAccentR, vrAccentG, vrAccentB)
	pdf.CellFormat(vrCW, 7, title, "", 1, "L", false, 0, "")

	pdf.SetDrawColor(vrAccentR, vrAccentG, vrAccentB)
	pdf.Line(vrML, pdf.GetY(), vrRE, pdf.GetY())
	pdf.Ln(3)
}

func vrDrawMetaRow(pdf *fpdf.Fpdf, colW float64, leftLabel, leftVal, rightLabel, rightVal string) {
	y := pdf.GetY()

	// Left column
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(130, 130, 130)
	pdf.SetXY(vrML, y)
	pdf.CellFormat(colW, 4, leftLabel, "", 0, "L", false, 0, "")
	pdf.SetXY(vrML, y+4)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(colW, 5, leftVal, "", 0, "L", false, 0, "")

	// Right column
	if rightLabel != "" {
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(130, 130, 130)
		pdf.SetXY(vrML+colW, y)
		pdf.CellFormat(colW, 4, rightLabel, "", 0, "L", false, 0, "")
		pdf.SetXY(vrML+colW, y+4)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		pdf.CellFormat(colW, 5, rightVal, "", 0, "L", false, 0, "")
	}

	pdf.SetY(y + 12)
}

func vrFormatDateTime(dt string) string {
	if dt == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, dt)
	if err != nil {
		return dt
	}
	return t.Format("02 Jan 2006 15:04")
}

func vrTruncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func vrFmtQty(q float64) string {
	if q == 0 {
		return "-"
	}
	if q == float64(int(q)) {
		return fmt.Sprintf("%d", int(q))
	}
	return fmt.Sprintf("%.2f", q)
}
