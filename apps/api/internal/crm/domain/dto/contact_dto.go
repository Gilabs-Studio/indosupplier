package dto

import "time"

// CreateContactRequest represents the request to create a contact
type CreateContactRequest struct {
	CustomerID    string  `json:"customer_id" binding:"required,uuid"`
	ContactRoleID *string `json:"contact_role_id" binding:"omitempty,uuid"`
	Name          string  `json:"name" binding:"required,min=2,max=200"`
	Phone         string  `json:"phone" binding:"omitempty,max=30"`
	Email         string  `json:"email" binding:"omitempty,email,max=100"`
	Notes         string  `json:"notes" binding:"max=1000"`
	IsActive      *bool   `json:"is_active"`
}

// UpdateContactRequest represents the request to update a contact
type UpdateContactRequest struct {
	CustomerID    string  `json:"customer_id" binding:"omitempty,uuid"`
	ContactRoleID *string `json:"contact_role_id" binding:"omitempty,uuid"`
	Name          string  `json:"name" binding:"omitempty,min=2,max=200"`
	Phone         string  `json:"phone" binding:"omitempty,max=30"`
	Email         string  `json:"email" binding:"omitempty,email,max=100"`
	Notes         string  `json:"notes" binding:"max=1000"`
	IsActive      *bool   `json:"is_active"`
}

// ContactResponse represents the response for a contact
type ContactResponse struct {
	ID            string               `json:"id"`
	CustomerID    string               `json:"customer_id"`
	Customer      *ContactCustomerInfo `json:"customer,omitempty"`
	ContactRoleID *string              `json:"contact_role_id"`
	ContactRole   *ContactRoleInfo     `json:"contact_role,omitempty"`
	Name          string               `json:"name"`
	Phone         string               `json:"phone"`
	Email         string               `json:"email"`
	Notes         string               `json:"notes"`
	IsActive      bool                 `json:"is_active"`
	CreatedBy     *string              `json:"created_by"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// ContactCustomerInfo contains minimal customer info for contact response
type ContactCustomerInfo struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// ContactRoleInfo contains minimal contact role info for contact response
type ContactRoleInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	BadgeColor string `json:"badge_color"`
}

// ContactFormDataResponse returns data needed for contact forms
type ContactFormDataResponse struct {
	Customers    []ContactCustomerOption    `json:"customers"`
	ContactRoles []ContactRoleOptionForForm `json:"contact_roles"`
}

// ContactCustomerOption represents a customer option for form dropdowns
type ContactCustomerOption struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// ContactRoleOptionForForm represents a contact role option for form dropdowns
type ContactRoleOptionForForm struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	BadgeColor string `json:"badge_color"`
}
