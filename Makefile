.DEFAULT_GOAL := help
COMPOSE ?= docker compose
SQLC_IMAGE := sqlc/sqlc:1.31.1

.PHONY: generate check-generated init help up down logs lint lint-fix test migrate migrate-status
help:
	@echo "make generate  Generate Go queries from SQL"
	@echo "make check-generated  Check generated queries are up to date"
	@echo "make init      Generate missing local secrets in .env"
	@echo "make migrate   Apply SQL migrations"
	@echo "make migrate-status  Show migration status"
	@echo "make up        Build and start frontend + API + worker + notifications + Kafka + PostgreSQL"
	@echo "make down      Stop containers (preserve volumes)"
	@echo "make logs      Follow service logs"
	@echo "make lint      Check Go and frontend using Docker"
	@echo "make lint-fix  Format and fix Go and frontend using Docker"
	@echo "make test      Run Go and frontend tests in Docker"

init:
	docker run --rm -v "$(CURDIR):/workspace" -w /workspace/backend golang:1.27.1-alpine go run ./cmd/setup

up: init
	$(COMPOSE) up --detach --wait --wait-timeout 180 postgres kafka
	$(MAKE) migrate
	$(COMPOSE) run --rm kafka-init
	$(COMPOSE) up --build --detach --wait --wait-timeout 180 backend frontend notifications worker

down:
	$(COMPOSE) down --remove-orphans

logs:
	$(COMPOSE) logs --follow backend frontend notifications worker kafka postgres

lint:
	$(COMPOSE) run --rm lint-go run ./...
	$(COMPOSE) run --rm lint-frontend 'npm ci && npm run lint'

lint-fix:
	$(COMPOSE) run --rm lint-go fmt
	$(COMPOSE) run --rm lint-go run --fix ./...
	$(COMPOSE) run --rm lint-frontend 'npm ci && npm run lint-fix'

test:
	$(COMPOSE) run --rm test-go
	$(COMPOSE) run --rm lint-frontend 'npm ci && npm test'

migrate:
	$(COMPOSE) run --build --rm migrate up

migrate-status:
	$(COMPOSE) run --build --rm migrate status

generate:
	docker run --rm --user "$$(id -u):$$(id -g)" -v "$(CURDIR)/backend:/src" -w /src $(SQLC_IMAGE) generate

check-generated:
	docker run --rm --user "$$(id -u):$$(id -g)" -v "$(CURDIR)/backend:/src:ro" -w /src $(SQLC_IMAGE) diff
