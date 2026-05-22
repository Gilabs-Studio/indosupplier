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

// SupplierInvoiceDPPrintHandler generates PDF documents for supplier invoice down payments.
type SupplierInvoiceDPPrintHandler struct {
	uc usecase.SupplierInvoiceDownPaymentUsecase
	db *gorm.DB
}

// NewSupplierInvoiceDPPrintHandler creates a new print handler instance.
func NewSupplierInvoiceDPPrintHandler(uc usecase.SupplierInvoiceDownPaymentUsecase, db *gorm.DB) *SupplierInvoiceDPPrintHandler {
	return &SupplierInvoiceDPPrintHandler{uc: uc, db: db}
}

type sidpPDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	Code              string
	InvoiceNumber     string
	Status            string
	InvoiceDate       string
	DueDate           string
	PurchaseOrderCode string
	Notes             string
	Amount            float64

	RegularInvoiceCodes []string
	PrintDate           string
}

// PrintSupplierInvoiceDP generates and streams a PDF for the given supplier invoice down payment.
func (h *SupplierInvoiceDPPrintHandler) PrintSupplierInvoiceDP(c *gin.Context) {
	id := c.Param("id")

	sidp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrSupplierInvoiceNotFound {
			c.String(http.StatusNotFound, "Supplier invoice down payment not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to load supplier invoice down payment")
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

	data := buildSIDPPDFData(sidp, &company)

	pdfBytes, err := buildSupplierInvoiceDPPDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", "supplier-invoice-dp-"+data.Code+".pdf"))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func buildSIDPPDFData(sidp *dto.SupplierInvoiceDownPaymentDetailResponse, company *orgModels.Company) *sidpPDFData {
	poCode := ""
	if sidp.PurchaseOrder != nil {
		poCode = sidp.PurchaseOrder.Code
	}
	notes := ""
	if sidp.Notes != nil {
		notes = *sidp.Notes
	}

	regularCodes := make([]string, 0, len(sidp.RegularInvoices))
	for _, ri := range sidp.RegularInvoices {
		regularCodes = append(regularCodes, ri.Code)
	}

	return &sidpPDFData{
		CompanyName:         company.Name,
		CompanyAddress:      company.Address,
		CompanyPhone:        company.Phone,
		CompanyEmail:        company.Email,
		Code:                sidp.Code,
		InvoiceNumber:       sidp.InvoiceNumber,
		Status:              strings.ToUpper(sidp.Status),
		InvoiceDate:         sidp.InvoiceDate,
		DueDate:             sidp.DueDate,
		PurchaseOrderCode:   poCode,
		Notes:               notes,
		Amount:              sidp.Amount,
		RegularInvoiceCodes: regularCodes,
		PrintDate:           apptime.Now().Format("02 January 2006"),
	}
}

func buildSupplierInvoiceDPPDF(d *sidpPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	pDrawDocHeader(pdf, "Down Payment Invoice", d.CompanyName)

	startY := pdf.GetY()
	leftX, leftW := pdfML, 82.0
	rightX := pdfML + leftW + 10
	rightW := pdfCW - leftW - 10

	pDrawCompanyBlock(pdf, leftX, startY, leftW, d.CompanyAddress, d.CompanyPhone, d.CompanyEmail)
	leftEndY := pdf.GetY()

	pdf.SetXY(rightX, startY)
	pMetaRow(pdf, rightX, 36, rightW, "Invoice No", d.InvoiceNumber)
	pMetaRow(pdf, rightX, 36, rightW, "Reference", d.Code)
	pMetaRow(pdf, rightX, 36, rightW, "Status", d.Status)
	pMetaRow(pdf, rightX, 36, rightW, "Invoice Date", d.InvoiceDate)
	pMetaRow(pdf, rightX, 36, rightW, "Due Date", d.DueDate)
	pMetaRow(pdf, rightX, 36, rightW, "PO Reference", d.PurchaseOrderCode)
	rightEndY := pdf.GetY()

	pdf.SetY(max(leftEndY, rightEndY) + 6)

	// Amount block
	pdf.SetFillColor(74, 94, 138)
	pdf.Rect(pdfML, pdf.GetY(), pdfCW, 0.5, "F")
	pdf.Ln(6)

	pDrawTotalsBlock(pdf, nil, "DOWN PAYMENT AMOUNT", pFmtMoney(d.Amount))

	// Linked regular invoices
	if len(d.RegularInvoiceCodes) > 0 {
		pdf.Ln(4)
		pSectionLabel(pdf, "LINKED REGULAR INVOICES")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(51, 51, 51)
		for _, code := range d.RegularInvoiceCodes {
			pdf.CellFormat(pdfCW, 5, "• "+code, "", 1, "L", false, 0, "")
		}
	}

	if d.Notes != "" {
		pdf.Ln(4)
		pSectionLabel(pdf, "NOTES")
		pdf.SetFont("Helvetica", "", 8.5)
		pdf.SetTextColor(85, 85, 85)
		pdf.MultiCell(pdfCW, 4, d.Notes, "", "L", false)
	}

	pdf.Ln(8)
	pDrawDocFooter(pdf, "THANK YOU FOR YOUR BUSINESS", d.CompanyName, d.CompanyPhone, d.CompanyEmail, d.PrintDate)

	return pdfToBytes(pdf)
}
