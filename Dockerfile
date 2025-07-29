FROM golang:1.24.4-alpine

# Set working directory
WORKDIR /app

# Install git, curl and ca-certificates (needed for some Go modules and IP checking)
RUN apk add --no-cache git ca-certificates curl

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy email templates
COPY files ./files

# Expose port 8080
EXPOSE 8080

# Show server IP then run the application
CMD sh -c "echo '=== Server IP Information ===' && curl -s ifconfig.co/json && echo '\n=== Starting MCP Server ===' && go run main.go helper.go"