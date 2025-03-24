FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN go build -o metrics-collector .

FROM alpine:latest
COPY --from=builder /app/metrics-collector /usr/local/bin/metrics-collector
COPY config.yaml /etc/metrics-collector/config.yaml
EXPOSE 80
EXPOSE 8080
ENTRYPOINT ["metrics-collector", "-config", "/etc/metrics-collector/config.yaml"]
