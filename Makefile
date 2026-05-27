.PHONY: run build tidy docker-up docker-down logs docker-run

IMAGE ?= $(DOCKERHUB_USERNAME)/discord-festival-bot:latest

run:
	go run ./cmd/bot

build:
	go build -o bin/bot ./cmd/bot

tidy:
	go mod tidy

# Local dev only — builds from source
docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

logs:
	docker compose logs -f bot

# Pull from Docker Hub and run (mirrors what the NAS does)
docker-run:
	docker run -d \
		--name festival-bot \
		--restart unless-stopped \
		--env-file .env \
		$(IMAGE)
