package xendit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL      = "https://api.xendit.co"
	invoicePath         = "/v2/invoices"
	subscriptionPath    = "/recurring/plans"
	subscriptionAPIPath = "/recurring/subscriptions"
)

// Client is a minimal Xendit API client scoped to the Invoice API.
// Extend with additional resources as needed.
type Client struct {
	secretKey  string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Xendit client.
// secretKey is the Xendit server-side secret key (xnd_production_... or xnd_development_...).
// baseURL is optional; defaults to https://api.xendit.co.
func NewClient(secretKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		secretKey: secretKey,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// IsConfigured returns true when the secret key is set and non-placeholder.
func (c *Client) IsConfigured() bool {
	return c.secretKey != "" &&
		c.secretKey != "xnd_development_your_secret_key_here"
}

// CreateInvoiceRequest holds the fields for creating a Xendit invoice.
type CreateInvoiceRequest struct {
	// ExternalID must be unique per invoice; use the pending registration token.
	ExternalID  string `json:"external_id"`
	Amount      int64  `json:"amount"` // In smallest IDR unit (IDR has no subunits, so 1 IDR = 1)
	PayerEmail  string `json:"payer_email"`
	Description string `json:"description"`
	// SuccessRedirectURL is where Xendit redirects the browser after payment.
	SuccessRedirectURL string `json:"success_redirect_url"`
	// FailureRedirectURL is where Xendit redirects the browser if the user cancels.
	FailureRedirectURL string `json:"failure_redirect_url"`
	// Currency defaults to IDR.
	Currency string `json:"currency"`
	// InvoiceDuration is the number of seconds the invoice stays payable (default 86400 = 24h).
	InvoiceDuration int `json:"invoice_duration"`
}

// CreateInvoiceResponse is the relevant subset of the Xendit invoice object.
type CreateInvoiceResponse struct {
	ID          string    `json:"id"`
	ExternalID  string    `json:"external_id"`
	Status      string    `json:"status"`
	InvoiceURL  string    `json:"invoice_url"`
	Amount      int64     `json:"amount"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	ExpiryDate  time.Time `json:"expiry_date"`
}

// WebhookPayload is the structure Xendit POSTs to your webhook endpoint.
type WebhookPayload struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
	Status     string `json:"status"` // "PAID", "EXPIRED", "SETTLED"
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	PayerEmail string `json:"payer_email"`
	// PaymentMethod and PaidAt are optional on failure callbacks.
	PaymentMethod string     `json:"payment_method,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
}

// CreateInvoice sends a create-invoice request to Xendit and returns the invoice.
func (c *Client) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	if req.InvoiceDuration == 0 {
		req.InvoiceDuration = 86400 // 24 hours
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("xendit: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+invoicePath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("xendit: build request: %w", err)
	}
	// Xendit uses HTTP Basic Auth with the secret key as the username and no password.
	httpReq.SetBasicAuth(c.secretKey, "")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("xendit: http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB cap
	if err != nil {
		return nil, fmt.Errorf("xendit: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("xendit: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result CreateInvoiceResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("xendit: parse response: %w", err)
	}
	return &result, nil
}

// ─── Recurring Subscription API ─────────────────────────────────────────────

// CreateRecurringPlanRequest holds the fields for creating a Xendit recurring plan.
// Reference: https://docs.xendit.co/docs/subscriptions-overview
type CreateRecurringPlanRequest struct {
	ReferenceID      string                   `json:"reference_id"`     // Unique plan reference (e.g. "indosupplier_{tenantID}_{planSlug}")
	CustomerID       string                   `json:"customer_id"`      // Xendit customer ID
	RecurringAction  string                   `json:"recurring_action"` // "PAYMENT"
	Currency         string                   `json:"currency"`         // "IDR"
	Amount           int64                    `json:"amount"`
	PaymentMethods   []RecurringPaymentMethod `json:"payment_methods"`
	Schedule         RecurringSchedule        `json:"schedule"`
	SuccessReturnURL string                   `json:"success_return_url"`
	FailureReturnURL string                   `json:"failure_return_url"`
	Metadata         map[string]interface{}   `json:"metadata,omitempty"`
}

// RecurringPaymentMethod specifies how the recurring charge is collected.
type RecurringPaymentMethod struct {
	Type string `json:"type"` // "CREDIT_CARD", "DEBIT_CARD", "EWALLET", "BANK_TRANSFER"
}

// RecurringSchedule defines the billing frequency.
type RecurringSchedule struct {
	ReferenceID   string `json:"reference_id"`
	Interval      string `json:"interval"`              // "MONTH" or "YEAR"
	IntervalCount int    `json:"interval_count"`        // 1 for monthly, 12 for yearly
	AnchorDate    string `json:"anchor_date,omitempty"` // ISO-8601 start date
}

// RecurringPlanResponse is the Xendit recurring plan object.
type RecurringPlanResponse struct {
	ID          string `json:"id"`
	ReferenceID string `json:"reference_id"`
	Status      string `json:"status"` // "ACTIVE", "INACTIVE", "STOPPED"
	CustomerID  string `json:"customer_id"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
}

// CreateRecurringSubscriptionRequest activates a recurring plan for a customer.
type CreateRecurringSubscriptionRequest struct {
	PlanID     string `json:"plan_id"`
	CustomerID string `json:"customer_id"`
}

// RecurringSubscriptionResponse is the Xendit subscription object.
type RecurringSubscriptionResponse struct {
	ID         string `json:"id"`
	PlanID     string `json:"plan_id"`
	Status     string `json:"status"`
	CustomerID string `json:"customer_id"`
}

// CreateCustomerRequest creates a Xendit customer for recurring billing.
type CreateCustomerRequest struct {
	ReferenceID string `json:"reference_id"`
	GivenNames  string `json:"given_names"`
	Email       string `json:"email"`
	Type        string `json:"type"` // "INDIVIDUAL" or "BUSINESS"
}

// CustomerResponse is the Xendit customer object.
type CustomerResponse struct {
	ID          string `json:"id"`
	ReferenceID string `json:"reference_id"`
	GivenNames  string `json:"given_names"`
	Email       string `json:"email"`
}

// CreateCustomer creates a Xendit customer record for recurring billing.
func (c *Client) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*CustomerResponse, error) {
	return doPost[CustomerResponse](ctx, c, "/customers", req)
}

// ─── Card Tokenization API ───────────────────────────────────────────────────

// CardTokenData is the card token object stored from Xendit.js tokenization.
// The frontend collects card details via Xendit.js, which returns a token_id.
// This token_id is sent to our backend and stored here — never the raw card data.
type CardTokenData struct {
	ID               string `json:"id"`                 // Xendit token ID (xnd_...)
	Status           string `json:"status"`             // "VALID", "INVALID"
	MaskedCardNumber string `json:"masked_card_number"` // e.g. "400000XXXXXX0002"
	CardBrand        string `json:"card_brand"`         // "VISA", "MASTERCARD", etc.
	CardHolderName   string `json:"card_holder_name"`
	ExpiryMonth      int    `json:"expiry_month"`
	ExpiryYear       int    `json:"expiry_year"`
}

// GetCardToken retrieves a card token by its ID to verify validity and card details.
func (c *Client) GetCardToken(ctx context.Context, tokenID string) (*CardTokenData, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/v2/tokens/"+tokenID, nil)
	if err != nil {
		return nil, fmt.Errorf("xendit: build get-token request: %w", err)
	}
	httpReq.SetBasicAuth(c.secretKey, "")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("xendit: http get-token: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("xendit: read get-token response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("xendit: API error %d on get-token: %s", resp.StatusCode, string(respBody))
	}

	var result CardTokenData
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("xendit: parse get-token response: %w", err)
	}
	return &result, nil
}

// CreateRecurringPlan registers a recurring billing plan in Xendit.
func (c *Client) CreateRecurringPlan(ctx context.Context, req CreateRecurringPlanRequest) (*RecurringPlanResponse, error) {
	return doPost[RecurringPlanResponse](ctx, c, subscriptionPath, req)
}

// CancelRecurringPlan deactivates a recurring plan (stops future charges).
func (c *Client) CancelRecurringPlan(ctx context.Context, planID string) (*RecurringPlanResponse, error) {
	return doPost[RecurringPlanResponse](ctx, c, subscriptionPath+"/"+planID+"/deactivate", nil)
}

// doPost is a typed helper that marshals req, sends a POST to path, and unmarshals into T.
func doPost[T any](ctx context.Context, c *Client, path string, req any) (*T, error) {
	var bodyReader *bytes.Reader
	if req != nil {
		b, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("xendit: marshal %s: %w", path, err)
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader([]byte("{}"))
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("xendit: build request %s: %w", path, err)
	}
	httpReq.SetBasicAuth(c.secretKey, "")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("xendit: http %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("xendit: read response %s: %w", path, err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("xendit: API error %d on %s: %s", resp.StatusCode, path, string(respBody))
	}

	var result T
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("xendit: parse response %s: %w", path, err)
	}
	return &result, nil
}
