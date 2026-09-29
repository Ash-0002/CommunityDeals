# Community Platform — Backend Makefile

.PHONY: run build seed test docker-up docker-down migrate-up migrate-down lint tidy

## ── Development ──────────────────────────────────────────────────────────────

# Start the server (expects a running postgres + redis)
run:
	go run ./cmd/server

# Build the binary
build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

# Populate demo data (users, a community, 3 campaigns with joins) — safe to re-run
seed:
	go run ./cmd/seed

# Run all tests
test:
	go test ./... -v -count=1

# Tidy and verify modules
tidy:
	go mod tidy
	go mod verify

## ── Docker ───────────────────────────────────────────────────────────────────

# Start all services (postgres, redis, api)
docker-up:
	docker compose up --build -d

# Stop all services
docker-down:
	docker compose down

# Stop and wipe volumes (clean slate)
docker-reset:
	docker compose down -v

## ── Database Migrations ──────────────────────────────────────────────────────

# Run all pending migrations
# Requires: migrate CLI  →  brew install golang-migrate
# or: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

DB_URL ?= postgres://postgres:postgres@localhost:5432/community_platform?sslmode=disable

migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-status:
	migrate -path ./migrations -database "$(DB_URL)" version

migrate-force:
	migrate -path ./migrations -database "$(DB_URL)" force $(VERSION)

## ── Code Quality ─────────────────────────────────────────────────────────────

lint:
	golangci-lint run ./...
