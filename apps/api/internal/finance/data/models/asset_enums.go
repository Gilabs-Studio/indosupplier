package models

import (
	"database/sql/driver"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AssetType enumeration
type AssetType string

const (
	AssetTypeTangible               AssetType = "TANGIBLE"
	AssetTypeIntangible             AssetType = "INTANGIBLE"
	AssetTypeNonDepreciable         AssetType = "NON_DEPRECIABLE"
	AssetTypeConstructionInProgress AssetType = "CONSTRUCTION"
)

// DisposalType enumeration
type DisposalType string

const (
	DisposalTypeSold        DisposalType = "SOLD"
	DisposalTypeScrapped    DisposalType = "SCRAPPED"
	DisposalTypeDonated     DisposalType = "DONATED"
	DisposalTypeTransferred DisposalType = "TRANSFERRED"
	DisposalTypeOther       DisposalType = "OTHER"
)

// AttachmentType enumeration
type AttachmentType string

const (
	AttachmentTypeInvoice  AttachmentType = "INVOICE"
	AttachmentTypeReceipt  AttachmentType = "RECEIPT"
	AttachmentTypePhoto    AttachmentType = "PHOTO"
	AttachmentTypeWarranty AttachmentType = "WARRANTY"
	AttachmentTypeManual   AttachmentType = "MANUAL"
	AttachmentTypeOther    AttachmentType = "OTHER"
)

// AuditAction enumeration
type AuditAction string

const (
	AuditActionCreated       AuditAction = "CREATED"
	AuditActionUpdated       AuditAction = "UPDATED"
	AuditActionActivated     AuditAction = "ACTIVATED"
	AuditActionDisposed      AuditAction = "DISPOSED"
	AuditActionTransferred   AuditAction = "TRANSFERRED"
	AuditActionStatusChanged AuditAction = "STATUS_CHANGED"
)

// Scan implements sql.Scanner interface for AssetType
func (at *AssetType) Scan(value interface{}) error {
	if value == nil {
		*at = ""
		return nil
	}
	bytes, _ := value.([]byte)
	*at = AssetType(strings.ToUpper(string(bytes)))
	return nil
}

// Value implements driver.Valuer interface for AssetType
func (at AssetType) Value() (driver.Value, error) {
	return string(at), nil
}

// helper empty uses to avoid unused import lint in files that might not reference all packages
var _ = time.Time{}
var _ = uuid.UUID{}
