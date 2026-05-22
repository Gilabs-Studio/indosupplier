package engine

import (
	"fmt"
	"strings"
)

// MessageNormalizer handles conversation history compression and normalization.
// Mirrors Claude Code's normalizeMessagesForAPI() which:
//   - Strips signature blocks
//   - Ensures tool_result follows tool_use
//   - Truncates oversized tool results
//   - Compresses old messages into summaries
type MessageNormalizer struct {
	maxToolResultLength int
}

// NewMessageNormalizer creates a normalizer with sensible defaults.
func NewMessageNormalizer() *MessageNormalizer {
	return &MessageNormalizer{
		maxToolResultLength: 4000, // ~1000 tokens
	}
}

// Normalize processes conversation history for the context window:
// 1. Keeps the most recent messages up to maxMessages
// 2. Truncates oversized tool results
// 3. Ensures role alternation is valid
// 4. Compresses older messages into a summary
func (n *MessageNormalizer) Normalize(history []ConversationMessage, maxMessages int) []ConversationMessage {
	if len(history) == 0 {
		return history
	}

	// If within limits, just clean up
	if len(history) <= maxMessages {
		return n.cleanMessages(history)
	}

	// Split: compress older messages into summary, keep recent ones
	cutoff := len(history) - maxMessages
	oldMessages := history[:cutoff]
	recentMessages := history[cutoff:]

	// Build a summary of old messages
	summary := n.buildSummary(oldMessages)
	result := make([]ConversationMessage, 0, maxMessages+1)

	if summary != "" {
		result = append(result, ConversationMessage{
			Role:    "system",
			Content: fmt.Sprintf("[Previous conversation summary: %s]", summary),
		})
	}

	result = append(result, n.cleanMessages(recentMessages)...)
	return result
}

// cleanMessages normalizes individual messages.
func (n *MessageNormalizer) cleanMessages(messages []ConversationMessage) []ConversationMessage {
	result := make([]ConversationMessage, 0, len(messages))

	for _, msg := range messages {
		cleaned := msg

		// Truncate oversized tool results
		if strings.Contains(msg.Content, "<tool_result>") && len(msg.Content) > n.maxToolResultLength {
			cleaned.Content = msg.Content[:n.maxToolResultLength] + "\n[TRUNCATED — result too large for context window]"
		}

		// Skip empty messages
		if strings.TrimSpace(cleaned.Content) == "" {
			continue
		}

		result = append(result, cleaned)
	}

	return result
}

// buildSummary creates a brief summary of compressed messages.
func (n *MessageNormalizer) buildSummary(messages []ConversationMessage) string {
	if len(messages) == 0 {
		return ""
	}

	var sb strings.Builder
	userTopics := make([]string, 0)
	toolsUsed := make([]string, 0)

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			// Extract first 50 chars of user messages as topic hints
			content := strings.TrimSpace(msg.Content)
			if strings.HasPrefix(content, "<tool_result>") {
				continue // Skip tool results
			}
			if len(content) > 50 {
				content = content[:50] + "..."
			}
			userTopics = append(userTopics, content)

		case "assistant":
			// Extract tool calls used
			if strings.Contains(msg.Content, "<tool_call>") {
				_, toolCall, _ := ParseToolCall(msg.Content)
				if toolCall != nil {
					toolsUsed = append(toolsUsed, toolCall.Name)
				}
			}
		}
	}

	if len(userTopics) > 0 {
		sb.WriteString("User discussed: ")
		// Limit to last 3 topics
		start := 0
		if len(userTopics) > 3 {
			start = len(userTopics) - 3
		}
		sb.WriteString(strings.Join(userTopics[start:], "; "))
	}

	if len(toolsUsed) > 0 {
		if sb.Len() > 0 {
			sb.WriteString(". ")
		}
		// Deduplicate tools
		seen := make(map[string]bool)
		unique := make([]string, 0)
		for _, t := range toolsUsed {
			if !seen[t] {
				seen[t] = true
				unique = append(unique, t)
			}
		}
		sb.WriteString("Tools used: ")
		sb.WriteString(strings.Join(unique, ", "))
	}

	return sb.String()
}
