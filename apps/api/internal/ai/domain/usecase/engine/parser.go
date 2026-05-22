package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gilabs/gims/api/internal/ai/domain/usecase/tools"
	"github.com/google/uuid"
)

const (
	toolCallOpen  = "<tool_call>"
	toolCallClose = "</tool_call>"
)

var legacyToolOpenTagPattern = regexp.MustCompile(`<([a-z][a-z0-9_]+)>`)

var trailingToolCallLabelPattern = regexp.MustCompile(`(?is)\*\*tool call\*\*\s*$`)

var trailingCodeFencePattern = regexp.MustCompile("(?is)```(?:json)?\\s*$")

// ParseToolCall extracts a tool call from the LLM response.
// Returns:
//   - textBefore: text content before the tool call
//   - toolCall: parsed tool call (nil if no tool call found)
//   - textAfter: text content after the tool call
//
// The LLM format is:
//
//	Some reasoning text...
//	<tool_call>
//	{"name": "query_stock", "parameters": {"low_stock": true}}
//	</tool_call>
//	Optional text after...
//
// Legacy compatibility:
// Some models may emit <create_sales_order> ... </create_sales_order> instead of
// the canonical <tool_call> wrapper. This parser accepts both formats.
func ParseToolCall(content string) (textBefore string, toolCall *tools.ToolCall, textAfter string) {
	if before, call, after, matched := parseCanonicalToolCall(content); matched {
		return before, call, after
	}

	if before, call, after, matched := parseLegacyToolCall(content); matched {
		return before, call, after
	}

	return content, nil, ""
}

func parseCanonicalToolCall(content string) (string, *tools.ToolCall, string, bool) {
	openIdx := strings.Index(content, toolCallOpen)
	if openIdx == -1 {
		return "", nil, "", false
	}

	closeIdx := strings.Index(content, toolCallClose)
	if closeIdx == -1 || closeIdx <= openIdx {
		// Incomplete tool call (stream truncated before </tool_call>).
		// Return only the text before the opening tag so callers never
		// receive partial XML fragments as visible content.
		return cleanToolCallPrelude(content[:openIdx]), nil, "", true
	}

	textBefore := cleanToolCallPrelude(content[:openIdx])
	textAfter := strings.TrimSpace(content[closeIdx+len(toolCallClose):])

	jsonStr := strings.TrimSpace(content[openIdx+len(toolCallOpen) : closeIdx])
	toolCall := buildToolCallFromJSON(jsonStr, "")

	return textBefore, toolCall, textAfter, true
}

func parseLegacyToolCall(content string) (string, *tools.ToolCall, string, bool) {
	openIdx, openTag, toolName, found := findLegacyToolOpenTag(content)
	if !found {
		return "", nil, "", false
	}

	afterOpenIdx := openIdx + len(openTag)
	closeTag := fmt.Sprintf("</%s>", toolName)
	closeRelIdx := strings.Index(content[afterOpenIdx:], closeTag)
	if closeRelIdx == -1 {
		// Legacy tool tag without a closing marker means the model stopped mid-output.
		// Keep only the user-facing explanation before the tool section.
		return cleanToolCallPrelude(content[:openIdx]), nil, "", true
	}

	body := content[afterOpenIdx : afterOpenIdx+closeRelIdx]
	textAfter := strings.TrimSpace(content[afterOpenIdx+closeRelIdx+len(closeTag):])
	toolCall := buildToolCallFromJSON(body, toolName)

	return cleanToolCallPrelude(content[:openIdx]), toolCall, textAfter, true
}

func findLegacyToolOpenTag(content string) (int, string, string, bool) {
	matches := legacyToolOpenTagPattern.FindAllStringSubmatchIndex(content, -1)
	for _, m := range matches {
		if len(m) < 4 {
			continue
		}

		openStart := m[0]
		openEnd := m[1]
		nameStart := m[2]
		nameEnd := m[3]
		if openStart < 0 || openEnd > len(content) || nameStart < 0 || nameEnd > len(content) {
			continue
		}

		name := strings.ToLower(content[nameStart:nameEnd])
		if name == "tool_call" || !isLikelyToolName(name) {
			continue
		}

		return openStart, content[openStart:openEnd], name, true
	}

	return 0, "", "", false
}

func isLikelyToolName(name string) bool {
	if !strings.Contains(name, "_") {
		return false
	}

	prefixes := []string{"create_", "update_", "delete_", "query_", "list_", "approve_", "reject_", "generate_"}
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}

	return false
}

func buildToolCallFromJSON(content string, fallbackName string) *tools.ToolCall {
	jsonStr, complete := extractJSONObject(content)
	if jsonStr == "" || !complete {
		return nil
	}

	var rawCall struct {
		Name       string                 `json:"name"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawCall); err != nil {
		return nil
	}

	if strings.TrimSpace(rawCall.Name) == "" {
		rawCall.Name = fallbackName
	}

	if strings.TrimSpace(rawCall.Name) == "" {
		return nil
	}

	toolCall := &tools.ToolCall{
		ID:         uuid.New().String(),
		Name:       rawCall.Name,
		Parameters: rawCall.Parameters,
	}

	if toolCall.Parameters == nil {
		toolCall.Parameters = make(map[string]interface{})
	}

	return toolCall
}

func extractJSONObject(content string) (string, bool) {
	start := strings.Index(content, "{")
	if start == -1 {
		return "", false
	}

	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(content); i++ {
		ch := content[i]

		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(content[start : i+1]), true
			}
		}
	}

	return strings.TrimSpace(content[start:]), false
}

func cleanToolCallPrelude(text string) string {
	cleaned := strings.TrimSpace(text)
	cleaned = trailingCodeFencePattern.ReplaceAllString(cleaned, "")
	cleaned = trailingToolCallLabelPattern.ReplaceAllString(cleaned, "")
	return strings.TrimSpace(cleaned)
}

// ContainsToolCall checks if the content has a tool call without fully parsing it.
func ContainsToolCall(content string) bool {
	_, toolCall, _ := ParseToolCall(content)
	return toolCall != nil
}
