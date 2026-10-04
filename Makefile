.DEFAULT_GOAL := help
COMPOSE ?= docker compose

.PHONY: help up down logs lint lint-fix test migrate migrate-status
help:
	@echo "make migrate   Apply SQL migrations"
	@echo "make migrate-status  Show migration status"
	@echo "make up        Build and start frontend + backend + notifications + Kafka + PostgreSQL"
	@echo "make down      Stop containers (preserve volumes)"
	@echo "make logs      Follow service logs"
	@echo "make lint      Check Go and frontend using Docker"
	@echo "make lint-fix  Format and fix Go and frontend using Docker"
	@echo "make test      Run Go tests in Docker"

up:
	$(COMPOSE) up --build --detach --wait --wait-timeout 180 backend frontend notifications kafka postgres

down:
	$(COMPOSE) down --remove-orphans

logs:
	$(COMPOSE) logs --follow backend frontend notifications kafka postgres

lint:
	$(COMPOSE) run --rm lint-go run ./...
	$(COMPOSE) run --rm lint-frontend 'npm ci && npm run lint'

lint-fix:
	$(COMPOSE) run --rm lint-go fmt
	$(COMPOSE) run --rm lint-go run --fix ./...
	$(COMPOSE) run --rm lint-frontend 'npm ci && npm run lint-fix'

test:
	$(COMPOSE) run --rm test-go

migrate:
	$(COMPOSE) run --build --rm migrate up

migrate-status:
	$(COMPOSE) run --build --rm migrate status
