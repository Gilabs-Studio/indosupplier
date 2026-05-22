package middleware

import (
	"net/http"
	"strings"
)

// isCRMLeadUpsertWebhookRequest returns true only for the exact CRM lead upsert webhook route.
func isCRMLeadUpsertWebhookRequest(method, path string) bool {
	if !strings.EqualFold(method, http.MethodPost) {
		return false
	}

	normalized := strings.TrimSpace(path)
	normalized = strings.TrimSuffix(normalized, "/")
	return normalized == "/api/v1/crm/leads/upsert"
}
