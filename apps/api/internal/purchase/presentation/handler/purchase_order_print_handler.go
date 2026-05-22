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

// PurchaseOrderPrintHandler generates PDF print documents for purchase orders.
type PurchaseOrderPrintHandler struct {
	uc usecase.PurchaseOrderUsecase
	db *gorm.DB
}

// NewPurchaseOrderPrintHandler creates a new print handler instance.
func NewPurchaseOrderPrintHandler(uc usecase.PurchaseOrderUsecase, db *gorm.DB) *PurchaseOrderPrintHandler {
	return &PurchaseOrderPrintHandler{uc: uc, db: db}
}

type poPDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	Code         string
	Status       string
	OrderDate    string
	DueDate      string
	SupplierName string
	PaymentTerms string
	BusinessUnit string
	Notes        string

	SubTotal     float64
	TaxRate      float64
	TaxAmount    float64
	DeliveryCost float64
	OtherCost    float64
	TotalAmount  float64

	Items     []poPDFItem
	PrintDate string
}

type poPDFItem struct {
	ProductCode string
	ProductName string
	Qty         float64
	Price       float64
	Discount    float64
	Subtotal    float64
}

// PrintPurchaseOrder generates and streams a PDF for the given purchase order.
func (h *PurchaseOrderPrintHandler) PrintPurchaseOrder(c *gin.Context) {
	id := c.Param("id")

	po, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrPurchaseOrderNotFound {
			c.String(http.StatusNotFound, "Purchase order not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to load purchase order")
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

	data := buildPOPDFData(po, &company)

	pdfBytes, err := buildPurchaseOrderPDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", "po-"+data.Code+".pdf"))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func buildPOPDFData(po *dto.PurchaseOrderDetailResponse, company *orgModels.Company) *poPDFData {
	dueDate := ""
	if po.DueDate != nil {
		dueDate = *po.DueDate
	}

	items := make([]poPDFItem, 0, len(po.Items))
	for _, it := range po.Items {
		items = append(items, poPDFItem{
			ProductCode: extractIfaceCode(it.Product),
			ProductName: extractIfaceName(it.Product),
			Qty:         it.Quantity,
			Price:       it.Price,
			Discount:    it.Discount,
			Subtotal:    it.Subtotal,
		})
	}

	return &poPDFData{
		CompanyName:    company.Name,
		CompanyAddress: company.Address,
		CompanyPhone:   company.Phone,
		CompanyEmail:   company.Email,
		Code:           po.Code,
		Status:         strings.ToUpper(string(po.Status)),
		OrderDate:      po.OrderDate,
		DueDate:        dueDate,
		SupplierName:   extractIfaceName(po.Supplier),
		PaymentTerms:   extractIfaceName(po.PaymentTerms),
		BusinessUnit:   extractIfaceName(po.BusinessUnit),
		Notes:          po.Notes,
		SubTotal:       po.SubTotal,
		TaxRate:        po.TaxRate,
		TaxAmount:      po.TaxAmount,
		DeliveryCost:   po.DeliveryCost,
		OtherCost:      po.OtherCost,
		TotalAmount:    po.TotalAmount,
		Items:          items,
		PrintDate:      apptime.Now().Format("02 January 2006"),
	}
}

func buildPurchaseOrderPDF(d *poPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	pDrawDocHeader(pdf, "Purchase Order", d.CompanyName)

	startY := pdf.GetY()
	leftX, leftW := pdfML, 82.0
	rightX := pdfML + leftW + 10
	rightW := pdfCW - leftW - 10

	pDrawCompanyBlock(pdf, leftX, startY, leftW, d.CompanyAddress, d.CompanyPhone, d.CompanyEmail)
	leftEndY := pdf.GetY()

	pdf.SetXY(rightX, startY)
	pMetaRow(pdf, rightX, 36, rightW, "PO Number", d.Code)
	pMetaRow(pdf, rightX, 36, rightW, "Status", d.Status)
	pMetaRow(pdf, rightX, 36, rightW, "Order Date", d.OrderDate)
	if d.DueDate != "" {
		pMetaRow(pdf, rightX, 36, rightW, "Due Date", d.DueDate)
	}
	pMetaRow(pdf, rightX, 36, rightW, "Supplier", d.SupplierName)
	pMetaRow(pdf, rightX, 36, rightW, "Payment Terms", d.PaymentTerms)
	pMetaRow(pdf, rightX, 36, rightW, "Business Unit", d.BusinessUnit)
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
	pDrawTotalsBlock(pdf, subtotals, "TOTAL AMOUNT", pFmtMoney(d.TotalAmount))

	pdf.Ln(6)
	pDrawDocFooter(pdf, "THANK YOU FOR YOUR BUSINESS", d.CompanyName, d.CompanyPhone, d.CompanyEmail, d.PrintDate)

	return pdfToBytes(pdf)
}
