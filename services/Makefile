lint: golint

golint:
	@for dir in $(shell find . -name "go.mod" -exec dirname {} \;); do \
		echo "Running golangci-lint in $$dir"; \
		(cd $$dir && golangci-lint run -E goimports,gocritic -v ./...); \
	done
