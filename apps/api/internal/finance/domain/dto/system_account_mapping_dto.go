package dto

import "time"

type SystemAccountMappingCOAInfo struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	IsPostable bool   `json:"is_postable"`
	IsActive   bool   `json:"is_active"`
}

type SystemAccountMappingResponse struct {
	ID        string                      `json:"id"`
	Key       string                      `json:"key"`
	CompanyID *string                     `json:"company_id"`
	COACode   string                      `json:"coa_code"`
	Label     string                      `json:"label"`
	COA       SystemAccountMappingCOAInfo `json:"coa"`
	CreatedAt time.Time                   `json:"created_at"`
	UpdatedAt time.Time                   `json:"updated_at"`
}

type UpsertSystemAccountMappingRequest struct {
	COACode string `json:"coa_code" binding:"required"`
	Label   string `json:"label"`
}
