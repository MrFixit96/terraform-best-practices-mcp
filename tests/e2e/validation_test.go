package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-mcp-server/pkg/hashicorp"
)

func TestValidateConfiguration(t *testing.T) {
	// Set up test environment
	env := SetupTestEnvironment(t, true)
	defer env.Cleanup()

	// Test a well-formed module
	goodModule := env.CreateTestTerraformModule()
	
	resp, err := env.ExecuteMCPRequest("ValidateConfiguration", hashicorp.ValidateConfigurationArgs{
		Files: goodModule,
	})
	require.NoError(t, err, "Failed to execute ValidateConfiguration request")
	require.Equal(t, "success", resp.Status, "Request should succeed")

	// Parse result
	var result hashicorp.ValidateConfigurationResult
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to unmarshal result")

	// Check summary - all should pass
	assert.Equal(t, len(result.Results), result.Summary.TotalCount, "Total count should match results length")
	assert.True(t, result.Summary.PassedCount > 0, "Some validations should pass")
	assert.Equal(t, 0, result.Summary.FailedCount, "No validations should fail for good module")
	
	// Now create a bad module with various issues
	badModule := map[string]string{
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
  # Missing description
  type        = string
  default     = "10.0.0.0/16"
}

variable "vpc_name" {
  # Missing description
  type        = string
  default     = "main"
}

variable "password" {
  # Sensitive variable not marked as sensitive
  type        = string
  default     = ""
}
`,
		"outputs.tf": `
output "vpc_id" {
  # Missing description
  value       = aws_vpc.main.id
}

output "static_value" {
  # Output without resource dependency
  value = "static"
}
`,
		// Missing README.md
	}
	
	resp, err = env.ExecuteMCPRequest("ValidateConfiguration", hashicorp.ValidateConfigurationArgs{
		Files: badModule,
	})
	require.NoError(t, err, "Failed to execute ValidateConfiguration request for bad module")
	require.Equal(t, "success", resp.Status, "Request should succeed")

	// Parse result
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to unmarshal result")

	// Verify there are failures
	assert.True(t, result.Summary.FailedCount > 0, "Bad module should have failures")
	
	// Check specific validation failures
	var moduleStructureFailure, variableDescriptionFailure, outputDescriptionFailure, 
		readmeFailure, outputDependencyFailure, sensitiveVariableFailure bool
	
	for _, res := range result.Results {
		if !res.Passed {
			switch res.Rule.ID {
			case "module-structure-files":
				moduleStructureFailure = true
				assert.Contains(t, res.Message, "README.md", "Should report missing README.md")
			case "variable-description":
				variableDescriptionFailure = true
				assert.Contains(t, res.Message, "missing descriptions", "Should report missing variable descriptions")
			case "output-description":
				outputDescriptionFailure = true
				assert.Contains(t, res.Message, "missing descriptions", "Should report missing output descriptions")
			case "readme-exists":
				readmeFailure = true
				// Rule might not fire since module-structure-files already covers it
			case "output-value-dependency":
				outputDependencyFailure = true
				assert.Contains(t, res.Message, "static_value", "Should report static_value as problematic")
			case "sensitive-variables":
				sensitiveVariableFailure = true
				assert.Contains(t, res.Message, "password", "Should report password variable as sensitive")
			}
		}
	}
	
	// Assert that at least some of the important validation failures are present
	assert.True(t, moduleStructureFailure, "Module structure failure should be detected")
	assert.True(t, variableDescriptionFailure, "Variable description failure should be detected")
	assert.True(t, outputDescriptionFailure, "Output description failure should be detected")
	assert.True(t, outputDependencyFailure, "Output dependency failure should be detected")
	assert.True(t, sensitiveVariableFailure, "Sensitive variable failure should be detected")
	
	// Test normalized file names
	normalizedModule := map[string]string{
		"main": badModule["main.tf"],        // Without .tf extension
		"vars": badModule["variables.tf"],   // Different name
		"output.tf": badModule["outputs.tf"], // Singular
	}
	
	resp, err = env.ExecuteMCPRequest("ValidateConfiguration", hashicorp.ValidateConfigurationArgs{
		Files: normalizedModule,
	})
	require.NoError(t, err, "Failed to execute ValidateConfiguration request with normalized file names")
	require.Equal(t, "success", resp.Status, "Request should succeed")
	
	// Parse result
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to unmarshal result")
	
	// Verify normalization worked - should have similar failures as the bad module
	assert.True(t, result.Summary.FailedCount > 0, "Normalized module should have failures")
	
	// Verify error counts by severity
	assert.True(t, result.Summary.ErrorCount > 0, "Should have error-level failures")
	assert.True(t, result.Summary.WarningCount > 0, "Should have warning-level failures")
}
