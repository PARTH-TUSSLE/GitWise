# syntax=docker/dockerfile:1
# Multi-stage production build for GitWise Go Backend

# Stage 1: Builder
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Download dependencies first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY . .

# Compile static binary with stripped debug symbols
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /bin/gitwise-server cmd/server/main.go

# Stage 2: Minimal runtime image
FROM alpine:3.20 AS runner

RUN apk add --no-cache ca-certificates tzdata git

# Create non-root service user
RUN addgroup -S gitwise && adduser -S -G gitwise -u 10001 gitwise

WORKDIR /home/gitwise

# Copy binary from builder
COPY --from=builder /bin/gitwise-server /usr/local/bin/gitwise-server

# Set ownership
USER gitwise:gitwise

# Expose default HTTP port
EXPOSE 8080

# Container healthcheck
HEALTHCHECK --interval=15s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

ENTRYPOINT ["gitwise-server"]
