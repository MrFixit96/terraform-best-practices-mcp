package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-mcp-server/pkg/hashicorp"
)

func TestGetBestPractices(t *testing.T) {
	// Set up test environment
	env := SetupTestEnvironment(t, false)
	defer env.Cleanup()

	// Create test documents
	env.CreateTestBestPracticeDocument(
		"module-structure",
		"Module Structure Best Practices",
		"Terraform modules should follow a standard structure with main.tf, variables.tf, outputs.tf, and README.md",
	)
	env.CreateTestBestPracticeDocument(
		"naming-conventions",
		"Naming Conventions Best Practices",
		"Use snake_case for resource names and descriptive names for all resources",
	)
	env.CreateTestBestPracticeDocument(
		"security-practices",
		"Security Best Practices",
		"Always use variables for sensitive data and mark them as sensitive",
	)

	// Test empty topic query
	resp, err := env.ExecuteMCPRequest("GetBestPractices", hashicorp.GetBestPracticesArgs{})
	require.NoError(t, err, "Failed to execute GetBestPractices request")
	require.Equal(t, "success", resp.Status, "Request should succeed")

	// Parse result
	var result hashicorp.GetBestPracticesResult
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to unmarshal result")

	// Verify all 3 documents are returned
	assert.Equal(t, 3, len(result.Practices), "Should return all 3 practices")

	// Test specific topic query
	resp, err = env.ExecuteMCPRequest("GetBestPractices", hashicorp.GetBestPracticesArgs{
		Topic: "security",
	})
	require.NoError(t, err, "Failed to execute GetBestPractices request with topic")
	require.Equal(t, "success", resp.Status, "Request should succeed")

	// Parse result
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to unmarshal result")

	// Verify only security-related documents are returned
	assert.Equal(t, 1, len(result.Practices), "Should return only security practices")
	
	// Check resources contain the correct data
	for _, practice := range result.Practices {
		var resource map[string]interface{}
		err = json.Unmarshal(practice, &resource)
		require.NoError(t, err, "Failed to unmarshal practice")
		
		data := resource["data"].(map[string]interface{})
		content := data["content"].(string)
		assert.Contains(t, content, "sensitive", "Content should contain the word 'sensitive'")
	}
}
