# =========================================================
# Build stage
# =========================================================
FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    libwebp-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build \
    -a \
    -trimpath \
    -ldflags="-s -w" \
    -o server \
    ./cmd/api/main.go

# =========================================================
# Runtime stage
# =========================================================
FROM alpine:3.22

RUN apk add --no-cache \
    ca-certificates \
    libwebp \
    tzdata \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/server /app/server

RUN mkdir -p /app/uploads \
    && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8089

CMD ["./server"]