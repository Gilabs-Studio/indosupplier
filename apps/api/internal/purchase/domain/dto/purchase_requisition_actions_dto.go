package dto

import "time"

type PurchaseRequisitionAddProduct struct {
	ID         string  `json:"id"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Stock      float64 `json:"stock"`
	CurrentHpp float64 `json:"current_hpp"`
	SupplierID *string `json:"supplier_id"`
	IsActive   bool    `json:"is_active"`
	IsApproved bool    `json:"is_approved"`
}

type PurchaseRequisitionAddSupplierContact struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Label       string `json:"label"`
	IsPrimary   bool   `json:"is_primary"`
}

type PurchaseRequisitionAddSupplier struct {
	ID             string                                  `json:"id"`
	Code           string                                  `json:"code"`
	Name           string                                  `json:"name"`
	PaymentTermsID *string                                 `json:"payment_terms_id"`
	BusinessUnitID *string                                 `json:"business_unit_id"`
	Contacts       []PurchaseRequisitionAddSupplierContact `json:"contacts"`
	Products       []PurchaseRequisitionAddProduct         `json:"products"`
}

type PurchaseRequisitionAddPaymentTerms struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Days int    `json:"days"`
}

type PurchaseRequisitionAddBusinessUnit struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PurchaseRequisitionAddEmployee struct {
	ID       string  `json:"id"`
	UserID   *string `json:"user_id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	IsActive bool    `json:"is_active"`
}

type PurchaseRequisitionAddResponse struct {
	Suppliers     []PurchaseRequisitionAddSupplier     `json:"suppliers"`
	PaymentTerms  []PurchaseRequisitionAddPaymentTerms `json:"payment_terms"`
	BusinessUnits []PurchaseRequisitionAddBusinessUnit `json:"business_units"`
	Employees     []PurchaseRequisitionAddEmployee     `json:"employees"`
}

type AuditTrailUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type PurchaseRequisitionAuditTrailEntry struct {
	ID             string                 `json:"id"`
	Action         string                 `json:"action"`
	PermissionCode string                 `json:"permission_code"`
	TargetID       string                 `json:"target_id"`
	Metadata       map[string]interface{} `json:"metadata"`
	User           *AuditTrailUser        `json:"user"`
	CreatedAt      time.Time              `json:"created_at"`
}
