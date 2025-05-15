// pkg/hashicorp/tfdocs/validation.go
package tfdocs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidationSeverity represents the severity of a validation issue
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
	SeverityInfo    ValidationSeverity = "info"
)

// ValidationCategory represents the category of a validation issue
type ValidationCategory string

const (
	CategoryStructure    ValidationCategory = "structure"
	CategoryNaming       ValidationCategory = "naming"
	CategorySecurity     ValidationCategory = "security"
	CategoryPerformance  ValidationCategory = "performance"
	CategoryMaintenance  ValidationCategory = "maintenance"
	CategoryDocumentation ValidationCategory = "documentation"
)

// ValidationIssue represents an issue found during validation
type ValidationIssue struct {
	Message      string             `json:"message"`
	Severity     ValidationSeverity  `json:"severity"`
	Category     ValidationCategory  `json:"category"`
	File         string              `json:"file,omitempty"`
	Line         int                 `json:"line,omitempty"`
	BestPractice string              `json:"best_practice,omitempty"`
	Suggestion   string              `json:"suggestion,omitempty"`
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Issues     []ValidationIssue `json:"issues"`
	FileCount  int               `json:"file_count"`
	ErrorCount int               `json:"error_count"`
	WarnCount  int               `json:"warn_count"`
	InfoCount  int               `json:"info_count"`
}

// TerraformConfiguration represents a Terraform configuration
type TerraformConfiguration struct {
	Files map[string]string
}

// ValidationEngine validates Terraform configurations against best practices
type ValidationEngine struct {
	docIndexer   *Indexer
	logger       Logger
	validators   []Validator
}

// Validator is the interface for validators
type Validator interface {
	Validate(config *TerraformConfiguration) []ValidationIssue
	Name() string
}

// NewValidationEngine creates a new validation engine
func NewValidationEngine(docIndexer *Indexer, logger Logger) *ValidationEngine {
	engine := &ValidationEngine{
		docIndexer: docIndexer,
		logger:     logger,
	}

	// Register validators
	engine.validators = []Validator{
		&StructureValidator{},
		&NamingValidator{},
		&SecurityValidator{},
		&DocumentationValidator{},
		&ModuleValidator{},
		&ResourceValidator{},
	}

	return engine
}

// ValidateConfiguration validates a Terraform configuration
func (e *ValidationEngine) ValidateConfiguration(config *TerraformConfiguration) (*ValidationResult, error) {
	e.logger.Info("Validating Terraform configuration")

	result := &ValidationResult{
		Issues: []ValidationIssue{},
	}

	// Count files
	result.FileCount = len(config.Files)

	// Run each validator
	for _, validator := range e.validators {
		e.logger.Debug("Running validator", "name", validator.Name())
		issues := validator.Validate(config)
		result.Issues = append(result.Issues, issues...)
	}

	// Count issues by severity
	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityError:
			result.ErrorCount++
		case SeverityWarning:
			result.WarnCount++
		case SeverityInfo:
			result.InfoCount++
		}
	}

	return result, nil
}

// SuggestImprovements suggests improvements for a Terraform configuration
func (e *ValidationEngine) SuggestImprovements(config *TerraformConfiguration) (map[string]string, error) {
	improvements := make(map[string]string)

	// Validate configuration
	result, err := e.ValidateConfiguration(config)
	if err != nil {
		return nil, err
	}

	// Generate improvements for common issues
	if !hasMainTF(config) {
		improvements["main.tf"] = "// main.tf\n" + generateMainTF(config)
	}

	if !hasVariablesTF(config) {
		improvements["variables.tf"] = "// variables.tf\n" + generateVariablesTF(config)
	}

	if !hasOutputsTF(config) {
		improvements["outputs.tf"] = "// outputs.tf\n" + generateOutputsTF(config)
	}

	if !hasReadmeMD(config) {
		improvements["README.md"] = generateReadmeMD(config)
	}

	// Add specific improvements based on validation issues
	for _, issue := range result.Issues {
		if issue.Severity == SeverityError || issue.Severity == SeverityWarning {
			if issue.File != "" && issue.Suggestion != "" {
				if _, ok := improvements[issue.File]; !ok {
					if content, ok := config.Files[issue.File]; ok {
						improvements[issue.File] = content
					} else {
						improvements[issue.File] = ""
					}
				}
				// Add comment with improvement suggestion
				improvements[issue.File] = fmt.Sprintf("// TODO: %s\n%s", issue.Message, improvements[issue.File])
			}
		}
	}

	return improvements, nil
}

// StructureValidator validates the structure of a Terraform configuration
type StructureValidator struct{}

// Name returns the name of the validator
func (v *StructureValidator) Name() string {
	return "StructureValidator"
}

// Validate validates the structure of a Terraform configuration
func (v *StructureValidator) Validate(config *TerraformConfiguration) []ValidationIssue {
	var issues []ValidationIssue

	// Check for essential files
	if !hasMainTF(config) {
		issues = append(issues, ValidationIssue{
			Message:      "Missing main.tf file",
			Severity:     SeverityError,
			Category:     CategoryStructure,
			BestPractice: "Include a main.tf file with core resource definitions",
			Suggestion:   "Create a main.tf file with core resource definitions",
		})
	}

	if !hasVariablesTF(config) {
		issues = append(issues, ValidationIssue{
			Message:      "Missing variables.tf file",
			Severity:     SeverityWarning,
			Category:     CategoryStructure,
			BestPractice: "Include a variables.tf file for input variable definitions",
			Suggestion:   "Create a variables.tf file with input variable definitions",
		})
	}

	if !hasOutputsTF(config) {
		issues = append(issues, ValidationIssue{
			Message:      "Missing outputs.tf file",
			Severity:     SeverityWarning,
			Category:     CategoryStructure,
			BestPractice: "Include an outputs.tf file for output definitions",
			Suggestion:   "Create an outputs.tf file with output definitions",
		})
	}

	// Check for monolithic files
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") {
			lineCount := len(strings.Split(content, "\n"))
			if lineCount > 500 {
				issues = append(issues, ValidationIssue{
					Message:      fmt.Sprintf("File %s is too large (%d lines). Consider splitting it into multiple files.", name, lineCount),
					Severity:     SeverityWarning,
					Category:     CategoryMaintenance,
					File:         name,
					BestPractice: "Keep Terraform files under 500 lines for better maintainability",
					Suggestion:   "Split the file into multiple logical files based on resource types or functionality",
				})
			}
		}
	}

	// Check for proper module structure
	moduleFiles := []string{"main.tf", "variables.tf", "outputs.tf", "README.md"}
	missingFiles := []string{}
	for _, file := range moduleFiles {
		if !hasFile(config, file) {
			missingFiles = append(missingFiles, file)
		}
	}
	if len(missingFiles) > 0 {
		issues = append(issues, ValidationIssue{
			Message:      fmt.Sprintf("Module is missing standard files: %s", strings.Join(missingFiles, ", ")),
			Severity:     SeverityInfo,
			Category:     CategoryStructure,
			BestPractice: "Follow standard module structure with main.tf, variables.tf, outputs.tf, and README.md",
			Suggestion:   "Add the missing files to follow the standard module structure",
		})
	}

	return issues
}

// NamingValidator validates naming conventions in a Terraform configuration
type NamingValidator struct{}

// Name returns the name of the validator
func (v *NamingValidator) Name() string {
	return "NamingValidator"
}

// Validate validates naming conventions in a Terraform configuration
func (v *NamingValidator) Validate(config *TerraformConfiguration) []ValidationIssue {
	var issues []ValidationIssue

	// Check variable naming conventions
	varPattern := regexp.MustCompile(`variable\s+"([^"]+)"\s+{`)
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") {
			matches := varPattern.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				varName := match[1]
				if strings.Contains(varName, "-") {
					issues = append(issues, ValidationIssue{
						Message:      fmt.Sprintf("Variable name '%s' uses hyphens instead of underscores", varName),
						Severity:     SeverityWarning,
						Category:     CategoryNaming,
						File:         name,
						BestPractice: "Use underscores, not hyphens, in variable names",
						Suggestion:   fmt.Sprintf("Rename variable '%s' to use underscores instead of hyphens", varName),
					})
				}
				if strings.ToLower(varName) != varName {
					issues = append(issues, ValidationIssue{
						Message:      fmt.Sprintf("Variable name '%s' uses uppercase letters", varName),
						Severity:     SeverityInfo,
						Category:     CategoryNaming,
						File:         name,
						BestPractice: "Use lowercase letters in variable names",
						Suggestion:   fmt.Sprintf("Rename variable '%s' to use all lowercase letters", varName),
					})
				}
			}
		}
	}

	// Check resource naming conventions
	resPattern := regexp.MustCompile(`resource\s+"([^"]+)"\s+"([^"]+)"\s+{`)
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") {
			matches := resPattern.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				resName := match[2]
				if strings.Contains(resName, "_") && !strings.Contains(resName, "-") {
					// This is following HashiCorp convention for resource names
					continue
				}
				issues = append(issues, ValidationIssue{
					Message:      fmt.Sprintf("Resource name '%s' doesn't follow naming convention", resName),
					Severity:     SeverityInfo,
					Category:     CategoryNaming,
					File:         name,
					BestPractice: "Use underscores in resource names for readability",
					Suggestion:   fmt.Sprintf("Rename resource '%s' to use underscores", resName),
				})
			}
		}
	}

	return issues
}

// SecurityValidator validates security practices in a Terraform configuration
type SecurityValidator struct{}

// Name returns the name of the validator
func (v *SecurityValidator) Name() string {
	return "SecurityValidator"
}

// Validate validates security practices in a Terraform configuration
func (v *SecurityValidator) Validate(config *TerraformConfiguration) []ValidationIssue {
	var issues []ValidationIssue

	// Check for hardcoded credentials
	secretPattern := regexp.MustCompile(`(?i)(password|secret|key|token|credential)s?\s*=\s*"[^"]+"`)
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") {
			matches := secretPattern.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				issues = append(issues, ValidationIssue{
					Message:      fmt.Sprintf("Possible hardcoded secret found: %s", match[0]),
					Severity:     SeverityError,
					Category:     CategorySecurity,
					File:         name,
					BestPractice: "Never hardcode sensitive values in Terraform configuration",
					Suggestion:   "Use variables with sensitive = true or integrate with a secrets management solution",
				})
			}
		}
	}

	// Check for sensitive variables
	sensitivePattern := regexp.MustCompile(`variable\s+"([^"]+)"\s+{(?:(?:.|\n)(?!^\}$))*sensitive\s*=\s*true`)
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") && strings.Contains(strings.ToLower(name), "variable") {
			if !sensitivePattern.MatchString(content) && secretPattern.MatchString(content) {
				issues = append(issues, ValidationIssue{
					Message:      "Sensitive variables should be marked with sensitive = true",
					Severity:     SeverityWarning,
					Category:     CategorySecurity,
					File:         name,
					BestPractice: "Mark sensitive variables with sensitive = true",
					Suggestion:   "Add sensitive = true to variable definitions containing sensitive information",
				})
			}
		}
	}

	// Check for overly permissive security groups
	sgPattern := regexp.MustCompile(`(?i)ingress\s+{[^}]*cidr_blocks\s*=\s*\[\s*"0\.0\.0\.0/0"\s*\]`)
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") {
			matches := sgPattern.FindAllString(content, -1)
			for range matches {
				issues = append(issues, ValidationIssue{
					Message:      "Security group allows access from 0.0.0.0/0 (any IP)",
					Severity:     SeverityWarning,
					Category:     CategorySecurity,
					File:         name,
					BestPractice: "Restrict security group access to specific IP ranges",
					Suggestion:   "Replace 0.0.0.0/0 with specific IP ranges or use a variable for allowed IPs",
				})
			}
		}
	}

	return issues
}

// DocumentationValidator validates documentation in a Terraform configuration
type DocumentationValidator struct{}

// Name returns the name of the validator
func (v *DocumentationValidator) Name() string {
	return "DocumentationValidator"
}

// Validate validates documentation in a Terraform configuration
func (v *DocumentationValidator) Validate(config *TerraformConfiguration) []ValidationIssue {
	var issues []ValidationIssue

	// Check for README.md
	if !hasReadmeMD(config) {
		issues = append(issues, ValidationIssue{
			Message:      "Missing README.md file",
			Severity:     SeverityWarning,
			Category:     CategoryDocumentation,
			BestPractice: "Include a README.md file with module documentation",
			Suggestion:   "Create a README.md file with module usage examples and documentation",
		})
	}

	// Check variable descriptions
	varPattern := regexp.MustCompile(`variable\s+"([^"]+)"\s+{(?:(?:.|\n)(?!^\}$))*}`)
	descPattern := regexp.MustCompile(`description\s*=\s*"[^"]+"`)
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") && strings.Contains(strings.ToLower(name), "variable") {
			varMatches := varPattern.FindAllStringSubmatch(content, -1)
			for _, varMatch := range varMatches {
				varDef := varMatch[0]
				varName := varMatch[1]
				if !descPattern.MatchString(varDef) {
					issues = append(issues, ValidationIssue{
						Message:      fmt.Sprintf("Variable '%s' is missing a description", varName),
						Severity:     SeverityWarning,
						Category:     CategoryDocumentation,
						File:         name,
						BestPractice: "Include descriptions for all variables",
						Suggestion:   fmt.Sprintf("Add a description attribute to variable '%s'", varName),
					})
				}
			}
		}
	}

	// Check output descriptions
	outPattern := regexp.MustCompile(`output\s+"([^"]+)"\s+{(?:(?:.|\n)(?!^\}$))*}`)
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") && strings.Contains(strings.ToLower(name), "output") {
			outMatches := outPattern.FindAllStringSubmatch(content, -1)
			for _, outMatch := range outMatches {
				outDef := outMatch[0]
				outName := outMatch[1]
				if !descPattern.MatchString(outDef) {
					issues = append(issues, ValidationIssue{
						Message:      fmt.Sprintf("Output '%s' is missing a description", outName),
						Severity:     SeverityInfo,
						Category:     CategoryDocumentation,
						File:         name,
						BestPractice: "Include descriptions for all outputs",
						Suggestion:   fmt.Sprintf("Add a description attribute to output '%s'", outName),
					})
				}
			}
		}
	}

	return issues
}

// ModuleValidator validates module usage in a Terraform configuration
type ModuleValidator struct{}

// Name returns the name of the validator
func (v *ModuleValidator) Name() string {
	return "ModuleValidator"
}

// Validate validates module usage in a Terraform configuration
func (v *ModuleValidator) Validate(config *TerraformConfiguration) []ValidationIssue {
	var issues []ValidationIssue

	// Check module version pinning
	modulePattern := regexp.MustCompile(`module\s+"([^"]+)"\s+{(?:(?:.|\n)(?!^\}$))*}`)
	sourcePattern := regexp.MustCompile(`source\s*=\s*"([^"]+)"`)
	versionPattern := regexp.MustCompile(`version\s*=\s*"([^"]+)"`)
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") {
			modMatches := modulePattern.FindAllStringSubmatch(content, -1)
			for _, modMatch := range modMatches {
				modDef := modMatch[0]
				modName := modMatch[1]
				sourceMatch := sourcePattern.FindStringSubmatch(modDef)
				if sourceMatch != nil {
					source := sourceMatch[1]
					if strings.Contains(source, "github.com") || strings.Contains(source, "terraform-aws-modules") || 
					   strings.Contains(source, "registry.terraform.io") {
						if !versionPattern.MatchString(modDef) {
							issues = append(issues, ValidationIssue{
								Message:      fmt.Sprintf("Module '%s' does not specify a version", modName),
								Severity:     SeverityWarning,
								Category:     CategoryMaintenance,
								File:         name,
								BestPractice: "Always pin module versions for consistency and stability",
								Suggestion:   fmt.Sprintf("Add version constraint to module '%s'", modName),
							})
						}
					}
				}
			}
		}
	}

	// Check for local modules
	localModulesDir := hasDir(config, "modules")
	moduleUsage := false
	for _, content := range config.Files {
		if strings.Contains(content, "module ") {
			moduleUsage = true
			break
		}
	}
	if localModulesDir && !moduleUsage {
		issues = append(issues, ValidationIssue{
			Message:      "Local modules directory exists but modules are not used",
			Severity:     SeverityInfo,
			Category:     CategoryMaintenance,
			BestPractice: "Use a modular approach for complex configurations",
			Suggestion:   "Consider using the modules in your configuration for better organization",
		})
	}

	return issues
}

// ResourceValidator validates resource usage in a Terraform configuration
type ResourceValidator struct{}

// Name returns the name of the validator
func (v *ResourceValidator) Name() string {
	return "ResourceValidator"
}

// Validate validates resource usage in a Terraform configuration
func (v *ResourceValidator) Validate(config *TerraformConfiguration) []ValidationIssue {
	var issues []ValidationIssue

	// Check for missing tags on resources
	tagPattern := regexp.MustCompile(`resource\s+"(aws_[^"]+)"\s+"([^"]+)"\s+{(?:(?:.|\n)(?!^\}$))*}`)
	tagsAttrPattern := regexp.MustCompile(`tags\s*=\s*`)
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") {
			resMatches := tagPattern.FindAllStringSubmatch(content, -1)
			for _, resMatch := range resMatches {
				resType := resMatch[1]
				resName := resMatch[2]
				resDef := resMatch[0]

				// Skip resources that don't support tags
				if strings.Contains(resType, "aws_iam_role_policy") || 
				   strings.Contains(resType, "aws_iam_policy") ||
				   strings.Contains(resType, "aws_route") {
					continue
				}

				// Check for resources that typically should have tags
				if (strings.HasPrefix(resType, "aws_") || 
					strings.HasPrefix(resType, "azurerm_") || 
					strings.HasPrefix(resType, "google_")) && 
					!tagsAttrPattern.MatchString(resDef) {
					issues = append(issues, ValidationIssue{
						Message:      fmt.Sprintf("Resource '%s' of type '%s' is missing tags", resName, resType),
						Severity:     SeverityInfo,
						Category:     CategoryMaintenance,
						File:         name,
						BestPractice: "Apply consistent tagging to all resources for better management",
						Suggestion:   fmt.Sprintf("Add tags to resource '%s'", resName),
					})
				}
			}
		}
	}

	// Check for resource count vs for_each
	countPattern := regexp.MustCompile(`resource\s+"([^"]+)"\s+"([^"]+)"\s+{(?:(?:.|\n)(?!^\}$))*\s+count\s*=\s*length\(([^)]+)\)`)
	for name, content := range config.Files {
		if strings.HasSuffix(name, ".tf") {
			countMatches := countPattern.FindAllStringSubmatch(content, -1)
			for _, countMatch := range countMatches {
				resType := countMatch[1]
				resName := countMatch[2]
				countVar := countMatch[3]
				issues = append(issues, ValidationIssue{
					Message:      fmt.Sprintf("Resource '%s' uses count with length(%s), consider using for_each", resName, countVar),
					Severity:     SeverityInfo,
					Category:     CategoryMaintenance,
					File:         name,
					BestPractice: "Use for_each instead of count when iterating over complex values",
					Suggestion:   fmt.Sprintf("Change 'count = length(%s)' to 'for_each = toset(%s)'", countVar, countVar),
				})
			}
		}
	}

	return issues
}

// Helper functions
func hasFile(config *TerraformConfiguration, name string) bool {
	_, ok := config.Files[name]
	return ok
}

func hasMainTF(config *TerraformConfiguration) bool {
	return hasFile(config, "main.tf")
}

func hasVariablesTF(config *TerraformConfiguration) bool {
	for name := range config.Files {
		if name == "variables.tf" || strings.HasSuffix(name, "_variables.tf") {
			return true
		}
	}
	return false
}

func hasOutputsTF(config *TerraformConfiguration) bool {
	for name := range config.Files {
		if name == "outputs.tf" || strings.HasSuffix(name, "_outputs.tf") {
			return true
		}
	}
	return false
}

func hasReadmeMD(config *TerraformConfiguration) bool {
	_, ok := config.Files["README.md"]
	if ok {
		return true
	}
	_, ok = config.Files["readme.md"]
	return ok
}

func hasDir(config *TerraformConfiguration, dir string) bool {
	for name := range config.Files {
		if strings.HasPrefix(name, dir+"/") {
			return true
		}
	}
	return false
}

// Generates a simple main.tf file
func generateMainTF(config *TerraformConfiguration) string {
	var sb strings.Builder

	sb.WriteString(`# Main configuration for Terraform
# Contains the primary resources defined in this module

provider "aws" {
  region = var.region
}

# Example resource:
# resource "aws_s3_bucket" "example" {
#   bucket = var.bucket_name
#   tags   = var.tags
# }

# Example module usage:
# module "vpc" {
#   source  = "terraform-aws-modules/vpc/aws"
#   version = "3.14.0"
#
#   name = var.vpc_name
#   cidr = var.vpc_cidr
#
#   azs             = var.availability_zones
#   private_subnets = var.private_subnets
#   public_subnets  = var.public_subnets
#
#   tags = var.tags
# }
`)

	return sb.String()
}

// Generates a simple variables.tf file
func generateVariablesTF(config *TerraformConfiguration) string {
	var sb strings.Builder

	sb.WriteString(`# Input variables for the module

variable "region" {
  description = "AWS region where resources will be created"
  type        = string
  default     = "us-west-2"
}

# Example variable:
# variable "bucket_name" {
#   description = "Name of the S3 bucket to create"
#   type        = string
# }

# Example variable with validation:
# variable "environment" {
#   description = "Environment where resources will be deployed"
#   type        = string
#   validation {
#     condition     = contains(["dev", "staging", "prod"], var.environment)
#     error_message = "Environment must be one of: dev, staging, prod."
#   }
# }

variable "tags" {
  description = "A map of tags to apply to all resources"
  type        = map(string)
  default     = {}
}
`)

	return sb.String()
}

// Generates a simple outputs.tf file
func generateOutputsTF(config *TerraformConfiguration) string {
	var sb strings.Builder

	sb.WriteString(`# Output values from the module

# Example output:
# output "bucket_id" {
#   description = "The ID of the S3 bucket"
#   value       = aws_s3_bucket.example.id
# }

# Example output with sensitive data:
# output "database_password" {
#   description = "The password for the database"
#   value       = aws_db_instance.example.password
#   sensitive   = true
# }
`)

	return sb.String()
}

// Generates a simple README.md file
func generateReadmeMD(config *TerraformConfiguration) string {
	var sb strings.Builder

	sb.WriteString(`# Terraform Module

This module provisions AWS resources following best practices.

## Usage

```hcl
module "example" {
  source = "./path/to/module"

  region = "us-west-2"
  
  tags = {
    Environment = "production"
    Project     = "example"
  }
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0.0 |
| aws | >= 4.0.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| region | AWS region where resources will be created | `string` | `"us-west-2"` | no |
| tags | A map of tags to apply to all resources | `map(string)` | `{}` | no |

## Outputs

No outputs.

## Resources

No resources.
`)

	return sb.String()
}

// ParseTerraformConfiguration parses a Terraform configuration from a string map
func ParseTerraformConfiguration(files map[string]string) (*TerraformConfiguration, error) {
	config := &TerraformConfiguration{
		Files: files,
	}
	return config, nil
}

// TerraformTools implements Terraform configuration manipulation tools
type TerraformTools struct {
	ValidationEngine *ValidationEngine
}

// NewTerraformTools creates a new TerraformTools instance
func NewTerraformTools(engine *ValidationEngine) *TerraformTools {
	return &TerraformTools{
		ValidationEngine: engine,
	}
}

// FormatValidationResult formats a ValidationResult as a string
func FormatValidationResult(result *ValidationResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Validation summary: %d files analyzed, %d errors, %d warnings, %d info\n\n",
		result.FileCount, result.ErrorCount, result.WarnCount, result.InfoCount))

	if len(result.Issues) == 0 {
		sb.WriteString("No issues found!")
		return sb.String()
	}

	for i, issue := range result.Issues {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, issue.Severity, issue.Message))
		if issue.File != "" {
			sb.WriteString(fmt.Sprintf("   File: %s\n", issue.File))
		}
		if issue.BestPractice != "" {
			sb.WriteString(fmt.Sprintf("   Best Practice: %s\n", issue.BestPractice))
		}
		if issue.Suggestion != "" {
			sb.WriteString(fmt.Sprintf("   Suggestion: %s\n", issue.Suggestion))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatImprovementSuggestions formats improvement suggestions as a string
func FormatImprovementSuggestions(improvements map[string]string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Suggested improvements for %d files:\n\n", len(improvements)))

	for file, content := range improvements {
		sb.WriteString(fmt.Sprintf("File: %s\n", file))
		
		// Limit the content length for display
		preview := content
		if len(content) > 500 {
			preview = content[:500] + "...\n(content truncated for display)"
		}
		
		sb.WriteString("```\n")
		sb.WriteString(preview)
		sb.WriteString("\n```\n\n")
	}

	return sb.String()
}
</content>
