package dto

import (
	"encoding/json"
	"time"
)

// ─── Form Schema (JSON Builder) ────────────────────────────────────────────────

// QuestionType enumerates supported question types.
type QuestionType string

const (
	QuestionTypeRating         QuestionType = "rating"
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeText           QuestionType = "text"
)

// RatingConfig holds configuration for a rating question.
type RatingConfig struct {
	Min  int    `json:"min"`
	Max  int    `json:"max"`
	Icon string `json:"icon"` // "star" | "heart" | "thumb"
}

// MultipleChoiceConfig holds configuration for a multiple-choice question.
type MultipleChoiceConfig struct {
	Options       []string `json:"options"`
	AllowMultiple bool     `json:"allow_multiple"`
}

// TextConfig holds configuration for a free-text question.
type TextConfig struct {
	Placeholder string `json:"placeholder,omitempty"`
	MaxLength   int    `json:"max_length,omitempty"`
}

// FormQuestion represents a single question within a feedback form schema.
type FormQuestion struct {
	ID       string          `json:"id"`
	Type     QuestionType    `json:"type"`
	Label    string          `json:"label"`
	Required bool            `json:"required"`
	Config   json.RawMessage `json:"config"`
}

// FormSchema is the top-level structure stored in FeedbackForm.SchemaJSON.
type FormSchema struct {
	Questions []FormQuestion `json:"questions"`
}

// ─── Form Requests ─────────────────────────────────────────────────────────────

type CreateFeedbackFormRequest struct {
	OutletID    string          `json:"outlet_id" binding:"required,uuid"`
	Title       string          `json:"title" binding:"required,min=2,max=255"`
	Description *string         `json:"description"`
	SchemaJSON  json.RawMessage `json:"schema_json" binding:"required"`
	IsActive    bool            `json:"is_active"`
}

type UpdateFeedbackFormRequest struct {
	Title       *string         `json:"title" binding:"omitempty,min=2,max=255"`
	Description *string         `json:"description"`
	SchemaJSON  json.RawMessage `json:"schema_json"`
	IsActive    *bool           `json:"is_active"`
}

// ─── Form Response ─────────────────────────────────────────────────────────────

type FeedbackFormResponse struct {
	ID          string          `json:"id"`
	OutletID    string          `json:"outlet_id"`
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	SchemaJSON  json.RawMessage `json:"schema_json"`
	Version     int             `json:"version"`
	IsActive    bool            `json:"is_active"`
	CreatedBy   *string         `json:"created_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ─── Token ─────────────────────────────────────────────────────────────────────

type GenerateTokenRequest struct {
	OutletID     string  `json:"outlet_id" binding:"required,uuid"`
	PosOrderID   *string `json:"pos_order_id"`
	CustomerName *string `json:"customer_name"`
}

type FeedbackTokenResponse struct {
	Token       string    `json:"token"`
	FeedbackURL string    `json:"feedback_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// ─── Public Form ───────────────────────────────────────────────────────────────

// PublicFormResponse is returned to the unauthenticated customer for rendering.
type PublicFormResponse struct {
	Token        string          `json:"token"`
	OutletName   string          `json:"outlet_name"`
	Title        string          `json:"title"`
	Description  *string         `json:"description,omitempty"`
	SchemaJSON   json.RawMessage `json:"schema_json"`
	CustomerName *string         `json:"customer_name,omitempty"`
}

// ─── Submission ────────────────────────────────────────────────────────────────

type SubmitFeedbackRequest struct {
	CustomerName *string `json:"customer_name"`
	// Answers must be an object keyed by question ID, e.g. {"q1":4,"q2":["Enak"],"q3":"Bagus!"}
	Answers json.RawMessage `json:"answers" binding:"required"`
}

// ─── Response (admin list) ─────────────────────────────────────────────────────

type FeedbackResponseItem struct {
	ID           string          `json:"id"`
	FormID       string          `json:"form_id"`
	FormTitle    string          `json:"form_title"`
	SchemaJSON   json.RawMessage `json:"schema_json,omitempty"`
	OutletID     string          `json:"outlet_id"`
	OutletName   string          `json:"outlet_name"`
	SalesOrderID *string         `json:"sales_order_id,omitempty"`
	PosOrderID   *string         `json:"pos_order_id,omitempty"`
	CustomerName *string         `json:"customer_name,omitempty"`
	Answers      json.RawMessage `json:"answers"`
	AvgScore     *float64        `json:"avg_score,omitempty"`
	SubmittedAt  time.Time       `json:"submitted_at"`
}

// CopyFeedbackFormRequest allows cloning an existing form to one or more outlets.
// If ApplyToAllOutlets is true, OutletIDs is ignored.
type CopyFeedbackFormRequest struct {
	OutletIDs         []string `json:"outlet_ids"`
	ApplyToAllOutlets bool     `json:"apply_to_all_outlets"`
	IsActive          bool     `json:"is_active"`
	Title             *string  `json:"title"`
}

// CopyFeedbackFormResponse returns clone operation summary.
type CopyFeedbackFormResponse struct {
	CopiedCount int                    `json:"copied_count"`
	Forms       []FeedbackFormResponse `json:"forms"`
}

// ─── List Responses Request ────────────────────────────────────────────────────

type ListFeedbackResponsesRequest struct {
	OutletID  string `form:"outlet_id"`
	FormID    string `form:"form_id"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Search    string `form:"search"`
	Page      int    `form:"page,default=1"`
	PerPage   int    `form:"per_page,default=20"`
}
