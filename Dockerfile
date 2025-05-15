FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o terraform-mcp-server ./cmd/terraform-mcp-server

# Use a small alpine image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder image
COPY --from=builder /app/terraform-mcp-server .

# Create data directory
RUN mkdir -p /app/data

# Set environment variables
ENV ADDR=:8080
ENV DATA_DIR=/app/data
ENV LOG_LEVEL=info

# Expose the port
EXPOSE 8080

# Run the application
ENTRYPOINT ["./terraform-mcp-server", "-addr", "${ADDR}", "-data-dir", "${DATA_DIR}", "-log-level", "${LOG_LEVEL}"]
</content>
