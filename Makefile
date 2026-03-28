APP_NAME=api
MIGRATIONS_PATH=./migrations
POSTGRES_DSN ?= postgres://app_user:app_pass@localhost:5432/app_db?sslmode=disable
GOPATH_BIN := $(shell go env GOPATH)/bin

.PHONY: run dev test migrate-up migrate-down migrate-down-1 swagger

# Live reload: PATH → $(GOPATH)/bin/air → go run (no install required; first run may download)
dev:
	@set -e; \
	if command -v air >/dev/null 2>&1; then \
		exec air; \
	elif [ -x "$(GOPATH_BIN)/air" ]; then \
		exec "$(GOPATH_BIN)/air"; \
	else \
		exec go run github.com/air-verse/air@latest; \
	fi

run:
	go run ./cmd/api

test:
	go test ./...

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_DSN)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_DSN)" down

migrate-down-1:
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_DSN)" down 1

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs --parseInternal
