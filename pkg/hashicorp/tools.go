// pkg/hashicorp/tools.go
package hashicorp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"terraform-mcp-server/pkg/hashicorp/tfdocs"
	"terraform-mcp-server/pkg/mcp"
)

// GetBestPracticesTool is a tool for retrieving best practice documentation
type GetBestPracticesTool struct {
	docIndexer       *tfdocs.Indexer
	resourceProvider *tfdocs.ResourceProvider
	logger           Logger
}

// GetBestPracticesArgs are the arguments for the GetBestPractices tool
type GetBestPracticesArgs struct {
	Topic     string   `json:"topic,omitempty"`
	Category  string   `json:"category,omitempty"`
	Provider  string   `json:"provider,omitempty"`
	Keywords  []string `json:"keywords,omitempty"`
}

// GetBestPracticesResult is the result of the GetBestPractices tool
type GetBestPracticesResult struct {
	Practices []BestPractice `json:"practices"`
}

// BestPractice represents a Terraform best practice
type BestPractice struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Provider    string   `json:"provider,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	References  []string `json:"references,omitempty"`
}

// NewGetBestPracticesTool creates a new GetBestPractices tool
func NewGetBestPracticesTool(indexer *tfdocs.Indexer, provider *tfdocs.ResourceProvider, logger Logger) *GetBestPracticesTool {
	return &GetBestPracticesTool{
		docIndexer:       indexer,
		resourceProvider: provider,
		logger:           logger,
	}
}

// Name returns the name of the tool
func (t *GetBestPracticesTool) Name() string {
	return "GetBestPractices"
}

// Describe returns a description of the tool
func (t *GetBestPracticesTool) Describe() mcp.ToolDescription {
	return mcp.ToolDescription{
		Name:        t.Name(),
		Description: "Retrieves Terraform best practices documentation, optionally filtered by topic, category, provider, or keywords",
		Parameters: map[string]mcp.ParameterDescription{
			"topic": {
				Type:        "string",
				Description: "The topic to filter by (e.g., 'module', 'security', 'naming')",
				Required:    false,
			},
			"category": {
				Type:        "string",
				Description: "The category to filter by (e.g., 'structure', 'organization', 'security')",
				Required:    false,
			},
			"provider": {
				Type:        "string",
				Description: "The provider to filter by (e.g., 'aws', 'azure', 'gcp')",
				Required:    false,
			},
			"keywords": {
				Type:        "array",
				Description: "Keywords to search for in best practices",
				Required:    false,
			},
		},
	}
}

// Execute executes the tool with the given arguments
func (t *GetBestPracticesTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a GetBestPracticesArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	t.logger.Debug("Executing GetBestPractices", "topic", a.Topic, "category", a.Category, "provider", a.Provider, "keywords", a.Keywords)

	practices, err := t.docIndexer.GetBestPractices(a.Topic, a.Category, a.Provider, a.Keywords)
	if err != nil {
		return nil, fmt.Errorf("failed to get best practices: %w", err)
	}

	// Convert to BestPractice structs
	var bestPractices []BestPractice
	for _, practice := range practices {
		bestPractices = append(bestPractices, BestPractice{
			ID:          practice.ID,
			Title:       practice.Title,
			Category:    practice.Category,
			Description: practice.Description,
			Content:     practice.Content,
			Provider:    practice.Provider,
			Tags:        practice.Tags,
			References:  practice.References,
		})
	}

	result := GetBestPracticesResult{
		Practices: bestPractices,
	}

	return json.Marshal(result)
}

// GetModuleStructureTool is a tool for retrieving module structure documentation
type GetModuleStructureTool struct {
	docIndexer       *tfdocs.Indexer
	resourceProvider *tfdocs.ResourceProvider
	logger           Logger
}

// GetModuleStructureArgs are the arguments for the GetModuleStructure tool
type GetModuleStructureArgs struct {
	Type     string `json:"type,omitempty"`
	Provider string `json:"provider,omitempty"`
}

// GetModuleStructureResult is the result of the GetModuleStructure tool
type GetModuleStructureResult struct {
	Structures []ModuleStructure `json:"structures"`
}

// ModuleStructure represents a Terraform module structure
type ModuleStructure struct {
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Files       []ModuleFile      `json:"files"`
	Examples    []string          `json:"examples,omitempty"`
	Provider    string            `json:"provider,omitempty"`
	References  []string          `json:"references,omitempty"`
}

// ModuleFile represents a file in a module structure
type ModuleFile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Content     string `json:"content,omitempty"`
}

// NewGetModuleStructureTool creates a new GetModuleStructure tool
func NewGetModuleStructureTool(indexer *tfdocs.Indexer, provider *tfdocs.ResourceProvider, logger Logger) *GetModuleStructureTool {
	return &GetModuleStructureTool{
		docIndexer:       indexer,
		resourceProvider: provider,
		logger:           logger,
	}
}

// Name returns the name of the tool
func (t *GetModuleStructureTool) Name() string {
	return "GetModuleStructure"
}

// Describe returns a description of the tool
func (t *GetModuleStructureTool) Describe() mcp.ToolDescription {
	return mcp.ToolDescription{
		Name:        t.Name(),
		Description: "Retrieves Terraform module structure documentation, optionally filtered by type or provider",
		Parameters: map[string]mcp.ParameterDescription{
			"type": {
				Type:        "string",
				Description: "The type of module to filter by (e.g., 'basic', 'advanced', 'nested')",
				Required:    false,
			},
			"provider": {
				Type:        "string",
				Description: "The provider to filter by (e.g., 'aws', 'azure', 'gcp')",
				Required:    false,
			},
		},
	}
}

// Execute executes the tool with the given arguments
func (t *GetModuleStructureTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a GetModuleStructureArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	t.logger.Debug("Executing GetModuleStructure", "type", a.Type, "provider", a.Provider)

	structures, err := t.docIndexer.GetModuleStructures(a.Type, a.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get module structures: %w", err)
	}

	// Convert to ModuleStructure structs
	var moduleStructures []ModuleStructure
	for _, structure := range structures {
		var files []ModuleFile
		for _, file := range structure.Files {
			files = append(files, ModuleFile{
				Name:        file.Name,
				Description: file.Description,
				Required:    file.Required,
				Content:     file.Content,
			})
		}

		moduleStructures = append(moduleStructures, ModuleStructure{
			Type:        structure.Type,
			Description: structure.Description,
			Files:       files,
			Examples:    structure.Examples,
			Provider:    structure.Provider,
			References:  structure.References,
		})
	}

	result := GetModuleStructureResult{
		Structures: moduleStructures,
	}

	return json.Marshal(result)
}

// GetPatternTemplateTool is a tool for retrieving code pattern templates
type GetPatternTemplateTool struct {
	patternRepo *tfdocs.PatternRepository
	logger      Logger
}

// GetPatternTemplateArgs are the arguments for the GetPatternTemplate tool
type GetPatternTemplateArgs struct {
	ID         string                  `json:"id,omitempty"`
	Category   *tfdocs.PatternCategory `json:"category,omitempty"`
	Provider   *tfdocs.CloudProvider   `json:"provider,omitempty"`
	Complexity *tfdocs.ComplexityLevel `json:"complexity,omitempty"`
	Tags       []string                `json:"tags,omitempty"`
	Query      string                  `json:"query,omitempty"`
}

// GetPatternTemplateResult is the result of the GetPatternTemplate tool
type GetPatternTemplateResult struct {
	Patterns []tfdocs.Pattern `json:"patterns"`
}

// NewGetPatternTemplateTool creates a new GetPatternTemplate tool
func NewGetPatternTemplateTool(repo *tfdocs.PatternRepository, logger Logger) *GetPatternTemplateTool {
	return &GetPatternTemplateTool{
		patternRepo: repo,
		logger:      logger,
	}
}

// Name returns the name of the tool
func (t *GetPatternTemplateTool) Name() string {
	return "GetPatternTemplate"
}

// Describe returns a description of the tool
func (t *GetPatternTemplateTool) Describe() mcp.ToolDescription {
	return mcp.ToolDescription{
		Name:        t.Name(),
		Description: "Retrieves Terraform code pattern templates, optionally filtered by ID, category, provider, complexity, tags, or search query",
		Parameters: map[string]mcp.ParameterDescription{
			"id": {
				Type:        "string",
				Description: "The ID of a specific pattern to retrieve",
				Required:    false,
			},
			"category": {
				Type:        "string",
				Description: "The category to filter by (e.g., 'compute', 'networking', 'storage')",
				Required:    false,
			},
			"provider": {
				Type:        "string",
				Description: "The cloud provider to filter by (e.g., 'aws', 'azure', 'gcp')",
				Required:    false,
			},
			"complexity": {
				Type:        "string",
				Description: "The complexity level to filter by (e.g., 'basic', 'intermediate', 'advanced')",
				Required:    false,
			},
			"tags": {
				Type:        "array",
				Description: "Tags to filter by",
				Required:    false,
			},
			"query": {
				Type:        "string",
				Description: "Text search query to filter patterns by name or description",
				Required:    false,
			},
		},
	}
}

// Execute executes the tool with the given arguments
func (t *GetPatternTemplateTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a GetPatternTemplateArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	t.logger.Debug("Executing GetPatternTemplate", 
		"id", a.ID, 
		"category", a.Category, 
		"provider", a.Provider, 
		"complexity", a.Complexity,
		"tags", a.Tags,
		"query", a.Query)

	var patterns []*tfdocs.Pattern
	var err error

	// If ID is specified, get the specific pattern
	if a.ID != "" {
		pattern, err := t.patternRepo.GetPatternByID(a.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get pattern: %w", err)
		}
		patterns = []*tfdocs.Pattern{pattern}
	} else {
		// Otherwise, use the filter criteria
		filter := tfdocs.PatternFilter{
			Category:   a.Category,
			Provider:   a.Provider,
			Complexity: a.Complexity,
			Tags:       a.Tags,
			Query:      a.Query,
		}

		patterns, err = t.patternRepo.FindPatterns(filter)
		if err != nil {
			return nil, fmt.Errorf("failed to find patterns: %w", err)
		}
	}

	// Convert to non-pointer pattern slice
	var resultPatterns []tfdocs.Pattern
	for _, pattern := range patterns {
		resultPatterns = append(resultPatterns, *pattern)
	}

	result := GetPatternTemplateResult{
		Patterns: resultPatterns,
	}

	return json.Marshal(result)
}

// ValidateConfigurationTool is a tool for validating Terraform configurations
type ValidateConfigurationTool struct {
	validationEngine *tfdocs.ValidationEngine
	logger           Logger
}

// ValidateConfigurationArgs are the arguments for the ValidateConfiguration tool
type ValidateConfigurationArgs struct {
	Files map[string]string `json:"files"`
}

// ValidateConfigurationResult is the result of the ValidateConfiguration tool
type ValidateConfigurationResult struct {
	Issues     []tfdocs.ValidationIssue `json:"issues"`
	Summary    ValidationSummary        `json:"summary"`
	Formatted  string                   `json:"formatted"`
	Successful bool                     `json:"successful"`
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	FileCount  int `json:"fileCount"`
	ErrorCount int `json:"errorCount"`
	WarnCount  int `json:"warnCount"`
	InfoCount  int `json:"infoCount"`
}

// NewValidateConfigurationTool creates a new ValidateConfiguration tool
func NewValidateConfigurationTool(engine *tfdocs.ValidationEngine, logger Logger) *ValidateConfigurationTool {
	return &ValidateConfigurationTool{
		validationEngine: engine,
		logger:           logger,
	}
}

// Name returns the name of the tool
func (t *ValidateConfigurationTool) Name() string {
	return "ValidateConfiguration"
}

// Describe returns a description of the tool
func (t *ValidateConfigurationTool) Describe() mcp.ToolDescription {
	return mcp.ToolDescription{
		Name:        t.Name(),
		Description: "Validates Terraform configurations against best practices and returns issues and improvement suggestions",
		Parameters: map[string]mcp.ParameterDescription{
			"files": {
				Type:        "object",
				Description: "Map of filenames to file contents to validate",
				Required:    true,
			},
		},
	}
}

// Execute executes the tool with the given arguments
func (t *ValidateConfigurationTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a ValidateConfigurationArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	t.logger.Debug("Executing ValidateConfiguration", "fileCount", len(a.Files))

	// Parse the configuration
	config, err := tfdocs.ParseTerraformConfiguration(a.Files)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Validate the configuration
	result, err := t.validationEngine.ValidateConfiguration(config)
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	// Format the validation result
	formatted := tfdocs.FormatValidationResult(result)

	// Prepare result
	validationResult := ValidateConfigurationResult{
		Issues: result.Issues,
		Summary: ValidationSummary{
			FileCount:  result.FileCount,
			ErrorCount: result.ErrorCount,
			WarnCount:  result.WarnCount,
			InfoCount:  result.InfoCount,
		},
		Formatted:  formatted,
		Successful: result.ErrorCount == 0,
	}

	return json.Marshal(validationResult)
}

// SuggestImprovementsTool is a tool for suggesting improvements to Terraform configurations
type SuggestImprovementsTool struct {
	validationEngine *tfdocs.ValidationEngine
	logger           Logger
}

// SuggestImprovementsArgs are the arguments for the SuggestImprovements tool
type SuggestImprovementsArgs struct {
	Files map[string]string `json:"files"`
}

// SuggestImprovementsResult is the result of the SuggestImprovements tool
type SuggestImprovementsResult struct {
	Improvements   map[string]string `json:"improvements"`
	FormattedGuide string            `json:"formattedGuide"`
}

// NewSuggestImprovementsTool creates a new SuggestImprovements tool
func NewSuggestImprovementsTool(engine *tfdocs.ValidationEngine, logger Logger) *SuggestImprovementsTool {
	return &SuggestImprovementsTool{
		validationEngine: engine,
		logger:           logger,
	}
}

// Name returns the name of the tool
func (t *SuggestImprovementsTool) Name() string {
	return "SuggestImprovements"
}

// Describe returns a description of the tool
func (t *SuggestImprovementsTool) Describe() mcp.ToolDescription {
	return mcp.ToolDescription{
		Name:        t.Name(),
		Description: "Suggests improvements to Terraform configurations based on best practices",
		Parameters: map[string]mcp.ParameterDescription{
			"files": {
				Type:        "object",
				Description: "Map of filenames to file contents to improve",
				Required:    true,
			},
		},
	}
}

// Execute executes the tool with the given arguments
func (t *SuggestImprovementsTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a SuggestImprovementsArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	t.logger.Debug("Executing SuggestImprovements", "fileCount", len(a.Files))

	// Parse the configuration
	config, err := tfdocs.ParseTerraformConfiguration(a.Files)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Generate improvements
	improvements, err := t.validationEngine.SuggestImprovements(config)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest improvements: %w", err)
	}

	// Format the improvement suggestions
	formattedGuide := tfdocs.FormatImprovementSuggestions(improvements)

	// Prepare result
	result := SuggestImprovementsResult{
		Improvements:   improvements,
		FormattedGuide: formattedGuide,
	}

	return json.Marshal(result)
}
</content>
