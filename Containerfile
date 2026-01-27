ARG GO_VERSION
FROM --platform=$BUILDPLATFORM golang:$GO_VERSION-alpine AS builder
ARG TARGETOS TARGETARCH

RUN apk add -U --no-cache ca-certificates tzdata

WORKDIR /build
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build github.com/alexbakker/alertmanager-ntfy/cmd/alertmanager-ntfy

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/ /usr/share/zoneinfo/
COPY --from=builder /build/alertmanager-ntfy /bin/

EXPOSE 8000
ENTRYPOINT ["/bin/alertmanager-ntfy"]

LABEL org.opencontainers.image.source=https://github.com/alexbakker/alertmanager-ntfy
LABEL org.opencontainers.image.description="Container image for alertmanager-ntfy"
LABEL org.opencontainers.image.licenses=GPL-3.0-only
