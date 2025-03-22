COMPOSE_FILE=./deploy/docker-compose.yaml

up:
	docker compose -f $(COMPOSE_FILE) up --build -d

down:
	docker compose -f $(COMPOSE_FILE) down

clean:
	docker compose -f $(COMPOSE_FILE) down -v

lint:
	make -C services lint