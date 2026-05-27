.PHONY: run build docker-up docker-down logs tidy

run:
	go run ./cmd/bot

build:
	go build -o bin/bot ./cmd/bot

tidy:
	go mod tidy

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

logs:
	docker compose logs -f bot