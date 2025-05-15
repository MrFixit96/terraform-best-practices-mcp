// cmd/terraform-mcp-server/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"terraform-mcp-server/pkg/hashicorp"
)

// Configuration options
type config struct {
	Addr            string
	DocSourcePath   string
	PatternPath     string
	DataDir         string
	UpdateInterval  time.Duration
	LogLevel        string
}

func main() {
	// Parse command line arguments
	cfg := parseFlags()
	
	// Initialize server
	logger := &hashicorp.DefaultLogger{
		Logger: log.New(os.Stdout, "terraform-mcp: ", log.LstdFlags),
	}
	
	logger.Info("Starting Terraform MCP Server")
	
	// Create server configuration
	serverConfig := hashicorp.Config{
		DocSourcePath:  cfg.DocSourcePath,
		PatternPath:    cfg.PatternPath,
		UpdateInterval: cfg.UpdateInterval,
	}
	
	// Create server
	server, err := hashicorp.NewServer(serverConfig, logger)
	if err != nil {
		logger.Error("Failed to create server", "error", err)
		os.Exit(1)
	}
	
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-signalCh
		logger.Info("Received shutdown signal")
		cancel()
	}()
	
	// Initialize server
	if err := server.Initialize(ctx); err != nil {
		logger.Error("Failed to initialize server", "error", err)
		os.Exit(1)
	}
	
	// Start HTTP server
	logger.Info("Starting HTTP server", "addr", cfg.Addr)
	if err := server.ListenAndServe(cfg.Addr); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}

// parseFlags parses the command line flags
func parseFlags() config {
	cfg := config{}
	
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to get executable path:", err)
	}
	exeDir := filepath.Dir(exePath)
	defaultDataDir := filepath.Join(exeDir, "data")
	
	// Define flags
	flag.StringVar(&cfg.Addr, "addr", ":8080", "Server address")
	flag.StringVar(&cfg.DataDir, "data-dir", defaultDataDir, "Data directory")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, error)")
	flag.DurationVar(&cfg.UpdateInterval, "update-interval", 24*time.Hour, "Update interval for documentation")
	
	// Parse flags
	flag.Parse()
	
	// Derive paths from data directory
	cfg.DocSourcePath = filepath.Join(cfg.DataDir, "docs")
	cfg.PatternPath = filepath.Join(cfg.DataDir, "patterns")
	
	return cfg
}