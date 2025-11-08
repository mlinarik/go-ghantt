# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Update to latest versions and build
RUN go get -u ./... && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o go-ghant

# Final stage
FROM alpine:latest

# Update package index and upgrade all packages, then install ca-certificates
RUN apk update && \
    apk upgrade && \
    apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/go-ghant ./go-ghant

# Copy static files
COPY static ./static

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/go-ghant"]
