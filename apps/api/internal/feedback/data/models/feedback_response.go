package models

import (
	"time"

	"gorm.io/datatypes"
)

// FeedbackResponse stores a single customer submission.
// Answers is a free-form JSONB map keyed by question ID, e.g.:
//
//	{"q1": 4, "q2": ["Makanan", "Suasana"], "q3": "Enak sekali!"}
type FeedbackResponse struct {
	ID           string         `gorm:"type:uuid;primary_key" json:"id"`
	TenantID string `gorm:"column:tenant_id;type:uuid;index" json:"tenant_id,omitempty"`
	FormID       string         `gorm:"type:uuid;not null;index" json:"form_id"`
	TokenID      string         `gorm:"type:uuid;not null;uniqueIndex" json:"token_id"`
	OutletID     string         `gorm:"type:uuid;not null;index" json:"outlet_id"`
	PosOrderID   *string        `gorm:"type:uuid;index" json:"pos_order_id,omitempty"`
	// SalesOrderID is a read-only projection populated via JOIN on sales_orders.source_pos_order_id.
	SalesOrderID *string        `gorm:"->;column:sales_order_id" json:"sales_order_id,omitempty"`
	CustomerName *string        `gorm:"type:varchar(255)" json:"customer_name,omitempty"`
	// Answers holds the submitted answers keyed by question ID.
	Answers      datatypes.JSON `gorm:"type:jsonb;not null" json:"answers"`
	// AvgScore is the pre-computed average of all numeric (rating) answers for
	// quick sorting/filtering without re-parsing JSONB in queries.
	AvgScore     *float64       `gorm:"type:decimal(5,2)" json:"avg_score,omitempty"`
	SubmittedAt  time.Time      `gorm:"type:timestamptz;not null;index" json:"submitted_at"`

	// Relations (preloaded on demand)
	Form  *FeedbackForm  `gorm:"foreignKey:FormID" json:"form,omitempty"`
	Token *FeedbackToken `gorm:"foreignKey:TokenID" json:"token,omitempty"`
}

func (FeedbackResponse) TableName() string { return "feedback_responses" }
