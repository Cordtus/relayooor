# Internal Documentation System Template

## Overview
This template provides a standardized approach for maintaining comprehensive internal documentation across all projects. It should be integrated into your global CLAUDE.md to ensure consistent documentation practices.

## Documentation Structure

### Required Directory Structure
```
.claude/
├── PROJECT_BLUEPRINT.md        # Overall project architecture
├── BUILD_AND_DEPLOYMENT.md     # Build and deployment procedures
├── FILE_MAPPING.md            # Complete file reference
├── API_INTERFACES.md          # External API documentation
├── TROUBLESHOOTING.md         # Common issues and solutions
├── INDEX.md                   # Documentation index
├── implementations/           # Module-specific blueprints
│   ├── [MODULE]_MODULE.md     # Detailed module documentation
│   └── ...
├── investigations/            # Research and technical analysis
│   ├── [TOPIC]_RESEARCH.md    # Investigation findings
│   └── ...
└── sessions/                  # Development session notes
    ├── DEVELOPMENT_CACHE.md   # Current work tracking
    └── [DATE]_SESSION.md      # Session-specific notes
```

### Core Documentation Files

#### 1. PROJECT_BLUEPRINT.md
**Purpose**: Provide comprehensive project overview
**Required Sections**:
- Project Overview
- System Architecture (with diagrams)
- Key Components
- Security Architecture
- Deployment Architecture
- Development Workflow
- Configuration Management
- Future Enhancements
- Success Metrics
- Risk Mitigation

#### 2. BUILD_AND_DEPLOYMENT.md
**Purpose**: Complete build and deployment guide
**Required Sections**:
- Prerequisites
- Build Process (all environments)
- Deployment Configurations
- Environment Configuration
- Database Management
- Monitoring Setup
- SSL/TLS Configuration
- Health Checks
- Rollback Procedures
- Troubleshooting Deployment

#### 3. FILE_MAPPING.md
**Purpose**: Comprehensive file reference
**Required Sections**:
- Configuration Files
- Test Files
- Helper Scripts
- Utility Files
- Environment-Specific Files
- Key File Locations
- Naming Conventions
- Quick Reference Commands

#### 4. API_INTERFACES.md
**Purpose**: Document all external APIs
**Required Sections**:
- API Overview
- Authentication Methods
- Endpoint Documentation
- Request/Response Formats
- Error Handling
- Rate Limiting
- WebSocket Interfaces
- Integration Examples
- Testing Endpoints

#### 5. TROUBLESHOOTING.md
**Purpose**: Common issues and solutions
**Required Sections**:
- Development Issues
- Service-Specific Issues
- Production Issues
- Debugging Tools
- Recovery Procedures
- Monitoring and Alerts
- Quick Reference

#### 6. INDEX.md
**Purpose**: Central documentation reference
**Required Sections**:
- Overview
- Core Documentation Links
- Module Documentation Links
- How to Use Documentation
- Quick Links
- Documentation Standards

### Module Blueprints (implementations/)
**Purpose**: Detailed module-specific documentation
**Required Sections**:
- Module Overview
- Architecture
- Technology Stack
- Directory Structure
- Key Components
- Configuration
- Testing Strategy
- Common Issues
- Performance Considerations
- Security Considerations

### Development Cache (sessions/)
**Purpose**: Track ongoing work
**Required Sections**:
- Current Development Status
- Active Work Items
- Recent Discoveries
- Temporary Workarounds
- Performance Observations
- Testing Gaps
- Next Session Priorities
- Questions for Team

## Integration Instructions

### Add to Global CLAUDE.md

```markdown
<!-- Add this to your global CLAUDE.md -->

### Internal Documentation System

For every new project, create and maintain comprehensive internal documentation:

1. **Create documentation structure**:
   ```bash
   mkdir -p .claude/{implementations,investigations,sessions}
   touch .claude/{PROJECT_BLUEPRINT.md,BUILD_AND_DEPLOYMENT.md,FILE_MAPPING.md,API_INTERFACES.md,TROUBLESHOOTING.md,INDEX.md}
   touch .claude/sessions/DEVELOPMENT_CACHE.md
   ```

2. **Document as you build**:
   - Update PROJECT_BLUEPRINT when making architectural decisions
   - Document build steps in BUILD_AND_DEPLOYMENT
   - Map new files in FILE_MAPPING
   - Add troubleshooting solutions as discovered
   - Track active work in DEVELOPMENT_CACHE

3. **Maintain documentation**:
   - Review and update after significant changes
   - Move items from cache to permanent docs
   - Keep index current with all documentation

4. **Use documentation**:
   - Consult before starting new features
   - Reference when debugging issues
   - Update when learning new information

### Documentation Update Triggers

Update documentation when:
- Starting a new project → Create all core files
- Making architectural changes → Update PROJECT_BLUEPRINT
- Adding new services → Create module blueprint
- Discovering issues → Update TROUBLESHOOTING
- Changing deployment → Update BUILD_AND_DEPLOYMENT
- Adding dependencies → Update relevant sections
- Completing features → Move from cache to permanent docs

### Documentation Review Process

At the end of each session:
1. Review DEVELOPMENT_CACHE
2. Move completed items to appropriate permanent docs
3. Update INDEX if structure changed
4. Commit documentation changes with code
```

## Best Practices

### Writing Guidelines
1. **Be specific**: Include exact commands, file paths, and error messages
2. **Add examples**: Show code snippets and command outputs
3. **Explain why**: Not just what, but why decisions were made
4. **Keep current**: Update docs with code changes
5. **Use diagrams**: Visual representations for architecture
6. **Include dates**: Last updated timestamps

### Organization Tips
1. **Consistent naming**: Use UPPERCASE for permanent docs
2. **Clear sections**: Use descriptive headers
3. **Cross-reference**: Link between related documents
4. **Version tracking**: Note documentation version
5. **Quick access**: Provide command snippets

### Common Patterns

#### For Configuration Documentation
```markdown
### Configuration: [Service Name]

**File Location**: `/path/to/config.ext`

**Key Settings**:
```yaml
setting_name: value  # Description of what this does
```

**Environment Variables**:
- `VAR_NAME` - Description (default: value)

**Common Issues**:
- Issue: Description
  Solution: Steps to fix
```

#### For API Documentation
```markdown
### Endpoint: [Name]

**URL**: `METHOD /path/to/endpoint`
**Authentication**: Required/Optional
**Rate Limit**: X requests per minute

**Request**:
```json
{
  "field": "value"
}
```

**Response (200)**:
```json
{
  "result": "data"
}
```

**Error Codes**:
- 400: Bad Request - Invalid parameters
- 401: Unauthorized - Missing/invalid auth
```

#### For Troubleshooting
```markdown
### Issue: [Description]

**Symptoms**: What user sees
**Root Cause**: Why it happens

**Solution**:
```bash
# Commands to fix
command --with-options
```

**Prevention**: How to avoid in future
```

## Example Usage

### Starting New Project
```bash
# 1. Create structure
cd new-project
mkdir -p .claude/{implementations,investigations,sessions}

# 2. Create initial blueprint
cat > .claude/PROJECT_BLUEPRINT.md << 'EOF'
# Project Blueprint: [Project Name]

## Overview
[Project description]

## Architecture
[System design]
EOF

# 3. Start development cache
cat > .claude/sessions/DEVELOPMENT_CACHE.md << 'EOF'
# Development Cache

## Current Status ($(date +%Y-%m-%d))
- [ ] Initial setup
- [ ] Core features
EOF
```

### During Development
```bash
# Update cache with findings
echo "## Discovered Issues" >> .claude/sessions/DEVELOPMENT_CACHE.md
echo "- Database connection requires service name not localhost" >> .claude/sessions/DEVELOPMENT_CACHE.md

# Document new API
echo "## New Endpoint: /api/feature" >> .claude/API_INTERFACES.md
```

### End of Session
```bash
# Review and move items from cache
grep -E "^- \[x\]" .claude/sessions/DEVELOPMENT_CACHE.md
# Move completed items to permanent docs
# Update INDEX.md if needed
```

## Benefits

1. **Knowledge Preservation**: Never lose important discoveries
2. **Faster Onboarding**: New developers can self-serve
3. **Reduced Debugging Time**: Solutions already documented
4. **Better Architecture**: Decisions are recorded
5. **Improved Collaboration**: Shared understanding

---

This template can be customized for specific project types while maintaining the core structure. The key is consistency across all projects.