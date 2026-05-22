package dto

import "time"

type AdjustmentApprovalActionRequest struct {
	Notes string `json:"notes"`
}

type AdjustmentJournalApprovalResponse struct {
	ID        string    `json:"id"`
	JournalID string    `json:"journal_id"`
	Action    string    `json:"action"`
	ActorID   string    `json:"actor_id"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

type ListJournalTemplatesRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PerPage   int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search    string `form:"search"`
	CompanyID string `form:"company_id" binding:"omitempty,uuid"`
}

type CreateJournalTemplateRequest struct {
	CompanyID    string               `json:"company_id" binding:"required,uuid"`
	TemplateName string               `json:"template_name" binding:"required,min=3"`
	JournalType  string               `json:"journal_type" binding:"required"`
	Description  string               `json:"description"`
	Lines        []JournalLineRequest `json:"lines" binding:"required,min=2"`
}

type JournalTemplateResponse struct {
	ID           string    `json:"id"`
	CompanyID    string    `json:"company_id"`
	TemplateName string    `json:"template_name"`
	JournalType  string    `json:"journal_type"`
	Description  string    `json:"description"`
	Lines        string    `json:"lines"`
	CreatedBy    *string   `json:"created_by,omitempty"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UseJournalTemplateResponse struct {
	Template JournalTemplateResponse   `json:"template"`
	Prefill  CreateAdjustmentJournalRequest `json:"prefill"`
}
