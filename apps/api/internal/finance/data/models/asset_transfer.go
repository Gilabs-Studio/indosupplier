package models

import (
	"time"

	"github.com/google/uuid"
)

// TransferStatus enumeration
type TransferStatus string

const (
	TransferStatusRequested             TransferStatus = "requested"
	TransferStatusPendingDepartmentHead TransferStatus = "pending_department_head"
	TransferStatusPendingFinance        TransferStatus = "pending_finance_controller"
	TransferStatusApproved              TransferStatus = "approved"
	TransferStatusRejected              TransferStatus = "rejected"
)

// AssetTransfer model
type AssetTransfer struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	TenantID            *uuid.UUID `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	AssetID             uuid.UUID  `gorm:"type:uuid;not null;index" json:"asset_id"`
	SourceTransactionID *uuid.UUID `gorm:"type:uuid;index" json:"source_transaction_id,omitempty"`

	FromLocationID   *uuid.UUID `gorm:"type:uuid" json:"from_location_id"`
	ToLocationID     *uuid.UUID `gorm:"type:uuid" json:"to_location_id"`
	FromCompanyID    *uuid.UUID `gorm:"type:uuid" json:"from_company_id"`
	ToCompanyID      *uuid.UUID `gorm:"type:uuid" json:"to_company_id"`
	FromDepartmentID *uuid.UUID `gorm:"type:uuid" json:"from_department_id"`
	ToDepartmentID   *uuid.UUID `gorm:"type:uuid" json:"to_department_id"`
	FromCustodianID  *uuid.UUID `gorm:"type:uuid" json:"from_custodian_id"`
	ToCustodianID    *uuid.UUID `gorm:"type:uuid" json:"to_custodian_id"`
	FromEmployeeID   *uuid.UUID `gorm:"type:uuid" json:"from_employee_id,omitempty"`
	ToEmployeeID     *uuid.UUID `gorm:"type:uuid" json:"to_employee_id,omitempty"`

	TransferDate time.Time `gorm:"type:date;not null;index" json:"transfer_date"`
	Reason       *string   `gorm:"type:text" json:"reason"`
	Notes        *string   `gorm:"type:text" json:"notes,omitempty"`

	Status              TransferStatus `gorm:"type:varchar(50);index" json:"status"`
	IsIntercompany      bool           `gorm:"type:boolean;default:false;index" json:"is_intercompany"`
	CurrentApprovalRole  string         `gorm:"type:varchar(50);index" json:"current_approval_role"`
	ApprovalStepIndex    int            `gorm:"type:integer;default:1" json:"approval_step_index"`
	ApprovalStepTotal    int            `gorm:"type:integer;default:1" json:"approval_step_total"`
	RequestedBy         *uuid.UUID     `gorm:"type:uuid;index" json:"requested_by,omitempty"`
	RequestedAt         time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"requested_at"`
	DhApprovedBy        *uuid.UUID     `gorm:"type:uuid" json:"dh_approved_by,omitempty"`
	DhApprovedAt        *time.Time     `gorm:"type:timestamp" json:"dh_approved_at,omitempty"`
	FcApprovedBy        *uuid.UUID     `gorm:"type:uuid" json:"fc_approved_by,omitempty"`
	FcApprovedAt        *time.Time     `gorm:"type:timestamp" json:"fc_approved_at,omitempty"`
	RejectedBy          *uuid.UUID     `gorm:"type:uuid" json:"rejected_by,omitempty"`
	RejectedAt          *time.Time     `gorm:"type:timestamp" json:"rejected_at,omitempty"`
	RejectionReason     *string        `gorm:"type:text" json:"rejection_reason,omitempty"`
	ApprovedBy          *uuid.UUID     `gorm:"type:uuid" json:"approved_by,omitempty"`
	ApprovedAt          *time.Time     `gorm:"type:timestamp" json:"approved_at,omitempty"`

	CreatedBy *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relations
	Asset            *Asset     `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
	RequestedByUser   *User      `gorm:"foreignKey:RequestedBy;references:ID" json:"requested_by_user,omitempty"`
	ToDepartment      *Department `gorm:"foreignKey:ToDepartmentID;references:ID" json:"to_department,omitempty"`
}

// TableName specifies the table name
func (AssetTransfer) TableName() string {
	return "asset_transfers"
}
