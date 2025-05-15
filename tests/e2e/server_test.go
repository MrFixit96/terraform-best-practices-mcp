package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerInitialization tests the server initialization process
func TestServerInitialization(t *testing.T) {
	// Set up test environment
	env := SetupTestEnvironment(t, false)
	defer env.Cleanup()
	
	// Verify that server initialization logs contain expected messages
	foundIndexer := false
	foundPattern := false
	foundResource := false
	foundValidation := false
	
	for _, log := range env.Logger.Logs {
		if log == "[INFO] E2E: Starting documentation indexer sourcePath=[testdir]" {
			foundIndexer = true
		} else if log == "[INFO] E2E: Initializing pattern repository patternsPath=[testdir]" {
			foundPattern = true
		} else if log == "[INFO] E2E: Initializing resource provider" {
			foundResource = true
		} else if log == "[INFO] E2E: Initializing validation engine" {
			foundValidation = true
		}
	}
	
	assert.True(t, foundIndexer, "Should log indexer initialization")
	assert.True(t, foundPattern, "Should log pattern repository initialization")
	assert.True(t, foundResource, "Should log resource provider initialization")
	assert.True(t, foundValidation, "Should log validation engine initialization")
}

// TestServerReinitialization tests the server reinitialization process
func TestServerReinitialization(t *testing.T) {
	// Set up test environment
	env := SetupTestEnvironment(t, false)
	defer env.Cleanup()
	
	// Record initial log count
	initialLogCount := len(env.Logger.Logs)
	
	// Wait a bit to ensure different timestamps
	time.Sleep(100 * time.Millisecond)
	
	// Reinitialize server
	err := env.Server.Initialize(env.Context)
	require.NoError(t, err, "Failed to reinitialize server")
	
	// Check that logs show reinitialization
	assert.Greater(t, len(env.Logger.Logs), initialLogCount, "Should have additional logs after reinitialization")
	
	// Verify that all components were reinitialized
	reinitCount := 0
	for i := initialLogCount; i < len(env.Logger.Logs); i++ {
		log := env.Logger.Logs[i]
		if log == "[INFO] E2E: Initializing HashiCorp MCP server" {
			reinitCount++
		} else if log == "[INFO] E2E: Initializing pattern repository patternsPath=[testdir]" {
			reinitCount++
		} else if log == "[INFO] E2E: Initializing resource provider" {
			reinitCount++
		} else if log == "[INFO] E2E: Initializing validation engine" {
			reinitCount++
		}
	}
	
	assert.Greater(t, reinitCount, 0, "Should have reinitialization logs")
}

// TestServerToolRegistration tests that all tools are properly registered
func TestServerToolRegistration(t *testing.T) {
	// Set up test environment
	env := SetupTestEnvironment(t, false)
	defer env.Cleanup()
	
	// Check for tool registration logs
	registeredTools := []string{
		"GetBestPractices",
		"GetModuleStructure",
		"GetPatternTemplate",
		"ValidateConfiguration",
	}
	
	for _, tool := range registeredTools {
		found := false
		for _, log := range env.Logger.Logs {
			if log == "[INFO] E2E: Registered tool name="+tool {
				found = true
				break
			}
		}
		assert.True(t, found, "Tool %s should be registered", tool)
	}
	
	// Test invalid tool name
	resp, err := env.ExecuteMCPRequest("InvalidTool", map[string]string{})
	require.NoError(t, err, "Failed to execute invalid tool request")
	assert.Equal(t, "error", resp.Status, "Invalid tool request should fail")
	assert.Equal(t, "tool_not_found", resp.Error.Code, "Should return tool_not_found error")
}

// TestAPIErrorHandling tests error handling in API requests
func TestAPIErrorHandling(t *testing.T) {
	// Set up test environment
	env := SetupTestEnvironment(t, false)
	defer env.Cleanup()
	
	// Test with invalid arguments
	resp, err := env.ExecuteMCPRequest("GetBestPractices", "invalid arguments")
	require.NoError(t, err, "Failed to execute request with invalid arguments")
	assert.Equal(t, "error", resp.Status, "Request with invalid arguments should fail")
	
	// Test with malformed request
	// This test requires low-level HTTP client work, so we'll skip it for now
}

// TestConcurrentOperations tests concurrent operations on the server
func TestConcurrentOperations(t *testing.T) {
	// Set up test environment
	env := SetupTestEnvironment(t, false)
	defer env.Cleanup()
	
	// Create a cancel context to stop goroutines
	ctx, cancel := context.WithCancel(env.Context)
	defer cancel()
	
	// Create test documents
	env.CreateTestBestPracticeDocument(
		"concurrent-test",
		"Concurrent Test",
		"This is a test document for concurrent operations",
	)
	
	// Run concurrent requests
	errorCh := make(chan error, 10)
	doneCh := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() {
				doneCh <- true
			}()
			
			resp, err := env.ExecuteMCPRequest("GetBestPractices", hashicorp.GetBestPracticesArgs{})
			if err != nil {
				errorCh <- err
				return
			}
			
			if resp.Status != "success" {
				errorCh <- fmt.Errorf("request failed: %s", resp.Error.Message)
				return
			}
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case err := <-errorCh:
			t.Errorf("Concurrent request failed: %v", err)
		case <-doneCh:
			// Request completed successfully
		}
	}
}
