FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN go build -o test-service-go .

FROM alpine:latest
COPY --from=builder /app/test-service-go /usr/local/bin/test-service-go
COPY config.yaml /etc/test-service-go/config.yaml
EXPOSE 8080
ENTRYPOINT ["test-service-go", "-config", "/etc/test-service-go/config.yaml"]
