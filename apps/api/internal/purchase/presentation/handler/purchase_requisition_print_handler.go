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

// PurchaseRequisitionPrintHandler generates PDF print documents for purchase requisitions.
type PurchaseRequisitionPrintHandler struct {
	uc usecase.PurchaseRequisitionUsecase
	db *gorm.DB
}

// NewPurchaseRequisitionPrintHandler creates a new print handler instance.
func NewPurchaseRequisitionPrintHandler(uc usecase.PurchaseRequisitionUsecase, db *gorm.DB) *PurchaseRequisitionPrintHandler {
	return &PurchaseRequisitionPrintHandler{uc: uc, db: db}
}

type prPDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	Code         string
	Status       string
	RequestDate  string
	SupplierName string
	PaymentTerms string
	BusinessUnit string
	EmployeeName string
	Address      string
	Notes        string

	Subtotal     float64
	TaxRate      float64
	TaxAmount    float64
	DeliveryCost float64
	OtherCost    float64
	TotalAmount  float64

	Items     []prPDFItem
	PrintDate string
}

type prPDFItem struct {
	ProductCode string
	ProductName string
	Qty         float64
	Price       float64
	Discount    float64
	Subtotal    float64
}

// PrintPurchaseRequisition generates and streams a PDF for the given purchase requisition.
func (h *PurchaseRequisitionPrintHandler) PrintPurchaseRequisition(c *gin.Context) {
	id := c.Param("id")

	pr, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrPurchaseRequisitionNotFound {
			c.String(http.StatusNotFound, "Purchase requisition not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to load purchase requisition")
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

	data := buildPRPDFData(pr, &company)

	pdfBytes, err := buildPurchaseRequisitionPDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", "pr-"+data.Code+".pdf"))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func buildPRPDFData(pr *dto.PurchaseRequisitionDetailResponse, company *orgModels.Company) *prPDFData {
	supplierName := ""
	if pr.Supplier != nil {
		supplierName = pr.Supplier.Name
	}
	paymentTermsName := ""
	if pr.PaymentTerms != nil {
		paymentTermsName = pr.PaymentTerms.Name
	}
	businessUnitName := ""
	if pr.BusinessUnit != nil {
		businessUnitName = pr.BusinessUnit.Name
	}
	employeeName := ""
	if pr.Employee != nil {
		employeeName = pr.Employee.Name
	}

	items := make([]prPDFItem, 0, len(pr.Items))
	for _, it := range pr.Items {
		productCode, productName := "", ""
		if it.Product != nil {
			productCode = it.Product.Code
			productName = it.Product.Name
		}
		items = append(items, prPDFItem{
			ProductCode: productCode,
			ProductName: productName,
			Qty:         it.Quantity,
			Price:       it.PurchasePrice,
			Discount:    it.Discount,
			Subtotal:    it.Subtotal,
		})
	}

	return &prPDFData{
		CompanyName:    company.Name,
		CompanyAddress: company.Address,
		CompanyPhone:   company.Phone,
		CompanyEmail:   company.Email,
		Code:           pr.Code,
		Status:         strings.ToUpper(pr.Status),
		RequestDate:    pr.RequestDate,
		SupplierName:   supplierName,
		PaymentTerms:   paymentTermsName,
		BusinessUnit:   businessUnitName,
		EmployeeName:   employeeName,
		Address:        pSafePtrStr(pr.Address),
		Notes:          pr.Notes,
		Subtotal:       pr.Subtotal,
		TaxRate:        pr.TaxRate,
		TaxAmount:      pr.TaxAmount,
		DeliveryCost:   pr.DeliveryCost,
		OtherCost:      pr.OtherCost,
		TotalAmount:    pr.TotalAmount,
		Items:          items,
		PrintDate:      apptime.Now().Format("02 January 2006"),
	}
}

func buildPurchaseRequisitionPDF(d *prPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	pDrawDocHeader(pdf, "Purchase Requisition", d.CompanyName)

	startY := pdf.GetY()
	leftX, leftW := pdfML, 82.0
	rightX := pdfML + leftW + 10
	rightW := pdfCW - leftW - 10

	pDrawCompanyBlock(pdf, leftX, startY, leftW, d.CompanyAddress, d.CompanyPhone, d.CompanyEmail)
	leftEndY := pdf.GetY()

	pdf.SetXY(rightX, startY)
	pMetaRow(pdf, rightX, 36, rightW, "PR Number", d.Code)
	pMetaRow(pdf, rightX, 36, rightW, "Status", d.Status)
	pMetaRow(pdf, rightX, 36, rightW, "Request Date", d.RequestDate)
	pMetaRow(pdf, rightX, 36, rightW, "Supplier", d.SupplierName)
	pMetaRow(pdf, rightX, 36, rightW, "Payment Terms", d.PaymentTerms)
	pMetaRow(pdf, rightX, 36, rightW, "Business Unit", d.BusinessUnit)
	pMetaRow(pdf, rightX, 36, rightW, "Requested By", d.EmployeeName)
	if d.Address != "" {
		pMetaRow(pdf, rightX, 36, rightW, "Delivery Address", d.Address)
	}
	rightEndY := pdf.GetY()

	pdf.SetY(max(leftEndY, rightEndY) + 6)

	if d.Notes != "" {
		pdf.SetFont("Helvetica", "", 8.5)
		pdf.SetTextColor(85, 85, 85)
		pdf.MultiCell(pdfCW, 4, "Notes: "+d.Notes, "", "L", false)
		pdf.Ln(2)
	}

	// Blue accent separator
	pdf.SetFillColor(74, 94, 138)
	pdf.Rect(pdfML, pdf.GetY(), pdfCW, 0.5, "F")
	pdf.Ln(4)

	// Items table — column widths sum to pdfCW (174 mm)
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
		{"Subtotal", pFmtMoney(d.Subtotal)},
		{fmt.Sprintf("Tax (%s%%)", pFmtTax(d.TaxRate)), pFmtMoney(d.TaxAmount)},
	}
	if d.DeliveryCost != 0 {
		subtotals = append(subtotals, [2]string{"Delivery Cost", pFmtMoney(d.DeliveryCost)})
	}
	if d.OtherCost != 0 {
		subtotals = append(subtotals, [2]string{"Other Costs", pFmtMoney(d.OtherCost)})
	}
	pDrawTotalsBlock(pdf, subtotals, "TOTAL AMOUNT", pFmtMoney(d.TotalAmount))

	pdf.Ln(6)
	pDrawDocFooter(pdf, "THANK YOU FOR YOUR BUSINESS", d.CompanyName, d.CompanyPhone, d.CompanyEmail, d.PrintDate)

	return pdfToBytes(pdf)
}
