# Build stage
FROM golang:1.19-alpine AS build

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o terraform-mcp-server ./cmd/terraform-mcp-server

# Final stage
FROM alpine:3.17

WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/terraform-mcp-server /app/terraform-mcp-server

# Create data directories
RUN mkdir -p /app/data/docs
RUN mkdir -p /app/data/patterns

# Copy example data if available
COPY --from=build /app/data/docs /app/data/docs
COPY --from=build /app/data/patterns /app/data/patterns

# Expose the default port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/terraform-mcp-server"]

# Default command line arguments
CMD ["-addr", ":8080", "-data-dir", "/app/data", "-log-level", "info"]