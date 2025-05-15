// pkg/hashicorp/tfdocs/indexer.go
package tfdocs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Default authority sources for Terraform best practices
var DefaultAuthoritySources = []string{
	"https://developer.hashicorp.com/terraform/language/modules/develop",
	"https://developer.hashicorp.com/terraform/language/style",
	"https://developer.hashicorp.com/validated-designs/terraform-operating-guides-adoption/terraform-workflows",
	"https://developer.hashicorp.com/terraform/tutorials/pro-cert/pro-review",
}

// Document represents a documentation document for Terraform best practices
type Document struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	URL         string            `json:"url"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
	LastUpdated time.Time         `json:"last_updated"`
}

// Indexer is responsible for indexing Terraform documentation
type Indexer struct {
	documents       map[string]Document
	categories      map[string][]string
	tags            map[string][]string
	mu              sync.RWMutex
	sourcePath      string
	updateInterval  time.Duration
	updateTicker    *time.Ticker
	logger          Logger
	lastIndexed     time.Time
	indexingActive  bool
	authoritySources []string
}

// Logger defines a simple interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// IndexerOption configures the Indexer
type IndexerOption func(*Indexer)

// WithUpdateInterval sets the update interval
func WithUpdateInterval(interval time.Duration) IndexerOption {
	return func(i *Indexer) {
		i.updateInterval = interval
	}
}

// WithAuthoritySources sets the authority sources
func WithAuthoritySources(sources []string) IndexerOption {
	return func(i *Indexer) {
		if len(sources) > 0 {
			i.authoritySources = sources
		}
	}
}

// NewIndexer creates a new documentation indexer
func NewIndexer(sourcePath string, logger Logger, opts ...IndexerOption) *Indexer {
	i := &Indexer{
		documents:      make(map[string]Document),
		categories:     make(map[string][]string),
		tags:           make(map[string][]string),
		sourcePath:     sourcePath,
		logger:         logger,
		updateInterval: 24 * time.Hour,
		authoritySources: DefaultAuthoritySources,
	}
	
	for _, opt := range opts {
		opt(i)
	}
	
	return i
}

// Start begins the indexing process and periodic updates
func (i *Indexer) Start(ctx context.Context) error {
	i.logger.Info("Starting documentation indexer", "sourcePath", i.sourcePath)
	
	if err := i.Index(ctx); err != nil {
		return fmt.Errorf("initial indexing failed: %w", err)
	}
	
	i.updateTicker = time.NewTicker(i.updateInterval)
	go func() {
		for {
			select {
			case <-i.updateTicker.C:
				i.logger.Debug("Running scheduled index update")
				if err := i.Index(ctx); err != nil {
					i.logger.Error("Scheduled index update failed", "error", err)
				}
			case <-ctx.Done():
				i.logger.Info("Stopping documentation indexer")
				i.updateTicker.Stop()
				return
			}
		}
	}()
	
	return nil
}

// Index indexes the documentation from the source path
func (i *Indexer) Index(ctx context.Context) error {
	i.mu.Lock()
	if i.indexingActive {
		i.mu.Unlock()
		i.logger.Debug("Indexing already in progress, skipping")
		return nil
	}
	
	i.indexingActive = true
	i.mu.Unlock()
	
	defer func() {
		i.mu.Lock()
		i.indexingActive = false
		i.mu.Unlock()
	}()
	
	i.logger.Info("Indexing documentation", "sourcePath", i.sourcePath)
	
	// First, try loading from the file system
	newDocs, err := i.loadFromFileSystem()
	if err != nil {
		i.logger.Error("Failed to load from file system", "error", err)
		// Continue anyway, we might have web sources
	}
	
	// Then, try to fetch from authority web sources
	webDocs, err := i.fetchFromAuthoritySources(ctx)
	if err != nil {
		i.logger.Error("Failed to fetch from web sources", "error", err)
		// Continue anyway, we might have file sources
	}
	
	// Merge the documents
	i.mu.Lock()
	defer i.mu.Unlock()
	
	// Reset the collections
	i.documents = make(map[string]Document)
	i.categories = make(map[string][]string)
	i.tags = make(map[string][]string)
	
	// Add file system documents
	for id, doc := range newDocs {
		i.documents[id] = doc
	}
	
	// Add web documents, potentially overriding file system documents
	for id, doc := range webDocs {
		i.documents[id] = doc
	}
	
	// Rebuild category and tag indices
	for id, doc := range i.documents {
		if _, exists := i.categories[doc.Category]; !exists {
			i.categories[doc.Category] = []string{}
		}
		i.categories[doc.Category] = append(i.categories[doc.Category], id)
		
		for _, tag := range doc.Tags {
			if _, exists := i.tags[tag]; !exists {
				i.tags[tag] = []string{}
			}
			i.tags[tag] = append(i.tags[tag], id)
		}
	}
	
	i.lastIndexed = time.Now()
	i.logger.Info("Indexing complete", "documentCount", len(i.documents), "authoritySources", len(i.authoritySources))
	
	return nil
}

// loadFromFileSystem loads documents from the file system
func (i *Indexer) loadFromFileSystem() (map[string]Document, error) {
	docs := make(map[string]Document)
	
	files, err := ioutil.ReadDir(i.sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		path := filepath.Join(i.sourcePath, file.Name())
		data, err := ioutil.ReadFile(path)
		if err != nil {
			i.logger.Error("Failed to read file", "path", path, "error", err)
			continue
		}
		
		var doc Document
		if err := json.Unmarshal(data, &doc); err != nil {
			i.logger.Error("Failed to unmarshal document", "path", path, "error", err)
			continue
		}
		
		// Use the filename without extension as the ID if not set
		if doc.ID == "" {
			doc.ID = strings.TrimSuffix(file.Name(), ".json")
		}
		
		docs[doc.ID] = doc
	}
	
	return docs, nil
}

// fetchFromAuthoritySources fetches documents from authority web sources
func (i *Indexer) fetchFromAuthoritySources(ctx context.Context) (map[string]Document, error) {
	docs := make(map[string]Document)
	
	for _, url := range i.authoritySources {
		// This is a simplified example - in a real implementation, 
		// you would use a proper HTML parser and extract structured data
		
		// Generate an ID from the URL
		id := GenerateIDFromURL(url)
		
		// Set up a descriptive title based on the URL
		title := "Terraform Best Practices"
		if strings.Contains(url, "modules/develop") {
			title = "Terraform Module Development Guide"
		} else if strings.Contains(url, "style") {
			title = "Terraform Style Guide"
		} else if strings.Contains(url, "workflows") {
			title = "Terraform Workflows"
		} else if strings.Contains(url, "pro-cert") {
			title = "Terraform Authoring and Operations Pro"
		}
		
		// Set up tags and category based on URL pattern
		tags := []string{"best-practice", "hashicorp"}
		category := "terraform"
		
		if strings.Contains(url, "modules") {
			tags = append(tags, "module", "structure")
			category = "module-structure"
		} else if strings.Contains(url, "style") {
			tags = append(tags, "style", "conventions")
			category = "style"
		} else if strings.Contains(url, "workflows") {
			tags = append(tags, "workflow", "process")
			category = "workflow"
		} else if strings.Contains(url, "pro-cert") {
			tags = append(tags, "certification", "best-practice")
			category = "certification"
		}
		
		doc := Document{
			ID:       id,
			Title:    title,
			URL:      url,
			Category: category,
			Tags:     tags,
			Metadata: map[string]string{
				"source": "HashiCorp",
			},
			LastUpdated: time.Now(),
		}
		
		// In a real implementation, you would fetch and parse the content
		resp, err := http.Get(url)
		if err != nil {
			i.logger.Error("Failed to fetch document", "url", url, "error", err)
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			i.logger.Error("Failed to fetch document", "url", url, "statusCode", resp.StatusCode)
			continue
		}
		
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			i.logger.Error("Failed to read document content", "url", url, "error", err)
			continue
		}
		
		// Set the content - in a real implementation, you would parse the HTML
		doc.Content = extractContentFromHTML(string(content))
		
		// Add to the collection
		docs[doc.ID] = doc
	}
	
	return docs, nil
}

// GenerateIDFromURL generates a unique ID from a URL
func GenerateIDFromURL(url string) string {
	// Strip protocol
	id := strings.ReplaceAll(
		strings.TrimPrefix(
			strings.TrimPrefix(url, "https://"), 
			"http://"),
		"/", "-")
	
	// Further cleanup to make a nice ID
	id = strings.ReplaceAll(id, ".", "-")
	id = strings.ReplaceAll(id, "--", "-")
	
	// Trim any trailing dash
	id = strings.TrimSuffix(id, "-")
	
	return id
}

// extractContentFromHTML extracts the main content from HTML
// This is a simplified implementation - in a real implementation, you would use a proper HTML parser
func extractContentFromHTML(html string) string {
	// Find main content area - this is very simplistic and would need to be improved
	// with a proper HTML parser in a real implementation
	mainStart := strings.Index(html, "<main")
	if mainStart == -1 {
		mainStart = strings.Index(html, "<article")
	}
	if mainStart == -1 {
		mainStart = strings.Index(html, "<div class=\"content")
	}
	if mainStart == -1 {
		return "Failed to extract content from HTML"
	}
	
	mainEnd := strings.Index(html[mainStart:], "</main>")
	if mainEnd == -1 {
		mainEnd = strings.Index(html[mainStart:], "</article>")
	}
	if mainEnd == -1 {
		mainEnd = strings.Index(html[mainStart:], "</div>")
	}
	if mainEnd == -1 {
		return "Failed to extract content from HTML"
	}
	
	content := html[mainStart:mainStart+mainEnd]
	
	// Strip HTML tags (very simplistic approach)
	for {
		tagStart := strings.Index(content, "<")
		if tagStart == -1 {
			break
		}
		
		tagEnd := strings.Index(content[tagStart:], ">")
		if tagEnd == -1 {
			break
		}
		
		content = content[:tagStart] + " " + content[tagStart+tagEnd+1:]
	}
	
	// Clean up whitespace
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\t", " ")
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}
	
	return strings.TrimSpace(content)
}

// GetDocument returns a document by ID
func (i *Indexer) GetDocument(id string) (Document, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	
	doc, exists := i.documents[id]
	return doc, exists
}

// FindDocumentsByCategory returns documents in a category
func (i *Indexer) FindDocumentsByCategory(category string) []Document {
	i.mu.RLock()
	defer i.mu.RUnlock()
	
	var docs []Document
	for _, id := range i.categories[category] {
		if doc, exists := i.documents[id]; exists {
			docs = append(docs, doc)
		}
	}
	
	return docs
}

// FindDocumentsByTag returns documents with a specific tag
func (i *Indexer) FindDocumentsByTag(tag string) []Document {
	i.mu.RLock()
	defer i.mu.RUnlock()
	
	var docs []Document
	for _, id := range i.tags[tag] {
		if doc, exists := i.documents[id]; exists {
			docs = append(docs, doc)
		}
	}
	
	return docs
}

// GetAllDocuments returns all indexed documents
func (i *Indexer) GetAllDocuments() []Document {
	i.mu.RLock()
	defer i.mu.RUnlock()
	
	docs := make([]Document, 0, len(i.documents))
	for _, doc := range i.documents {
		docs = append(docs, doc)
	}
	
	return docs
}

// GetAuthoritySources returns the configured authority sources
func (i *Indexer) GetAuthoritySources() []string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	
	// Return a copy to avoid external modification
	sources := make([]string, len(i.authoritySources))
	copy(sources, i.authoritySources)
	
	return sources
}