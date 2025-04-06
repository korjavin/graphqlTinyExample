.PHONY: up down restart logs migrate seed clean build test run-client run-server

# Docker commands
up:
	podman-compose up -d

down:
	podman-compose down

restart: down up

logs:
	podman-compose logs -f

# Database commands
postgres-up:
	podman-compose up -d postgres

migrate: postgres-up
	podman-compose exec postgres psql -U postgres -d graphql_example -f /var/lib/postgresql/data/01_schema.sql
	@echo "Database schema migrated successfully"

seed: migrate
	podman-compose exec postgres psql -U postgres -d graphql_example -f /var/lib/postgresql/data/02_fixtures.sql
	@echo "Test data loaded successfully"

load-migrations: postgres-up
	podman cp migrations/01_schema.sql graphqltinyexample_postgres_1:/var/lib/postgresql/data/
	podman cp migrations/02_fixtures.sql graphqltinyexample_postgres_1:/var/lib/postgresql/data/
	@echo "Migration files copied to PostgreSQL container"

db-setup: load-migrations migrate seed

# Development commands
clean:
	rm -rf bin/
	podman-compose down -v

build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

test:
	go test -v ./...

run-server: build
	./bin/server

run-client: build
	./bin/client

# Using pre-built images from GHCR
up-ghcr:
	podman-compose -f docker-compose.gchr.yaml up -d

down-ghcr:
	podman-compose -f docker-compose.gchr.yaml down