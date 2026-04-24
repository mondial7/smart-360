# Stage 1: Build the Go application
FROM golang:1.25.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a static binary that doesn't depend on libc
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smart360-server .

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install runtime dependencies
# ca-certificates: required for HTTPS calls (Google OAuth, Gemini API)
# tzdata: proper timezone handling
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/smart360-server .

# Copy .env.example for reference
COPY --from=builder /build/.env.example .

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./smart360-server"]
