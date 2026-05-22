package simulator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/jwt"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleManager  Role = "manager"
	RoleStaff    Role = "staff"
	RoleViewer   Role = "viewer"
	RoleNoAccess Role = "no_access"
)

type ApiClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Tokens     map[Role]string
	coaCache   map[string]string // map[code]uuid
}

func NewApiClient() *ApiClient {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/api/v1"
	}

	client := &ApiClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Tokens:     make(map[Role]string),
		coaCache:   make(map[string]string),
	}

	// Generate localized tokens using the application's JWT service
	// Using hardcoded UUIDs matching the seeders to avoid circular/CGO dependencies
	client.Tokens[RoleAdmin] = generateSimulatorToken("ee0b14e0-c651-4814-a5a2-e7398f81dcf4", string(RoleAdmin))
	client.Tokens[RoleManager] = generateSimulatorToken("2b83a042-45e3-46d4-a957-3f8d22384784", string(RoleManager))
	client.Tokens[RoleStaff] = generateSimulatorToken("51d45763-8a39-4d6b-b4dc-7d2e057865c6", string(RoleStaff))
	client.Tokens[RoleViewer] = generateSimulatorToken("98f45a2d-3c21-41b5-82e6-1234567890ab", string(RoleViewer))
	client.Tokens[RoleNoAccess] = "invalid_or_expired_token_for_negative_testing"

	return client
}

func generateSimulatorToken(userID, roleCode string) string {
	if config.AppConfig == nil {
		_ = config.Load()
	}
	if config.AppConfig == nil {
		return ""
	}

	accessSecret := config.AppConfig.JWT.AccessSecretKey
	if accessSecret == "" {
		accessSecret = config.AppConfig.JWT.SecretKey
	}

	opts := jwt.Options{
		AccessSecretKey: accessSecret,
		Issuer:          config.AppConfig.JWT.Issuer,
		AccessTokenTTL:  1 * time.Hour,
	}
	manager := jwt.NewJWTManager(opts)

	// QA simulator tokens are platform-scoped (no tenant isolation in automated tests)
	token, _ := manager.GenerateAccessToken(userID, "qa.simulator@example.com", roleCode, "")
	return token
}

func (c *ApiClient) GetCOAID(code string) string {
	if len(c.coaCache) == 0 {
		c.FetchCOAs()
	}
	return c.coaCache[code]
}

// GetAnyCOAPair returns any two distinct COA UUIDs from the cache as a fallback.
func (c *ApiClient) GetAnyCOAPair() (string, string) {
	if len(c.coaCache) == 0 {
		c.FetchCOAs()
	}
	var first, second string
	for _, id := range c.coaCache {
		if first == "" {
			first = id
		} else if id != first && second == "" {
			second = id
			break
		}
	}
	return first, second
}

func (c *ApiClient) FetchCOAs() {
	status, body, err := c.Request("GET", "/finance/journal-entries/form-data", nil, RoleAdmin)
	if err != nil || status != http.StatusOK {
		return
	}

	data, ok := body["data"].(map[string]interface{})
	if !ok {
		return
	}

	coas, ok := data["chart_of_accounts"].([]interface{})
	if !ok {
		return
	}

	for _, coa := range coas {
		m, ok := coa.(map[string]interface{})
		if !ok {
			continue
		}
		cID, _ := m["id"].(string)
		cCode, _ := m["code"].(string)
		if cID != "" && cCode != "" {
			c.coaCache[cCode] = cID
		}
	}
}

// Request performs an HTTP request and returns the status, body, and error
func (c *ApiClient) Request(method, endpoint string, body interface{}, role Role) (int, map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	fullURL := c.BaseURL + endpoint
	if !strings.HasPrefix(endpoint, "/") {
		fullURL = c.BaseURL + "/" + endpoint
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Inject CSRF token to pass double-submit cookie validation in GIMS
	mockCsrf := "simulator-qa-csrf-token-bypass"
	req.AddCookie(&http.Cookie{Name: "gims_csrf_token", Value: mockCsrf})
	req.Header.Set("X-CSRF-Token", mockCsrf)

	// Inject JWT token
	if token, ok := c.Tokens[role]; ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return 0, nil, fmt.Errorf("request timeout: %w", err)
		}
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			// Not a JSON response, maybe a 500 error page
			return resp.StatusCode, nil, fmt.Errorf("failed to unmarshal JSON (status %d): %s", resp.StatusCode, string(respBody))
		}
	}

	return resp.StatusCode, result, nil
}
