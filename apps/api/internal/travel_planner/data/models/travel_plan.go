package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TravelMode string

const (
	TravelModeLogistic  TravelMode = "logistic"
	TravelModeCargo     TravelMode = "cargo"
	TravelModeVessel    TravelMode = "vessel"
	TravelModeMilestone TravelMode = "milestone"
)

type TravelPlanStatus string

const (
	TravelPlanStatusDraft     TravelPlanStatus = "draft"
	TravelPlanStatusActive    TravelPlanStatus = "active"
	TravelPlanStatusCompleted TravelPlanStatus = "completed"
	TravelPlanStatusCancelled TravelPlanStatus = "cancelled"
)

type TravelPlanType string

const (
	TravelPlanTypeUpCountryCost TravelPlanType = "up_country_cost"
	TravelPlanTypeVisitReport   TravelPlanType = "visit_report"
)

type TravelStopCategory string

const (
	TravelStopCategoryPickup     TravelStopCategory = "pickup"
	TravelStopCategoryDropoff    TravelStopCategory = "dropoff"
	TravelStopCategoryRefuel     TravelStopCategory = "refuel"
	TravelStopCategoryCheckpoint TravelStopCategory = "checkpoint"
	TravelStopCategoryRest       TravelStopCategory = "rest"
	TravelStopCategoryCustom     TravelStopCategory = "custom"
)

type TravelStopSource string

const (
	TravelStopSourceManual        TravelStopSource = "manual"
	TravelStopSourceGooglePlaces  TravelStopSource = "google_places"
	TravelStopSourceOpenStreetMap TravelStopSource = "open_street_map"
)

type TravelExpenseType string

const (
	TravelExpenseTypeTransport     TravelExpenseType = "transport"
	TravelExpenseTypeAccommodation TravelExpenseType = "accommodation"
	TravelExpenseTypeMeal          TravelExpenseType = "meal"
	TravelExpenseTypeFuel          TravelExpenseType = "fuel"
	TravelExpenseTypeToll          TravelExpenseType = "toll"
	TravelExpenseTypeParking       TravelExpenseType = "parking"
	TravelExpenseTypeOther         TravelExpenseType = "other"
)

type TravelPlan struct {
	ID           string           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     string           `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	Code         string           `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Title        string           `gorm:"type:varchar(255);not null" json:"title"`
	PlanType     TravelPlanType   `gorm:"type:varchar(30);default:'up_country_cost';index;not null" json:"plan_type"`
	Mode         TravelMode       `gorm:"type:varchar(30);index;not null" json:"mode"`
	StartDate    time.Time        `gorm:"type:date;not null" json:"start_date"`
	EndDate      time.Time        `gorm:"type:date;not null" json:"end_date"`
	Status       TravelPlanStatus `gorm:"type:varchar(30);default:'draft';index" json:"status"`
	BudgetAmount float64          `gorm:"type:numeric(18,2);default:0" json:"budget_amount"`
	Notes        string           `gorm:"type:text" json:"notes"`

	Days     []TravelPlanDay     `gorm:"foreignKey:TravelPlanID;constraint:OnDelete:CASCADE" json:"days,omitempty"`
	Expenses []TravelPlanExpense `gorm:"foreignKey:TravelPlanID;constraint:OnDelete:CASCADE" json:"expenses,omitempty"`

	CreatedBy *string        `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TravelPlan) TableName() string {
	return "travel_plans"
}

func (t *TravelPlan) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

type TravelPlanDay struct {
	ID           string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	TravelPlanID string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_travel_plan_day_order,priority:1" json:"travel_plan_id"`
	DayIndex     int       `gorm:"not null;uniqueIndex:idx_travel_plan_day_order,priority:2" json:"day_index"`
	DayDate      time.Time `gorm:"type:date;not null" json:"day_date"`
	Summary      string    `gorm:"type:text" json:"summary"`

	Stops []TravelPlanStop    `gorm:"foreignKey:TravelPlanDayID;constraint:OnDelete:CASCADE" json:"stops,omitempty"`
	Notes []TravelPlanDayNote `gorm:"foreignKey:TravelPlanDayID;constraint:OnDelete:CASCADE" json:"notes,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TravelPlanDay) TableName() string {
	return "travel_plan_days"
}

func (t *TravelPlanDay) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

type TravelPlanStop struct {
	ID              string             `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID        string             `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	TravelPlanDayID string             `gorm:"type:uuid;not null;index" json:"travel_plan_day_id"`
	PlaceName       string             `gorm:"type:varchar(255);not null" json:"place_name"`
	Latitude        float64            `gorm:"type:numeric(10,6);not null" json:"latitude"`
	Longitude       float64            `gorm:"type:numeric(10,6);not null" json:"longitude"`
	Category        TravelStopCategory `gorm:"type:varchar(30);index" json:"category"`
	OrderIndex      int                `gorm:"not null;index" json:"order_index"`
	IsLocked        bool               `gorm:"not null;default:false" json:"is_locked"`
	Source          TravelStopSource   `gorm:"type:varchar(40);default:'manual'" json:"source"`
	PhotoURL        string             `gorm:"type:text" json:"photo_url"`
	Note            string             `gorm:"type:text" json:"note"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TravelPlanStop) TableName() string {
	return "travel_plan_stops"
}

func (t *TravelPlanStop) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

type TravelPlanDayNote struct {
	ID              string `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID        string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	TravelPlanDayID string `gorm:"type:uuid;not null;index" json:"travel_plan_day_id"`
	IconTag         string `gorm:"type:varchar(40)" json:"icon_tag"`
	NoteText        string `gorm:"type:text;not null" json:"note_text"`
	NoteTime        string `gorm:"type:varchar(5)" json:"note_time"`
	OrderIndex      int    `gorm:"not null;index" json:"order_index"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TravelPlanDayNote) TableName() string {
	return "travel_plan_day_notes"
}

func (t *TravelPlanDayNote) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

type TravelPlanExpense struct {
	ID           string            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     string            `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	TravelPlanID string            `gorm:"type:uuid;not null;index" json:"travel_plan_id"`
	ExpenseType  TravelExpenseType `gorm:"type:varchar(30);not null;index" json:"expense_type"`
	Description  string            `gorm:"type:text" json:"description"`
	Amount       float64           `gorm:"type:numeric(18,2);not null" json:"amount"`
	ExpenseDate  time.Time         `gorm:"type:date;not null;index" json:"expense_date"`
	ReceiptURL   string            `gorm:"type:text" json:"receipt_url"`
	CreatedBy    *string           `gorm:"type:uuid" json:"created_by"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TravelPlanExpense) TableName() string {
	return "travel_plan_expenses"
}

func (t *TravelPlanExpense) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
