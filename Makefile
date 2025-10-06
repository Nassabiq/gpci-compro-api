SHELL := /bin/bash
.DEFAULT_GOAL := help

ifneq (,$(wildcard .env))
include .env
export $(shell sed -n 's/^\([A-Za-z_][A-Za-z0-9_]*\)=.*/\1/p' .env)
endif

APP_BIN_DIR := bin
API_BIN := $(APP_BIN_DIR)/api
WORKER_BIN := $(APP_BIN_DIR)/worker
IMAGE_NAME ?= gpci-api:latest
COMPOSE ?= docker compose
GOOSE := go run github.com/pressly/goose/v3/cmd/goose
MIGRATIONS_DIR ?= migrations
DB_DSN ?= host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)

.PHONY: help deps tidy test run-api run-worker build build-api build-worker clean docker-build compose-up compose-down migrator migrate-up migrate-down migrate-reset migration

help:
	@echo "Make targets:"
	@echo "  deps           Install Go module dependencies."
	@echo "  tidy           Run go mod tidy to clean module file."
	@echo "  test           Run unit tests."
	@echo "  run-api        Start the HTTP API locally."
	@echo "  run-worker     Start the background worker locally."
	@echo "  build          Build API and worker binaries into bin/."
	@echo "  docker-build   Build the Docker image ($(IMAGE_NAME))."
	@echo "  compose-up     Start api + worker services with docker compose."
	@echo "  compose-down   Stop docker compose services."
	@echo "  migrator       Run the one-shot migrator service via docker compose."
	@echo "  migrate-up     Apply all pending database migrations."
	@echo "  migrate-down   Roll back the most recent database migration."
	@echo "  migrate-reset  Reset database schema by rolling back then reapplying migrations."
	@echo "  migration      Create a new goose SQL migration (usage: make migration name=add_users)."

deps:
	go mod download

tidy:
	go mod tidy

test:
	go test ./...

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

build: build-api build-worker

build-api:
	mkdir -p $(APP_BIN_DIR)
	go build -o $(API_BIN) ./cmd/api

build-worker:
	mkdir -p $(APP_BIN_DIR)
	go build -o $(WORKER_BIN) ./cmd/worker

clean:
	rm -rf $(APP_BIN_DIR)

docker-build:
	docker build -t $(IMAGE_NAME) .

compose-up:
	$(COMPOSE) up api worker

compose-down:
	$(COMPOSE) down

migrator:
	$(COMPOSE) --profile migrate run --rm migrator

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down

migrate-reset:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" reset

migration:
ifeq ($(strip $(name)),)
	$(error name is required, e.g. make migration name=add_users_table)
endif
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql
