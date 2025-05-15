// tests/patterns_test.go
package tests

import (
	"testing"

	"terraform-mcp-server/pkg/hashicorp/tfdocs"
)

type mockLogger struct{}

func (l *mockLogger) Info(msg string, fields ...interface{})  {}
func (l *mockLogger) Error(msg string, fields ...interface{}) {}
func (l *mockLogger) Debug(msg string, fields ...interface{}) {}

func TestPatternRepository(t *testing.T) {
	// Create a temporary directory for the patterns
	tempDir := t.TempDir()

	// Create a mock logger
	logger := &mockLogger{}

	// Create a pattern repository
	repo := tfdocs.NewPatternRepository(tempDir, logger)

	// Initialize the repository
	err := repo.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pattern repository: %v", err)
	}

	// Get all patterns
	filter := tfdocs.PatternFilter{}
	patterns, err := repo.FindPatterns(filter)
	if err != nil {
		t.Fatalf("Failed to find patterns: %v", err)
	}

	// Verify that we have patterns
	if len(patterns) == 0 {
		t.Fatalf("Expected patterns, got none")
	}

	// Get a specific pattern
	pattern, err := repo.GetPatternByID(patterns[0].ID)
	if err != nil {
		t.Fatalf("Failed to get pattern by ID: %v", err)
	}

	// Verify that the pattern has the expected fields
	if pattern.ID == "" {
		t.Errorf("Expected pattern ID, got empty string")
	}
	if pattern.Name == "" {
		t.Errorf("Expected pattern name, got empty string")
	}
	if pattern.Description == "" {
		t.Errorf("Expected pattern description, got empty string")
	}
	if len(pattern.Files) == 0 {
		t.Errorf("Expected pattern files, got none")
	}

	// Test filtering by category
	category := patterns[0].Category
	filter = tfdocs.PatternFilter{
		Category: &category,
	}
	categoryPatterns, err := repo.FindPatterns(filter)
	if err != nil {
		t.Fatalf("Failed to find patterns by category: %v", err)
	}
	if len(categoryPatterns) == 0 {
		t.Errorf("Expected patterns for category %s, got none", category)
	}
	for _, p := range categoryPatterns {
		if p.Category != category {
			t.Errorf("Expected pattern with category %s, got %s", category, p.Category)
		}
	}

	// Test filtering by provider
	provider := patterns[0].Provider
	filter = tfdocs.PatternFilter{
		Provider: &provider,
	}
	providerPatterns, err := repo.FindPatterns(filter)
	if err != nil {
		t.Fatalf("Failed to find patterns by provider: %v", err)
	}
	if len(providerPatterns) == 0 {
		t.Errorf("Expected patterns for provider %s, got none", provider)
	}
	for _, p := range providerPatterns {
		if p.Provider != provider {
			t.Errorf("Expected pattern with provider %s, got %s", provider, p.Provider)
		}
	}

	// Test filtering by query
	filter = tfdocs.PatternFilter{
		Query: patterns[0].Name[:5], // Use first 5 chars of the name
	}
	queryPatterns, err := repo.FindPatterns(filter)
	if err != nil {
		t.Fatalf("Failed to find patterns by query: %v", err)
	}
	if len(queryPatterns) == 0 {
		t.Errorf("Expected patterns for query %s, got none", patterns[0].Name[:5])
	}
}
</content>
