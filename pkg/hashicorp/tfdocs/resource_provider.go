// pkg/hashicorp/tfdocs/resource_provider.go
package tfdocs

import (
	"context"
	"encoding/json"
	"fmt"
)

// ResourceProvider provides resources for MCP
type ResourceProvider struct {
	docIndexer *Indexer
	logger     Logger
}

// NewResourceProvider creates a new resource provider
func NewResourceProvider(indexer *Indexer, logger Logger) *ResourceProvider {
	return &ResourceProvider{
		docIndexer: indexer,
		logger:     logger,
	}
}

// Initialize initializes the resource provider
func (rp *ResourceProvider) Initialize() error {
	rp.logger.Info("Initializing resource provider")
	return nil
}

// ListResources lists resources matching a pattern
func (rp *ResourceProvider) ListResources(ctx context.Context, pattern string) ([]string, error) {
	rp.logger.Debug("Listing resources", "pattern", pattern)
	return rp.docIndexer.ListResources(ctx, pattern)
}

// GetResource gets a resource by URI
func (rp *ResourceProvider) GetResource(ctx context.Context, uri string) (json.RawMessage, error) {
	rp.logger.Debug("Getting resource", "uri", uri)
	return rp.docIndexer.GetResource(ctx, uri)
}
</content>
