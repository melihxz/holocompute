# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git build-base

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o holo ./cmd/holo

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh holouser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/holo .

# Change ownership to non-root user
RUN chown -R holouser:holouser /app

# Switch to non-root user
USER holouser

# Expose default port
EXPOSE 8443

# Command to run the application
ENTRYPOINT ["./holo"]
CMD ["agent"]