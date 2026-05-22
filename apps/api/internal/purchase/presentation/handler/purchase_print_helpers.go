package handler

import (
"encoding/json"
"fmt"
"strings"

"github.com/go-pdf/fpdf"
"golang.org/x/text/language"
"golang.org/x/text/message"
)

// A4 portrait layout constants (millimetres) — shared across all purchase print handlers.
const (
pdfML = 18.0  // left margin
pdfMT = 18.0  // top margin
pdfMR = 18.0  // right margin
pdfMB = 14.0  // bottom margin
pdfPW = 210.0 // page width
pdfCW = 174.0 // content width  (pdfPW - pdfML - pdfMR)
pdfRE = 192.0 // right edge     (pdfML + pdfCW)
)

var pMoneyFmt = message.NewPrinter(language.English)

// pFmtMoney formats a float64 as a 2-decimal number with thousand separators.
func pFmtMoney(v float64) string { return pMoneyFmt.Sprintf("%.2f", v) }

// pFmtQty formats a quantity trimming trailing decimal zeros up to 3 places.
func pFmtQty(v float64) string {
s := fmt.Sprintf("%.3f", v)
s = strings.TrimRight(s, "0")
s = strings.TrimRight(s, ".")
return s
}

// pFmtTax formats a tax / discount rate trimming trailing zeros up to 2 places.
func pFmtTax(v float64) string {
s := fmt.Sprintf("%.2f", v)
s = strings.TrimRight(s, "0")
s = strings.TrimRight(s, ".")
return s
}

// extractIfaceName reads the "name" key from an interface{} struct snapshot.
func extractIfaceName(v interface{}) string {
if v == nil {
return ""
}
b, err := json.Marshal(v)
if err != nil {
return ""
}
var m map[string]interface{}
if err := json.Unmarshal(b, &m); err != nil {
return ""
}
name, _ := m["name"].(string)
return name
}

// extractIfaceCode reads the "code" key from an interface{} struct snapshot.
func extractIfaceCode(v interface{}) string {
if v == nil {
return ""
}
b, _ := json.Marshal(v)
var m map[string]interface{}
json.Unmarshal(b, &m)
code, _ := m["code"].(string)
return code
}

// pSafePtrStr returns the dereferenced value of a *string, or "" if nil.
func pSafePtrStr(p *string) string {
if p == nil {
return ""
}
return *p
}

// ─── Shared PDF drawing helpers ───────────────────────────────────────────────
// All colours match the sales module exactly: accent blue (74,94,138).

// pDrawDocHeader renders the italic grey title (left), bold company name (right),
// and the blue accent bar — identical to the sales print handlers.
func pDrawDocHeader(pdf *fpdf.Fpdf, docTitle, companyName string) {
pdf.SetFont("Helvetica", "I", 22)
pdf.SetTextColor(122, 122, 122)
pdf.CellFormat(pdfCW/2, 10, docTitle, "", 0, "L", false, 0, "")

pdf.SetFont("Helvetica", "B", 14)
pdf.SetTextColor(51, 51, 51)
pdf.CellFormat(pdfCW/2, 10, strings.ToUpper(companyName), "", 1, "R", false, 0, "")

pdf.Ln(1)
pdf.SetFillColor(74, 94, 138)
pdf.Rect(pdfML, pdf.GetY(), pdfCW, 1.5, "F")
pdf.Ln(5)
}

// pDrawCompanyBlock renders the company address/phone/email block at position (x, y).
func pDrawCompanyBlock(pdf *fpdf.Fpdf, x, startY, w float64, address, phone, email string) {
pdf.SetXY(x, startY)
pdf.SetFont("Helvetica", "", 8.5)
pdf.SetTextColor(85, 85, 85)
if address != "" {
pdf.MultiCell(w, 4, address, "", "L", false)
}
if phone != "" {
pdf.SetX(x)
pdf.CellFormat(w, 4, "Tel: "+phone, "", 1, "L", false, 0, "")
}
if email != "" {
pdf.SetX(x)
pdf.CellFormat(w, 4, email, "", 1, "L", false, 0, "")
}
}

// pMetaRow renders a label : value pair with an aligned colon separator.
func pMetaRow(pdf *fpdf.Fpdf, x, labelW, blockW float64, label, value string) {
const lineH = 5.0
pdf.SetX(x)
pdf.SetFont("Helvetica", "", 9)
pdf.SetTextColor(102, 102, 102)
pdf.CellFormat(labelW, lineH, label, "", 0, "L", false, 0, "")
pdf.CellFormat(3, lineH, ":", "", 0, "L", false, 0, "")
pdf.SetFont("Helvetica", "B", 9)
pdf.SetTextColor(0, 0, 0)
pdf.CellFormat(blockW-labelW-3, lineH, value, "", 1, "R", false, 0, "")
}

// pSectionLabel renders a small bold uppercase section heading.
func pSectionLabel(pdf *fpdf.Fpdf, label string) {
pdf.SetFont("Helvetica", "B", 8.5)
pdf.SetTextColor(51, 51, 51)
pdf.CellFormat(0, 4, label, "", 1, "L", false, 0, "")
pdf.Ln(1)
}

// pDrawItemsTableHeader draws the blue table header row used in documents with line items.
func pDrawItemsTableHeader(pdf *fpdf.Fpdf, labels []string, widths []float64, aligns []string) {
hdrH := 7.0
pdf.SetFillColor(74, 94, 138)
pdf.SetTextColor(255, 255, 255)
pdf.SetFont("Helvetica", "B", 8)
for i, lbl := range labels {
align := "L"
if i < len(aligns) {
align = aligns[i]
}
pdf.CellFormat(widths[i], hdrH, lbl, "", 0, align, true, 0, "")
}
pdf.Ln(-1)
}

// pDrawTotalsBlock renders subtotal rows + a grand total with double blue border lines.
func pDrawTotalsBlock(pdf *fpdf.Fpdf, subtotals [][2]string, totalLabel, totalValue string) {
tx := pdfRE - 80
lw := 45.0
vw := 35.0
tw := lw + vw

if pdf.GetY()+50 > 283 {
pdf.AddPage()
}

separator := func() {
pdf.SetDrawColor(238, 238, 238)
pdf.SetLineWidth(0.2)
y := pdf.GetY()
pdf.Line(tx, y, tx+tw, y)
}

for _, r := range subtotals {
pdf.SetX(tx)
pdf.SetFont("Helvetica", "", 9)
pdf.SetTextColor(51, 51, 51)
pdf.CellFormat(lw, 5, r[0], "", 0, "L", false, 0, "")
pdf.CellFormat(vw, 5, r[1], "", 1, "R", false, 0, "")
separator()
}

pdf.Ln(1)
pdf.SetDrawColor(74, 94, 138)
pdf.SetLineWidth(0.4)
y := pdf.GetY()
pdf.Line(tx, y, tx+tw, y)
pdf.Ln(1)

pdf.SetX(tx)
pdf.SetFont("Helvetica", "B", 10)
pdf.SetTextColor(0, 0, 0)
pdf.CellFormat(lw, 7, totalLabel, "", 0, "L", false, 0, "")
pdf.CellFormat(vw, 7, totalValue, "", 1, "R", false, 0, "")

y = pdf.GetY()
pdf.Line(tx, y, tx+tw, y)
pdf.Line(tx, y+0.6, tx+tw, y+0.6)
pdf.SetLineWidth(0.2)
pdf.Ln(8)
}

// pDrawDocFooter renders the "THANK YOU" banner, grey separator, and generation timestamp.
func pDrawDocFooter(pdf *fpdf.Fpdf, thankMsg, companyName, phone, email, printDate string) {
parts := make([]string, 0, 2)
if phone != "" {
parts = append(parts, phone)
}
if email != "" {
parts = append(parts, email)
}
if len(parts) > 0 {
pdf.SetFont("Helvetica", "", 8)
pdf.SetTextColor(85, 85, 85)
pdf.MultiCell(pdfCW, 4, "If you have any questions, please contact us at "+strings.Join(parts, ", "), "", "C", false)
pdf.Ln(4)
}

pdf.SetFont("Helvetica", "B", 10)
pdf.SetTextColor(74, 94, 138)
pdf.CellFormat(pdfCW, 6, thankMsg, "", 1, "C", false, 0, "")
pdf.Ln(6)

pdf.SetDrawColor(204, 204, 204)
pdf.Line(pdfML, pdf.GetY(), pdfRE, pdf.GetY())
pdf.Ln(3)

pdf.SetFont("Helvetica", "", 7.5)
pdf.SetTextColor(153, 153, 153)
pdf.CellFormat(pdfCW, 4, fmt.Sprintf("This document was generated on %s  -  %s", printDate, companyName), "", 1, "C", false, 0, "")
}

// pdfToBytes serialises a completed fpdf document into a byte slice.
func pdfToBytes(pdf *fpdf.Fpdf) ([]byte, error) {
var buf pdfBytesWriter
if err := pdf.Output(&buf); err != nil {
return nil, err
}
return buf.buf, nil
}

// pdfBytesWriter collects PDF output bytes without allocating an intermediate string.
type pdfBytesWriter struct{ buf []byte }

func (w *pdfBytesWriter) Write(p []byte) (int, error) {
w.buf = append(w.buf, p...)
return len(p), nil
}
