package engine_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gilabs/gims/api/internal/ai/domain/usecase/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseToolCall_SalesOrderJSON_HappyPath verifies ParseToolCall handles a
// create_sales_order payload where items is a proper JSON array.
func TestParseToolCall_SalesOrderJSON_HappyPath(t *testing.T) {
	llmResponse := "Saya akan membuat sales order.\n<tool_call>\n" +
		`{"name":"create_sales_order","parameters":{"customer_name":"Kimia Farma","order_date":"2026-04-10","items":[{"product_name":"Paracetamol 500mg","quantity":3,"price":50000,"discount":0}],"notes":""}}` +
		"\n</tool_call>"

	textBefore, toolCall, textAfter := engine.ParseToolCall(llmResponse)

	assert.Contains(t, textBefore, "Saya akan membuat")
	require.NotNil(t, toolCall, "tool call must be parsed")
	assert.Equal(t, "create_sales_order", toolCall.Name)
	assert.Equal(t, "", textAfter)

	// items must be []interface{}, not a string
	rawItems := toolCall.Parameters["items"]
	items, ok := rawItems.([]interface{})
	require.True(t, ok, "items must be []interface{}, got %T", rawItems)
	require.Len(t, items, 1)

	item, ok := items[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Paracetamol 500mg", item["product_name"])
	assert.Equal(t, float64(3), item["quantity"])
	assert.Equal(t, float64(50000), item["price"])
}

// TestParseToolCall_SalesOrderItemsAsString confirms that when the AI (mis)encodes
// items as a JSON string, ParseToolCall returns it as-is; the sanitizer in
// normalizeSalesOrderItems is responsible for converting it later.
func TestParseToolCall_SalesOrderItemsAsString(t *testing.T) {
	rawJSON := `{"name":"create_sales_order","parameters":{"customer_name":"Kimia Farma","order_date":"2026-04-10","items":"[{\"product_name\":\"Paracetamol 500mg\",\"quantity\":3,\"price\":50000}]","notes":""}}`
	llmResponse := "<tool_call>\n" + rawJSON + "\n</tool_call>"

	_, toolCall, _ := engine.ParseToolCall(llmResponse)
	require.NotNil(t, toolCall)
	assert.Equal(t, "create_sales_order", toolCall.Name)

	rawItems := toolCall.Parameters["items"]
	itemsStr, isString := rawItems.(string)
	require.True(t, isString, "items should still be a string before sanitization, got %T", rawItems)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(itemsStr), "["))
}

// TestSanitizeItemsJSONString verifies the sanitization logic (mirrors normalizeSalesOrderItems).
func TestSanitizeItemsJSONString(t *testing.T) {
	params := map[string]interface{}{
		"customer_name": "Kimia Farma",
		"order_date":    "2026-04-10",
		"items":         `[{"product_name":"Paracetamol 500mg","quantity":3,"price":50000,"discount":0}]`,
		"notes":         "",
	}

	sanitizeItemsField(t, params)

	items, ok := params["items"].([]interface{})
	require.True(t, ok, "after sanitization items must be []interface{}")
	require.Len(t, items, 1)

	item, ok := items[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Paracetamol 500mg", item["product_name"])
	assert.Equal(t, float64(3), item["quantity"])
	assert.Equal(t, float64(50000), item["price"])
}

// TestSanitizeItemsMalformedString verifies a malformed items string is removed from
// params to prevent a downstream unmarshal panic.
func TestSanitizeItemsMalformedString(t *testing.T) {
	params := map[string]interface{}{
		"customer_name": "Kimia Farma",
		"items":         "not valid json at all",
	}

	sanitizeItemsField(t, params)

	_, exists := params["items"]
	assert.False(t, exists, "malformed items string should be removed from params")
}

// TestSanitizeItemsAlreadyArray verifies that a properly-typed []interface{} passes
// through the sanitizer unchanged.
func TestSanitizeItemsAlreadyArray(t *testing.T) {
	params := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"product_name": "Produk A", "quantity": float64(2), "price": float64(10000)},
		},
	}

	sanitizeItemsField(t, params)

	items, ok := params["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 1)
}

// TestBuildMarshalRoundTrip verifies the params->JSON->struct round-trip succeeds
// when items is a proper []interface{} (as produced after sanitization).
func TestBuildMarshalRoundTrip(t *testing.T) {
	params := map[string]interface{}{
		"customer_name": "Kimia Farma",
		"order_date":    "2026-04-10",
		"items": []interface{}{
			map[string]interface{}{
				"product_name": "Paracetamol 500mg",
				"quantity":     float64(3),
				"price":        float64(50000),
				"discount":     float64(0),
			},
		},
		"notes": "",
	}

	paramJSON, err := json.Marshal(params)
	require.NoError(t, err)

	var dest struct {
		CustomerName string `json:"customer_name"`
		OrderDate    string `json:"order_date"`
		Items        []struct {
			ProductName string  `json:"product_name"`
			Quantity    float64 `json:"quantity"`
			Price       float64 `json:"price"`
			Discount    float64 `json:"discount"`
		} `json:"items"`
		Notes string `json:"notes"`
	}

	err = json.Unmarshal(paramJSON, &dest)
	require.NoError(t, err, "unmarshal must succeed — this is the shape of CreateSalesOrderRequest")

	assert.Equal(t, "Kimia Farma", dest.CustomerName)
	require.Len(t, dest.Items, 1)
	assert.Equal(t, "Paracetamol 500mg", dest.Items[0].ProductName)
	assert.Equal(t, float64(3), dest.Items[0].Quantity)
	assert.Equal(t, float64(50000), dest.Items[0].Price)
}

// TestBuildMarshalRoundTrip_StringItems demonstrates that WITHOUT sanitization the
// JSON->struct unmarshal fails — confirming why the sanitizer is critical.
func TestBuildMarshalRoundTrip_StringItems(t *testing.T) {
	params := map[string]interface{}{
		"customer_name": "Kimia Farma",
		"order_date":    "2026-04-10",
		"items":         `[{"product_name":"Paracetamol 500mg","quantity":3,"price":50000}]`,
		"notes":         "",
	}

	paramJSON, err := json.Marshal(params)
	require.NoError(t, err)

	var dest struct {
		Items []struct {
			ProductName string  `json:"product_name"`
			Quantity    float64 `json:"quantity"`
			Price       float64 `json:"price"`
		} `json:"items"`
	}

	err = json.Unmarshal(paramJSON, &dest)
	assert.Error(t, err, "WITHOUT sanitization the unmarshal MUST fail — this confirms the bug")
	assert.Contains(t, err.Error(), "cannot unmarshal string")
}

// TestParseToolCall_TruncatedStreamNeverLeaksXML verifies that when the LLM stream
// ends before </tool_call> (e.g. max_tokens reached mid-call), ParseToolCall returns
// only the text that appeared BEFORE the opening tag — never raw XML fragments.
func TestParseToolCall_TruncatedStreamNeverLeaksXML(t *testing.T) {
	// Simulate a stream that was cut off inside the JSON payload.
	truncated := "Baik, saya akan membuat sales order untuk Anda.\n<tool_call>\n{\"name\": \"create_sales_order"

	textBefore, toolCall, textAfter := engine.ParseToolCall(truncated)

	assert.Nil(t, toolCall, "incomplete tool call must not be parsed")
	assert.Equal(t, "", textAfter)
	// textBefore must be the human-readable text only — no <tool_call> XML.
	assert.NotContains(t, textBefore, "<tool_call>", "raw XML must not appear in textBefore")
	assert.Contains(t, textBefore, "Baik, saya akan membuat")
}

// TestParseToolCall_NoToolCallPassthrough ensures content with no tool call marker
// is returned as-is in textBefore.
func TestParseToolCall_NoToolCallPassthrough(t *testing.T) {
	plain := "Ini adalah respons biasa tanpa tool call."
	textBefore, toolCall, textAfter := engine.ParseToolCall(plain)

	assert.Equal(t, plain, textBefore)
	assert.Nil(t, toolCall)
	assert.Equal(t, "", textAfter)
}

// TestParseToolCall_LegacyTag_HappyPath verifies compatibility with legacy
// tool wrappers like <create_sales_order>...</create_sales_order>.
func TestParseToolCall_LegacyTag_HappyPath(t *testing.T) {
	llmResponse := "Saya akan membuat sales order untuk Anda.\n\n**Tool Call**\n```json\n<create_sales_order>\n" +
		`{"name":"create_sales_order","parameters":{"customer_name":"Kimia Farma","order_date":"2026-04-10","items":[{"product_name":"Paracetamol 500mg","quantity":3,"price":50000}]}}` +
		"\n</create_sales_order>\n```"

	textBefore, toolCall, textAfter := engine.ParseToolCall(llmResponse)

	assert.Equal(t, "Saya akan membuat sales order untuk Anda.", textBefore)
	require.NotNil(t, toolCall)
	assert.Equal(t, "create_sales_order", toolCall.Name)
	assert.Equal(t, "```", textAfter)
}

// TestParseToolCall_TruncatedLegacyTagNeverLeaksXML ensures malformed legacy
// tag output does not leak raw XML/JSON in assistant visible text.
func TestParseToolCall_TruncatedLegacyTagNeverLeaksXML(t *testing.T) {
	truncated := "Saya akan membuat sales order.\n\n**Tool Call**\n```json\n<create_sales_order>\n{\"name\":\"create_sales_order\",\"parameters\":{\"customer_name\":\"Kimia Farma\""

	textBefore, toolCall, textAfter := engine.ParseToolCall(truncated)

	assert.Equal(t, "Saya akan membuat sales order.", textBefore)
	assert.Nil(t, toolCall)
	assert.Equal(t, "", textAfter)
	assert.NotContains(t, textBefore, "<create_sales_order>")
}

// ─── helper ──────────────────────────────────────────────────────────────────

// sanitizeItemsField mirrors the logic in normalizeSalesOrderItems so tests can
// invoke it without accessing the private function.
func sanitizeItemsField(t *testing.T, params map[string]interface{}) {
	t.Helper()
	rawItems, ok := params["items"]
	if !ok || rawItems == nil {
		return
	}
	itemsStr, isString := rawItems.(string)
	if !isString {
		return
	}
	trimmed := strings.TrimSpace(itemsStr)
	var parsed []interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		params["items"] = parsed
	} else {
		delete(params, "items")
	}
}
