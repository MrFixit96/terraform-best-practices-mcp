# Terraform Best Practices MCP Server

A Model Context Protocol (MCP) server that exposes Terraform best practices, module structure recommendations, code pattern templates, and validation tools to AI assistants.

## Features

- **Best Practices Documentation**: Exposes HashiCorp's Terraform best practices documentation through structured resources
- **Module Structure Guidelines**: Provides standardized module structure recommendations for different use cases
- **Pattern Library**: Offers reusable code pattern templates that follow best practices
- **Validation Engine**: Validates Terraform configurations against best practices and provides improvement suggestions

## Architecture

The server follows a modular architecture with clear separation of concerns:

```
terraform-mcp-server/
├── cmd/
│   └── terraform-mcp-server/ # Main application
├── pkg/
│   ├── hashicorp/            # HashiCorp-specific components
│   │   ├── tfdocs/           # Documentation and validation
│   │   │   ├── indexer.go    # Documentation indexer
│   │   │   ├── resources.go  # Resource definitions
│   │   │   ├── patterns.go   # Code pattern templates
│   │   │   └── validation.go # Validation engine
│   │   ├── server.go         # Server implementation
│   │   └── tools.go          # MCP tool implementations
│   └── mcp/                  # MCP protocol components
│       ├── protocol.go       # Protocol definitions
│       ├── server.go         # Generic MCP server
│       └── tool.go           # Tool interface definitions
```

## Building the Server

To build the server, you need Go 1.17 or later:

```bash
# Clone the repository
git clone https://github.com/yourusername/terraform-mcp-server.git
cd terraform-mcp-server

# Build the binary
go build -o terraform-mcp-server ./cmd/terraform-mcp-server
```

## Running the Server

To run the server:

```bash
# Run with default settings
./terraform-mcp-server

# Run with custom settings
./terraform-mcp-server -addr :8081 -data-dir ./custom-data -log-level debug
```

### Command-line Options

- `-addr`: Server address (default: `:8080`)
- `-data-dir`: Data directory for documentation and patterns (default: `./data`)
- `-log-level`: Log level (`debug`, `info`, `error`) (default: `info`)
- `-update-interval`: Update interval for documentation (default: `24h`)

## MCP Protocol

The server implements the [Model Context Protocol (MCP)](https://github.com/mcp-core/mcp) to provide a standardized interface for AI assistants. The following tools are available:

### GetBestPractices

Retrieves best practice documentation for Terraform.

### GetModuleStructure

Provides standardized module structure recommendations.

### GetPatternTemplate

Retrieves code pattern templates for common infrastructure scenarios.

### ValidateConfiguration

Validates Terraform configurations against best practices.

## License

This project is licensed under the MIT License - see the LICENSE file for details.