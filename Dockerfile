FROM golang:1.24.4-alpine

# Set working directory
WORKDIR /app

# Install git (needed for some Go modules)
RUN apk add --no-cache git ca-certificates

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

# Run the application directly with go run
CMD ["go", "run", "main.go", "helper.go"]