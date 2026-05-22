package dto

// CreateVisitReportRequest defines the request body for creating a visit report
type CreateVisitReportRequest struct {
	VisitDate     string                           `json:"visit_date" binding:"required"`
	ScheduledTime *string                          `json:"scheduled_time"`
	EmployeeID    string                           `json:"employee_id" binding:"required,uuid"`
	CustomerID    *string                          `json:"customer_id" binding:"omitempty,uuid"`
	ContactID     *string                          `json:"contact_id" binding:"omitempty,uuid"`
	DealID        *string                          `json:"deal_id" binding:"omitempty,uuid"`
	LeadID        *string                          `json:"lead_id" binding:"omitempty,uuid"`
	TravelPlanID  *string                          `json:"travel_plan_id" binding:"omitempty,uuid"`
	ContactPerson string                           `json:"contact_person" binding:"max=200"`
	ContactPhone  string                           `json:"contact_phone" binding:"max=20"`
	Address       string                           `json:"address"`
	VillageID     *string                          `json:"village_id" binding:"omitempty,uuid"`
	Purpose       string                           `json:"purpose"`
	Notes         string                           `json:"notes"`
	Details       []CreateVisitReportDetailRequest `json:"details"`
}

// CreateVisitReportDetailRequest defines a product interest item
type CreateVisitReportDetailRequest struct {
	ProductID     string                                   `json:"product_id" binding:"required,uuid"`
	InterestLevel int                                      `json:"interest_level" binding:"min=0,max=5"`
	Notes         string                                   `json:"notes"`
	Quantity      *float64                                 `json:"quantity" binding:"omitempty,gte=0"`
	Price         *float64                                 `json:"price" binding:"omitempty,gte=0"`
	Answers       []CreateVisitReportInterestAnswerRequest `json:"answers"`
}

// CreateVisitReportInterestAnswerRequest defines a survey answer
type CreateVisitReportInterestAnswerRequest struct {
	QuestionID string `json:"question_id" binding:"required,uuid"`
	OptionID   string `json:"option_id" binding:"required,uuid"`
}

// UpdateVisitReportRequest defines the request body for updating a visit report
type UpdateVisitReportRequest struct {
	VisitDate     *string                           `json:"visit_date"`
	ScheduledTime *string                           `json:"scheduled_time"`
	EmployeeID    *string                           `json:"employee_id" binding:"omitempty,uuid"`
	CustomerID    *string                           `json:"customer_id" binding:"omitempty,uuid"`
	ContactID     *string                           `json:"contact_id" binding:"omitempty,uuid"`
	DealID        *string                           `json:"deal_id" binding:"omitempty,uuid"`
	LeadID        *string                           `json:"lead_id" binding:"omitempty,uuid"`
	TravelPlanID  *string                           `json:"travel_plan_id" binding:"omitempty,uuid"`
	ContactPerson *string                           `json:"contact_person" binding:"omitempty,max=200"`
	ContactPhone  *string                           `json:"contact_phone" binding:"omitempty,max=20"`
	Address       *string                           `json:"address"`
	VillageID     *string                           `json:"village_id" binding:"omitempty,uuid"`
	Purpose       *string                           `json:"purpose"`
	Notes         *string                           `json:"notes"`
	Result        *string                           `json:"result"`
	Outcome       *string                           `json:"outcome" binding:"omitempty,oneof=positive neutral negative very_positive"`
	NextSteps     *string                           `json:"next_steps"`
	Details       *[]CreateVisitReportDetailRequest `json:"details"`
}

// ListVisitReportsRequest defines the query parameters for listing visit reports
type ListVisitReportsRequest struct {
	Page              int    `form:"page" binding:"omitempty,min=1"`
	PerPage           int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search            string `form:"search"`
	CustomerID        string `form:"customer_id" binding:"omitempty,uuid"`
	EmployeeID        string `form:"employee_id" binding:"omitempty,uuid"`
	ContactID         string `form:"contact_id" binding:"omitempty,uuid"`
	DealID            string `form:"deal_id" binding:"omitempty,uuid"`
	LeadID            string `form:"lead_id" binding:"omitempty,uuid"`
	TravelPlanID      string `form:"travel_plan_id" binding:"omitempty,uuid"`
	WithoutTravelPlan bool   `form:"without_travel_plan"`
	Outcome           string `form:"outcome" binding:"omitempty,oneof=positive neutral negative very_positive"`
	DateFrom          string `form:"date_from"`
	DateTo            string `form:"date_to"`
	SortBy            string `form:"sort_by"`
	SortDir           string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// CheckInRequest defines the GPS check-in request
type CheckInVisitRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Accuracy  *float64 `json:"accuracy"`
}

// CheckOutVisitRequest defines the GPS check-out request
type CheckOutVisitRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Accuracy  *float64 `json:"accuracy"`
	Result    string   `json:"result"`
	Outcome   string   `json:"outcome" binding:"omitempty,oneof=positive neutral negative very_positive"`
	NextSteps string   `json:"next_steps"`
}

// SubmitVisitReportRequest defines the submit for approval request
type SubmitVisitReportRequest struct {
	Notes string `json:"notes"`
}

// ApproveVisitReportRequest defines the approval request
type ApproveVisitReportRequest struct {
	Notes string `json:"notes"`
}

// RejectVisitReportRequest defines the rejection request
type RejectVisitReportRequest struct {
	Reason string `json:"reason" binding:"required,min=5"`
	Notes  string `json:"notes"`
}

// VisitReportResponse defines the response body for a visit report
type VisitReportResponse struct {
	ID               string                      `json:"id"`
	Code             string                      `json:"code"`
	CustomerID       *string                     `json:"customer_id"`
	Customer         *VisitCustomerBrief         `json:"customer,omitempty"`
	ContactID        *string                     `json:"contact_id"`
	Contact          *VisitContactBrief          `json:"contact,omitempty"`
	DealID           *string                     `json:"deal_id"`
	Deal             *VisitDealBrief             `json:"deal,omitempty"`
	LeadID           *string                     `json:"lead_id"`
	TravelPlanID     *string                     `json:"travel_plan_id"`
	Lead             *VisitLeadBrief             `json:"lead,omitempty"`
	EmployeeID       string                      `json:"employee_id"`
	Employee         *VisitEmployeeBrief         `json:"employee,omitempty"`
	VisitDate        string                      `json:"visit_date"`
	ScheduledTime    *string                     `json:"scheduled_time"`
	ActualTime       *string                     `json:"actual_time"`
	CheckInAt        *string                     `json:"check_in_at"`
	CheckOutAt       *string                     `json:"check_out_at"`
	CheckInLocation  *string                     `json:"check_in_location"`
	CheckOutLocation *string                     `json:"check_out_location"`
	Address          string                      `json:"address"`
	VillageID        *string                     `json:"village_id"`
	Village          *VisitVillageResponse       `json:"village,omitempty"`
	Latitude         *float64                    `json:"latitude"`
	Longitude        *float64                    `json:"longitude"`
	Purpose          string                      `json:"purpose"`
	Notes            string                      `json:"notes"`
	Result           string                      `json:"result"`
	Outcome          string                      `json:"outcome"`
	NextSteps        string                      `json:"next_steps"`
	ContactPerson    string                      `json:"contact_person"`
	ContactPhone     string                      `json:"contact_phone"`
	Photos           *string                     `json:"photos"`
	CreatedBy        *string                     `json:"created_by"`
	Details          []VisitReportDetailResponse `json:"details,omitempty"`
	CreatedAt        string                      `json:"created_at"`
	UpdatedAt        string                      `json:"updated_at"`
}

// VisitReportDetailResponse defines the response for a visit report detail
type VisitReportDetailResponse struct {
	ID            string                              `json:"id"`
	VisitReportID string                              `json:"visit_report_id"`
	ProductID     string                              `json:"product_id"`
	Product       *VisitProductBrief                  `json:"product,omitempty"`
	InterestLevel int                                 `json:"interest_level"`
	Notes         string                              `json:"notes"`
	Quantity      *float64                            `json:"quantity"`
	Price         *float64                            `json:"price"`
	Answers       []VisitReportInterestAnswerResponse `json:"answers,omitempty"`
	CreatedAt     string                              `json:"created_at"`
	UpdatedAt     string                              `json:"updated_at"`
}

// VisitReportInterestAnswerResponse defines the response for an interest answer
type VisitReportInterestAnswerResponse struct {
	ID           string `json:"id"`
	QuestionID   string `json:"question_id"`
	QuestionText string `json:"question_text"`
	OptionID     string `json:"option_id"`
	OptionText   string `json:"option_text"`
	Score        int    `json:"score"`
}

// VisitReportProgressHistoryResponse defines the response for progress history
type VisitReportProgressHistoryResponse struct {
	ID            string  `json:"id"`
	VisitReportID string  `json:"visit_report_id"`
	FromStatus    string  `json:"from_status"`
	ToStatus      string  `json:"to_status"`
	Notes         string  `json:"notes"`
	ChangedBy     *string `json:"changed_by"`
	CreatedAt     string  `json:"created_at"`
}

// VisitInterestQuestionResponse defines the response for an interest question
type VisitInterestQuestionResponse struct {
	ID           string                        `json:"id"`
	QuestionText string                        `json:"question_text"`
	Sequence     int                           `json:"sequence"`
	Options      []VisitInterestOptionResponse `json:"options"`
}

// VisitInterestOptionResponse defines the response for an interest option
type VisitInterestOptionResponse struct {
	ID         string `json:"id"`
	OptionText string `json:"option_text"`
	Score      int    `json:"score"`
}

// Brief responses for related entities

// VisitCustomerBrief is a brief customer representation
type VisitCustomerBrief struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

// VisitContactBrief is a brief contact representation
type VisitContactBrief struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// VisitDealBrief is a brief deal representation
type VisitDealBrief struct {
	ID    string `json:"id"`
	Code  string `json:"code"`
	Title string `json:"title"`
}

// VisitLeadBrief is a brief lead representation
type VisitLeadBrief struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// VisitEmployeeBrief is a brief employee representation
type VisitEmployeeBrief struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}

// VisitProductBrief is a brief product representation
type VisitProductBrief struct {
	ID           string  `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	SellingPrice float64 `json:"selling_price"`
	ImageURL     string  `json:"image_url"`
}

// VisitVillageResponse is the village with location hierarchy
type VisitVillageResponse struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	District *VisitDistrictResponse `json:"district,omitempty"`
}

// VisitDistrictResponse is the district representation
type VisitDistrictResponse struct {
	ID      string                `json:"id"`
	Name    string                `json:"name"`
	Regency *VisitRegencyResponse `json:"regency,omitempty"`
}

// VisitRegencyResponse is the regency/city representation
type VisitRegencyResponse struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Province *VisitProvinceResponse `json:"province,omitempty"`
}

// VisitProvinceResponse is the province representation
type VisitProvinceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// VisitReportFormDataResponse contains form data for creating/editing visit reports
type VisitReportFormDataResponse struct {
	Customers         []VisitFormDataCustomer         `json:"customers"`
	Contacts          []VisitFormDataContact          `json:"contacts"`
	Employees         []VisitFormDataEmployee         `json:"employees"`
	Deals             []VisitFormDataDeal             `json:"deals"`
	Leads             []VisitFormDataLead             `json:"leads"`
	Products          []VisitFormDataProduct          `json:"products"`
	Outcomes          []VisitFormDataOption           `json:"outcomes"`
	InterestQuestions []VisitInterestQuestionResponse `json:"interest_questions"`
}

// VisitFormDataCustomer is a customer option for forms
type VisitFormDataCustomer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// VisitFormDataContact is a contact option for forms
type VisitFormDataContact struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CustomerID string `json:"customer_id"`
}

// VisitFormDataEmployee is an employee option for forms
type VisitFormDataEmployee struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
}

// VisitFormDataDeal is a deal option for forms
type VisitFormDataDeal struct {
	ID    string `json:"id"`
	Code  string `json:"code"`
	Title string `json:"title"`
}

// VisitFormDataLead is a lead option for forms
type VisitFormDataLead struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// VisitFormDataProduct is a product option for forms
type VisitFormDataProduct struct {
	ID           string  `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	SellingPrice float64 `json:"selling_price"`
}

// VisitFormDataOption is a generic enum option for forms
type VisitFormDataOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ListByEmployeeRequest defines query parameters for listing visit report summaries grouped by employee
type ListByEmployeeRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search"`
}

// VisitReportStatusCounts contains per-status counts for aggregated views
type VisitReportStatusCounts struct {
	Draft     int64 `json:"draft"`
	Submitted int64 `json:"submitted"`
	Approved  int64 `json:"approved"`
	Rejected  int64 `json:"rejected"`
}

// VisitReportEmployeeSummary represents an employee with their visit report metrics
type VisitReportEmployeeSummary struct {
	EmployeeID   string                  `json:"employee_id"`
	EmployeeCode string                  `json:"employee_code"`
	EmployeeName string                  `json:"employee_name"`
	TotalReports int64                   `json:"total_reports"`
	LatestVisit  string                  `json:"latest_visit,omitempty"`
	StatusCounts VisitReportStatusCounts `json:"status_counts"`
}
