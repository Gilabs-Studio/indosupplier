# Build stage
FROM golang:1.25.4-alpine AS builder

# Install build dependencies including libwebp for CGO
RUN apk add --no-cache git gcc musl-dev libwebp-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for libwebp support
RUN CGO_ENABLED=1 GOOS=linux go build -a -o server ./cmd/api/main.go

# Build seed tool (for manual seeding modes)
RUN CGO_ENABLED=1 GOOS=linux go build -a -o seed ./cmd/tools/seed/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies for libwebp and timezones
RUN apk --no-cache add ca-certificates libwebp tzdata

WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/server .
COPY --from=builder /app/seed .

# Create uploads directory
RUN mkdir -p /app/uploads && chmod 755 /app/uploads

# Expose port
EXPOSE 8087

# Run the binary
CMD ["./server"]

