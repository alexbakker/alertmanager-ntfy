FROM alpine:latest

# Add non-root user with configurable UID/GID
ENV USER_ID=1000
ENV GROUP_ID=1000

WORKDIR /app
# Build stage
FROM golang:alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o alertmanager-ntfy ./cmd/alertmanager-ntfy

# Final stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/alertmanager-ntfy .

# Create config directory and set up user
RUN mkdir -p /config && \
    addgroup -g ${GROUP_ID} appuser && \
    adduser -D -u ${USER_ID} -G appuser appuser && \
    chown -R appuser:appuser /app && \
    chown -R appuser:appuser /config

# Expose port 8111
EXPOSE 8111

# Set default config path - can be overridden
ENV CONFIG_FILES="/config/config.yml"

# Switch to non-root user
USER appuser:appuser

# Use exec form of ENTRYPOINT with environment variable expansion
ENTRYPOINT ["/app/alertmanager-ntfy", "--configs", "${CONFIG_FILES}"]