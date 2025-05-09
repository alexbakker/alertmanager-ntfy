FROM golang:1.24.3-alpine3.21 AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /src

COPY go.mod go.sum ./
RUN apk add --no-cache git && go mod download

COPY . .
RUN go build -o alertmanager-ntfy ./cmd/alertmanager-ntfy

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /src/alertmanager-ntfy /usr/local/bin/alertmanager-ntfy
COPY --from=builder /src/config.example.yml /etc/alertmanager-ntfy/config.yml

EXPOSE 8000

ENTRYPOINT ["/usr/local/bin/alertmanager-ntfy"]
CMD ["--configs", "/etc/alertmanager-ntfy/config.yml"]
