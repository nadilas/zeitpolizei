# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled for SQLite
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o zeitpolizei ./cmd/zeitpolizei

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates sqlite-libs tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/zeitpolizei /app/zeitpolizei

# Create data directory
RUN mkdir -p /data

# Default configuration
ENV ZEITPOLIZEI_CONFIG=/app/config.yaml

# Expose default port
EXPOSE 8765

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8765/health || exit 1

# Run as non-root user
RUN adduser -D -u 1000 zeitpolizei
RUN chown -R zeitpolizei:zeitpolizei /app /data
USER zeitpolizei

VOLUME ["/data"]

ENTRYPOINT ["/app/zeitpolizei"]
CMD ["-config", "/app/config.yaml"]
