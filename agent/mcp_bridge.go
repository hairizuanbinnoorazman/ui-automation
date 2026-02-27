package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hairizuan-noorazman/ui-automation/logger"
)

// MCPBridge connects to a Playwright MCP server and bridges tool calls.
type MCPBridge struct {
	serverURL string
	logger    logger.Logger
	connected bool
}

// NewMCPBridge creates a new MCP bridge.
func NewMCPBridge(serverURL string, log logger.Logger) *MCPBridge {
	return &MCPBridge{
		serverURL: serverURL,
		logger:    log,
	}
}

// Connect establishes connection to the MCP server.
func (b *MCPBridge) Connect(ctx context.Context) error {
	b.logger.Info(ctx, "connecting to MCP server", map[string]interface{}{
		"url": b.serverURL,
	})
	// TODO: Implement SSE client connection using mcp-go
	b.connected = true
	return nil
}

// Close closes the MCP connection.
func (b *MCPBridge) Close() error {
	b.connected = false
	return nil
}

// ListTools queries the MCP server for available tools.
func (b *MCPBridge) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	if !b.connected {
		return nil, fmt.Errorf("MCP bridge not connected")
	}
	// TODO: Query MCP server for tools and convert to ToolDefinition
	return nil, nil
}

// CallTool forwards a tool call to the MCP server.
func (b *MCPBridge) CallTool(ctx context.Context, name string, input json.RawMessage) (*ToolResult, error) {
	if !b.connected {
		return nil, fmt.Errorf("MCP bridge not connected")
	}
	// TODO: Forward tool call to MCP server
	return &ToolResult{
		Content: "tool call not yet implemented",
	}, nil
}

// ToolDefinition represents a tool available from the MCP server.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// ToolResult represents the result of a tool call.
type ToolResult struct {
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
	Screenshot []byte `json:"-"` // Raw screenshot data if tool returns one
}
