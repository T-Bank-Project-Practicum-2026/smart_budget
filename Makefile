.DEFAULT_GOAL := help
COMPOSE ?= docker compose
SQLC_IMAGE := sqlc/sqlc:1.31.1
SWAGGER_IMAGE := smart-budget-swagger:local
SWAGGER_DOCKER = docker run --rm --user "$$(id -u):$$(id -g)" -e GOCACHE=/tmp/go-build -e GOPATH=/tmp/go

.PHONY: generate generate-sql swagger swagger-tool check-swagger check-generated init help up down logs lint lint-fix test migrate migrate-status
help:
	@echo "make swagger   Обновить api/openapi.yaml из Go-кода"
	@echo "make check-swagger  Проверить актуальность OpenAPI"
	@echo "make generate  Generate SQL queries and OpenAPI documentation"
	@echo "make check-generated  Check generated SQL queries and OpenAPI are up to date"
	@echo "make init      Generate missing local secrets in .env"
	@echo "make migrate   Apply SQL migrations"
	@echo "make migrate-status  Show migration status"
	@echo "make up        Build and start frontend + API + worker + notifications + Redpanda + PostgreSQL"
	@echo "make down      Stop containers (preserve volumes)"
	@echo "make logs      Follow service logs"
	@echo "make lint      Check Go and frontend using Docker"
	@echo "make lint-fix  Format and fix Go and frontend using Docker"
	@echo "make test      Run Go tests in Docker"

init:
	docker run --rm --user "$$(id -u):$$(id -g)" -e GOCACHE=/tmp/go-build -e GOPATH=/tmp/go -v "$(CURDIR):/workspace" -w /workspace/backend golang:1.27-alpine go run ./cmd/setup

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

migrate:
	$(COMPOSE) run --build --rm migrate up

migrate-status:
	$(COMPOSE) run --build --rm migrate status

generate: generate-sql swagger

generate-sql:
	docker run --rm --user "$$(id -u):$$(id -g)" -v "$(CURDIR)/backend:/src" -w /src $(SQLC_IMAGE) generate

check-generated: check-swagger
	docker run --rm --user "$$(id -u):$$(id -g)" -v "$(CURDIR)/backend:/src:ro" -w /src $(SQLC_IMAGE) diff

swagger-tool:
	docker build -f scripts/swagger.Dockerfile -t $(SWAGGER_IMAGE) scripts

swagger: swagger-tool
	$(SWAGGER_DOCKER) -v "$(CURDIR):/workspace" $(SWAGGER_IMAGE) sh ../scripts/openapi.sh generate

check-swagger: swagger-tool
	$(SWAGGER_DOCKER) -v "$(CURDIR):/workspace:ro" $(SWAGGER_IMAGE) sh ../scripts/openapi.sh check
