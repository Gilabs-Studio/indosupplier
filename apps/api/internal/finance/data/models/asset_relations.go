package models

// Placeholder models untuk relasi Asset
// File ini berisi type aliases untuk model yang ada di module lain
// untuk menghindari circular imports

// Company represents a company entity (from organization module)
type Company struct {
	ID       string `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name     string `gorm:"type:varchar(200)" json:"name"`
	Code     string `gorm:"type:varchar(50)" json:"code,omitempty"`
}

func (Company) TableName() string {
	return "companies"
}

// BusinessUnit represents a business unit (from organization module)
type BusinessUnit struct {
	ID   string `gorm:"type:uuid;primary_key" json:"id"`
	Name string `gorm:"type:varchar(200)" json:"name"`
	Code string `gorm:"type:varchar(50)" json:"code,omitempty"`
}

func (BusinessUnit) TableName() string {
	return "business_units"
}

// Department represents a department (from organization module)
type Department struct {
	ID   string `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Name string `gorm:"type:varchar(200)" json:"name"`
	Code string `gorm:"type:varchar(50)" json:"code,omitempty"`
}

func (Department) TableName() string {
	return "divisions"
}

// Employee represents an employee (from organization module)
type Employee struct {
	ID           string `gorm:"type:uuid;primary_key" json:"id"`
	Name         string `gorm:"type:varchar(200)" json:"name"`
	EmployeeCode string `gorm:"type:varchar(50)" json:"employee_code,omitempty"`
}

func (Employee) TableName() string {
	return "employees"
}

// User represents a system user (from user module)
type User struct {
	ID    string `gorm:"type:uuid;primary_key" json:"id"`
	Name  string `gorm:"type:varchar(200)" json:"name"`
	Email string `gorm:"type:varchar(200)" json:"email"`
}

func (User) TableName() string {
	return "users"
}

// Contact represents a contact/supplier (from purchase module)
type Contact struct {
	ID   string `gorm:"type:uuid;primary_key" json:"id"`
	Name string `gorm:"type:varchar(200)" json:"name"`
}

func (Contact) TableName() string {
	return "contacts"
}

// PurchaseOrder represents a purchase order (from purchase module)
type PurchaseOrder struct {
	ID       string `gorm:"type:uuid;primary_key" json:"id"`
	PONumber string `gorm:"type:varchar(50)" json:"po_number"`
	PODate   string `gorm:"type:date" json:"po_date,omitempty"`
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

// SupplierInvoice represents a supplier invoice (from purchase module)
type SupplierInvoice struct {
	ID            string `gorm:"type:uuid;primary_key" json:"id"`
	InvoiceNumber string `gorm:"type:varchar(50)" json:"invoice_number"`
	InvoiceDate   string `gorm:"type:date" json:"invoice_date,omitempty"`
}

func (SupplierInvoice) TableName() string {
	return "supplier_invoices"
}
