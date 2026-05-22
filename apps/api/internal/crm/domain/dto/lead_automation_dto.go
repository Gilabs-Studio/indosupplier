package dto

// LeadAutomationTriggerRequest defines request payload for triggering n8n lead generation.
type LeadAutomationTriggerRequest struct {
	Type         *int   `json:"type" binding:"required,oneof=0 1"`
	Keyword      string `json:"keyword" binding:"required,min=1,max=200"`
	City         string `json:"city" binding:"required,min=1,max=100"`
	Limit        int    `json:"limit" binding:"omitempty,min=1,max=100"`
	LeadSourceID string `json:"lead_source_id" binding:"omitempty,uuid"`
	ERPBaseURL   string `json:"erp_base_url" binding:"required,url"`
}

// LeadAutomationConnectionResponse describes n8n connectivity check result.
type LeadAutomationConnectionResponse struct {
	Reachable    bool   `json:"reachable"`
	Status       int    `json:"status"`
	WebhookURL   string `json:"webhook_url"`
	N8NBaseURL   string `json:"n8n_base_url"`
	Message      string `json:"message"`
}

// LeadAutomationTriggerResponse returns normalized execution data from n8n.
type LeadAutomationTriggerResponse struct {
	Triggered          bool        `json:"triggered"`
	UpstreamStatus     int         `json:"upstream_status"`
	ExecutedWebhookURL string      `json:"executed_webhook_url"`
	Result             interface{} `json:"result,omitempty"`
}
