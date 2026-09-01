# ── Stage 1: Build ─────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server ./cmd/server

# ── Stage 2: Runtime ───────────────────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

# Add CA certificates (needed for HTTPS calls to SMS providers, etc.)
RUN apk add --no-cache ca-certificates tzdata

# Copy compiled binary only
COPY --from=builder /app/server /app/server

# Copy migrations (needed at runtime if using migrate CLI inside container)
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

# Run as non-root for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

ENTRYPOINT ["/app/server"]
