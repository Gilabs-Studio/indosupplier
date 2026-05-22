package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/crm/domain/dto"
)

var (
	// ErrN8NNotConfigured indicates the n8n webhook URL is not configured.
	ErrN8NNotConfigured = errors.New("n8n is not configured")
)

// LeadAutomationUpstreamError wraps non-2xx responses from n8n.
type LeadAutomationUpstreamError struct {
	Status             int
	Body               string
	ExecutedWebhookURL string
}

func (e *LeadAutomationUpstreamError) Error() string {
	return fmt.Sprintf("n8n upstream request failed with status %d", e.Status)
}

// LeadAutomationUsecase defines business operations for n8n lead automation.
type LeadAutomationUsecase interface {
	TestConnection(ctx context.Context) (*dto.LeadAutomationConnectionResponse, error)
	Trigger(ctx context.Context, req dto.LeadAutomationTriggerRequest) (*dto.LeadAutomationTriggerResponse, error)
}

type leadAutomationUsecase struct {
	client      *http.Client
	webhookURL  string
	serperKey   string
	apifyKey    string
	apolloKey   string
}

// NewLeadAutomationUsecase creates a usecase that proxies automation requests to n8n.
func NewLeadAutomationUsecase() LeadAutomationUsecase {
	timeoutSec := 30
	if rawTimeout := strings.TrimSpace(os.Getenv("N8N_TIMEOUT_SEC")); rawTimeout != "" {
		if parsed, err := strconv.Atoi(rawTimeout); err == nil && parsed > 0 {
			timeoutSec = parsed
		}
	}

	return &leadAutomationUsecase{
		client: &http.Client{Timeout: time.Duration(timeoutSec) * time.Second},
		webhookURL: strings.TrimSpace(os.Getenv("N8N_LEADS_WEBHOOK_URL")),
		serperKey: strings.TrimSpace(os.Getenv("SERPER_API_KEY")),
		apifyKey: strings.TrimSpace(os.Getenv("APIFY_API_KEY")),
		apolloKey: strings.TrimSpace(os.Getenv("APOLLO_API_KEY")),
	}
}

func (u *leadAutomationUsecase) TestConnection(ctx context.Context) (*dto.LeadAutomationConnectionResponse, error) {
	webhookURL, err := u.resolveWebhookURL()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, webhookURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reachable := resp.StatusCode >= 200 && resp.StatusCode < 500
	message := "n8n webhook endpoint is reachable"
	if resp.StatusCode == http.StatusNotFound {
		message = "webhook not registered or workflow is inactive"
	}

	return &dto.LeadAutomationConnectionResponse{
		Reachable:  reachable,
		Status:     resp.StatusCode,
		WebhookURL: webhookURL,
		N8NBaseURL: normalizeN8NBaseURL(webhookURL),
		Message:    message,
	}, nil
}

func (u *leadAutomationUsecase) Trigger(ctx context.Context, req dto.LeadAutomationTriggerRequest) (*dto.LeadAutomationTriggerResponse, error) {
	webhookURL, err := u.resolveWebhookURL()
	if err != nil {
		return nil, err
	}

	sourceType := normalizeSourceType(req.Type)

	payload := map[string]interface{}{
		"type":           sourceType,
		"source":         sourceTypeToName(sourceType),
		"keyword":        req.Keyword,
		"city":           req.City,
		"limit":          normalizeLeadLimit(sourceType, req.Limit),
		"lead_source_id": strings.TrimSpace(req.LeadSourceID),
		"erp_base_url":   req.ERPBaseURL,
	}

	u.mergeAPIKeys(payload)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := u.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed interface{}
	if len(rawBody) > 0 {
		_ = json.Unmarshal(rawBody, &parsed)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &LeadAutomationUpstreamError{
			Status:             resp.StatusCode,
			Body:               string(rawBody),
			ExecutedWebhookURL: webhookURL,
		}
	}

	return &dto.LeadAutomationTriggerResponse{
		Triggered:          true,
		UpstreamStatus:     resp.StatusCode,
		ExecutedWebhookURL: webhookURL,
		Result:             parsed,
	}, nil
}

func (u *leadAutomationUsecase) resolveWebhookURL() (string, error) {
	webhookURL := strings.TrimSpace(u.webhookURL)
	if webhookURL == "" {
		return "", ErrN8NNotConfigured
	}

	return webhookURL, nil
}


func normalizeN8NBaseURL(webhookURL string) string {
	parsedURL, err := url.Parse(strings.TrimSpace(webhookURL))
	if err != nil {
		return ""
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return ""
	}

	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
}

func (u *leadAutomationUsecase) mergeAPIKeys(payload map[string]interface{}) {
	if u.serperKey != "" {
		payload["serper_api_key"] = u.serperKey
	}
	if u.apifyKey != "" {
		payload["apify_api_key"] = u.apifyKey
	}
	if u.apolloKey != "" {
		payload["apollo_api_key"] = u.apolloKey
	}
}

func normalizeLeadLimit(sourceType int, requestedLimit int) int {
	if requestedLimit <= 0 {
		requestedLimit = 10
	}

	maxLimit := 20 // Google Maps
	if sourceType == 1 {
		maxLimit = 25 // LinkedIn
	}

	if requestedLimit > maxLimit {
		return maxLimit
	}
	return requestedLimit
}

func normalizeSourceType(sourceType *int) int {
	if sourceType == nil {
		return 0
	}

	if *sourceType == 1 {
		return 1
	}

	return 0
}

func sourceTypeToName(sourceType int) string {
	if sourceType == 1 {
		return "linkedin"
	}

	return "google_maps"
}
