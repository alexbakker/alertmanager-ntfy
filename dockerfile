# Build stage
# ARG GO_VERSION
# FROM golang:$GO_VERSION-alpine AS builder
FROM golang:alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o alertmanager-ntfy ./cmd/alertmanager-ntfy

# Final stage
FROM alpine:latest

# Re-declare build args in final stage with defaults
ARG USER_ID=1000
ARG GROUP_ID=1000

WORKDIR /app
COPY --from=builder /build/alertmanager-ntfy .

# Create config directory and set up user
RUN mkdir -p /config && \
    addgroup -g "${GROUP_ID}" appuser && \
    adduser -D -u "${USER_ID}" -G appuser -s /sbin/nologin appuser && \
    chown -R appuser:appuser /app && \
    chown -R appuser:appuser /config

# Expose port 8000
EXPOSE 8000

# Set default config path - can be overridden
ENV CONFIG_FILES="/config/config.yml"

# Switch to non-root user
USER appuser:appuser

# Use shell form to allow environment variable expansion
ENTRYPOINT /app/alertmanager-ntfy --configs $CONFIG_FILES