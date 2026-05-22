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

// PurchasePaymentPrintHandler generates PDF print documents for purchase payments.
type PurchasePaymentPrintHandler struct {
	uc usecase.PurchasePaymentUsecase
	db *gorm.DB
}

// NewPurchasePaymentPrintHandler creates a new print handler instance.
func NewPurchasePaymentPrintHandler(uc usecase.PurchasePaymentUsecase, db *gorm.DB) *PurchasePaymentPrintHandler {
	return &PurchasePaymentPrintHandler{uc: uc, db: db}
}

type ppPDFData struct {
	CompanyName    string
	CompanyAddress string
	CompanyPhone   string
	CompanyEmail   string

	PaymentID       string
	PaymentDate     string
	Amount          float64
	Method          string
	Status          string
	ReferenceNumber string
	Notes           string

	InvoiceCode    string
	InvoiceNumber  string
	InvoiceDate    string
	InvoiceDueDate string
	InvoiceAmount  float64

	BankAccountName   string
	BankAccountNumber string
	AccountHolder     string
	Currency          string

	PrintDate string
}

// PrintPurchasePayment generates and streams a PDF for the given purchase payment.
func (h *PurchasePaymentPrintHandler) PrintPurchasePayment(c *gin.Context) {
	id := c.Param("id")

	payment, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrPurchasePaymentNotFound {
			c.String(http.StatusNotFound, "Purchase payment not found")
			return
		}
		c.String(http.StatusInternalServerError, "Failed to load purchase payment")
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

	data := buildPPPDFData(payment, &company)

	pdfBytes, err := buildPurchasePaymentPDF(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate PDF")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", "payment-"+data.PaymentID+".pdf"))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func buildPPPDFData(payment *dto.PurchasePaymentDetailResponse, company *orgModels.Company) *ppPDFData {
	invoiceCode, invoiceNumber, invoiceDate, invoiceDueDate := "", "", "", ""
	var invoiceAmount float64
	if payment.Invoice != nil {
		invoiceCode = payment.Invoice.Code
		invoiceNumber = payment.Invoice.InvoiceNumber
		invoiceDate = payment.Invoice.InvoiceDate
		invoiceDueDate = payment.Invoice.DueDate
		invoiceAmount = payment.Invoice.Amount
	}

	bankName, bankNumber, accountHolder, currency := "", "", "", ""
	if payment.BankAccount != nil {
		bankName = payment.BankAccount.Name
		bankNumber = payment.BankAccount.AccountNumber
		accountHolder = payment.BankAccount.AccountHolder
		currency = payment.BankAccount.Currency
	}

	notes := ""
	if payment.Notes != nil {
		notes = *payment.Notes
	}

	return &ppPDFData{
		CompanyName:       company.Name,
		CompanyAddress:    company.Address,
		CompanyPhone:      company.Phone,
		CompanyEmail:      company.Email,
		PaymentID:         payment.ID,
		PaymentDate:       payment.PaymentDate,
		Amount:            payment.Amount,
		Method:            strings.ToUpper(payment.Method),
		Status:            strings.ToUpper(payment.Status),
		ReferenceNumber:   pSafePtrStr(payment.ReferenceNumber),
		Notes:             notes,
		InvoiceCode:       invoiceCode,
		InvoiceNumber:     invoiceNumber,
		InvoiceDate:       invoiceDate,
		InvoiceDueDate:    invoiceDueDate,
		InvoiceAmount:     invoiceAmount,
		BankAccountName:   bankName,
		BankAccountNumber: bankNumber,
		AccountHolder:     accountHolder,
		Currency:          currency,
		PrintDate:         apptime.Now().Format("02 January 2006"),
	}
}

func buildPurchasePaymentPDF(d *ppPDFData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfML, pdfMT, pdfMR)
	pdf.SetAutoPageBreak(true, pdfMB)
	pdf.AddPage()

	pDrawDocHeader(pdf, "Payment Receipt", d.CompanyName)

	startY := pdf.GetY()
	leftX, leftW := pdfML, 82.0
	rightX := pdfML + leftW + 10
	rightW := pdfCW - leftW - 10

	pDrawCompanyBlock(pdf, leftX, startY, leftW, d.CompanyAddress, d.CompanyPhone, d.CompanyEmail)
	leftEndY := pdf.GetY()

	pdf.SetXY(rightX, startY)
	pMetaRow(pdf, rightX, 40, rightW, "Payment ID", d.PaymentID)
	pMetaRow(pdf, rightX, 40, rightW, "Payment Date", d.PaymentDate)
	pMetaRow(pdf, rightX, 40, rightW, "Method", d.Method)
	pMetaRow(pdf, rightX, 40, rightW, "Status", d.Status)
	if d.ReferenceNumber != "" {
		pMetaRow(pdf, rightX, 40, rightW, "Reference No", d.ReferenceNumber)
	}
	rightEndY := pdf.GetY()

	pdf.SetY(max(leftEndY, rightEndY) + 6)

	// — Payment amount block
	pdf.SetFillColor(74, 94, 138)
	pdf.Rect(pdfML, pdf.GetY(), pdfCW, 0.5, "F")
	pdf.Ln(6)

	pDrawTotalsBlock(pdf, nil, "PAYMENT AMOUNT", pFmtMoney(d.Amount))

	// — Invoice reference section
	if d.InvoiceCode != "" {
		pdf.Ln(4)
		pSectionLabel(pdf, "INVOICE REFERENCE")
		pdf.SetFillColor(74, 94, 138)
		pdf.Rect(pdfML, pdf.GetY(), pdfCW, 0.5, "F")
		pdf.Ln(4)

		leftW2 := 82.0
		rightX2 := pdfML + leftW2 + 10
		rightW2 := pdfCW - leftW2 - 10

		pMetaRow(pdf, pdfML, 36, leftW2, "Invoice Code", d.InvoiceCode)
		pdf.SetXY(rightX2, pdf.GetY()-5)
		pMetaRow(pdf, rightX2, 40, rightW2, "Invoice Amount", pFmtMoney(d.InvoiceAmount))

		saveY := pdf.GetY()
		pdf.SetXY(pdfML, saveY)
		pMetaRow(pdf, pdfML, 36, leftW2, "Invoice No", d.InvoiceNumber)
		pdf.SetXY(rightX2, pdf.GetY()-5)
		pMetaRow(pdf, rightX2, 40, rightW2, "Invoice Date", d.InvoiceDate)

		saveY = pdf.GetY()
		pdf.SetXY(pdfML, saveY)
		pMetaRow(pdf, pdfML, 36, leftW2, "Due Date", d.InvoiceDueDate)
		pdf.Ln(2)
	}

	// — Bank account section
	if d.BankAccountName != "" {
		pdf.Ln(4)
		pSectionLabel(pdf, "BANK ACCOUNT")
		pdf.SetFillColor(74, 94, 138)
		pdf.Rect(pdfML, pdf.GetY(), pdfCW, 0.5, "F")
		pdf.Ln(4)

		leftW2 := 82.0
		rightX2 := pdfML + leftW2 + 10
		rightW2 := pdfCW - leftW2 - 10

		pMetaRow(pdf, pdfML, 36, leftW2, "Bank Name", d.BankAccountName)
		pdf.SetXY(rightX2, pdf.GetY()-5)
		pMetaRow(pdf, rightX2, 40, rightW2, "Account No", d.BankAccountNumber)

		saveY := pdf.GetY()
		pdf.SetXY(pdfML, saveY)
		pMetaRow(pdf, pdfML, 36, leftW2, "Account Holder", d.AccountHolder)
		pdf.SetXY(rightX2, pdf.GetY()-5)
		currencyStr := d.Currency
		if currencyStr == "" {
			currencyStr = "-"
		}
		pMetaRow(pdf, rightX2, 40, rightW2, "Currency", currencyStr)
		pdf.Ln(2)
	}

	// — Notes
	if d.Notes != "" {
		pdf.Ln(4)
		pSectionLabel(pdf, "NOTES")
		pdf.SetFont("Helvetica", "", 8.5)
		pdf.SetTextColor(85, 85, 85)
		pdf.MultiCell(pdfCW, 4, d.Notes, "", "L", false)
	}

	pdf.Ln(8)
	pDrawDocFooter(pdf, "THANK YOU FOR YOUR PAYMENT", d.CompanyName, d.CompanyPhone, d.CompanyEmail, d.PrintDate)

	return pdfToBytes(pdf)
}
