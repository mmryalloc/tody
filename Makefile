include .env  
export

BINARY_DIR  := ./bin
BINARY      := $(BINARY_DIR)/$(APP_NAME)
MIGRATIONS  := ./migrations
DB_URL      ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

.PHONY: build run tidy migrate-up migrate-down migrate-force migrate-create docker-up docker-down docker-build

# Build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BINARY_DIR)
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/api

run:
	go run ./cmd/api

# Quality
tidy:
	go mod tidy
	go mod verify

test:
	go test ./... -v

# Database
migrate-up:
	@migrate -path $(MIGRATIONS) -database "$(DB_URL)" up

migrate-down:
	@migrate -path $(MIGRATIONS) -database "$(DB_URL)" down

migrate-force:	
	@migrate -path $(MIGRATIONS) -database "$(DB_URL)" force $(version)

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS) -seq $(name)

# Docker
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose build
