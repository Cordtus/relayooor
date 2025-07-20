# File Mapping Reference

## Configuration Files

### Root Level Configuration
- **`.env`** - Main environment variables (create from .env.example)
- **`.env.example`** - Template for environment configuration
- **`.gitignore`** - Git ignore patterns
- **`Makefile`** - Make commands for development
- **`Makefile.docker`** - Docker-specific make commands

### Docker Configuration
- **`docker-compose.yml`** - Base Docker Compose configuration
- **`docker-compose.full.yml`** - Full stack with monitoring
- **`docker-compose.local.yml`** - Local development overrides
- **`docker-compose.override.yml`** - Local custom overrides
- **`docker-compose.prod.yml`** - Production configuration
- **`docker-compose.webapp.yml`** - Frontend-only configuration
- **`Dockerfile`** - Main application Dockerfile

### Service-Specific Configuration

#### Chainpulse Configuration (`/config/`)
- **`chainpulse.toml`** - Main Chainpulse configuration
- **`chainpulse-mainnet.toml`** - Mainnet chain configuration
- **`chainpulse-testnet.toml`** - Testnet chain configuration
- **`chainpulse-local.toml`** - Local development configuration
- **`chainpulse-proxy.toml`** - Proxy-enabled configuration

#### Hermes Configuration (`/config/hermes/`)
- **`config.toml`** - Main Hermes relayer configuration
- **`test-config.toml`** - Test environment configuration
- **`entrypoint.sh`** - Docker entrypoint script
- **`entrypoint-with-proxy.sh`** - Proxy-enabled entrypoint

#### Monitoring Configuration (`/config/`)
- **`prometheus.yml`** - Prometheus scrape configuration
- **`nodes.toml.example`** - RPC node configuration template

### Frontend Configuration (`/webapp/`)
- **`package.json`** - Node.js dependencies and scripts
- **`tsconfig.json`** - TypeScript configuration
- **`vite.config.ts`** - Vite build configuration
- **`vitest.config.ts`** - Vitest test configuration
- **`tailwind.config.js`** - Tailwind CSS configuration
- **`postcss.config.js`** - PostCSS configuration
- **`.prettierrc`** - Code formatting rules
- **`.eslintrc.js`** - Linting rules

### API Configuration (`/relayer-middleware/api/`)
- **`go.mod`** - Go module dependencies
- **`go.sum`** - Go dependency checksums
- **`.air.toml`** - Hot reload configuration

### Packet Manager Configuration (`/packet-manager/`)
- **`package.json`** - Node.js dependencies
- **`Dockerfile`** - Container configuration

## Test Files

### Frontend Tests (`/webapp/`)
```
src/
├── components/
│   ├── __tests__/          # Component unit tests
│   │   ├── PacketCard.spec.ts
│   │   └── WalletConnect.spec.ts
│   └── clearing/
│       └── __tests__/      # Clearing component tests
│           ├── ClearingWizard.spec.ts
│           └── PaymentStep.spec.ts
├── services/
│   └── __tests__/          # Service tests
│       ├── api.spec.ts
│       └── wallet.spec.ts
└── stores/
    └── __tests__/          # Store tests
        ├── packets.spec.ts
        └── auth.spec.ts
```

### API Tests (`/relayer-middleware/api/`)
```
pkg/
├── clearing/
│   ├── service_test.go     # Clearing service tests
│   ├── token_test.go       # Token generation tests
│   └── validator_test.go   # Validation tests
├── handlers/
│   ├── auth_test.go        # Auth handler tests
│   ├── clearing_test.go    # Clearing endpoint tests
│   └── packets_test.go     # Packet query tests
├── database/
│   └── queries_test.go     # Database query tests
└── services/
    ├── chainpulse_test.go  # Chainpulse client tests
    └── hermes_test.go      # Hermes client tests
```

### Integration Tests (`/tests/`)
- **`integration/`** - End-to-end integration tests
- **`e2e/`** - Full stack E2E tests

### Simple API Tests (`/api/`)
- **`cmd/server/main_test.go`** - Basic API tests

## Helper Scripts

### Development Scripts (`/scripts/`)

#### Setup and Launch
- **`setup-and-launch.sh`** - Complete setup and start script
  ```bash
  # Builds frontend, creates containers, starts services
  ./scripts/setup-and-launch.sh
  ```

#### Testing Scripts
- **`run-tests.sh`** - Run all test suites
  ```bash
  # Runs frontend, backend, and integration tests
  ./scripts/run-tests.sh
  ```

- **`test-packet-clearing-scenarios.sh`** - Test clearing scenarios
  ```bash
  # Tests various packet clearing scenarios
  ./scripts/test-packet-clearing-scenarios.sh
  ```

- **`test-chainpulse-integration.sh`** - Test Chainpulse integration
  ```bash
  # Verifies Chainpulse is collecting data correctly
  ./scripts/test-chainpulse-integration.sh
  ```

#### Chain Management
- **`check-chain-compatibility.sh`** - Verify chain compatibility
  ```bash
  # Check if a chain is compatible
  ./scripts/check-chain-compatibility.sh cosmoshub-4
  ```

- **`update-api-chains.sh`** - Update chain configurations
  ```bash
  # Updates chain list in API
  ./scripts/update-api-chains.sh
  ```

- **`debug-neutron.sh`** - Debug Neutron-specific issues
  ```bash
  # Helps diagnose Neutron ABCI++ issues
  ./scripts/debug-neutron.sh
  ```

#### Configuration Generation
- **`generate-chainpulse-config.js`** - Generate Chainpulse configs
  ```bash
  # Creates Chainpulse configuration from template
  node scripts/generate-chainpulse-config.js
  ```

### Database Scripts (`/relayer-middleware/api/scripts/`)
- **`create-migration.sh`** - Create new database migration
- **`backup-db.sh`** - Backup database
- **`restore-db.sh`** - Restore database from backup

### Deployment Scripts (`/deploy/`)
- **`deploy-staging.sh`** - Deploy to staging environment
- **`deploy-production.sh`** - Deploy to production
- **`rollback.sh`** - Rollback to previous version

## Utility Files

### Documentation (`/docs/`)
- **`neutron-slinky-issue.md`** - Neutron compatibility documentation
- **`api-spec.yaml`** - OpenAPI specification
- **`architecture.md`** - System architecture documentation

### Monitoring (`/monitoring/`)
- **`grafana/dashboards/`** - Grafana dashboard JSON files
  - `api-metrics.json`
  - `chainpulse-ibc.json`
  - `hermes-relayer.json`
- **`alerts/`** - Prometheus alerting rules
  - `alerts.yml`

### Migration Files (`/relayer-middleware/api/migrations/`)
- **`001_initial_schema.up.sql`** - Initial database schema
- **`001_initial_schema.down.sql`** - Rollback initial schema
- **`002_add_user_tables.up.sql`** - User management tables
- **`003_add_clearing_tables.up.sql`** - Clearing request tables

## Environment-Specific Files

### Development Files
- **`.env.development`** - Development environment variables
- **`docker-compose.override.yml`** - Local Docker overrides
- **`.vscode/`** - VS Code workspace settings
  - `launch.json` - Debug configurations
  - `settings.json` - Workspace settings

### Production Files
- **`.env.production`** - Production environment variables
- **`fly.toml`** - Fly.io deployment configuration
- **`k8s/`** - Kubernetes manifests
  - `deployment.yaml`
  - `service.yaml`
  - `configmap.yaml`

## Key File Locations Reference

### Critical Configuration Files
1. **Main env config**: `/.env`
2. **Docker compose**: `/docker-compose.yml`
3. **Chainpulse config**: `/config/chainpulse.toml`
4. **Hermes config**: `/config/hermes/config.toml`
5. **Frontend config**: `/webapp/.env`

### Main Entry Points
1. **Frontend**: `/webapp/src/main.ts`
2. **API**: `/relayer-middleware/api/cmd/server/main.go`
3. **Simple API**: `/api/cmd/server/main.go`

### Test Entry Points
1. **Frontend tests**: Run `yarn test` in `/webapp`
2. **API tests**: Run `go test ./...` in `/relayer-middleware/api`
3. **Integration tests**: Run `/scripts/run-tests.sh`

### Build Outputs
1. **Frontend build**: `/webapp/dist/`
2. **API binary**: `/relayer-middleware/api/api`
3. **Docker images**: Check with `docker images | grep relayooor`

## File Naming Conventions

### Frontend Files
- Components: PascalCase (e.g., `PacketCard.vue`)
- Composables: camelCase with 'use' prefix (e.g., `usePackets.ts`)
- Stores: camelCase (e.g., `packets.ts`)
- Types: PascalCase with '.types.ts' suffix

### Backend Files
- Go files: snake_case (e.g., `packet_handler.go`)
- Test files: '_test.go' suffix
- SQL migrations: numbered prefix (e.g., `001_initial_schema.up.sql`)

### Configuration Files
- TOML configs: kebab-case (e.g., `chainpulse-mainnet.toml`)
- Docker files: 'docker-compose' prefix
- Environment files: '.env' prefix

## Quick Reference Commands

### Find all config files
```bash
find . -name "*.toml" -o -name "*.json" -o -name "*.yml" -o -name "*.yaml" | grep -E "(config|conf)" | sort
```

### Find all test files
```bash
# Go tests
find . -name "*_test.go" | sort

# JavaScript/TypeScript tests
find . -name "*.spec.ts" -o -name "*.spec.js" -o -name "*.test.ts" -o -name "*.test.js" | sort
```

### Find all scripts
```bash
find . -name "*.sh" -type f | sort
```

### List Docker-related files
```bash
find . -name "Dockerfile*" -o -name "docker-compose*.yml" | sort
```