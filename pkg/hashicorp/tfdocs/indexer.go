// pkg/hashicorp/tfdocs/indexer.go
package tfdocs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ResourceType represents a type of resource
type ResourceType string

const (
	ResourceTypeBestPractice   ResourceType = "bestpractice"
	ResourceTypeModuleStructure ResourceType = "modulestructure"
)

// Resource represents a documentation resource
type Resource struct {
	URI     string          `json:"uri"`
	Type    ResourceType    `json:"type"`
	Content json.RawMessage `json:"content"`
}

// BestPractice represents a Terraform best practice
type BestPracticeDoc struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Provider    string   `json:"provider,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	References  []string `json:"references,omitempty"`
}

// ModuleStructureFile represents a file in a module structure
type ModuleStructureFile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Content     string `json:"content,omitempty"`
}

// ModuleStructureDoc represents a Terraform module structure
type ModuleStructureDoc struct {
	Type        string               `json:"type"`
	Description string               `json:"description"`
	Files       []ModuleStructureFile `json:"files"`
	Examples    []string             `json:"examples,omitempty"`
	Provider    string               `json:"provider,omitempty"`
	References  []string             `json:"references,omitempty"`
}

// IndexerOption is a function that configures an Indexer
type IndexerOption func(*Indexer)

// WithUpdateInterval sets the update interval for the indexer
func WithUpdateInterval(interval time.Duration) IndexerOption {
	return func(i *Indexer) {
		i.updateInterval = interval
	}
}

// WithAuthoritySources sets the authority sources for the indexer
func WithAuthoritySources(sources []string) IndexerOption {
	return func(i *Indexer) {
		i.authoritySources = sources
	}
}

// Indexer manages the indexing of Terraform documentation
type Indexer struct {
	docSourcePath    string
	resources        map[string]*Resource
	authoritySources []string
	updateInterval   time.Duration
	mutex            sync.RWMutex
	logger           Logger
}

// NewIndexer creates a new indexer
func NewIndexer(docSourcePath string, logger Logger, options ...IndexerOption) *Indexer {
	indexer := &Indexer{
		docSourcePath:    docSourcePath,
		resources:        make(map[string]*Resource),
		authoritySources: DefaultAuthoritySources,
		updateInterval:   24 * time.Hour,
		logger:           logger,
	}

	// Apply options
	for _, option := range options {
		option(indexer)
	}

	return indexer
}

// Initialize initializes the indexer
func (i *Indexer) Initialize(ctx context.Context) error {
	i.logger.Info("Initializing documentation indexer", "path", i.docSourcePath)

	// Create doc source directory if it doesn't exist
	if err := os.MkdirAll(i.docSourcePath, 0755); err != nil {
		return fmt.Errorf("failed to create doc source directory: %w", err)
	}

	// Check if index file exists
	indexPath := filepath.Join(i.docSourcePath, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		i.logger.Info("Index file not found, initializing with default documentation")
		return i.initializeDefaultDocs(ctx)
	}

	// Load index file
	data, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	var resources map[string]*Resource
	if err := json.Unmarshal(data, &resources); err != nil {
		return fmt.Errorf("failed to unmarshal index file: %w", err)
	}

	i.mutex.Lock()
	i.resources = resources
	i.mutex.Unlock()

	i.logger.Info("Documentation indexer initialized", "resourceCount", len(resources))
	return nil
}

// initializeDefaultDocs initializes the indexer with default documentation
func (i *Indexer) initializeDefaultDocs(ctx context.Context) error {
	i.logger.Info("Fetching documentation from authority sources", "count", len(i.authoritySources))

	// Create a channel for best practices
	bestPractices := make(chan BestPracticeDoc)
	moduleStructures := make(chan ModuleStructureDoc)
	errCh := make(chan error, len(i.authoritySources))

	// Start workers to fetch documentation
	var wg sync.WaitGroup
	for _, source := range i.authoritySources {
		wg.Add(1)
		go func(source string) {
			defer wg.Done()
			if err := i.fetchDocumentation(ctx, source, bestPractices, moduleStructures); err != nil {
				errCh <- fmt.Errorf("failed to fetch documentation from %s: %w", source, err)
			}
		}(source)
	}

	// Close channels when all workers are done
	go func() {
		wg.Wait()
		close(bestPractices)
		close(moduleStructures)
		close(errCh)
	}()

	// Initialize resources map
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.resources = make(map[string]*Resource)

	// Process best practices
	for practice := range bestPractices {
		// Generate URI
		uri := fmt.Sprintf("%s:%s/%s", ResourceTypeBestPractice, practice.Category, practice.ID)

		// Marshal to JSON
		content, err := json.Marshal(practice)
		if err != nil {
			i.logger.Error("Failed to marshal best practice", "id", practice.ID, "error", err)
			continue
		}

		// Add to resources
		i.resources[uri] = &Resource{
			URI:     uri,
			Type:    ResourceTypeBestPractice,
			Content: content,
		}
	}

	// Process module structures
	for structure := range moduleStructures {
		// Generate URI
		provider := structure.Provider
		if provider == "" {
			provider = "generic"
		}
		uri := fmt.Sprintf("%s:%s/%s", ResourceTypeModuleStructure, provider, structure.Type)

		// Marshal to JSON
		content, err := json.Marshal(structure)
		if err != nil {
			i.logger.Error("Failed to marshal module structure", "type", structure.Type, "error", err)
			continue
		}

		// Add to resources
		i.resources[uri] = &Resource{
			URI:     uri,
			Type:    ResourceTypeModuleStructure,
			Content: content,
		}
	}

	// Check for errors
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors while fetching documentation: %v", len(errs), errs)
	}

	// Generate default resources if no documentation was fetched
	if len(i.resources) == 0 {
		i.generateDefaultResources()
	}

	// Save index file
	indexPath := filepath.Join(i.docSourcePath, "index.json")
	indexData, err := json.MarshalIndent(i.resources, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index file: %w", err)
	}

	if err := ioutil.WriteFile(indexPath, indexData, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	i.logger.Info("Documentation initialized with default resources", "count", len(i.resources))
	return nil
}

// fetchDocumentation fetches documentation from a source URL
func (i *Indexer) fetchDocumentation(ctx context.Context, source string, bestPractices chan<- BestPracticeDoc, moduleStructures chan<- ModuleStructureDoc) error {
	i.logger.Debug("Fetching documentation", "source", source)

	// For now, we'll use a simple approach and just check if the source starts with http
	if strings.HasPrefix(source, "http") {
		// TODO: Implement HTTP fetching
		// This is a placeholder for future implementation
		return nil
	}

	// Otherwise, assume it's a local file
	// TODO: Implement local file loading
	// This is a placeholder for future implementation
	return nil
}

// generateDefaultResources creates default resources when no documentation is available
func (i *Indexer) generateDefaultResources() {
	i.logger.Info("Generating default resources")

	// Add default best practices
	i.addDefaultBestPractices()

	// Add default module structures
	i.addDefaultModuleStructures()
}

// addDefaultBestPractices adds default best practices
func (i *Indexer) addDefaultBestPractices() {
	// Module structure best practice
	moduleStructurePractice := BestPracticeDoc{
		ID:          "module-structure",
		Title:       "Standard Module Structure",
		Category:    "structure",
		Description: "Follow the standard module structure for Terraform modules",
		Content:     "Terraform modules should follow a standard structure with main.tf, variables.tf, outputs.tf, and README.md. This makes modules easier to understand, use, and maintain. The main.tf file should contain the primary resources, variables.tf should define all input variables, outputs.tf should define all outputs, and README.md should provide documentation on how to use the module. For larger modules, consider using additional files like providers.tf and versions.tf.",
		Tags:        []string{"modules", "structure", "organization"},
		References:  []string{"https://developer.hashicorp.com/terraform/language/modules/develop/structure"},
	}

	// Marshal to JSON
	content, err := json.Marshal(moduleStructurePractice)
	if err != nil {
		i.logger.Error("Failed to marshal best practice", "id", moduleStructurePractice.ID, "error", err)
		return
	}

	// Add to resources
	uri := fmt.Sprintf("%s:%s/%s", ResourceTypeBestPractice, moduleStructurePractice.Category, moduleStructurePractice.ID)
	i.resources[uri] = &Resource{
		URI:     uri,
		Type:    ResourceTypeBestPractice,
		Content: content,
	}

	// Variables documentation best practice
	variablesPractice := BestPracticeDoc{
		ID:          "variables-documentation",
		Title:       "Document Variables with Description",
		Category:    "documentation",
		Description: "Always include a description for all variables",
		Content:     "All variables in a Terraform module should include a description attribute that explains the purpose of the variable, expected values, and any constraints. This helps users understand how to use the module correctly. Additionally, variables should have an explicit type and, where appropriate, a default value or validation rules.",
		Tags:        []string{"variables", "documentation"},
		References:  []string{"https://developer.hashicorp.com/terraform/language/values/variables"},
	}

	// Marshal to JSON
	content, err = json.Marshal(variablesPractice)
	if err != nil {
		i.logger.Error("Failed to marshal best practice", "id", variablesPractice.ID, "error", err)
		return
	}

	// Add to resources
	uri = fmt.Sprintf("%s:%s/%s", ResourceTypeBestPractice, variablesPractice.Category, variablesPractice.ID)
	i.resources[uri] = &Resource{
		URI:     uri,
		Type:    ResourceTypeBestPractice,
		Content: content,
	}

	// Tagging best practice
	taggingPractice := BestPracticeDoc{
		ID:          "consistent-tagging",
		Title:       "Consistent Resource Tagging",
		Category:    "organization",
		Description: "Apply consistent tags to all resources",
		Content:     "Apply a consistent set of tags to all resources for easier management, cost allocation, and resource organization. Use a map variable for tags that can be set at the root module level and passed to all nested modules. This allows for centralized tag management and ensures consistency across resources. Consider implementing mandatory tags for environment, project, owner, and cost center.",
		Tags:        []string{"tagging", "organization"},
		References:  []string{"https://developer.hashicorp.com/terraform/tutorials/modules/pattern-module-composition"},
	}

	// Marshal to JSON
	content, err = json.Marshal(taggingPractice)
	if err != nil {
		i.logger.Error("Failed to marshal best practice", "id", taggingPractice.ID, "error", err)
		return
	}

	// Add to resources
	uri = fmt.Sprintf("%s:%s/%s", ResourceTypeBestPractice, taggingPractice.Category, taggingPractice.ID)
	i.resources[uri] = &Resource{
		URI:     uri,
		Type:    ResourceTypeBestPractice,
		Content: content,
	}

	// AWS specific best practice
	awsSecurityPractice := BestPracticeDoc{
		ID:          "security-group-rules",
		Title:       "Security Group Rules",
		Category:    "security",
		Description: "Follow security best practices for security group rules",
		Content:     "When defining security group rules, always follow the principle of least privilege. Avoid overly permissive rules such as allowing all ingress traffic (0.0.0.0/0) for ports other than HTTP/HTTPS. Use specific CIDR blocks or security group references instead. Separate security groups by function, and document each rule with a description attribute. Use a dedicated security groups module for reusable patterns.",
		Provider:    "aws",
		Tags:        []string{"security", "aws"},
		References:  []string{"https://docs.aws.amazon.com/vpc/latest/userguide/VPC_SecurityGroups.html"},
	}

	// Marshal to JSON
	content, err = json.Marshal(awsSecurityPractice)
	if err != nil {
		i.logger.Error("Failed to marshal best practice", "id", awsSecurityPractice.ID, "error", err)
		return
	}

	// Add to resources
	uri = fmt.Sprintf("%s:%s/%s", ResourceTypeBestPractice, awsSecurityPractice.Category, awsSecurityPractice.ID)
	i.resources[uri] = &Resource{
		URI:     uri,
		Type:    ResourceTypeBestPractice,
		Content: content,
	}

	// Version pinning best practice
	versionPinningPractice := BestPracticeDoc{
		ID:          "version-pinning",
		Title:       "Version Pinning",
		Category:    "stability",
		Description: "Pin provider and module versions for stability",
		Content:     "Always pin provider and module versions to ensure stability and predictability. Use the version attribute in the provider block to specify the provider version. For modules, use the version attribute in the module block to specify the module version. This prevents automatic updates that could introduce breaking changes. Use semantic versioning constraints to allow compatible updates while preventing breaking changes.",
		Tags:        []string{"versioning", "stability"},
		References:  []string{"https://developer.hashicorp.com/terraform/language/providers/requirements"},
	}

	// Marshal to JSON
	content, err = json.Marshal(versionPinningPractice)
	if err != nil {
		i.logger.Error("Failed to marshal best practice", "id", versionPinningPractice.ID, "error", err)
		return
	}

	// Add to resources
	uri = fmt.Sprintf("%s:%s/%s", ResourceTypeBestPractice, versionPinningPractice.Category, versionPinningPractice.ID)
	i.resources[uri] = &Resource{
		URI:     uri,
		Type:    ResourceTypeBestPractice,
		Content: content,
	}
}

// addDefaultModuleStructures adds default module structures
func (i *Indexer) addDefaultModuleStructures() {
	// Basic module structure
	basicModuleStructure := ModuleStructureDoc{
		Type:        "basic",
		Description: "Standard structure for a basic Terraform module",
		Files: []ModuleStructureFile{
			{
				Name:        "main.tf",
				Description: "Contains the main resources of the module",
				Required:    true,
				Content:     "# main.tf\n# Contains the main resources of the module\n\nresource \"aws_example\" \"this\" {\n  name = var.name\n  # other attributes\n}",
			},
			{
				Name:        "variables.tf",
				Description: "Contains the input variables for the module",
				Required:    true,
				Content:     "# variables.tf\n# Contains the input variables for the module\n\nvariable \"name\" {\n  description = \"The name to be used for resources created by this module\"\n  type        = string\n}\n\nvariable \"tags\" {\n  description = \"A map of tags to add to all resources\"\n  type        = map(string)\n  default     = {}\n}",
			},
			{
				Name:        "outputs.tf",
				Description: "Contains the outputs from the module",
				Required:    true,
				Content:     "# outputs.tf\n# Contains the outputs from the module\n\noutput \"id\" {\n  description = \"The ID of the resource\"\n  value       = aws_example.this.id\n}",
			},
			{
				Name:        "README.md",
				Description: "Contains documentation for the module",
				Required:    true,
				Content:     "# Example Module\n\nThis module manages an example resource.\n\n## Usage\n\n```hcl\nmodule \"example\" {\n  source = \"./example\"\n\n  name = \"example\"\n  tags = {\n    Environment = \"production\"\n  }\n}\n```\n\n## Requirements\n\n| Name | Version |\n|------|--------|\n| terraform | >= 1.0 |\n| aws | >= 4.0 |\n\n## Inputs\n\n| Name | Description | Type | Default | Required |\n|------|-------------|------|---------|:--------:|\n| name | The name to be used for resources created by this module | `string` | n/a | yes |\n| tags | A map of tags to add to all resources | `map(string)` | `{}` | no |\n\n## Outputs\n\n| Name | Description |\n|------|-------------|\n| id | The ID of the resource |",
			},
			{
				Name:        "versions.tf",
				Description: "Contains provider and terraform version constraints",
				Required:    false,
				Content:     "# versions.tf\n# Contains provider and terraform version constraints\n\nterraform {\n  required_version = \">= 1.0.0\"\n\n  required_providers {\n    aws = {\n      source  = \"hashicorp/aws\"\n      version = \">= 4.0.0\"\n    }\n  }\n}",
			},
		},
		Examples: []string{
			"module \"example\" {\n  source = \"./example\"\n\n  name = \"example\"\n  tags = {\n    Environment = \"production\"\n  }\n}",
		},
		References: []string{
			"https://developer.hashicorp.com/terraform/language/modules/develop/structure",
		},
	}

	// Marshal to JSON
	content, err := json.Marshal(basicModuleStructure)
	if err != nil {
		i.logger.Error("Failed to marshal module structure", "type", basicModuleStructure.Type, "error", err)
		return
	}

	// Add to resources
	uri := fmt.Sprintf("%s:generic/%s", ResourceTypeModuleStructure, basicModuleStructure.Type)
	i.resources[uri] = &Resource{
		URI:     uri,
		Type:    ResourceTypeModuleStructure,
		Content: content,
	}

	// AWS module structure
	awsModuleStructure := ModuleStructureDoc{
		Type:        "aws",
		Description: "Standard structure for an AWS-focused Terraform module",
		Files: []ModuleStructureFile{
			{
				Name:        "main.tf",
				Description: "Contains the main resources of the module",
				Required:    true,
				Content:     "# main.tf\n# Contains the main resources of the module\n\nresource \"aws_example\" \"this\" {\n  name = var.name\n  # other attributes\n}\n\nresource \"aws_security_group\" \"this\" {\n  name        = \"${var.name}-sg\"\n  description = \"Security group for ${var.name}\"\n  vpc_id      = var.vpc_id\n\n  tags = merge(\n    {\n      Name = \"${var.name}-sg\"\n    },\n    var.tags\n  )\n}",
			},
			{
				Name:        "variables.tf",
				Description: "Contains the input variables for the module",
				Required:    true,
				Content:     "# variables.tf\n# Contains the input variables for the module\n\nvariable \"name\" {\n  description = \"The name to be used for resources created by this module\"\n  type        = string\n}\n\nvariable \"vpc_id\" {\n  description = \"The ID of the VPC where resources will be created\"\n  type        = string\n}\n\nvariable \"tags\" {\n  description = \"A map of tags to add to all resources\"\n  type        = map(string)\n  default     = {}\n}",
			},
			{
				Name:        "outputs.tf",
				Description: "Contains the outputs from the module",
				Required:    true,
				Content:     "# outputs.tf\n# Contains the outputs from the module\n\noutput \"id\" {\n  description = \"The ID of the resource\"\n  value       = aws_example.this.id\n}\n\noutput \"security_group_id\" {\n  description = \"The ID of the security group\"\n  value       = aws_security_group.this.id\n}",
			},
			{
				Name:        "README.md",
				Description: "Contains documentation for the module",
				Required:    true,
				Content:     "# AWS Example Module\n\nThis module manages AWS resources.\n\n## Usage\n\n```hcl\nmodule \"example\" {\n  source = \"./example\"\n\n  name   = \"example\"\n  vpc_id = \"vpc-12345678\"\n  tags   = {\n    Environment = \"production\"\n  }\n}\n```\n\n## Requirements\n\n| Name | Version |\n|------|--------|\n| terraform | >= 1.0 |\n| aws | >= 4.0 |\n\n## Inputs\n\n| Name | Description | Type | Default | Required |\n|------|-------------|------|---------|:--------:|\n| name | The name to be used for resources created by this module | `string` | n/a | yes |\n| vpc_id | The ID of the VPC where resources will be created | `string` | n/a | yes |\n| tags | A map of tags to add to all resources | `map(string)` | `{}` | no |\n\n## Outputs\n\n| Name | Description |\n|------|-------------|\n| id | The ID of the resource |\n| security_group_id | The ID of the security group |",
			},
			{
				Name:        "versions.tf",
				Description: "Contains provider and terraform version constraints",
				Required:    true,
				Content:     "# versions.tf\n# Contains provider and terraform version constraints\n\nterraform {\n  required_version = \">= 1.0.0\"\n\n  required_providers {\n    aws = {\n      source  = \"hashicorp/aws\"\n      version = \">= 4.0.0\"\n    }\n  }\n}",
			},
		},
		Provider: "aws",
		Examples: []string{
			"module \"example\" {\n  source = \"./example\"\n\n  name   = \"example\"\n  vpc_id = \"vpc-12345678\"\n  tags   = {\n    Environment = \"production\"\n  }\n}",
		},
		References: []string{
			"https://developer.hashicorp.com/terraform/language/modules/develop/structure",
			"https://registry.terraform.io/providers/hashicorp/aws/latest/docs",
		},
	}

	// Marshal to JSON
	content, err = json.Marshal(awsModuleStructure)
	if err != nil {
		i.logger.Error("Failed to marshal module structure", "type", awsModuleStructure.Type, "error", err)
		return
	}

	// Add to resources
	uri = fmt.Sprintf("%s:%s/%s", ResourceTypeModuleStructure, awsModuleStructure.Provider, awsModuleStructure.Type)
	i.resources[uri] = &Resource{
		URI:     uri,
		Type:    ResourceTypeModuleStructure,
		Content: content,
	}
}

// ListResources lists resources matching a pattern
func (i *Indexer) ListResources(ctx context.Context, pattern string) ([]string, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	var uris []string
	for uri := range i.resources {
		if strings.HasPrefix(uri, pattern) {
			uris = append(uris, uri)
		}
	}

	return uris, nil
}

// GetResource gets a resource by URI
func (i *Indexer) GetResource(ctx context.Context, uri string) (json.RawMessage, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	resource, ok := i.resources[uri]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", uri)
	}

	return resource.Content, nil
}

// GetBestPractices gets best practices
func (i *Indexer) GetBestPractices(topic, category, provider string, keywords []string) ([]BestPracticeDoc, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	var practices []BestPracticeDoc
	for uri, resource := range i.resources {
		if resource.Type != ResourceTypeBestPractice {
			continue
		}

		// Parse the resource content
		var practice BestPracticeDoc
		if err := json.Unmarshal(resource.Content, &practice); err != nil {
			i.logger.Error("Failed to unmarshal best practice", "uri", uri, "error", err)
			continue
		}

		// Apply filters
		if topic != "" && !strings.Contains(strings.ToLower(practice.Title), strings.ToLower(topic)) && !strings.Contains(strings.ToLower(practice.Description), strings.ToLower(topic)) {
			continue
		}

		if category != "" && practice.Category != category {
			continue
		}

		if provider != "" && practice.Provider != provider {
			continue
		}

		if len(keywords) > 0 {
			match := false
			for _, keyword := range keywords {
				keyword = strings.ToLower(keyword)
				if strings.Contains(strings.ToLower(practice.Title), keyword) ||
					strings.Contains(strings.ToLower(practice.Description), keyword) ||
					strings.Contains(strings.ToLower(practice.Content), keyword) {
					match = true
					break
				}

				// Check tags
				for _, tag := range practice.Tags {
					if strings.Contains(strings.ToLower(tag), keyword) {
						match = true
						break
					}
				}
			}

			if !match {
				continue
			}
		}

		practices = append(practices, practice)
	}

	return practices, nil
}

// GetModuleStructures gets module structures
func (i *Indexer) GetModuleStructures(structureType, provider string) ([]ModuleStructureDoc, error) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	var structures []ModuleStructureDoc
	for uri, resource := range i.resources {
		if resource.Type != ResourceTypeModuleStructure {
			continue
		}

		// Parse the resource content
		var structure ModuleStructureDoc
		if err := json.Unmarshal(resource.Content, &structure); err != nil {
			i.logger.Error("Failed to unmarshal module structure", "uri", uri, "error", err)
			continue
		}

		// Apply filters
		if structureType != "" && structure.Type != structureType {
			continue
		}

		if provider != "" && structure.Provider != provider {
			continue
		}

		structures = append(structures, structure)
	}

	return structures, nil
}
</content>
