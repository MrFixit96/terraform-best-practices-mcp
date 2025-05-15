// pkg/mcp/server.go
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Server represents an MCP server that handles requests from AI assistants
type Server struct {
	tools     map[string]Tool
	resources ResourceProvider
	mu        sync.RWMutex
	logger    Logger
}

// Logger defines a simple interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewServer creates a new MCP server
func NewServer(resources ResourceProvider, logger Logger) *Server {
	return &Server{
		tools:     make(map[string]Tool),
		resources: resources,
		logger:    logger,
	}
}

// AddTool registers a tool with the server
func (s *Server) AddTool(tool Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	toolName := tool.Name()
	s.tools[toolName] = tool
	s.logger.Info("Registered tool", "name", toolName)
}

// HandleRequest processes an MCP request and returns a response
func (s *Server) HandleRequest(ctx context.Context, req Request) Response {
	s.logger.Debug("Handling request", "id", req.ID, "tool", req.Tool)
	
	s.mu.RLock()
	tool, exists := s.tools[req.Tool]
	s.mu.RUnlock()
	
	if !exists {
		s.logger.Error("Tool not found", "tool", req.Tool)
		return Response{
			ID:     req.ID,
			Status: "error",
			Error: &ErrorDetail{
				Code:    "tool_not_found",
				Message: fmt.Sprintf("Tool %q not found", req.Tool),
			},
		}
	}
	
	result, err := tool.Execute(ctx, req.Arguments)
	if err != nil {
		s.logger.Error("Tool execution failed", "tool", req.Tool, "error", err)
		return Response{
			ID:     req.ID,
			Status: "error",
			Error: &ErrorDetail{
				Code:    "execution_error",
				Message: err.Error(),
			},
		}
	}
	
	s.logger.Debug("Request completed successfully", "id", req.ID)
	return Response{
		ID:     req.ID,
		Status: "success",
		Result: result,
	}
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode request", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	
	ctx := r.Context()
	resp := s.HandleRequest(ctx, req)
	
	w.Header().Set("Content-Type", "application/json")
	if resp.Status == "error" {
		w.WriteHeader(http.StatusInternalServerError)
	}
	
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("Failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}