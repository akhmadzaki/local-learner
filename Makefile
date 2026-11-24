.PHONY: build run test clean docker-build docker-up lint swagger swagger-fmt
.PHONY: migrate-up migrate-down migrate-to migrate-drop migrate-version migrate-create

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags="-X github.com/akhmadzaki/go-backend/config.Version=$(VERSION) -s -w" -o go-backend .

run:
	go run main.go

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

# ── Swagger / OpenAPI docs ─────────────────────────────────────────────────────

# Regenerate the OpenAPI documentation from controller annotations.
# The swag CLI must be installed first: go install github.com/swaggo/swag/cmd/swag@latest
swagger:
	swag init -g main.go -o docs --parseDependency --parseInternal

# Format/validate the generated OpenAPI spec.
swagger-fmt:
	swag fmt -g main.go

# ── Database migrations (golang-migrate) ───────────────────────────────────────

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