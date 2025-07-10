.PHONY: help build test clean install dev up down logs

# Default target
help:
	@echo "Relayooor Monorepo Commands:"
	@echo "  make build       - Build all projects"
	@echo "  make test        - Run all tests"
	@echo "  make clean       - Clean all build artifacts"
	@echo "  make install     - Install dependencies for all projects"
	@echo "  make dev         - Start development environment"
	@echo "  make up          - Start all services with docker-compose"
	@echo "  make down        - Stop all services"
	@echo "  make logs        - Show logs from all services"

# Build all projects
build: build-hermes build-relayer build-middleware build-monitoring build-webapp

build-hermes:
	@echo "Building Hermes..."
	cd hermes && cargo build --release

build-relayer:
	@echo "Building Go Relayer..."
	cd relayer && go build -o build/rly main.go

build-middleware:
	@echo "Building Relayer Middleware..."
	cd relayer-middleware && go build -o build/relayer-middleware api/cmd/server/main.go

build-monitoring:
	@echo "Building Chainpulse..."
	cd monitoring/chainpulse && cargo build --release

build-webapp:
	@echo "Building Webapp..."
	cd webapp/web && npm run build

# Run tests
test: test-hermes test-relayer test-middleware test-monitoring test-webapp

test-hermes:
	@echo "Testing Hermes..."
	cd hermes && cargo test

test-relayer:
	@echo "Testing Go Relayer..."
	cd relayer && go test ./...

test-middleware:
	@echo "Testing Relayer Middleware..."
	cd relayer-middleware && go test ./...

test-monitoring:
	@echo "Testing Chainpulse..."
	cd monitoring/chainpulse && cargo test

test-webapp:
	@echo "Testing Webapp..."
	cd webapp/web && npm test

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf hermes/target
	rm -rf monitoring/chainpulse/target
	rm -rf relayer/build
	rm -rf relayer-middleware/build
	rm -rf webapp/web/dist

# Install dependencies
install: install-webapp

install-webapp:
	@echo "Installing webapp dependencies..."
	cd webapp/web && npm install

# Development environment
dev:
	@echo "Starting development environment..."
	docker-compose -f docker-compose.yml up -d

# Docker compose commands
up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

# Individual service commands
hermes-start:
	cd hermes && cargo run -- start

relayer-start:
	cd relayer && ./build/rly start

middleware-start:
	cd relayer-middleware && go run api/cmd/server/main.go

monitoring-start:
	cd monitoring/chainpulse && cargo run

webapp-dev:
	cd webapp/web && npm run dev