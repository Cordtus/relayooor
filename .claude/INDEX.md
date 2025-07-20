# Documentation Index

## Overview
This index provides a complete reference to all internal documentation for the Relayooor project. These documents are maintained in the `.claude` directory and provide comprehensive guidance for development, deployment, and operations.

## Core Documentation

### 📋 [PROJECT_BLUEPRINT.md](./PROJECT_BLUEPRINT.md)
**Purpose**: High-level project overview and architecture
**Contents**:
- System architecture diagrams
- Component interactions
- Security architecture
- Deployment patterns
- Success metrics
- Risk mitigation strategies

**When to use**: 
- Understanding overall system design
- Planning new features
- Architectural decisions
- Onboarding new team members

### 🏗️ [BUILD_AND_DEPLOYMENT.md](./BUILD_AND_DEPLOYMENT.md)
**Purpose**: Complete build and deployment procedures
**Contents**:
- Build process for all environments
- Deployment configurations
- Environment setup
- Database management
- SSL/TLS configuration
- Monitoring setup
- Rollback procedures

**When to use**:
- Setting up development environment
- Deploying to staging/production
- Configuring CI/CD pipelines
- Troubleshooting deployment issues

### 🗺️ [FILE_MAPPING.md](./FILE_MAPPING.md)
**Purpose**: Complete reference of all project files
**Contents**:
- Configuration file locations
- Test file organization
- Helper script descriptions
- Key file paths
- Naming conventions
- Quick reference commands

**When to use**:
- Finding specific files
- Understanding project structure
- Locating configuration
- Finding test files

### 🔌 [API_INTERFACES.md](./API_INTERFACES.md)
**Purpose**: External API documentation
**Contents**:
- Hermes REST API endpoints
- Chainpulse API endpoints
- WebSocket interfaces
- Authentication details
- Error formats
- Integration examples

**When to use**:
- Integrating with Hermes
- Using Chainpulse data
- Implementing WebSocket features
- Debugging API issues

### 🔧 [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)
**Purpose**: Common issues and solutions
**Contents**:
- Development issues
- Service-specific problems
- Production issues
- Debugging tools
- Recovery procedures
- Quick reference

**When to use**:
- Encountering errors
- Service failures
- Performance issues
- System recovery

## Module Blueprints

### 🎨 [implementations/FRONTEND_MODULE.md](./implementations/FRONTEND_MODULE.md)
**Purpose**: Frontend architecture and implementation
**Contents**:
- Vue 3 architecture
- Component structure
- State management (Pinia)
- API integration patterns
- Testing strategies
- Build configuration

**When to use**:
- Frontend development
- Adding new UI features
- Component testing
- Performance optimization

### ⚙️ [implementations/API_MODULE.md](./implementations/API_MODULE.md)
**Purpose**: Backend API architecture
**Contents**:
- Go service structure
- Authentication system
- Clearing service logic
- Database patterns
- WebSocket implementation
- Testing approaches

**When to use**:
- Backend development
- Adding API endpoints
- Database changes
- Security implementation

### 📊 [implementations/CHAINPULSE_MODULE.md](./implementations/CHAINPULSE_MODULE.md)
**Purpose**: Chainpulse monitoring service
**Contents**:
- Rust architecture
- IBC packet monitoring
- Custom modifications
- Database schema
- Performance optimization
- Known limitations

**When to use**:
- Understanding monitoring
- Debugging chain issues
- Adding chain support
- Performance tuning

### 🔄 [implementations/HERMES_MODULE.md](./implementations/HERMES_MODULE.md)
**Purpose**: Hermes relayer integration
**Contents**:
- Configuration details
- API endpoints
- Docker integration
- Operational procedures
- Error handling
- Security setup

**When to use**:
- Configuring Hermes
- Clearing packets
- Adding new chains
- Troubleshooting relaying

## Session Documentation

### 💾 [sessions/DEVELOPMENT_CACHE.md](./sessions/DEVELOPMENT_CACHE.md)
**Purpose**: Active development tracking
**Contents**:
- Current work items
- Recent discoveries
- Temporary workarounds
- Performance notes
- Testing gaps
- Next priorities

**When to use**:
- Resuming development
- Tracking progress
- Recording findings
- Planning next steps

## How to Use This Documentation

### For New Developers
1. Start with [PROJECT_BLUEPRINT.md](./PROJECT_BLUEPRINT.md) for system overview
2. Read [BUILD_AND_DEPLOYMENT.md](./BUILD_AND_DEPLOYMENT.md) to set up environment
3. Review relevant module blueprint for your area of work
4. Check [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) when encountering issues

### For Specific Tasks

#### Setting up development environment:
1. [BUILD_AND_DEPLOYMENT.md](./BUILD_AND_DEPLOYMENT.md) - Section 1: Local Development Build
2. [FILE_MAPPING.md](./FILE_MAPPING.md) - Configuration Files section
3. [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) - Development Issues section

#### Adding new features:
1. [PROJECT_BLUEPRINT.md](./PROJECT_BLUEPRINT.md) - Architecture overview
2. Relevant module blueprint (Frontend/API/etc.)
3. [API_INTERFACES.md](./API_INTERFACES.md) - If integrating with external services

#### Debugging issues:
1. [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) - Find specific issue
2. [FILE_MAPPING.md](./FILE_MAPPING.md) - Locate relevant files
3. Module blueprint - Understand component architecture

#### Deployment:
1. [BUILD_AND_DEPLOYMENT.md](./BUILD_AND_DEPLOYMENT.md) - Deployment procedures
2. [FILE_MAPPING.md](./FILE_MAPPING.md) - Environment-specific files
3. [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) - Production issues

### Documentation Maintenance

#### When to update documentation:
- After implementing significant features
- When discovering new issues/solutions
- After architectural changes
- When deployment processes change

#### Update process:
1. Make changes in relevant document
2. Update this INDEX if structure changes
3. Move items from DEVELOPMENT_CACHE to permanent docs
4. Keep documentation current with code

### Quick Links

#### Critical Files:
- Main config: `/.env`
- Docker setup: `/docker-compose.yml`
- Frontend entry: `/webapp/src/main.ts`
- API entry: `/relayer-middleware/api/cmd/server/main.go`

#### Key Commands:
```bash
# Start everything
./scripts/setup-and-launch.sh

# View documentation
ls -la .claude/

# Search documentation
grep -r "search-term" .claude/
```

#### Emergency References:
- [TROUBLESHOOTING.md#quick-reference](./TROUBLESHOOTING.md) - Service URLs and commands
- [BUILD_AND_DEPLOYMENT.md#rollback-procedures](./BUILD_AND_DEPLOYMENT.md) - Emergency rollback
- [API_INTERFACES.md#error-response-format](./API_INTERFACES.md) - Error codes

## Documentation Standards

### File Organization:
- Core docs in `.claude/`
- Module blueprints in `.claude/implementations/`
- Session notes in `.claude/sessions/`
- Research in `.claude/investigations/`

### Naming Conventions:
- UPPERCASE.md for permanent documentation
- Descriptive names indicating content
- Module blueprints include MODULE suffix
- Date prefixes for session files (YYYY-MM-DD)

### Content Guidelines:
- Clear section headers
- Code examples where relevant
- Practical "when to use" guidance
- Keep updated with code changes
- Include last updated dates

---
**Documentation Version**: 1.0
**Last Updated**: 2025-07-19
**Maintained By**: Development Team