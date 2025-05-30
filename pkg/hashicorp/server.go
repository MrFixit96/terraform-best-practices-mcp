// pkg/hashicorp/server.go
package hashicorp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"terraform-mcp-server/pkg/hashicorp/tfdocs"
	"terraform-mcp-server/pkg/mcp"
)

// Server represents a HashiCorp MCP server implementation
type Server struct {
	mcpServer        *mcp.Server
	docIndexer       *tfdocs.Indexer
	patternRepo      *tfdocs.PatternRepository
	resourceProvider *tfdocs.ResourceProvider
	validationEngine *tfdocs.ValidationEngine
	logger           Logger
}

// Logger defines a simple interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// Config represents the configuration for the HashiCorp MCP server
type Config struct {
	DocSourcePath    string
	PatternPath      string
	UpdateInterval   time.Duration
	AuthoritySources []string
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		DocSourcePath:    "data/docs",
		PatternPath:      "data/patterns",
		UpdateInterval:   24 * time.Hour,
		AuthoritySources: tfdocs.DefaultAuthoritySources,
	}
}

// NewServer creates a new HashiCorp MCP server
func NewServer(config Config, logger Logger) (*Server, error) {
	// Create the logger if not provided
	if logger == nil {
		logger = &DefaultLogger{
			Logger: log.New(os.Stdout, "terraform-mcp: ", log.LstdFlags),
		}
	}

	// Create directories if they don't exist
	if err := os.MkdirAll(config.DocSourcePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create doc source directory: %w", err)
	}
	
	if err := os.MkdirAll(config.PatternPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create pattern directory: %w", err)
	}

	// Create core components
	// Pass authority sources to the indexer if provided
	indexerOptions := []tfdocs.IndexerOption{
		tfdocs.WithUpdateInterval(config.UpdateInterval),
	}
	
	// Add authority sources if provided
	if len(config.AuthoritySources) > 0 {
		indexerOptions = append(indexerOptions, tfdocs.WithAuthoritySources(config.AuthoritySources))
	}
	
	docIndexer := tfdocs.NewIndexer(
		config.DocSourcePath, 
		logger, 
		indexerOptions...,
	)
	
	patternRepo := tfdocs.NewPatternRepository(config.PatternPath, logger)
	resourceProvider := tfdocs.NewResourceProvider(docIndexer, logger)
	validationEngine := tfdocs.NewValidationEngine(docIndexer, logger)
	
	// Create MCP server
	mcpServer := mcp.NewServer(logger)
	
	return &Server{
		mcpServer:        mcpServer,
		docIndexer:       docIndexer,
		patternRepo:      patternRepo,
		resourceProvider: resourceProvider,
		validationEngine: validationEngine,
		logger:           logger,
	}, nil
}

// Initialize initializes the server components
func (s *Server) Initialize(ctx context.Context) error {
	s.logger.Info("Initializing HashiCorp MCP server")
	
	// Initialize the documentation indexer
	if err := s.docIndexer.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize documentation indexer: %w", err)
	}
	
	// Initialize the pattern repository
	if err := s.patternRepo.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize pattern repository: %w", err)
	}
	
	// Register the tools
	s.registerTools()
	
	s.logger.Info("HashiCorp MCP server initialized")
	return nil
}

// registerTools registers the MCP tools
func (s *Server) registerTools() {
	// Register the documentation tools
	s.mcpServer.AddTool(NewGetBestPracticesTool(s.docIndexer, s.resourceProvider, s.logger))
	s.mcpServer.AddTool(NewGetModuleStructureTool(s.docIndexer, s.resourceProvider, s.logger))
	s.mcpServer.AddTool(NewGetPatternTemplateTool(s.patternRepo, s.logger))
	s.mcpServer.AddTool(NewValidateConfigurationTool(s.validationEngine, s.logger))
	s.mcpServer.AddTool(NewSuggestImprovementsTool(s.validationEngine, s.logger))
}

// AddTool registers a tool with the server
func (s *Server) AddTool(tool mcp.Tool) {
	s.mcpServer.AddTool(tool)
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mcpServer.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe(addr string) error {
	s.logger.Info("Starting HTTP server", "addr", addr)
	return http.ListenAndServe(addr, s)
}

// DefaultLogger is a simple logger implementation
type DefaultLogger struct {
	*log.Logger
}

// Info logs an info message
func (l *DefaultLogger) Info(msg string, fields ...interface{}) {
	if len(fields) == 0 {
		l.Printf("INFO: %s", msg)
		return
	}
	
	l.Printf("INFO: %s %v", msg, fields)
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	if len(fields) == 0 {
		l.Printf("ERROR: %s", msg)
		return
	}
	
	l.Printf("ERROR: %s %v", msg, fields)
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, fields ...interface{}) {
	if len(fields) == 0 {
		l.Printf("DEBUG: %s", msg)
		return
	}
	
	l.Printf("DEBUG: %s %v", msg, fields)
}
</content>
