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

// GetBestPracticesToolTool is a tool for retrieving best practice documentation
type GetBestPracticesToolTool struct {
	resourceProvider *tfdocs.ResourceProvider
	logger           Logger
}

// GetBestPracticesArgs are the arguments for the GetBestPractices tool
type GetBestPracticesArgs struct {
	Topic string `json:"topic,omitempty"`
}

// GetBestPracticesResult is the result of the GetBestPractices tool
type GetBestPracticesResult struct {
	Practices []json.RawMessage `json:"practices"`
}

// NewGetBestPracticesToolTool creates a new GetBestPractices tool
func NewGetBestPracticesToolTool(provider *tfdocs.ResourceProvider, logger Logger) *GetBestPracticesToolTool {
	return &GetBestPracticesToolTool{
		resourceProvider: provider,
		logger:           logger,
	}
}

// Name returns the name of the tool
func (t *GetBestPracticesToolTool) Name() string {
	return "GetBestPractices"
}

// Execute executes the tool with the given arguments
func (t *GetBestPracticesToolTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a GetBestPracticesArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	t.logger.Debug("Executing GetBestPractices", "topic", a.Topic)

	// Get resources based on the topic
	pattern := string(tfdocs.ResourceTypeBestPractice) + ":"
	if a.Topic != "" {
		pattern += a.Topic + "/"
	}

	uris, err := t.resourceProvider.ListResources(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	var practices []json.RawMessage
	for _, uri := range uris {
		resource, err := t.resourceProvider.GetResource(ctx, uri)
		if err != nil {
			t.logger.Error("Failed to get resource", "uri", uri, "error", err)
			continue
		}

		practices = append(practices, resource)
	}

	result := GetBestPracticesResult{
		Practices: practices,
	}

	return json.Marshal(result)
}

// GetModuleStructureTool is a tool for retrieving module structure documentation
type GetModuleStructureTool struct {
	resourceProvider *tfdocs.ResourceProvider
	logger           Logger
}

// GetModuleStructureArgs are the arguments for the GetModuleStructure tool
type GetModuleStructureArgs struct {
	Type string `json:"type,omitempty"`
}

// GetModuleStructureResult is the result of the GetModuleStructure tool
type GetModuleStructureResult struct {
	Structures []json.RawMessage `json:"structures"`
}

// NewGetModuleStructureTool creates a new GetModuleStructure tool
func NewGetModuleStructureTool(provider *tfdocs.ResourceProvider, logger Logger) *GetModuleStructureTool {
	return &GetModuleStructureTool{
		resourceProvider: provider,
		logger:           logger,
	}
}

// Name returns the name of the tool
func (t *GetModuleStructureTool) Name() string {
	return "GetModuleStructure"
}

// Execute executes the tool with the given arguments
func (t *GetModuleStructureTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a GetModuleStructureArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	t.logger.Debug("Executing GetModuleStructure", "type", a.Type)

	// Get resources based on the type
	pattern := string(tfdocs.ResourceTypeModuleStructure) + ":"
	if a.Type != "" {
		pattern += a.Type + "/"
	}

	uris, err := t.resourceProvider.ListResources(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	var structures []json.RawMessage
	for _, uri := range uris {
		resource, err := t.resourceProvider.GetResource(ctx, uri)
		if err != nil {
			t.logger.Error("Failed to get resource", "uri", uri, "error", err)
			continue
		}

		structures = append(structures, resource)
	}

	result := GetModuleStructureResult{
		Structures: structures,
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
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
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

// Execute executes the tool with the given arguments
func (t *GetPatternTemplateTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a GetPatternTemplateArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	t.logger.Debug("Executing GetPatternTemplate", "category", a.Category, "tags", a.Tags)

	var patterns []tfdocs.Pattern

	// Filter patterns based on the arguments
	if a.Category != "" {
		patterns = t.patternRepo.FindPatternsByCategory(a.Category)
	} else if len(a.Tags) > 0 {
		// For each tag, get the patterns and merge
		patternMap := make(map[string]tfdocs.Pattern)
		
		for _, tag := range a.Tags {
			tagPatterns := t.patternRepo.FindPatternsByTag(tag)
			for _, pattern := range tagPatterns {
				patternMap[pattern.ID] = pattern
			}
		}

		// Convert map to slice
		for _, pattern := range patternMap {
			patterns = append(patterns, pattern)
		}
	} else {
		// No filters, return all patterns
		patterns = t.patternRepo.GetAllPatterns()
	}

	result := GetPatternTemplateResult{
		Patterns: patterns,
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
	Results []tfdocs.ValidationResult `json:"results"`
	Summary ValidationSummary         `json:"summary"`
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	TotalCount     int `json:"totalCount"`
	PassedCount    int `json:"passedCount"`
	FailedCount    int `json:"failedCount"`
	ErrorCount     int `json:"errorCount"`
	WarningCount   int `json:"warningCount"`
	InfoCount      int `json:"infoCount"`
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

// Execute executes the tool with the given arguments
func (t *ValidateConfigurationTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a ValidateConfigurationArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	t.logger.Debug("Executing ValidateConfiguration", "fileCount", len(a.Files))

	// Normalize file names
	normalizedFiles := make(map[string]string)
	for name, content := range a.Files {
		// Convert to lowercase for case-insensitive comparison
		normalizedName := strings.ToLower(name)
		
		// Handle common variations in file naming
		normalizedName = strings.TrimSuffix(normalizedName, ".tf")
		
		// Map common file names to standard names
		switch {
		case normalizedName == "main" || normalizedName == "resources":
			normalizedFiles["main.tf"] = content
		case normalizedName == "variables" || normalizedName == "vars":
			normalizedFiles["variables.tf"] = content
		case normalizedName == "outputs" || normalizedName == "output":
			normalizedFiles["outputs.tf"] = content
		case normalizedName == "readme" || normalizedName == "readme.md":
			normalizedFiles["README.md"] = content
		default:
			// Keep original name if no mapping
			normalizedFiles[name] = content
		}
	}

	// Validate the configuration
	results, err := t.validationEngine.ValidateConfiguration(ctx, normalizedFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	// Calculate summary statistics
	summary := ValidationSummary{
		TotalCount: len(results),
	}

	for _, result := range results {
		if result.Passed {
			summary.PassedCount++
		} else {
			summary.FailedCount++
			
			switch result.Rule.Severity {
			case "error":
				summary.ErrorCount++
			case "warning":
				summary.WarningCount++
			case "info":
				summary.InfoCount++
			}
		}
	}

	// Prepare result
	validationResult := ValidateConfigurationResult{
		Results: results,
		Summary: summary,
	}

	return json.Marshal(validationResult)
}