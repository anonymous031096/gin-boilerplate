.PHONY: run build test migrate-up migrate-down swag docker

DATABASE_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

include .env
export

run:
	air

build:
	go build -o bin/api cmd/api/main.go

test:
	go test ./...

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

swag:
	swag init -g cmd/api/main.go -o docs --parseDependency

docker:
	docker compose up -d
