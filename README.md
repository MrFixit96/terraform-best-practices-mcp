# Terraform Best Practices MCP Server

This Model Context Protocol (MCP) server exposes Terraform best practices, module structure recommendations, code pattern templates, and validation tools to AI assistants.

## Overview

The Terraform Best Practices MCP server provides AI assistants with access to:

1. **Best Practices Documentation**: HashiCorp's Terraform best practices for code structure and organization
2. **Module Structure Guidelines**: Standardized module structure recommendations for different use cases
3. **Pattern Library**: Reusable code pattern templates that follow best practices
4. **Validation Engine**: Tools to validate Terraform configurations against best practices

## Features

### 1. Best Practices Documentation

Provides access to Terraform best practices on:
- Module structure and organization
- Variable and output documentation
- Resource naming conventions
- Security best practices
- Tagging strategies
- Version pinning

### 2. Module Structure Guidelines

Offers standardized module structures for:
- Basic Terraform modules
- AWS-specific modules
- Azure-specific modules
- GCP-specific modules

### 3. Pattern Library

Provides code templates for common infrastructure components:
- AWS VPC with public and private subnets
- EC2 web server with security groups
- Azure Virtual Network
- GCP VPC
- Standard Terraform module structure

### 4. Validation Engine

Validates Terraform configurations against best practices:
- File structure validation
- Naming convention checks
- Security best practices validation
- Documentation completeness checks
- Module usage validation
- Resource organization validation

## Installation

### Prerequisites

- Go 1.17 or later
- Git

### Platform-Specific Setup Instructions

#### Windows (PowerShell 7)

1. **Install Prerequisites with Windows Package Manager (winget)**

   Install Go:
   ```powershell
   # Install Go using winget
   winget install GoLang.Go
   
   # Verify installation
   go version
   ```

   Install Git:
   ```powershell
   # Install Git using winget
   winget install Git.Git
   
   # Verify installation
   git --version
   ```

   Alternative using Scoop package manager:
   ```powershell
   # Install Scoop if not already installed
   if (-not (Get-Command scoop -ErrorAction SilentlyContinue)) {
     Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
     Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression
   }
   
   # Install Go and Git using Scoop
   scoop install go git
   
   # Verify installations
   go version
   git --version
   ```

2. **Clone and Build**

   ```powershell
   # Clone the repository
   git clone https://github.com/MrFixit96/terraform-best-practices-mcp.git
   cd terraform-best-practices-mcp
   
   # Build the binary
   go build -o terraform-mcp-server.exe .\cmd\terraform-mcp-server
   ```

3. **Run the Server**

   ```powershell
   # Start the server on port 8080
   .\terraform-mcp-server.exe -addr :8080
   
   # With custom data directory
   .\terraform-mcp-server.exe -addr :8080 -data-dir .\my-data
   
   # With debug logging
   .\terraform-mcp-server.exe -addr :8080 -log-level debug
   ```

#### macOS

1. **Install Prerequisites**

   Install Go:
   ```bash
   # Use Homebrew
   brew install go
   
   # Verify installation
   go version
   ```

   Install Git:
   ```bash
   # Use Homebrew
   brew install git
   
   # Verify installation
   git --version
   ```

2. **Clone and Build**

   ```bash
   # Clone the repository
   git clone https://github.com/MrFixit96/terraform-best-practices-mcp.git
   cd terraform-best-practices-mcp
   
   # Build the binary
   go build -o terraform-mcp-server ./cmd/terraform-mcp-server
   ```

3. **Run the Server**

   ```bash
   # Start the server on port 8080
   ./terraform-mcp-server -addr :8080
   
   # With custom data directory
   ./terraform-mcp-server -addr :8080 -data-dir ./my-data
   
   # With debug logging
   ./terraform-mcp-server -addr :8080 -log-level debug
   ```

#### Linux (Bash)

1. **Install Prerequisites**

   Install Go:
   ```bash
   # Debian/Ubuntu
   sudo apt update
   sudo apt install golang-go
   
   # RHEL/CentOS/Fedora
   sudo dnf install golang
   
   # Verify installation
   go version
   ```

   Install Git:
   ```bash
   # Debian/Ubuntu
   sudo apt update
   sudo apt install git
   
   # RHEL/CentOS/Fedora
   sudo dnf install git
   
   # Verify installation
   git --version
   ```

2. **Clone and Build**

   ```bash
   # Clone the repository
   git clone https://github.com/MrFixit96/terraform-best-practices-mcp.git
   cd terraform-best-practices-mcp
   
   # Build the binary
   go build -o terraform-mcp-server ./cmd/terraform-mcp-server
   ```

3. **Run the Server**

   ```bash
   # Start the server on port 8080
   ./terraform-mcp-server -addr :8080
   
   # With custom data directory
   ./terraform-mcp-server -addr :8080 -data-dir ./my-data
   
   # With debug logging
   ./terraform-mcp-server -addr :8080 -log-level debug
   
   # Run as a background service
   nohup ./terraform-mcp-server -addr :8080 > terraform-mcp.log 2>&1 &
   ```

### Docker Installation

```bash
# Build the Docker image
docker build -t terraform-mcp-server .

# Run the Docker container
docker run -p 8080:8080 terraform-mcp-server

# With custom configuration
docker run -p 8080:8080 -e LOG_LEVEL=debug terraform-mcp-server
```

## Usage

### Command-line Options

- `-addr`: Server address (default: `:8080`)
- `-data-dir`: Data directory for documentation and patterns (default: `./data`)
- `-log-level`: Log level (`debug`, `info`, `error`) (default: `info`)
- `-update-interval`: Update interval for documentation (default: `24h`)
- `-authority-sources`: Comma-separated list of authority sources for Terraform documentation (default: built-in list)

### Integration with AI Assistants

#### 1. Claude Desktop Integration

Claude Desktop supports MCP servers through its configuration file:

1. Start the server: `./terraform-mcp-server -addr :8080`
2. Configure Claude Desktop:
   - Open Settings → Developer → Edit Config
   - Add the following to your configuration:

```json
{
  "mcpServers": {
    "terraform-best-practices": {
      "command": "/path/to/terraform-mcp-server",
      "args": ["-addr", ":8080", "-log-level", "info"]
    }
  }
}
```

#### 2. VS Code Integration

Visual Studio Code supports MCP servers through GitHub Copilot's Agent mode:

1. Create or edit `.vscode/mcp.json` in your workspace:

```json
{
  "servers": {
    "terraform-best-practices": {
      "type": "stdio",
      "command": "/path/to/terraform-mcp-server",
      "args": ["-log-level", "info"]
    }
  }
}
```

## MCP Tools Provided

The server provides these tools to AI assistants:

### 1. GetBestPractices

```json
{
  "topic": "module",
  "category": "structure",
  "provider": "aws",
  "keywords": ["organization", "structure"]
}
```

### 2. GetModuleStructure

```json
{
  "type": "basic",
  "provider": "aws"
}
```

### 3. GetPatternTemplate

```json
{
  "id": "aws-vpc-basic",
  "category": "networking",
  "provider": "aws",
  "complexity": "basic",
  "tags": ["vpc", "networking"],
  "query": "vpc"
}
```

### 4. ValidateConfiguration

```json
{
  "files": {
    "main.tf": "resource \"aws_instance\" \"example\" { ... }",
    "variables.tf": "variable \"name\" { ... }"
  }
}
```

### 5. SuggestImprovements

```json
{
  "files": {
    "main.tf": "resource \"aws_instance\" \"example\" { ... }",
    "variables.tf": "variable \"name\" { ... }"
  }
}
```

## Development

### Project Structure

```
terraform-mcp-server/
├── cmd/
│   └── terraform-mcp-server/ # Main application entry point
├── pkg/
│   ├── hashicorp/           # HashiCorp-specific components
│   │   ├── tfdocs/          # Documentation and validation
│   │   │   ├── indexer.go   # Documentation indexer
│   │   │   ├── patterns.go  # Code pattern templates
│   │   │   ├── validation.go # Validation engine
│   │   │   └── resource_provider.go # Resource provider
│   │   ├── server.go        # Server implementation
│   │   └── tools.go         # MCP tool implementations
│   └── mcp/                 # MCP protocol components
├── tests/                   # Test suite
└── data/                    # Data files
```

### Adding New Components

#### 1. Adding a New Best Practice

Edit the `addDefaultBestPractices` method in `pkg/hashicorp/tfdocs/indexer.go`.

#### 2. Adding a New Pattern Template

Edit the `initializeDefaultPatterns` method in `pkg/hashicorp/tfdocs/patterns.go`.

#### 3. Adding a New Validator

Create a new validator that implements the `Validator` interface in `pkg/hashicorp/tfdocs/validation.go`.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
</content>
