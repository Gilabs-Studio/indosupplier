package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gilabs/gims/api/internal/core/apptime"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	orgModels "github.com/gilabs/gims/api/internal/organization/data/models"
	"github.com/gilabs/gims/api/internal/purchase/domain/dto"
	"github.com/gilabs/gims/api/internal/purchase/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"gorm.io/gorm"
)

// SupplierInvoicePrintHandler generates PDF print documents for supplier invoices.
type SupplierInvoicePrintHandler struct {
	uc usecase.SupplierInvoiceUsecase
	db *gorm.DB
}

// NewSupplierInvoicePrintHandler creates a new print handler instance.
func NewSupplierInvoicePrintHandler(uc usecase.SupplierInvoiceUsecase, db *gorm.DB) *SupplierInvoicePrintHandler {
	return &SupplierInvoicePrintHandler{uc: uc, db: db}
}

type siPDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	Code              string
	InvoiceNumber     string
	Type              string
	Status            string
	InvoiceDate       string
	DueDate           string
	PurchaseOrderCode string
	PaymentTermsName  string
	Notes             string

	SubTotal          float64
	TaxRate           float64
	TaxAmount         float64
	DeliveryCost      float64
	OtherCost         float64
	Amount            float64
	PaidAmount        float64
	RemainingAmount   float64
	DownPaymentAmount float64

	Items     []siPDFItem
	PrintDate string
}

type siPDFItem struct {
	ProductCode string
	ProductName string
	Qty         float64
	Price       float64
	Discount    float64
	Subtotal    float64
}

// PrintSupplierInvoice generates and streams a PDF for the given supplier invoice.
func (h *SupplierInvoicePrintHandler) PrintSupplierInvoice(c *gin.Context) {
	id := c.Param("id")

	si, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSupplierInvoiceNotFound {
			c.String(http.StatusNotFound, "Supplier invoice not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to load supplier invoice")
		return
	}

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

	data := buildSIPDFData(si, &company)

	pdfBytes, err := buildSupplierInvoicePDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", "supplier-invoice-"+data.Code+".pdf"))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func buildSIPDFData(si *dto.SupplierInvoiceDetailResponse, company *orgModels.Company) *siPDFData {
	poCode := ""
	if si.PurchaseOrder != nil {
		poCode = si.PurchaseOrder.Code
	}
	paymentTermsName := ""
	if si.PaymentTerms != nil {
		paymentTermsName = si.PaymentTerms.Name
	}
	notes := ""
	if si.Notes != nil {
		notes = *si.Notes
	}

	items := make([]siPDFItem, 0, len(si.Items))
	for _, it := range si.Items {
		items = append(items, siPDFItem{
			ProductCode: extractIfaceCode(it.Product),
			ProductName: extractIfaceName(it.Product),
			Qty:         it.Quantity,
			Price:       it.Price,
			Discount:    it.Discount,
			Subtotal:    it.SubTotal,
		})
	}

	return &siPDFData{
		CompanyName:       company.Name,
		CompanyAddress:    company.Address,
		CompanyPhone:      company.Phone,
		CompanyEmail:      company.Email,
		Code:              si.Code,
		InvoiceNumber:     si.InvoiceNumber,
		Type:              strings.ToUpper(si.Type),
		Status:            strings.ToUpper(si.Status),
		InvoiceDate:       si.InvoiceDate,
		DueDate:           si.DueDate,
		PurchaseOrderCode: poCode,
		PaymentTermsName:  paymentTermsName,
		Notes:             notes,
		SubTotal:          si.SubTotal,
		TaxRate:           si.TaxRate,
		TaxAmount:         si.TaxAmount,
		DeliveryCost:      si.DeliveryCost,
		OtherCost:         si.OtherCost,
		Amount:            si.Amount,
		PaidAmount:        si.PaidAmount,
		RemainingAmount:   si.RemainingAmount,
		DownPaymentAmount: si.DownPaymentAmount,
		Items:             items,
		PrintDate:         apptime.Now().Format("02 January 2006"),
	}
}

func buildSupplierInvoicePDF(d *siPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	pDrawDocHeader(pdf, "Supplier Invoice", d.CompanyName)

	startY := pdf.GetY()
	leftX, leftW := pdfML, 82.0
	rightX := pdfML + leftW + 10
	rightW := pdfCW - leftW - 10

	pDrawCompanyBlock(pdf, leftX, startY, leftW, d.CompanyAddress, d.CompanyPhone, d.CompanyEmail)
	leftEndY := pdf.GetY()

	pdf.SetXY(rightX, startY)
	pMetaRow(pdf, rightX, 36, rightW, "Invoice No", d.InvoiceNumber)
	pMetaRow(pdf, rightX, 36, rightW, "Reference", d.Code)
	pMetaRow(pdf, rightX, 36, rightW, "Type", d.Type)
	pMetaRow(pdf, rightX, 36, rightW, "Status", d.Status)
	pMetaRow(pdf, rightX, 36, rightW, "Invoice Date", d.InvoiceDate)
	pMetaRow(pdf, rightX, 36, rightW, "Due Date", d.DueDate)
	pMetaRow(pdf, rightX, 36, rightW, "PO Reference", d.PurchaseOrderCode)
	pMetaRow(pdf, rightX, 36, rightW, "Payment Terms", d.PaymentTermsName)
	rightEndY := pdf.GetY()

	pdf.SetY(max(leftEndY, rightEndY) + 6)

	if d.Notes != "" {
		pdf.SetFont("Helvetica", "", 8.5)
		pdf.SetTextColor(85, 85, 85)
		pdf.MultiCell(pdfCW, 4, "Notes: "+d.Notes, "", "L", false)
		pdf.Ln(2)
	}

	pdf.SetFillColor(74, 94, 138)
	pdf.Rect(pdfML, pdf.GetY(), pdfCW, 0.5, "F")
	pdf.Ln(4)

	colW := []float64{20, 70, 14, 28, 16, 26}
	pDrawItemsTableHeader(pdf,
		[]string{"CODE", "PRODUCT", "QTY", "UNIT PRICE", "DISC %", "SUBTOTAL"},
		colW,
		[]string{"L", "L", "R", "R", "R", "R"},
	)

	rowH := 6.0
	for _, it := range d.Items {
		pdf.SetTextColor(51, 51, 51)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(colW[0], rowH, it.ProductCode, "", 0, "L", false, 0, "")
		pdf.CellFormat(colW[1], rowH, it.ProductName, "", 0, "L", false, 0, "")
		pdf.CellFormat(colW[2], rowH, pFmtQty(it.Qty), "", 0, "R", false, 0, "")
		pdf.CellFormat(colW[3], rowH, pFmtMoney(it.Price), "", 0, "R", false, 0, "")
		pdf.CellFormat(colW[4], rowH, pFmtTax(it.Discount)+"%", "", 0, "R", false, 0, "")
		pdf.CellFormat(colW[5], rowH, pFmtMoney(it.Subtotal), "", 1, "R", false, 0, "")
		pdf.SetDrawColor(221, 221, 221)
		pdf.SetLineWidth(0.1)
		pdf.Line(pdfML, pdf.GetY(), pdfRE, pdf.GetY())
	}
	pdf.Ln(4)

	subtotals := [][2]string{
		{"Subtotal", pFmtMoney(d.SubTotal)},
		{fmt.Sprintf("Tax (%s%%)", pFmtTax(d.TaxRate)), pFmtMoney(d.TaxAmount)},
	}
	if d.DeliveryCost != 0 {
		subtotals = append(subtotals, [2]string{"Delivery Cost", pFmtMoney(d.DeliveryCost)})
	}
	if d.OtherCost != 0 {
		subtotals = append(subtotals, [2]string{"Other Costs", pFmtMoney(d.OtherCost)})
	}
	if d.DownPaymentAmount != 0 {
		subtotals = append(subtotals, [2]string{"Down Payment", "-" + pFmtMoney(d.DownPaymentAmount)})
	}
	pDrawTotalsBlock(pdf, subtotals, "TOTAL AMOUNT", pFmtMoney(d.Amount))

	// Payment summary
	if d.PaidAmount > 0 || d.RemainingAmount > 0 {
		pdf.Ln(2)
		tx := pdfRE - 80
		lw := 45.0
		vw := 35.0
		pdf.SetX(tx)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(85, 85, 85)
		pdf.CellFormat(lw, 5, "Paid Amount", "", 0, "L", false, 0, "")
		pdf.CellFormat(vw, 5, pFmtMoney(d.PaidAmount), "", 1, "R", false, 0, "")
		pdf.SetX(tx)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(74, 94, 138)
		pdf.CellFormat(lw, 5, "Remaining Amount", "", 0, "L", false, 0, "")
		pdf.CellFormat(vw, 5, pFmtMoney(d.RemainingAmount), "", 1, "R", false, 0, "")
	}

	pdf.Ln(6)
	pDrawDocFooter(pdf, "THANK YOU FOR YOUR BUSINESS", d.CompanyName, d.CompanyPhone, d.CompanyEmail, d.PrintDate)

	return pdfToBytes(pdf)
}
