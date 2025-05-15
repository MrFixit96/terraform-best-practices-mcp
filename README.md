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
git clone https://github.com/yourusername/terraform-best-practices-mcp.git
cd terraform-best-practices-mcp

# Build the binary
go build -o terraform-mcp-server ./cmd/terraform-mcp-server
```

## Usage

### 1. Claude Desktop Integration

Claude Desktop supports MCP servers through its configuration file, making it easy to integrate Terraform best practices into your AI workflows:

```bash
# Start the server
./terraform-mcp-server -addr :8080
```

Then, configure Claude Desktop:

1. Open Claude Desktop on your machine
2. Click on Claude menu and select "Settings..."
3. Navigate to "Developer" in the left sidebar
4. Click "Edit Config"
5. Add the following to your configuration:

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

6. Save and restart Claude Desktop
7. You'll see a tool icon in the input area when the server is active
8. Click the tool icon to access Terraform best practices tools

### 2. VS Code Integration

Visual Studio Code now supports MCP servers through GitHub Copilot's Agent mode:

1. Install VS Code with GitHub Copilot extension
2. Create or edit `.vscode/mcp.json` in your workspace:

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

3. Enable agent mode in GitHub Copilot
4. Access Terraform best practices tools in the Copilot chat by clicking the tools icon or typing "#"
5. Use prompts like "Validate my Terraform module" or "Show best practices for AWS VPC modules"

### 3. Cursor IDE Integration

Cursor IDE, a popular AI-enhanced code editor, also supports MCP integration:

1. Install Cursor IDE from [cursor.sh](https://cursor.sh)
2. Start the Terraform MCP server
3. In Cursor, go to Settings → AI → MCP Servers
4. Add a new server with:
   - Name: Terraform Best Practices
   - Command: /path/to/terraform-mcp-server
   - Arguments: -log-level info
5. Save settings and restart Cursor
6. Access Terraform tools in the AI chat panel

## MCP Tools Provided

The Terraform Best Practices MCP server provides these tools:

- **GetBestPractices**: Retrieves best practice documentation for Terraform
- **GetModuleStructure**: Provides standardized module structure recommendations
- **GetPatternTemplate**: Retrieves code pattern templates for common infrastructure scenarios
- **ValidateConfiguration**: Validates Terraform configurations against best practices

## Example Prompts

Here are some example prompts to use with the MCP server:

- "Show me best practices for structuring Terraform modules"
- "What's the recommended way to handle variables for AWS EC2 modules?"
- "Give me a template for a VPC module"
- "Validate this terraform code [paste code]"
- "What security best practices should I follow in my Terraform configurations?"

## Command-line Options

- `-addr`: Server address (default: `:8080`)
- `-data-dir`: Data directory for documentation and patterns (default: `./data`)
- `-log-level`: Log level (`debug`, `info`, `error`) (default: `info`)
- `-update-interval`: Update interval for documentation (default: `24h`)

## Docker Support

You can also run the server using Docker:

```bash
# Build the Docker image
docker build -t terraform-mcp-server .

# Run the Docker container
docker run -p 8080:8080 terraform-mcp-server
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.