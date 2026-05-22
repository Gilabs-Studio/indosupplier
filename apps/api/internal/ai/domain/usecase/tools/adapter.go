package tools

import (
	"context"
)

// AdaptedTool wraps an existing action handler function into the Tool interface.
// This bridges the existing ActionExecutor methods to the new tool system without
// requiring a rewrite of all 80+ handlers. Over time, handlers can be migrated to
// native Tool implementations.
type AdaptedTool struct {
	name              string
	description       string
	module            string
	category          string
	permission        string
	needsConfirmation bool
	concurrencySafe   bool
	params            []ToolParameter
	executeFn         func(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (*ToolResult, error)
}

// AdaptedToolConfig holds configuration for creating an adapted tool.
type AdaptedToolConfig struct {
	Name              string
	Description       string
	Module            string
	Category          string
	Permission        string
	NeedsConfirmation bool
	ConcurrencySafe   bool
	Params            []ToolParameter
	ExecuteFn         func(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (*ToolResult, error)
}

func NewAdaptedTool(cfg AdaptedToolConfig) *AdaptedTool {
	return &AdaptedTool{
		name:              cfg.Name,
		description:       cfg.Description,
		module:            cfg.Module,
		category:          cfg.Category,
		permission:        cfg.Permission,
		needsConfirmation: cfg.NeedsConfirmation,
		concurrencySafe:   cfg.ConcurrencySafe,
		params:            cfg.Params,
		executeFn:         cfg.ExecuteFn,
	}
}

func (t *AdaptedTool) Name() string              { return t.name }
func (t *AdaptedTool) Description() string        { return t.description }
func (t *AdaptedTool) Module() string             { return t.module }
func (t *AdaptedTool) Category() string           { return t.category }
func (t *AdaptedTool) Permission() string         { return t.permission }
func (t *AdaptedTool) NeedsConfirmation() bool    { return t.needsConfirmation }
func (t *AdaptedTool) IsConcurrencySafe() bool    { return t.concurrencySafe }
func (t *AdaptedTool) Parameters() []ToolParameter { return t.params }

func (t *AdaptedTool) Execute(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (*ToolResult, error) {
	return t.executeFn(ctx, params, execCtx)
}
