# syntax=docker/dockerfile:1.7

# ============================================================
# Stage 1: Build the jane-ai binary
# ============================================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source and build
COPY Makefile ./
COPY cmd ./cmd
COPY pkg ./pkg
COPY workspace ./workspace
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make build

# ============================================================
# Stage 2: Minimal runtime image
# ============================================================
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata curl

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:18790/health || exit 1

# Copy binary
COPY --from=builder /src/build/jane-ai /usr/local/bin/jane-ai

# Create non-root user and group
RUN addgroup -g 1000 jane-ai && \
    adduser -D -u 1000 -G jane-ai jane-ai

# Switch to non-root user
USER jane-ai

# Run onboard to create initial directories and config
RUN /usr/local/bin/jane-ai onboard

ENTRYPOINT ["jane-ai"]
CMD ["gateway"]
