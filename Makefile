COMPOSE_FILE=./deploy/docker-compose.yaml

up:
	docker compose -f $(COMPOSE_FILE) up --build -d

down:
	docker compose -f $(COMPOSE_FILE) down

clean:
	docker compose -f $(COMPOSE_FILE) down -v

run-tests: 
	docker run --rm --network=host tests:latest

test:
	make clean
	make up
	@echo wait cluster to start && sleep 1
	make run-tests
	make clean
	@echo "test finished"

protolint:
	protolint ./proto

lint:
	make protolint
	make -C services lint

protobuf:
	protoc --go_out=./services/metrics-collector/adapters/grpc --go_opt=paths=source_relative \
		--go-grpc_out=./services/metrics-collector/adapters/grpc --go-grpc_opt=paths=source_relative \
		./proto/metrics_collector.proto
	protoc --go_out=./tests/metrics-collector --go_opt=paths=source_relative \
		--go-grpc_out=./tests/metrics-collector --go-grpc_opt=paths=source_relative \
		./proto/metrics_collector.proto

tools:
	go install github.com/yoheimuta/protolint/cmd/protolint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.5
