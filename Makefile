# Relayooor Makefile

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: docs
docs: ## Open documentation files for editing
	@echo "Opening documentation files..."
	@${EDITOR:-code} CLAUDE.md PROJECT_STATUS.md

.PHONY: status
status: ## Show project status
	@echo "📊 Project Status"
	@echo "================"
	@echo ""
	@echo "Documentation:"
	@echo "  CLAUDE.md - Last modified: $$(stat -c %y CLAUDE.md 2>/dev/null || stat -f %Sm CLAUDE.md)"
	@echo "  PROJECT_STATUS.md - Last modified: $$(stat -c %y PROJECT_STATUS.md 2>/dev/null || stat -f %Sm PROJECT_STATUS.md)"
	@echo ""
	@echo "Services:"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(relayooor|grafana|prometheus|monitor)" || echo "  No services running"
	@echo ""
	@echo "Recent commits:"
	@git log --oneline -5

.PHONY: check-docs
check-docs: ## Check if docs are up to date
	@echo "📅 Checking documentation freshness..."
	@bash -c 'if [[ $$(git diff --name-only | grep -E "(CLAUDE\.md|PROJECT_STATUS\.md)" | wc -l) -gt 0 ]]; then \
		echo "❌ Documentation has uncommitted changes"; \
	else \
		echo "✅ Documentation is committed"; \
	fi'
	@echo ""
	@echo "Last documentation updates:"
	@git log -1 --pretty=format:"%h %s (%cr)" -- CLAUDE.md PROJECT_STATUS.md

.PHONY: update-context
update-context: ## Update all context documentation
	@echo "📝 Updating context documentation..."
	@echo ""
	@echo "Current tasks from PROJECT_STATUS.md:"
	@grep -E "^- \[ \]" PROJECT_STATUS.md | head -5 || echo "No pending tasks found"

.PHONY: start
start: ## Start all services
	./start.sh

.PHONY: start-monitoring
start-monitoring: ## Start just monitoring stack
	docker-compose -f docker-compose.minimal.yml up -d

.PHONY: logs
logs: ## Show logs from all services
	docker-compose logs -f

.PHONY: stop
stop: ## Stop all services
	docker-compose down
	docker stop $$(docker ps -q --filter name=relayooor) 2>/dev/null || true

.PHONY: clean
clean: stop ## Stop services and clean up volumes
	docker-compose down -v
	rm -rf hermes-home chainpulse-data
	docker rmi $$(docker images -q --filter reference=relayooor*) 2>/dev/null || true

.PHONY: build
build: ## Build all Docker images
	docker-compose build

.PHONY: test-monitor
test-monitor: ## Test monitoring endpoints
	@echo "🔍 Testing monitoring endpoints..."
	@curl -s http://localhost:3002/metrics > /dev/null && echo "✅ IBC Monitor: OK" || echo "❌ IBC Monitor: Failed"
	@curl -s http://localhost:9090/-/healthy > /dev/null && echo "✅ Prometheus: OK" || echo "❌ Prometheus: Failed"
	@curl -s http://localhost:3000/api/health > /dev/null && echo "✅ Grafana: OK" || echo "❌ Grafana: Failed"

# Development workflow commands
.PHONY: dev-backend
dev-backend: ## Start backend development
	@echo "Starting backend development..."
	@echo "Remember to update CLAUDE.md with any architecture changes!"
	cd api && go run cmd/server/main.go

.PHONY: dev-frontend
dev-frontend: ## Start frontend development
	@echo "Starting frontend development..."
	@echo "Remember to update PROJECT_STATUS.md when completing features!"
	cd webapp && yarn dev

.PHONY: commit
commit: check-docs ## Commit with documentation check
	@git add -A
	@git commit

# Testing commands
.PHONY: test
test: ## Run all tests
	@./run-tests.sh

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@cd api && go test ./... -v
	@cd relayer-middleware/api && go test ./... -v
	@cd webapp && npm test

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@cd tests/integration && npm test

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	@cd tests/integration && npm run test:e2e

.PHONY: test-security
test-security: ## Run security scans
	@echo "Running security scans..."
	@govulncheck ./... || echo "Install govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest"
	@cd webapp && npm audit

.PHONY: test-coverage
test-coverage: ## Generate test coverage reports
	@echo "Generating coverage reports..."
	@cd api && go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
	@cd relayer-middleware/api && go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
	@cd webapp && npm run test:coverage

.PHONY: test-quick
test-quick: ## Quick test (unit tests with race detection)
	@cd api && go test -race -short ./...
	@cd relayer-middleware/api && go test -race -short ./...

.PHONY: test-bench
test-bench: ## Run benchmark tests
	@cd api && go test -bench=. -benchmem ./...
	@cd relayer-middleware/api && go test -bench=. -benchmem ./...

