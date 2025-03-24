FROM golang:1.24 AS builder
WORKDIR /tests
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .

FROM builder AS tester
ENTRYPOINT [ "go", "test", "-race", "-v", "./..." ]
