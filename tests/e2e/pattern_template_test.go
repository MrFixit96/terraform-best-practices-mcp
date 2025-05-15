package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-mcp-server/pkg/hashicorp"
	"terraform-mcp-server/pkg/hashicorp/tfdocs"
)

func TestGetPatternTemplate(t *testing.T) {
	// Set up test environment
	env := SetupTestEnvironment(t, false)
	defer env.Cleanup()

	// Test initial patterns (default seeded pattern)
	resp, err := env.ExecuteMCPRequest("GetPatternTemplate", hashicorp.GetPatternTemplateArgs{})
	require.NoError(t, err, "Failed to execute GetPatternTemplate request")
	require.Equal(t, "success", resp.Status, "Request should succeed")

	// Parse result
	var result hashicorp.GetPatternTemplateResult
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to unmarshal result")

	// Verify default pattern is returned
	assert.Equal(t, 1, len(result.Patterns), "Should return default pattern")
	assert.Equal(t, "aws-vpc", result.Patterns[0].ID, "Default pattern should be aws-vpc")
	assert.Equal(t, "networking", result.Patterns[0].Category, "Default pattern should be in networking category")

	// Add a new pattern to test category filtering
	pattern := tfdocs.Pattern{
		ID:          "gcp-gke",
		Name:        "GCP GKE Cluster Module",
		Description: "A Terraform module for creating a GKE cluster on Google Cloud",
		Category:    "kubernetes",
		Tags:        []string{"gcp", "kubernetes", "gke"},
		Files: map[string]string{
			"main.tf": `
resource "google_container_cluster" "primary" {
  name     = var.cluster_name
  location = var.location
  
  node_config {
    machine_type = var.machine_type
  }
}
`,
			"variables.tf": `
variable "cluster_name" {
  description = "The name of the GKE cluster"
  type        = string
}

variable "location" {
  description = "The location of the GKE cluster"
  type        = string
  default     = "us-central1"
}

variable "machine_type" {
  description = "The machine type for the GKE nodes"
  type        = string
  default     = "e2-medium"
}
`,
			"README.md": `
# GCP GKE Cluster Module

This module creates a GKE cluster on Google Cloud.
`,
		},
		Metadata: map[string]string{
			"provider":   "gcp",
			"complexity": "medium",
		},
	}

	// Save pattern (using reflection to access private method)
	patternJSON, err := json.Marshal(pattern)
	require.NoError(t, err, "Failed to marshal pattern")
	
	patternPath := env.PatternsDir + "/" + pattern.ID + ".json"
	err = ioutil.WriteFile(patternPath, patternJSON, 0644)
	require.NoError(t, err, "Failed to write pattern file")

	// Refresh the pattern repository
	err = env.Server.Initialize(env.Context)
	require.NoError(t, err, "Failed to reinitialize server")

	// Test category filter
	resp, err = env.ExecuteMCPRequest("GetPatternTemplate", hashicorp.GetPatternTemplateArgs{
		Category: "kubernetes",
	})
	require.NoError(t, err, "Failed to execute GetPatternTemplate request with category")
	require.Equal(t, "success", resp.Status, "Request should succeed")

	// Parse result
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to unmarshal result")

	// Verify only kubernetes patterns are returned
	assert.Equal(t, 1, len(result.Patterns), "Should return only kubernetes patterns")
	assert.Equal(t, "gcp-gke", result.Patterns[0].ID, "Should return gcp-gke pattern")

	// Test tag filter
	resp, err = env.ExecuteMCPRequest("GetPatternTemplate", hashicorp.GetPatternTemplateArgs{
		Tags: []string{"gcp"},
	})
	require.NoError(t, err, "Failed to execute GetPatternTemplate request with tags")
	require.Equal(t, "success", resp.Status, "Request should succeed")

	// Parse result
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to unmarshal result")

	// Verify only patterns with gcp tag are returned
	assert.Equal(t, 1, len(result.Patterns), "Should return only gcp patterns")
	assert.Equal(t, "gcp-gke", result.Patterns[0].ID, "Should return gcp-gke pattern")

	// Test multiple tags
	resp, err = env.ExecuteMCPRequest("GetPatternTemplate", hashicorp.GetPatternTemplateArgs{
		Tags: []string{"gcp", "networking"},
	})
	require.NoError(t, err, "Failed to execute GetPatternTemplate request with multiple tags")
	require.Equal(t, "success", resp.Status, "Request should succeed")

	// Parse result
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to unmarshal result")

	// Verify both patterns are returned (one for each tag)
	assert.Equal(t, 2, len(result.Patterns), "Should return both matching patterns")
	
	// Check that the patterns contain the expected files and content
	for _, p := range result.Patterns {
		if p.ID == "gcp-gke" {
			assert.Contains(t, p.Files["main.tf"], "google_container_cluster", "GKE pattern should contain cluster resource")
			assert.Contains(t, p.Files["variables.tf"], "cluster_name", "GKE pattern should have cluster_name variable")
		} else if p.ID == "aws-vpc" {
			assert.Contains(t, p.Files["main.tf"], "aws_vpc", "VPC pattern should contain VPC resource")
		}
	}
}
