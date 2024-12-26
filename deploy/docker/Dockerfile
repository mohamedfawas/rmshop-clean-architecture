# Start with golang alpine image for building
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and required dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main ./cmd/api

# Start a new stage from scratch
FROM alpine:latest

WORKDIR /root/

# Copy only necessary files from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/static ./static

# Create necessary directories
RUN mkdir -p ./internal/config

# Expose port
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
