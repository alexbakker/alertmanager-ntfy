FROM golang:latest as builder
WORKDIR /build
COPY ./ ./
RUN go mod download
WORKDIR /build/cmd/alertmanager-ntfy
RUN CGO_ENABLED=0 GOOS=linux go build

FROM scratch
WORKDIR /app
COPY --from=builder /build/cmd/alertmanager-ntfy/alertmanager-ntfy ./alertmanager-ntfy
EXPOSE 8000
ENTRYPOINT ["./alertmanager-ntfy"]
