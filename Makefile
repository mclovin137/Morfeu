.PHONY: help up down gen migrate lint test build run

help:
	@echo "Morfeu makefile targets:"
	@echo "  up       - Start docker-compose (PG + Redis)"
	@echo "  down     - Stop docker-compose"
	@echo "  gen      - Generate sqlc code"
	@echo "  migrate  - Run database migrations"
	@echo "  lint     - Run golangci-lint"
	@echo "  test     - Run tests"
	@echo "  build    - Build the application"
	@echo "  run      - Run the application"

up:
	docker-compose up -d

down:
	docker-compose down

gen:
	sqlc generate

migrate:
	migrate -path ./migrations -database $(DATABASE_URL) up

lint:
	golangci-lint run ./...

test:
	go test -race -cover ./...

build:
	CGO_ENABLED=0 go build -o app ./cmd/app

run: build
	./app
