package tools

import (
	"fmt"
	"strings"
	"sync"
)

// Registry holds all registered tools and provides lookup, filtering, and
// system prompt generation. Mirrors Claude Code's `getAllBaseTools()` with
// dynamic filtering via `filterToolsByDenyRules()`.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool // keyed by tool name
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry. Panics on duplicate names.
func (r *Registry) Register(t Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[t.Name()]; exists {
		panic(fmt.Sprintf("duplicate tool registration: %s", t.Name()))
	}
	r.tools[t.Name()] = t
}

// Get returns a tool by name, or nil if not found.
func (r *Registry) Get(name string) Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tools[name]
}

// All returns every registered tool.
func (r *Registry) All() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

// FilterByPermissions returns only the tools the user is allowed to see,
// following Claude Code's pattern of removing denied tools BEFORE sending to the model.
// This prevents the model from ever attempting to call tools the user cannot access.
func (r *Registry) FilterByPermissions(userPermissions map[string]bool, isAdmin bool) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		if isAdmin {
			result = append(result, t)
			continue
		}
		perm := t.Permission()
		if perm == "" || userPermissions[perm] {
			result = append(result, t)
		}
	}
	return result
}

// FilterByModule returns tools belonging to a specific ERP module.
func (r *Registry) FilterByModule(module string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Tool, 0)
	for _, t := range r.tools {
		if strings.EqualFold(t.Module(), module) {
			result = append(result, t)
		}
	}
	return result
}

// FormatForPrompt generates the tool description block for the LLM system prompt.
// Each tool is described with its name, description, module, category, and parameter schema.
// This is the equivalent of Claude Code's tool schema injection into the system prompt.
func (r *Registry) FormatForPrompt(tools []Tool) string {
	if len(tools) == 0 {
		return "No tools available."
	}

	var sb strings.Builder
	for i, t := range tools {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(fmt.Sprintf("### %s\n", t.Name()))
		sb.WriteString(fmt.Sprintf("- **Description**: %s\n", t.Description()))
		sb.WriteString(fmt.Sprintf("- **Module**: %s\n", t.Module()))
		sb.WriteString(fmt.Sprintf("- **Category**: %s\n", t.Category()))
		if t.NeedsConfirmation() {
			sb.WriteString("- **Requires Confirmation**: yes\n")
		}

		params := t.Parameters()
		if len(params) > 0 {
			sb.WriteString("- **Parameters**:\n")
			for _, p := range params {
				required := ""
				if p.Required {
					required = " (REQUIRED)"
				}
				enumStr := ""
				if len(p.Enum) > 0 {
					enumStr = fmt.Sprintf(" [values: %s]", strings.Join(p.Enum, ", "))
				}
				sb.WriteString(fmt.Sprintf("  - `%s` (%s)%s: %s%s\n", p.Name, p.Type, required, p.Description, enumStr))
			}
		}
	}
	return sb.String()
}

// Count returns the total number of registered tools.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}
