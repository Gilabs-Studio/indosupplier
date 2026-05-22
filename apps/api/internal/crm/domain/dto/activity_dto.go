package dto

import "encoding/json"

// CreateActivityRequest represents the request to create an activity
type CreateActivityRequest struct {
	Type           string  `json:"type" binding:"required,min=1"`
	ActivityTypeID *string `json:"activity_type_id" binding:"omitempty,uuid"`
	EmployeeID     *string `json:"employee_id"`
	UserID         *string `json:"user_id"`
	CustomerID     *string `json:"customer_id" binding:"omitempty,uuid"`
	ContactID      *string `json:"contact_id" binding:"omitempty,uuid"`
	DealID         *string `json:"deal_id" binding:"omitempty,uuid"`
	LeadID         *string `json:"lead_id" binding:"omitempty,uuid"`
	VisitReportID  *string `json:"visit_report_id" binding:"omitempty,uuid"`
	Description    string          `json:"description" binding:"required,min=1"`
	Timestamp      *string         `json:"timestamp"` // ISO 8601 format, defaults to now if empty
	Metadata       json.RawMessage `json:"metadata"`
}

// ActivityResponse represents the activity data returned to client
type ActivityResponse struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	ActivityTypeID *string           `json:"activity_type_id"`
	ActivityType   *ActivityTypeInfo `json:"activity_type,omitempty"`
	CustomerID     *string           `json:"customer_id"`
	ContactID      *string           `json:"contact_id"`
	DealID         *string           `json:"deal_id"`
	LeadID         *string           `json:"lead_id"`
	VisitReportID  *string           `json:"visit_report_id"`
	EmployeeID     string            `json:"employee_id"`
	Employee       *ActivityEmployeeInfo `json:"employee,omitempty"`
	Description    string            `json:"description"`
	Timestamp      string            `json:"timestamp"`
	Metadata       json.RawMessage   `json:"metadata"`
	CreatedAt      string            `json:"created_at"`
}

// ActivityTypeInfo holds compact activity type info for responses
type ActivityTypeInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	Icon       string `json:"icon"`
	BadgeColor string `json:"badge_color"`
}

// ActivityEmployeeInfo holds compact employee info for activity responses
type ActivityEmployeeInfo struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}
