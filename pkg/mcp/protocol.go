// pkg/mcp/protocol.go
package mcp

import (
	"context"
	"encoding/json"
)

// Request represents an MCP request from an AI assistant
type Request struct {
	ID        string          `json:"id"`
	Tool      string          `json:"tool"`
	Arguments json.RawMessage `json:"arguments"`
}

// Response represents an MCP response to an AI assistant
type Response struct {
	ID     string          `json:"id"`
	Status string          `json:"status"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ErrorDetail    `json:"error,omitempty"`
}

// ErrorDetail contains error details for an MCP response
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Tool defines the interface for an MCP tool implementation
type Tool interface {
	// Name returns the name of the tool
	Name() string
	
	// Execute executes the tool with the given arguments
	Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
}

// ResourceProvider defines the interface for an MCP resource provider
type ResourceProvider interface {
	// GetResource returns a resource by its URI
	GetResource(ctx context.Context, uri string) (json.RawMessage, error)
	
	// ListResources lists resources matching a pattern
	ListResources(ctx context.Context, pattern string) ([]string, error)
}