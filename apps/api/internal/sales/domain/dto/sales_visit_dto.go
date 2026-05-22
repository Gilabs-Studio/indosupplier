package dto

// CreateSalesVisitRequest represents the request to create a sales visit
type CreateSalesVisitRequest struct {
	VisitDate     string   `json:"visit_date" binding:"required"`
	ScheduledTime *string  `json:"scheduled_time"`
	EmployeeID    string   `json:"employee_id" binding:"required,uuid"`
	CompanyID     *string  `json:"company_id" binding:"omitempty,uuid"`
	ContactPerson string   `json:"contact_person"`
	ContactPhone  string   `json:"contact_phone"`
	Address       string   `json:"address"`
	VillageID     *string  `json:"village_id" binding:"omitempty,uuid"`
	Purpose       string   `json:"purpose"`
	Notes         string   `json:"notes"`
	Details       []CreateSalesVisitDetailRequest `json:"details"`
}

// CreateSalesVisitDetailRequest represents a detail item in the visit
type CreateSalesVisitDetailRequest struct {
	ProductID     string   `json:"product_id" binding:"required,uuid"`
	InterestLevel int      `json:"interest_level" binding:"min=0,max=5"`
	Notes         string   `json:"notes"`
	Quantity      *float64 `json:"quantity" binding:"omitempty,gte=0"`
	Price         *float64 `json:"price" binding:"omitempty,gte=0"`
	Answers       []CreateSalesVisitInterestAnswerRequest `json:"answers"`
}

// CreateSalesVisitInterestAnswerRequest represents the answer to a survey question
type CreateSalesVisitInterestAnswerRequest struct {
	QuestionID string `json:"question_id" binding:"required,uuid"`
	OptionID   string `json:"option_id" binding:"required,uuid"`
}

// UpdateSalesVisitRequest represents the request to update a sales visit
type UpdateSalesVisitRequest struct {
	VisitDate     *string  `json:"visit_date"`
	ScheduledTime *string  `json:"scheduled_time"`
	EmployeeID    *string  `json:"employee_id" binding:"omitempty,uuid"`
	CompanyID     *string  `json:"company_id" binding:"omitempty,uuid"`
	ContactPerson *string  `json:"contact_person"`
	ContactPhone  *string  `json:"contact_phone"`
	Address       *string  `json:"address"`
	VillageID     *string  `json:"village_id" binding:"omitempty,uuid"`
	Purpose       *string  `json:"purpose"`
	Notes         *string  `json:"notes"`
	Result        *string  `json:"result"`
	Details       *[]CreateSalesVisitDetailRequest `json:"details"`
}

// ListSalesVisitsRequest represents the request to list sales visits
type ListSalesVisitsRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PerPage    int    `form:"per_page" binding:"omitempty,min=1,max=1000"`
	Search     string `form:"search"`
	Status     string `form:"status" binding:"omitempty,oneof=planned in_progress completed cancelled"`
	EmployeeID string `form:"employee_id" binding:"omitempty,uuid"`
	CompanyID  string `form:"company_id" binding:"omitempty,uuid"`
	DateFrom   string `form:"date_from"`
	DateTo     string `form:"date_to"`
	SortBy     string `form:"sort_by"`
	SortDir    string `form:"sort_dir" binding:"omitempty,oneof=asc desc"`
}

// ListSalesVisitDetailsRequest represents the request to list visit details
type ListSalesVisitDetailsRequest struct {
	Page    int `form:"page" binding:"omitempty,min=1"`
	PerPage int `form:"per_page" binding:"omitempty,min=1,max=100"`
}

// ListSalesVisitProgressHistoryRequest represents the request to list progress history
type ListSalesVisitProgressHistoryRequest struct {
	Page    int `form:"page" binding:"omitempty,min=1"`
	PerPage int `form:"per_page" binding:"omitempty,min=1,max=100"`
}

// UpdateSalesVisitStatusRequest represents the request to update visit status
type UpdateSalesVisitStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=planned in_progress completed cancelled"`
	Notes  string `json:"notes"`
}

// CheckInRequest represents the request to check in to a visit
type CheckInRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// CheckOutRequest represents the request to check out from a visit
type CheckOutRequest struct {
	Result string `json:"result"`
}

// SalesVisitResponse represents the response for a sales visit
type SalesVisitResponse struct {
	ID            string                        `json:"id"`
	Code          string                        `json:"code"`
	VisitDate     string                        `json:"visit_date"`
	ScheduledTime *string                       `json:"scheduled_time"`
	ActualTime    *string                       `json:"actual_time"`
	EmployeeID    string                        `json:"employee_id"`
	Employee      *EmployeeBriefResponse        `json:"employee,omitempty"`
	CompanyID     *string                       `json:"company_id"`
	Company       *CompanyBriefResponse         `json:"company,omitempty"`
	ContactPerson string                        `json:"contact_person"`
	ContactPhone  string                        `json:"contact_phone"`
	Address       string                        `json:"address"`
	VillageID     *string                       `json:"village_id"`
	Village       *VillageResponse              `json:"village,omitempty"`
	Latitude      *float64                      `json:"latitude"`
	Longitude     *float64                      `json:"longitude"`
	Purpose       string                        `json:"purpose"`
	Notes         string                        `json:"notes"`
	Result        string                        `json:"result"`
	Status        string                        `json:"status"`
	CheckInAt     *string                       `json:"check_in_at"`
	CheckOutAt    *string                       `json:"check_out_at"`
	CreatedBy     *string                       `json:"created_by"`
	CancelledBy   *string                       `json:"cancelled_by"`
	CancelledAt   *string                       `json:"cancelled_at"`
	Details       []SalesVisitDetailResponse    `json:"details,omitempty"`
	CreatedAt     string                        `json:"created_at"`
	UpdatedAt     string                        `json:"updated_at"`
}

// SalesVisitDetailResponse represents a detail item in the visit response
type SalesVisitDetailResponse struct {
	ID            string           `json:"id"`
	SalesVisitID  string           `json:"sales_visit_id"`
	ProductID     string           `json:"product_id"`
	Product       *ProductResponse `json:"product,omitempty"`
	InterestLevel int              `json:"interest_level"`
	Notes         string           `json:"notes"`
	Quantity      *float64         `json:"quantity"`
	Price         *float64 `json:"price"`
	Answers       []SalesVisitInterestAnswerResponse `json:"answers,omitempty"`
	CreatedAt     string           `json:"created_at"`
	UpdatedAt     string           `json:"updated_at"`
}

// SalesVisitInterestAnswerResponse represents the response for a survey answer
type SalesVisitInterestAnswerResponse struct {
	ID           string `json:"id"`
	QuestionID   string `json:"question_id"`
	QuestionText string `json:"question_text"`
	OptionID     string `json:"option_id"`
	OptionText   string `json:"option_text"`
	Score        int    `json:"score"`
}

// SalesVisitInterestQuestionResponse represents the response for a survey question
type SalesVisitInterestQuestionResponse struct {
	ID           string                             `json:"id"`
	QuestionText string                             `json:"question_text"`
	Sequence     int                                `json:"sequence"`
	Options      []SalesVisitInterestOptionResponse `json:"options"`
}

// SalesVisitInterestOptionResponse represents the response for a survey option
type SalesVisitInterestOptionResponse struct {
	ID         string `json:"id"`
	OptionText string `json:"option_text"`
	Score      int    `json:"score"`
}

// SalesVisitProgressHistoryResponse represents a progress history item
type SalesVisitProgressHistoryResponse struct {
	ID           string  `json:"id"`
	SalesVisitID string  `json:"sales_visit_id"`
	FromStatus   string  `json:"from_status"`
	ToStatus     string  `json:"to_status"`
	Notes        string  `json:"notes"`
	ChangedBy    *string `json:"changed_by"`
	CreatedAt    string  `json:"created_at"`
}

// EmployeeBriefResponse represents a brief employee info in response
type EmployeeBriefResponse struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}

// CompanyBriefResponse represents a brief company info in response
type CompanyBriefResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

// VillageResponse represents a village with geographic hierarchy
type VillageResponse struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	District *DistrictResponse `json:"district,omitempty"`
}

// DistrictResponse represents a district in response
type DistrictResponse struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Regency *RegencyResponse `json:"regency,omitempty"`
}

// RegencyResponse represents a regency in response
type RegencyResponse struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Province *ProvinceResponse `json:"province,omitempty"`
}

// ProvinceResponse represents a province in response
type ProvinceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
