package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-mcp-server/pkg/hashicorp"
	"terraform-mcp-server/pkg/hashicorp/tfdocs"
	"terraform-mcp-server/pkg/mcp"
)

// TestLogger implements a test logger that captures log output
type TestLogger struct {
	Logs   []string
	t      *testing.T
	prefix string
}

func NewTestLogger(t *testing.T, prefix string) *TestLogger {
	return &TestLogger{
		Logs:   []string{},
		t:      t,
		prefix: prefix,
	}
}

func (l *TestLogger) Info(msg string, fields ...interface{}) {
	logMsg := fmt.Sprintf("[INFO] %s: %s %v", l.prefix, msg, fields)
	l.Logs = append(l.Logs, logMsg)
	l.t.Log(logMsg)
}

func (l *TestLogger) Error(msg string, fields ...interface{}) {
	logMsg := fmt.Sprintf("[ERROR] %s: %s %v", l.prefix, msg, fields)
	l.Logs = append(l.Logs, logMsg)
	l.t.Log(logMsg)
}

func (l *TestLogger) Debug(msg string, fields ...interface{}) {
	logMsg := fmt.Sprintf("[DEBUG] %s: %s %v", l.prefix, msg, fields)
	l.Logs = append(l.Logs, logMsg)
	l.t.Log(logMsg)
}

// TestEnvironment represents a test environment for E2E tests
type TestEnvironment struct {
	t              *testing.T
	Server         *hashicorp.Server
	TestDir        string
	DocsDir        string
	PatternsDir    string
	Logger         *TestLogger
	HTTPServer     *httptest.Server
	Context        context.Context
	CancelFunc     context.CancelFunc
	ValidationTest bool
}

// SetupTestEnvironment creates a new test environment
func SetupTestEnvironment(t *testing.T, validationTest bool) *TestEnvironment {
	// Create a test context
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create test directories
	testDir, err := ioutil.TempDir("", "terraform-mcp-test-")
	require.NoError(t, err, "Failed to create test directory")
	
	docsDir := filepath.Join(testDir, "docs")
	patternsDir := filepath.Join(testDir, "patterns")
	
	err = os.MkdirAll(docsDir, 0755)
	require.NoError(t, err, "Failed to create docs directory")
	
	err = os.MkdirAll(patternsDir, 0755)
	require.NoError(t, err, "Failed to create patterns directory")
	
	// Create a test logger
	logger := NewTestLogger(t, "E2E")
	
	// Create server configuration
	config := hashicorp.Config{
		DocSourcePath:  docsDir,
		PatternPath:    patternsDir,
		UpdateInterval: 1 * time.Minute,
	}
	
	// Create server
	server, err := hashicorp.NewServer(config, logger)
	require.NoError(t, err, "Failed to create server")
	
	// Initialize server
	err = server.Initialize(ctx)
	require.NoError(t, err, "Failed to initialize server")
	
	// Create HTTP test server
	httpServer := httptest.NewServer(server)
	
	return &TestEnvironment{
		t:              t,
		Server:         server,
		TestDir:        testDir,
		DocsDir:        docsDir,
		PatternsDir:    patternsDir,
		Logger:         logger,
		HTTPServer:     httpServer,
		Context:        ctx,
		CancelFunc:     cancel,
		ValidationTest: validationTest,
	}
}

// Cleanup cleans up the test environment
func (e *TestEnvironment) Cleanup() {
	e.HTTPServer.Close()
	e.CancelFunc()
	os.RemoveAll(e.TestDir)
}

// CreateTestBestPracticeDocument creates a test best practice document
func (e *TestEnvironment) CreateTestBestPracticeDocument(id, title, content string) {
	doc := tfdocs.Document{
		ID:          id,
		Title:       title,
		Content:     content,
		URL:         fmt.Sprintf("https://example.com/%s", id),
		Category:    "best-practice",
		Tags:        []string{"best-practice", "terraform"},
		Metadata:    map[string]string{"source": "test"},
		LastUpdated: time.Now(),
	}
	
	data, err := json.MarshalIndent(doc, "", "  ")
	require.NoError(e.t, err, "Failed to marshal document")
	
	err = ioutil.WriteFile(filepath.Join(e.DocsDir, id+".json"), data, 0644)
	require.NoError(e.t, err, "Failed to write document file")
}

// CreateTestTerraformModule creates test Terraform files for validation
func (e *TestEnvironment) CreateTestTerraformModule() map[string]string {
	return map[string]string{
		"main.tf": `
resource "aws_vpc" "main" {
  cidr_block = var.vpc_cidr
  
  tags = {
    Name = var.vpc_name
  }
}
`,
		"variables.tf": `
variable "vpc_cidr" {
  description = "The CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "vpc_name" {
  description = "The name of the VPC"
  type        = string
  default     = "main"
}
`,
		"outputs.tf": `
output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.main.id
}
`,
		"README.md": `
# Test Module

This is a test module for validation.

## Usage

\`\`\`hcl
module "vpc" {
  source = "./module"
  
  vpc_cidr = "10.0.0.0/16"
  vpc_name = "main"
}
\`\`\`
`,
	}
}

// ExecuteMCPRequest executes an MCP request
func (e *TestEnvironment) ExecuteMCPRequest(toolName string, args interface{}) (*mcp.Response, error) {
	// Prepare request
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}
	
	req := mcp.Request{
		ID:        "test-request",
		Tool:      toolName,
		Arguments: json.RawMessage(argsJSON),
	}
	
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Send request to HTTP server
	resp, err := http.Post(e.HTTPServer.URL, "application/json", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Parse response
	var mcpResp mcp.Response
	err = json.Unmarshal(respBody, &mcpResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &mcpResp, nil
}
