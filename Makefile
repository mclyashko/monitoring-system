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

lint:
	make -C services lint