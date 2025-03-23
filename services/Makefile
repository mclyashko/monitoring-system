lint: golint

golint:
	@for dir in $(shell find . -name "go.mod" -exec dirname {} \;); do \
		echo "Running golangci-lint in $$dir"; \
		(cd $$dir && golangci-lint run -E goimports,gocritic -v ./...); \
	done

tools:
	go install github.com/yoheimuta/protolint/cmd/protolint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.5