package dto

type NavigationOptimizeMode string

type NavigationOptimizeFor string

type VisitActionType string

const (
	NavigationOptimizeModeDriving NavigationOptimizeMode = "driving"
	NavigationOptimizeModeWalking NavigationOptimizeMode = "walking"

	NavigationOptimizeForTime     NavigationOptimizeFor = "time"
	NavigationOptimizeForDistance NavigationOptimizeFor = "distance"

	VisitActionCheckIn     VisitActionType = "check_in"
	VisitActionCheckOut    VisitActionType = "check_out"
	VisitActionSubmitVisit VisitActionType = "submit_visit"
)

type NavigationCheckpointInput struct {
	ID    *string  `json:"id"`
	Type  string   `json:"type" binding:"required,oneof=lead deal customer pipeline"`
	RefID *string  `json:"ref_id"`
	Lat   *float64 `json:"lat"`
	Lng   *float64 `json:"lng"`
}

type NavigationOptimizeOptions struct {
	Mode        NavigationOptimizeMode `json:"mode" binding:"omitempty,oneof=driving walking"`
	OptimizeFor NavigationOptimizeFor  `json:"optimizeFor" binding:"omitempty,oneof=time distance"`
}

type OptimizeNavigationRequest struct {
	EmployeeID  *string                     `json:"employee_id" binding:"omitempty,uuid"`
	Checkpoints []NavigationCheckpointInput `json:"checkpoints" binding:"required,min=1,dive"`
	Options     *NavigationOptimizeOptions  `json:"options"`
}

type OptimizedNavigationCheckpoint struct {
	CheckpointID string  `json:"checkpoint_id"`
	Type         string  `json:"type"`
	RefID        *string `json:"ref_id,omitempty"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	Sequence     int     `json:"sequence"`
	LegDistanceM int64   `json:"leg_distance_m"`
	LegDurationS int64   `json:"leg_duration_s"`
	Warning      string  `json:"warning,omitempty"`
}

type OptimizeNavigationSummary struct {
	TotalDistanceM int64 `json:"total_distance_m"`
	TotalDurationS int64 `json:"total_duration_s"`
}

type OptimizeNavigationResponse struct {
	OrderedCheckpoints []OptimizedNavigationCheckpoint `json:"ordered_checkpoints"`
	Polyline           string                          `json:"polyline"`
	Summary            OptimizeNavigationSummary       `json:"summary"`
	Warnings           []string                        `json:"warnings"`
}

type VisitLogLocationInput struct {
	Lat float64 `json:"lat" binding:"required"`
	Lng float64 `json:"lng" binding:"required"`
}

type UpsertVisitLogRequest struct {
	EmployeeID       *string                     `json:"employee_id" binding:"omitempty,uuid"`
	VisitID          *string                     `json:"visit_id" binding:"omitempty,uuid"`
	RouteID          *string                     `json:"route_id" binding:"omitempty,uuid"`
	CheckpointID     *string                     `json:"checkpoint_id"`
	LeadID           *string                     `json:"lead_id" binding:"omitempty,uuid"`
	DealID           *string                     `json:"deal_id" binding:"omitempty,uuid"`
	CustomerID       *string                     `json:"customer_id" binding:"omitempty,uuid"`
	Action           VisitActionType             `json:"action" binding:"required,oneof=check_in check_out submit_visit"`
	Timestamp        *string                     `json:"timestamp"`
	Location         *VisitLogLocationInput      `json:"location"`
	Photos           []string                    `json:"photos"`
	Notes            string                      `json:"notes"`
	Outcome          string                      `json:"outcome"`
	ActivityType     string                      `json:"activity_type"`
	DistanceM        *float64                    `json:"distance_m"`
	ProductInterests []VisitProductInterestInput `json:"product_interests"`
}

type VisitProductInterestInput struct {
	ProductID     string   `json:"product_id" binding:"required,uuid"`
	InterestLevel int      `json:"interest_level" binding:"omitempty,min=0,max=5"`
	Notes         string   `json:"notes"`
	Quantity      *float64 `json:"quantity" binding:"omitempty,gte=0"`
	Price         *float64 `json:"price" binding:"omitempty,gte=0"`
}

type VisitLogResponse struct {
	Action string                  `json:"action"`
	Visit  TravelPlanVisitResponse `json:"visit"`
}

type LocationSubscriptionRequest struct {
	EmployeeIDs []string `json:"employee_ids"`
	AreaBBox    *string  `json:"area_bbox"`
}

type LocationUpdateRequest struct {
	EmployeeID   *string  `json:"employee_id" binding:"omitempty,uuid"`
	RouteID      *string  `json:"route_id" binding:"omitempty,uuid"`
	CheckpointID *string  `json:"checkpoint_id"`
	Lat          float64  `json:"lat" binding:"required"`
	Lng          float64  `json:"lng" binding:"required"`
	Heading      *float64 `json:"heading"`
}

type LocationUpdateResponse struct {
	EmployeeID       string   `json:"employee_id"`
	RouteID          *string  `json:"route_id,omitempty"`
	CheckpointID     *string  `json:"checkpoint_id,omitempty"`
	Lat              float64  `json:"lat"`
	Lng              float64  `json:"lng"`
	Heading          *float64 `json:"heading,omitempty"`
	NavigationStatus string   `json:"navigation_status,omitempty"`
	Timestamp        string   `json:"timestamp"`
}

// StartNavigationRequest is sent by the sales employee's device when they begin
// navigating to their first checkpoint.  The server broadcasts a navigation_started
// WebSocket event so supervisors with the appropriate scope can see it in real time.
type StartNavigationRequest struct {
	EmployeeID *string  `json:"employee_id" binding:"omitempty,uuid"`
	RouteID    *string  `json:"route_id" binding:"omitempty,uuid"`
	Lat        float64  `json:"lat" binding:"required"`
	Lng        float64  `json:"lng" binding:"required"`
	Heading    *float64 `json:"heading"`
}

// StopNavigationRequest is sent when the sales employee ends their navigation session.
type StopNavigationRequest struct {
	EmployeeID *string `json:"employee_id" binding:"omitempty,uuid"`
	RouteID    *string `json:"route_id" binding:"omitempty,uuid"`
}

// NavigationStatusResponse is returned for both start and stop navigation calls.
type NavigationStatusResponse struct {
	EmployeeID string   `json:"employee_id"`
	RouteID    *string  `json:"route_id,omitempty"`
	Lat        *float64 `json:"lat,omitempty"`
	Lng        *float64 `json:"lng,omitempty"`
	Status     string   `json:"status"` // "navigating" | "idle"
	Timestamp  string   `json:"timestamp"`
}

type ActiveVisitRouteCheckpoint struct {
	VisitID              string   `json:"visit_id"`
	CheckpointID         string   `json:"checkpoint_id"`
	Type                 string   `json:"type"`
	RefID                *string  `json:"ref_id,omitempty"`
	Label                string   `json:"label"`
	Lat                  *float64 `json:"lat,omitempty"`
	Lng                  *float64 `json:"lng,omitempty"`
	Status               string   `json:"status"`
	Warning              string   `json:"warning,omitempty"`
	ProductInterestCount int      `json:"product_interest_count"`
	DocumentationCount   int      `json:"documentation_count"`
}

type ActiveVisitRouteResponse struct {
	RouteID           string                       `json:"route_id"`
	PlanCode          string                       `json:"plan_code"`
	PlanTitle         string                       `json:"plan_title"`
	EmployeeID        string                       `json:"employee_id"`
	EmployeeName      string                       `json:"employee_name"`
	EmployeeAvatarURL string                       `json:"employee_avatar_url"`
	CheckpointTotal   int                          `json:"checkpoint_total"`
	CompletedTotal    int                          `json:"completed_total"`
	InProgressTotal   int                          `json:"in_progress_total"`
	CurrentETAS       int64                        `json:"current_eta_s"`
	Polyline          string                       `json:"polyline"`
	Checkpoints       []ActiveVisitRouteCheckpoint `json:"checkpoints"`
}

type ListVisitPlannerRoutesRequest struct {
	Page       int     `form:"page" binding:"omitempty,min=1"`
	PerPage    int     `form:"per_page" binding:"omitempty,min=1,max=100"`
	EmployeeID *string `form:"employee_id" binding:"omitempty,uuid"`
	DivisionID *string `form:"division_id" binding:"omitempty,uuid"`
	RouteDate  string  `form:"route_date"`
}

type VisitPlannerCandidate struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Label       string   `json:"label"`
	AssignedTo  *string  `json:"assigned_to,omitempty"`
	Lat         *float64 `json:"lat,omitempty"`
	Lng         *float64 `json:"lng,omitempty"`
	HasLocation bool     `json:"has_location"`
	Warning     string   `json:"warning,omitempty"`
}

type VisitPlannerFormDataRequest struct {
	EmployeeID *string `form:"employee_id" binding:"omitempty,uuid"`
	Search     string  `form:"search"`
}

type VisitPlannerFormDataResponse struct {
	Employees []EmployeeFormOption        `json:"employees"`
	Leads     []VisitPlannerCandidate     `json:"leads"`
	Deals     []VisitPlannerCandidate     `json:"deals"`
	Customers []VisitPlannerCandidate     `json:"customers"`
	Products  []VisitPlannerProductOption `json:"products"`
	Warnings  []string                    `json:"warnings"`
}

type VisitPlannerProductOption struct {
	ID           string  `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	SellingPrice float64 `json:"selling_price"`
}

type CreateVisitPlannerPlanRequest struct {
	Title       string                      `json:"title" binding:"max=255"`
	RouteDate   string                      `json:"route_date" binding:"required"`
	EmployeeID  *string                     `json:"employee_id" binding:"omitempty,uuid"`
	Checkpoints []NavigationCheckpointInput `json:"checkpoints" binding:"required,min=1,dive"`
}

type CreateVisitPlannerPlanResponse struct {
	RouteID         string   `json:"route_id"`
	PlanCode        string   `json:"plan_code"`
	PlanTitle       string   `json:"plan_title"`
	EmployeeID      string   `json:"employee_id"`
	CheckpointTotal int      `json:"checkpoint_total"`
	VisitIDs        []string `json:"visit_ids"`
}
