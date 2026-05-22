// Package context implements Claude Code's multi-layer context assembly pipeline.
// The system prompt is built from modular sections with explicit cache boundaries:
//   - STATIC sections (identity, rules, output format) are cacheable across sessions
//   - DYNAMIC sections (tools, user context, ERP state) are session-specific
package context

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/ai/domain/usecase/tools"
	"github.com/gilabs/gims/api/internal/core/apptime"
)

// SectionScope determines caching behavior (mirrors Claude Code's cache boundaries).
type SectionScope string

const (
	// ScopeGlobal sections are stable across sessions and can be cached aggressively.
	ScopeGlobal SectionScope = "global"
	// ScopeSession sections are session-specific and must be rebuilt per conversation.
	ScopeSession SectionScope = "session"

	// DynamicBoundary marks the transition from cacheable to non-cacheable prompt sections.
	DynamicBoundary = "__SYSTEM_PROMPT_DYNAMIC_BOUNDARY__"
)

// PromptSection represents a single logical block of the system prompt.
type PromptSection struct {
	ID      string
	Scope   SectionScope
	Content string
}

// UserContext holds session-specific context that changes per request.
type UserContext struct {
	UserID          string
	UserName        string
	UserRole        string
	IsAdmin         bool
	Permissions     map[string]bool
	Locale          string // "id" or "en"
	CompanyName     string
	CompanyTimezone string
	RecentActions   []RecentAction
}

// RecentAction summarizes a recent AI action for context continuity.
type RecentAction struct {
	Intent    string
	Module    string
	Status    string
	CreatedAt time.Time
}

// Builder assembles the multi-section system prompt and user context.
type Builder struct {
	mu           sync.RWMutex
	cachedStatic string
	cacheExpiry  time.Time
}

// NewBuilder creates a new context builder.
func NewBuilder(_ *tools.Registry) *Builder {
	return &Builder{}
}

// BuildSystemPrompt assembles the complete system prompt from modular sections.
// Static sections are cached for 10 minutes; dynamic sections rebuild every call.
func (b *Builder) BuildSystemPrompt(availableTools []tools.Tool, userCtx *UserContext, toolRegistry *tools.Registry) []string {
	sections := make([]string, 0, 8)

	// --- STATIC SECTIONS (cacheable) ---
	sections = append(sections, b.getOrCacheStatic())

	// --- DYNAMIC BOUNDARY ---
	sections = append(sections, DynamicBoundary)

	// --- DYNAMIC SECTIONS (session-specific) ---
	sections = append(sections, b.buildToolSection(availableTools, toolRegistry))
	sections = append(sections, b.buildUserContextSection(userCtx))
	sections = append(sections, b.buildDatetimeSection())
	sections = append(sections, b.buildRecentActionsSection(userCtx))

	return sections
}

// BuildFlatSystemPrompt returns the system prompt as a single string (for non-caching APIs).
func (b *Builder) BuildFlatSystemPrompt(availableTools []tools.Tool, userCtx *UserContext, toolRegistry *tools.Registry) string {
	sections := b.BuildSystemPrompt(availableTools, userCtx, toolRegistry)

	// Filter out the boundary marker
	parts := make([]string, 0, len(sections))
	for _, s := range sections {
		if s != DynamicBoundary && strings.TrimSpace(s) != "" {
			parts = append(parts, s)
		}
	}

	return strings.Join(parts, "\n\n")
}

// getOrCacheStatic returns the cached static prompt or rebuilds it.
func (b *Builder) getOrCacheStatic() string {
	b.mu.RLock()
	if b.cachedStatic != "" && apptime.Now().Before(b.cacheExpiry) {
		defer b.mu.RUnlock()
		return b.cachedStatic
	}
	b.mu.RUnlock()

	// Double-check after acquiring write lock
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cachedStatic != "" && apptime.Now().Before(b.cacheExpiry) {
		return b.cachedStatic
	}

	b.cachedStatic = buildStaticSections()
	b.cacheExpiry = apptime.Now().Add(10 * time.Minute)
	return b.cachedStatic
}

// buildStaticSections constructs the stable, cacheable portion of the system prompt.
func buildStaticSections() string {
	var sb strings.Builder

	// Section 1: Identity & Persona
	sb.WriteString(sectionIdentity)
	sb.WriteString("\n\n")

	// Section 2: Core System Rules
	sb.WriteString(sectionSystemRules)
	sb.WriteString("\n\n")

	// Section 3: Tool Usage Protocol
	sb.WriteString(sectionToolProtocol)
	sb.WriteString("\n\n")

	// Section 4: Output Format & Style
	sb.WriteString(sectionOutputStyle)
	sb.WriteString("\n\n")

	// Section 5: Safety & Security
	sb.WriteString(sectionSecurity)
	sb.WriteString("\n\n")

	// Section 6: Payload Templates (correct JSON format for complex CREATE tools)
	sb.WriteString(sectionPayloadTemplates)
	sb.WriteString("\n\n")

	// Section 7: Application Navigation (links to frontend pages)
	sb.WriteString(sectionNavigation)

	return sb.String()
}

// buildToolSection generates the dynamic tool catalog for the LLM.
func (b *Builder) buildToolSection(availableTools []tools.Tool, registry *tools.Registry) string {
	if len(availableTools) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Available Tools\n\n")
	sb.WriteString("You have access to the following tools to help the user. ")
	sb.WriteString("Choose the most appropriate tool based on the user's request. ")
	sb.WriteString("If no tool matches, respond conversationally.\n\n")
	sb.WriteString(registry.FormatForPrompt(availableTools))

	return sb.String()
}

// buildUserContextSection injects user-specific context.
func (b *Builder) buildUserContextSection(userCtx *UserContext) string {
	if userCtx == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Current User Context\n\n")

	if userCtx.UserName != "" {
		sb.WriteString(fmt.Sprintf("- **User**: %s\n", userCtx.UserName))
	}
	if userCtx.UserRole != "" {
		sb.WriteString(fmt.Sprintf("- **Role**: %s\n", userCtx.UserRole))
	}
	if userCtx.IsAdmin {
		sb.WriteString("- **Access Level**: Administrator (full access)\n")
	}
	if userCtx.CompanyName != "" {
		sb.WriteString(fmt.Sprintf("- **Company**: %s\n", userCtx.CompanyName))
	}
	if userCtx.Locale != "" {
		lang := "Bahasa Indonesia"
		if userCtx.Locale == "en" {
			lang = "English"
		}
		sb.WriteString(fmt.Sprintf("- **Preferred Language**: %s\n", lang))
	}

	return sb.String()
}

// buildDatetimeSection injects the current datetime for temporal awareness.
func (b *Builder) buildDatetimeSection() string {
	now := apptime.Now()
	return fmt.Sprintf("## Current Datetime\n\n- **Now**: %s (WIB)\n- **Year**: %d\n- **Date**: %s",
		now.Format("2006-01-02 15:04:05"),
		now.Year(),
		now.Format("Monday, 02 January 2006"),
	)
}

// buildRecentActionsSection provides context about recent actions for continuity.
func (b *Builder) buildRecentActionsSection(userCtx *UserContext) string {
	if userCtx == nil || len(userCtx.RecentActions) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Recent Actions (for context continuity)\n\n")

	for _, a := range userCtx.RecentActions {
		sb.WriteString(fmt.Sprintf("- [%s] %s (%s) → %s\n",
			a.CreatedAt.Format("15:04"),
			a.Intent, a.Module, a.Status,
		))
	}

	return sb.String()
}
