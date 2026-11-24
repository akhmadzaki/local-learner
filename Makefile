.PHONY: build run stop test clean docker-build docker-up lint swagger swagger-fmt
.PHONY: migrate-up migrate-down migrate-to migrate-drop migrate-version migrate-create
.PHONY: pull-models pull-models-gpu
.PHONY: run-gpu

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# ── GPU backend selection ──────────────────────────────────────────────────
# Override OLLAMA_PROFILE to switch between CPU, NVIDIA, or MPS:
#
#   CPU    (default) : make run                    → --profile cpu
#   NVIDIA GPU       : make run-gpu                → --profile nvidia
#   Apple MPS        : OLLAMA_PROFILE=mps make run → --profile mps
#
OLLAMA_PROFILE ?= cpu
OLLAMA_SERVICE  = ollama$(if $(filter nvidia,$(OLLAMA_PROFILE)),-gpu)$(if $(filter mps,$(OLLAMA_PROFILE)),-mps)

# ── Production build / run ──────────────────────────────────────────────────

build:
	go build -ldflags="-X github.com/akhmadzaki/go-backend/config.Version=$(VERSION) -s -w" -o go-backend .

run:
	docker compose --profile $(OLLAMA_PROFILE) up --build -d postgres $(OLLAMA_SERVICE) backend frontend
	@echo ""
	@echo "╔══════════════════════════════════════════════════════════════╗"
	@echo "║  Pulling AI models into the Ollama container…              ║"
	@echo "║  This downloads ~1-2 GB on first run.                      ║"
	@echo "╚══════════════════════════════════════════════════════════════╝"
	@echo ""
	docker compose exec -T $(OLLAMA_SERVICE) ollama pull nomic-embed-text
	docker compose exec -T $(OLLAMA_SERVICE) ollama pull llama3.2:1b-instruct-q2_K
	go run ./cmd/migrate up
	@echo ""
	@echo "╔══════════════════════════════════════════════════════════════╗"
	@echo "║  All services are running!                                 ║"
	@echo "║  Frontend : http://localhost:3000                          ║"
	@echo "║  Backend  : http://localhost:8080                          ║"
	@echo "║  Ollama   : http://localhost:11434                         ║"
	@echo "╚══════════════════════════════════════════════════════════════╝"
	@echo ""

# Production run with NVIDIA GPU acceleration
run-gpu:
	$(MAKE) run OLLAMA_PROFILE=nvidia

stop:
	docker compose down
	@echo ""
	@echo "All services stopped. To also delete data volumes:"
	@echo "  docker compose down -v"
	@echo ""

# Pull models into the running CPU Ollama container
pull-models:
	docker compose exec -T ollama ollama pull nomic-embed-text
	docker compose exec -T ollama ollama pull llama3.2:1b-instruct-q2_K

# Pull models into the running GPU Ollama container
pull-models-gpu:
	docker compose exec -T ollama-gpu ollama pull nomic-embed-text
	docker compose exec -T ollama-gpu ollama pull llama3.2:1b-instruct-q2_K

test:
	go test ./... -v -race -count=1

clean:
	rm -f go-backend
	rm -rf build/

docker-build:
	docker compose build

docker-up:
	docker compose up -d

lint:
	go vet ./...

# ── Swagger / OpenAPI docs ─────────────────────────────────────────────────

# Regenerate the OpenAPI documentation from controller annotations.
# The swag CLI must be installed first: go install github.com/swaggo/swag/cmd/swag@latest
swagger:
	swag init -g main.go -o docs --parseDependency --parseInternal

# Format/validate the generated OpenAPI spec.
swagger-fmt:
	swag fmt -g main.go

# ── Database migrations (golang-migrate) ───────────────────────────────────

# Migrate commands use the embedded cmd/migrate binary, which reads the standard
# DB_* environment variables (same defaults as the server).

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-to:
	@test -n "$(VERSION)" || (echo "Usage: make migrate-to VERSION=<n>" && exit 1)
	go run ./cmd/migrate to $(VERSION)

migrate-drop:
	go run ./cmd/migrate drop

migrate-version:
	go run ./cmd/migrate version

# Create a new migration using the golang-migrate CLI tool.
# Usage: make migrate-create NAME=add_user_preferences
migrate-create:
	@test -n "$(NAME)" || (echo "Usage: make migrate-create NAME=<name>" && exit 1)
	go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1 create -ext sql -dir db/migrations -seq $(NAME)