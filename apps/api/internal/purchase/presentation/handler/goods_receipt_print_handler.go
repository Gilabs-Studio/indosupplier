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

// GoodsReceiptPrintHandler generates PDF print documents for goods receipts.
type GoodsReceiptPrintHandler struct {
	uc usecase.GoodsReceiptUsecase
	db *gorm.DB
}

// NewGoodsReceiptPrintHandler creates a new print handler instance.
func NewGoodsReceiptPrintHandler(uc usecase.GoodsReceiptUsecase, db *gorm.DB) *GoodsReceiptPrintHandler {
	return &GoodsReceiptPrintHandler{uc: uc, db: db}
}

type grPDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	Code              string
	Status            string
	ReceiptDate       string
	PurchaseOrderCode string
	SupplierName      string
	Notes             string

	Items     []grPDFItem
	PrintDate string
}

type grPDFItem struct {
	ProductSKU  string
	ProductName string
	QtyReceived float64
	Notes       string
}

// PrintGoodsReceipt generates and streams a PDF for the given goods receipt.
func (h *GoodsReceiptPrintHandler) PrintGoodsReceipt(c *gin.Context) {
	id := c.Param("id")

	gr, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrGoodsReceiptNotFound {
			c.String(http.StatusNotFound, "Goods receipt not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to load goods receipt")
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

	data := buildGRPDFData(gr, &company)

	pdfBytes, err := buildGoodsReceiptPDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", "gr-"+data.Code+".pdf"))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func buildGRPDFData(gr *dto.GoodsReceiptDetailResponse, company *orgModels.Company) *grPDFData {
	poCode := ""
	if gr.PurchaseOrder != nil {
		poCode = gr.PurchaseOrder.Code
	}
	supplierName := ""
	if gr.Supplier != nil {
		supplierName = gr.Supplier.Name
	}

	items := make([]grPDFItem, 0, len(gr.Items))
	for _, it := range gr.Items {
		sku, name := "", ""
		if it.Product != nil {
			sku = pSafePtrStr(it.Product.SKU)
			name = it.Product.Name
		}
		items = append(items, grPDFItem{
			ProductSKU:  sku,
			ProductName: name,
			QtyReceived: it.QuantityReceived,
			Notes:       pSafePtrStr(it.Notes),
		})
	}

	return &grPDFData{
		CompanyName:       company.Name,
		CompanyAddress:    company.Address,
		CompanyPhone:      company.Phone,
		CompanyEmail:      company.Email,
		Code:              gr.Code,
		Status:            strings.ToUpper(gr.Status),
		ReceiptDate:       pSafePtrStr(gr.ReceiptDate),
		PurchaseOrderCode: poCode,
		SupplierName:      supplierName,
		Notes:             pSafePtrStr(gr.Notes),
		Items:             items,
		PrintDate:         apptime.Now().Format("02 January 2006"),
	}
}

func buildGoodsReceiptPDF(d *grPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	pDrawDocHeader(pdf, "Goods Receipt", d.CompanyName)

	startY := pdf.GetY()
	leftX, leftW := pdfML, 82.0
	rightX := pdfML + leftW + 10
	rightW := pdfCW - leftW - 10

	pDrawCompanyBlock(pdf, leftX, startY, leftW, d.CompanyAddress, d.CompanyPhone, d.CompanyEmail)
	leftEndY := pdf.GetY()

	pdf.SetXY(rightX, startY)
	pMetaRow(pdf, rightX, 36, rightW, "GR Number", d.Code)
	pMetaRow(pdf, rightX, 36, rightW, "Status", d.Status)
	pMetaRow(pdf, rightX, 36, rightW, "Receipt Date", d.ReceiptDate)
	pMetaRow(pdf, rightX, 36, rightW, "PO Reference", d.PurchaseOrderCode)
	pMetaRow(pdf, rightX, 36, rightW, "Supplier", d.SupplierName)
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

	// GR item columns: SKU(25) + Name(90) + Qty(20) + Notes(39) = 174
	colW := []float64{25, 90, 20, 39}
	pDrawItemsTableHeader(pdf,
		[]string{"SKU", "PRODUCT", "QTY RECEIVED", "NOTES"},
		colW,
		[]string{"L", "L", "R", "L"},
	)

	rowH := 6.0
	for _, it := range d.Items {
		pdf.SetTextColor(51, 51, 51)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(colW[0], rowH, it.ProductSKU, "", 0, "L", false, 0, "")
		pdf.CellFormat(colW[1], rowH, it.ProductName, "", 0, "L", false, 0, "")
		pdf.CellFormat(colW[2], rowH, pFmtQty(it.QtyReceived), "", 0, "R", false, 0, "")
		pdf.CellFormat(colW[3], rowH, it.Notes, "", 1, "L", false, 0, "")
		pdf.SetDrawColor(221, 221, 221)
		pdf.SetLineWidth(0.1)
		pdf.Line(pdfML, pdf.GetY(), pdfRE, pdf.GetY())
	}
	pdf.Ln(8)

	pDrawDocFooter(pdf, "GOODS RECEIVED — THANK YOU", d.CompanyName, d.CompanyPhone, d.CompanyEmail, d.PrintDate)

	return pdfToBytes(pdf)
}
