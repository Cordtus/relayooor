# CSS Optimization Plan

## Overview
This plan outlines the systematic optimization of CSS and styling in the Vue.js frontend to achieve consistency, reduce code bloat, and align with Tailwind/shadcn patterns.

## Current State Analysis

### Key Issues Identified
1. **Duplicate Components**: Multiple card implementations with inconsistent styling
2. **Hardcoded Colors**: 441+ instances of hardcoded color classes instead of semantic tokens
3. **Inconsistent Patterns**: Different styling approaches for similar UI elements
4. **Underutilized shadcn**: Only 4 components use the cn() utility function
5. **No Design System**: Lack of standardized variants for common patterns

### Code Bloat Statistics
- 15+ components implement their own card styles
- 5 instances of inline styles
- Multiple button style variations without reusable component
- Repeated status color mappings across components

## Optimization Strategy

### Phase 1: Component Standardization
1. Remove duplicate Card.vue, standardize on shadcn card
2. Create core UI components:
   - Button with variants (primary, secondary, ghost, danger)
   - Badge with status variants
   - EmptyState for consistent placeholder UI
   - MetricCard for dashboard widgets

### Phase 2: Design Token Implementation
1. Extend Tailwind configuration with semantic tokens
2. Define consistent color system for statuses
3. Standardize spacing scale usage
4. Implement typography hierarchy

### Phase 3: Pattern Extraction
1. Convert repeated patterns to utility classes
2. Remove inline styles
3. Consolidate animation/transition patterns
4. Standardize hover/focus states

### Phase 4: Component Migration
1. Update all components to use cn() utility
2. Replace conditional classes with variant props
3. Remove custom CSS where Tailwind utilities exist
4. Ensure TypeScript interfaces for all props

## Implementation Checklist

### Immediate Actions
- [ ] Create base UI components (Button, Badge, EmptyState)
- [ ] Update Tailwind config with semantic colors
- [ ] Define component utility classes
- [ ] Remove duplicate Card component

### Short-term Goals
- [ ] Migrate status colors to centralized system
- [ ] Standardize spacing across all components
- [ ] Implement consistent hover/focus states
- [ ] Create style guide documentation

### Long-term Improvements
- [ ] Full shadcn integration
- [ ] Dark mode support
- [ ] Component library documentation
- [ ] Automated style linting

## Expected Outcomes
- 50% reduction in CSS-related code
- Consistent UI across all components
- Easier theme customization
- Improved maintainability
- Better developer experience