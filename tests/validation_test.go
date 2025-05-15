// tests/validation_test.go
package tests

import (
	"testing"

	"terraform-mcp-server/pkg/hashicorp/tfdocs"
)

func TestValidationEngine(t *testing.T) {
	// Create a temporary directory for the indexer
	tempDir := t.TempDir()

	// Create a mock logger
	logger := &mockLogger{}

	// Create an indexer
	indexer := tfdocs.NewIndexer(tempDir, logger)

	// Create a validation engine
	engine := tfdocs.NewValidationEngine(indexer, logger)

	// Create a sample Terraform configuration
	config := &tfdocs.TerraformConfiguration{
		Files: map[string]string{
			"main.tf": `
resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}
`,
		},
	}

	// Validate the configuration
	result, err := engine.ValidateConfiguration(config)
	if err != nil {
		t.Fatalf("Failed to validate configuration: %v", err)
	}

	// Check that the result has issues
	if len(result.Issues) == 0 {
		t.Errorf("Expected validation issues, got none")
	}

	// Check missing files issues
	hasStructureIssue := false
	for _, issue := range result.Issues {
		if issue.Category == tfdocs.CategoryStructure {
			hasStructureIssue = true
			break
		}
	}
	if !hasStructureIssue {
		t.Errorf("Expected structure issues, got none")
	}

	// Test with a more complete configuration
	config = &tfdocs.TerraformConfiguration{
		Files: map[string]string{
			"main.tf": `
resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
  tags = var.tags
}
`,
			"variables.tf": `
variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}
`,
			"outputs.tf": `
output "instance_id" {
  description = "The ID of the EC2 instance"
  value       = aws_instance.example.id
}
`,
			"README.md": `
# Example Module

This module creates an EC2 instance.
`,
		},
	}

	// Validate the configuration
	result, err = engine.ValidateConfiguration(config)
	if err != nil {
		t.Fatalf("Failed to validate configuration: %v", err)
	}

	// The configuration should pass most structure checks
	hasStructureError := false
	for _, issue := range result.Issues {
		if issue.Category == tfdocs.CategoryStructure && issue.Severity == tfdocs.SeverityError {
			hasStructureError = true
			break
		}
	}
	if hasStructureError {
		t.Errorf("Expected no structure errors, got some")
	}

	// Test improvement suggestions
	improvements, err := engine.SuggestImprovements(config)
	if err != nil {
		t.Fatalf("Failed to suggest improvements: %v", err)
	}

	// Check that we got improvement suggestions
	if len(improvements) == 0 {
		t.Errorf("Expected improvement suggestions, got none")
	}
}

func TestValidationWithSecurityIssues(t *testing.T) {
	// Create a temporary directory for the indexer
	tempDir := t.TempDir()

	// Create a mock logger
	logger := &mockLogger{}

	// Create an indexer
	indexer := tfdocs.NewIndexer(tempDir, logger)

	// Create a validation engine
	engine := tfdocs.NewValidationEngine(indexer, logger)

	// Create a sample Terraform configuration with security issues
	config := &tfdocs.TerraformConfiguration{
		Files: map[string]string{
			"main.tf": `
resource "aws_security_group" "example" {
  name        = "example"
  description = "Example security group"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
  
  user_data = <<EOF
#!/bin/bash
echo "password=secret123" > /tmp/config
EOF
}
`,
		},
	}

	// Validate the configuration
	result, err := engine.ValidateConfiguration(config)
	if err != nil {
		t.Fatalf("Failed to validate configuration: %v", err)
	}

	// Check for security issues
	hasSecurityIssue := false
	for _, issue := range result.Issues {
		if issue.Category == tfdocs.CategorySecurity {
			hasSecurityIssue = true
			break
		}
	}
	if !hasSecurityIssue {
		t.Errorf("Expected security issues, got none")
	}
}
</content>
