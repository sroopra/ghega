# Ghega — Multi-stage Dockerfile
# https://github.com/ghega/ghega
#
# Build constraints:
#   - Final image must NOT contain Java, JVM, JRE, JDK, Node.js, npm, yarn, or pnpm.
#   - Minimal attack surface: distroless static final stage.

# ------------------------------------------------------------------------------
# Stage 1: Builder
# ------------------------------------------------------------------------------
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Copy module files first for layer caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/ghega ./cmd/ghega

# ------------------------------------------------------------------------------
# Stage 2: Final (distroless static)
# ------------------------------------------------------------------------------
FROM gcr.io/distroless/static:nonroot

# OCI labels — Ghega branding
LABEL org.opencontainers.image.title="Ghega"
LABEL org.opencontainers.image.description="Open-source healthcare integration engine"
LABEL org.opencontainers.image.source="https://github.com/ghega/ghega"
LABEL org.opencontainers.image.vendor="Ghega"
LABEL org.opencontainers.image.licenses="Apache-2.0"

# Copy binary and CA certificates from builder
COPY --from=builder /bin/ghega /bin/ghega
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/bin/ghega"]
CMD ["serve"]
