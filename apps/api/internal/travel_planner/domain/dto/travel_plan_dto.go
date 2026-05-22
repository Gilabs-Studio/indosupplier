package dto

type EnumOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type TravelPlanStopRequest struct {
	PlaceName  string  `json:"place_name" binding:"required,max=255"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Longitude  float64 `json:"longitude" binding:"required"`
	Category   string  `json:"category" binding:"required"`
	OrderIndex int     `json:"order_index"`
	IsLocked   bool    `json:"is_locked"`
	Source     string  `json:"source"`
	PhotoURL   string  `json:"photo_url"`
	Note       string  `json:"note"`
}

type TravelPlanDayNoteRequest struct {
	IconTag    string `json:"icon_tag"`
	NoteText   string `json:"note_text" binding:"required"`
	NoteTime   string `json:"note_time"`
	OrderIndex int    `json:"order_index"`
}

type TravelPlanDayRequest struct {
	DayIndex int                        `json:"day_index" binding:"required,min=1"`
	DayDate  string                     `json:"day_date" binding:"required"`
	Summary  string                     `json:"summary"`
	Stops    []TravelPlanStopRequest    `json:"stops" binding:"required,min=1"`
	Notes    []TravelPlanDayNoteRequest `json:"notes"`
}

type CreateTravelPlanRequest struct {
	Title        string                 `json:"title" binding:"required,max=255"`
	Mode         string                 `json:"mode" binding:"required"`
	StartDate    string                 `json:"start_date" binding:"required"`
	EndDate      string                 `json:"end_date" binding:"required"`
	BudgetAmount float64                `json:"budget_amount" binding:"omitempty,min=0"`
	Notes        string                 `json:"notes"`
	Days         []TravelPlanDayRequest `json:"days" binding:"required,min=1"`
}

type UpdateTravelPlanRequest struct {
	Title        string                 `json:"title" binding:"required,max=255"`
	Mode         string                 `json:"mode" binding:"required"`
	StartDate    string                 `json:"start_date" binding:"required"`
	EndDate      string                 `json:"end_date" binding:"required"`
	Status       string                 `json:"status"`
	BudgetAmount float64                `json:"budget_amount" binding:"omitempty,min=0"`
	Notes        string                 `json:"notes"`
	Days         []TravelPlanDayRequest `json:"days" binding:"required,min=1"`
}

type UpdateTravelPlanParticipantsRequest struct {
	ParticipantIDs []string `json:"participant_ids" binding:"omitempty,dive,uuid"`
}

type ListTravelPlansRequest struct {
	Page      int     `form:"page" binding:"omitempty,min=1"`
	PerPage   int     `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search    string  `form:"search"`
	PlanType  *string `form:"plan_type"`
	Mode      *string `form:"mode"`
	Status    *string `form:"status"`
	StartDate *string `form:"start_date"`
	EndDate   *string `form:"end_date"`
}

type ListTravelPlanParticipantsRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search"`
}

type TravelPlanStopResponse struct {
	ID         string  `json:"id"`
	PlaceName  string  `json:"place_name"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Category   string  `json:"category"`
	OrderIndex int     `json:"order_index"`
	IsLocked   bool    `json:"is_locked"`
	Source     string  `json:"source"`
	PhotoURL   string  `json:"photo_url"`
	Note       string  `json:"note"`
}

type TravelPlanDayNoteResponse struct {
	ID         string `json:"id"`
	IconTag    string `json:"icon_tag"`
	NoteText   string `json:"note_text"`
	NoteTime   string `json:"note_time"`
	OrderIndex int    `json:"order_index"`
}

type TravelPlanDayResponse struct {
	ID       string                      `json:"id"`
	DayIndex int                         `json:"day_index"`
	DayDate  string                      `json:"day_date"`
	Summary  string                      `json:"summary"`
	Stops    []TravelPlanStopResponse    `json:"stops"`
	Notes    []TravelPlanDayNoteResponse `json:"notes"`
}

type TravelPlanResponse struct {
	ID           string                  `json:"id"`
	Code         string                  `json:"code"`
	Title        string                  `json:"title"`
	PlanType     string                  `json:"plan_type"`
	Mode         string                  `json:"mode"`
	StartDate    string                  `json:"start_date"`
	EndDate      string                  `json:"end_date"`
	Status       string                  `json:"status"`
	BudgetAmount float64                 `json:"budget_amount"`
	Notes        string                  `json:"notes"`
	Days         []TravelPlanDayResponse `json:"days"`
	CreatedBy    *string                 `json:"created_by"`
	CreatedAt    string                  `json:"created_at"`
	UpdatedAt    string                  `json:"updated_at"`
}

type CreateTravelExpenseRequest struct {
	ExpenseType string  `json:"expense_type" binding:"required"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ExpenseDate string  `json:"expense_date" binding:"required"`
	ReceiptURL  string  `json:"receipt_url"`
}

type TravelExpenseResponse struct {
	ID           string  `json:"id"`
	TravelPlanID string  `json:"travel_plan_id"`
	ExpenseType  string  `json:"expense_type"`
	Description  string  `json:"description"`
	Amount       float64 `json:"amount"`
	ExpenseDate  string  `json:"expense_date"`
	ReceiptURL   string  `json:"receipt_url"`
	CreatedBy    *string `json:"created_by"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type TravelExpenseListResponse struct {
	Items       []TravelExpenseResponse `json:"items"`
	TotalAmount float64                 `json:"total_amount"`
}

type LinkTravelPlanVisitsRequest struct {
	VisitIDs []string `json:"visit_ids" binding:"required,min=1,dive,uuid"`
}

type CreateTravelPlanVisitRequest struct {
	VisitDate     string  `json:"visit_date" binding:"required"`
	EmployeeID    string  `json:"employee_id" binding:"required,uuid"`
	CustomerID    *string `json:"customer_id" binding:"omitempty,uuid"`
	ContactID     *string `json:"contact_id" binding:"omitempty,uuid"`
	DealID        *string `json:"deal_id" binding:"omitempty,uuid"`
	LeadID        *string `json:"lead_id" binding:"omitempty,uuid"`
	VillageID     *string `json:"village_id" binding:"omitempty,uuid"`
	ContactPerson string  `json:"contact_person" binding:"max=200"`
	ContactPhone  string  `json:"contact_phone" binding:"max=20"`
	Address       string  `json:"address"`
	Purpose       string  `json:"purpose"`
	Notes         string  `json:"notes"`
}

type TravelPlanVisitResponse struct {
	ID                string                   `json:"id"`
	Code              string                   `json:"code"`
	TravelPlanID      *string                  `json:"travel_plan_id"`
	VisitDate         string                   `json:"visit_date"`
	EmployeeID        string                   `json:"employee_id"`
	EmployeeName      string                   `json:"employee_name"`
	EmployeeAvatarURL string                   `json:"employee_avatar_url"`
	CustomerID        *string                  `json:"customer_id"`
	CustomerName      string                   `json:"customer_name"`
	Status            string                   `json:"status"`
	Purpose           string                   `json:"purpose"`
	Outcome           string                   `json:"outcome"`
	CreatedAt         string                   `json:"created_at"`
	// Visit execution details
	CheckInAt        *string                  `json:"check_in_at,omitempty"`
	CheckOutAt       *string                  `json:"check_out_at,omitempty"`
	CheckInLocation  *string                  `json:"check_in_location,omitempty"`
	CheckOutLocation *string                  `json:"check_out_location,omitempty"`
	// Product interests count
	ProductInterestCount int                   `json:"product_interest_count"`
	// Documentation
	Photos           *string                  `json:"photos,omitempty"`
	Notes            string                   `json:"notes"`
	Result           string                   `json:"result"`
}

type PlaceSearchResult struct {
	Provider  string   `json:"provider"`
	PlaceName string   `json:"place_name"`
	Address   string   `json:"address"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Category  string   `json:"category"`
	PhotoURL  string   `json:"photo_url"`
	Rating    *float64 `json:"rating"`
}

type RouteOptimizationDaySummary struct {
	DayID            string   `json:"day_id"`
	DayIndex         int      `json:"day_index"`
	TotalDistanceKM  float64  `json:"total_distance_km"`
	GoogleMapsURL    string   `json:"google_maps_url"`
	OptimizedStopIDs []string `json:"optimized_stop_ids"`
}

type RouteOptimizationResponse struct {
	PlanID      string                        `json:"plan_id"`
	OptimizedAt string                        `json:"optimized_at"`
	Days        []RouteOptimizationDaySummary `json:"days"`
}

type DayGoogleMapsLink struct {
	DayID    string `json:"day_id"`
	DayIndex int    `json:"day_index"`
	URL      string `json:"url"`
}

type EmployeeFormOption struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url"`
}

type TravelPlannerFormDataResponse struct {
	Modes        []EnumOption         `json:"modes"`
	Categories   []EnumOption         `json:"categories"`
	Sources      []EnumOption         `json:"sources"`
	Employees    []EmployeeFormOption `json:"employees"`
	ExpenseTypes []EnumOption         `json:"expense_types"`
}
