GO ?= go
CONFIG ?= config/config.yaml

.PHONY: build run test cover vet fmt tidy migrate migration-up up down

build: ## Build both binaries into ./bin
	$(GO) build -o bin/core ./cmd/core
	$(GO) build -o bin/migration ./cmd/migration

run: ## Run the gRPC service
	$(GO) run ./cmd/core -config $(CONFIG)

test: ## Run all tests
	$(GO) test ./...

cover: ## Run tests with a coverage report
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

vet: ## Run go vet
	$(GO) vet ./...

fmt: ## Format the code
	$(GO) fmt ./...

tidy: ## Tidy go.mod/go.sum
	$(GO) mod tidy

migrate: ## Apply migrations
	$(GO) run ./cmd/migration -config $(CONFIG)

migration-up: ## Create a migration, e.g. make migration-up name=add_products
	$(GO) run ./cmd/migration -create "$(name)"

up: ## Start local postgres and redis
	docker compose up -d

down: ## Stop local infrastructure
	docker compose down
